//go:build ignore
// +build ignore

//
package main

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/Civil/mlx5fw-go/pkg/dev/pcie"
)

func CreateDebugCommand() *cobra.Command {
	debugCmd := &cobra.Command{
		Use:   "debug",
		Short: "Low-level device debug helpers",
		Long:  "Raw PCICONF gateway access for development and bring-up.",
	}

	// read32
	var (
		r32SpaceStr  string
		r32OffsetStr string
		r32Count     int
	)
	read32Cmd := &cobra.Command{
		Use:   "read32",
		Short: "Read 32-bit words via PCICONF",
		RunE: func(cmd *cobra.Command, args []string) error {
			if deviceBDF == "" {
				return errors.New("-d <BDF> is required for debug commands")
			}
			space, err := parseSpace(r32SpaceStr)
			if err != nil {
				return fmt.Errorf("invalid --space: %w", err)
			}
			offset, err := parseUint(r32OffsetStr)
			if err != nil {
				return fmt.Errorf("invalid --offset: %w", err)
			}

			// Prefer direct MST path when provided
			openSpec := deviceBDF
			if mstPath != "" {
				openSpec = mstPath
			}
			dev, err := pcie.Open(openSpec, logger)
			if err != nil {
				return fmt.Errorf("open device: %w", err)
			}
			defer dev.Close()
			logger.Info("Device opened", zap.String("backend", dev.Type()))

			if r32Count <= 0 {
				r32Count = 1
			}
			logger.Info("pciconf.read32.begin",
				zap.String("device", openSpec),
				zap.String("backend", dev.Type()),
				zap.Uint16("space", uint16(space)),
				zap.Uint32("offset", uint32(offset)),
				zap.Int("count", r32Count),
			)
			for i := 0; i < r32Count; i++ {
				off := uint32(offset) + uint32(i*4)
				v, err := dev.Read32(uint16(space), off)
				if err != nil {
					logger.Error("pciconf.read32.error",
						zap.String("device", openSpec),
						zap.String("backend", dev.Type()),
						zap.Uint16("space", uint16(space)),
						zap.Uint32("offset", off),
						zap.Error(err),
					)
					return fmt.Errorf("read32 failed at offset 0x%08x: %w", off, err)
				}
				logger.Info("pciconf.read32.ok",
					zap.String("device", openSpec),
					zap.String("backend", dev.Type()),
					zap.Uint16("space", uint16(space)),
					zap.Uint32("offset", off),
					zap.Uint32("value", v),
				)
				fmt.Printf("0x%08x: 0x%08x\n", off, v)
			}
			return nil
		},
	}
	read32Cmd.Flags().StringVarP(&r32SpaceStr, "space", "s", "cr", "Address space (name: cr, icmd, pci_cr, ... or hex/decimal)")
	read32Cmd.Flags().StringVarP(&r32OffsetStr, "offset", "o", "0", "Offset (hex like 0x58 or decimal)")
	read32Cmd.Flags().IntVarP(&r32Count, "count", "c", 1, "Number of dwords to read")
	debugCmd.AddCommand(read32Cmd)

	// write32
	var (
		w32SpaceStr  string
		w32OffsetStr string
		w32ValueStr  string
	)
	write32Cmd := &cobra.Command{
		Use:   "write32",
		Short: "Write a 32-bit word via PCICONF",
		RunE: func(cmd *cobra.Command, args []string) error {
			if deviceBDF == "" {
				return errors.New("-d <BDF> is required for debug commands")
			}
			space, err := parseSpace(w32SpaceStr)
			if err != nil {
				return fmt.Errorf("invalid --space: %w", err)
			}
			offset, err := parseUint(w32OffsetStr)
			if err != nil {
				return fmt.Errorf("invalid --offset: %w", err)
			}
			val, err := parseUint(w32ValueStr)
			if err != nil {
				return fmt.Errorf("invalid --value: %w", err)
			}

			openSpec := deviceBDF
			if mstPath != "" {
				openSpec = mstPath
			}
			dev, err := pcie.Open(openSpec, logger)
			if err != nil {
				return fmt.Errorf("open device: %w", err)
			}
			defer dev.Close()
			logger.Info("Device opened", zap.String("backend", dev.Type()))
			logger.Info("pciconf.write32.begin",
				zap.String("device", openSpec),
				zap.String("backend", dev.Type()),
				zap.Uint16("space", uint16(space)),
				zap.Uint32("offset", uint32(offset)),
				zap.Uint32("value", uint32(val)),
			)
			if err := dev.Write32(uint16(space), uint32(offset), uint32(val)); err != nil {
				logger.Error("pciconf.write32.error",
					zap.String("device", openSpec),
					zap.String("backend", dev.Type()),
					zap.Uint16("space", uint16(space)),
					zap.Uint32("offset", uint32(offset)),
					zap.Uint32("value", uint32(val)),
					zap.Error(err),
				)
				return fmt.Errorf("write32 failed: %w", err)
			}
			logger.Info("pciconf.write32.ok",
				zap.String("device", openSpec),
				zap.String("backend", dev.Type()),
				zap.Uint16("space", uint16(space)),
				zap.Uint32("offset", uint32(offset)),
				zap.Uint32("value", uint32(val)),
			)
			fmt.Printf("Wrote 0x%08x to offset 0x%08x (space 0x%x)\n", uint32(val), uint32(offset), uint16(space))
			return nil
		},
	}
	write32Cmd.Flags().StringVarP(&w32SpaceStr, "space", "s", "cr", "Address space (name: cr, icmd, pci_cr, ... or hex/decimal)")
	write32Cmd.Flags().StringVarP(&w32OffsetStr, "offset", "o", "0", "Offset (hex like 0x58 or decimal)")
	write32Cmd.Flags().StringVarP(&w32ValueStr, "value", "v", "0", "Value (hex like 0xdeadbeef or decimal)")
	debugCmd.AddCommand(write32Cmd)

	// readblock
	var (
		rbSpaceStr  string
		rbOffsetStr string
		rbSize      int
	)
	readBlockCmd := &cobra.Command{
		Use:   "readblock",
		Short: "Read a byte block via PCICONF",
		RunE: func(cmd *cobra.Command, args []string) error {
			if deviceBDF == "" {
				return errors.New("-d <BDF> is required for debug commands")
			}
			space, err := parseSpace(rbSpaceStr)
			if err != nil {
				return fmt.Errorf("invalid --space: %w", err)
			}
			offset, err := parseUint(rbOffsetStr)
			if err != nil {
				return fmt.Errorf("invalid --offset: %w", err)
			}
			if rbSize <= 0 {
				return fmt.Errorf("--size must be > 0")
			}

			openSpec := deviceBDF
			if mstPath != "" {
				openSpec = mstPath
			}
			dev, err := pcie.Open(openSpec, logger)
			if err != nil {
				return fmt.Errorf("open device: %w", err)
			}
			defer dev.Close()
			logger.Info("Device opened", zap.String("backend", dev.Type()))
			logger.Info("pciconf.readblock.begin",
				zap.String("device", openSpec),
				zap.String("backend", dev.Type()),
				zap.Uint16("space", uint16(space)),
				zap.Uint32("offset", uint32(offset)),
				zap.Int("size", rbSize),
			)
			data, err := dev.ReadBlock(uint16(space), uint32(offset), rbSize)
			if err != nil {
				logger.Error("pciconf.readblock.error",
					zap.String("device", openSpec),
					zap.String("backend", dev.Type()),
					zap.Uint16("space", uint16(space)),
					zap.Uint32("offset", uint32(offset)),
					zap.Int("size", rbSize),
					zap.Error(err),
				)
				return fmt.Errorf("readblock failed: %w", err)
			}
			logger.Info("pciconf.readblock.ok",
				zap.String("device", openSpec),
				zap.String("backend", dev.Type()),
				zap.Uint16("space", uint16(space)),
				zap.Uint32("offset", uint32(offset)),
				zap.Int("size", len(data)),
			)
			// hex dump, 16 bytes per line
			for i := 0; i < len(data); i += 16 {
				end := i + 16
				if end > len(data) {
					end = len(data)
				}
				fmt.Printf("0x%08x:", uint32(offset)+uint32(i))
				for j := i; j < end; j++ {
					if (j-i)%4 == 0 {
						fmt.Printf(" ")
					}
					fmt.Printf(" %02x", data[j])
				}
				fmt.Println()
			}
			return nil
		},
	}
	readBlockCmd.Flags().StringVarP(&rbSpaceStr, "space", "s", "cr", "Address space (name: cr, icmd, pci_cr, ... or hex/decimal)")
	readBlockCmd.Flags().StringVarP(&rbOffsetStr, "offset", "o", "0", "Offset (hex like 0x58 or decimal)")
	readBlockCmd.Flags().IntVarP(&rbSize, "size", "n", 64, "Number of bytes to read")
	debugCmd.AddCommand(readBlockCmd)

	// Add mst-diff comparator (if built)
	debugCmd.AddCommand(createDebugMSTDiffCommand())

	// mst-params
	paramsCmd := &cobra.Command{
		Use:   "mst-params",
		Short: "Show MST kernel device parameters",
		RunE: func(cmd *cobra.Command, args []string) error {
			if deviceBDF == "" {
				return errors.New("-d <BDF> is required for debug commands")
			}
			openSpec := deviceBDF
			if mstPath != "" {
				openSpec = mstPath
			}
			dev, err := pcie.Open(openSpec, logger)
			if err != nil {
				return fmt.Errorf("open device: %w", err)
			}
			defer dev.Close()
			logger.Info("Device opened", zap.String("backend", dev.Type()))
			p, err := pcie.MSTParams(dev)
			if err != nil {
				return err
			}
			fmt.Printf("Domain: 0x%x Bus: 0x%x Slot: 0x%x Func: 0x%x\n", p.Domain, p.Bus, p.Slot, p.Func)
			fmt.Printf("Vendor: 0x%04x Device: 0x%04x Subsys: 0x%04x/0x%04x\n", p.Vendor, p.Device, p.SubsystemVendor, p.SubsystemDevice)
			fmt.Printf("Functional VSEC offset: 0x%x\n", p.FunctionalVsecOffset)
			return nil
		},
	}
	debugCmd.AddCommand(paramsCmd)

	// cfg-read
	var crOffsetStr string
	cfgReadCmd := &cobra.Command{
		Use:   "cfg-read",
		Short: "Read PCI config dword via MST",
		RunE: func(cmd *cobra.Command, args []string) error {
			if deviceBDF == "" {
				return errors.New("-d <BDF>|/dev/mst/<node> is required")
			}
			off, err := parseUint(crOffsetStr)
			if err != nil {
				return fmt.Errorf("invalid --offset: %w", err)
			}
			openSpec := deviceBDF
			if mstPath != "" {
				openSpec = mstPath
			}
			dev, err := pcie.Open(openSpec, logger)
			if err != nil {
				return fmt.Errorf("open device: %w", err)
			}
			defer dev.Close()
			logger.Info("Device opened", zap.String("backend", dev.Type()))
			logger.Info("pciconf.cfg_read.begin",
				zap.String("device", openSpec),
				zap.String("backend", dev.Type()),
				zap.Uint32("offset", uint32(off)),
			)
			v, err := pcie.ReadPCIConfigDword(dev, uint32(off))
			if err != nil {
				// Fallback to sysfs config read when BDF is provided
				if pcie.IsBDF(deviceBDF) {
					v2, ferr := pcie.ReadPCIConfigSysfs(deviceBDF, uint32(off))
					if ferr != nil {
						return fmt.Errorf("cfg-read failed (mst=%v, sysfs=%v)", err, ferr)
					}
					logger.Info("pciconf.cfg_read.sysfs",
						zap.String("device", deviceBDF),
						zap.Uint32("offset", uint32(off)),
						zap.Uint32("value", v2),
					)
					fmt.Printf("cfg[0x%02x] = 0x%08x (sysfs)\n", uint32(off), v2)
					return nil
				}
				logger.Error("pciconf.cfg_read.error",
					zap.String("device", openSpec),
					zap.String("backend", dev.Type()),
					zap.Uint32("offset", uint32(off)),
					zap.Error(err),
				)
				return err
			}
			logger.Info("pciconf.cfg_read.mst",
				zap.String("device", openSpec),
				zap.String("backend", dev.Type()),
				zap.Uint32("offset", uint32(off)),
				zap.Uint32("value", v),
			)
			fmt.Printf("cfg[0x%02x] = 0x%08x (mst)\n", uint32(off), v)
			return nil
		},
	}
	cfgReadCmd.Flags().StringVarP(&crOffsetStr, "offset", "o", "0", "PCI config offset (hex or decimal)")
	debugCmd.AddCommand(cfgReadCmd)

	// list-mst
	listMstCmd := &cobra.Command{
		Use:   "list-mst",
		Short: "List /dev/mst nodes and attempt MST_PARAMS",
		RunE: func(cmd *cobra.Command, args []string) error {
			entries, err := pcie.ListMSTNodes()
			if err != nil {
				return err
			}
			if len(entries) == 0 {
				fmt.Println("No /dev/mst entries found")
				return nil
			}
			for _, path := range entries {
				bdf, params, perr := pcie.ProbeMSTNode(path)
				if perr != nil {
					fmt.Printf("%s: error=%v\n", path, perr)
				} else {
					fmt.Printf("%s: BDF=%s vendor=0x%04x device=0x%04x vsec=0x%x\n", path, bdf, params.Vendor, params.Device, params.FunctionalVsecOffset)
				}
			}
			return nil
		},
	}
	debugCmd.AddCommand(listMstCmd)

	// space-check
	spaceCheckCmd := &cobra.Command{
		Use:   "space-check",
		Short: "Probe supported address spaces via PCICONF",
		RunE: func(cmd *cobra.Command, args []string) error {
			if deviceBDF == "" {
				return errors.New("-d <BDF>|/dev/mst/<node> is required")
			}
			openSpec := deviceBDF
			if mstPath != "" {
				openSpec = mstPath
			}
			dev, err := pcie.Open(openSpec, logger)
			if err != nil {
				return fmt.Errorf("open device: %w", err)
			}
			defer dev.Close()
			logger.Info("Device opened", zap.String("backend", dev.Type()))
			results := pcie.ProbeSpaces(dev)
			for _, line := range results {
				fmt.Println(line)
			}
			return nil
		},
	}
	debugCmd.AddCommand(spaceCheckCmd)
	// Nest AR subcommands under debug
	debugCmd.AddCommand(createDebugARCommands())

	return debugCmd
}

func parseUint(s string) (uint64, error) {
	if len(s) > 2 && (s[0:2] == "0x" || s[0:2] == "0X") {
		v, err := strconv.ParseUint(s[2:], 16, 64)
		if err != nil {
			return 0, err
		}
		return v, nil
	}
	return strconv.ParseUint(s, 10, 64)
}

func parseSpace(s string) (uint64, error) {
	// Support common names matching mstflint address_space_t
	switch s {
	case "cr":
		return 0x2, nil // AS_CR_SPACE
	case "icmd":
		return 0x3, nil // AS_ICMD
	case "icmd_ext":
		return 0x1, nil // AS_ICMD_EXT
	case "nodnic_init", "init_seg":
		return 0x4, nil // AS_NODNIC_INIT_SEG
	case "exprom", "rom":
		return 0x5, nil // AS_EXPANSION_ROM
	case "nd_cr":
		return 0x6, nil // AS_ND_CRSPACE
	case "scan_cr":
		return 0x7, nil // AS_SCAN_CRSPACE
	case "sem", "semaphore":
		return 0xa, nil // AS_SEMAPHORE
	case "recovery":
		return 0x0c, nil // AS_RECOVERY
	case "mac":
		return 0x0f, nil // AS_MAC
	case "pci_icmd":
		return 0x101, nil // AS_PCI_ICMD
	case "pci_cr":
		return 0x102, nil // AS_PCI_CRSPACE
	case "pci_all_icmd":
		return 0x103, nil // AS_PCI_ALL_ICMD
	case "pci_scan_cr":
		return 0x107, nil // AS_PCI_SCAN_CRSPACE
	case "pci_gsem":
		return 0x10a, nil // AS_PCI_GLOBAL_SEMAPHORE
	}
	return parseUint(s)
}
