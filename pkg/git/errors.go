package git

import (
	"fmt"

	"github.com/dustin/go-humanize/english"
	"github.com/fioncat/gitzombie/pkg/term"
)

type ExecError struct {
	Stderr string

	Err error

	Cmd string

	Opts *Options
}

func (e *ExecError) Error() string {
	if e.Opts.QuietCmd {
		return fmt.Sprintf("failed to exec %q: %v", e.Cmd, e.Err)
	}
	return fmt.Sprintf("failed to exec git command: %v", e.Err)
}

func (e *ExecError) Out() string {
	return e.Stderr
}

func (e *ExecError) Extra() {
	if e.Stderr != "" {
		term.Println(e.Stderr)
	}
}

type UncommittedChangeError struct {
	changes []string
}

func (e *UncommittedChangeError) Error() string {
	changeWord := english.Plural(len(e.changes), "uncommitted change", "")
	var who string
	if len(e.changes) == 1 {
		who = "it"
	} else {
		who = "them"
	}
	return fmt.Sprintf("you have %s in current repo, please handle %s first",
		changeWord, who)
}
