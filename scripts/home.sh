
gz() {
	local ret=$(gitzombie home repo $@)
	if [[ -d $ret ]]; then
		cd $ret
		return
	fi
}

gzs() {
	local ret=$(gitzombie home remote $@)
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
complete -F _gz gzs

alias gzb="gitzombie"