package config

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func (c *ConfigManager) GetBool(key string) bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	if v, ok := c.combinedConfig[strings.ToLower(key)]; ok {
		val := v.Value
		switch val := val.(type) {
		case bool:
			return val
		case string:
			if strings.ToLower(val) == "true" {
				return true
			}
			return false
		case int:
			if val == 0 {
				return false
			}
			return true
		case float32, float64:
			if val == 0 {
				return false
			}
			return true
		case nil:
			return false
		case time.Duration:
			if val == 0 || val < 0 {
				return false
			}
			return true
		default:
			return val.(bool)
		}
	}
	return false
}

func (c *ConfigManager) GetDuration(key string) time.Duration {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	if v, ok := c.combinedConfig[strings.ToLower(key)]; ok {
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
		case float32:
			return time.Duration(val)
		case float64:
			return time.Duration(val)
		case nil:
			return 0
		default:
			return val.(time.Duration)
		}
	}
	return 0
}

func (c *ConfigManager) GetString(key string) string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	if v, ok := c.combinedConfig[strings.ToLower(key)]; ok {
		switch val := v.Value.(type) {
		case string:
			return val
		default:
			return fmt.Sprintf("%v", v.Value)
		}
	}
	return ""
}

func (c *ConfigManager) GetStringMap(key string) map[string]any {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	if v, ok := c.combinedConfig[strings.ToLower(key)]; ok {
		switch val := v.Value.(type) {
		case map[string]any:
			return val
		default:
			return nil
		}
	}
	return nil
}

func (c *ConfigManager) GetStringSlice(key string) []string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	if v, ok := c.combinedConfig[strings.ToLower(key)]; ok {
		switch val := v.Value.(type) {
		case []string:
			return val
		default:
			return nil
		}
	}
	return nil
}

func (c *ConfigManager) GetInt(key string) int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	if v, ok := c.combinedConfig[strings.ToLower(key)]; ok {
		switch val := v.Value.(type) {
		case int:
			return val
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
			return val.(int)
		}
	}
	return 0
}

func (c *ConfigManager) GetIntSlice(key string) []int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	if v, ok := c.combinedConfig[strings.ToLower(key)]; ok {
		switch val := v.Value.(type) {
		case []int:
			return val
		default:
			return nil
		}
	}
	return nil
}
