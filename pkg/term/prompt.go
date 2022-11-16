package term

import (
	"fmt"
	"os"
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
