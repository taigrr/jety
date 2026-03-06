package jety

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// resolve looks up a key in combinedConfig, falling back to envConfig.
func (c *ConfigManager) resolve(key string) (ConfigMap, bool) {
	lower := strings.ToLower(key)
	if v, ok := c.combinedConfig[lower]; ok {
		return v, true
	}
	if v, ok := c.envConfig[lower]; ok {
		return v, true
	}
	return ConfigMap{}, false
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
		return val
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
