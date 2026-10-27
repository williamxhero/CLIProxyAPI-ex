package resourcecatalog_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/router-for-me/CLIProxyAPI/v8/internal/config"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/resourcecatalog"
)

func TestResourceInventoryRejectsInvalidManualScalars(t *testing.T) {
	for _, values := range []string{
		"disabled: definitely-not-a-bool, keys: [{api-key: synthetic}]",
		"disabled: [], keys: [{api-key: synthetic}]",
		"keys: [{api-key: synthetic, weight: 0.0}]",
		"keys: [{api-key: synthetic, weight: '0'}]",
		"keys: [{api-key: synthetic, weight: true}]",
	} {
		t.Run(values, func(t *testing.T) {
			dir := t.TempDir()
			authDir := filepath.Join(dir, "auth")
			if err := os.Mkdir(authDir, 0700); err != nil {
				t.Fatal(err)
			}
			text := fmt.Sprintf("oauth: {auth-dir: %q}\nresource-catalog: {enabled: true}\napi-keys:\n  openai-compatibility: [{name: synthetic, base-url: https://synthetic.invalid, %s}]\n", filepath.ToSlash(authDir), values)
			path := filepath.Join(dir, "config.yaml")
			if err := os.WriteFile(path, []byte(text), 0600); err != nil {
				t.Fatal(err)
			}
			if _, err := config.ParseConfigBytes([]byte(text)); err == nil {
				t.Fatal("fixture is not invalid under the runtime parser")
			}
			c, err := resourcecatalog.Open(path, authDir, "")
			if err == nil {
				if errClose := c.Close(); errClose != nil {
					t.Error(errClose)
				}
				t.Fatal("invalid runtime manual scalar was silently coerced into inventory")
			}
		})
	}
}
