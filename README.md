# JETY

JSON, ENV, YAML, TOML

This is a package for collapsing multiple configuration stores (env+json, env+yaml, env+toml) and writing them back to a centralized config.

It should behave similarly to the AutomaticEnv functionality of viper, but without some of the extra heft of the depedendencies it carries.


.AutomaticEnv
.ConfigFileUsed
.GetDuration
.GetString
.GetStringMap
.GetStringSlice
.ReadInConfig
.SetConfigFile
.SetConfigName
.SetConfigType
.SetDefault
.Set("privkey", string
viper.ConfigFileNotFoundError); ok {
.WriteConfig
