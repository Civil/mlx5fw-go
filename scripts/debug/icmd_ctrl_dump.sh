#!/bin/bash
set -euo pipefail

# Usage: BDF=0000:07:00.0 scripts/debug/icmd_ctrl_dump.sh
BDF=${BDF:-${1:-}}
if [[ -z "${BDF}" ]]; then
  echo "Usage: BDF=<BDF> $0 or $0 <BDF>" >&2
  exit 1
fi

read_ctrl() {
  # Uses mlx5fw-go to read ICMD ctrl (VCR) and prints the raw value
  local line
  line=$(sudo ./mlx5fw-go debug read32 -d "$BDF" -s icmd -o 0x0 2>/dev/null | tail -n1)
  # Expected format: 0x00000000: 0xXXXXXXXX
  echo "$line" | awk '{print $2}'
}

dump_mbx_head() {
  # Dump 4 dwords from VCR mailbox head at 0x100000
  sudo ./mlx5fw-go debug read32 -d "$BDF" -s icmd -o 0x100000 -c 4 2>/dev/null | awk '{print $2}' | tr '\n' ' '
}

decode_fields() {
  local v=$1
  # v like 0x12345678
  local dec=$((v))
  local busy=$((dec & 0x1))
  local exmb=$(((dec >> 1) & 0x1))
  local status=$(((dec >> 8) & 0xff))
  local opcode=$(((dec >> 16) & 0xffff))
  echo "busy=$busy status=0x$(printf %02x $status) opcode=0x$(printf %04x $opcode) exmb=$exmb"
}

echo "bdf=$BDF"
CTRL_BEFORE=$(read_ctrl)
MBX_BEFORE=$(dump_mbx_head)
echo "ctrl_before=$CTRL_BEFORE $(decode_fields $CTRL_BEFORE)"
echo -n "mbx_head_before=[ "; echo -n "$MBX_BEFORE"; echo "]"

FW_COMPS_DEBUG=1 MFT_DEBUG=1 sudo -E mstflint -d "$BDF" query full || true

CTRL_AFTER=$(read_ctrl)
MBX_AFTER=$(dump_mbx_head)
echo "ctrl_after=$CTRL_AFTER $(decode_fields $CTRL_AFTER)"
echo -n "mbx_head_after=[ "; echo -n "$MBX_AFTER"; echo "]"
