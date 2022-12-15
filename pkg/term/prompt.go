package term

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"syscall"

	"github.com/jedib0t/go-pretty/v6/text"
	"golang.org/x/term"
)

var (
	AlwaysYes bool
)

func Confirm(msg string, args ...any) bool {
	if AlwaysYes {
		return true
	}
	msg = fmt.Sprintf(msg, args...)
	msg = fmt.Sprintf("%s? (y/n) ", msg)
	fmt.Fprint(os.Stderr, msg)
	var input string
	fmt.Scanf("%s", &input)
	return input == "y"
}

func ConfirmExit(msg string, args ...any) {
	if !Confirm(msg, args...) {
		os.Exit(2)
	}
}

func Open(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}

func InputPassword(msg string, args ...any) (string, error) {
	msg = fmt.Sprintf(msg, args...)
	fmt.Fprintf(os.Stderr, "%s: ", msg)
	bytesPassword, err := term.ReadPassword(syscall.Stdin)
	if err != nil {
		return "", fmt.Errorf("failed to read password: %v", err)
	}
	Println()
	return string(bytesPassword), nil
}

func InputNewPassword(msg string, args ...any) (string, error) {
	password, err := InputPassword(msg, args...)
	if err != nil {
		return "", err
	}

	rePassword, err := InputPassword("Please confirm password")
	if err != nil {
		return "", err
	}

	if password != rePassword {
		return "", errors.New("the two passwords you entered were inconsistent")
	}
	return password, nil
}

func InputErase(msg string, args ...any) string {
	msg = fmt.Sprintf(msg, args...)
	fmt.Fprintf(os.Stderr, "%s: ", msg)
	var input string
	fmt.Scanf("%s", &input)
	fmt.Fprint(os.Stderr, text.CursorUp.Sprint())
	fmt.Fprint(os.Stderr, text.EraseLine.Sprint())
	fmt.Fprintln(os.Stderr, msg)
	return input
}
