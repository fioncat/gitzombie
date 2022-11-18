package config

import _ "embed"

//go:embed config.toml
var DefaultConfig string

//go:embed remote.toml
var DefaultRemote string

//go:embed builder.yaml
var DefaultBuilder string
