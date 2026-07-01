package jety

import (
	"os"
	"path/filepath"
	"testing"
)

func TestUnmarshal(t *testing.T) {
	type ServerConfig struct {
		Host string `json:"host"`
		Port int    `json:"port"`
	}
	type AppConfig struct {
		Name   string       `json:"name"`
		Debug  bool         `json:"debug"`
		Server ServerConfig `json:"server"`
	}

	dir := t.TempDir()
	configFile := filepath.Join(dir, "config.json")
	err := os.WriteFile(configFile, []byte(`{
		"name": "myapp",
		"debug": true,
		"server": {
			"host": "localhost",
			"port": 8080
		}
	}`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	cm := NewConfigManager()
	cm.SetConfigFile(configFile)
	if err := cm.SetConfigType("json"); err != nil {
		t.Fatal(err)
	}
	if err := cm.ReadInConfig(); err != nil {
		t.Fatal(err)
	}

	var cfg AppConfig
	if err := cm.Unmarshal(&cfg); err != nil {
		t.Fatal(err)
	}

	if cfg.Name != "myapp" {
		t.Errorf("expected name 'myapp', got %q", cfg.Name)
	}
	if !cfg.Debug {
		t.Error("expected debug to be true")
	}
	if cfg.Server.Host != "localhost" {
		t.Errorf("expected host 'localhost', got %q", cfg.Server.Host)
	}
	if cfg.Server.Port != 8080 {
		t.Errorf("expected port 8080, got %d", cfg.Server.Port)
	}
}

func TestUnmarshalWithDefaults(t *testing.T) {
	type Config struct {
		Host    string `json:"host"`
		Port    int    `json:"port"`
		Timeout int    `json:"timeout"`
	}

	dir := t.TempDir()
	configFile := filepath.Join(dir, "config.json")
	err := os.WriteFile(configFile, []byte(`{
		"host": "example.com"
	}`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	cm := NewConfigManager()
	cm.SetDefault("port", 3000)
	cm.SetDefault("timeout", 30)
	cm.SetConfigFile(configFile)
	if err := cm.SetConfigType("json"); err != nil {
		t.Fatal(err)
	}
	if err := cm.ReadInConfig(); err != nil {
		t.Fatal(err)
	}

	var cfg Config
	if err := cm.Unmarshal(&cfg); err != nil {
		t.Fatal(err)
	}

	if cfg.Host != "example.com" {
		t.Errorf("expected host 'example.com', got %q", cfg.Host)
	}
	if cfg.Port != 3000 {
		t.Errorf("expected port 3000, got %d", cfg.Port)
	}
	if cfg.Timeout != 30 {
		t.Errorf("expected timeout 30, got %d", cfg.Timeout)
	}
}

func TestUnmarshalKey(t *testing.T) {
	type DatabaseConfig struct {
		Host string `json:"host"`
		Port int    `json:"port"`
		Name string `json:"name"`
	}

	dir := t.TempDir()
	configFile := filepath.Join(dir, "config.json")
	err := os.WriteFile(configFile, []byte(`{
		"database": {
			"host": "db.example.com",
			"port": 5432,
			"name": "mydb"
		},
		"cache": {
			"host": "cache.example.com",
			"port": 6379
		}
	}`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	cm := NewConfigManager()
	cm.SetConfigFile(configFile)
	if err := cm.SetConfigType("json"); err != nil {
		t.Fatal(err)
	}
	if err := cm.ReadInConfig(); err != nil {
		t.Fatal(err)
	}

	var dbCfg DatabaseConfig
	if err := cm.UnmarshalKey("database", &dbCfg); err != nil {
		t.Fatal(err)
	}

	if dbCfg.Host != "db.example.com" {
		t.Errorf("expected host 'db.example.com', got %q", dbCfg.Host)
	}
	if dbCfg.Port != 5432 {
		t.Errorf("expected port 5432, got %d", dbCfg.Port)
	}
	if dbCfg.Name != "mydb" {
		t.Errorf("expected name 'mydb', got %q", dbCfg.Name)
	}
}

func TestUnmarshalKeyNotFound(t *testing.T) {
	cm := NewConfigManager()
	var target struct{}
	err := cm.UnmarshalKey("nonexistent", &target)
	if err == nil {
		t.Error("expected error for missing key")
	}
}

func TestUnmarshalKeyScalarValue(t *testing.T) {
	cm := NewConfigManager()
	cm.Set("count", 42)

	var count int
	if err := cm.UnmarshalKey("count", &count); err != nil {
		t.Fatal(err)
	}
	if count != 42 {
		t.Errorf("expected 42, got %d", count)
	}
}

func TestUnmarshalKeyNestedDotNotation(t *testing.T) {
	type Inner struct {
		Value string `json:"value"`
	}

	dir := t.TempDir()
	configFile := filepath.Join(dir, "config.json")
	err := os.WriteFile(configFile, []byte(`{
		"services": {
			"api": {
				"value": "hello"
			}
		}
	}`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	cm := NewConfigManager()
	cm.SetConfigFile(configFile)
	if err := cm.SetConfigType("json"); err != nil {
		t.Fatal(err)
	}
	if err := cm.ReadInConfig(); err != nil {
		t.Fatal(err)
	}

	var inner Inner
	if err := cm.UnmarshalKey("services.api", &inner); err != nil {
		t.Fatal(err)
	}
	if inner.Value != "hello" {
		t.Errorf("expected 'hello', got %q", inner.Value)
	}
}

func TestUnmarshalTOML(t *testing.T) {
	type Config struct {
		Title string `json:"title"`
		Port  int    `json:"port"`
	}

	dir := t.TempDir()
	configFile := filepath.Join(dir, "config.toml")
	err := os.WriteFile(configFile, []byte(`
title = "My App"
port = 9090
`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	cm := NewConfigManager()
	cm.SetConfigFile(configFile)
	if err := cm.SetConfigType("toml"); err != nil {
		t.Fatal(err)
	}
	if err := cm.ReadInConfig(); err != nil {
		t.Fatal(err)
	}

	var cfg Config
	if err := cm.Unmarshal(&cfg); err != nil {
		t.Fatal(err)
	}

	if cfg.Title != "My App" {
		t.Errorf("expected title 'My App', got %q", cfg.Title)
	}
	if cfg.Port != 9090 {
		t.Errorf("expected port 9090, got %d", cfg.Port)
	}
}

func TestUnmarshalYAML(t *testing.T) {
	type Config struct {
		Name    string   `json:"name"`
		Tags    []string `json:"tags"`
		Enabled bool     `json:"enabled"`
	}

	dir := t.TempDir()
	configFile := filepath.Join(dir, "config.yaml")
	err := os.WriteFile(configFile, []byte(`
name: testapp
tags:
  - web
  - api
enabled: true
`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	cm := NewConfigManager()
	cm.SetConfigFile(configFile)
	if err := cm.SetConfigType("yaml"); err != nil {
		t.Fatal(err)
	}
	if err := cm.ReadInConfig(); err != nil {
		t.Fatal(err)
	}

	var cfg Config
	if err := cm.Unmarshal(&cfg); err != nil {
		t.Fatal(err)
	}

	if cfg.Name != "testapp" {
		t.Errorf("expected name 'testapp', got %q", cfg.Name)
	}
	if len(cfg.Tags) != 2 || cfg.Tags[0] != "web" || cfg.Tags[1] != "api" {
		t.Errorf("expected tags [web api], got %v", cfg.Tags)
	}
	if !cfg.Enabled {
		t.Error("expected enabled to be true")
	}
}

func TestDefaultUnmarshal(t *testing.T) {
	// Test package-level Unmarshal function
	type Config struct {
		Key string `json:"key"`
	}

	// Reset default manager
	defaultConfigManager = NewConfigManager()
	Set("key", "value")

	var cfg Config
	if err := Unmarshal(&cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.Key != "value" {
		t.Errorf("expected 'value', got %q", cfg.Key)
	}

	// Restore
	defaultConfigManager = NewConfigManager()
}

func TestDefaultUnmarshalKey(t *testing.T) {
	defaultConfigManager = NewConfigManager()
	Set("section", map[string]any{"field": "data"})

	type Section struct {
		Field string `json:"field"`
	}
	var sec Section
	if err := UnmarshalKey("section", &sec); err != nil {
		t.Fatal(err)
	}
	if sec.Field != "data" {
		t.Errorf("expected 'data', got %q", sec.Field)
	}

	defaultConfigManager = NewConfigManager()
}

func TestUnmarshalNestedEnvOverrides(t *testing.T) {
	t.Setenv("SERVICES_API_HOST", "env.example.com")
	t.Setenv("SERVICES_API_PORT", "9090")

	cm := NewConfigManager()
	cm.SetDefault("services", map[string]any{
		"api": map[string]any{
			"host": "file.example.com",
			"port": 8080,
		},
	})

	type Config struct {
		Services struct {
			API struct {
				Host string `json:"host"`
				Port string `json:"port"`
			} `json:"api"`
		} `json:"services"`
	}

	var cfg Config
	if err := cm.Unmarshal(&cfg); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if cfg.Services.API.Host != "env.example.com" {
		t.Errorf("cfg.Services.API.Host = %q, want %q", cfg.Services.API.Host, "env.example.com")
	}
	if cfg.Services.API.Port != "9090" {
		t.Errorf("cfg.Services.API.Port = %q, want %q", cfg.Services.API.Port, "9090")
	}

	type APIConfig struct {
		Host string `json:"host"`
		Port string `json:"port"`
	}
	var apiCfg APIConfig
	if err := cm.UnmarshalKey("services.api", &apiCfg); err != nil {
		t.Fatalf("UnmarshalKey failed: %v", err)
	}
	if apiCfg.Host != "env.example.com" || apiCfg.Port != "9090" {
		t.Errorf("UnmarshalKey result = %#v, want env overrides", apiCfg)
	}
}
