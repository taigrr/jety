package jety

import "time"

var defaultConfigManager = NewConfigManager()

func GetIntSlice(key string) []int {
	return defaultConfigManager.GetIntSlice(key)
}

func ReadInConfig() error {
	return defaultConfigManager.ReadInConfig()
}

func SetConfigFile(file string) {
	defaultConfigManager.SetConfigFile(file)
}

func SetConfigName(name string) {
	defaultConfigManager.SetConfigName(name)
}

func GetFloat64(key string) float64 {
	return defaultConfigManager.GetFloat64(key)
}

func GetInt64(key string) int64 {
	return defaultConfigManager.GetInt64(key)
}

func GetInt(key string) int {
	return defaultConfigManager.GetInt(key)
}

func SetEnvPrefix(prefix string) {
	defaultConfigManager.SetEnvPrefix(prefix)
}

func SetConfigType(configType string) error {
	return defaultConfigManager.SetConfigType(configType)
}

func SetDefault(key string, value any) {
	defaultConfigManager.SetDefault(key, value)
}

func Set(key string, value any) {
	defaultConfigManager.Set(key, value)
}

func WriteConfig() error {
	return defaultConfigManager.WriteConfig()
}

func ConfigFileUsed() string {
	return defaultConfigManager.ConfigFileUsed()
}

func GetBool(key string) bool {
	return defaultConfigManager.GetBool(key)
}

func GetDuration(key string) time.Duration {
	return defaultConfigManager.GetDuration(key)
}

func GetString(key string) string {
	return defaultConfigManager.GetString(key)
}

func GetStringMap(key string) map[string]any {
	return defaultConfigManager.GetStringMap(key)
}

func GetStringSlice(key string) []string {
	return defaultConfigManager.GetStringSlice(key)
}

func Get(key string) any {
	return defaultConfigManager.Get(key)
}

func SetBool(key string, value bool) {
	defaultConfigManager.SetBool(key, value)
}

func SetString(key string, value string) {
	defaultConfigManager.SetString(key, value)
}

func Delete(key string) {
	defaultConfigManager.Delete(key)
}

func Sub(key string) *ConfigManager {
	return defaultConfigManager.Sub(key)
}

func SetConfigDir(path string) {
	defaultConfigManager.SetConfigDir(path)
}

func WithEnvPrefix(prefix string) *ConfigManager {
	return defaultConfigManager.WithEnvPrefix(prefix)
}

func IsSet(key string) bool {
	return defaultConfigManager.IsSet(key)
}

func AllKeys() []string {
	return defaultConfigManager.AllKeys()
}

func AllSettings() map[string]any {
	return defaultConfigManager.AllSettings()
}
