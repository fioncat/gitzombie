package term

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/fatih/color"
	"github.com/fioncat/gitzombie/pkg/errors"
)

// colorPlaceholder is the color placeholder "color|**|" regex object.
var colorPlaceholder = regexp.MustCompile(`(cyan|blue|red|yellow|magenta|green)\|[^\|]+\|`)

func Color(s string, args ...any) string {
	msg := fmt.Sprintf(s, args...)
	return colorString(msg)
}

func colorString(s string) string {
	return colorPlaceholder.ReplaceAllStringFunc(s, func(ph string) string {
		tmp := strings.Split(ph, "|")
		if len(tmp) <= 1 {
			return ph
		}
		colorType := tmp[0]
		content := tmp[1]

		switch colorType {
		case "cyan":
			return color.CyanString(content)

		case "blue":
			return color.BlueString(content)

		case "red":
			return color.RedString(content)

		case "yellow":
			return color.YellowString(content)

		case "magenta":
			return color.MagentaString(content)

		case "green":
			return color.GreenString(content)
		}
		return ph
	})
}

// Print prints message to stderr. The color placeholder "color|**|"
// will be expanded.
func Print(msg string, args ...any) {
	msg = fmt.Sprintf(msg, args...)
	fmt.Fprintln(os.Stderr, colorString(msg))
}

func Warn(msg string, args ...any) {
	msg = fmt.Sprintf(msg, args...)
	msg = fmt.Sprintf("yellow|WARN: %s|", msg)
	Print(msg)
}

func PrintError(err error) {
	msg := fmt.Sprintf("red|fatal:| %v", err)
	Print(msg)
	if ext, ok := err.(errors.Extra); ok {
		ext.Extra()
	}
}
