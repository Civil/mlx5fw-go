#!/bin/bash

# Check against sample sections
declare -A mlx5fw-go_parsed

unset FW_COMPS_DEBUG
unset MFT_DEBUG

failed=0
total=0
out=""
verbose=1
t="FW Release Date"
pattern="^FW Release Date"
pr=$(echo "${pattern}" | tr -cd ' \t' | wc -c)
pr=$((pr+2))
integration=1

list=""
if [[ ${integration} -ne 0 ]]; then
	verbose=0
	declare -a files_list
	mapfile -d $'\0' -t files_list < <(find /home/civil/fws/ -name '*.xz' -print0)

	for i in {1..20}; do
		randfile=$(python3 -S -c "import random; print(random.randrange(0,${#files_list[@]}))")
		list="${list} ${files_list[$randfile]}"
	done
fi

for c_f in sample_firmwares/*.bin ${list}; do
	f="${c_f}"
	if [[ ${c_f} == *.xz ]]; then
		xz -kd "${c_f}"
		f="${c_f/.xz/}"
	fi
	total=$((total+1))
	mlx5fw-go_ver=$(./mlx5fw-go -metadata "${f}" | grep "${pattern}" | awk -vpr=${pr} '{print $pr}')
	if [[ $? -ne 0 ]]; then
		failed=$((failed+1))
		grep -q 'panic' <<< "${mlx5fw-go_out}"
		if [[ $? -ne 0 ]]; then
			out="${out}$(echo -e '\n')$(echo "${mlx5fw-go_out}" | awk 'BEGIN{pr=0}{if ($0 ~ /panic/){pr=1;}; if (pr == 1) { print $0; };}')"
		fi
		continue
	fi

	mft_ver=$(mstflint -i "${f}" query full | grep "${pattern}" | awk -vpr=${pr} '{print $pr}')
	if [[ "${mlx5fw-go_ver}" != "${mft_ver}" ]]; then
		if [[ ${verbose} -ne 0 ]]; then
			echo "${t} mismatch for ${f}. Got '${mlx5fw-go_ver}', expected '${mft_ver}'"
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
