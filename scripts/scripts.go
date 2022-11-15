package scripts

import _ "embed"

//go:embed zsh-comp.zsh
var ZshComp string

//go:embed bash-comp.sh
var BashComp string

//go:embed home.sh
var Home string
