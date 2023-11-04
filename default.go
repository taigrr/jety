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

func WriteConfig() {
	defaultConfigManager.WriteConfig()
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
