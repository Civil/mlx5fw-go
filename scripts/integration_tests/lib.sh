#!/bin/bash

declare -A mlx5fw-go_parsed

unset FW_COMPS_DEBUG
unset MFT_DEBUG

failed=0
total=0
verbose=1
pr=""
if [[ ! -z "${pattern}" ]]; then
	pr=$(echo "${pattern}" | tr -cd ' \t' | wc -c)
	pr=$((pr+2))
fi

function find_firmwares() {
	declare -A files_list
	for f in $(find -L -not -path './reference/*' -name '*.bin'); do
		f_b=$(basename ${f})
		if [[ ${files_list[${f_b}]} == 1 ]]; then
			continue
		else
			files_list[${f_b}]=1
			echo "${f}"
		fi
	done
}

list="$(find_firmwares)"
if [[ ${integration} -ne 0 ]]; then
        verbose=0
        declare -a files_list
        mapfile -d $'\0' -t files_list < <(find /home/civil/fws/ -name '*.xz' -print0)

        for i in {1..20}; do
                randfile=$(python3 -S -c "import random; print(random.randrange(0,${#files_list[@]}))")
                list="${list} ${files_list[$randfile]}"
        done
fi
