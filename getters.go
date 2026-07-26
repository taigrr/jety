package jety

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func envKeyForPath(key string) string {
	return strings.ReplaceAll(strings.ToLower(key), ".", "_")
}

func cloneMap(input map[string]any) map[string]any {
	cloned := make(map[string]any, len(input))
	for key, value := range input {
		switch typed := value.(type) {
		case map[string]any:
			cloned[key] = cloneMap(typed)
		default:
			cloned[key] = typed
		}
	}
	return cloned
}

func findMapKeyFold(input map[string]any, needle string) (string, bool) {
	for key := range input {
		if strings.EqualFold(key, needle) {
			return key, true
		}
	}
	return "", false
}

func applyScopedEnvOverrides(target map[string]any, prefix string, envConfig map[string]ConfigMap) {
	basePrefix := envKeyForPath(prefix)
	for envKey, entry := range envConfig {
		if basePrefix != "" {
			var ok bool
			envKey, ok = strings.CutPrefix(envKey, basePrefix+"_")
			if !ok {
				continue
			}
		}
		parts := strings.Split(envKey, "_")
		current := target
		if len(parts) == 1 {
			leafKey, ok := findMapKeyFold(current, parts[0])
			if !ok {
				continue
			}
			current[leafKey] = entry.Value
			continue
		}
		for _, part := range parts[:len(parts)-1] {
			matchedKey, ok := findMapKeyFold(current, part)
			if !ok {
				current = nil
				break
			}
			next, ok := current[matchedKey].(map[string]any)
			if !ok {
				current = nil
				break
			}
			current = next
		}
		if current == nil {
			continue
		}
		leafKey, ok := findMapKeyFold(current, parts[len(parts)-1])
		if !ok {
			continue
		}
		current[leafKey] = entry.Value
	}
}

// resolve looks up a key in combinedConfig, falling back to envConfig.
// It supports dot notation (e.g., "services.mas.server") to traverse nested maps.
func (c *ConfigManager) resolve(key string) (ConfigMap, bool) {
	lower := strings.ToLower(key)

	// First, try direct lookup (for top-level keys or keys without dots)
	if v, ok := c.combinedConfig[lower]; ok {
		return v, true
	}
	if v, ok := c.envConfig[lower]; ok {
		return v, true
	}

	// If key contains dots, allow underscore-separated env vars to override
	// nested config paths (e.g. SERVICES_API_PORT -> services.api.port).
	if strings.Contains(lower, ".") {
		if v, ok := c.envConfig[envKeyForPath(lower)]; ok {
			return v, true
		}
		if v, ok := c.resolveNested(lower, c.combinedConfig); ok {
			return v, true
		}
		if v, ok := c.resolveNested(lower, c.envConfig); ok {
			return v, true
		}
	}

	return ConfigMap{}, false
}

// resolveNested traverses nested maps using dot-separated key paths.
func (c *ConfigManager) resolveNested(key string, config map[string]ConfigMap) (ConfigMap, bool) {
	parts := strings.Split(key, ".")
	if len(parts) < 2 {
		return ConfigMap{}, false
	}

	// Look up the first part in the config
	firstPart := parts[0]
	entry, ok := config[firstPart]
	if !ok {
		return ConfigMap{}, false
	}

	// Traverse the remaining parts through nested maps
	current := entry.Value
	for _, part := range parts[1:] {
		m, ok := current.(map[string]any)
		if !ok {
			return ConfigMap{}, false
		}

		// Try case-insensitive lookup in the nested map
		matchedKey, ok := findMapKeyFold(m, part)
		if !ok {
			return ConfigMap{}, false
		}
		current = m[matchedKey]
	}

	return ConfigMap{Key: key, Value: current}, true
}

func (c *ConfigManager) Get(key string) any {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	v, ok := c.resolve(key)
	if !ok {
		return nil
	}
	return v.Value
}

func (c *ConfigManager) GetBool(key string) bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	v, ok := c.resolve(key)
	if !ok {
		return false
	}
	val := v.Value
	switch val := val.(type) {
	case bool:
		return val
	case string:
		return strings.EqualFold(val, "true")
	case int:
		return val != 0
	case float32:
		return val != 0
	case float64:
		return val != 0
	case time.Duration:
		return val > 0
	case nil:
		return false
	default:
		return false
	}
}

func (c *ConfigManager) GetDuration(key string) time.Duration {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	v, ok := c.resolve(key)
	if !ok {
		return 0
	}
	val := v.Value
	switch val := val.(type) {
	case time.Duration:
		return val
	case string:
		d, err := time.ParseDuration(val)
		if err != nil {
			return 0
		}
		return d
	case int:
		return time.Duration(val)
	case int64:
		return time.Duration(val)
	case float32:
		return time.Duration(val)
	case float64:
		return time.Duration(val)
	case nil:
		return 0
	default:
		return 0
	}
}

func (c *ConfigManager) GetString(key string) string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	v, ok := c.resolve(key)
	if !ok {
		return ""
	}

	switch val := v.Value.(type) {
	case string:
		return val
	default:
		return fmt.Sprintf("%v", v.Value)
	}
}

func (c *ConfigManager) GetStringMap(key string) map[string]any {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	v, ok := c.resolve(key)
	if !ok {
		return nil
	}
	switch val := v.Value.(type) {
	case map[string]any:
		cloned := cloneMap(val)
		applyScopedEnvOverrides(cloned, key, c.envConfig)
		return cloned
	default:
		return nil
	}
}

func (c *ConfigManager) GetStringSlice(key string) []string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	v, ok := c.resolve(key)
	if !ok {
		return nil
	}
	switch val := v.Value.(type) {
	case []string:
		return val
	case []any:
		var ret []string
		for _, v := range val {
			switch v := v.(type) {
			case string:
				ret = append(ret, v)
			default:
				ret = append(ret, fmt.Sprintf("%v", v))
			}
		}
		return ret
	default:
		return nil
	}
}

func (c *ConfigManager) GetFloat64(key string) float64 {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	v, ok := c.resolve(key)
	if !ok {
		return 0
	}
	switch val := v.Value.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case string:
		f, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return 0
		}
		return f
	case nil:
		return 0
	default:
		return 0
	}
}

func (c *ConfigManager) GetInt64(key string) int64 {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	v, ok := c.resolve(key)
	if !ok {
		return 0
	}
	switch val := v.Value.(type) {
	case int64:
		return val
	case int:
		return int64(val)
	case string:
		i, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return 0
		}
		return i
	case float32:
		return int64(val)
	case float64:
		return int64(val)
	case nil:
		return 0
	default:
		return 0
	}
}

func (c *ConfigManager) GetInt(key string) int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	v, ok := c.resolve(key)
	if !ok {
		return 0
	}
	switch val := v.Value.(type) {
	case int:
		return val
	case int64:
		return int(val)
	case string:
		i, err := strconv.Atoi(val)
		if err != nil {
			return 0
		}
		return i
	case float32:
		return int(val)
	case float64:
		return int(val)
	case nil:
		return 0
	default:
		return 0
	}
}

func (c *ConfigManager) GetIntSlice(key string) []int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	v, ok := c.resolve(key)
	if !ok {
		return nil
	}
	switch val := v.Value.(type) {
	case []int:
		return val
	case []any:
		var ret []int
		for _, v := range val {
			switch v := v.(type) {
			case int:
				ret = append(ret, v)
			case int64:
				ret = append(ret, int(v))
			case string:
				i, err := strconv.Atoi(v)
				if err != nil {
					continue
				}
				ret = append(ret, i)
			case float32:
				ret = append(ret, int(v))
			case float64:
				ret = append(ret, int(v))
			case nil:
				continue
			default:
				continue
			}
		}
		return ret
	default:
		return nil
	}
}
