package term

import (
	"fmt"
	"os"

	"github.com/manifoldco/promptui"
)

var (
	AlwaysYes bool
)

func Confirm(msg string, args ...any) bool {
	if AlwaysYes {
		return true
	}
	msg = fmt.Sprintf(msg, args...)
	p := &promptui.Prompt{
		Label: msg,

		IsConfirm: true,
		Stdout:    os.Stderr,
	}
	_, err := p.Run()
	return err == nil
}

func ConfirmExit(msg string, args ...any) {
	if !Confirm(msg, args...) {
		os.Exit(2)
	}
}
