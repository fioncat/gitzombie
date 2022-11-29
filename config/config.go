package config

import (
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fioncat/gitzombie/pkg/errors"
	"github.com/fioncat/gitzombie/pkg/osutil"
	"github.com/fioncat/gitzombie/pkg/validate"
	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	Workspace  string `toml:"workspace" default:"$HOME/dev/src" env:"true"`
	Playground string `toml:"playground" default:"$HOME/dev/play" env:"true"`

	SearchLimit int `toml:"search_limit" default:"200"`

	Editor string `toml:"editor" default:"vim"`
}

var (
	baseDir  string
	homeDir  string
	localDir string

	instance *Config

	initOnce sync.Once
)

func Init() error {
	var err error
	initOnce.Do(func() {
		homeDir, err = os.UserHomeDir()
		if err != nil {
			err = errors.Trace(err, "get home dir")
			return
		}

		// The config file should be created by user manually, so we donot need
		// to ensure it.
		baseDir = filepath.Join(homeDir, ".config", "gitzombie")

		// The local dir must be ensured before using because we need to store
		// some data under it.
		localDir = filepath.Join(homeDir, ".local", "share", "gitzombie")
		err = osutil.EnsureDir(localDir)
		if err != nil {
			err = errors.Trace(err, "ensure local dir")
			return
		}

		configPath := filepath.Join(baseDir, "config.toml")
		_, err = os.Stat(configPath)
		switch {
		case os.IsNotExist(err):
			// Okay, user did not create config file, let's use the default one.
			// The field does not need to be assigned here since the normalize()
			// method would do it for us.
			instance = &Config{}
			err = nil // The NotExist error should be discarded

		case err == nil:
			// User have created config file, read and parse it.
			var data []byte
			data, err = os.ReadFile(configPath)
			if err != nil {
				err = errors.Trace(err, "read config file")
				return
			}
			instance = new(Config)
			err = toml.Unmarshal(data, instance)
			if err != nil {
				err = errors.Trace(err, "parse config file")
				return
			}

		default:
			err = errors.Trace(err, "stat config file")
			return
		}
		validate.ExpandDefault(instance)
	})
	return err
}

func ensureInit() {
	if instance == nil || homeDir == "" || localDir == "" || baseDir == "" {
		panic("please call Init first")
	}
}

func Get() *Config {
	ensureInit()
	return instance
}

func getDir(dir string, names ...string) string {
	ensureInit()
	if len(names) == 0 {
		return dir
	}
	names = append([]string{dir}, names...)
	return filepath.Join(names...)
}

func GetDir(names ...string) string {
	return getDir(baseDir, names...)
}

func GetLocalDir(names ...string) string {
	return getDir(localDir, names...)
}

var (
	nowUnix int64
	nowOnce sync.Once
)

var (
	HourSeconds int64 = int64(time.Hour.Seconds())
	DaySeconds  int64 = HourSeconds * 24
	WeekSeconds int64 = DaySeconds * 7
)

func Now() int64 {
	nowOnce.Do(func() { nowUnix = time.Now().Unix() })
	return nowUnix
}
