#!/bin/bash
# Script to run mstflint query --full with debug flags enabled

echo "Running mstflint -d 07:00.0 query --full with debug flags..."
echo "Debug flags enabled: FW_COMPS_DEBUG, MFT_DEBUG, FLASH_DEBUG, MLXFWOPS_ERRMSG_DEBUG"
echo "================================================"

sudo env FW_COMPS_DEBUG=1 \
         MFT_DEBUG=1 \
         FLASH_DEBUG=1 \
         MLXFWOPS_ERRMSG_DEBUG=1 \
         FLASH_ACCESS_DEBUG=1 \
         mstflint -d 07:00.0 query full 2>&1 | tee docs/mstflint-pcie-investigation/logs/01-query-full.log

echo "================================================"
echo "Log saved to: docs/mstflint-pcie-investigation/logs/01-query-full.log"