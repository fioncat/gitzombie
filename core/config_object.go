package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fioncat/gitzombie/config"
	"github.com/fioncat/gitzombie/pkg/errors"
	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
)

func listConfigObjectNames(dir, oext string) ([]string, error) {
	dir = config.BaseDir(dir)
	files, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, errors.Trace(err, "read %s", dir)
	}
	names := make([]string, 0, len(files))
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		name := file.Name()
		ext := filepath.Ext(name)
		if ext == oext {
			name = strings.TrimSuffix(name, ext)
			if name != "" {
				names = append(names, name)
			}
		}
	}
	return names, nil
}

func getConfigObject[T any](name string, dir, otype, oext string, validate func(val *T) error) (*T, error) {
	filename := name + oext
	path := config.BaseDir(dir, filename)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("cannot find %s %s", otype, name)
		}
		return nil, errors.Trace(err, "read %s file", otype)
	}

	var val T
	switch oext {
	case ".toml":
		err = toml.Unmarshal(data, &val)
	case ".yaml":
		err = yaml.Unmarshal(data, &val)

	default:
		panic("unknown config ext " + oext)
	}
	if err != nil {
		return nil, errors.Trace(err, "parse %s file", otype)
	}

	err = validate(&val)
	if err != nil {
		return nil, errors.Trace(err, "validate %s %s", otype, name)
	}

	return &val, nil
}
