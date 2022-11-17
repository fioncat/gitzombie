gz() {
	local ret=$(gitzombie home repo $@)
	if [[ -d $ret ]]; then
		cd $ret
		return
	fi
}

_gz() {
	if [[ "${COMP_CWORD}" = "1" ]]; then
		# Complete remote
		gitzombie list repo
		return
	fi
	if [[ "${COMP_CWORD}" = "2" ]]; then
		# Complete repo
		local remote="${COMP_WORDS[1]}"
		gitzombie list repo $remote
		return
	fi
	COMPREPLY=()
}

complete -F _gz gz
alias gzh="gz github"

alias gzb="gitzombie"
alias gzo="gitzombie open repo"
alias gzk="gitzombie switch"
alias gzs="gitzombie sync branch"
alias gze="gitzombie edit"
