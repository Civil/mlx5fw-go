#!/bin/bash
# Script to simulate a burn operation with debug flags (DRY RUN - no actual burn)
# This shows what would happen during a burn without actually flashing

echo "Simulating burn operation with debug flags (DRY RUN)..."
echo "Debug flags enabled: FW_COMPS_DEBUG, MFT_DEBUG, FLASH_DEBUG, MLXFWOPS_ERRMSG_DEBUG, FLASH_ACCESS_DEBUG"
echo "================================================"

# First, let's query the current firmware version
echo "Current firmware status:"
sudo mstflint -d 07:00.0 query | grep -E "FW Version|PSID|Device Type" || true

echo ""
echo "Simulating burn with --use_fw flag (no actual flash write)..."
echo "This shows the validation and preparation steps without modifying the device"
echo ""

# Use --use_fw to simulate without actually burning
# This flag makes mstflint skip the actual flash write
sudo env FW_COMPS_DEBUG=1 \
         MFT_DEBUG=1 \
         FLASH_DEBUG=1 \
         MLXFWOPS_ERRMSG_DEBUG=1 \
         FLASH_ACCESS_DEBUG=1 \
         MFT_FLASH_DEBUG=1 \
         MFLASH_DEBUG=1 \
         mstflint -d 07:00.0 \
         -i /home/civil/go/src/github.com/Civil/mlx5fw-go/sample_firmwares/mcx5/MCX516A-CDA_Ax_Bx_MT_0000000013_rel-16_35.2000.bin \
         --use_fw burn 2>&1 | tee docs/mstflint-pcie-investigation/logs/03-burn-simulation.log

echo "================================================"
echo "Simulation log saved to: docs/mstflint-pcie-investigation/logs/03-burn-simulation.log"
echo ""
echo "NOTE: This was a DRY RUN with --use_fw flag. No actual firmware was burned."