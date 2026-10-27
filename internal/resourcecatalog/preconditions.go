package resourcecatalog

import (
	"github.com/google/uuid"
	"github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/resource"
)

// Check validates observation preconditions only. Success is NOT an enforcement ACK.
func (c *Catalog) Check(version resource.Version, target resource.Target) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.currentLocked() {
		return ErrUnavailable
	}
	if version != c.versionLocked() {
		return ErrConflict
	}
	record := c.state.Records[target.ResourceID]
	if record == nil || record.Deleted || record.Blocked || record.Target != target {
		return ErrConflict
	}
	for _, row := range c.snapshot.Resources {
		if row.ResourceID == target.ResourceID && row.Readiness == "ready" && row.Identity == "valid" {
			return nil
		}
	}
	return ErrConflict
}

// Rebind is an explicit owner migration primitive for cloned/restored state.
// No HTTP route or automatic startup path calls it; it never changes routing.
func (c *Catalog) Rebind(expected resource.Version) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if expected != c.versionLocked() {
		return ErrConflict
	}
	if !c.currentLocked() {
		return ErrUnavailable
	}
	next := *c.state
	if next.Revision == ^uint64(0) {
		return ErrUnavailable
	}
	next.Revision++
	next.InstanceIncarnation = uuid.NewString()
	values, err := sourceContentValues(c.configPath, c.authDir, c.key)
	if err != nil {
		return err
	}
	if err = c.publish(&next, nil, nil, values); err != nil {
		c.failed = true
		return err
	}
	c.state = &next
	c.snapshot.Version = c.versionLocked()
	return nil
}
