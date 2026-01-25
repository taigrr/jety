# JETY

JSON, ENV, TOML, YAML

A lightweight Go configuration management library supporting JSON, ENV, TOML, and YAML formats.
It provides viper-like `AutomaticEnv` functionality with fewer dependencies.
Originally built to support [grlx](http://github.com/gogrlx/grlx).

## Installation

```bash
go get github.com/taigrr/jety
```

Requires Go 1.25.5 or later.

## Quick Start

```go
package main

import "github.com/taigrr/jety"

func main() {
    // Set defaults
    jety.SetDefault("port", 8080)
    jety.SetDefault("host", "localhost")

    // Environment variables are loaded automatically
    // e.g., PORT=9000 overrides the default

    // Read from config file
    jety.SetConfigFile("config.toml")
    jety.SetConfigType("toml")
    if err := jety.ReadInConfig(); err != nil {
        // handle error
    }

    // Get values (config file > env > default)
    port := jety.GetInt("port")
    host := jety.GetString("host")
}
```

## Features

- **Multiple formats**: JSON, TOML, YAML
- **Automatic env loading**: Environment variables loaded on init
- **Prefix filtering**: Filter env vars by prefix (e.g., `MYAPP_`)
- **Case-insensitive keys**: Keys normalized to lowercase
- **Type coercion**: Getters handle type conversion gracefully
- **Thread-safe**: Safe for concurrent access
- **Config precedence**: config file > environment > defaults

## Nested Configuration

For nested config structures like:

```toml
[services.cloud]
var = "xyz"
timeout = "30s"

[services.cloud.auth]
client_id = "abc123"
```

Access nested values using `GetStringMap` and type assertions:

```go
services := jety.GetStringMap("services")
cloud := services["cloud"].(map[string]any)
varValue := cloud["var"].(string)  // "xyz"

// For deeper nesting
auth := cloud["auth"].(map[string]any)
clientID := auth["client_id"].(string)  // "abc123"
```

### Environment Variable Overrides

Environment variables use uppercase keys. For nested config, the env var name is the key in uppercase:

```bash
# Override top-level key
export PORT=9000

# For nested keys, use the full key name in uppercase
export SERVICES_CLOUD_VAR=override_value
```

With a prefix:

```go
cm := jety.NewConfigManager().WithEnvPrefix("MYAPP_")
```

```bash
export MYAPP_PORT=9000
export MYAPP_SERVICES_CLOUD_VAR=override_value
```

**Note**: Environment variables override defaults but config files take highest precedence.

## Migration Guide

### From v0.x to v1.x

#### Breaking Changes

1. **`WriteConfig()` now returns `error`**

   ```go
   // Before
   jety.WriteConfig()

   // After
   if err := jety.WriteConfig(); err != nil {
       // handle error
   }
   // Or if you want to ignore the error:
   _ = jety.WriteConfig()
   ```

2. **Go 1.25.5 minimum required**

   Update your Go version or pin to an older jety release.

#### Non-Breaking Improvements

- Getters (`GetBool`, `GetInt`, `GetDuration`) now return zero values instead of panicking on unknown types
- Added `int64` support in `GetInt`, `GetIntSlice`, and `GetDuration`
- Improved env var parsing (handles values containing `=`)

## API

### Configuration

| Function              | Description                                   |
| --------------------- | --------------------------------------------- |
| `SetConfigFile(path)` | Set config file path                          |
| `SetConfigDir(dir)`   | Set config directory                          |
| `SetConfigName(name)` | Set config file name (without extension)      |
| `SetConfigType(type)` | Set config type: `"toml"`, `"yaml"`, `"json"` |
| `ReadInConfig()`      | Read config file                              |
| `WriteConfig()`       | Write config to file                          |

### Values

| Function                 | Description              |
| ------------------------ | ------------------------ |
| `Set(key, value)`        | Set a value              |
| `SetDefault(key, value)` | Set a default value      |
| `Get(key)`               | Get raw value            |
| `GetString(key)`         | Get as string            |
| `GetInt(key)`            | Get as int               |
| `GetBool(key)`           | Get as bool              |
| `GetDuration(key)`       | Get as time.Duration     |
| `GetStringSlice(key)`    | Get as []string          |
| `GetIntSlice(key)`       | Get as []int             |
| `GetStringMap(key)`      | Get as map[string]string |

### Environment

| Function                | Description                                         |
| ----------------------- | --------------------------------------------------- |
| `WithEnvPrefix(prefix)` | Filter env vars by prefix (strips prefix from keys) |
| `SetEnvPrefix(prefix)`  | Set prefix for env var lookups                      |

## License

See [LICENSE](LICENSE) file.
