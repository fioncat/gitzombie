package exec

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"

	"github.com/fioncat/gitzombie/pkg/errors"
	"github.com/fioncat/gitzombie/pkg/term"
	"github.com/kballard/go-shellquote"
)

func Run(cmd string, mute bool) error {
	cmd = os.ExpandEnv(cmd)
	args, err := shellquote.Split(cmd)
	if err != nil {
		return errors.Trace(err, "shell split %q", cmd)
	}
	if len(args) == 0 {
		return fmt.Errorf("get empty string after spliting shell %q", cmd)
	}
	if !mute {
		term.PrintCmd(cmd)
	}
	name := args[0]
	args = args[1:]
	c := exec.Command(name, args...)

	var stdoutBuffer bytes.Buffer
	var stderrBuffer bytes.Buffer
	if mute {
		c.Stdout = &stdoutBuffer
		c.Stderr = &stderrBuffer
	} else {
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		c.Stdin = os.Stdin
	}

	err = c.Run()
	if err != nil {
		return &Error{
			Cmd: cmd,

			Stdout: stdoutBuffer.String(),
			Stderr: stderrBuffer.String(),

			Err: err,
		}
	}

	return nil
}
