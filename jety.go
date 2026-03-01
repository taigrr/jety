package jety

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"
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

	ConfigMap struct {
		Key   string
		Value any
	}

	ConfigManager struct {
		configName       string
		configPath       string
		configFileUsed   string
		configType       configType
		mapConfig        map[string]ConfigMap
		defaultConfig    map[string]ConfigMap
		envConfig        map[string]ConfigMap
		combinedConfig   map[string]ConfigMap
		mutex            sync.RWMutex
		explicitDefaults bool
	}
)

var (
	ErrConfigFileNotFound = errors.New("config file not found")
	ErrConfigFileEmpty    = errors.New("config file is empty")
)

func NewConfigManager() *ConfigManager {
	cm := ConfigManager{}
	cm.envConfig = make(map[string]ConfigMap)
	cm.mapConfig = make(map[string]ConfigMap)
	cm.defaultConfig = make(map[string]ConfigMap)
	cm.combinedConfig = make(map[string]ConfigMap)
	envSet := os.Environ()
	for _, env := range envSet {
		key, value, found := strings.Cut(env, "=")
		if !found {
			continue
		}
		lower := strings.ToLower(key)
		cm.envConfig[lower] = ConfigMap{Key: key, Value: value}
	}
	return &cm
}

func (c *ConfigManager) WithEnvPrefix(prefix string) *ConfigManager {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	envSet := os.Environ()
	c.envConfig = make(map[string]ConfigMap)
	for _, env := range envSet {
		key, value, found := strings.Cut(env, "=")
		if !found {
			continue
		}
		if withoutPrefix, ok := strings.CutPrefix(key, prefix); ok {
			lower := strings.ToLower(withoutPrefix)
			c.envConfig[lower] = ConfigMap{Key: withoutPrefix, Value: value}
		}
	}
	return c
}

func (c *ConfigManager) ConfigFileUsed() string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.configFileUsed
}

func (c *ConfigManager) UseExplicitDefaults(enable bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.explicitDefaults = enable
}

func (c *ConfigManager) collapse() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	ccm := make(map[string]ConfigMap)
	for k, v := range c.defaultConfig {
		ccm[k] = v
		if _, ok := c.envConfig[k]; ok {
			ccm[k] = c.envConfig[k]
		}
	}
	maps.Copy(ccm, c.mapConfig)
	c.combinedConfig = ccm
}

func (c *ConfigManager) WriteConfig() error {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
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
		err = enc.Encode(flattenedConfig)
		return err
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

func (c *ConfigManager) SetConfigType(configType string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
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

func (c *ConfigManager) SetEnvPrefix(prefix string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	// Re-read environment variables, stripping the prefix from matching keys.
	// This mirrors WithEnvPrefix behavior so that prefixed env vars are
	// accessible by their unprefixed key name.
	envSet := os.Environ()
	c.envConfig = make(map[string]ConfigMap)
	for _, env := range envSet {
		key, value, found := strings.Cut(env, "=")
		if !found {
			continue
		}
		if withoutPrefix, ok := strings.CutPrefix(key, prefix); ok {
			lower := strings.ToLower(withoutPrefix)
			c.envConfig[lower] = ConfigMap{Key: withoutPrefix, Value: value}
		}
	}
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
	c.mapConfig = conf
	c.configFileUsed = configFile
	c.mutex.Unlock()
	c.collapse()
	return nil
}

func readFile(filename string, fileType configType) (map[string]any, error) {
	fileData := make(map[string]any)
	if d, err := os.Stat(filename); os.IsNotExist(err) {
		return nil, ErrConfigFileNotFound
	} else if d.Size() == 0 {
		return nil, ErrConfigFileEmpty
	}

	switch fileType {
	case ConfigTypeTOML:
		_, err := toml.DecodeFile(filename, &fileData)
		return fileData, err
	case ConfigTypeYAML:
		f, err := os.Open(filename)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		d := yaml.NewDecoder(f)
		err = d.Decode(&fileData)
		return fileData, err
	case ConfigTypeJSON:
		f, err := os.Open(filename)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		err = json.NewDecoder(f).Decode(&fileData)
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
