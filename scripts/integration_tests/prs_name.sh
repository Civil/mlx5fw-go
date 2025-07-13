#!/bin/bash

declare -A mlx5fw_go_parsed
failed=0
total=0
out=""
t="PRS Name"
pattern="^${t}"
integration=1

. $(dirname ${0})/lib.sh

do_test "${t}" "${pattern}" "${verbose}"
