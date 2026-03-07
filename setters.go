package jety

import (
	"strings"
)

func (c *ConfigManager) SetBool(key string, value bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	lower := strings.ToLower(key)
	c.overrideConfig[lower] = ConfigMap{Key: key, Value: value}
	c.combinedConfig[lower] = ConfigMap{Key: key, Value: value}
}

func (c *ConfigManager) SetString(key string, value string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	lower := strings.ToLower(key)
	c.overrideConfig[lower] = ConfigMap{Key: key, Value: value}
	c.combinedConfig[lower] = ConfigMap{Key: key, Value: value}
}

func (c *ConfigManager) Set(key string, value any) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	lower := strings.ToLower(key)
	c.overrideConfig[lower] = ConfigMap{Key: key, Value: value}
	c.combinedConfig[lower] = ConfigMap{Key: key, Value: value}
}

// Delete removes a key from all configuration layers (overrides, file,
// defaults) and rebuilds the combined configuration. Environment variables
// are not affected since they are loaded from the process environment.
func (c *ConfigManager) Delete(key string) {
	c.mutex.Lock()
	lower := strings.ToLower(key)
	delete(c.overrideConfig, lower)
	delete(c.fileConfig, lower)
	delete(c.defaultConfig, lower)
	delete(c.combinedConfig, lower)
	c.mutex.Unlock()
}

func (c *ConfigManager) SetDefault(key string, value any) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	lower := strings.ToLower(key)
	c.defaultConfig[lower] = ConfigMap{Key: key, Value: value}
	// Update combinedConfig respecting precedence: override > env > file > default
	if v, ok := c.overrideConfig[lower]; ok {
		c.combinedConfig[lower] = v
	} else if v, ok := c.envConfig[lower]; ok {
		c.combinedConfig[lower] = ConfigMap{Key: key, Value: v.Value}
	} else if v, ok := c.fileConfig[lower]; ok {
		c.combinedConfig[lower] = v
	} else {
		c.combinedConfig[lower] = ConfigMap{Key: key, Value: value}
	}
}
