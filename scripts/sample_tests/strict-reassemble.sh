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

	f_sha=$(sha256sum ${f} | awk '{print $1}')
	n_f_sha=$(sha256sum ${n_f} | awk '{print $1}')
	
	total=$((total+1))
	if [[ "${f_sha}" != "${n_f_sha}" ]]; then
		failed=$((failed+1))
		[[ ${verbose} == 1 ]] && echo "sha256sum mismatch for ${n_f} (${n_f_sha}) vs ${f} (${f_sha})"
	fi

	rm -rf "${out_dir}"
	if [[ ${c_f} == *.xz ]]; then
		rm -f "${f}"
	fi
done

score=$(echo | awk -vtotal=${total} -vfailed=${failed} 'END{print (total+0.0-failed)/total}')
ec="0"
if [[ ${score} != "1" ]]; then
	ec=1
fi
echo "${score}"
if [[ ! -z ${out} ]]; then
	echo "${out}"
fi

exit ${ec}
