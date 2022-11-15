package main

import (
	"os"

	"github.com/fatih/color"
	"github.com/fioncat/gitzombie/cmd"
	"github.com/fioncat/gitzombie/pkg/term"
)

var Version string

func main() {
	color.NoColor = false
	if os.Getenv("GITZOMBIE_NO_COLOR") != "" {
		color.NoColor = true
	}
	cmd.Root.Version = Version
	err := cmd.Root.Execute()
	if err != nil {
		term.PrintError(err)
		os.Exit(1)
	}
}
