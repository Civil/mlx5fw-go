#!/bin/bash

# Check against sample sections
unset FW_COMPS_DEBUG
unset MFT_DEBUG

failed=0
total=0
out=""
verbose=1

list=""
declare -a files_list
mapfile -d $'\0' -t files_list < <(find /home/civil/fws/ -name '*.xz' -print0)

for i in {1..20}; do
	randfile=$(python3 -S -c "import random; print(random.randrange(0,${#files_list[@]}))")
	list="${list} ${files_list[$randfile]}"
done

for c_f in sample_firmwares/*.bin ${list}; do
	unset mlx5fw-go_parsed
	declare -A mlx5fw-go_parsed
	f="${c_f}"
	if [[ ${c_f} == *.xz ]]; then
		xz -kd "${c_f}"
		f="${c_f/.xz/}"
	fi
	total=$((total+1))
	mlx5fw-go_out=$(./mlx5fw-go -sections "${f}" | grep '^[A-Z_0-9]\+\s\+0x[a-fA-F0-9]\+\s\+[0-9]\+\s\+0x[0-9a-fA-F]\+' 2>&1)
	if [[ $? -ne 0 ]]; then
		failed=$((failed+1))
		grep -q 'panic' <<< "${mlx5fw-go_out}"
		if [[ $? -ne 0 ]]; then
			out="${out}$(echo -e '\n')$(echo "${mlx5fw-go_out}" | awk 'BEGIN{pr=0}{if ($0 ~ /panic/){pr=1;}; if (pr == 1) { print $0; };}')"
		fi
		continue
	fi

	OLD_IFS="${IFS}"
	IFS=$'\n'
	for l in ${mlx5fw-go_out}; do
		IFS="${OLD_IFS}"
		set -- ${l}
		addr="${mlx5fw-go_parsed[${1}]}"
		mlx5fw-go_parsed[${1}]="${addr} ${4}"
	done
	IFS="${OLD_IFS}"

	mft_out=$(mstflint -i "${f}" v | head -n -3 | tail -n +4)

	OLD_IFS="${IFS}"
	IFS=$'\n'
	test_failed=0
	for l in ${mft_out}; do
		IFS="${OLD_IFS}"
		set -- ${l}
		addr=$(echo "${1}" | cut -d'/' -f 2 | cut -d'-' -f 1)
		name=$(echo "${3}" | cut -d'(' -f 2 | cut -d')' -f 1)
		mlx5fw-go_addrs="${mlx5fw-go_parsed[${name}]}"
		grep -q "${addr}" <<< "${mlx5fw-go_addrs}"
		if [[ ${?} -ne 0 ]]; then
			test_failed=1
			if [[ ${verbose} == 1 ]]; then
				echo "Error for section ${name} for file ${f}:"
				if [[ -z "${mlx5fw-go_addrs}" ]]; then
					echo " mlx5fw-go output doesn't have section '${name}'"
				else
					echo " mlx5fw-go output for section ${name} have a wrong address/adresses. got '${mlx5fw-go_parsed[${name}]}', expected '${addr}'"
				fi
			fi
		fi
	done
	if [[ ${test_failed} -eq 1 ]]; then
		failed=$((failed + 1))
	fi
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
----------------------------------------------------
IRON_PREP_CODE       0x05    116420    0x00000038
RESET_INFO           0x20    256       0x00000118
MAIN_CODE            0x03    5404368   0x0000011c
PCIE_LINK_CODE       0x04    91692     0x00002a58
POST_IRON_BOOT_CODE  0x06    2608      0x00002b08
PCI_CODE             0x02    344148    0x00002b10
UPGRADE_CODE         0x07    9704      0x00002db0
PHY_UC_CODE          0x0a    101104    0x00002dc4
PCIE_PHY_UC_CODE     0x0c    78432     0x00002e88
CCIR_INFRA_CODE      0x0d    8212      0x00002f20
CCIR_ALGO_CODE       0x0e    6008      0x00002f30
IMAGE_INFO           0x10    1024      0x00002f3c
FW_MAIN_CFG          0x12    3072      0x00002f40
FW_BOOT_CFG          0x11    1216      0x00002f44
HW_MAIN_CFG          0x09    4608      0x00002f48
HW_BOOT_CFG          0x08    1472      0x00002f50
PHY_UC_CONSTS        0x0b    1280      0x00002f54
IMAGE_SIGNATURE_256  0xa0    320       0x00002f54
PUBLIC_KEYS_2048     0xa1    2304      0x00002f58
FORBIDDEN_VERSIONS   0xa2    144       0x00002f5c
IMAGE_SIGNATURE_512  0xa3    576       0x00002f5c
PUBLIC_KEYS_4096     0xa4    4352      0x00002f5c
RSA_PUBLIC_KEY       0xa6    4352      0x00002f64
RSA_4096_SIGNATURES  0xa7    1536      0x00002f6c
ROM_CODE             0x18    871524    0x00002f70
DGB_LOG_MAP          0x30    2672      0x00003618
DGB_FW_PARAMS        0x32    8         0x0000361c
NV_DATA              0xe9    116640    0x0000361c
exit 0
     /0x00000018-0x0000001f (0x000008)/ (HW_POINTERS) - OK
     /0x00000020-0x00000027 (0x000008)/ (HW_POINTERS) - OK
     /0x00000028-0x0000002f (0x000008)/ (HW_POINTERS) - OK
     /0x00000030-0x00000037 (0x000008)/ (HW_POINTERS) - OK
     /0x00000038-0x0000003f (0x000008)/ (HW_POINTERS) - OK
     /0x00000040-0x00000047 (0x000008)/ (HW_POINTERS) - OK
     /0x00000048-0x0000004f (0x000008)/ (HW_POINTERS) - OK
     /0x00000050-0x00000057 (0x000008)/ (HW_POINTERS) - OK
     /0x00000058-0x0000005f (0x000008)/ (HW_POINTERS) - OK
     /0x00000060-0x00000067 (0x000008)/ (HW_POINTERS) - OK
     /0x00000068-0x0000006f (0x000008)/ (HW_POINTERS) - OK
     /0x00000070-0x00000077 (0x000008)/ (HW_POINTERS) - OK
     /0x00000078-0x0000007f (0x000008)/ (HW_POINTERS) - OK
     /0x00000080-0x00000087 (0x000008)/ (HW_POINTERS) - OK
     /0x00000088-0x0000008f (0x000008)/ (HW_POINTERS) - OK
     /0x00000090-0x00000097 (0x000008)/ (HW_POINTERS) - OK
     /0x00000500-0x0000053f (0x000040)/ (TOOLS_AREA) - OK
     /0x00001000-0x00004013 (0x003014)/ (BOOT2) - OK
     /0x00005000-0x0000501f (0x000020)/ (ITOC_HEADER) - OK
     /0x00007000-0x000236c3 (0x01c6c4)/ (IRON_PREP_CODE) - OK
     /0x00023700-0x000237ff (0x000100)/ (RESET_INFO) - OK
     /0x00023980-0x0054b04f (0x5276d0)/ (MAIN_CODE) - OK
     /0x0054b080-0x005616ab (0x01662c)/ (PCIE_LINK_CODE) - OK
     /0x00561700-0x0056212f (0x000a30)/ (POST_IRON_BOOT_CODE) - OK
     /0x00562180-0x005b61d3 (0x054054)/ (PCI_CODE) - OK
     /0x005b6200-0x005b87e7 (0x0025e8)/ (UPGRADE_CODE) - OK
     /0x005b8800-0x005d12ef (0x018af0)/ (PHY_UC_CODE) - OK
     /0x005d1300-0x005e455f (0x013260)/ (PCIE_PHY_UC_CODE) - OK
     /0x005e4580-0x005e6593 (0x002014)/ (CCIR_INFRA_CODE) - OK
     /0x005e6600-0x005e7d77 (0x001778)/ (CCIR_ALGO_CODE) - OK
     /0x005e7d80-0x005e817f (0x000400)/ (IMAGE_INFO) - OK
     /0x005e8180-0x005e8d7f (0x000c00)/ (FW_MAIN_CFG) - OK
     /0x005e8d80-0x005e923f (0x0004c0)/ (FW_BOOT_CFG) - OK
     /0x005e9280-0x005ea47f (0x001200)/ (HW_MAIN_CFG) - OK
     /0x005ea480-0x005eaa3f (0x0005c0)/ (HW_BOOT_CFG) - OK
     /0x005eaa80-0x005eaf7f (0x000500)/ (PHY_UC_CONSTS) - OK
     /0x005eaf80-0x005eb0bf (0x000140)/ (IMAGE_SIGNATURE_256) - CRC IGNORED
     /0x005eb100-0x005eb9ff (0x000900)/ (PUBLIC_KEYS_2048) - OK
     /0x005eba00-0x005eba8f (0x000090)/ (FORBIDDEN_VERSIONS) - OK
     /0x005ebb00-0x005ebd3f (0x000240)/ (IMAGE_SIGNATURE_512) - CRC IGNORED
     /0x005ebd80-0x005ece7f (0x001100)/ (PUBLIC_KEYS_4096) - OK
     /0x005ece80-0x005edf7f (0x001100)/ (RSA_PUBLIC_KEY) - OK
     /0x005edf80-0x005ee57f (0x000600)/ (RSA_4096_SIGNATURES) - CRC IGNORED
     /0x005ee580-0x006c31e3 (0x0d4c64)/ (ROM_CODE) - OK
     /0x006c3200-0x006c3c6f (0x000a70)/ (DBG_FW_INI) - OK
     /0x006c3c80-0x006c3c87 (0x000008)/ (DBG_FW_PARAMS) - OK
     /0x006c3d00-0x006e049f (0x01c7a0)/ (CRDUMP_MASK_DATA) - OK
     /0x01fff000-0x01fff01f (0x000020)/ (DTOC_HEADER) - OK
     /0x01f00000-0x01f1ffff (0x020000)/ (FW_NV_LOG) - CRC IGNORED
     /0x01f20000-0x01f2ffff (0x010000)/ (NV_DATA) - CRC IGNORED
     /0x01f40000-0x01f4ffff (0x010000)/ (NV_DATA) - CRC IGNORED
     /0x01f60000-0x01f601ff (0x000200)/ (DEV_INFO) - OK
     /0x01f80000-0x01f9ffff (0x020000)/ (FW_INTERNAL_USAGE) - CRC IGNORED
     /0x01fa0000-0x01fbffff (0x020000)/ (PROGRAMMABLE_HW_FW) - CRC IGNORED
     /0x01fc0000-0x01fdffff (0x020000)/ (PROGRAMMABLE_HW_FW) - CRC IGNORED
     /0x01fe0000-0x01fedfff (0x00e000)/ (DIGITAL_CERT_RW) - OK
     /0x01fee000-0x01ff1fff (0x004000)/ (DIGITAL_CACERT_RW) - OK
     /0x01ff2000-0x01ff3fff (0x002000)/ (CERT_CHAIN_0) - OK
     /0x01ff4000-0x01ff4027 (0x000028)/ (DIGITAL_CERT_PTR) - OK
     /0x01ff8000-0x01ff813f (0x000140)/ (MFG_INFO) - OK
     /0x01ff8140-0x01ff813f (0x000000)/ (VPD_R0) - OK
