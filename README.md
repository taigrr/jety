# JETY

JSON, ENV, TOML, YAML

This is a package for collapsing multiple configuration stores (env+json, env+yaml, env+toml) and writing them back to a centralized config.

It should behave similarly to the AutomaticEnv functionality of viper, but without some of the extra heft of the depedendencies it carries.

The inital purpose of this repo is to support the configuration requirements of [grlx](http://github.com/gogrlx/grlx), but development may continue to expand until more viper use cases and functionality are covered.
