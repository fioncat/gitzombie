package delete

import (
	"fmt"
	"os"

	"github.com/fioncat/gitzombie/pkg/osutil"
	"github.com/fioncat/gitzombie/pkg/term"
)

func do(name, path string) error {
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
