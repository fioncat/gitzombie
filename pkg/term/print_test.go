package term

import (
	"fmt"
	"testing"
)

func TestColor(t *testing.T) {
	fmt.Println(colorString("This is a red|red| color"))
	fmt.Println(colorString("yellow|hello, world| middle blue|some message|"))
	fmt.Println(colorString("green|[info]| A message to show for magenta|user|"))
}
