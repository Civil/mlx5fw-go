#!/bin/bash

declare -A mlx5fw_go_parsed
failed=0
total=0
out=""
integration=0

. $(dirname ${0})/lib.sh

verbose=1

for c_f in ${list}; do
	grep -F -q "*" <<< ${c_f} && continue
	f="${c_f}"
	if [[ ${c_f} == *.xz ]]; then
		xz -kd "${c_f}"
		f="${c_f/.xz/}"
	fi
	
	unset mlx5fw_go_parsed
	declare -A mlx5fw_go_parsed

	[[ ${verbose} == 1 ]] && echo -e "\n\n\ntest report for ${f}"
	out_dir=$(mktemp -d)
	n_f="${out_dir}/$(basename "${f}")"
	mlx5fw_go_out=$(./${mlx5fw} extract -f "${f}" -o "${out_dir}/extract")
	if [[ $? -ne 0 ]]; then
		total=$((total+30))
		failed=$((failed+30))
		echo "  FATAL(${f}): mlx5fw-go failed to extract ${f}:"
		echo "${mlx5fw_go_out}"
		echo
		rm -rf "${out_dir}"
		continue
	fi
	
	mlx5fw_go_out=$(./${mlx5fw} reassemble -i "${out_dir}/extract" -o "${n_f}")
	if [[ $? -ne 0 ]]; then
		total=$((total+30))
		failed=$((failed+30))
		echo "  FATAL(${f}): mlx5fw-go failed to reassemble ${f}:"
		echo "${mlx5fw_go_out}"
		echo
		rm -rf "${out_dir}"
		continue
	fi

	# Do exact same comparison as for sections
	mlx5fw_go_out=$(./${mlx5fw} sections -f "${f}" 2>&1)
	if [[ $? -ne 0 ]]; then
		mstflint -i "${f}" v &>/dev/null
		if [[ $? -ne 0 ]]; then
			total=$((total+1))
			[[ ${verbose} == 1 ]] && echo "OK: both mstflint and mlx5fw-go refused to parse file"
			continue
		fi
		total=$((total+30))
		failed=$((failed+30))
		echo "  FATAL(${f}): mlx5fw-go failed to run for ${f}:"
		echo "${mlx5fw_go_out}"
		echo
		continue
	fi

	OLD_IFS="${IFS}"
	IFS=$'\n'
	for l in ${mlx5fw_go_out}; do
		IFS="${OLD_IFS}"
		set -- ${l}
		addr=$(echo "${1}" | cut -d'/' -f 2 | cut -d'-' -f 1)
		name=$(echo "${3}" | cut -d'(' -f 2 | cut -d')' -f 1)
		# addr="${2}"
		# name="${1}"
		addrs="${mlx5fw_go_parsed[${name}]}"
		mlx5fw_go_parsed[${name}]="${addrs} ${addr}"
	done
	IFS="${OLD_IFS}"

	mft_out=$(${mstflint} -i "${f}" v | head -n -3 | tail -n +4)

	OLD_IFS="${IFS}"
	IFS=$'\n'
	for l in ${mft_out}; do
		IFS="${OLD_IFS}"
		set -- ${l}
		addr=$(echo "${1}" | cut -d'/' -f 2 | cut -d'-' -f 1)
		name=$(echo "${3}" | cut -d'(' -f 2 | cut -d')' -f 1)
		mlx5fw_go_addrs="${mlx5fw_go_parsed[${name}]}"
		total=$((total+1))
		if [[ -z "${mlx5fw_go_addrs}" ]]; then
			failed=$((failed+1))
			[[ ${verbose} == 1 ]] && echo "  ERROR(${f}): mlx5fw-go output doesn't have section '${name}'" ||:
			continue
		else
			[[ ${verbose} == 1 ]] && echo "  OK(${f}): section '${name}' is present"
		fi

		total=$((total+1))
		grep -q "${addr}" <<< "${mlx5fw_go_addrs}"
		if [[ ${?} -ne 0 ]]; then
			failed=$((failed+1))
			[[ ${verbose} == 1 ]] && echo "  ERROR(${f}: mlx5fw-go output for section ${name} have a wrong address/adresses. got '${mlx5fw_go_parsed[${name}]}', expected '${addr}'"
		else
			[[ ${verbose} == 1 ]] && echo "  OK(${f}): Address for section '${name}' present"
		fi
	done
	IFS="${OLD_IFS}"
	rm -rf "${out_dir}"
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
