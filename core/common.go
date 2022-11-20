package core

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/fioncat/gitzombie/config"
	"github.com/fioncat/gitzombie/pkg/errors"
	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
)

const (
	tomlExt = ".toml"
	yamlExt = ".yaml"
	shExt   = ".sh"
)

func listConfigObjects(configType, configExt string) ([]string, error) {
	rootDir := config.GetDir(configType)
	var names []string
	err := filepath.Walk(rootDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		ext := filepath.Ext(path)
		if ext != configExt {
			return nil
		}

		name, err := filepath.Rel(rootDir, path)
		if err != nil {
			return errors.Trace(err, "get rel path for config object %s", path)
		}
		name = strings.TrimSuffix(name, ext)
		names = append(names, name)
		return nil
	})
	return names, err
}

func getConfigObjectPath(configType, configExt string, name string) string {
	rootDir := config.GetDir(configType)
	name = name + configExt
	return filepath.Join(rootDir, name)
}

func getConfigObject[T any](configType, configExt, configName string, name string, validate func(*T) error) (*T, error) {
	path := getConfigObjectPath(configType, configExt, name)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("cannot find %s %s", configName, name)
		}
		return nil, err
	}

	var obj T
	switch configExt {
	case yamlExt:
		err = yaml.Unmarshal(data, &obj)
		if err != nil {
			return nil, errors.Trace(err, "parse yaml")
		}

	case tomlExt:
		err = toml.Unmarshal(data, &obj)
		if err != nil {
			return nil, errors.Trace(err, "parse toml")
		}

	default:
		panic("getConfigObject: invalid ext: " + configExt)
	}
	if validate != nil {
		err = validate(&obj)
		if err != nil {
			return nil, errors.Trace(err, "validate")
		}
	}
	return &obj, nil
}
