package git

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fioncat/gitzombie/pkg/errors"
	"github.com/fioncat/gitzombie/pkg/term"
)

type Options struct {
	QuietCmd bool

	QuietStderr bool

	Path string

	NoTrimLines bool
}

var (
	Default = &Options{}

	QuietOutput = &Options{QuietStderr: true}

	Mute = &Options{
		QuietStderr: true,
		QuietCmd:    true,
	}
)

func Output(args []string, opts *Options) (string, error) {
	if opts.Path != "" {
		args = append([]string{"-C", opts.Path}, args...)
	}

	cmdStr := fmt.Sprintf("git %s", strings.Join(args, " "))
	if !opts.QuietCmd {
		term.PrintCmd(cmdStr)
	}

	cmd := exec.Command("git", args...)

	var stderrOut bytes.Buffer
	if opts.QuietStderr {
		cmd.Stderr = &stderrOut
	} else {
		cmd.Stderr = os.Stderr
	}

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stdin = os.Stdin

	err := cmd.Run()
	if err != nil {
		return "", &ExecError{
			Stderr: stderrOut.String(),

			Err: err,

			Cmd: cmdStr,

			Opts: opts,
		}
	}

	outStr := out.String()
	return strings.TrimSpace(outStr), nil
}

func Exec(args []string, opts *Options) error {
	_, err := Output(args, opts)
	return err
}

func OutputItems(args []string, opts *Options) ([]string, error) {
	out, err := Output(args, opts)
	if err != nil {
		return nil, err
	}
	rawLines := strings.Split(out, "\n")
	if opts.NoTrimLines {
		return rawLines, nil
	}
	lines := make([]string, 0, len(rawLines))
	for _, line := range rawLines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		lines = append(lines, line)
	}
	return lines, nil
}

func Clone(url, path string, opts *Options) error {
	return Exec([]string{"clone", url, path}, opts)
}

func SetRemoteURL(remote, url string, opts *Options) error {
	return Exec([]string{"remote", "set-url", remote, url}, opts)
}

func Config(name, value string, opts *Options) error {
	return Exec([]string{"config", name, value}, opts)
}

func ListRemotes(opts *Options) ([]string, error) {
	return OutputItems([]string{"remote"}, opts)
}

func Checkout(name string, create bool, opts *Options) error {
	var args = []string{"checkout"}
	if create {
		args = append(args, "-b")
	}
	args = append(args, name)
	return Exec(args, opts)
}

func Switch(local, target string, opts *Options) error {
	return Exec([]string{"switch", "-c", local, target}, opts)
}

func EnsureNoUncommitted(opts *Options) error {
	changes, err := OutputItems([]string{"status", "-s"}, opts)
	if err != nil {
		return err
	}
	if len(changes) > 0 {
		return &UncommittedChangeError{
			changes: changes,
		}
	}
	return nil
}

func Fetch(remote string, branch, tag bool, opts *Options) error {
	args := []string{"fetch", remote}
	if branch {
		args = append(args, "--prune")
	}
	if tag {
		args = append(args, "--prune-tags")
	}
	return Exec(args, opts)
}

func GetMainBranch(remote string, opts *Options) (string, error) {
	ref := fmt.Sprintf("refs/remotes/%s/", remote)
	headRef := filepath.Join(ref, "HEAD")

	out, err := Output([]string{"symbolic-ref", headRef}, opts)
	if err != nil {
		return "", err
	}

	if out == "" {
		return "", errors.New("main branch is empty")
	}
	if !strings.HasPrefix(out, ref) {
		return "", fmt.Errorf("invalid ref %q", out)
	}
	main := strings.TrimPrefix(out, ref)
	return main, nil
}

func GetCurrentBranch(opts *Options) (string, error) {
	return Output([]string{"branch", "--show-current"}, opts)
}

func CreateTag(tag string, opts *Options) error {
	return Exec([]string{"tag", tag}, opts)
}

func DeleteTag(tag string, opts *Options) error {
	return Exec([]string{"tag", "-d", tag}, opts)
}

func PushTag(tag string, remote string, remove bool, opts *Options) error {
	if remove {
		return Exec([]string{"push", "--delete", remote, tag}, opts)
	}
	return Exec([]string{"push", remote, tag}, opts)
}

func ListTags(opts *Options) ([]string, error) {
	return OutputItems([]string{"tag"}, opts)
}
