#!/bin/bash

declare -A mlx5fw_go_parsed
failed=0
total=0
out=""
t="FW Version"
pattern="^${t}"
integration=0

. $(dirname ${0})/lib.sh

do_test "${t}" "${pattern}" "${verbose}"
