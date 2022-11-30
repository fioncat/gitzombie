package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/fatih/color"
	"github.com/fioncat/gitzombie/cmd"
	"github.com/fioncat/gitzombie/pkg/term"
	"github.com/spf13/cobra"
)

var (
	Version   string
	Commit    string
	BuildDate string
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show full version, include build info",

	Run: func(_ *cobra.Command, _ []string) {
		fmt.Printf("Version:   %s\n", Version)
		fmt.Printf("Commit:    %s\n", Commit)
		fmt.Printf("BuildTime: %s\n", BuildDate)
		fmt.Printf("Platform:  %s\n", fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH))
		fmt.Printf("GoVersion: %s\n", runtime.Version())
	},
}

func main() {
	color.NoColor = false
	if os.Getenv("GITZOMBIE_NO_COLOR") != "" {
		color.NoColor = true
	}
	cmd.Root.Version = Version
	cmd.Root.AddCommand(versionCmd)
	err := cmd.Root.Execute()
	if err != nil {
		term.PrintError(err)
		os.Exit(1)
	}
}
