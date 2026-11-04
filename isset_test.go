package jety

import "testing"

func TestIsSetResolvesNestedPaths(t *testing.T) {
	t.Setenv("JETY_ISSET_SERVICES_API_ENVONLY", "present")
	manager := NewConfigManager().WithEnvPrefix("JETY_ISSET_")
	manager.SetDefault("services", map[string]any{
		"API": map[string]any{
			"Port":    8080,
			"Enabled": false,
			"Empty":   "",
			"Null":    nil,
		},
	})
	manager.Set("services.api.override", "explicit")

	for _, key := range []string{
		"services", "services.api", "SERVICES.API.PORT",
		"services.api.enabled", "services.api.empty", "services.api.null",
		"services.api.envonly", "services.api.override",
	} {
		t.Run(key, func(t *testing.T) {
			if !manager.IsSet(key) {
				t.Errorf("IsSet(%q) = false, want true", key)
			}
		})
	}
	for _, key := range []string{"missing", "services.missing", "services.api.port.child"} {
		if manager.IsSet(key) {
			t.Errorf("IsSet(%q) = true, want false", key)
		}
	}
}
