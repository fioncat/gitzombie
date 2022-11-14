package term

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/fatih/color"
)

// colorPlaceholder is the color placeholder "color|**|" regex object.
var colorPlaceholder = regexp.MustCompile(`(cyan|blue|red|yellow|magenta|green)\|[^\|]+\|`)

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

type extraError interface {
	Extra()
}

// PrintError prints error to stderr. If error has Extra() function, it
// will be called after printing.
func PrintError(err error) {
	msg := fmt.Sprintf("red|fatal:| %v", err)
	Print(msg)
	if ext, ok := err.(extraError); ok {
		ext.Extra()
	}
}
