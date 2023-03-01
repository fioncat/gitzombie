package osutil

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fioncat/gitzombie/pkg/errors"
)

func ListEmptyDir(dir string, exclude []string) ([]string, error) {
	stat, err := os.Stat(dir)
	if err != nil {
		return nil, err
	}
	if !stat.IsDir() {
		return nil, fmt.Errorf("clean: %s is not a directory", dir)
	}

	root, err := cleanScan(dir, exclude)
	if err != nil {
		return nil, err
	}

	cleanMark(root)

	var emptyDirs []string
	cleanList(root, &emptyDirs)
	return emptyDirs, nil
}

type cleanItem struct {
	path string
	mark bool

	subs []*cleanItem

	forceNoMark bool
}

func cleanScan(path string, exclude []string) (*cleanItem, error) {
	excludeMap := make(map[string]struct{}, len(exclude))
	for _, dir := range exclude {
		excludeMap[dir] = struct{}{}
	}

	root := &cleanItem{
		path: path,
	}
	stack := []*cleanItem{root}

	for len(stack) > 0 {
		cur := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		subEntries, err := os.ReadDir(cur.path)
		if err != nil {
			return nil, errors.Trace(err, "read clean dir")
		}

		if len(subEntries) == 0 {
			cur.mark = true
			continue
		}

		forceNoMark := false
		for _, subEntry := range subEntries {
			if !subEntry.IsDir() {
				forceNoMark = true
				continue
			}
			subPath := filepath.Join(cur.path, subEntry.Name())
			if _, ok := excludeMap[subPath]; ok {
				forceNoMark = true
				continue
			}
			subItem := &cleanItem{path: subPath}
			cur.subs = append(cur.subs, subItem)
			stack = append(stack, subItem)
		}
		cur.forceNoMark = forceNoMark
	}

	return root, nil
}

func cleanMark(item *cleanItem) {
	if len(item.subs) == 0 {
		return
	}
	mark := true
	for _, sub := range item.subs {
		cleanMark(sub)
		if !sub.mark {
			mark = false
		}
	}
	if !item.forceNoMark {
		item.mark = mark
	}
}

func cleanList(item *cleanItem, toRemove *[]string) {
	if item.mark {
		*toRemove = append(*toRemove, item.path)
		return
	}
	for _, sub := range item.subs {
		cleanList(sub, toRemove)
	}
}
