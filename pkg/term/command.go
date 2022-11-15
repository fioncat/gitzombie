package term

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const FilePlaceholder = "{file}"

func FuzzySearch(name string, items []string) (int, error) {
	PrintSearch("use fzf to search %s", name)
	var inputBuf bytes.Buffer
	inputBuf.Grow(len(items))
	for _, item := range items {
		inputBuf.WriteString(item + "\n")
	}

	var outputBuf bytes.Buffer
	cmd := exec.Command("fzf")
	cmd.Stdin = &inputBuf
	cmd.Stderr = os.Stderr
	cmd.Stdout = &outputBuf

	err := cmd.Run()
	if err != nil {
		return 0, err
	}

	result := outputBuf.String()
	result = strings.TrimSpace(result)
	for idx, item := range items {
		if item == result {
			return idx, nil
		}
	}

	return 0, fmt.Errorf("cannot find %q", result)
}

func Exec(cmds ...string) error {
	_, err := ExecCmd(cmds, "", false)
	return err
}

func ExecOutput(cmds ...string) (string, error) {
	return ExecCmd(cmds, "", false)
}

func ExecCmd(cmds []string, in string, quiet bool) (string, error) {
	if !quiet {
		PrintCmd(cmds)
	}
	name := cmds[0]
	var args []string
	if len(cmds) > 1 {
		args = cmds[1:]
	}

	cmd := exec.Command(name, args...)
	if in != "" {
		var inBuffer bytes.Buffer
		inBuffer.WriteString(in)
		cmd.Stdin = &inBuffer
	} else {
		cmd.Stdin = os.Stdin
	}
	var outBuffer bytes.Buffer
	cmd.Stdout = &outBuffer

	var errBuffer bytes.Buffer
	if quiet {
		cmd.Stderr = &errBuffer
	} else {
		cmd.Stderr = os.Stderr
	}

	err := cmd.Run()
	if err != nil {
		return "", &CommandError{
			Cmds:   cmds,
			Err:    err,
			Output: errBuffer.String(),
		}
	}

	out := outBuffer.String()
	return strings.TrimSpace(out), nil
}

type CommandError struct {
	Cmds   []string
	Err    error
	Output string
}

func (err *CommandError) Error() string {
	return Color("failed to execute command: %v", err.Err)
}

func (err *CommandError) Extra() {
	if err.Output != "" {
		Print(err.Output)
	}
}

func PrintSearch(msg string, args ...any) {
	msg = fmt.Sprintf(msg, args...)
	Print("=> green|%s|", msg)
}

func PrintCmd(args []string) {
	msg := strings.Join(args, " ")
	Print("==> cyan|%s|", msg)
}
