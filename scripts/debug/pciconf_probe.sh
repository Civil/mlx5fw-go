#!/bin/bash
set -euo pipefail

# Usage: BDF=0000:07:00.0 scripts/debug/pciconf_probe.sh
BDF=${BDF:-${1:-}}
if [[ -z "${BDF}" ]]; then
  echo "Usage: BDF=<BDF> $0 or $0 <BDF>" >&2
  exit 1
fi

LOG=${LOG:-/tmp/pciconf_probe.${BDF//:/_}.log}
echo "Writing logs to $LOG" >&2

{
  echo "# space-check"
  sudo ./mlx5fw-go debug space-check -d "$BDF"
  echo "# icmd ctrl dwords"
  sudo ./mlx5fw-go debug read32 -d "$BDF" -s icmd -o 0x0 -c 4
  echo "# icmd mailbox head"
  sudo ./mlx5fw-go debug readblock -d "$BDF" -s icmd -o 0x100000 -n 32
  echo "# ar mgir raw"
  sudo ./mlx5fw-go debug ar get --reg mgir --raw -d "$BDF" -v
} 2>&1 | tee "$LOG"

echo "Probe complete. Log: $LOG"
