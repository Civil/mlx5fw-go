#!/usr/bin/env bash
set -euo pipefail

# Compares mlx5fw-go device-mode query against mstflint query and reports parity.
# Usage: BDF=0000:07:00.0 scripts/integration_tests/device_query_parity.sh

BDF=${BDF:-${1:-}}
if [[ -z "$BDF" ]]; then
  echo "Usage: BDF=<BDF> $0 or $0 <BDF>" >&2
  exit 1
fi

TMPDIR=${TMPDIR:-/tmp}
OUR_JSON="$TMPDIR/mlx5fw-go.query.$(echo "$BDF" | tr ':.' '_').json"
MST_TXT="$TMPDIR/mstflint.query.$(echo "$BDF" | tr ':.' '_').txt"

# Run our tool (device-mode JSON)
sudo ./mlx5fw-go query -d "$BDF" --json > "$OUR_JSON"

# Run mstflint text
sudo mstflint -d "$BDF" query > "$MST_TXT" || true

# Extract mstflint fields
mst_get() {
  local key="$1"
  awk -v k="$key" -F':' 'tolower($1) ~ tolower(k) { sub(/^ */,"",$2); print $2; exit }' "$MST_TXT"
}

# Fields to compare: keep to ones we can reasonably match
declare -A fields=(
  [fw_version]="FW Version"
  [fw_release_date]="FW Release Date"
  [product_version]="Product Version"
  [psid]="PSID"
  [image_type]="Image type"
  [image_vsd]="Image VSD"
  [device_vsd]="Device VSD"
  [part_number]="Part Number"
  [description]="Description"
  [security_attributes]="Security Attributes"
)

ok=0; total=0
printf "Comparing fields for %s\n" "$BDF"
for json_key in "${!fields[@]}"; do
  mst_key="${fields[$json_key]}"
  total=$((total+1))
  our_val=$(jq -r --arg k "$json_key" '.[$k] // ""' "$OUR_JSON" 2>/dev/null || echo "")
  mst_val=$(mst_get "$mst_key")
  # normalize trivial N/A spacings
  our_val_norm=$(echo "$our_val" | sed 's/^\s\+//; s/\s\+$//')
  mst_val_norm=$(echo "$mst_val" | sed 's/^\s\+//; s/\s\+$//')
  if [[ "$json_key" == "description" ]] && echo "$mst_val_norm" | grep -Eq '^UID[[:space:]]+GuidsNumber'; then
    printf "OK   %-22s == (header-only)\n" "$json_key"
    ok=$((ok+1))
    continue
  fi
  if [[ "$our_val_norm" == "$mst_val_norm" ]]; then
    printf "OK   %-22s == %s\n" "$json_key" "$our_val_norm"
    ok=$((ok+1))
  else
    printf "DIFF %-22s ours=[%s] mst=[%s]\n" "$json_key" "$our_val_norm" "$mst_val_norm"
  fi
done

score=$(python3 - <<PY
ok=$ok; total=$total
print(f"{ok}/{total}={ok/total if total else 0:.2f}")
PY
)
echo "score=$score"
echo "our_json=$OUR_JSON"
echo "mst_txt=$MST_TXT"

# Compare ROM Info entries (order-insensitive)
echo
echo "Comparing ROM Info:"
awk '/^Rom Info:/ {print; p=1; next} p && $0 !~ /^\s+type=/ {p=0} p {print}' "$MST_TXT" > "$MST_TXT.rom" || true
mst_rom_json=$(awk '/^\s+type=/{
  t=$0; sub(/^\s+type=/, "", t);
  ver=""; cpu=""; type="";
  n=split(t, a, /[[:space:]]+/);
  type=a[1];
  for (i=2;i<=n;i++) {
    if (a[i] ~ /^version=/) { sub(/^version=/, "", a[i]); ver=a[i] }
    if (a[i] ~ /^cpu=/) { sub(/^cpu=/, "", a[i]); cpu=a[i] }
  }
  printf("{\"type\":\"%s\",\"version\":\"%s\",\"cpu\":\"%s\"}\n", type, ver, cpu)
}' "$MST_TXT.rom" | jq -s '.' )
our_rom_json=$(jq '.rom_info // []' "$OUR_JSON" 2>/dev/null || echo '[]')
mst_sorted=$(echo "$mst_rom_json" | jq 'sort_by(.type,.version,.cpu)')
our_sorted=$(echo "$our_rom_json" | jq 'sort_by(.type,.version,.cpu)')
if [[ "$mst_sorted" == "$our_sorted" ]]; then
  echo "OK   rom_info matches"
else
  echo "DIFF rom_info"
  echo "  ours=$(echo "$our_sorted" | jq -c '.')"
  echo "  mst =$(echo "$mst_sorted" | jq -c '.')"
fi
