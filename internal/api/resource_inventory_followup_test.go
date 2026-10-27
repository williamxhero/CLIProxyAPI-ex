package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/router-for-me/CLIProxyAPI/v8/internal/config"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/registry"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/runtime/executor"
	coreauth "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/auth"
	coreexecutor "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/executor"

	"gopkg.in/yaml.v3"
)

func TestResourceInventoryFirstInitializationRepair(t *testing.T) {
	dir := t.TempDir()
	authDir := filepath.Join(dir, "auth")
	if err := os.Mkdir(authDir, 0700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(authDir, "synthetic.json")
	if err := os.WriteFile(path, []byte(`{"type":"codex",`), 0600); err != nil {
		t.Fatal(err)
	}
	f := newInventoryHarness(t, dir)
	f.request("resources", 503)
	f.close()
	if err := os.WriteFile(path, []byte(`{"type":"codex","disabled":true,"access_token":"synthetic-token"}`), 0600); err != nil {
		t.Fatal(err)
	}
	f = newInventoryHarness(t, dir)
	repaired := f.snapshot()
	if len(repaired.Resources) != 5 {
		t.Fatalf("repaired initialization rows=%d want=5", len(repaired.Resources))
	}
	f.close()
	f = newInventoryHarness(t, dir)
	if restarted := f.snapshot(); restarted.InstanceID != repaired.InstanceID || restarted.InstanceIncarnation != repaired.InstanceIncarnation {
		t.Fatal("repaired initialization did not retain its issued instance")
	}
}

func TestResourceInventoryEstablishedStateLossRefused(t *testing.T) {
	for _, name := range []string{"catalog.json", "seal.key"} {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			f := newInventoryHarness(t, dir)
			f.snapshot()
			f.close()
			source, err := os.ReadFile(f.path)
			if err != nil {
				t.Fatal(err)
			}
			lost := filepath.Join(dir, ".resource-catalog", name)
			if err = os.Remove(lost); err != nil {
				t.Fatal(err)
			}
			f = newInventoryHarness(t, dir)
			f.request("resources", 503)
			if _, err = os.Stat(lost); !os.IsNotExist(err) {
				t.Fatal("lost established state was recreated")
			}
			after, err := os.ReadFile(f.path)
			if err != nil {
				t.Fatal(err)
			}
			if string(after) != string(source) {
				t.Fatal("failed reopen rewrote marked sources")
			}
		})
	}
}

func TestResourceInventoryManualScalarSpellings(t *testing.T) {
	for _, tc := range []struct{ name, disabled, weight string }{
		{"upper-hex", "TRUE", "0x0"},
		{"title-octal", "True", "0o0"},
		{"legacy-binary", "yes", "0b0"},
		{"legacy-plus", "ON", "+0"},
		{"negative-zero", "true", "-0"},
		{"leading-zero", "true", "00"},
		{"non-positive", "true", "-1"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			authDir := filepath.Join(dir, "auth")
			text := fmt.Sprintf(`server: {port: 0}
management: {secret-key: test-password}
oauth: {auth-dir: %q}
resource-catalog: {enabled: true}
api-keys:
  openai-compatibility:
    - name: synthetic-disabled
      base-url: https://synthetic.invalid
      disabled: %s
      keys: [{api-key: synthetic-A, weight: %s}, {api-key: synthetic-B, weight: 2}]
  codex:
    - name: synthetic-excluded
      base-url: https://synthetic.invalid
      excluded-models: ['*']
      keys: [{api-key: synthetic-C}]
`, filepath.ToSlash(authDir), tc.disabled, tc.weight)
			if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(text), 0600); err != nil {
				t.Fatal(err)
			}
			f := newInventoryHarness(t, dir)
			cfg := f.server.cfg
			if len(cfg.OpenAICompatibility) != 1 || !cfg.OpenAICompatibility[0].Disabled || cfg.OpenAICompatibility[0].APIKeyEntries[0].Weight == nil || *cfg.OpenAICompatibility[0].APIKeyEntries[0].Weight > 0 {
				t.Fatal("literal fixture did not express disabled/non-positive runtime state")
			}
			snapshot := f.snapshot()
			var compat, excluded int
			for _, row := range snapshot.Resources {
				switch row.Family {
				case "openai-compatibility":
					compat++
					if !row.Manual.Disabled {
						t.Error("inventory lost runtime disabled spelling or inheritance")
					}
					if row.Kind == "provider" && row.Manual.ZeroWeight {
						t.Error("child weight leaked into parent manual state")
					}
					if row.Kind == "provider" {
						first := rowByID(t, snapshot, row.Members[0])
						second := rowByID(t, snapshot, row.Members[1])
						if !first.Manual.ZeroWeight || second.Manual.ZeroWeight {
							t.Error("inventory lost explicit non-positive/positive weights")
						}
					}
				case "codex":
					excluded++
					if !row.Manual.Excluded {
						t.Error("inventory lost inherited exclusions")
					}
				}
			}
			if compat != 3 || excluded != 2 {
				t.Fatalf("incomplete literal inventory: compat=%d excluded=%d", compat, excluded)
			}
		})
	}
}

func TestResourceInventoryManualInheritanceOverrides(t *testing.T) {
	dir := t.TempDir()
	authDir := filepath.Join(dir, "auth")
	text := fmt.Sprintf(`server: {port: 0}
management: {secret-key: test-password}
oauth: {auth-dir: %q}
resource-catalog: {enabled: true}
api-keys:
  codex:
    - name: synthetic-excluded
      base-url: https://synthetic.invalid
      excluded-models: ['*']
      keys:
        - {api-key: synthetic-inherited}
        - {api-key: synthetic-override, excluded-models: []}
        - {api-key: synthetic-null, excluded-models: null}
`, filepath.ToSlash(authDir))
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(text), 0600); err != nil {
		t.Fatal(err)
	}
	f := newInventoryHarness(t, dir)
	if keys := f.server.cfg.CodexKey; len(keys) != 3 || len(keys[0].ExcludedModels) != 1 || len(keys[1].ExcludedModels) != 0 || len(keys[2].ExcludedModels) != 1 {
		t.Fatal("literal fixture did not express runtime inheritance/override")
	}
	snapshot := f.snapshot()
	group := snapshot.Resources[0]
	if !group.Manual.Excluded {
		t.Fatal("source group exclusion lost")
	}
	for i, want := range []bool{true, false, true} {
		child := rowByID(t, snapshot, group.Members[i])
		if child.Manual.Excluded != want {
			t.Errorf("child %d excluded=%v want=%v", i, child.Manual.Excluded, want)
		}
	}
}

func TestResourceInventoryRecoveryPhysicalReceipts(t *testing.T) {
	for _, copied := range []bool{false, true} {
		t.Run(fmt.Sprintf("copied=%v", copied), func(t *testing.T) {
			dir := t.TempDir()
			f := newInventoryHarness(t, dir)
			before := f.snapshot()
			authPath := filepath.Join(f.authDir, "synthetic-new.json")
			if err := os.WriteFile(authPath, []byte(`{"type":"codex","access_token":"synthetic-new-token"}`), 0600); err != nil {
				t.Fatal(err)
			}
			obstruction := fmt.Sprintf("%s.%s.%d.catalog-v1-backup", authPath, before.InstanceID, before.Revision+1)
			if err := os.Mkdir(obstruction, 0700); err != nil {
				t.Fatal(err)
			}
			editInventoryConfig(t, f.path, func(root *yaml.Node) {
				keys := testField(testField(testField(root, "api-keys"), "codex").Content[0], "keys")
				var added yaml.Node
				if err := yaml.Unmarshal([]byte("{api-key: synthetic-new-key}"), &added); err != nil {
					t.Fatal(err)
				}
				keys.Content = append(keys.Content, added.Content[0])
			})
			f.watcher.ReloadConfigIfChanged()
			f.request("resources", 503)
			published, err := os.ReadFile(f.path)
			if err != nil {
				t.Fatal(err)
			}
			var doc yaml.Node
			if err = yaml.Unmarshal(published, &doc); err != nil {
				t.Fatal(err)
			}
			keys := testField(testField(testField(doc.Content[0], "api-keys"), "codex").Content[0], "keys")
			if testField(keys.Content[1], "resource-id") == nil {
				t.Fatal("fixture did not publish the owned config markers before failure")
			}
			f.close()
			if err = os.Remove(obstruction); err != nil {
				t.Fatal(err)
			}
			if copied {
				data, err := os.ReadFile(f.path)
				if err != nil {
					t.Fatal(err)
				}
				replacement := filepath.Join(dir, "synthetic-copy.yaml")
				if err = os.WriteFile(replacement, data, 0600); err != nil {
					t.Fatal(err)
				}
				if err = os.Rename(replacement, f.path); err != nil {
					t.Fatal(err)
				}
			}
			f = newInventoryHarness(t, dir)
			if copied {
				f.request("resources", 503)
				return
			}
			after := f.snapshot()
			if after.InstanceID != before.InstanceID || len(after.Resources) != 6 {
				t.Fatal("owned source replacement did not recover its full inventory")
			}
			for _, old := range before.Resources {
				next := rowByID(t, after, old.ResourceID)
				if next.Identity != "valid" || next.Readiness != "ready" {
					t.Fatal("authorized replacement made existing resources ambiguous")
				}
				if old.Kind == "credential" && next.Target != old.Target {
					t.Fatal("owned marker insertion rebound an existing credential")
				}
			}
			f.close()
			f = newInventoryHarness(t, dir)
			if f.snapshot().Revision != after.Revision {
				t.Fatal("owned recovery was replayed")
			}
		})
	}
}

func TestResourceInventoryAmbiguousGroupsAllowManualSave(t *testing.T) {
	dir := t.TempDir()
	text := fmt.Sprintf("server: {port: 0}\nmanagement: {secret-key: test-password}\noauth: {auth-dir: %q}\nresource-catalog: {enabled: true}\napi-keys:\n  meta: [{name: first, base-url: 'https://synthetic.invalid', keys: [{api-key: synthetic-A}]}, {name: second, base-url: 'https://synthetic.invalid', keys: [{api-key: synthetic-B}]}]\n", filepath.ToSlash(filepath.Join(dir, "auth")))
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(text), 0600); err != nil {
		t.Fatal(err)
	}
	f := newInventoryHarness(t, dir)
	editInventoryConfig(t, f.path, func(root *yaml.Node) {
		groups := testField(testField(root, "api-keys"), "meta")
		var key, empty yaml.Node
		if err := yaml.Unmarshal([]byte("{api-key: ''}"), &key); err != nil {
			t.Fatal(err)
		}
		if err := yaml.Unmarshal([]byte("{name: empty, base-url: https://synthetic.invalid, keys: []}"), &empty); err != nil {
			t.Fatal(err)
		}
		keys := testField(groups.Content[0], "keys")
		keys.Content = append(keys.Content, key.Content[0])
		groups.Content = append(groups.Content, empty.Content[0])
	})
	if err := f.handler.ReconcileResourceInventory(); err != nil {
		t.Fatal(err)
	}
	f.close()
	var duplicateID string
	editInventoryConfig(t, f.path, func(root *yaml.Node) {
		groups := testField(testField(root, "api-keys"), "meta")
		duplicateID = testField(groups.Content[0], "resource-id").Value
		data, err := yaml.Marshal(groups.Content[0])
		if err != nil {
			t.Fatal(err)
		}
		var copied yaml.Node
		if err = yaml.Unmarshal(data, &copied); err != nil {
			t.Fatal(err)
		}
		groups.Content = append(groups.Content, copied.Content[0])
	})
	f = newInventoryHarness(t, dir)
	before := f.snapshot()
	data, err := os.ReadFile(f.path)
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := config.ParseConfigBytes(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.MetaKey) != 3 {
		t.Fatal("fixture did not keep both copied runtime keys")
	}
	cfg.MetaKey[0].ExcludedModels = []string{"*"}
	cfg.MetaKey[0], cfg.MetaKey[2] = cfg.MetaKey[2], cfg.MetaKey[0]
	if err = config.SaveConfigPreserveComments(f.path, cfg); err != nil {
		t.Fatalf("ambiguous identity blocked existing manual save: %v", err)
	}
	saved, err := os.ReadFile(f.path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(saved), "resource-source-ref") {
		t.Fatal("transient manual provenance escaped into source configuration")
	}
	var doc yaml.Node
	if err = yaml.Unmarshal(saved, &doc); err != nil {
		t.Fatal(err)
	}
	groups := testField(testField(doc.Content[0], "api-keys"), "meta")
	if len(groups.Content) != 4 {
		t.Fatalf("manual save dropped/regrouped providers: %d", len(groups.Content))
	}
	var copiedGroups, emptyKeys, emptyGroups, exclusions int
	for _, group := range groups.Content {
		if marker := testField(group, "resource-id"); marker != nil && marker.Value == duplicateID {
			copiedGroups++
		}
		keys := testField(group, "keys")
		if len(keys.Content) == 0 {
			emptyGroups++
		}
		for _, key := range keys.Content {
			if testField(key, "api-key").Value == "" {
				emptyKeys++
			}
			if excluded := testField(key, "excluded-models"); excluded != nil && len(excluded.Content) > 0 {
				exclusions++
			}
		}
	}
	if copiedGroups != 2 || emptyKeys != 2 || emptyGroups != 1 || exclusions != 1 {
		t.Fatalf("manual save lost or cross-bound copied sources: groups=%d keyless=%d empty=%d exclusions=%d", copiedGroups, emptyKeys, emptyGroups, exclusions)
	}
	f.watcher.ReloadConfigIfChanged()
	after := f.snapshot()
	if len(after.Resources) != len(before.Resources) {
		t.Fatal("manual save lost complete source inventory")
	}
	for _, row := range after.Resources {
		if row.ResourceID == duplicateID && (row.Identity != "ambiguous" || row.Readiness != "not_ready") {
			t.Fatal("manual save falsely validated duplicate identity")
		}
	}
	f.close()
	f = newInventoryHarness(t, dir)
	for _, row := range f.snapshot().Resources {
		if row.ResourceID == duplicateID && row.Identity != "ambiguous" {
			t.Fatal("restart falsely validated duplicate identity")
		}
	}
}

func TestResourceInventoryManualAuthScalars(t *testing.T) {
	for _, tc := range []struct {
		name, payload           string
		disabled, zero, invalid bool
	}{
		{"float-zero", `{"type":"codex","disabled":true,"weight":0.0}`, true, true, false},
		{"exponent-zero", `{"type":"codex","weight":0e0}`, false, true, false},
		{"negative-zero", `{"type":"codex","weight":-0.0}`, false, true, false},
		{"string-zero", `{"type":"codex","disabled":"true","weight":"0"}`, false, true, false},
		{"positive", `{"type":"codex","disabled":1,"weight":2}`, false, false, false},
		{"invalid-weight", `{"type":"codex","weight":"0.0"}`, false, false, true},
		{"fractional-weight", `{"type":"codex","weight":0.5}`, false, false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			authDir := filepath.Join(dir, "auth")
			if err := os.Mkdir(authDir, 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(authDir, "synthetic.json"), []byte(tc.payload), 0600); err != nil {
				t.Fatal(err)
			}
			f := newInventoryHarness(t, dir)
			if tc.invalid {
				f.request("resources", 503)
				return
			}
			runtime, ok := f.manager.GetByID("synthetic.json")
			if !ok || runtime.Disabled != tc.disabled {
				t.Fatal("literal auth fixture did not match runtime disabled semantics")
			}
			snapshot := f.snapshot()
			var found bool
			for _, row := range snapshot.Resources {
				if row.Kind == "credential" && row.ParentID == "" {
					found = true
					if row.Manual.Disabled != tc.disabled || row.Manual.ZeroWeight != tc.zero {
						t.Errorf("manual auth state=%+v want disabled=%v zero=%v", row.Manual, tc.disabled, tc.zero)
					}
				}
			}
			if !found {
				t.Fatal("literal auth source omitted")
			}
		})
	}
}

func resourceSaveFixture(t *testing.T, providers string) *inventoryHarness {
	t.Helper()
	dir := t.TempDir()
	text := fmt.Sprintf("server: {port: 0}\nmanagement: {secret-key: test-password}\noauth: {auth-dir: %q}\nresource-catalog: {enabled: true}\napi-keys:\n%s", filepath.ToSlash(filepath.Join(dir, "auth")), providers)
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(text), 0600); err != nil {
		t.Fatal(err)
	}
	f := newInventoryHarness(t, dir)
	f.snapshot()
	return f
}

func resourceSavedGroups(t *testing.T, path, family string) *yaml.Node {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "resource-source-ref") {
		t.Fatal("transient provenance leaked into persisted YAML")
	}
	var doc yaml.Node
	if err = yaml.Unmarshal(data, &doc); err != nil {
		t.Fatal(err)
	}
	return testField(testField(doc.Content[0], "api-keys"), family)
}

func resourceSourceValue(t *testing.T, node *yaml.Node) any {
	t.Helper()
	var value any
	if err := node.Decode(&value); err != nil {
		t.Fatal(err)
	}
	return value
}

func resourceSaveConfig(t *testing.T, f *inventoryHarness) *config.Config {
	t.Helper()
	data, err := os.ReadFile(f.path)
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := config.ParseConfigBytes(data)
	if err != nil {
		t.Fatal(err)
	}
	return cfg
}

func TestResourceInventoryCompatibilityFilteredProviderSave(t *testing.T) {
	for _, state := range []string{"disabled", "missing-base-url", "disabled-first", "missing-base-url-first"} {
		t.Run(state, func(t *testing.T) {
			var aCalls, bCalls atomic.Int32
			upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("Authorization") == "Bearer synthetic-B" {
					bCalls.Add(1)
				} else if r.Header.Get("Authorization") == "Bearer synthetic-A" {
					aCalls.Add(1)
				} else {
					t.Error("unexpected synthetic credential")
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"id":"synthetic","object":"chat.completion","choices":[{"index":0,"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}],"usage":{"prompt_tokens":1,"completion_tokens":1,"total_tokens":2}}`))
			}))
			defer upstream.Close()
			bState := "      disabled: true\n      base-url: " + upstream.URL + "/B\n"
			if strings.HasPrefix(state, "missing-base-url") {
				bState = ""
			}
			f := resourceSaveFixture(t, fmt.Sprintf("  openai-compatibility:\n    - name: first\n      base-url: %s\n      headers: {X-Synthetic: A}\n      models: [{name: synthetic-model}]\n      keys: [{api-key: synthetic-A}]\n    - name: filtered\n%s      headers: {X-Synthetic: B}\n      models: [{name: synthetic-model}]\n      keys: [{api-key: synthetic-B, weight: 3}]\n", upstream.URL, bState))
			f.close()
			var duplicateID string
			editInventoryConfig(t, f.path, func(root *yaml.Node) {
				groups := testField(testField(root, "api-keys"), "openai-compatibility")
				duplicateID = testField(groups.Content[0], "resource-id").Value
				testField(groups.Content[1], "resource-id").Value = duplicateID
				if strings.HasSuffix(state, "-first") {
					groups.Content[0], groups.Content[1] = groups.Content[1], groups.Content[0]
				}
			})
			f = newInventoryHarness(t, f.dir)
			before := f.snapshot()
			findB := func(groups *yaml.Node) *yaml.Node {
				for _, group := range groups.Content {
					if testField(group, "name").Value == "filtered" {
						return group
					}
				}
				t.Fatal("filtered provider source disappeared")
				return nil
			}
			originalB := resourceSourceValue(t, findB(resourceSavedGroups(t, f.path, "openai-compatibility")))
			if got := len(f.watcher.SnapshotCoreAuths()); got != 1 {
				t.Fatalf("fixture runtime auths=%d want=1", got)
			}
			cfg := resourceSaveConfig(t, f)
			for i := range cfg.OpenAICompatibility {
				if cfg.OpenAICompatibility[i].Name == "first" {
					cfg.OpenAICompatibility[i].Priority = 1
				}
			}
			if err := config.SaveConfigPreserveComments(f.path, cfg); err != nil {
				t.Fatal(err)
			}
			f.close()
			f = newInventoryHarness(t, f.dir)
			after := f.snapshot()
			groups := resourceSavedGroups(t, f.path, "openai-compatibility")
			if len(groups.Content) != 2 {
				t.Errorf("typed edit dropped filtered provider: groups=%d want=2", len(groups.Content))
			} else if !reflect.DeepEqual(resourceSourceValue(t, findB(groups)), originalB) {
				t.Error("typed edit changed filtered provider endpoint/manual/header/key data")
			}
			if len(after.Resources) != len(before.Resources) {
				t.Error("typed edit lost source inventory")
			}
			for _, row := range after.Resources {
				if row.ResourceID == duplicateID && (row.Identity != "ambiguous" || row.Readiness != "not_ready") {
					t.Error("duplicate provider became valid")
				}
			}
			auths := f.watcher.SnapshotCoreAuths()
			if len(auths) != 1 {
				t.Errorf("filtered B became routable: synthesized auths=%d want=1", len(auths))
			}
			f.manager.SetConfig(f.server.cfg)
			for _, auth := range f.manager.List() {
				if auth.Attributes["api_key"] != "synthetic-A" || auth.Attributes["base_url"] != upstream.URL {
					t.Error("filtered B acquired an enabled runtime auth")
				}
				f.manager.RegisterExecutor(executor.NewOpenAICompatExecutor(auth.Provider, f.server.cfg))
				registry.GetGlobalRegistry().RegisterClient(auth.ID, auth.Provider, []*registry.ModelInfo{{ID: "synthetic-model"}})
				id := auth.ID
				t.Cleanup(func() { registry.GetGlobalRegistry().UnregisterClient(id) })
			}
			for range 4 {
				_, err := f.manager.Execute(context.Background(), []string{"openai-compatible-first"}, coreexecutor.Request{Model: "synthetic-model", Payload: []byte(`{"model":"synthetic-model","messages":[{"role":"user","content":"synthetic"}]}`)}, coreexecutor.Options{SourceFormat: "openai"})
				if err != nil {
					t.Fatalf("controlled A request failed: %v", err)
				}
			}
			if aCalls.Load() != 4 || bCalls.Load() != 0 {
				t.Errorf("controlled routing: A=%d B=%d want A=4 B=0", aCalls.Load(), bCalls.Load())
			}
			encoded, err := json.Marshal(f.server.cfg)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(string(encoded), "resource-source-ref") {
				t.Fatal("transient provenance leaked into JSON")
			}
		})
	}
}

func TestResourceInventoryFilteredMemberProjection(t *testing.T) {
	for _, shape := range []string{"across-groups", "within-group", "missing-ids", "missing-child-id", "missing-group-id"} {
		t.Run(shape, func(t *testing.T) {
			f := resourceSaveFixture(t, "  meta:\n    - name: live\n      base-url: https://synthetic.invalid\n      keys: [{api-key: synthetic-A}]\n    - name: filtered\n      base-url: https://synthetic.invalid\n      keys: [{api-key: '', headers: {X-Synthetic: filtered}}]\n    - name: other\n      base-url: https://synthetic.invalid\n      keys: [{api-key: synthetic-C}]\n    - name: empty\n      keys: []\n")
			f.close()
			editInventoryConfig(t, f.path, func(root *yaml.Node) {
				groups := testField(testField(root, "api-keys"), "meta")
				live := testField(groups.Content[0], "keys").Content[0]
				filtered := testField(groups.Content[1], "keys").Content[0]
				testField(filtered, "resource-id").Value = testField(live, "resource-id").Value
				if shape == "within-group" {
					keys := testField(groups.Content[0], "keys")
					keys.Content = append(keys.Content, filtered)
					groups.Content = append(groups.Content[:1], groups.Content[2:]...)
				}
				if strings.HasPrefix(shape, "missing-") {
					for _, group := range groups.Content {
						for i := 0; i+1 < len(group.Content); i += 2 {
							if shape != "missing-child-id" && group.Content[i].Value == "resource-id" {
								group.Content = append(group.Content[:i], group.Content[i+2:]...)
								break
							}
						}
						for _, key := range testField(group, "keys").Content {
							for i := 0; i+1 < len(key.Content); i += 2 {
								if shape != "missing-group-id" && key.Content[i].Value == "resource-id" {
									key.Content = append(key.Content[:i], key.Content[i+2:]...)
									break
								}
							}
						}
					}
				}
			})
			if !strings.HasPrefix(shape, "missing-") {
				f = newInventoryHarness(t, f.dir)
				f.request("config/api-keys", 200)
			}
			groups := resourceSavedGroups(t, f.path, "meta")
			wantGroups := len(groups.Content)
			var filteredValue any
			for _, group := range groups.Content {
				for _, key := range testField(group, "keys").Content {
					if testField(key, "api-key").Value == "" {
						filteredValue = resourceSourceValue(t, key)
					}
				}
			}
			cfg := resourceSaveConfig(t, f)
			if len(cfg.MetaKey) != 2 {
				t.Fatal("fixture must contain two represented live rows")
			}
			weight := 4
			cfg.MetaKey[0].Weight = &weight
			cfg.MetaKey[0].APIKey = "synthetic-A-edited"
			cfg.MetaKey[0], cfg.MetaKey[1] = cfg.MetaKey[1], cfg.MetaKey[0]
			if err := config.SaveConfigPreserveComments(f.path, cfg); err != nil {
				t.Fatal(err)
			}
			groups = resourceSavedGroups(t, f.path, "meta")
			if len(groups.Content) != wantGroups {
				t.Errorf("typed edit/reorder lost providers: got=%d want=%d", len(groups.Content), wantGroups)
			}
			checkFiltered := func() {
				t.Helper()
				var filtered, empty int
				for _, group := range resourceSavedGroups(t, f.path, "meta").Content {
					keys := testField(group, "keys")
					if len(keys.Content) == 0 {
						empty++
					}
					for _, key := range keys.Content {
						if testField(key, "api-key").Value == "" {
							filtered++
							if !reflect.DeepEqual(resourceSourceValue(t, key), filteredValue) {
								t.Error("filtered child source data changed")
							}
						}
					}
				}
				if filtered != 1 || empty != 1 {
					t.Errorf("typed save lost unrepresented source entries: keyless=%d empty=%d want=1,1", filtered, empty)
				}
			}
			checkFiltered()
			cfg = resourceSaveConfig(t, f)
			if len(cfg.MetaKey) != 2 || cfg.MetaKey[0].APIKey != "synthetic-C" || cfg.MetaKey[1].APIKey != "synthetic-A-edited" || cfg.MetaKey[1].Weight == nil || *cfg.MetaKey[1].Weight != 4 {
				t.Fatal("typed edit/reorder did not survive parsing")
			}
			cfg.MetaKey = cfg.MetaKey[:1]
			if err := config.SaveConfigPreserveComments(f.path, cfg); err != nil {
				t.Fatal(err)
			}
			checkFiltered()
			cfg = resourceSaveConfig(t, f)
			if len(cfg.MetaKey) != 1 || cfg.MetaKey[0].APIKey != "synthetic-C" {
				t.Error("explicit represented deletion was restored or deleted another row")
			}
		})
	}
}

func resourceProjectedRows(t *testing.T, data []byte, field string, edit bool) []string {
	t.Helper()
	var cfg config.Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		t.Fatal(err)
	}
	keys := reflect.ValueOf(&cfg).Elem().FieldByName(field)
	if edit {
		weight := 4
		keys.Index(0).FieldByName("Weight").Set(reflect.ValueOf(&weight))
	}
	var rows []string
	for i := 0; i < keys.Len(); i++ {
		data, err := json.Marshal(keys.Index(i).Interface())
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(data), "resource-source-ref") {
			t.Fatal("transient provenance leaked into public typed JSON")
		}
		rows = append(rows, string(data))
	}
	sort.Strings(rows)
	return rows
}

func TestResourceInventoryFilteredNullInheritance(t *testing.T) {
	for _, family := range []struct{ name, field string }{
		{"gemini", "GeminiKey"}, {"interactions", "InteractionsKey"}, {"vertex", "VertexCompatAPIKey"}, {"codex", "CodexKey"}, {"claude", "ClaudeKey"}, {"xai", "XAIKey"}, {"meta", "MetaKey"},
	} {
		for _, duplicate := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/duplicate=%v", family.name, duplicate), func(t *testing.T) {
				f := resourceSaveFixture(t, fmt.Sprintf(`  %s:
    - name: inherited
      base-url: https://synthetic.invalid
      excluded-models: ['*']
      headers: {X-Synthetic: inherited}
      priority: 7
      prefix: inherited
      proxy-url: http://synthetic.invalid:80
      models: [{name: synthetic-model, alias: synthetic-model}]
      disable-cooling: true
      request-retry: 2
      request-scoped-errors: [{status: 400, match: [synthetic], action: stop}]
      keys:
        - {api-key: synthetic-A}
        - {api-key: '', excluded-models: null, headers: null, priority: null, prefix: null, proxy-url: null, models: null, disable-cooling: null, request-retry: null, request-scoped-errors: null}
        - {api-key: ''}
        - {api-key: '', excluded-models: [], headers: {}, models: [], request-scoped-errors: []}
        - {api-key: synthetic-C, excluded-models: [], headers: {}, models: [], request-scoped-errors: []}
`, family.name))
				if duplicate {
					f.close()
					editInventoryConfig(t, f.path, func(root *yaml.Node) {
						groups := testField(testField(root, "api-keys"), family.name)
						data, err := yaml.Marshal(groups.Content[0])
						if err != nil {
							t.Fatal(err)
						}
						var copy yaml.Node
						if err = yaml.Unmarshal(data, &copy); err != nil {
							t.Fatal(err)
						}
						groups.Content = append(groups.Content, copy.Content[0])
					})
					f = newInventoryHarness(t, f.dir)
				}
				before := f.snapshot()
				shared := map[string]any{}
				for _, field := range []string{"excluded-models", "headers", "priority", "prefix", "proxy-url", "models", "disable-cooling", "request-retry", "request-scoped-errors"} {
					shared[field] = resourceSourceValue(t, testField(resourceSavedGroups(t, f.path, family.name).Content[0], field))
				}
				data, err := os.ReadFile(f.path)
				if err != nil {
					t.Fatal(err)
				}
				want := resourceProjectedRows(t, data, family.field, true)
				cfg := resourceSaveConfig(t, f)
				keys := reflect.ValueOf(cfg).Elem().FieldByName(family.field)
				if keys.Len() == 0 {
					t.Fatal("fixture has no represented live key")
				}
				weight := 4
				keys.Index(0).FieldByName("Weight").Set(reflect.ValueOf(&weight))
				if err = config.SaveConfigPreserveComments(f.path, cfg); err != nil {
					t.Fatal(err)
				}
				data, err = os.ReadFile(f.path)
				if err != nil {
					t.Fatal(err)
				}
				if got := resourceProjectedRows(t, data, family.field, false); !reflect.DeepEqual(got, want) {
					t.Error("typed weight save changed effective inherited/explicit-empty fields")
				}
				for _, group := range resourceSavedGroups(t, f.path, family.name).Content {
					for field, want := range shared {
						value := testField(group, field)
						if value == nil || !reflect.DeepEqual(resourceSourceValue(t, value), want) {
							t.Errorf("unrelated weight edit lost group default %s", field)
						}
					}
				}
				f.close()
				f = newInventoryHarness(t, f.dir)
				after := f.snapshot()
				beforeExcluded, afterExcluded := 0, 0
				for _, row := range before.Resources {
					if row.Kind == "credential" && row.Manual.Excluded {
						beforeExcluded++
					}
				}
				for _, row := range after.Resources {
					if row.Kind == "credential" && row.Manual.Excluded {
						afterExcluded++
					}
				}
				wantExcluded := 3
				if duplicate {
					wantExcluded *= 2
				}
				if beforeExcluded != wantExcluded {
					t.Fatalf("fixture exclusions=%d want=%d", beforeExcluded, wantExcluded)
				}
				if len(after.Resources) != len(before.Resources) || afterExcluded != wantExcluded {
					t.Errorf("typed weight save changed credential exclusions: got=%d want=%d", afterExcluded, wantExcluded)
				}
			})
		}
	}
}

func TestResourceInventoryManualProjectionRefusal(t *testing.T) {
	for _, family := range []string{"meta", "openai-compatibility"} {
		for _, reason := range []string{"unknown", "missing", "stale", "stale-delete-all", "stale-delete-sibling", "unknown-child", "missing-child", "wrong-provider", "group-ref-is-child"} {
			if family == "meta" && strings.Contains(reason, "child") || family == "meta" && reason == "wrong-provider" {
				continue
			}
			t.Run(family+"/"+reason, func(t *testing.T) {
				f := resourceSaveFixture(t, fmt.Sprintf("  %s:\n    - name: first\n      base-url: https://synthetic.invalid\n      keys: [{api-key: synthetic-A}]\n    - name: second\n      base-url: https://synthetic.invalid\n      keys: [{api-key: synthetic-B}]\n", family))
				cfg := resourceSaveConfig(t, f).CloneForRuntime()
				weight := 4
				if family == "meta" {
					cfg.MetaKey[0].Weight = &weight
					switch reason {
					case "unknown":
						cfg.MetaKey[0].ResourceSourceRef = "unrecognized"
					case "missing":
						cfg.MetaKey[0].ResourceSourceRef = ""
					}
				} else {
					group := &cfg.OpenAICompatibility[0]
					group.Priority = 4
					switch reason {
					case "unknown":
						group.ResourceSourceRef = "unrecognized"
					case "missing":
						group.ResourceSourceRef = ""
					case "unknown-child":
						group.APIKeyEntries[0].ResourceSourceRef = "unrecognized"
					case "missing-child":
						group.APIKeyEntries[0].ResourceSourceRef = ""
					case "wrong-provider":
						group.APIKeyEntries[0].ResourceSourceRef = cfg.OpenAICompatibility[1].APIKeyEntries[0].ResourceSourceRef
					case "group-ref-is-child":
						group.ResourceSourceRef = group.APIKeyEntries[0].ResourceSourceRef
						group.APIKeyEntries = nil
					}
				}
				if reason == "stale-delete-all" {
					if family == "meta" {
						cfg.MetaKey = nil
					} else {
						cfg.OpenAICompatibility = nil
					}
				}
				if reason == "stale-delete-sibling" {
					if family == "meta" {
						cfg.MetaKey = cfg.MetaKey[1:]
					} else {
						cfg.OpenAICompatibility = cfg.OpenAICompatibility[1:]
					}
				}
				if strings.HasPrefix(reason, "stale") {
					editInventoryConfig(t, f.path, func(root *yaml.Node) {
						group := testField(testField(root, "api-keys"), family).Content[0]
						testField(group, "name").Value = "external-edit"
						if strings.HasPrefix(reason, "stale-delete-") {
							var added yaml.Node
							if err := yaml.Unmarshal([]byte("{api-key: synthetic-concurrent}"), &added); err != nil {
								t.Fatal(err)
							}
							keys := testField(group, "keys")
							keys.Content = append(keys.Content, added.Content[0])
						}
					})
				}
				before, err := os.ReadFile(f.path)
				if err != nil {
					t.Fatal(err)
				}
				err = config.SaveConfigPreserveComments(f.path, cfg)
				if err == nil {
					t.Fatal("unknown/missing/stale projection was silently accepted")
				}
				if strings.Contains(err.Error(), "synthetic") || strings.Contains(err.Error(), "unrecognized") {
					t.Error("projection refusal exposed source material")
				}
				after, errRead := os.ReadFile(f.path)
				if errRead != nil {
					t.Fatal(errRead)
				}
				if string(before) != string(after) {
					t.Fatal("refused save modified source bytes")
				}
			})
		}
	}
}

func TestResourceInventoryCompatibilityProjectionEdits(t *testing.T) {
	for _, shape := range []string{"unique", "duplicate", "missing-ids"} {
		t.Run(shape, func(t *testing.T) {
			f := resourceSaveFixture(t, `  openai-compatibility:
    - name: first
      base-url: https://synthetic.invalid/A
      keys: [{api-key: synthetic-A1}, {api-key: synthetic-A2}]
    - name: filtered
      headers: {X-Synthetic: B}
      keys: [{api-key: synthetic-B}]
    - name: disabled
      disabled: true
      base-url: https://synthetic.invalid/C
      keys: [{api-key: synthetic-C}]
    - name: deleted
      base-url: https://synthetic.invalid/D
      keys: [{api-key: synthetic-D}]
    - name: empty
      keys: []
`)
			f.close()
			if shape != "unique" {
				editInventoryConfig(t, f.path, func(root *yaml.Node) {
					groups := testField(testField(root, "api-keys"), "openai-compatibility")
					if shape == "duplicate" {
						testField(groups.Content[1], "resource-id").Value = testField(groups.Content[0], "resource-id").Value
						keys := testField(groups.Content[0], "keys")
						testField(keys.Content[1], "resource-id").Value = testField(keys.Content[0], "resource-id").Value
					} else {
						var removeMarkers func(*yaml.Node)
						removeMarkers = func(n *yaml.Node) {
							if n.Kind == yaml.MappingNode {
								for i := 0; i+1 < len(n.Content); i += 2 {
									if n.Content[i].Value == "resource-id" {
										n.Content = append(n.Content[:i], n.Content[i+2:]...)
										i -= 2
									}
								}
							}
							for _, child := range n.Content {
								removeMarkers(child)
							}
						}
						removeMarkers(groups)
					}
				})
			}
			groups := resourceSavedGroups(t, f.path, "openai-compatibility")
			originalB := resourceSourceValue(t, groups.Content[1])
			originalC := resourceSourceValue(t, groups.Content[2])
			originalEmpty := resourceSourceValue(t, groups.Content[4])
			cfg := resourceSaveConfig(t, f)
			if len(cfg.OpenAICompatibility) != 3 {
				t.Fatal("fixture must expose A, disabled C and D to typed saves")
			}
			group := &cfg.OpenAICompatibility[0]
			group.Name = "renamed"
			group.BaseURL = "https://synthetic.invalid/edited"
			group.Headers = map[string]string{"X-Synthetic": "edited"}
			group.APIKeyEntries = group.APIKeyEntries[1:]
			group.APIKeyEntries[0].APIKey = "synthetic-A2-edited"
			cfg.OpenAICompatibility = []config.OpenAICompatibility{cfg.OpenAICompatibility[1], cfg.OpenAICompatibility[0]}
			if err := config.SaveConfigPreserveComments(f.path, cfg); err != nil {
				t.Fatal(err)
			}
			groups = resourceSavedGroups(t, f.path, "openai-compatibility")
			if len(groups.Content) != 4 {
				t.Fatalf("explicit D deletion or filtered groups lost: groups=%d want=4", len(groups.Content))
			}
			if !reflect.DeepEqual(resourceSourceValue(t, groups.Content[0]), originalC) || !reflect.DeepEqual(resourceSourceValue(t, groups.Content[2]), originalB) || !reflect.DeepEqual(resourceSourceValue(t, groups.Content[3]), originalEmpty) {
				t.Fatal("typed reorder/edit modified untouched disabled/filtered/empty groups")
			}
			keys := testField(groups.Content[1], "keys")
			if len(keys.Content) != 1 || testField(keys.Content[0], "api-key").Value != "synthetic-A2-edited" {
				t.Fatal("explicit child deletion/key edit was not preserved")
			}
			f = newInventoryHarness(t, f.dir)
			f.snapshot()
			auths := f.watcher.SnapshotCoreAuths()
			if len(auths) != 1 || auths[0].Attributes["api_key"] != "synthetic-A2-edited" || auths[0].Attributes["base_url"] != "https://synthetic.invalid/edited" || auths[0].Label != "renamed" {
				t.Fatal("typed projection edits changed unrelated routing eligibility or lost rename/endpoint edit")
			}
		})
	}
}

func TestResourceInventoryManualProjectionDeleteAll(t *testing.T) {
	for _, family := range []string{"meta", "openai-compatibility"} {
		t.Run(family, func(t *testing.T) {
			f := resourceSaveFixture(t, fmt.Sprintf("  %s:\n    - name: represented\n      base-url: https://synthetic.invalid\n      keys: [{api-key: synthetic-A}]\n    - name: filtered\n      keys: [{api-key: '', headers: {X-Synthetic: retained}}]\n    - name: empty\n      keys: []\n", family))
			groups := resourceSavedGroups(t, f.path, family)
			filtered := resourceSourceValue(t, groups.Content[1])
			empty := resourceSourceValue(t, groups.Content[2])
			cfg := resourceSaveConfig(t, f).CloneForRuntime()
			if family == "meta" {
				cfg.MetaKey = nil
			} else {
				cfg.OpenAICompatibility = nil
			}
			if err := config.SaveConfigPreserveComments(f.path, cfg); err != nil {
				t.Fatalf("fresh explicit delete-all was blocked: %v", err)
			}
			groups = resourceSavedGroups(t, f.path, family)
			if len(groups.Content) != 2 || !reflect.DeepEqual(resourceSourceValue(t, groups.Content[0]), filtered) || !reflect.DeepEqual(resourceSourceValue(t, groups.Content[1]), empty) {
				t.Fatal("fresh delete-all lost filtered/empty source groups or restored a represented deletion")
			}
			f.close()
			f = newInventoryHarness(t, f.dir)
			f.snapshot()
			if len(f.watcher.SnapshotCoreAuths()) != 0 || len(f.manager.List()) != 0 {
				t.Fatal("delete-all made a filtered source routable or restored a represented auth")
			}
		})
	}
}

func TestResourceInventoryInheritedManualClears(t *testing.T) {
	for _, duplicate := range []bool{false, true} {
		t.Run(fmt.Sprintf("duplicate=%v", duplicate), func(t *testing.T) {
			f := resourceSaveFixture(t, `  meta:
    - name: inherited
      base-url: https://synthetic.invalid
      priority: 7
      prefix: inherited
      proxy-url: http://synthetic.invalid:80
      excluded-models: ['*']
      headers: {X-Synthetic: inherited}
      models: [{name: synthetic-model, alias: synthetic-model}]
      request-scoped-errors: [{status: 400, match: [synthetic], action: stop}]
      disable-cooling: true
      request-retry: 2
      keys:
        - {api-key: synthetic-A}
        - {api-key: '', disable-cooling: null, request-retry: null}
`)
			if duplicate {
				f.close()
				editInventoryConfig(t, f.path, func(root *yaml.Node) {
					groups := testField(testField(root, "api-keys"), "meta")
					data, err := yaml.Marshal(groups.Content[0])
					if err != nil {
						t.Fatal(err)
					}
					var copy yaml.Node
					if err = yaml.Unmarshal(data, &copy); err != nil {
						t.Fatal(err)
					}
					groups.Content = append(groups.Content, copy.Content[0])
				})
				f = newInventoryHarness(t, f.dir)
				f.snapshot()
			}
			cfg := resourceSaveConfig(t, f)
			weight := 4
			key := &cfg.MetaKey[0]
			key.Weight = &weight
			key.Priority = 0
			key.Prefix = ""
			key.ProxyURL = ""
			key.ExcludedModels = nil
			key.Headers = nil
			key.Models = nil
			key.RequestScopedErrors = nil
			key.DisableCooling = nil
			key.RequestRetry = nil
			if err := config.SaveConfigPreserveComments(f.path, cfg); err != nil {
				t.Fatal(err)
			}
			data, err := os.ReadFile(f.path)
			if err != nil {
				t.Fatal(err)
			}
			var full config.Config
			if err = yaml.Unmarshal(data, &full); err != nil {
				t.Fatal(err)
			}
			changed := 0
			for _, key := range full.MetaKey {
				if key.Weight != nil && *key.Weight == 4 {
					changed++
					if key.Priority != 0 || key.Prefix != "" || key.ProxyURL != "" || len(key.ExcludedModels) != 0 || len(key.Headers) != 0 || len(key.Models) != 0 || len(key.RequestScopedErrors) != 0 || key.DisableCooling != nil || key.RequestRetry != nil {
						t.Fatal("typed explicit clear re-inherited an old shared setting")
					}
				} else if key.Priority != 7 || key.Prefix != "inherited" || key.ProxyURL != "http://synthetic.invalid:80" || len(key.ExcludedModels) != 1 || key.Headers["X-Synthetic"] != "inherited" || len(key.Models) != 1 || len(key.RequestScopedErrors) != 1 || key.DisableCooling == nil || !*key.DisableCooling || key.RequestRetry == nil || *key.RequestRetry != 2 {
					t.Fatal("typed clear changed an untouched represented/filtered sibling")
				}
			}
			if changed != 1 {
				t.Fatalf("explicit clear changed %d rows want=1", changed)
			}
			for _, group := range resourceSavedGroups(t, f.path, "meta").Content {
				if field := testField(group, "excluded-models"); field == nil || len(field.Content) != 1 || field.Content[0].Value != "*" {
					t.Fatal("explicit child empty override erased the shared default")
				}
			}
		})
	}
}

// Reusing a loaded writer is supported; only an external edit makes it stale.
func TestResourceInventoryOwnedSaveContinuity(t *testing.T) {
	for _, family := range []string{"gemini", "openai-compatibility"} {
		t.Run(family, func(t *testing.T) {
			f := resourceSaveFixture(t, fmt.Sprintf("  %s:\n    - name: live\n      base-url: https://synthetic.invalid\n      keys: [{api-key: synthetic-A}, {api-key: synthetic-B}]\n", family))
			cfg := resourceSaveConfig(t, f)
			weight := 2
			if family == "gemini" {
				cfg.GeminiKey[0].Weight = &weight
			} else {
				cfg.OpenAICompatibility[0].APIKeyEntries[0].Weight = &weight
			}
			if err := config.SaveConfigPreserveComments(f.path, cfg); err != nil {
				t.Fatalf("first owned save: %v", err)
			}
			if family == "gemini" {
				cfg.GeminiKey[0].Priority = 3
			} else {
				cfg.OpenAICompatibility[0].Priority = 3
			}
			if err := config.SaveConfigPreserveComments(f.path, cfg); err != nil {
				t.Fatalf("second owned save without external mutation: %v", err)
			}
			loaded := resourceSaveConfig(t, f)
			if family == "gemini" {
				if loaded.GeminiKey[0].Priority != 3 || *loaded.GeminiKey[0].Weight != 2 {
					t.Fatal("successive native edits were not persisted")
				}
			} else if loaded.OpenAICompatibility[0].Priority != 3 || *loaded.OpenAICompatibility[0].APIKeyEntries[0].Weight != 2 {
				t.Fatal("successive compatibility edits were not persisted")
			}
		})
	}
}

func TestResourceInventoryManualSaveContinuity(t *testing.T) {
	for _, family := range []string{"gemini", "openai-compatibility"} {
		t.Run(family, func(t *testing.T) {
			f := resourceSaveFixture(t, fmt.Sprintf("  %s:\n    - name: live\n      base-url: https://synthetic.invalid\n      keys: [{api-key: synthetic-A}]\n", family))
			f.handler.SetConfig(resourceSaveConfig(t, f))
			route := "openai-compatibility"
			if family == "gemini" {
				route = "gemini-api-key"
			}
			for _, priority := range []int{2, 3} {
				resourceManualPatch(t, f, route, map[string]any{"priority": priority})
			}
		})
	}
}

func TestResourceInventoryOwnedSaveProjectionLifecycle(t *testing.T) {
	for _, family := range []string{"gemini", "openai-compatibility"} {
		t.Run(family, func(t *testing.T) {
			f := resourceSaveFixture(t, fmt.Sprintf("  %s:\n    - name: first\n      base-url: https://synthetic.invalid\n      keys: [{api-key: synthetic-A}, {api-key: synthetic-B}]\n    - name: second\n      base-url: https://synthetic.invalid\n      keys: [{api-key: synthetic-C}]\n    - name: filtered\n      keys: [{api-key: '', weight: 0}]\n", family))
			groups := resourceSavedGroups(t, f.path, family)
			filtered := resourceSourceValue(t, groups.Content[2])
			cfg := resourceSaveConfig(t, f)
			cfg.Home.Enabled = true
			cfg.CodexResponseSteering = true
			cfg.Debug = true
			older := cfg.CloneForRuntime()
			if family == "gemini" {
				cfg.GeminiKey[0].Priority = 2
			} else {
				cfg.OpenAICompatibility[0].Priority = 2
			}
			if err := config.SaveConfigPreserveComments(f.path, cfg); err != nil {
				t.Fatal(err)
			}
			if !cfg.Home.Enabled || !cfg.CodexResponseSteering || !cfg.Debug {
				t.Fatal("owned save overwrote runtime-only or unrelated caller values")
			}
			refuse := func(writer *config.Config) {
				t.Helper()
				before, err := os.ReadFile(f.path)
				if err != nil {
					t.Fatal(err)
				}
				if err = config.SaveConfigPreserveComments(f.path, writer); err == nil {
					t.Fatal("a different writer's stale basis was accepted")
				}
				after, err := os.ReadFile(f.path)
				if err != nil || string(after) != string(before) {
					t.Fatal("refused writer changed persisted source")
				}
			}
			refuse(older)
			fresh := cfg.CloneForRuntime()
			if family == "gemini" {
				fresh.GeminiKey[0], fresh.GeminiKey[2] = fresh.GeminiKey[2], fresh.GeminiKey[0]
			} else {
				fresh.OpenAICompatibility[0], fresh.OpenAICompatibility[1] = fresh.OpenAICompatibility[1], fresh.OpenAICompatibility[0]
				keys := fresh.OpenAICompatibility[1].APIKeyEntries
				keys[0], keys[1] = keys[1], keys[0]
			}
			if err := config.SaveConfigPreserveComments(f.path, fresh); err != nil {
				t.Fatalf("fresh runtime clone reorder: %v", err)
			}
			refuse(cfg)
			cfg = fresh
			if family == "gemini" {
				cfg.GeminiKey = cfg.GeminiKey[:2]
			} else {
				cfg.OpenAICompatibility[1].APIKeyEntries = cfg.OpenAICompatibility[1].APIKeyEntries[:1]
			}
			if err := config.SaveConfigPreserveComments(f.path, cfg); err != nil {
				t.Fatalf("represented child deletion after reorder: %v", err)
			}
			loaded := resourceSaveConfig(t, f)
			var keys []string
			if family == "gemini" {
				for _, key := range loaded.GeminiKey {
					keys = append(keys, key.APIKey)
				}
			} else {
				for _, group := range loaded.OpenAICompatibility {
					for _, key := range group.APIKeyEntries {
						keys = append(keys, key.APIKey)
					}
				}
			}
			if !reflect.DeepEqual(keys, []string{"synthetic-C", "synthetic-B"}) {
				t.Fatalf("reordered deletion cross-bound rows: got %v", keys)
			}
			if family == "gemini" {
				cfg.GeminiKey = nil
			} else {
				cfg.OpenAICompatibility = nil
			}
			if err := config.SaveConfigPreserveComments(f.path, cfg); err != nil {
				t.Fatalf("delete all represented rows: %v", err)
			}
			if err := config.SaveConfigPreserveComments(f.path, cfg.CloneForRuntime()); err != nil {
				t.Fatalf("fresh empty projection: %v", err)
			}
			groups = resourceSavedGroups(t, f.path, family)
			if len(groups.Content) != 1 || !reflect.DeepEqual(resourceSourceValue(t, groups.Content[0]), filtered) {
				t.Fatal("successive saves changed a filtered source")
			}
			editInventoryConfig(t, f.path, func(root *yaml.Node) {
				group := testField(testField(root, "api-keys"), family).Content[0]
				var added yaml.Node
				if err := yaml.Unmarshal([]byte("{api-key: synthetic-external}"), &added); err != nil {
					t.Fatal(err)
				}
				list := testField(group, "keys")
				list.Content = append(list.Content, added.Content[0])
			})
			refuse(cfg)
			refuse(cfg.CloneForRuntime())
		})
	}
}

func TestResourceInventoryNativeEndpointClear(t *testing.T) {
	for _, family := range []struct{ name, field string }{
		{"gemini", "GeminiKey"}, {"interactions", "InteractionsKey"}, {"vertex", "VertexCompatAPIKey"}, {"codex", "CodexKey"}, {"claude", "ClaudeKey"}, {"xai", "XAIKey"}, {"meta", "MetaKey"},
	} {
		for _, duplicate := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/duplicate=%v", family.name, duplicate), func(t *testing.T) {
				f := resourceSaveFixture(t, fmt.Sprintf("  %s:\n    - name: first\n      base-url: https://synthetic.invalid/original\n      keys: [{api-key: synthetic-A}]\n    - name: second\n      base-url: https://synthetic.invalid/untouched\n      keys: [{api-key: synthetic-B}]\n", family.name))
				if duplicate {
					f.close()
					editInventoryConfig(t, f.path, func(root *yaml.Node) {
						groups := testField(testField(root, "api-keys"), family.name)
						testField(groups.Content[1], "resource-id").Value = testField(groups.Content[0], "resource-id").Value
						testField(testField(groups.Content[1], "keys").Content[0], "resource-id").Value = testField(testField(groups.Content[0], "keys").Content[0], "resource-id").Value
					})
					f = newInventoryHarness(t, f.dir)
				}
				before := resourceSavedGroups(t, f.path, family.name)
				untouched := resourceSourceValue(t, before.Content[1])
				childID := testField(testField(before.Content[0], "keys").Content[0], "resource-id").Value
				cfg := resourceSaveConfig(t, f)
				for _, endpoint := range []string{"https://synthetic.invalid/edited", ""} {
					rows := reflect.ValueOf(cfg).Elem().FieldByName(family.field)
					rows.Index(0).FieldByName("BaseURL").SetString(endpoint)
					if err := config.SaveConfigPreserveComments(f.path, cfg); err != nil {
						t.Fatalf("save endpoint %q: %v", endpoint, err)
					}
					data, err := os.ReadFile(f.path)
					if err != nil {
						t.Fatal(err)
					}
					var raw config.Config
					if err = yaml.Unmarshal(data, &raw); err != nil {
						t.Fatal(err)
					}
					got := reflect.ValueOf(raw).FieldByName(family.field).Index(0).FieldByName("BaseURL").String()
					if got != endpoint {
						t.Fatalf("saved endpoint reloaded as %q, want explicit %q", got, endpoint)
					}
				}
				after := resourceSavedGroups(t, f.path, family.name)
				if !reflect.DeepEqual(resourceSourceValue(t, after.Content[1]), untouched) || testField(testField(after.Content[0], "keys").Content[0], "resource-id").Value != childID {
					t.Fatal("endpoint edit cross-bound a source or replaced its marker")
				}
				f.close()
				f = newInventoryHarness(t, f.dir)
				wantCount := 2
				wantBase := ""
				if family.name == "codex" || family.name == "xai" {
					wantCount = 1
				}
				if family.name == "meta" {
					wantBase = "https://api.meta.ai/v1"
				}
				auths := f.watcher.SnapshotCoreAuths()
				if len(auths) != wantCount {
					t.Fatalf("provider runtime after clear auths=%d want=%d", len(auths), wantCount)
				}
				for _, auth := range auths {
					if auth.Attributes["api_key"] == "synthetic-A" && auth.Attributes["base_url"] != wantBase {
						t.Fatalf("clear did not restore parser/provider default: %q want %q", auth.Attributes["base_url"], wantBase)
					}
				}
			})
		}
	}
}

func TestResourceInventoryOwnedSaveFilteredTransition(t *testing.T) {
	for _, family := range []string{"codex", "openai-compatibility"} {
		t.Run(family, func(t *testing.T) {
			f := resourceSaveFixture(t, fmt.Sprintf("  %s:\n    - name: first\n      base-url: https://synthetic.invalid/original\n      keys: [{api-key: synthetic-A}]\n    - name: second\n      base-url: https://synthetic.invalid/untouched\n      keys: [{api-key: synthetic-B}]\n", family))
			cfg := resourceSaveConfig(t, f)
			if family == "codex" {
				cfg.CodexKey[0].BaseURL = ""
			} else {
				cfg.OpenAICompatibility[0].BaseURL = ""
			}
			if err := config.SaveConfigPreserveComments(f.path, cfg); err != nil {
				t.Fatal(err)
			}
			if family == "codex" {
				cfg.CodexKey[0].Priority = 2
			} else {
				cfg.OpenAICompatibility[0].Priority = 2
			}
			if err := config.SaveConfigPreserveComments(f.path, cfg); err != nil {
				t.Fatalf("repeat owned save of a now-filtered row: %v", err)
			}
			groups := resourceSavedGroups(t, f.path, family)
			if len(groups.Content) != 2 || len(testField(groups.Content[0], "keys").Content) != 1 {
				t.Fatalf("owned filtered transition duplicated its source: groups=%d first-members=%d", len(groups.Content), len(testField(groups.Content[0], "keys").Content))
			}
			f.close()
			f = newInventoryHarness(t, f.dir)
			auths := f.watcher.SnapshotCoreAuths()
			if len(auths) != 1 || auths[0].Attributes["api_key"] != "synthetic-B" {
				t.Fatal("successive save re-enabled the filtered credential/provider")
			}
		})
	}
}

func TestResourceInventoryEffectiveCredentialContinuity(t *testing.T) {
	for _, family := range []struct{ name, field string }{
		{"gemini", "GeminiKey"}, {"interactions", "InteractionsKey"}, {"vertex", "VertexCompatAPIKey"}, {"codex", "CodexKey"}, {"claude", "ClaudeKey"}, {"xai", "XAIKey"}, {"meta", "MetaKey"},
	} {
		for _, tc := range []struct {
			name, child string
			wantHeaders map[string]string
		}{
			{"null-inherits", ", headers: null", map[string]string{"X-Synthetic": "inherited"}},
			{"omitted-inherits", "", map[string]string{"X-Synthetic": "inherited"}},
			{"empty-overrides", ", headers: {}", nil},
		} {
			t.Run(family.name+"/"+tc.name, func(t *testing.T) {
				f := resourceSaveFixture(t, fmt.Sprintf("  %s:\n    - name: inherited\n      base-url: https://synthetic.invalid\n      headers: {X-Synthetic: inherited}\n      keys: [{api-key: synthetic-A%s}]\n", family.name, tc.child))
				before := f.snapshot().Resources[1]
				if before.Kind != "credential" || before.IdentityVersion != 1 || before.CredentialVersion != 1 || before.ConfigVersion != 1 {
					t.Fatalf("initial issued credential versions: %+v", before.Target)
				}
				cfg := resourceSaveConfig(t, f)
				key := reflect.ValueOf(cfg).Elem().FieldByName(family.field).Index(0)
				if got := key.FieldByName("Headers").Interface(); !reflect.DeepEqual(got, tc.wantHeaders) {
					t.Fatalf("fixture parser headers=%v want=%v", got, tc.wantHeaders)
				}
				weight := 2
				key.FieldByName("Weight").Set(reflect.ValueOf(&weight))
				if err := config.SaveConfigPreserveComments(f.path, cfg); err != nil {
					t.Fatal(err)
				}
				if err := f.handler.ReconcileResourceInventory(); err != nil {
					t.Fatal(err)
				}
				after := rowByID(t, f.snapshot(), before.ResourceID)
				if after.Incarnation != before.Incarnation || after.IdentityVersion != 1 || after.CredentialVersion != 1 || after.ConfigVersion != 2 {
					t.Fatalf("weight-only save replaced effective credential: before=%+v after=%+v", before.Target, after.Target)
				}
				if got := reflect.ValueOf(resourceSaveConfig(t, f)).Elem().FieldByName(family.field).Index(0).FieldByName("Headers").Interface(); !reflect.DeepEqual(got, tc.wantHeaders) {
					t.Fatalf("saved parser headers=%v want=%v", got, tc.wantHeaders)
				}
				f.close()
				f = newInventoryHarness(t, f.dir)
				if restarted := rowByID(t, f.snapshot(), before.ResourceID); restarted.Target != after.Target {
					t.Fatal("unchanged restart altered effective credential fences")
				}
			})
		}
	}
}

func TestResourceInventoryEffectiveCredentialDefaults(t *testing.T) {
	for _, tc := range []struct{ name, family, settings string }{
		{"null-group-headers", "gemini", "      base-url: https://synthetic.invalid\n      headers: null\n"},
		{"empty-group-headers", "gemini", "      base-url: https://synthetic.invalid\n      headers: {}\n"},
		{"null-endpoint", "gemini", "      base-url: null\n"},
		{"meta-parser-default", "meta", ""},
		{"normalized-headers", "gemini", "      base-url: https://synthetic.invalid\n      headers: {' X-Synthetic ': ' inherited '}\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := resourceSaveFixture(t, fmt.Sprintf("  %s:\n    - name: effective\n%s      keys: [{api-key: synthetic-A}]\n", tc.family, tc.settings))
			before := f.snapshot().Resources[1]
			cfg := resourceSaveConfig(t, f)
			weight := 2
			if tc.family == "gemini" {
				cfg.GeminiKey[0].Weight = &weight
				want := map[string]string(nil)
				if tc.name == "normalized-headers" {
					want = map[string]string{"X-Synthetic": "inherited"}
				}
				if !reflect.DeepEqual(cfg.GeminiKey[0].Headers, want) {
					t.Fatal("fixture did not have the independently expected effective headers")
				}
			} else {
				cfg.MetaKey[0].Weight = &weight
				if cfg.MetaKey[0].BaseURL != "https://api.meta.ai/v1" {
					t.Fatal("fixture did not have the literal Meta parser default")
				}
			}
			if err := config.SaveConfigPreserveComments(f.path, cfg); err != nil {
				t.Fatal(err)
			}
			if err := f.handler.ReconcileResourceInventory(); err != nil {
				t.Fatal(err)
			}
			after := rowByID(t, f.snapshot(), before.ResourceID)
			if after.Incarnation != before.Incarnation || after.IdentityVersion != 1 || after.CredentialVersion != 1 || after.ConfigVersion != 2 {
				t.Fatalf("materializing effective defaults replaced credentials: before=%+v after=%+v", before.Target, after.Target)
			}
		})
	}
}

func TestResourceInventoryNativeEndpointConflict(t *testing.T) {
	f := resourceSaveFixture(t, "  gemini:\n    - name: shared\n      base-url: https://synthetic.invalid/original\n      keys: [{api-key: synthetic-A}, {api-key: synthetic-B}]\n")
	cfg := resourceSaveConfig(t, f)
	before, err := os.ReadFile(f.path)
	if err != nil {
		t.Fatal(err)
	}
	cfg.GeminiKey[0].BaseURL = ""
	if err = config.SaveConfigPreserveComments(f.path, cfg); err == nil {
		t.Fatal("a per-member endpoint conflict was silently accepted")
	}
	after, err := os.ReadFile(f.path)
	if err != nil || string(after) != string(before) {
		t.Fatal("rejected endpoint conflict changed source bytes")
	}
	cfg.GeminiKey[1].BaseURL = ""
	if err = config.SaveConfigPreserveComments(f.path, cfg); err != nil {
		t.Fatalf("clearing the whole shared endpoint: %v", err)
	}
	cfg.GeminiKey[0].Priority = 2
	if err = config.SaveConfigPreserveComments(f.path, cfg); err != nil {
		t.Fatalf("repeat save after clearing a shared endpoint: %v", err)
	}
	for _, key := range resourceSaveConfig(t, f).GeminiKey {
		if key.BaseURL != "" {
			t.Fatal("shared endpoint clear restored the old endpoint")
		}
	}
}

// The read-only attribute belongs only to this synthetic fixture, not an ACL.
func TestResourceInventoryOwnedSaveWriteFailure(t *testing.T) {
	for _, family := range []string{"gemini", "openai-compatibility"} {
		t.Run(family, func(t *testing.T) {
			f := resourceSaveFixture(t, fmt.Sprintf("  %s:\n    - name: first\n      base-url: https://synthetic.invalid\n      keys: [{api-key: synthetic-A}]\n", family))
			cfg := resourceSaveConfig(t, f)
			if family == "gemini" {
				cfg.GeminiKey[0].Priority = 2
			} else {
				cfg.OpenAICompatibility[0].Priority = 2
			}
			before, err := os.ReadFile(f.path)
			if err != nil {
				t.Fatal(err)
			}
			if err = os.Chmod(f.path, 0400); err != nil {
				t.Fatal(err)
			}
			errSave := config.SaveConfigPreserveComments(f.path, cfg)
			if err = os.Chmod(f.path, 0600); err != nil {
				t.Fatal(err)
			}
			if errSave == nil {
				t.Fatal("fixture did not reject the write")
			}
			after, err := os.ReadFile(f.path)
			if err != nil || string(after) != string(before) {
				t.Fatal("failed write changed source bytes")
			}
			if err = config.SaveConfigPreserveComments(f.path, cfg); err != nil {
				t.Fatalf("failed write adopted a false basis: %v", err)
			}
			if err = config.SaveConfigPreserveComments(f.path, cfg.CloneForRuntime()); err != nil {
				t.Fatalf("fresh clone after successful persistence: %v", err)
			}
		})
	}
}

func TestResourceInventoryManualCredentialChanges(t *testing.T) {
	type observedRequest struct{ path, header, key string }
	observed := make(chan observedRequest, 8)
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		observed <- observedRequest{r.URL.Path, r.Header.Get("X-Synthetic"), r.Header.Get("x-goog-api-key")}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"candidates":[{"content":{"role":"model","parts":[{"text":"ok"}]},"finishReason":"STOP"}],"usageMetadata":{"promptTokenCount":1,"candidatesTokenCount":1,"totalTokenCount":2}}`))
	}))
	defer upstream.Close()
	f := resourceSaveFixture(t, fmt.Sprintf("  gemini:\n    - name: inherited\n      base-url: %s/initial\n      headers: {X-Synthetic: inherited}\n      keys: [{api-key: synthetic-A, headers: null}]\n", upstream.URL))
	f.handler.SetConfig(resourceSaveConfig(t, f))
	before := f.snapshot().Resources[1]
	previous := before
	for _, step := range []struct {
		body, path, header, key   string
		credential, configuration uint64
		rotate                    bool
	}{
		{`{"weight":2}`, "/initial/", "inherited", "synthetic-A", 1, 2, false},
		{`{"headers":{"X-Synthetic":"changed"}}`, "/initial/", "changed", "synthetic-A", 2, 3, true},
		{fmt.Sprintf(`{"base-url":%q}`, upstream.URL+"/next"), "/next/", "changed", "synthetic-A", 3, 4, true},
		{`{"api-key":"synthetic-B"}`, "/next/", "changed", "synthetic-B", 4, 5, true},
		{`{"priority":3}`, "/next/", "changed", "synthetic-B", 4, 6, false},
	} {
		var patch map[string]any
		if err := json.Unmarshal([]byte(step.body), &patch); err != nil {
			t.Fatal(err)
		}
		resourceManualPatch(t, f, "gemini", patch)
		if err := f.handler.ReconcileResourceInventory(); err != nil {
			t.Fatal(err)
		}
		after := rowByID(t, f.snapshot(), before.ResourceID)
		if after.CredentialVersion != step.credential || after.IdentityVersion != step.credential || after.ConfigVersion != step.configuration || (after.Incarnation != previous.Incarnation) != step.rotate {
			t.Fatalf("manual credential/config fences for %s: before=%+v after=%+v", step.body, previous.Target, after.Target)
		}
		view := resourceSaveConfig(t, f)
		f.watcher.SetConfig(view)
		for _, auth := range f.manager.List() {
			f.manager.Remove(context.Background(), auth.ID)
			registry.GetGlobalRegistry().UnregisterClient(auth.ID)
		}
		for _, auth := range f.watcher.SnapshotCoreAuths() {
			if _, err := f.manager.Register(coreauth.WithSkipPersist(context.Background()), auth); err != nil {
				t.Fatal(err)
			}
			registry.GetGlobalRegistry().RegisterClient(auth.ID, auth.Provider, []*registry.ModelInfo{{ID: "synthetic-model"}})
			id := auth.ID
			t.Cleanup(func() { registry.GetGlobalRegistry().UnregisterClient(id) })
		}
		f.manager.RegisterExecutor(executor.NewGeminiExecutor(view))
		if _, err := f.manager.Execute(context.Background(), []string{"gemini"}, coreexecutor.Request{Model: "synthetic-model", Payload: []byte(`{"contents":[{"role":"user","parts":[{"text":"synthetic"}]}]}`)}, coreexecutor.Options{SourceFormat: "gemini"}); err != nil {
			t.Fatalf("controlled upstream after manual edit: %v", err)
		}
		got := <-observed
		if !strings.HasPrefix(got.path, step.path) || got.header != step.header || got.key != step.key {
			t.Fatalf("effective upstream credential=%+v want path=%s header=%s key=%s", got, step.path, step.header, step.key)
		}
		previous = after
	}
	f.close()
	f = newInventoryHarness(t, f.dir)
	if restarted := rowByID(t, f.snapshot(), before.ResourceID); restarted.Target != previous.Target {
		t.Fatal("restart changed the final manual credential/config fences")
	}
}

func resourceManualJSON(t *testing.T, f *inventoryHarness, method, route string, body any, want int) []byte {
	t.Helper()
	family := strings.TrimSuffix(route, "-api-key")
	var data []byte
	if body != nil {
		var err error
		data, err = json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
	}
	req := httptest.NewRequest(method, "/v8/management/config/api-keys/"+family, strings.NewReader(string(data)))
	req.RemoteAddr = "127.0.0.1:1234"
	req.Header.Set("Authorization", "Bearer test-password")
	req.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	f.server.engine.ServeHTTP(response, req)
	if response.Code != want {
		t.Fatalf("%s %s status=%d want=%d body=%s", method, route, response.Code, want, response.Body.String())
	}
	return response.Body.Bytes()
}

// Native v8 lists are source documents, not filtered runtime replacement arrays.
// Editing a child uses GET/edit/PATCH of that real canonical list.
func resourceManualPatch(t *testing.T, f *inventoryHarness, route string, patch map[string]any) {
	t.Helper()
	rows := resourceManualArray(t, f, route)
	for key, value := range patch {
		if key == "base-url" || route == "openai-compatibility" {
			rows[0][key] = value
		} else {
			rows[0]["keys"].([]any)[0].(map[string]any)[key] = value
		}
	}
	resourceManualJSON(t, f, http.MethodPatch, route, rows, http.StatusOK)
}

func resourceManualArray(t *testing.T, f *inventoryHarness, route string) []map[string]any {
	t.Helper()
	data := resourceManualJSON(t, f, http.MethodGet, route, nil, http.StatusOK)
	if strings.Contains(string(data), "resource-source-ref") || strings.Contains(string(data), "resource-group-") {
		t.Fatal("manual GET exposed private provenance")
	}
	var result []map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatal(err)
	}
	return result
}

func TestResourceInventoryManualJSONRoundTrip(t *testing.T) {
	for _, family := range []string{"gemini", "meta", "openai-compatibility"} {
		t.Run(family, func(t *testing.T) {
			f := resourceSaveFixture(t, fmt.Sprintf("  %s:\n    - name: first\n      base-url: https://synthetic.invalid/A\n      keys: [{api-key: synthetic-A}, {api-key: synthetic-B}, {api-key: '', weight: 0}]\n    - name: second\n      base-url: https://synthetic.invalid/C\n      keys: [{api-key: synthetic-C}]\n    - name: filtered\n      headers: {X-Synthetic: retained}\n      keys: [{api-key: ''}]\n", family))
			f.handler.SetConfig(resourceSaveConfig(t, f))
			groups := resourceSavedGroups(t, f.path, family)
			groupID := testField(groups.Content[0], "resource-id").Value
			childID := testField(testField(groups.Content[0], "keys").Content[0], "resource-id").Value
			filteredKey := resourceSourceValue(t, testField(groups.Content[0], "keys").Content[2])
			filteredGroup := resourceSourceValue(t, groups.Content[2])
			rows := resourceManualArray(t, f, family)
			if len(rows) != 3 || len(rows[0]["keys"].([]any)) != 3 {
				t.Fatal("v8 GET omitted filtered source rows")
			}
			if family == "openai-compatibility" {
				rows[0]["disabled"] = true
			} else {
				rows[0]["keys"].([]any)[0].(map[string]any)["excluded-models"] = []string{"*"}
			}
			resourceManualJSON(t, f, http.MethodPut, family, rows, http.StatusOK)
			if err := f.handler.ReconcileResourceInventory(); err != nil {
				t.Fatal(err)
			}
			snapshot := f.snapshot()
			if family == "openai-compatibility" {
				if !rowByID(t, snapshot, groupID).Manual.Disabled || !rowByID(t, snapshot, childID).Manual.Disabled {
					t.Fatal("JSON round-trip lost provider/child manual disablement")
				}
			} else if !rowByID(t, snapshot, childID).Manual.Excluded {
				t.Fatal("JSON round-trip lost native manual exclusion")
			}
			rows = resourceManualArray(t, f, family)
			if family == "openai-compatibility" {
				rows[0]["disabled"] = false
			} else {
				rows[0]["keys"].([]any)[0].(map[string]any)["excluded-models"] = []string{}
			}
			resourceManualJSON(t, f, http.MethodPatch, family, rows, http.StatusOK)
			rows = resourceManualArray(t, f, family)
			rows[0], rows[1] = rows[1], rows[0]
			resourceManualJSON(t, f, http.MethodPut, family, rows, http.StatusOK)
			rows = resourceManualArray(t, f, family)
			children := rows[1]["keys"].([]any)
			children[0].(map[string]any)["api-key"] = "synthetic-replaced"
			children = append(children[:1], children[2:]...)
			children = append(children, map[string]any{"api-key": "synthetic-new"})
			rows[1]["keys"] = children
			rows = append(rows, map[string]any{"name": "new-provider", "base-url": "https://synthetic.invalid/new", "keys": []map[string]any{{"api-key": "synthetic-new-provider"}}})
			resourceManualJSON(t, f, http.MethodPut, family, rows, http.StatusOK)
			groups = resourceSavedGroups(t, f.path, family)
			var foundFiltered, foundOriginal bool
			var replaced, added int
			for _, group := range groups.Content {
				for _, key := range testField(group, "keys").Content {
					switch testField(key, "api-key").Value {
					case "synthetic-replaced":
						replaced++
					case "synthetic-new":
						added++
					}
				}
				if marker := testField(group, "resource-id"); marker != nil && marker.Value == groupID {
					foundOriginal = true
					var kept bool
					for _, key := range testField(group, "keys").Content {
						if reflect.DeepEqual(resourceSourceValue(t, key), filteredKey) {
							kept = true
						}
					}
					if !kept {
						t.Fatal("JSON edits lost the keyless sibling")
					}
				}
				if reflect.DeepEqual(resourceSourceValue(t, group), filteredGroup) {
					foundFiltered = true
				}
			}
			if !foundOriginal || replaced != 1 || added != 1 || !foundFiltered {
				t.Fatal("JSON replacement/addition lost, duplicated or changed source context")
			}
			// Native API edits explicitly retain raw filtered sources in the source list.
			rows = resourceManualArray(t, f, family)
			resourceManualJSON(t, f, http.MethodPut, family, rows[2:3], http.StatusOK)
			if err := f.handler.ReconcileResourceInventory(); err != nil {
				t.Fatal(err)
			}
			f.watcher.SetConfig(resourceSaveConfig(t, f))
			if auths := f.watcher.SnapshotCoreAuths(); len(auths) != 0 {
				t.Fatalf("JSON deletion retained represented runtime auths: %d", len(auths))
			}
			groups = resourceSavedGroups(t, f.path, family)
			if len(groups.Content) != 1 || !reflect.DeepEqual(resourceSourceValue(t, groups.Content[0]), filteredGroup) {
				t.Fatal("JSON deletion restored represented providers or changed external filtered source")
			}
		})
	}
}

func TestResourceInventoryManualJSONRefusal(t *testing.T) {
	for _, family := range []string{"gemini", "openai-compatibility"} {
		for _, reason := range []string{"unknown-id", "copied-id", "stale", "cross-provider"} {
			t.Run(family+"/"+reason, func(t *testing.T) {
				f := resourceSaveFixture(t, fmt.Sprintf("  %s:\n    - name: first\n      base-url: https://synthetic.invalid/A\n      keys: [{api-key: synthetic-A}]\n    - name: second\n      base-url: https://synthetic.invalid/B\n      keys: [{api-key: synthetic-B}]\n", family))
				f.handler.SetConfig(resourceSaveConfig(t, f))
				rows := resourceManualArray(t, f, family)
				child := rows[0]["keys"].([]any)[0].(map[string]any)
				switch reason {
				case "unknown-id":
					child["resource-id"] = "not-an-issued-source"
				case "copied-id":
					rows[1]["keys"].([]any)[0].(map[string]any)["resource-id"] = child["resource-id"]
				case "cross-provider":
					rows[1]["keys"] = append(rows[1]["keys"].([]any), child)
				case "stale":
					editInventoryConfig(t, f.path, func(root *yaml.Node) {
						testField(testField(testField(root, "api-keys"), family).Content[0], "name").Value = "external-edit"
					})
				}
				before, err := os.ReadFile(f.path)
				if err != nil {
					t.Fatal(err)
				}
				resourceManualJSON(t, f, http.MethodPut, family, rows, http.StatusInternalServerError)
				after, err := os.ReadFile(f.path)
				if err != nil || string(after) != string(before) {
					t.Fatal("refused JSON replacement changed source bytes")
				}
			})
		}
	}
}

func TestResourceInventoryManualJSONRuntime(t *testing.T) {
	for _, family := range []string{"gemini", "openai-compatibility"} {
		t.Run(family, func(t *testing.T) {
			var calls atomic.Int32
			observed := make(chan string, 1)
			upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				w.Header().Set("Content-Type", "application/json")
				if family == "gemini" {
					observed <- r.Header.Get("x-goog-api-key")
					_, _ = w.Write([]byte(`{"candidates":[{"content":{"role":"model","parts":[{"text":"ok"}]},"finishReason":"STOP"}],"usageMetadata":{"promptTokenCount":1,"candidatesTokenCount":1,"totalTokenCount":2}}`))
				} else {
					observed <- strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
					_, _ = w.Write([]byte(`{"id":"synthetic","object":"chat.completion","choices":[{"index":0,"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}],"usage":{"prompt_tokens":1,"completion_tokens":1,"total_tokens":2}}`))
				}
			}))
			defer upstream.Close()
			f := resourceSaveFixture(t, fmt.Sprintf("  %s:\n    - name: first\n      base-url: %s\n      models: [{name: synthetic-model}]\n      keys: [{api-key: synthetic-A}]\n", family, upstream.URL))
			f.handler.SetConfig(resourceSaveConfig(t, f))
			before := f.snapshot().Resources[1]
			route := "openai-compatibility"
			if family == "gemini" {
				route = "gemini-api-key"
			}
			rows := resourceManualArray(t, f, route)
			if family == "gemini" {
				rows[0]["keys"].([]any)[0].(map[string]any)["api-key"] = "synthetic-replaced"
			} else {
				rows[0]["keys"].([]any)[0].(map[string]any)["api-key"] = "synthetic-replaced"
			}
			resourceManualJSON(t, f, http.MethodPut, route, rows, http.StatusOK)
			if err := f.handler.ReconcileResourceInventory(); err != nil {
				t.Fatal(err)
			}
			after := rowByID(t, f.snapshot(), before.ResourceID)
			if after.Identity != "valid" || after.CredentialVersion != before.CredentialVersion+1 || after.IdentityVersion != before.IdentityVersion+1 || after.Incarnation == before.Incarnation {
				t.Fatal("JSON credential replacement reused old credential evidence")
			}
			view := resourceSaveConfig(t, f)
			f.watcher.SetConfig(view)
			f.manager.SetConfig(view)
			for _, auth := range f.manager.List() {
				f.manager.Remove(context.Background(), auth.ID)
				registry.GetGlobalRegistry().UnregisterClient(auth.ID)
			}
			auths := f.watcher.SnapshotCoreAuths()
			if len(auths) != 1 {
				t.Fatalf("JSON replacement synthesized %d auths, want 1", len(auths))
			}
			auth := auths[0]
			if _, err := f.manager.Register(coreauth.WithSkipPersist(context.Background()), auth); err != nil {
				t.Fatal(err)
			}
			registry.GetGlobalRegistry().RegisterClient(auth.ID, auth.Provider, []*registry.ModelInfo{{ID: "synthetic-model"}})
			t.Cleanup(func() { registry.GetGlobalRegistry().UnregisterClient(auth.ID) })
			payload := []byte(`{"model":"synthetic-model","messages":[{"role":"user","content":"synthetic"}]}`)
			options := coreexecutor.Options{SourceFormat: "openai"}
			if family == "gemini" {
				f.manager.RegisterExecutor(executor.NewGeminiExecutor(view))
				payload = []byte(`{"contents":[{"role":"user","parts":[{"text":"synthetic"}]}]}`)
				options.SourceFormat = "gemini"
			} else {
				f.manager.RegisterExecutor(executor.NewOpenAICompatExecutor(auth.Provider, view))
			}
			if _, err := f.manager.Execute(context.Background(), []string{auth.Provider}, coreexecutor.Request{Model: "synthetic-model", Payload: payload}, options); err != nil {
				t.Fatal(err)
			}
			if key := <-observed; key != "synthetic-replaced" || calls.Load() != 1 {
				t.Fatal("real auth manager did not send exactly the replaced synthetic credential")
			}
			rows = resourceManualArray(t, f, route)
			if family == "gemini" {
				rows[0]["keys"].([]any)[0].(map[string]any)["excluded-models"] = []string{"*"}
			} else {
				rows[0]["disabled"] = true
			}
			resourceManualJSON(t, f, http.MethodPut, route, rows, http.StatusOK)
			if err := f.handler.ReconcileResourceInventory(); err != nil {
				t.Fatal(err)
			}
			f.watcher.SetConfig(resourceSaveConfig(t, f))
			auths = f.watcher.SnapshotCoreAuths()
			if family == "gemini" {
				if len(auths) != 1 || auths[0].Attributes["excluded_models"] != "*" {
					t.Fatal("JSON manual exclusion did not reach real watcher synthesis")
				}
			} else if len(auths) != 0 {
				t.Fatal("JSON disabled provider still synthesized a runtime auth")
			}
			var caps struct{ Gate, Query bool }
			if err := json.Unmarshal(f.request("resource-capabilities", 200), &caps); err != nil || caps.Gate || caps.Query || calls.Load() != 1 {
				t.Fatal("manual operations fabricated gate/query capability or contacted upstream")
			}
		})
	}
}

func TestResourceInventoryOwnedFilteredDeletion(t *testing.T) {
	for _, family := range []string{"codex", "openai-compatibility"} {
		for _, writerKind := range []string{"same-writer", "fresh-clone"} {
			t.Run(family+"/"+writerKind, func(t *testing.T) {
				f := resourceSaveFixture(t, fmt.Sprintf("  %s:\n    - name: first\n      base-url: https://synthetic.invalid/A\n      keys: [{api-key: synthetic-A}]\n    - name: second\n      base-url: https://synthetic.invalid/B\n      keys: [{api-key: synthetic-B}]\n    - name: external-filtered\n      keys: [{api-key: synthetic-external, headers: {X-Synthetic: retained}}]\n    - name: empty\n      keys: []\n", family))
				groups := resourceSavedGroups(t, f.path, family)
				filtered := resourceSourceValue(t, groups.Content[2])
				empty := resourceSourceValue(t, groups.Content[3])
				cfg := resourceSaveConfig(t, f)
				f.handler.SetConfig(cfg)
				if err := config.SaveConfigPreserveCommentsForManagement(f.path, cfg); err != nil {
					t.Fatal(err)
				}
				if family == "codex" {
					cfg.CodexKey[0].BaseURL = ""
				} else {
					cfg.OpenAICompatibility[0].BaseURL = ""
				}
				if err := config.SaveConfigPreserveComments(f.path, cfg); err != nil {
					t.Fatalf("owned endpoint clear: %v", err)
				}
				older := cfg.CloneForRuntime()
				if writerKind == "fresh-clone" {
					cfg = cfg.CloneForRuntime()
				}
				if err := f.handler.ReconcileResourceInventory(); err != nil {
					t.Fatal(err)
				}
				f.watcher.SetConfig(resourceSaveConfig(t, f))
				for _, auth := range f.manager.List() {
					f.manager.Remove(context.Background(), auth.ID)
				}
				for _, auth := range f.watcher.SnapshotCoreAuths() {
					if _, err := f.manager.Register(coreauth.WithSkipPersist(context.Background()), auth); err != nil {
						t.Fatal(err)
					}
				}
				if auths := f.manager.List(); len(auths) != 1 || auths[0].Attributes["api_key"] != "synthetic-B" {
					t.Fatal("owned endpoint clear did not leave only B in the real auth manager")
				}
				if family == "codex" {
					cfg.CodexKey = cfg.CodexKey[1:]
				} else {
					cfg.OpenAICompatibility = cfg.OpenAICompatibility[1:]
				}
				before, err := os.ReadFile(f.path)
				if err != nil {
					t.Fatal(err)
				}
				if err = os.Chmod(f.path, 0400); err != nil {
					t.Fatal(err)
				}
				errSave := config.SaveConfigPreserveComments(f.path, cfg)
				if err = os.Chmod(f.path, 0600); err != nil {
					t.Fatal(err)
				}
				if errSave == nil {
					t.Fatal("synthetic read-only file did not reject deletion persistence")
				}
				after, err := os.ReadFile(f.path)
				if err != nil || string(after) != string(before) {
					t.Fatal("failed deletion changed source bytes")
				}
				if err = config.SaveConfigPreserveComments(f.path, cfg); err != nil {
					t.Fatalf("delete owned newly filtered A after failed write: %v", err)
				}
				groups = resourceSavedGroups(t, f.path, family)
				if len(groups.Content) != 3 || testField(testField(groups.Content[0], "keys").Content[0], "api-key").Value != "synthetic-B" || !reflect.DeepEqual(resourceSourceValue(t, groups.Content[1]), filtered) || !reflect.DeepEqual(resourceSourceValue(t, groups.Content[2]), empty) {
					t.Fatalf("successful owned deletion retained filtered A or changed external sources: groups=%d want=3 (B, external-filtered, empty)", len(groups.Content))
				}
				before, err = os.ReadFile(f.path)
				if err != nil {
					t.Fatal(err)
				}
				if err = config.SaveConfigPreserveComments(f.path, older); err == nil {
					t.Fatal("an independent older filtered view was accepted after deletion")
				}
				after, err = os.ReadFile(f.path)
				if err != nil || string(after) != string(before) {
					t.Fatal("refused older clone changed the source")
				}
				fresh := cfg.CloneForRuntime()
				f.handler.SetConfig(fresh)
				fresh.CodexKey = nil
				fresh.OpenAICompatibility = nil
				if err := config.SaveConfigPreserveCommentsForManagement(f.path, fresh); err != nil {
					t.Fatal(err)
				}
				if err = config.SaveConfigPreserveComments(f.path, fresh.CloneForRuntime()); err != nil {
					t.Fatalf("fresh empty clone lost the owned represented-row basis: %v", err)
				}
				groups = resourceSavedGroups(t, f.path, family)
				if len(groups.Content) != 2 || !reflect.DeepEqual(resourceSourceValue(t, groups.Content[0]), filtered) || !reflect.DeepEqual(resourceSourceValue(t, groups.Content[1]), empty) {
					t.Fatal("delete-all restored owned A/B or removed externally filtered/empty sources")
				}
				if err = f.handler.ReconcileResourceInventory(); err != nil {
					t.Fatal(err)
				}
				f.snapshot()
				f.watcher.SetConfig(resourceSaveConfig(t, f))
				if len(f.watcher.SnapshotCoreAuths()) != 0 {
					t.Fatal("filtered deletion or delete-all made an external source routable")
				}
			})
		}
	}
}

func TestResourceInventoryV0SaveCompatibility(t *testing.T) {
	f := resourceSaveFixture(t, "  codex:\n    - name: legacy\n      base-url: https://synthetic.invalid/v1\n      keys: [{api-key: synthetic-A}]\n")
	f.handler.SetConfig(resourceSaveConfig(t, f))
	req := httptest.NewRequest(http.MethodPut, "/v0/management/codex-api-key", strings.NewReader(`[{"api-key":"synthetic-replacement","base-url":"https://synthetic.invalid/v1","resource-id":"legacy-client-id"}]`))
	req.RemoteAddr = "127.0.0.1:1234"
	req.Header.Set("Authorization", "Bearer test-password")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	f.server.engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("deprecated v0 replacement acquired new provenance refusal: status=%d body=%s", rec.Code, rec.Body.String())
	}
	saved := resourceSaveConfig(t, f)
	if len(saved.CodexKey) != 1 || saved.CodexKey[0].APIKey != "synthetic-replacement" {
		t.Fatal("v0 replacement did not retain its historical save behavior")
	}
}
func TestResourceInventoryTypedInvalidRecovery(t *testing.T) {
	for _, layout := range []string{"grouped", "native"} {
		for _, name := range []string{"api-key", "base-url", "headers"} {
			t.Run(layout+"/"+name, func(t *testing.T) {
				f := resourceSaveFixture(t, "  codex:\n    - name: live\n      base-url: https://synthetic.invalid/v1\n      headers: {X-Synthetic: kept}\n      keys: [{api-key: synthetic-A}, {api-key: '', weight: 0}]\n    - name: filtered\n      keys: [{api-key: '', disabled: true}]\n")
				before := f.snapshot()
				valid, err := os.ReadFile(f.path)
				if err != nil {
					t.Fatal(err)
				}
				var doc yaml.Node
				if err = yaml.Unmarshal(valid, &doc); err != nil {
					t.Fatal(err)
				}
				root := doc.Content[0]
				groups := testField(testField(root, "api-keys"), "codex")
				group := groups.Content[0]
				if layout == "native" {
					// The parser also accepts native typed flat keys. Keep issued IDs in the
					// external edit so type validation, not enrollment, determines readiness.
					key := testField(group, "keys").Content[0]
					for _, field := range []string{"base-url", "headers"} {
						key.Content = append(key.Content, &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: field}, testField(group, field))
					}
					for i := 0; i+1 < len(root.Content); i += 2 {
						if root.Content[i].Value == "api-keys" {
							root.Content = append(root.Content[:i], root.Content[i+2:]...)
							break
						}
					}
					root.Content = append(root.Content, &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: "codex-api-key"}, &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq", Content: []*yaml.Node{key}})
					group = key
				}
				target := group
				if name == "api-key" && layout == "grouped" {
					target = testField(group, "keys").Content[0]
				}
				field := testField(target, name)
				*field = yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq", Content: []*yaml.Node{{Kind: yaml.ScalarNode, Tag: "!!str", Value: "synthetic-bad"}}}
				invalid, err := yaml.Marshal(&doc)
				if err != nil {
					t.Fatal(err)
				}
				if _, err = config.ParseConfigBytes(invalid); err == nil {
					t.Fatal("fixture must be rejected by native typed parser")
				}
				if err = os.WriteFile(f.path, invalid, 0600); err != nil {
					t.Fatal(err)
				}
				f.request("resources", http.StatusServiceUnavailable)
				errReconcile := f.watcher.ReconcileResourceInventory()
				if errReconcile == nil {
					t.Error("typed-invalid external source was reconciled as complete/ready")
				}
				f.request("resources", http.StatusServiceUnavailable)
				var caps struct{ Inventory, Durable, Gate, Query bool }
				if err = json.Unmarshal(f.request("resource-capabilities", http.StatusOK), &caps); err != nil {
					t.Fatal(err)
				}
				if caps.Inventory || caps.Durable || caps.Gate || caps.Query {
					t.Fatal("typed-invalid source advertised ready inventory or extended capabilities")
				}
				untouched, err := os.ReadFile(f.path)
				if err != nil || string(untouched) != string(invalid) {
					t.Fatal("invalid reconciliation or GET rewrote malformed source")
				}
				if err = os.WriteFile(f.path, valid, 0600); err != nil {
					t.Fatal(err)
				}
				if err = f.watcher.ReconcileResourceInventory(); err != nil {
					t.Fatalf("valid recovery: %v", err)
				}
				after := f.snapshot()
				if !after.Complete || len(after.Resources) != len(before.Resources) {
					t.Fatal("recovery omitted raw filtered inventory")
				}
				for _, old := range before.Resources {
					next := rowByID(t, after, old.ResourceID)
					if next.Target != old.Target || next.Manual != old.Manual || next.Readiness != "ready" || next.Routing != "unknown" || next.Capabilities.Query || next.Capabilities.Gate {
						t.Fatal("failed typed edit changed durable identity/manual state or recovery readiness")
					}
				}
			})
		}
	}
}

func TestResourceInventoryV8EnrollmentEdit(t *testing.T) {
	f := newInventoryHarness(t, t.TempDir())
	before := f.snapshot()
	// Do not reload or inject a replacement handler config after enrollment.
	// The first native GET/edit/PATCH must work after catalog-issued markers.
	resourceManualPatch(t, f, "codex", map[string]any{"weight": 0})
	if err := f.watcher.ReconcileResourceInventory(); err != nil {
		t.Fatal(err)
	}
	after := f.snapshot()
	old := before.Resources[1]
	next := rowByID(t, after, old.ResourceID)
	if !next.Manual.ZeroWeight || next.Incarnation != old.Incarnation || next.CredentialVersion != old.CredentialVersion {
		t.Fatal("first native edit lost manual ownership or identity")
	}
}
func TestResourceInventoryTypedInvalidEmptyGroupRecovery(t *testing.T) {
	for _, name := range []string{"base-url", "headers"} {
		t.Run(name, func(t *testing.T) {
			f := resourceSaveFixture(t, "  codex:\n    - name: empty\n      base-url: https://synthetic.invalid\n      headers: {X-Synthetic: kept}\n      keys: []\n")
			before := f.snapshot()
			valid, err := os.ReadFile(f.path)
			if err != nil {
				t.Fatal(err)
			}
			editInventoryConfig(t, f.path, func(root *yaml.Node) {
				group := testField(testField(root, "api-keys"), "codex").Content[0]
				*testField(group, name) = yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq", Content: []*yaml.Node{{Kind: yaml.ScalarNode, Tag: "!!str", Value: "synthetic-bad"}}}
			})
			// There is no flattened runtime row to type-check. Catalog seals must still
			// validate raw provider credential fields instead of coercing them to empty.
			if err = f.watcher.ReconcileResourceInventory(); err == nil {
				t.Fatal("raw empty provider accepted malformed credential type")
			}
			f.request("resources", http.StatusServiceUnavailable)
			if err = os.WriteFile(f.path, valid, 0600); err != nil {
				t.Fatal(err)
			}
			if err = f.watcher.ReconcileResourceInventory(); err != nil {
				t.Fatal(err)
			}
			after := f.snapshot()
			if !after.Complete || len(after.Resources) != 1 || after.Resources[0].Target != before.Resources[0].Target {
				t.Fatal("empty raw provider did not recover without identity drift")
			}
		})
	}
}
