#!/bin/bash

declare -A mlx5fw_go_parsed
failed=0
total=0
out=""
t="PSID"
pattern="^${t}"
integration=1

. $(dirname ${0})/lib.sh

verbose=1

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
	mft_out=$(${mstflint} -i "${f}" query full | grep "${pattern}")

	mft_out_f="${f/.bin/.mstflint.out}"
	mlx5fw_go_f="${f/.bin/.mlx5fw_go.out}"

	echo "${mft_out}" > "${mft_out_f}"
	echo "${mlx5fw_go_out}" > "${mlx5fw_go_f}"

	echo
	echo
	echo "Results for ${f}:"
	diff -u "${mft_out_f}" "${mlx5fw_go_f}"
	if [[ $? -ne 0 ]]; then
		failed=$((failed+1))
		[[ ${verbose} ]] && echo "  ERROR(${f}): There is a diff between mstflint(-) and mlx5fw(+) go"
	else
		echo "  OK(${f}): There is no diff between mstflint and mlx5fw"
	fi
	rm -f "${mft_out_f}"
	rm -f "${mlx5fw_go_f}"

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
