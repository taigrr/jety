package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

type ConfigType string

const (
	ConfigTypeTOML ConfigType = "toml"
	ConfigTypeYAML ConfigType = "yaml"
	ConfigTypeJSON ConfigType = "json"
)

type ConfigManager struct {
	configFileUsed string
	configType     ConfigType
	envPrefix      string
	mapConfig      map[string]any
	defaultConfig  map[string]any
	envConfig      map[string]string
}

var (
	ErrConfigFileNotFound = errors.New("config File Not Found")
)

func NewConfigManager(automaticEnv bool) *ConfigManager {
	cm := ConfigManager{}
	cm.envConfig = make(map[string]string)
	cm.mapConfig = make(map[string]any)
	cm.defaultConfig = make(map[string]any)
	cm.envPrefix = ""
	envSet := os.Environ()
	for _, env := range envSet {
		kv := strings.Split(env, "=")
		cm.envConfig[kv[0]] = kv[1]
		lowerKey := strings.ToLower(kv[0])
		if cm.envConfig[lowerKey] == "" {
			// if the key is not set, set it as the lower case of the key
			// but don't clobber any existing, more specific (already lowercase) value
			cm.envConfig[lowerKey] = kv[1]
		}
	}
	return &cm
}

func (c *ConfigManager) ConfigFileUsed() string {
	return c.configFileUsed
}

func (c *ConfigManager) WriteConfig() {
}

func (c *ConfigManager) GetBool(key string) bool {
}

func (c *ConfigManager) GetDuration(key string) time.Duration {
}

func (c *ConfigManager) GetString(key string) string {
}

func (c *ConfigManager) GetStringMap(key string) map[string]any {
}

func (c *ConfigManager) GetStringSlice(key string) []string {
}

func (c *ConfigManager) GetInt(key string) int {
}

func (c *ConfigManager) SetConfigType(configType string) error {
	switch configType {
	case "toml":
		c.configType = ConfigTypeTOML
	case "yaml":
		c.configType = ConfigTypeYAML
	case "json":
		c.configType = ConfigTypeJSON
	default:
		return fmt.Errorf("config type %s not supported", configType)
	}
	return nil
}

func (c *ConfigManager) SetDefault(key string, value any) {
}

func (c *ConfigManager) SetEnvPrefix(prefix string) {
}

func (c *ConfigManager) Set(key string, value any) {
}

func (c *ConfigManager) GetIntSlice(key string) []int {
}

func (c *ConfigManager) ReadInConfig() error {
	return nil
}

func (c *ConfigManager) SetConfigName(name string) {
}

func (c *ConfigManager) SetConfigFile(file string) {
	c.configFileUsed = file
}
