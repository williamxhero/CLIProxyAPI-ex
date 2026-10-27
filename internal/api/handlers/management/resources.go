package management

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/config"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/resourcecatalog"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/util"
	sdkAuth "github.com/router-for-me/CLIProxyAPI/v8/sdk/auth"
	"github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/resource"
)

func (h *Handler) initializeResourceCatalog() {
	h.resourceCatalogReason = h.resourceModeReason()
	if h.resourceCatalogReason != "" {
		return
	}
	h.resourceCatalogReason = "not_ready"
	catalog, err := resourcecatalog.Open(h.configFilePath, h.cfg.AuthDir, h.cfg.ResourceCatalog.StateDir)
	if err != nil {
		if errors.Is(err, resourcecatalog.ErrUnsupported) {
			h.resourceCatalogReason = "unsupported_filesystem"
		}
		return
	}
	h.resourceCatalog = catalog
	// Opening an opted-in catalog can enroll markers in the source. Bind only
	// that writer receipt after migration, without replacing runtime settings.
	if data, errRead := os.ReadFile(h.configFilePath); errRead == nil {
		if loaded, errParse := config.ParseConfigBytes(data); errParse == nil {
			if _, errSnapshot := catalog.Snapshot(); errSnapshot == nil {
				h.cfg.ResourceCatalog = loaded.ResourceCatalog
			}
		}
	}
	h.resourceCatalogReason = "available"
}

// CloseResourceCatalog releases instance-local catalog resources at shutdown.
func (h *Handler) CloseResourceCatalog() error {
	if h.resourceCatalog == nil {
		return nil
	}
	return h.resourceCatalog.Close()
}

// ResourceCapabilities reports only capabilities established by this instance.
func (h *Handler) ResourceCapabilities(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	reason := h.resourceModeReason()
	caps := resource.Capabilities{Version: 1, PolicyLocked: true, Reason: h.resourceCatalogReason}
	if h.resourceCatalog != nil && reason == "" {
		snapshot, err := h.resourceCatalog.Snapshot()
		caps.Inventory = err == nil
		caps.Durable = err == nil
		if err == nil {
			scope := snapshot.Version
			caps.Scope = &scope
		}
		if err != nil {
			caps.Reason = "not_ready"
		}
	}
	if reason != "" {
		caps.Reason = reason
	}
	c.JSON(http.StatusOK, caps)
}

// Resources never migrates sources or contacts providers as a side effect of a read.
func (h *Handler) Resources(c *gin.Context) {
	reason := h.resourceModeReason()
	c.Header("Cache-Control", "no-store")
	if h.resourceCatalog != nil && reason == "" {
		if snapshot, err := h.resourceCatalog.Snapshot(); err == nil {
			c.JSON(http.StatusOK, snapshot)
			return
		}
	}
	c.JSON(http.StatusServiceUnavailable, gin.H{"error": "resource_inventory_unavailable"})
}

// ReconcileResourceInventory is a local source-write hook, not a management action API.
func (h *Handler) ReconcileResourceInventory() error {
	if h.resourceCatalog != nil && h.resourceModeReason() != "" {
		return resourcecatalog.ErrUnavailable
	}
	if h.resourceCatalog == nil {
		return nil
	}
	return h.resourceCatalog.Reconcile(h.resourceCatalog.Version())
}

// resourceModeReason takes no upstream action and does not infer manual ownership.
func (h *Handler) resourceModeReason() string {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.cfg == nil || !h.cfg.ResourceCatalog.Enabled {
		return "not_enabled"
	}
	if h.cfg.Home.Enabled || h.cfg.Plugins.Enabled {
		return "unsupported_mode"
	}
	if h.authManager == nil {
		return "unsupported_storage"
	}
	store, ok := h.authManager.Storage().(*sdkAuth.FileTokenStore)
	if !ok || store == nil {
		return "unsupported_storage"
	}
	actual, err := util.ResolveAuthDir(store.BaseDir())
	if err != nil || actual == "" {
		return "unsupported_storage"
	}
	expected, err := util.ResolveAuthDir(h.cfg.AuthDir)
	if err != nil || expected == "" {
		return "unsupported_storage"
	}
	actual, err = filepath.EvalSymlinks(actual)
	if err != nil {
		return "unsupported_storage"
	}
	expected, err = filepath.EvalSymlinks(expected)
	if err != nil || actual != expected {
		return "unsupported_storage"
	}
	return ""
}
