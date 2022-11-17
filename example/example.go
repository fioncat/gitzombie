package example

import _ "embed"

//go:embed config.toml
var Config string

//go:embed remote.toml
var Remote string
