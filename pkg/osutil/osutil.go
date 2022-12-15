package osutil

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type Env map[string]string

func (env Env) Expand(s string) string {
	return os.Expand(s, func(key string) string {
		if env == nil {
			return key
		}
		return env[key]
	})
}

func (env Env) SetCmd(cmd *exec.Cmd) {
	for key, val := range env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, val))
	}
}

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
	var err error
	if !filepath.IsAbs(path) {
		path, err = filepath.Abs(path)
		if err != nil {
			return fmt.Errorf("failed to convert abs path: %v", err)
		}
	}
	dir := filepath.Dir(path)
	// Make sure that the dir of file is created
	err = EnsureDir(dir)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
