#!/bin/bash

declare -A mlx5fw-go_parsed
failed=0
total=0
out=""
t="FW Release Date"
pattern="^FW Release Date"
integration=0

. $(dirname ${0})/lib.sh

for c_f in ${list}; do
	f="${c_f}"
	if [[ ${c_f} == *.xz ]]; then
		xz -kd "${c_f}"
		f="${c_f/.xz/}"
	fi
	total=$((total+1))
	mlx5fw-go_ver=$(./mlx5fw-go query -f "${f}" | grep "${pattern}" | awk -vpr=${pr} '{print $pr}')
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
