//go:build ignore
// +build linux

//
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/Civil/mlx5fw-go/pkg/dev/pcie"
	"github.com/Civil/mlx5fw-go/pkg/interfaces"
)

// runQueryDeviceCommand opens a PCIe device via MST/sysfs backend.
// Phase 1: Just open/close to validate discovery; actual query comes in Phase 2.
func runQueryDeviceCommand(cmd *cobra.Command, args []string, fullOutput bool, jsonOutput bool) error {
	logger.Debug("Querying device via PCIe backend", zap.String("device", deviceBDF))

	openSpec := deviceBDF
	if mstPath != "" {
		openSpec = mstPath
	}
	dev, err := pcie.Open(openSpec, logger)
	if err != nil {
		return fmt.Errorf("failed to open device %s: %w", deviceBDF, err)
	}
	defer dev.Close()

	// Ensure MST transport uses LE ctrl on this NIC (observed requirement for ICMD VCR)
	if dev.Type() == "mst" && os.Getenv("MLX5FW_CTRL_LE") == "" {
		_ = os.Setenv("MLX5FW_CTRL_LE", "1")
		logger.Info("mst.ctrl_le.enabled")
	}
	// Prefer MCQI dword-based packers on MST for better compatibility
	if dev.Type() == "mst" && os.Getenv("MLX5FW_MCQI_PACK") == "" {
		_ = os.Setenv("MLX5FW_MCQI_PACK", "dwords")
		logger.Info("mst.mcqi_pack.dwords.enabled")
	}

	logger.Debug("Device opened successfully", zap.String("backend", dev.Type()))

	// Derive BDF from MST when not provided, to allow sysfs augmentation in AR client
	bdfForAug := deviceBDF
	if dev.Type() == "mst" && bdfForAug == "" {
		if params, err := pcie.MSTParams(dev); err == nil {
			bdfForAug = fmt.Sprintf("%04x:%02x:%02x.%d", params.Domain, params.Bus, params.Slot, params.Func)
			logger.Info("mst.derived_bdf", zap.String("bdf", bdfForAug))
		}
	}

	// Phase 2 scaffold: attempt Access Register based query
	ar := pcie.NewARClient(dev, bdfForAug, logger)
	info, err := ar.QueryFirmwareInfo()
	if err != nil {
		return fmt.Errorf("device-mode query not implemented: %w", err)
	}
	// Build FirmwareInfo for mstflint-parity display
	fwInfo := &interfaces.FirmwareInfo{}
	// Default to FS4 for NIC families in device-mode; MGIR may override
	fwInfo.Format = "FS4"
	if fw, ok := info["FWVersion"].(string); ok {
		fwInfo.FWVersion = fw
		fwInfo.ProductVersion = fw
	}
	if psid, ok := info["PSID"].(string); ok {
		fwInfo.PSID = psid
	}
	if sa, ok := info["SecurityAttrs"].(string); ok {
		fwInfo.SecurityAttrs = sa
	}
	if sv, ok := info["SecurityVer"].(uint16); ok {
		fwInfo.SecurityVer = int(sv)
	}
	// Optional richer fields
	if pn, ok := info["PartNumber"].(string); ok {
		fwInfo.PartNumber = pn
	}
	if desc, ok := info["Description"].(string); ok {
		fwInfo.Description = desc
	}
	if prs, ok := info["PRSName"].(string); ok {
		fwInfo.PRSName = prs
	}
	if ivsd, ok := info["ImageVSD"].(string); ok {
		fwInfo.ImageVSD = ivsd
	}
	if dvsd, ok := info["DeviceVSD"].(string); ok {
		fwInfo.DeviceVSD = dvsd
	}
	if rdate, ok := info["FWReleaseDate"].(string); ok {
		fwInfo.FWReleaseDate = rdate
	}
	if it, ok := info["ImageType"].(string); ok && it != "" {
		fwInfo.Format = it
	}
	if guid, ok := info["BaseGUID"].(uint64); ok {
		fwInfo.BaseGUID = guid
	}
	if gnum, ok := info["BaseGUIDNum"].(int); ok {
		fwInfo.BaseGUIDNum = gnum
	}
	if mac, ok := info["BaseMAC"].(uint64); ok {
		fwInfo.BaseMAC = mac
	}
	if mnum, ok := info["BaseMACNum"].(int); ok {
		fwInfo.BaseMACNum = mnum
	}
	// Vendor/Device from hex-like strings 0xVVVV,0xDDDD
	if vstr, ok := info["VendorID"].(string); ok && len(vstr) >= 4 {
		fmt.Sscanf(strings.TrimPrefix(vstr, "0x"), "%x", &fwInfo.VendorID)
	}
	if dstr, ok := info["DeviceID"].(string); ok && len(dstr) >= 4 {
		fmt.Sscanf(strings.TrimPrefix(dstr, "0x"), "%x", &fwInfo.DeviceID)
	}
	// Render device-mode output
	if am, ok := info["ActivationMethod"].(string); ok {
		fwInfo.ActivationMethod = am
	}
	// ROM info from device-mode (if present)
	if romAny, ok := info["ROMInfo"]; ok {
		if arr, ok := romAny.([]map[string]string); ok {
			for _, e := range arr {
				fwInfo.RomInfo = append(fwInfo.RomInfo, interfaces.RomInfo{
					Type:    e["type"],
					Version: e["version"],
					CPU:     e["cpu"],
				})
			}
		}
	}
	if jsonOutput {
		return outputDeviceJSON(fwInfo, info)
	}
	return displayQueryInfo(fwInfo, fullOutput, jsonOutput)
}
