package secret

import (
	"fmt"
	"os"
	"strings"

	"github.com/fioncat/gitzombie/cmd/app"
	"github.com/fioncat/gitzombie/pkg/crypto"
	"github.com/fioncat/gitzombie/pkg/osutil"
	"github.com/fioncat/gitzombie/pkg/term"
	"github.com/spf13/cobra"
)

type FileFlags struct {
	Write  bool
	Output string
}

func prepareFileCmd(cmd *cobra.Command, flags *FileFlags) {
	cmd.Flags().BoolVarP(&flags.Write, "write", "w", false, "write output to file")
	cmd.Flags().StringVarP(&flags.Output, "output", "o", "", "output file")

	cmd.Args = cobra.ExactArgs(1)
}

var Encrypt = app.Register(&app.Command[FileFlags, app.Empty]{
	Use:  "encrypt [-w] [-o output] {file}",
	Desc: "Encrypt a file use given password",

	Prepare: prepareFileCmd,

	Run: func(ctx *app.Context[FileFlags, app.Empty]) error {
		inPath := ctx.Arg(0)

		data, err := os.ReadFile(inPath)
		if err != nil {
			return err
		}

		password, err := term.InputNewPassword("Please input new password")
		if err != nil {
			return err
		}

		encrypted, _, err := crypto.Encrypt(password, data, true)
		if err != nil {
			return err
		}

		var outPath string
		if ctx.Flags.Write {
			outPath = inPath
		} else {
			outPath = ctx.Flags.Output
		}
		if outPath != "" {
			return osutil.WriteFile(outPath, []byte(encrypted))
		}

		fmt.Println(encrypted)
		return nil
	},
})

var Decrypt = app.Register(&app.Command[FileFlags, app.Empty]{
	Use:  "decrypt [-w] [-o output] {file}",
	Desc: "Decrypt a file use given password",

	Prepare: prepareFileCmd,

	Run: func(ctx *app.Context[FileFlags, app.Empty]) error {
		inPath := ctx.Arg(0)

		data, err := os.ReadFile(inPath)
		if err != nil {
			return err
		}

		password, err := term.InputPassword("Please input password")
		if err != nil {
			return err
		}

		value := strings.TrimSpace(string(data))
		raw, err := crypto.Decrypt(password, "", value)
		if err != nil {
			return err
		}

		var outPath string
		if ctx.Flags.Write {
			outPath = inPath
		} else {
			outPath = ctx.Flags.Output
		}
		if outPath != "" {
			return osutil.WriteFile(outPath, raw)
		}

		fmt.Println(string(raw))
		return nil
	},
})
