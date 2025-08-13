#!/bin/bash
set -euo pipefail

# Usage: BDF=0000:07:00.0 scripts/debug/icmd_trace.sh
BDF=${BDF:-${1:-}}
if [[ -z "${BDF}" ]]; then
  echo "Usage: BDF=<BDF> $0 or $0 <BDF>" >&2
  exit 1
fi

OUT_DIR=${OUT_DIR:-/tmp}
PREFIX=${PREFIX:-mst_icmd}
LOG_BASE="$OUT_DIR/$PREFIX.strace"

sudo -E strace -ff -yy -e trace=read,write,pread64,pwrite64,lseek,_llseek -o "$LOG_BASE" mstflint -d "$BDF" query full || true

# Summarize relevant offsets usage
files=("${LOG_BASE}".*)
reads_ctrl=0; writes_ctrl=0
reads_mbx=0;  writes_mbx=0
reads_aux=0;  writes_aux=0
for f in "${files[@]}"; do
  # pread/pwrite with offset patterns
  rc=$(grep -E "\bpread(64)?\(.*0x0(\b|[,\)])" -c "$f" || true); reads_ctrl=$((reads_ctrl+rc))
  wc=$(grep -E "\bpwrite(64)?\(.*0x0(\b|[,\)])" -c "$f" || true); writes_ctrl=$((writes_ctrl+wc))
  rm=$(grep -E "\bpread(64)?\(.*0x100000(\b|[,\)])" -c "$f" || true); reads_mbx=$((reads_mbx+rm))
  wm=$(grep -E "\bpwrite(64)?\(.*0x100000(\b|[,\)])" -c "$f" || true); writes_mbx=$((writes_mbx+wm))
  # lseek to offsets before read/write
  rc2=$(grep -E "\blseek\(.*0x0(\b|[,\)])|\b_llseek\(.*0x0(\b|[,\)])" -c "$f" || true); reads_ctrl=$((reads_ctrl+rc2))
  rm2=$(grep -E "\blseek\(.*0x100000(\b|[,\)])|\b_llseek\(.*0x100000(\b|[,\)])" -c "$f" || true); reads_mbx=$((reads_mbx+rm2))
  ra2=$(grep -E "\blseek\(.*0x1000(\b|[,\)])|\b_llseek\(.*0x1000(\b|[,\)])" -c "$f" || true); reads_aux=$((reads_aux+ra2))
done

echo "trace_dir=$OUT_DIR prefix=$PREFIX"
echo "icmd_ctrl_reads=$reads_ctrl icmd_ctrl_writes=$writes_ctrl"
echo "mbx_reads=$reads_mbx mbx_writes=$writes_mbx"
echo "aux_reads=$reads_aux aux_writes=$writes_aux"

# Print a couple of sample lines for quick inspection
echo "samples:"
grep -E "\bpwrite(64)?\(.*0x0(\b|[,\)])|\blseek\(.*0x0(\b|[,\)])" -n "${files[@]}" | head -n 3 || true
grep -E "\bpwrite(64)?\(.*0x100000(\b|[,\)])|\blseek\(.*0x100000(\b|[,\)])" -n "${files[@]}" | head -n 3 || true
