gz() {
	action=$1
	case "${action}" in
		home|play|jump)
			ret_path=$(gitzombie $@)
			if [[ -d $ret_path ]]; then
				cd $ret_path
				return
			fi
			;;
		*)
			gitzombie $@
			;;
	esac
	return $?
}
