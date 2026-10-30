package watcher_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/router-for-me/CLIProxyAPI/v8/internal/config"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/resourcecatalog"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/watcher"
	"gopkg.in/yaml.v3"
)

func watcherField(n *yaml.Node, name string) *yaml.Node {
	for i := 0; i+1 < len(n.Content); i += 2 {
		if n.Content[i].Value == name {
			return n.Content[i+1]
		}
	}
	return nil
}

func editWatchedSource(t *testing.T, path string, replace bool, edit func(*yaml.Node)) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var doc yaml.Node
	if err = yaml.Unmarshal(data, &doc); err != nil {
		t.Fatal(err)
	}
	edit(watcherField(watcherField(doc.Content[0], "api-keys"), "openai-compatibility").Content[0])
	data, err = yaml.Marshal(&doc)
	if err != nil {
		t.Fatal(err)
	}
	destination := path
	if replace {
		destination += ".external-save"
	}
	if err = os.WriteFile(destination, data, 0600); err != nil {
		t.Fatal(err)
	}
	if replace {
		if err = os.Rename(destination, path); err != nil {
			t.Fatal(err)
		}
	}
}

func TestResourceInventoryStartedWatcherSurvivesReplacements(t *testing.T) {
	dir := t.TempDir()
	authDir := filepath.Join(dir, "auth")
	if err := os.Mkdir(authDir, 0700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "config.yaml")
	text := fmt.Sprintf("server: {port: 0}\noauth: {auth-dir: %q}\nresource-catalog: {enabled: true}\napi-keys:\n  openai-compatibility: [{name: synthetic, base-url: https://synthetic.invalid, disabled: false, keys: [{api-key: synthetic-A}]}]\n", filepath.ToSlash(authDir))
	if err := os.WriteFile(path, []byte(text), 0600); err != nil {
		t.Fatal(err)
	}
	catalog, err := resourcecatalog.Open(path, authDir, "")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := catalog.Close(); err != nil {
			t.Error(err)
		}
	})
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := config.ParseConfigBytes(data)
	if err != nil {
		t.Fatal(err)
	}
	changes := make(chan *config.Config, 64)
	w, err := watcher.NewWatcher(path, authDir, func(next *config.Config) { changes <- next.CloneForRuntime() })
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(func() {
		cancel()
		if err := w.Stop(); err != nil {
			t.Error(err)
		}
	})
	w.SetConfig(cfg)
	w.SetResourceInventoryHook(func() error { return catalog.Reconcile(catalog.Version()) })
	if err = w.Start(ctx); err != nil {
		t.Fatal(err)
	}
	waitFor := func(keys int, disabled bool) {
		t.Helper()
		deadline := time.NewTimer(5 * time.Second)
		defer deadline.Stop()
		for {
			select {
			case next := <-changes:
				if len(next.OpenAICompatibility) == 1 && len(next.OpenAICompatibility[0].APIKeyEntries) == keys && next.OpenAICompatibility[0].Disabled == disabled {
					return
				}
			case <-deadline.C:
				t.Fatalf("started watcher stopped delivering runtime state: keys=%d disabled=%v", keys, disabled)
			}
		}
	}
	waitFor(1, false)
	editWatchedSource(t, path, false, func(group *yaml.Node) {
		var key yaml.Node
		if err := yaml.Unmarshal([]byte("{api-key: synthetic-B}"), &key); err != nil {
			t.Fatal(err)
		}
		keys := watcherField(group, "keys")
		keys.Content = append(keys.Content, key.Content[0])
	})
	waitFor(2, false)
	snapshot, err := catalog.Snapshot()
	if err != nil || len(snapshot.Resources) != 3 {
		t.Fatal("started watcher did not publish real marker migration")
	}
	setDisabled := func(value string) func(*yaml.Node) {
		return func(group *yaml.Node) {
			*watcherField(group, "disabled") = yaml.Node{Kind: yaml.ScalarNode, Tag: "!!bool", Value: value}
		}
	}
	editWatchedSource(t, path, false, setDisabled("true"))
	waitFor(2, true)
	if len(w.SnapshotCoreAuths()) != 0 {
		t.Fatal("manual disablement after owned replacement did not reach runtime")
	}
	editWatchedSource(t, path, true, setDisabled("false"))
	waitFor(2, false)
	if len(w.SnapshotCoreAuths()) != 2 {
		t.Fatal("external replacement did not reach runtime")
	}
	editWatchedSource(t, path, false, setDisabled("true"))
	waitFor(2, true)
	if len(w.SnapshotCoreAuths()) != 0 {
		t.Fatal("manual disablement after external replacement did not reach runtime")
	}
}
