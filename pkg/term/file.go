package term

import (
	"fmt"
	"os"
)

func EnsureDir(dir string) error {
	stat, err := os.Stat(dir)
	if os.IsNotExist(err) {
		return os.MkdirAll(dir, os.ModePerm)
	}
	if err != nil {
		return err
	}
	if !stat.IsDir() {
		return fmt.Errorf("%s is not a directory", dir)
	}
	return nil
}
