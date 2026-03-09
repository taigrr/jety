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
