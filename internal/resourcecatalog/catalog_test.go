package resourcecatalog_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/router-for-me/CLIProxyAPI/v8/internal/resourcecatalog"
)

func fixture(t *testing.T) (*resourcecatalog.Catalog, string, string) {
	t.Helper()
	dir := t.TempDir()
	authDir := filepath.Join(dir, "auth")
	if err := os.Mkdir(authDir, 0700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "config.yaml")
	text := fmt.Sprintf("server: {port: 0}\noauth: {auth-dir: %q}\nresource-catalog: {enabled: true}\napi-keys:\n  codex: [{name: synthetic, base-url: 'https://synthetic.invalid', keys: [{api-key: synthetic-A}]}]\n", filepath.ToSlash(authDir))
	if err := os.WriteFile(path, []byte(text), 0600); err != nil {
		t.Fatal(err)
	}
	c, err := resourcecatalog.Open(path, authDir, "")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := c.Close(); err != nil {
			t.Error(err)
		}
	})
	return c, path, authDir
}
func TestResourceIdentityCASAndObservationFences(t *testing.T) {
	c, path, _ := fixture(t)
	before, err := c.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(path, []byte(strings.Replace(string(data), "synthetic-A", "synthetic-B", 1)), 0600); err != nil {
		t.Fatal(err)
	}
	results := make(chan error, 2)
	for range 2 {
		go func() { results <- c.Reconcile(before.Version) }()
	}
	ok, conflict := 0, 0
	for range 2 {
		err := <-results
		if err == nil {
			ok++
		} else if errors.Is(err, resourcecatalog.ErrConflict) {
			conflict++
		} else {
			t.Fatal(err)
		}
	}
	if ok != 1 || conflict != 1 {
		t.Fatalf("CAS results: success=%d conflict=%d", ok, conflict)
	}
	after, err := c.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if err = c.Check(after.Version, after.Resources[1].Target); err != nil {
		t.Fatal(err)
	}
	if !errors.Is(c.Check(before.Version, after.Resources[1].Target), resourcecatalog.ErrConflict) {
		t.Fatal("late snapshot was accepted")
	}
	if !errors.Is(c.Check(after.Version, before.Resources[1].Target), resourcecatalog.ErrConflict) {
		t.Fatal("stale credential target was accepted")
	}
	other, _, _ := fixture(t)
	scope := other.Version()
	if !errors.Is(c.Reconcile(scope), resourcecatalog.ErrConflict) {
		t.Fatal("cross-instance CAS was accepted")
	}
}
func TestResourceIdentityExplicitRebind(t *testing.T) {
	c, path, authDir := fixture(t)
	before, err := c.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if err = c.Rebind(before.Version); err != nil {
		t.Fatal(err)
	}
	after, err := c.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if before.InstanceID != after.InstanceID || before.InstanceIncarnation == after.InstanceIncarnation || before.BootEpoch != after.BootEpoch || after.Revision <= before.Revision {
		t.Fatal("rebind confused persistent identity and boot epoch")
	}
	if !errors.Is(c.Check(before.Version, before.Resources[1].Target), resourcecatalog.ErrConflict) {
		t.Fatal("pre-rebind observation was accepted")
	}
	if err = c.Close(); err != nil {
		t.Fatal(err)
	}
	next, err := resourcecatalog.Open(path, authDir, "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := next.Close(); err != nil {
			t.Error(err)
		}
	}()
	restarted, err := next.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if restarted.InstanceIncarnation != after.InstanceIncarnation || restarted.BootEpoch == after.BootEpoch {
		t.Fatal("rebound instance was not durable")
	}
	if !errors.Is(next.Reconcile(after.Version), resourcecatalog.ErrConflict) {
		t.Fatal("cross-boot writer was accepted")
	}
}
func TestResourceInventoryWriterProcess(t *testing.T) {
	if path := os.Getenv("CPA_SPEC66_LOCK_TEST_CONFIG"); path != "" {
		c, err := resourcecatalog.Open(path, os.Getenv("CPA_SPEC66_LOCK_TEST_AUTH"), "")
		if err == nil {
			_ = c.Close()
			t.Fatal("another process acquired the writer lock")
		}
		return
	}
	_, path, authDir := fixture(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestResourceInventoryWriterProcess$")
	cmd.Env = append(os.Environ(), "CPA_SPEC66_LOCK_TEST_CONFIG="+path, "CPA_SPEC66_LOCK_TEST_AUTH="+authDir)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("writer process: %v %s", err, output)
	}
}
