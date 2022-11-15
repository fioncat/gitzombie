package main

import (
	"os"

	"github.com/fioncat/gitzombie/cmd"
	"github.com/fioncat/gitzombie/pkg/term"
)

func main() {
	err := cmd.Root.Execute()
	if err != nil {
		term.PrintError(err)
		os.Exit(1)
	}
}
