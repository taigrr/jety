package jety

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// Test file contents
const (
	tomlConfig = `
port = 8080
host = "localhost"
debug = true
timeout = "30s"
rate = 1.5
tags = ["api", "v1"]
counts = [1, 2, 3]

[database]
host = "db.example.com"
port = 5432
`

	yamlConfig = `
port: 9090
host: "yaml-host"
debug: false
timeout: "1m"
rate: 2.5
tags:
  - web
  - v2
counts:
  - 10
  - 20
database:
  host: "yaml-db.example.com"
  port: 3306
`

	jsonConfig = `{
  "port": 7070,
  "host": "json-host",
  "debug": true,
  "timeout": "15s",
  "rate": 3.5,
  "tags": ["json", "v3"],
  "counts": [100, 200],
  "database": {
    "host": "json-db.example.com",
    "port": 27017
  }
}`
)

func TestNewConfigManager(t *testing.T) {
	cm := NewConfigManager()
	if cm == nil {
		t.Fatal("NewConfigManager returned nil")
	}
	if cm.envConfig == nil {
		t.Error("envConfig not initialized")
	}
	if cm.overrideConfig == nil {
		t.Error("mapConfig not initialized")
	}
	if cm.defaultConfig == nil {
		t.Error("defaultConfig not initialized")
	}
	if cm.combinedConfig == nil {
		t.Error("combinedConfig not initialized")
	}
}

func TestSetAndGetString(t *testing.T) {
	cm := NewConfigManager()
	cm.Set("name", "test-value")

	got := cm.GetString("name")
	if got != "test-value" {
		t.Errorf("GetString() = %q, want %q", got, "test-value")
	}

	// Case insensitive
	got = cm.GetString("NAME")
	if got != "test-value" {
		t.Errorf("GetString(NAME) = %q, want %q", got, "test-value")
	}
}

func TestSetAndGetInt(t *testing.T) {
	cm := NewConfigManager()

	tests := []struct {
		name  string
		value any
		want  int
	}{
		{"int", 42, 42},
		{"string", "123", 123},
		{"float64", 99.9, 99},
		{"float32", float32(50.5), 50},
		{"invalid string", "not-a-number", 0},
		{"nil", nil, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm.Set("key", tt.value)
			got := cm.GetInt("key")
			if got != tt.want {
				t.Errorf("GetInt() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestSetAndGetInt64(t *testing.T) {
	cm := NewConfigManager()

	tests := []struct {
		name  string
		value any
		want  int64
	}{
		{"int64", int64(9223372036854775807), 9223372036854775807},
		{"int", 42, 42},
		{"string", "123456789012345", 123456789012345},
		{"float64", 99.9, 99},
		{"float32", float32(50.5), 50},
		{"invalid string", "not-a-number", 0},
		{"nil", nil, 0},
		{"unknown type", struct{}{}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm.Set("key", tt.value)
			got := cm.GetInt64("key")
			if got != tt.want {
				t.Errorf("GetInt64() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestSetAndGetFloat64(t *testing.T) {
	cm := NewConfigManager()

	tests := []struct {
		name  string
		value any
		want  float64
	}{
		{"float64", 3.14159, 3.14159},
		{"float32", float32(2.5), 2.5},
		{"int", 42, 42.0},
		{"int64", int64(100), 100.0},
		{"string", "1.618", 1.618},
		{"invalid string", "not-a-float", 0},
		{"nil", nil, 0},
		{"unknown type", struct{}{}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm.Set("key", tt.value)
			got := cm.GetFloat64("key")
			if got != tt.want {
				t.Errorf("GetFloat64() = %f, want %f", got, tt.want)
			}
		})
	}
}

func TestGetFloat64NotSet(t *testing.T) {
	cm := NewConfigManager()
	if got := cm.GetFloat64("nonexistent"); got != 0 {
		t.Errorf("GetFloat64(nonexistent) = %f, want 0", got)
	}
}

func TestGetInt64NotSet(t *testing.T) {
	cm := NewConfigManager()
	if got := cm.GetInt64("nonexistent"); got != 0 {
		t.Errorf("GetInt64(nonexistent) = %d, want 0", got)
	}
}

func TestSetAndGetBool(t *testing.T) {
	cm := NewConfigManager()

	tests := []struct {
		name  string
		value any
		want  bool
	}{
		{"true bool", true, true},
		{"false bool", false, false},
		{"string true", "true", true},
		{"string TRUE", "TRUE", true},
		{"string false", "false", false},
		{"string other", "yes", false},
		{"int zero", 0, false},
		{"int nonzero", 1, true},
		{"float32 zero", float32(0), false},
		{"float32 nonzero", float32(1.5), true},
		{"float64 zero", float64(0), false},
		{"float64 nonzero", float64(1.5), true},
		{"duration zero", time.Duration(0), false},
		{"duration positive", time.Second, true},
		{"nil", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm.Set("key", tt.value)
			got := cm.GetBool("key")
			if got != tt.want {
				t.Errorf("GetBool() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetAndGetDuration(t *testing.T) {
	cm := NewConfigManager()

	tests := []struct {
		name  string
		value any
		want  time.Duration
	}{
		{"duration", 5 * time.Second, 5 * time.Second},
		{"string", "10s", 10 * time.Second},
		{"string minutes", "2m", 2 * time.Minute},
		{"invalid string", "not-duration", 0},
		{"int", 1000, time.Duration(1000)},
		{"int64", int64(2000), time.Duration(2000)},
		{"float64", float64(3000), time.Duration(3000)},
		{"float32", float32(4000), time.Duration(4000)},
		{"nil", nil, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm.Set("key", tt.value)
			got := cm.GetDuration("key")
			if got != tt.want {
				t.Errorf("GetDuration() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetAndGetStringSlice(t *testing.T) {
	cm := NewConfigManager()

	// Direct string slice
	cm.Set("tags", []string{"a", "b", "c"})
	got := cm.GetStringSlice("tags")
	if len(got) != 3 || got[0] != "a" || got[1] != "b" || got[2] != "c" {
		t.Errorf("GetStringSlice() = %v, want [a b c]", got)
	}

	// []any slice
	cm.Set("mixed", []any{"x", 123, "z"})
	got = cm.GetStringSlice("mixed")
	if len(got) != 3 || got[0] != "x" || got[1] != "123" || got[2] != "z" {
		t.Errorf("GetStringSlice() = %v, want [x 123 z]", got)
	}

	// Non-slice returns nil
	cm.Set("notslice", "single")
	got = cm.GetStringSlice("notslice")
	if got != nil {
		t.Errorf("GetStringSlice() = %v, want nil", got)
	}
}

func TestSetAndGetIntSlice(t *testing.T) {
	cm := NewConfigManager()

	// Direct int slice
	cm.Set("nums", []int{1, 2, 3})
	got := cm.GetIntSlice("nums")
	if len(got) != 3 || got[0] != 1 || got[1] != 2 || got[2] != 3 {
		t.Errorf("GetIntSlice() = %v, want [1 2 3]", got)
	}

	// []any slice with mixed types
	cm.Set("mixed", []any{10, "20", float64(30), float32(40)})
	got = cm.GetIntSlice("mixed")
	if len(got) != 4 || got[0] != 10 || got[1] != 20 || got[2] != 30 || got[3] != 40 {
		t.Errorf("GetIntSlice() = %v, want [10 20 30 40]", got)
	}

	// Invalid entries skipped
	cm.Set("invalid", []any{1, "not-a-number", nil, 2})
	got = cm.GetIntSlice("invalid")
	if len(got) != 2 || got[0] != 1 || got[1] != 2 {
		t.Errorf("GetIntSlice() = %v, want [1 2]", got)
	}
}

func TestSetAndGetStringMap(t *testing.T) {
	cm := NewConfigManager()

	m := map[string]any{"foo": "bar", "num": 123}
	cm.Set("config", m)
	got := cm.GetStringMap("config")
	if got["foo"] != "bar" || got["num"] != 123 {
		t.Errorf("GetStringMap() = %v, want %v", got, m)
	}

	// Non-map returns nil
	cm.Set("notmap", "string")
	got = cm.GetStringMap("notmap")
	if got != nil {
		t.Errorf("GetStringMap() = %v, want nil", got)
	}
}

func TestGet(t *testing.T) {
	cm := NewConfigManager()
	cm.Set("key", "value")

	got := cm.Get("key")
	if got != "value" {
		t.Errorf("Get() = %v, want %q", got, "value")
	}

	// Non-existent key
	got = cm.Get("nonexistent")
	if got != nil {
		t.Errorf("Get(nonexistent) = %v, want nil", got)
	}
}

func TestSetDefault(t *testing.T) {
	cm := NewConfigManager()

	cm.SetDefault("port", 8080)
	if got := cm.GetInt("port"); got != 8080 {
		t.Errorf("GetInt(port) = %d, want 8080", got)
	}

	// Set overrides default
	cm.Set("port", 9090)
	if got := cm.GetInt("port"); got != 9090 {
		t.Errorf("GetInt(port) after Set = %d, want 9090", got)
	}

	// New default doesn't override existing value
	cm.SetDefault("port", 7070)
	if got := cm.GetInt("port"); got != 9090 {
		t.Errorf("GetInt(port) after second SetDefault = %d, want 9090", got)
	}
}

func TestReadTOMLConfig(t *testing.T) {
	dir := t.TempDir()
	configFile := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(configFile, []byte(tomlConfig), 0o644); err != nil {
		t.Fatal(err)
	}

	cm := NewConfigManager()
	cm.SetConfigFile(configFile)
	if err := cm.SetConfigType("toml"); err != nil {
		t.Fatal(err)
	}

	if err := cm.ReadInConfig(); err != nil {
		t.Fatalf("ReadInConfig() error = %v", err)
	}

	if got := cm.GetInt("port"); got != 8080 {
		t.Errorf("GetInt(port) = %d, want 8080", got)
	}
	if got := cm.GetString("host"); got != "localhost" {
		t.Errorf("GetString(host) = %q, want %q", got, "localhost")
	}
	if got := cm.GetBool("debug"); got != true {
		t.Errorf("GetBool(debug) = %v, want true", got)
	}
	if got := cm.ConfigFileUsed(); got != configFile {
		t.Errorf("ConfigFileUsed() = %q, want %q", got, configFile)
	}
}

func TestReadYAMLConfig(t *testing.T) {
	dir := t.TempDir()
	configFile := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(configFile, []byte(yamlConfig), 0o644); err != nil {
		t.Fatal(err)
	}

	cm := NewConfigManager()
	cm.SetConfigFile(configFile)
	if err := cm.SetConfigType("yaml"); err != nil {
		t.Fatal(err)
	}

	if err := cm.ReadInConfig(); err != nil {
		t.Fatalf("ReadInConfig() error = %v", err)
	}

	if got := cm.GetInt("port"); got != 9090 {
		t.Errorf("GetInt(port) = %d, want 9090", got)
	}
	if got := cm.GetString("host"); got != "yaml-host" {
		t.Errorf("GetString(host) = %q, want %q", got, "yaml-host")
	}
	if got := cm.GetBool("debug"); got != false {
		t.Errorf("GetBool(debug) = %v, want false", got)
	}
}

func TestReadJSONConfig(t *testing.T) {
	dir := t.TempDir()
	configFile := filepath.Join(dir, "config.json")
	if err := os.WriteFile(configFile, []byte(jsonConfig), 0o644); err != nil {
		t.Fatal(err)
	}

	cm := NewConfigManager()
	cm.SetConfigFile(configFile)
	if err := cm.SetConfigType("json"); err != nil {
		t.Fatal(err)
	}

	if err := cm.ReadInConfig(); err != nil {
		t.Fatalf("ReadInConfig() error = %v", err)
	}

	if got := cm.GetInt("port"); got != 7070 {
		t.Errorf("GetInt(port) = %d, want 7070", got)
	}
	if got := cm.GetString("host"); got != "json-host" {
		t.Errorf("GetString(host) = %q, want %q", got, "json-host")
	}
	if got := cm.GetBool("debug"); got != true {
		t.Errorf("GetBool(debug) = %v, want true", got)
	}
}

func TestReadConfigWithDirAndName(t *testing.T) {
	dir := t.TempDir()
	configFile := filepath.Join(dir, "myconfig.yaml")
	if err := os.WriteFile(configFile, []byte(yamlConfig), 0o644); err != nil {
		t.Fatal(err)
	}

	cm := NewConfigManager()
	cm.SetConfigDir(dir)
	cm.SetConfigName("myconfig")
	if err := cm.SetConfigType("yaml"); err != nil {
		t.Fatal(err)
	}

	if err := cm.ReadInConfig(); err != nil {
		t.Fatalf("ReadInConfig() error = %v", err)
	}

	if got := cm.GetInt("port"); got != 9090 {
		t.Errorf("GetInt(port) = %d, want 9090", got)
	}
	if got := cm.ConfigFileUsed(); got != configFile {
		t.Errorf("ConfigFileUsed() = %q, want %q", got, configFile)
	}
}

func TestReadConfigNoFileSpecified(t *testing.T) {
	cm := NewConfigManager()
	if err := cm.SetConfigType("yaml"); err != nil {
		t.Fatal(err)
	}

	err := cm.ReadInConfig()
	if err == nil {
		t.Error("ReadInConfig() expected error, got nil")
	}
}

func TestReadConfigFileNotFound(t *testing.T) {
	cm := NewConfigManager()
	cm.SetConfigFile("/nonexistent/path/config.yaml")
	if err := cm.SetConfigType("yaml"); err != nil {
		t.Fatal(err)
	}

	err := cm.ReadInConfig()
	if err != ErrConfigFileNotFound {
		t.Errorf("ReadInConfig() error = %v, want ErrConfigFileNotFound", err)
	}
}

func TestReadConfigFileEmpty(t *testing.T) {
	dir := t.TempDir()
	configFile := filepath.Join(dir, "empty.yaml")
	if err := os.WriteFile(configFile, []byte{}, 0o644); err != nil {
		t.Fatal(err)
	}

	cm := NewConfigManager()
	cm.SetConfigFile(configFile)
	if err := cm.SetConfigType("yaml"); err != nil {
		t.Fatal(err)
	}

	err := cm.ReadInConfig()
	if err != ErrConfigFileEmpty {
		t.Errorf("ReadInConfig() error = %v, want ErrConfigFileEmpty", err)
	}
}

func TestWriteConfig(t *testing.T) {
	dir := t.TempDir()

	tests := []struct {
		name       string
		configType string
		ext        string
	}{
		{"TOML", "toml", ".toml"},
		{"YAML", "yaml", ".yaml"},
		{"JSON", "json", ".json"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configFile := filepath.Join(dir, "write_test"+tt.ext)

			cm := NewConfigManager()
			cm.SetConfigFile(configFile)
			if err := cm.SetConfigType(tt.configType); err != nil {
				t.Fatal(err)
			}

			cm.Set("port", 8080)
			cm.Set("host", "example.com")

			if err := cm.WriteConfig(); err != nil {
				t.Fatalf("WriteConfig() error = %v", err)
			}

			// Read it back
			cm2 := NewConfigManager()
			cm2.SetConfigFile(configFile)
			if err := cm2.SetConfigType(tt.configType); err != nil {
				t.Fatal(err)
			}
			if err := cm2.ReadInConfig(); err != nil {
				t.Fatalf("ReadInConfig() error = %v", err)
			}

			if got := cm2.GetInt("port"); got != 8080 {
				t.Errorf("GetInt(port) = %d, want 8080", got)
			}
			if got := cm2.GetString("host"); got != "example.com" {
				t.Errorf("GetString(host) = %q, want %q", got, "example.com")
			}
		})
	}
}

func TestSetConfigTypeInvalid(t *testing.T) {
	cm := NewConfigManager()
	err := cm.SetConfigType("xml")
	if err == nil {
		t.Error("SetConfigType(xml) expected error, got nil")
	}
}

func TestEnvPrefix(t *testing.T) {
	// Set env vars BEFORE creating ConfigManager
	os.Setenv("TESTAPP_PORT", "3000")
	os.Setenv("TESTAPP_HOST", "envhost")
	os.Setenv("OTHER_VAR", "other")
	defer func() {
		os.Unsetenv("TESTAPP_PORT")
		os.Unsetenv("TESTAPP_HOST")
		os.Unsetenv("OTHER_VAR")
	}()

	// Create new manager AFTER setting env vars, then apply prefix
	cm := NewConfigManager().WithEnvPrefix("TESTAPP_")

	if got := cm.GetString("port"); got != "3000" {
		t.Errorf("GetString(port) = %q, want %q", got, "3000")
	}
	if got := cm.GetString("host"); got != "envhost" {
		t.Errorf("GetString(host) = %q, want %q", got, "envhost")
	}
	// OTHER_VAR should not be accessible without prefix
	if got := cm.GetString("other_var"); got != "" {
		t.Errorf("GetString(other_var) = %q, want empty", got)
	}
}

func TestEnvVarWithEqualsInValue(t *testing.T) {
	os.Setenv("TEST_CONN", "host=localhost;user=admin")
	defer os.Unsetenv("TEST_CONN")

	cm := NewConfigManager()
	if got := cm.GetString("test_conn"); got != "host=localhost;user=admin" {
		t.Errorf("GetString(test_conn) = %q, want %q", got, "host=localhost;user=admin")
	}
}

func TestEnvOverridesDefault(t *testing.T) {
	os.Setenv("MYPORT", "5000")
	defer os.Unsetenv("MYPORT")

	cm := NewConfigManager()
	cm.SetDefault("myport", 8080)

	if got := cm.GetInt("myport"); got != 5000 {
		t.Errorf("GetInt(myport) = %d, want 5000 (from env)", got)
	}
}

func TestEnvOverridesConfigFile(t *testing.T) {
	os.Setenv("PORT", "5000")
	defer os.Unsetenv("PORT")

	dir := t.TempDir()
	configFile := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(configFile, []byte("port: 9000"), 0o644); err != nil {
		t.Fatal(err)
	}

	cm := NewConfigManager()
	cm.SetConfigFile(configFile)
	if err := cm.SetConfigType("yaml"); err != nil {
		t.Fatal(err)
	}
	cm.SetDefault("port", 8080)

	if err := cm.ReadInConfig(); err != nil {
		t.Fatal(err)
	}

	// Env should override config file (env > file > defaults)
	if got := cm.GetInt("port"); got != 5000 {
		t.Errorf("GetInt(port) = %d, want 5000 (env overrides file)", got)
	}
}

func TestCaseInsensitiveKeys(t *testing.T) {
	cm := NewConfigManager()
	cm.Set("MyKey", "value")

	tests := []string{"MyKey", "mykey", "MYKEY", "mYkEy"}
	for _, key := range tests {
		if got := cm.GetString(key); got != "value" {
			t.Errorf("GetString(%q) = %q, want %q", key, got, "value")
		}
	}
}

func TestGetNonExistentKey(t *testing.T) {
	cm := NewConfigManager()

	if got := cm.GetString("nonexistent"); got != "" {
		t.Errorf("GetString(nonexistent) = %q, want empty", got)
	}
	if got := cm.GetInt("nonexistent"); got != 0 {
		t.Errorf("GetInt(nonexistent) = %d, want 0", got)
	}
	if got := cm.GetBool("nonexistent"); got != false {
		t.Errorf("GetBool(nonexistent) = %v, want false", got)
	}
	if got := cm.GetDuration("nonexistent"); got != 0 {
		t.Errorf("GetDuration(nonexistent) = %v, want 0", got)
	}
	if got := cm.GetStringSlice("nonexistent"); got != nil {
		t.Errorf("GetStringSlice(nonexistent) = %v, want nil", got)
	}
	if got := cm.GetIntSlice("nonexistent"); got != nil {
		t.Errorf("GetIntSlice(nonexistent) = %v, want nil", got)
	}
	if got := cm.GetStringMap("nonexistent"); got != nil {
		t.Errorf("GetStringMap(nonexistent) = %v, want nil", got)
	}
}

func TestGetBoolUnknownType(t *testing.T) {
	cm := NewConfigManager()
	cm.Set("key", struct{}{})

	// Should not panic, should return false
	got := cm.GetBool("key")
	if got != false {
		t.Errorf("GetBool(struct) = %v, want false", got)
	}
}

func TestGetDurationUnknownType(t *testing.T) {
	cm := NewConfigManager()
	cm.Set("key", struct{}{})

	// Should not panic, should return 0
	got := cm.GetDuration("key")
	if got != 0 {
		t.Errorf("GetDuration(struct) = %v, want 0", got)
	}
}

func TestConcurrentAccess(t *testing.T) {
	cm := NewConfigManager()
	var wg sync.WaitGroup

	// Concurrent writes
	for i := range 100 {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			cm.Set("key", n)
			cm.SetDefault("default", n)
			cm.SetString("str", "value")
			cm.SetBool("bool", true)
		}(i)
	}

	// Concurrent reads
	for range 100 {
		wg.Go(func() {
			_ = cm.GetInt("key")
			_ = cm.GetString("str")
			_ = cm.GetBool("bool")
			_ = cm.Get("key")
			_ = cm.ConfigFileUsed()
		})
	}

	wg.Wait()
}

func TestConcurrentReadWrite(t *testing.T) {
	dir := t.TempDir()
	configFile := filepath.Join(dir, "concurrent.yaml")
	if err := os.WriteFile(configFile, []byte("port: 8080"), 0o644); err != nil {
		t.Fatal(err)
	}

	cm := NewConfigManager()
	cm.SetConfigFile(configFile)
	if err := cm.SetConfigType("yaml"); err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup

	// Reader goroutines
	for range 50 {
		wg.Go(func() {
			for range 10 {
				_ = cm.GetInt("port")
				_ = cm.GetString("host")
			}
		})
	}

	// Writer goroutines
	for i := range 50 {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			for range 10 {
				cm.Set("port", n)
				cm.SetDefault("host", "localhost")
			}
		}(i)
	}

	// Config operations
	for range 10 {
		wg.Go(func() {
			_ = cm.ReadInConfig()
		})
	}

	wg.Wait()
}

// Package-level function tests (default.go)

func TestPackageLevelFunctions(t *testing.T) {
	// Reset default manager for this test
	defaultConfigManager = NewConfigManager()

	dir := t.TempDir()
	configFile := filepath.Join(dir, "pkg_test.yaml")
	if err := os.WriteFile(configFile, []byte("port: 8888\nhost: pkghost"), 0o644); err != nil {
		t.Fatal(err)
	}

	SetConfigFile(configFile)
	if err := SetConfigType("yaml"); err != nil {
		t.Fatal(err)
	}
	SetDefault("timeout", "30s")

	if err := ReadInConfig(); err != nil {
		t.Fatalf("ReadInConfig() error = %v", err)
	}

	// Set() must be called AFTER ReadInConfig to override file values
	Set("debug", true)

	if got := GetInt("port"); got != 8888 {
		t.Errorf("GetInt(port) = %d, want 8888", got)
	}
	if got := GetString("host"); got != "pkghost" {
		t.Errorf("GetString(host) = %q, want %q", got, "pkghost")
	}
	if got := GetBool("debug"); got != true {
		t.Errorf("GetBool(debug) = %v, want true", got)
	}
	if got := GetDuration("timeout"); got != 30*time.Second {
		t.Errorf("GetDuration(timeout) = %v, want 30s", got)
	}
	if got := ConfigFileUsed(); got != configFile {
		t.Errorf("ConfigFileUsed() = %q, want %q", got, configFile)
	}
}

func TestSetString(t *testing.T) {
	cm := NewConfigManager()
	cm.SetString("name", "test")

	if got := cm.GetString("name"); got != "test" {
		t.Errorf("GetString(name) = %q, want %q", got, "test")
	}
}

func TestSetBool(t *testing.T) {
	cm := NewConfigManager()
	cm.SetBool("enabled", true)

	if got := cm.GetBool("enabled"); got != true {
		t.Errorf("GetBool(enabled) = %v, want true", got)
	}
}

func TestWriteConfigUnsupportedType(t *testing.T) {
	dir := t.TempDir()
	configFile := filepath.Join(dir, "test.txt")

	cm := NewConfigManager()
	cm.SetConfigFile(configFile)
	// Don't set config type

	err := cm.WriteConfig()
	if err == nil {
		t.Error("WriteConfig() expected error for unsupported type, got nil")
	}
}

func TestSetEnvPrefix(t *testing.T) {
	cm := NewConfigManager()
	cm.SetEnvPrefix("PREFIX_")
}

func TestDeeplyNestedConfig(t *testing.T) {
	const nestedYAML = `
app:
  name: myapp
  server:
    host: localhost
    port: 8080
    tls:
      enabled: true
      cert: /path/to/cert.pem
      key: /path/to/key.pem
  database:
    primary:
      host: db1.example.com
      port: 5432
      credentials:
        username: admin
        password: secret
    replicas:
      - host: db2.example.com
        port: 5432
      - host: db3.example.com
        port: 5432
  features:
    - name: feature1
      enabled: true
      config:
        timeout: 30s
        retries: 3
    - name: feature2
      enabled: false
`

	const nestedTOML = `
[app]
name = "myapp"

[app.server]
host = "localhost"
port = 8080

[app.server.tls]
enabled = true
cert = "/path/to/cert.pem"
key = "/path/to/key.pem"

[app.database.primary]
host = "db1.example.com"
port = 5432

[app.database.primary.credentials]
username = "admin"
password = "secret"

[[app.database.replicas]]
host = "db2.example.com"
port = 5432

[[app.database.replicas]]
host = "db3.example.com"
port = 5432

[[app.features]]
name = "feature1"
enabled = true

[app.features.config]
timeout = "30s"
retries = 3

[[app.features]]
name = "feature2"
enabled = false
`

	const nestedJSON = `{
  "app": {
    "name": "myapp",
    "server": {
      "host": "localhost",
      "port": 8080,
      "tls": {
        "enabled": true,
        "cert": "/path/to/cert.pem",
        "key": "/path/to/key.pem"
      }
    },
    "database": {
      "primary": {
        "host": "db1.example.com",
        "port": 5432,
        "credentials": {
          "username": "admin",
          "password": "secret"
        }
      },
      "replicas": [
        {"host": "db2.example.com", "port": 5432},
        {"host": "db3.example.com", "port": 5432}
      ]
    },
    "features": [
      {
        "name": "feature1",
        "enabled": true,
        "config": {
          "timeout": "30s",
          "retries": 3
        }
      },
      {
        "name": "feature2",
        "enabled": false
      }
    ]
  }
}`

	tests := []struct {
		name       string
		configType string
		content    string
		ext        string
	}{
		{"YAML", "yaml", nestedYAML, ".yaml"},
		{"TOML", "toml", nestedTOML, ".toml"},
		{"JSON", "json", nestedJSON, ".json"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			configFile := filepath.Join(dir, "nested"+tt.ext)
			if err := os.WriteFile(configFile, []byte(tt.content), 0o644); err != nil {
				t.Fatal(err)
			}

			cm := NewConfigManager()
			cm.SetConfigFile(configFile)
			if err := cm.SetConfigType(tt.configType); err != nil {
				t.Fatal(err)
			}

			if err := cm.ReadInConfig(); err != nil {
				t.Fatalf("ReadInConfig() error = %v", err)
			}

			// Test that we can retrieve the top-level nested structure
			app := cm.GetStringMap("app")
			if app == nil {
				t.Fatal("GetStringMap(app) = nil, want nested map")
			}

			// Verify app.name exists
			if name, ok := app["name"].(string); !ok || name != "myapp" {
				t.Errorf("app.name = %v, want %q", app["name"], "myapp")
			}

			// Verify nested server config
			server, ok := app["server"].(map[string]any)
			if !ok {
				t.Fatalf("app.server is not a map: %T", app["server"])
			}
			if server["host"] != "localhost" {
				t.Errorf("app.server.host = %v, want %q", server["host"], "localhost")
			}

			// Verify deeply nested TLS config
			tls, ok := server["tls"].(map[string]any)
			if !ok {
				t.Fatalf("app.server.tls is not a map: %T", server["tls"])
			}
			if tls["enabled"] != true {
				t.Errorf("app.server.tls.enabled = %v, want true", tls["enabled"])
			}
			if tls["cert"] != "/path/to/cert.pem" {
				t.Errorf("app.server.tls.cert = %v, want %q", tls["cert"], "/path/to/cert.pem")
			}

			// Verify database.primary.credentials (4 levels deep)
			database, ok := app["database"].(map[string]any)
			if !ok {
				t.Fatalf("app.database is not a map: %T", app["database"])
			}
			primary, ok := database["primary"].(map[string]any)
			if !ok {
				t.Fatalf("app.database.primary is not a map: %T", database["primary"])
			}
			creds, ok := primary["credentials"].(map[string]any)
			if !ok {
				t.Fatalf("app.database.primary.credentials is not a map: %T", primary["credentials"])
			}
			if creds["username"] != "admin" {
				t.Errorf("credentials.username = %v, want %q", creds["username"], "admin")
			}

			// Verify array of nested objects (replicas)
			// TOML decodes to []map[string]interface{}, YAML/JSON to []any
			var replicaHost any
			switch r := database["replicas"].(type) {
			case []any:
				if len(r) != 2 {
					t.Errorf("len(replicas) = %d, want 2", len(r))
				}
				if len(r) > 0 {
					replica0, ok := r[0].(map[string]any)
					if !ok {
						t.Fatalf("replicas[0] is not a map: %T", r[0])
					}
					replicaHost = replica0["host"]
				}
			case []map[string]any:
				if len(r) != 2 {
					t.Errorf("len(replicas) = %d, want 2", len(r))
				}
				if len(r) > 0 {
					replicaHost = r[0]["host"]
				}
			default:
				t.Fatalf("app.database.replicas unexpected type: %T", database["replicas"])
			}
			if replicaHost != "db2.example.com" {
				t.Errorf("replicas[0].host = %v, want %q", replicaHost, "db2.example.com")
			}

			// Verify features array with nested config
			// TOML decodes to []map[string]interface{}, YAML/JSON to []any
			var featureName any
			switch f := app["features"].(type) {
			case []any:
				if len(f) < 1 {
					t.Fatal("features slice is empty")
				}
				feature0, ok := f[0].(map[string]any)
				if !ok {
					t.Fatalf("features[0] is not a map: %T", f[0])
				}
				featureName = feature0["name"]
			case []map[string]any:
				if len(f) < 1 {
					t.Fatal("features slice is empty")
				}
				featureName = f[0]["name"]
			default:
				t.Fatalf("app.features unexpected type: %T", app["features"])
			}
			if featureName != "feature1" {
				t.Errorf("features[0].name = %v, want %q", featureName, "feature1")
			}
		})
	}
}

func TestPackageLevelGetIntSlice(t *testing.T) {
	defaultConfigManager = NewConfigManager()
	Set("nums", []int{1, 2, 3})
	got := GetIntSlice("nums")
	if len(got) != 3 || got[0] != 1 {
		t.Errorf("GetIntSlice() = %v, want [1 2 3]", got)
	}
}

func TestPackageLevelSetConfigName(t *testing.T) {
	defaultConfigManager = NewConfigManager()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "app.json"), []byte(`{"port": 1234}`), 0o644); err != nil {
		t.Fatal(err)
	}
	SetConfigName("app")
	defaultConfigManager.SetConfigDir(dir)
	if err := SetConfigType("json"); err != nil {
		t.Fatal(err)
	}
	if err := ReadInConfig(); err != nil {
		t.Fatal(err)
	}
	if got := GetInt("port"); got != 1234 {
		t.Errorf("GetInt(port) = %d, want 1234", got)
	}
}

func TestPackageLevelSetEnvPrefix(t *testing.T) {
	defaultConfigManager = NewConfigManager()
	SetEnvPrefix("JETY_TEST_")
}

func TestPackageLevelWriteConfig(t *testing.T) {
	defaultConfigManager = NewConfigManager()
	dir := t.TempDir()
	f := filepath.Join(dir, "out.json")
	SetConfigFile(f)
	if err := SetConfigType("json"); err != nil {
		t.Fatal(err)
	}
	Set("key", "value")
	if err := WriteConfig(); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(f)
	if len(data) == 0 {
		t.Error("WriteConfig produced empty file")
	}
}

func TestPackageLevelGetStringMap(t *testing.T) {
	defaultConfigManager = NewConfigManager()
	Set("m", map[string]any{"a": 1})
	got := GetStringMap("m")
	if got == nil || got["a"] != 1 {
		t.Errorf("GetStringMap() = %v", got)
	}
}

func TestPackageLevelGetStringSlice(t *testing.T) {
	defaultConfigManager = NewConfigManager()
	Set("s", []string{"a", "b"})
	got := GetStringSlice("s")
	if len(got) != 2 || got[0] != "a" {
		t.Errorf("GetStringSlice() = %v", got)
	}
}

func TestGetStringNonStringValue(t *testing.T) {
	cm := NewConfigManager()
	cm.Set("num", 42)
	if got := cm.GetString("num"); got != "42" {
		t.Errorf("GetString(num) = %q, want %q", got, "42")
	}
}

func TestGetIntInt64(t *testing.T) {
	cm := NewConfigManager()
	cm.Set("key", int64(999))
	if got := cm.GetInt("key"); got != 999 {
		t.Errorf("GetInt(int64) = %d, want 999", got)
	}
}

func TestGetIntUnknownType(t *testing.T) {
	cm := NewConfigManager()
	cm.Set("key", struct{}{})
	if got := cm.GetInt("key"); got != 0 {
		t.Errorf("GetInt(struct) = %d, want 0", got)
	}
}

func TestGetIntSliceInt64Values(t *testing.T) {
	cm := NewConfigManager()
	cm.Set("key", []any{int64(10), int64(20)})
	got := cm.GetIntSlice("key")
	if len(got) != 2 || got[0] != 10 || got[1] != 20 {
		t.Errorf("GetIntSlice(int64) = %v, want [10 20]", got)
	}
}

func TestGetIntSliceNonSlice(t *testing.T) {
	cm := NewConfigManager()
	cm.Set("key", "notaslice")
	if got := cm.GetIntSlice("key"); got != nil {
		t.Errorf("GetIntSlice(string) = %v, want nil", got)
	}
}

func TestGetIntSliceUnknownInnerType(t *testing.T) {
	cm := NewConfigManager()
	cm.Set("key", []any{struct{}{}, true})
	got := cm.GetIntSlice("key")
	if len(got) != 0 {
		t.Errorf("GetIntSlice(unknown types) = %v, want []", got)
	}
}

func TestDeeplyNestedWriteConfig(t *testing.T) {
	dir := t.TempDir()

	tests := []struct {
		name       string
		configType string
		ext        string
	}{
		{"YAML", "yaml", ".yaml"},
		{"TOML", "toml", ".toml"},
		{"JSON", "json", ".json"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configFile := filepath.Join(dir, "nested_write"+tt.ext)

			cm := NewConfigManager()
			cm.SetConfigFile(configFile)
			if err := cm.SetConfigType(tt.configType); err != nil {
				t.Fatal(err)
			}

			// Set a deeply nested structure
			nested := map[string]any{
				"server": map[string]any{
					"host": "localhost",
					"port": 8080,
					"tls": map[string]any{
						"enabled": true,
						"cert":    "/path/to/cert.pem",
					},
				},
				"database": map[string]any{
					"primary": map[string]any{
						"host": "db.example.com",
						"port": 5432,
					},
				},
			}
			cm.Set("app", nested)

			if err := cm.WriteConfig(); err != nil {
				t.Fatalf("WriteConfig() error = %v", err)
			}

			// Read it back
			cm2 := NewConfigManager()
			cm2.SetConfigFile(configFile)
			if err := cm2.SetConfigType(tt.configType); err != nil {
				t.Fatal(err)
			}
			if err := cm2.ReadInConfig(); err != nil {
				t.Fatalf("ReadInConfig() error = %v", err)
			}

			// Verify nested structure was preserved
			app := cm2.GetStringMap("app")
			if app == nil {
				t.Fatal("GetStringMap(app) = nil after read")
			}

			server, ok := app["server"].(map[string]any)
			if !ok {
				t.Fatalf("app.server is not a map: %T", app["server"])
			}
			if server["host"] != "localhost" {
				t.Errorf("app.server.host = %v, want %q", server["host"], "localhost")
			}

			tls, ok := server["tls"].(map[string]any)
			if !ok {
				t.Fatalf("app.server.tls is not a map: %T", server["tls"])
			}
			if tls["enabled"] != true {
				t.Errorf("app.server.tls.enabled = %v, want true", tls["enabled"])
			}
		})
	}
}

func TestSetEnvPrefixOverridesDefault(t *testing.T) {
	// Subprocess test: env vars must exist before NewConfigManager is called.
	if os.Getenv("TEST_SET_ENV_PREFIX") == "1" {
		cm := NewConfigManager()
		cm.SetEnvPrefix("MYAPP_")
		cm.SetDefault("port", 8080)

		if got := cm.GetInt("port"); got != 9999 {
			fmt.Fprintf(os.Stderr, "GetInt(port) = %d, want 9999\n", got)
			os.Exit(1)
		}
		if got := cm.GetString("host"); got != "envhost" {
			fmt.Fprintf(os.Stderr, "GetString(host) = %q, want %q\n", got, "envhost")
			os.Exit(1)
		}
		// Unprefixed var should not be visible.
		if got := cm.GetString("other"); got != "" {
			fmt.Fprintf(os.Stderr, "GetString(other) = %q, want empty\n", got)
			os.Exit(1)
		}
		os.Exit(0)
	}

	cmd := exec.Command(os.Args[0], "-test.run=^TestSetEnvPrefixOverridesDefault$")
	cmd.Env = append(os.Environ(),
		"TEST_SET_ENV_PREFIX=1",
		"MYAPP_PORT=9999",
		"MYAPP_HOST=envhost",
		"OTHER=should_not_see",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("subprocess failed: %v\n%s", err, out)
	}
}

func TestSetEnvPrefixWithSetDefault(t *testing.T) {
	// SetDefault should pick up prefixed env vars after SetEnvPrefix.
	if os.Getenv("TEST_SET_ENV_PREFIX_DEFAULT") == "1" {
		cm := NewConfigManager()
		cm.SetEnvPrefix("APP_")
		cm.SetDefault("database_host", "localhost")

		if got := cm.GetString("database_host"); got != "db.example.com" {
			fmt.Fprintf(os.Stderr, "GetString(database_host) = %q, want %q\n", got, "db.example.com")
			os.Exit(1)
		}
		os.Exit(0)
	}

	cmd := exec.Command(os.Args[0], "-test.run=^TestSetEnvPrefixWithSetDefault$")
	cmd.Env = append(os.Environ(),
		"TEST_SET_ENV_PREFIX_DEFAULT=1",
		"APP_DATABASE_HOST=db.example.com",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("subprocess failed: %v\n%s", err, out)
	}
}

func TestPackageLevelSetEnvPrefixOverrides(t *testing.T) {
	// Package-level SetEnvPrefix should work the same way.
	if os.Getenv("TEST_PKG_SET_ENV_PREFIX") == "1" {
		// Reset the default manager to pick up our env vars.
		defaultConfigManager = NewConfigManager()
		SetEnvPrefix("PKG_")
		SetDefault("val", "default")

		if got := GetString("val"); got != "from_env" {
			fmt.Fprintf(os.Stderr, "GetString(val) = %q, want %q\n", got, "from_env")
			os.Exit(1)
		}
		os.Exit(0)
	}

	cmd := exec.Command(os.Args[0], "-test.run=^TestPackageLevelSetEnvPrefixOverrides$")
	cmd.Env = append(os.Environ(),
		"TEST_PKG_SET_ENV_PREFIX=1",
		"PKG_VAL=from_env",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("subprocess failed: %v\n%s", err, out)
	}
}

func TestPrecedenceChain(t *testing.T) {
	// Verify: Set > env > file > defaults
	os.Setenv("PORT", "5000")
	os.Setenv("HOST", "envhost")
	os.Setenv("LOG", "envlog")
	defer os.Unsetenv("PORT")
	defer os.Unsetenv("HOST")
	defer os.Unsetenv("LOG")

	dir := t.TempDir()
	configFile := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(configFile, []byte("port: 9000\nhost: filehost\nlog: filelog\nname: filename"), 0o644); err != nil {
		t.Fatal(err)
	}

	cm := NewConfigManager()
	cm.SetDefault("port", 8080)
	cm.SetDefault("host", "defaulthost")
	cm.SetDefault("log", "defaultlog")
	cm.SetDefault("name", "defaultname")
	cm.SetDefault("extra", "defaultextra")

	cm.SetConfigFile(configFile)
	if err := cm.SetConfigType("yaml"); err != nil {
		t.Fatal(err)
	}
	if err := cm.ReadInConfig(); err != nil {
		t.Fatal(err)
	}

	cm.Set("port", 1111) // Set overrides everything

	// port: Set(1111) > env(5000) > file(9000) > default(8080) → 1111
	if got := cm.GetInt("port"); got != 1111 {
		t.Errorf("port: got %d, want 1111 (Set overrides all)", got)
	}
	// host: env(envhost) > file(filehost) > default(defaulthost) → envhost
	if got := cm.GetString("host"); got != "envhost" {
		t.Errorf("host: got %q, want envhost (env overrides file)", got)
	}
	// name: file(filename) > default(defaultname) → filename
	if got := cm.GetString("name"); got != "filename" {
		t.Errorf("name: got %q, want filename (file overrides default)", got)
	}
	// extra: only default → defaultextra
	if got := cm.GetString("extra"); got != "defaultextra" {
		t.Errorf("extra: got %q, want defaultextra (default)", got)
	}
}

func TestIsSet(t *testing.T) {
	cm := NewConfigManager()
	cm.Set("exists", "yes")

	if !cm.IsSet("exists") {
		t.Error("IsSet(exists) = false, want true")
	}
	if cm.IsSet("nope") {
		t.Error("IsSet(nope) = true, want false")
	}
}

func TestAllKeys(t *testing.T) {
	cm := NewConfigManager()
	cm.SetDefault("a", 1)
	cm.Set("b", 2)

	keys := cm.AllKeys()
	found := make(map[string]bool)
	for _, k := range keys {
		found[k] = true
	}
	if !found["a"] || !found["b"] {
		t.Errorf("AllKeys() = %v, want to contain a and b", keys)
	}
}

func TestAllSettings(t *testing.T) {
	cm := NewConfigManager()
	cm.Set("port", 8080)
	cm.Set("host", "localhost")

	settings := cm.AllSettings()
	if settings["port"] != 8080 || settings["host"] != "localhost" {
		t.Errorf("AllSettings() = %v", settings)
	}
}

func TestEnvOverridesFileWithoutDefault(t *testing.T) {
	// Bug fix: env should override file even when no default is set for that key
	os.Setenv("HOST", "envhost")
	defer os.Unsetenv("HOST")

	dir := t.TempDir()
	configFile := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(configFile, []byte("host: filehost"), 0o644); err != nil {
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

	// No SetDefault("host", ...) was called — env should still win
	if got := cm.GetString("host"); got != "envhost" {
		t.Errorf("GetString(host) = %q, want envhost (env overrides file even without default)", got)
	}
}

func TestPackageLevelIsSet(t *testing.T) {
	defaultConfigManager = NewConfigManager()
	Set("x", 1)
	if !IsSet("x") {
		t.Error("IsSet(x) = false")
	}
}

func TestPackageLevelGet(t *testing.T) {
	defaultConfigManager = NewConfigManager()
	Set("key", "val")
	if Get("key") != "val" {
		t.Error("Get(key) failed")
	}
}

func TestPackageLevelGetFloat64(t *testing.T) {
	defaultConfigManager = NewConfigManager()
	Set("rate", 3.14)
	if got := GetFloat64("rate"); got != 3.14 {
		t.Errorf("GetFloat64(rate) = %f, want 3.14", got)
	}
}

func TestPackageLevelGetInt64(t *testing.T) {
	defaultConfigManager = NewConfigManager()
	Set("bignum", int64(9223372036854775807))
	if got := GetInt64("bignum"); got != 9223372036854775807 {
		t.Errorf("GetInt64(bignum) = %d, want 9223372036854775807", got)
	}
}

func TestSub(t *testing.T) {
	dir := t.TempDir()
	configFile := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(configFile, []byte(tomlConfig), 0o644); err != nil {
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

	sub := cm.Sub("database")
	if sub == nil {
		t.Fatal("Sub(database) returned nil")
	}

	host := sub.GetString("host")
	if host != "db.example.com" {
		t.Errorf("Sub(database).GetString(host) = %q, want %q", host, "db.example.com")
	}

	port := sub.GetInt("port")
	if port != 5432 {
		t.Errorf("Sub(database).GetInt(port) = %d, want 5432", port)
	}
}

func TestSubNonExistentKey(t *testing.T) {
	cm := NewConfigManager()
	cm.SetDefault("simple", "value")

	sub := cm.Sub("nonexistent")
	if sub != nil {
		t.Error("Sub(nonexistent) should return nil")
	}
}

func TestSubNonMapKey(t *testing.T) {
	cm := NewConfigManager()
	cm.Set("name", "plain-string")

	sub := cm.Sub("name")
	if sub != nil {
		t.Error("Sub on a non-map key should return nil")
	}
}

func TestSubIsIndependent(t *testing.T) {
	cm := NewConfigManager()
	cm.Set("section", map[string]any{
		"key1": "val1",
		"key2": "val2",
	})

	sub := cm.Sub("section")
	if sub == nil {
		t.Fatal("Sub(section) returned nil")
	}

	// Modifying sub should not affect parent
	sub.Set("key1", "modified")
	if sub.GetString("key1") != "modified" {
		t.Error("sub should reflect the Set")
	}

	parentSection := cm.Get("section").(map[string]any)
	if parentSection["key1"] != "val1" {
		t.Error("modifying sub should not affect parent")
	}
}

func TestPackageLevelSub(t *testing.T) {
	defaultConfigManager = NewConfigManager()
	Set("db", map[string]any{"host": "localhost", "port": 5432})

	sub := Sub("db")
	if sub == nil {
		t.Fatal("package-level Sub returned nil")
	}
	if sub.GetString("host") != "localhost" {
		t.Errorf("Sub(db).GetString(host) = %q, want %q", sub.GetString("host"), "localhost")
	}
	if sub.GetInt("port") != 5432 {
		t.Errorf("Sub(db).GetInt(port) = %d, want 5432", sub.GetInt("port"))
	}
}
