package osutil

import (
	"fmt"
	"os"
	"path/filepath"
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

func FileExists(dir string) (bool, error) {
	stat, err := os.Stat(dir)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if stat.IsDir() {
		return false, fmt.Errorf("%s is not a file", dir)
	}
	return true, nil
}

func WriteFile(path string, data []byte) error {
	if !filepath.IsAbs(path) {
		// The caller should make sure that path is abs, here we
		// make a double-check.
		return fmt.Errorf("failed to write to %s: path is not abs", path)
	}
	dir := filepath.Dir(path)
	// Make sure that the dir of file is created
	err := EnsureDir(dir)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
