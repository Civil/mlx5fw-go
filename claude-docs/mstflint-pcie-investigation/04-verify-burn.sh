#!/bin/bash
# Script to verify a firmware image against device with debug output

echo "Verifying firmware image against device with debug flags..."
echo "Debug flags enabled: FW_COMPS_DEBUG, MFT_DEBUG, FLASH_DEBUG, MLXFWOPS_ERRMSG_DEBUG"
echo "================================================"

# Verify command checks if the image is compatible without burning
sudo env FW_COMPS_DEBUG=1 \
         MFT_DEBUG=1 \
         FLASH_DEBUG=1 \
         MLXFWOPS_ERRMSG_DEBUG=1 \
         FLASH_ACCESS_DEBUG=1 \
         mstflint -d 07:00.0 \
         -i /home/civil/go/src/github.com/Civil/mlx5fw-go/sample_firmwares/mcx5/MCX516A-CDA_Ax_Bx_MT_0000000013_rel-16_35.2000.bin \
         verify 2>&1 | tee docs/mstflint-pcie-investigation/logs/04-verify-burn.log

echo "================================================"
echo "Verification log saved to: docs/mstflint-pcie-investigation/logs/04-verify-burn.log"