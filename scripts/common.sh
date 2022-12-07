__gz_home() {
	ret_path=$(gitzombie $@)
	if [[ -d $ret_path ]]; then
		cd $ret_path
		return
	fi
	echo $ret_path
	return $?
}

gz() {
	action=$1
	case "${action}" in
		home|play|jump|template)
			__gz_home $@
			;;

		create)
			target=$2
			case "${target}" in
			play|repo)
				__gz_home $@
				;;
			*)
				gitzombie $@
				;;
			esac
			;;

		*)
			gitzombie $@
			;;
	esac
	return $?
}
