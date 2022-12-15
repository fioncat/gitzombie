package play

import (
	"fmt"

	"github.com/fioncat/gitzombie/cmd/app"
)

var List = app.Register(&app.Command[app.Empty, app.Empty]{
	Use:    "play",
	Desc:   "List playgrounds",
	Action: "List",

	RunNoContext: func() error {
		repos, err := list()
		if err != nil {
			return err
		}
		for _, repo := range repos {
			fmt.Println(repo.Name)
		}
		return nil
	},
})
