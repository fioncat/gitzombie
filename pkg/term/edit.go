package term

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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

func EditItems[T any](editor string, items []*T, getKey func(*T) string) ([]*T, error) {
	itemMap := make(map[string]*T, len(items))
	lines := make([]string, len(items))
	for i, item := range items {
		key := getKey(item)
		lines[i] = key
		itemMap[key] = item
	}
	content := strings.Join(lines, "\n")
	content, err := EditContent(editor, content, "items.txt")
	if err != nil {
		return nil, err
	}
	lines = strings.Split(content, "\n")

	editedItems := make([]*T, 0, len(lines))
	set := make(map[string]struct{}, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		name := line
		item := itemMap[name]
		if item == nil {
			return nil, fmt.Errorf("edit: cannot find item %q", name)
		}
		if _, ok := set[name]; ok {
			continue
		}
		editedItems = append(editedItems, item)
		set[name] = struct{}{}
	}
	if len(editedItems) == 0 {
		return nil, errors.New("nothing to do after editing")
	}
	return editedItems, nil
}
