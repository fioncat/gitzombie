zb() {
	zb_out=$(gitzombie $@)
	if [[ ! -z $zb_out ]]; then
		if [[ -d $zb_out ]]; then
			cd $zb_out
			return
		fi
		echo $zb_out
	fi
}
