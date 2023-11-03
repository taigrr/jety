package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
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
		envPrefix        string
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
	cm.envPrefix = ""
	envSet := os.Environ()
	for _, env := range envSet {
		kv := strings.Split(env, "=")
		lower := strings.ToLower(kv[0])
		cm.envConfig[lower] = ConfigMap{Key: kv[0], Value: kv[1]}
	}
	return &cm
}

func (c *ConfigManager) WithEnvPrefix(prefix string) *ConfigManager {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.envPrefix = prefix
	envSet := os.Environ()
	c.envConfig = make(map[string]ConfigMap)
	for _, env := range envSet {
		kv := strings.Split(env, "=")
		if strings.HasPrefix(kv[0], prefix) {
			withoutPrefix := strings.TrimPrefix(kv[0], prefix)
			lower := strings.ToLower(withoutPrefix)
			c.envConfig[lower] = ConfigMap{Key: withoutPrefix, Value: kv[1]}
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
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	ccm := make(map[string]ConfigMap)
	for k, v := range c.defaultConfig {
		ccm[k] = v
		if _, ok := c.envConfig[k]; ok {
			ccm[k] = c.envConfig[k]
		}
	}
	for k, v := range c.mapConfig {
		ccm[k] = v
	}
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
	c.envPrefix = prefix
}

func (c *ConfigManager) ReadInConfig() error {
	// assume config = map[string]any
	confFileData, err := readFile(c.configFileUsed, c.configType)
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
