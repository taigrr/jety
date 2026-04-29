package jety

import (
	"encoding/json"
	"fmt"
)

// Unmarshal unmarshals the full combined configuration into the provided
// struct pointer. It works by marshaling the settings map to JSON and then
// unmarshaling into the target, so struct field tags should use `json:"key"`.
// The target must be a non-nil pointer to a struct.
func (c *ConfigManager) Unmarshal(target any) error {
	settings := c.AllSettings()
	return mapToStruct(settings, target)
}

// UnmarshalKey unmarshals a specific key's value into the provided struct
// pointer. The key must refer to a map value (e.g., a nested TOML/YAML/JSON
// section). Supports dot notation for nested keys.
func (c *ConfigManager) UnmarshalKey(key string, target any) error {
	if settings := c.GetStringMap(key); settings != nil {
		return mapToStruct(settings, target)
	}

	c.mutex.RLock()
	v, ok := c.resolve(key)
	c.mutex.RUnlock()
	if !ok {
		return fmt.Errorf("key %q not found", key)
	}

	// If the value is not a map, try direct JSON round-trip.
	data, err := json.Marshal(v.Value)
	if err != nil {
		return fmt.Errorf("cannot marshal value for key %q: %w", key, err)
	}
	return json.Unmarshal(data, target)
}

// mapToStruct converts a map[string]any to a struct via JSON round-trip.
func mapToStruct(m map[string]any, target any) error {
	data, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("cannot marshal config: %w", err)
	}
	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("cannot unmarshal config: %w", err)
	}
	return nil
}
