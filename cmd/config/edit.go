package config

import (
	"github.com/fioncat/gitzombie/cmd/app"
	"github.com/fioncat/gitzombie/config"
	"github.com/fioncat/gitzombie/pkg/errors"
	"github.com/pelletier/go-toml/v2"
)

var Config = app.Register(&app.Command[app.Empty, app.Empty]{
	Use:    "config",
	Desc:   "Edit config file",
	Action: "Edit",

	RunNoContext: func() error {
		path := config.GetDir("config.toml")
		return app.Edit(path, config.DefaultConfig, "config.toml", func(s string) error {
			data := []byte(s)
			var cfg config.Config
			err := toml.Unmarshal(data, &cfg)
			return errors.Trace(err, "parse toml")
		})
	},
})
