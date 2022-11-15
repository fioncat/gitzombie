package main

import (
	"os"

	"github.com/fatih/color"
	"github.com/fioncat/gitzombie/cmd"
	"github.com/fioncat/gitzombie/pkg/term"
)

func main() {
	color.NoColor = false
	err := cmd.Root.Execute()
	if err != nil {
		term.PrintError(err)
		os.Exit(1)
	}
}
