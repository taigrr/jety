package config

import (
	"errors"
	"fmt"
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
	autoEnvEnabled bool
	envPrefix      string
	mapConfig      map[string]any
	defaultConfig  map[string]any
}

var (
	ErrConfigFileNotFound = errors.New("config File Not Found")
	defaultConfigManager  = NewConfigManager()
)

func NewConfigManager() *ConfigManager {
	return &ConfigManager{}
}

func AutoMaticEnv(enabled bool) {
	defaultConfigManager.autoEnvEnabled = enabled
}

func (c *ConfigManager) ConfigFileUsed() string {
	return c.configFileUsed
}

func ConfigFileUsed() string {
	return defaultConfigManager.ConfigFileUsed()
}

func (c *ConfigManager) GetBool(key string) bool {
}

func GetBool(key string) bool {
	return defaultConfigManager.GetBool(key)
}

func (c *ConfigManager) GetDuration(key string) time.Duration {
}

func GetDuration(key string) time.Duration {
	return defaultConfigManager.GetDuration(key)
}

func (c *ConfigManager) GetString(key string) string {
}

func GetString(key string) string {
	return defaultConfigManager.GetString(key)
}

func (c *ConfigManager) GetStringMap(key string) map[string]any {
}

func GetStringMap(key string) map[string]any {
	return defaultConfigManager.GetStringMap(key)
}

func (c *ConfigManager) GetStringSlice(key string) []string {
}

func GetStringSlice(key string) []string {
	return defaultConfigManager.GetStringSlice(key)
}

func (c *ConfigManager) GetInt(key string) int {
}

func GetInt(key string) int {
	return defaultConfigManager.GetInt(key)
}

func (c *ConfigManager) GetIntSlice(key string) []int {
}

func GetIntSlice(key string) []int {
	return defaultConfigManager.GetIntSlice(key)
}

func (c *ConfigManager) ReadInConfig() error {
	return nil
}

func ReadInConfig() error {
	return defaultConfigManager.ReadInConfig()
}

func (c *ConfigManager) SetConfigFile(file string) {
	c.configFileUsed = file
}

func SetConfigFile(file string) {
	defaultConfigManager.SetConfigFile(file)
}

func (c *ConfigManager) SetConfigName(name string) {
}

func SetConfigName(name string) {
	defaultConfigManager.SetConfigName(name)
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

func SetConfigType(configType string) error {
	return defaultConfigManager.SetConfigType(configType)
}

func (c *ConfigManager) SetDefault(key string, value any) {
}

func SetDefault(key string, value any) {
	defaultConfigManager.SetDefault(key, value)
}

func (c *ConfigManager) SetEnvPrefix(prefix string) {
}

func SetEnvPrefix(prefix string) {
	defaultConfigManager.SetEnvPrefix(prefix)
}

func (c *ConfigManager) Set(key string, value any) {
}

func Set(key string, value any) {
	defaultConfigManager.Set(key, value)
}

func (c *ConfigManager) WriteConfig() {
}

func WriteConfig() {
	defaultConfigManager.WriteConfig()
}
