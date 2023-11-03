package config

import (
	"strings"
)

func (c *ConfigManager) SetBool(key string, value bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	lower := strings.ToLower(key)
	c.mapConfig[lower] = ConfigMap{Key: key, Value: value}
	c.combinedConfig[lower] = ConfigMap{Key: key, Value: value}
}

func (c *ConfigManager) SetString(key string, value string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	lower := strings.ToLower(key)
	c.mapConfig[lower] = ConfigMap{Key: key, Value: value}
	c.combinedConfig[lower] = ConfigMap{Key: key, Value: value}
}

func (c *ConfigManager) Set(key string, value any) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	lower := strings.ToLower(key)
	c.mapConfig[lower] = ConfigMap{Key: key, Value: value}
	c.combinedConfig[lower] = ConfigMap{Key: key, Value: value}
}

func (c *ConfigManager) SetDefault(key string, value any) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	lower := strings.ToLower(key)
	c.defaultConfig[lower] = ConfigMap{Key: key, Value: value}
	if _, ok := c.mapConfig[lower]; !ok {
		if envVal, ok := c.envConfig[lower]; !ok {
			c.mapConfig[lower] = envVal
			c.combinedConfig[lower] = envVal
		} else {
			c.combinedConfig[lower] = ConfigMap{Key: key, Value: value}
		}
	}
}
