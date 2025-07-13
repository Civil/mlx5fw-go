#!/bin/bash

declare -A mlx5fw_go_parsed

unset FW_COMPS_DEBUG
unset MFT_DEBUG
mlx5fw="mlx5fw-go"
mstflint="mstflint"

verbose=1

function find_firmwares() {
	echo "sample_firmwares/broken_fw.bin"
	return
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

function do_test() {
	failed=0
	total=0
	t=${1}
	pattern=${2}
	verbose=${3}
	pr=""
	if [[ ! -z "${pattern}" ]]; then
		pr=$(echo "${pattern}" | tr -cd ' \t' | wc -c)
		pr=$((pr+2))
	fi

	for c_f in ${list}; do
		f="${c_f}"
		if [[ ${c_f} == *.xz ]]; then
			xz -kd "${c_f}"
			f="${c_f/.xz/}"
		fi
		total=$((total+1))
		mlx5fw_go_out=$(./${mlx5fw} query -f "${f}")
		if [[ $? -ne 0 ]]; then
			${mstflint} -i "${f}" query full &>/dev/null
			if [[ $? -ne 0 ]]; then
				[[ ${verbose} -ne 0 ]] && echo "both mlx5fw and mstflint fails"
				continue
			fi
			failed=$((failed+30))
			grep -q 'panic' <<< "${mlx5fw_go_out}"
			if [[ $? -ne 0 ]]; then
				out="${out}$(echo -e '\n')$(echo "${mlx5fw_go_out}" | awk 'BEGIN{pr=0}{if ($0 ~ /panic/){pr=1;}; if (pr == 1) { print $0; };}')"
			fi
			continue
		fi
		mlx5fw_go_ver=$(grep "${pattern}" <<< "${mlx5fw_go_out}" | head -n 1 | awk -vpr=${pr} '{for (i=1; i < pr; i++) { $i=""; }; print $0}' | sed 's/^ \+//g')
	
		mft_ver=$(${mstflint} -i "${f}" query full | grep "${pattern}" | head -n 1 | awk -vpr=${pr} '{for (i=1; i < pr; i++) { $i=""; }; print $0}' | sed 's/^ \+//g')
		if [[ "${mlx5fw_go_ver}" != "${mft_ver}" ]]; then
			if [[ ${verbose} -ne 0 ]]; then
				echo "${t} mismatch for ${f}. Got '${mlx5fw_go_ver}', expected '${mft_ver}'"
			fi
			failed=$((failed+1))
		fi
	
		if [[ ${c_f} == *.xz ]]; then
			rm -f "${f}"
		fi
	done
	
	score=$(echo | awk -vtotal=${total} -vfailed=${failed} 'END{print (total+0.0-failed)/total}')
	echo "${score}"
	if [[ ! -z ${out} ]]; then
		echo "${out}"
	fi
	
	exit 0
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
