gz() {
	gz_out=$(gitzombie $@)
	ret_code=$?
	if [[ ! -z $gz_out ]]; then
		if [[ -d $gz_out ]]; then
			cd $gz_out
			return
		fi
		echo $gz_out
	fi
	return $ret_code
}
