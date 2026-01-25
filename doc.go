// Package jety provides configuration management supporting JSON, ENV, TOML, and YAML formats.
//
// It offers viper-like AutomaticEnv functionality with minimal dependencies, allowing
// configuration to be loaded from files and environment variables with automatic merging.
//
// Configuration sources are layered with the following precedence (highest to lowest):
//   - Values set via Set() or SetString()/SetBool()
//   - Environment variables (optionally filtered by prefix)
//   - Values from config file via ReadInConfig()
//   - Default values set via SetDefault()
//
// Basic usage:
//
//	jety.SetConfigFile("/etc/myapp/config.yaml")
//	jety.SetConfigType("yaml")
//	jety.SetEnvPrefix("MYAPP_")
//	jety.SetDefault("port", 8080)
//
//	if err := jety.ReadInConfig(); err != nil {
//	    log.Fatal(err)
//	}
//
//	port := jety.GetInt("port")
//
// For multiple independent configurations, create separate ConfigManager instances:
//
//	cm := jety.NewConfigManager()
//	cm.SetConfigFile("/etc/myapp/config.toml")
//	cm.SetConfigType("toml")
//	cm.ReadInConfig()
package jety
