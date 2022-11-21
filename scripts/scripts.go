package scripts

import _ "embed"

//go:embed zsh-comp.zsh
var zshComp string

//go:embed bash-comp.sh
var bashComp string

//go:embed common.sh
var Common string

//go:embed alias.sh
var Alias string

var Comps = map[string]string{
	"zsh":  zshComp,
	"bash": bashComp,
}
