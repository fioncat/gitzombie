package app

import (
	"fmt"
	"os"

	"github.com/fioncat/gitzombie/config"
	"github.com/fioncat/gitzombie/pkg/errors"
	"github.com/fioncat/gitzombie/pkg/osutil"
	"github.com/fioncat/gitzombie/pkg/term"
)

func Delete(name, path string) error {
	exists, err := osutil.FileExists(path)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("cannot find %s", name)
	}
	term.ConfirmExit("Do you want to delete %s", path)
	return os.Remove(path)
}

func Edit(path, defaultContent, name string, validate func(s string) error) error {
	var content string
	_, err := os.Stat(path)
	switch {
	case os.IsNotExist(err):
		content = defaultContent

	case err == nil:
		var data []byte
		data, err = os.ReadFile(path)
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
	if validate != nil {
		err = validate(content)
		if err != nil {
			return errors.Trace(err, "validate edit content")
		}
	}

	return errors.Trace(osutil.WriteFile(path, []byte(content)), "write file")
}
