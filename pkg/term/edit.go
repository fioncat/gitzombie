package term

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/fioncat/gitzombie/pkg/errors"
	"github.com/fioncat/gitzombie/pkg/osutil"
)

func EditContent(editor, content string, name string) (string, error) {
	tmpDir := os.TempDir()
	tmpDir = filepath.Join(tmpDir, "gitzombie")
	err := osutil.EnsureDir(tmpDir)
	if err != nil {
		return "", errors.Trace(err, "ensure tmp dir")
	}

	path := filepath.Join(tmpDir, name)
	err = osutil.WriteFile(path, []byte(content))
	if err != nil {
		return "", errors.Trace(err, "write tmp file to edit")
	}

	err = Edit(editor, path)
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", errors.Trace(err, "read edited tmp file")
	}

	err = os.Remove(path)
	if err != nil {
		Warn("failed to remove edited tmp file: %v", err)
	}

	return string(data), nil
}

func Edit(editor, path string) error {
	cmd := exec.Command(editor, path)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("edit failed: %v", err)
	}
	return nil
}
