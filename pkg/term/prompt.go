package term

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

var (
	AlwaysYes bool
)

func Confirm(msg string, args ...any) bool {
	if AlwaysYes {
		return true
	}
	msg = Color(msg, args...)
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
