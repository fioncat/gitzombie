package osutil

import (
	"fmt"
	"os"
)

func EnsureDir(dir string) error {
	exists, err := DirExists(dir)
	if err != nil {
		return err
	}
	if !exists {
		return os.MkdirAll(dir, os.ModePerm)
	}
	return nil
}

func DirExists(dir string) (bool, error) {
	stat, err := os.Stat(dir)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if !stat.IsDir() {
		return false, fmt.Errorf("%s is not a directory", dir)
	}
	return true, nil
}

func Setenv(env map[string]string) error {
	for key, val := range env {
		err := os.Setenv(key, val)
		if err != nil {
			return err
		}
	}
	return nil
}
