package jety

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v3"
)

const (
	ConfigTypeTOML configType = "toml"
	ConfigTypeYAML configType = "yaml"
	ConfigTypeJSON configType = "json"
)

type (
	configType string

	// ConfigMap holds a configuration entry with its original key name and value.
	ConfigMap struct {
		Key   string
		Value any
	}

	// ConfigManager manages layered configuration from defaults, files,
	// environment variables, and programmatic overrides.
	ConfigManager struct {
		configName     string
		configPath     string
		configFileUsed string
		configType     configType
		overrideConfig map[string]ConfigMap
		fileConfig     map[string]ConfigMap
		defaultConfig  map[string]ConfigMap
		envConfig      map[string]ConfigMap
		combinedConfig map[string]ConfigMap
		mutex          sync.RWMutex
	}
)

var (
	ErrConfigFileNotFound = errors.New("config file not found")
	ErrConfigFileEmpty    = errors.New("config file is empty")
)

// parseEnv reads environment variables, optionally filtering by prefix,
// and returns a map keyed by lowercased (and prefix-stripped) variable names.
func parseEnv(prefix string) map[string]ConfigMap {
	result := make(map[string]ConfigMap)
	for _, env := range os.Environ() {
		key, value, found := strings.Cut(env, "=")
		if !found {
			continue
		}
		if prefix != "" {
			stripped, ok := strings.CutPrefix(key, prefix)
			if !ok {
				continue
			}
			key = stripped
		}
		result[strings.ToLower(key)] = ConfigMap{Key: key, Value: value}
	}
	return result
}

// NewConfigManager creates a new ConfigManager with all environment
// variables loaded. Use [ConfigManager.WithEnvPrefix] or
// [ConfigManager.SetEnvPrefix] to filter by prefix.
func NewConfigManager() *ConfigManager {
	return &ConfigManager{
		envConfig:      parseEnv(""),
		overrideConfig: make(map[string]ConfigMap),
		fileConfig:     make(map[string]ConfigMap),
		defaultConfig:  make(map[string]ConfigMap),
		combinedConfig: make(map[string]ConfigMap),
	}
}

// WithEnvPrefix filters environment variables to only those starting with
// the given prefix, stripping the prefix from key names. Returns the
// ConfigManager for chaining.
func (c *ConfigManager) WithEnvPrefix(prefix string) *ConfigManager {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.envConfig = parseEnv(prefix)
	return c
}

func (c *ConfigManager) ConfigFileUsed() string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.configFileUsed
}

// IsSet checks whether a key has been set in any configuration source.
func (c *ConfigManager) IsSet(key string) bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	lower := strings.ToLower(key)
	if _, ok := c.combinedConfig[lower]; ok {
		return true
	}
	_, ok := c.envConfig[lower]
	return ok
}

// AllKeys returns all keys from all configuration sources, deduplicated.
func (c *ConfigManager) AllKeys() []string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	seen := make(map[string]struct{})
	var keys []string
	for k := range c.combinedConfig {
		if _, ok := seen[k]; !ok {
			seen[k] = struct{}{}
			keys = append(keys, k)
		}
	}
	for k := range c.envConfig {
		if _, ok := seen[k]; !ok {
			seen[k] = struct{}{}
			keys = append(keys, k)
		}
	}
	return keys
}

// AllSettings returns all settings as a flat map of key to value.
func (c *ConfigManager) AllSettings() map[string]any {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	result := make(map[string]any, len(c.combinedConfig))
	for k, v := range c.combinedConfig {
		result[k] = v.Value
	}
	return result
}

func (c *ConfigManager) collapse() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	ccm := make(map[string]ConfigMap)
	// Precedence (highest to lowest): overrides (Set) > env > file > defaults
	for k, v := range c.defaultConfig {
		ccm[k] = v
	}
	for k, v := range c.fileConfig {
		ccm[k] = v
	}
	for k, v := range c.envConfig {
		if _, inDefaults := c.defaultConfig[k]; inDefaults {
			ccm[k] = v
		} else if _, inFile := c.fileConfig[k]; inFile {
			ccm[k] = v
		}
	}
	for k, v := range c.overrideConfig {
		ccm[k] = v
	}
	c.combinedConfig = ccm
}

func (c *ConfigManager) WriteConfig() error {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	if c.configFileUsed == "" {
		return errors.New("no config file specified")
	}
	flattenedConfig := make(map[string]any)
	for _, v := range c.combinedConfig {
		flattenedConfig[v.Key] = v.Value
	}
	switch c.configType {
	case ConfigTypeTOML:
		f, err := os.Create(c.configFileUsed)
		if err != nil {
			return err
		}
		defer f.Close()
		enc := toml.NewEncoder(f)
		err = enc.Encode(flattenedConfig)
		return err
	case ConfigTypeYAML:
		f, err := os.Create(c.configFileUsed)
		if err != nil {
			return err
		}
		defer f.Close()
		enc := yaml.NewEncoder(f)
		if err = enc.Encode(flattenedConfig); err != nil {
			return err
		}
		return enc.Close()
	case ConfigTypeJSON:
		f, err := os.Create(c.configFileUsed)
		if err != nil {
			return err
		}
		defer f.Close()
		enc := json.NewEncoder(f)
		return enc.Encode(flattenedConfig)
	default:
		return fmt.Errorf("config type %s not supported", c.configType)
	}
}

func (c *ConfigManager) SetConfigType(ct string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	switch ct {
	case "toml":
		c.configType = ConfigTypeTOML
	case "yaml":
		c.configType = ConfigTypeYAML
	case "json":
		c.configType = ConfigTypeJSON
	default:
		return fmt.Errorf("config type %s not supported", ct)
	}
	return nil
}

// SetEnvPrefix filters environment variables to only those starting with
// the given prefix, stripping the prefix from key names.
func (c *ConfigManager) SetEnvPrefix(prefix string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.envConfig = parseEnv(prefix)
}

func (c *ConfigManager) ReadInConfig() error {
	c.mutex.RLock()
	configFile := c.configFileUsed
	if configFile == "" && c.configPath != "" && c.configName != "" {
		ext := ""
		switch c.configType {
		case ConfigTypeTOML:
			ext = ".toml"
		case ConfigTypeYAML:
			ext = ".yaml"
		case ConfigTypeJSON:
			ext = ".json"
		}
		configFile = filepath.Join(c.configPath, c.configName+ext)
	}
	configType := c.configType
	c.mutex.RUnlock()

	if configFile == "" {
		return errors.New("no config file specified: use SetConfigFile or SetConfigDir + SetConfigName")
	}

	confFileData, err := readFile(configFile, configType)
	if err != nil {
		return err
	}
	conf := make(map[string]ConfigMap)
	for k, v := range confFileData {
		lower := strings.ToLower(k)
		conf[lower] = ConfigMap{Key: k, Value: v}
	}
	c.mutex.Lock()
	c.fileConfig = conf
	c.configFileUsed = configFile
	c.mutex.Unlock()
	c.collapse()
	return nil
}

func readFile(filename string, fileType configType) (map[string]any, error) {
	f, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrConfigFileNotFound
		}
		return nil, err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if info.Size() == 0 {
		return nil, ErrConfigFileEmpty
	}

	fileData := make(map[string]any)
	switch fileType {
	case ConfigTypeTOML:
		_, err := toml.NewDecoder(f).Decode(&fileData)
		return fileData, err
	case ConfigTypeYAML:
		err := yaml.NewDecoder(f).Decode(&fileData)
		return fileData, err
	case ConfigTypeJSON:
		err := json.NewDecoder(f).Decode(&fileData)
		return fileData, err
	default:
		return nil, fmt.Errorf("config type %s not supported", fileType)
	}
}

func (c *ConfigManager) SetConfigDir(path string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.configPath = path
}

func (c *ConfigManager) SetConfigName(name string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.configName = name
}

func (c *ConfigManager) SetConfigFile(file string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.configFileUsed = file
}
