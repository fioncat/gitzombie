package edit

import (
	"os"
	"path/filepath"

	"github.com/fioncat/gitzombie/config"
	"github.com/fioncat/gitzombie/pkg/errors"
	"github.com/fioncat/gitzombie/pkg/osutil"
	"github.com/fioncat/gitzombie/pkg/term"
)

func Do(path, defaultContent, name string, validate func(s string) error) error {
	var content string
	_, err := os.Stat(path)
	switch {
	case os.IsNotExist(err):
		content = defaultContent

	case err == nil:
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		content = string(data)

	default:
		return err
	}

	content, err = term.EditContent(config.Get().Editor, content, name)
	if err != nil {
		return err
	}
	err = validate(content)
	if err != nil {
		return errors.Trace(err, "validate edit content")
	}

	dir := filepath.Dir(path)
	err = osutil.EnsureDir(dir)
	if err != nil {
		return errors.Trace(err, "ensure dir")
	}

	return errors.Trace(os.WriteFile(path, []byte(content), 0644), "write file")
}
