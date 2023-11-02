package config

import "errors"

type ConfigManager struct{}

var ConfigFileNotFoundError = errors.New("config File Not Found")

func NewConfigManager() *ConfigManager {
	return &ConfigManager{}
}

func AutoMaticEnv() {
}

func ConfigFileUsed() {
}

func GetDuration() {
}

func GetString() {
}

func GetStringMap() {
}

func GetStringSlice() {
}

func ReadInConfig() {
}

func SetConfigFile() {
}

func SetConfigName() {
}

func SetConfigType() {
}

func SetDefault() {
}

func Set() {
}

func WriteConfig() {
}
