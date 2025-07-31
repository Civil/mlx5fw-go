#!/bin/bash

extra_verbose=0
if [[ ${1} == "-v" ]]; then
	extra_verbose=1
fi

declare -A mlx5fw_go_parsed
failed=0
total=0
out=""
integration=0

. $(dirname ${0})/lib.sh

verbose=1

for c_f in ${list}; do
	grep -F -q "*" <<< ${c_f} && continue
	unset mlx5fw_go_parsed
	declare -A mlx5fw_go_parsed

	unset mlx5fw_go_crc
	declare -A mlx5fw_go_crc
	f="${c_f}"
	if [[ ${c_f} == *.xz ]]; then
		xz -kd "${c_f}"
		f="${c_f/.xz/}"
	fi
	[[ ${verbose} == 1 ]] && echo -e "\n\n\ntest report for ${f}"
	#mlx5fw_go_out=$(./mlx5fw-go sections -f "${f}" | grep "^0x" | awk '($1 ~ /^[A-Z_0-9]+/){print $2" "$3}' 2>&1)
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
	fi

	OLD_IFS="${IFS}"
	IFS=$'\n'
	for l in ${mlx5fw_go_out}; do
		IFS="${OLD_IFS}"
		set -- ${l}
		if [[ "${1}" == "-I-" ]] || [[ "${1}" == "-E-" ]]; then
			continue
		fi
		addr=$(echo "${1}" | cut -d'/' -f 2 | cut -d'-' -f 1)
		name=$(echo "${3}" | cut -d'(' -f 2 | cut -d')' -f 1)
		crc_state="${5} ${6}"
		# addr="${2}"
		# name="${1}"
		addrs="${mlx5fw_go_parsed[${name}]}"
		mlx5fw_go_parsed[${name}]="${addrs} ${addr}"
		mlx5fw_go_crc["${addr}"]="${crc_state}"
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
		crc_state="${5} ${6}"
		mlx5fw_go_addrs="${mlx5fw_go_parsed[${name}]}"
		mlx5fw_go_crc_state="${mlx5fw_go_crc["${addr}"]}"
		total=$((total+1))
		if [[ -z "${mlx5fw_go_addrs}" ]]; then
			failed=$((failed+1))
			[[ ${verbose} == 1 ]] && echo "  ERROR(${f}): mlx5fw-go output doesn't have section '${name}'" ||:
			continue
		else
			[[ ${extra_verbose} == 1 ]] && echo "  OK(${f}): section '${name}' is present"
		fi

		total=$((total+1))
		grep -q "${addr}" <<< "${mlx5fw_go_addrs}"
		if [[ ${?} -ne 0 ]]; then
			failed=$((failed+1))
			[[ ${verbose} == 1 ]] && echo "  ERROR(${f}: mlx5fw-go output for section ${name} have a wrong address/adresses. got '${mlx5fw_go_parsed[${name}]}', expected '${addr}'"
		else
			[[ ${extra_verbose} == 1 ]] && echo "  OK(${f}): Address for section '${name}' present"
		fi

		total=$((total+1))
		if [[ -z "${mlx5fw_go_crc_state}" ]]; then
			failed=$((failed+1))
			[[ ${verbose} == 1 ]] && echo "  ERROR(${f}): mlx5fw-go output doesn't have section at address '${addr}'" ||:
			continue
		else
			[[ ${extra_verbose} == 1 ]] && echo "  OK(${f}): section with '${addr}' is present"
		fi

		if [[ "${mlx5fw_go_crc_state}" != "${crc_state}" ]]; then
			failed=$((failed+1))
			[[ ${verbose} == 1 ]] && echo "  ERROR(${f}): mlx5fw-go crc state doesn't match for section '${name}' at address '${addr}, got "${mlx5fw_go_crc_state}", expected "${crc_state}"'" ||:
		else
			[[ ${extra_verbose} == 1 ]] && echo "  OK(${f}): section '${name}' have same CRC state"
		fi
	done
	IFS="${OLD_IFS}"
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
