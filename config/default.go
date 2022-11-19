package config

import _ "embed"

//go:embed default/config.toml
var DefaultConfig string

//go:embed default/remote.toml
var DefaultRemote string
