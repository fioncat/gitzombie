
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

gzs() {
	local ret=$(gitzombie home remote $@)
	if [[ -d $ret ]]; then
		cd $ret
		return
	fi
}

_gzs() {
	if [[ "${COMP_CWORD}" = "1" ]]; then
		# Complete remote
		gitzombie list repo
		return
	fi
	if [[ "${COMP_CWORD}" = "2" ]]; then
		# Complete repo
		local remote="${COMP_WORDS[1]}"
		gitzombie list repo $remote --group
		return
	fi
	COMPREPLY=()
}

complete -F _gz gz
complete -F _gzs gzs

alias gzb="gitzombie"
alias gzo="gitzombie open repo"
