package term

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/fioncat/gitzombie/pkg/errors"
)

func Style(a any, style string) string {
	styleAttrStrs := strings.Split(style, ",")
	attrs := make([]color.Attribute, len(styleAttrStrs))
	for i, styleAttrStr := range styleAttrStrs {
		var attr color.Attribute
		switch styleAttrStr {
		case "red":
			attr = color.FgRed

		case "green":
			attr = color.FgGreen

		case "magenta":
			attr = color.FgMagenta

		case "yellow":
			attr = color.FgYellow

		case "bold":
			attr = color.Bold

		case "blue":
			attr = color.FgBlue

		default:
			panic("unknown style attr " + styleAttrStr)
		}
		attrs[i] = attr
	}
	c := color.New(attrs...)
	return c.Sprint(a)
}

func Printf(msg string, args ...any) {
	msg = fmt.Sprintf(msg, args...)
	fmt.Fprintln(os.Stderr, msg)
}

func Println(args ...any) {
	fields := make([]string, len(args))
	for i, arg := range args {
		fields[i] = fmt.Sprint(arg)
	}
	msg := strings.Join(fields, " ")
	fmt.Fprintln(os.Stderr, msg)
}

func Warn(msg string, args ...any) {
	msg = fmt.Sprintf(msg, args...)
	Println(Style("WARN:", "yellow,bold"), msg)
}

func PrintError(err error) {
	Println(Style("error:", "red,bold"), err)
	if ext, ok := err.(errors.Extra); ok {
		ext.Extra()
	}
}

func PrintOperation(msg string, args ...any) {
	msg = fmt.Sprintf(msg, args...)
	Println(Style("==>", "blue"), Style(msg, "bold"))
}

func PrintCmd(cmd string, args ...any) {
	cmd = fmt.Sprintf(cmd, args...)
	Println(Style("==>", "green"), Style(cmd, "bold"))
}
