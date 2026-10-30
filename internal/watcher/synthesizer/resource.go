package synthesizer

import coreauth "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/auth"

func applyResourceBinding(id, parent string, attrs map[string]string) {
	if id != "" {
		attrs[coreauth.AttributeResourceID] = id
	}
	if parent != "" {
		attrs[coreauth.AttributeProviderResourceID] = parent
	}
}
