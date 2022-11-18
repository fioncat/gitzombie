package del

import (
	"fmt"
	"os"
	"strings"

	"github.com/fioncat/gitzombie/cmd/app"
	"github.com/fioncat/gitzombie/config"
	"github.com/fioncat/gitzombie/pkg/osutil"
	"github.com/fioncat/gitzombie/pkg/term"
)

func Run(name, path string) error {
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

func comp(name, ext string) app.CompAction {
	return func(_ []string) (*app.CompResult, error) {
		dir := config.BaseDir(name)
		files, err := os.ReadDir(dir)
		if err != nil {
			return nil, err
		}
		var items []string
		for _, file := range files {
			if file.IsDir() {
				continue
			}
			if strings.HasSuffix(file.Name(), ext) {
				item := strings.TrimSuffix(file.Name(), ext)
				items = append(items, item)
			}
		}
		return &app.CompResult{Items: items}, nil
	}
}
