//go:build linux
// +build linux

//
package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/Civil/mlx5fw-go/pkg/dev/pcie"
)

// small shims to access pcie's BE conversion without import cycles
func pcieFromBE(dst, src []byte) {
	// replicate pcie.fromBE behavior (u32 chunks)
	n := len(src) / 4
	for i := 0; i < n; i++ {
		off := i * 4
		v := uint32(src[off])<<24 | uint32(src[off+1])<<16 | uint32(src[off+2])<<8 | uint32(src[off+3])
		dst[off] = byte(v)
		dst[off+1] = byte(v >> 8)
		dst[off+2] = byte(v >> 16)
		dst[off+3] = byte(v >> 24)
	}
}

// setBitsBE sets bitLen bits of val into dst starting at bitOff (MSB-first within each byte).
func setBitsBE(dst []byte, bitOff uint32, bitLen uint32, val uint32) {
	for i := uint32(0); i < bitLen; i++ {
		bit := (val >> (bitLen - 1 - i)) & 1
		pos := bitOff + i
		byteIdx := pos / 8
		bitInByte := pos % 8
		if int(byteIdx) >= len(dst) {
			break
		}
		mask := byte(1 << (7 - bitInByte))
		if bit != 0 {
			dst[byteIdx] |= mask
		}
	}
}

// parseStraceOps extracts a simplified sequence of MST block IOCTLs from strace output.
// We keep only READ4_BUFFER/WRITE4_BUFFER and classify by size: >=160 as mailbox, <160 as ctrl.
type traceOp struct {
	kind string
	size int
}

// rawOp captures any MST_BLOCK_ACCESS ioctl from strace (dir, nr, size)
type rawOp struct {
	dir  string // READ or WRITE
	nr   int
	size int // ioctl payload size
	ret  int // return value trailing '= N' if parsed, else 0
}

// parseStraceRawOps extracts raw MST_BLOCK_ACCESS ioctls (0xD2 magic) with dir, nr, size
func parseStraceRawOps(path string, maxOps int) ([]rawOp, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	ops := make([]rawOp, 0, 256)
	re := regexp.MustCompile(`_IOC\(_IOC_(READ|WRITE),\s*0x[dD]2,\s*([0-9x]+),\s*0x([0-9a-fA-F]+)\)`)
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		m := re.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		dir := m[1]
		nrStr := strings.ToLower(m[2])
		sizeHex := m[3]
		nrv := 0
		if strings.HasPrefix(nrStr, "0x") {
			if v, e := strconv.ParseInt(nrStr, 0, 32); e == nil {
				nrv = int(v)
			}
		} else {
			if v, e := strconv.Atoi(nrStr); e == nil {
				nrv = v
			}
		}
		sz := 0
		if v, e := strconv.ParseInt("0x"+sizeHex, 0, 32); e == nil {
			sz = int(v)
		}
		// parse trailing return value "= N" if present
		rv := 0
		if i := strings.LastIndex(line, ") = "); i >= 0 {
			tail := strings.TrimSpace(line[i+4:])
			if n, e := strconv.Atoi(strings.TrimPrefix(tail, "+")); e == nil {
				rv = n
			}
		}
		ops = append(ops, rawOp{dir: dir, nr: nrv, size: sz, ret: rv})
		if maxOps > 0 && len(ops) >= maxOps {
			break
		}
	}
	return ops, nil
}

// chooseSetStrategy parses MLX5FW_MST_SET_STRATEGY and returns a function producing space IDs.
// Supported forms:
//
//	rot:10,2,3  (default)
//	const:3
//	rot:3,2,10
/*
// NOTE: Previous session left this function body corrupted.
// Commenting out the broken implementation to restore buildability.
func chooseSetStrategy() func() uint16 {
	s := os.Getenv("MLX5FW_MST_SET_STRATEGY")
	if s == "" {
		// default rotation 10,2,3
		seq := []uint16{10, 2, 3}
		i := 0
		return func() uint16 { v := seq[i%len(seq)]; i++; return v }
	}
	parts := strings.SplitN(s, ":", 2)
	mode := parts[0]
	arg := ""
	if len(parts) > 1 {
		arg = parts[1]
	}
	switch mode {
	case "const":
		if arg == "" {
			return func() uint16 { return 3 }
		}
		if n, err := strconv.Atoi(strings.TrimSpace(arg)); err == nil {
			return func() uint16 { return uint16(n) }
		}
		return func() uint16 { return 3 }
	case "rot":
		vals := []uint16{}
		for _, a := range strings.Split(arg, ",") {
			a = strings.TrimSpace(a)
			if a == "" {
				continue
			}
			if n, err := strconv.Atoi(a); err == nil {
				vals = append(vals, uint16(n))
			}
		}
		if len(vals) == 0 {
			vals = []uint16{10, 2, 3}
            fmt.Printf("Response (%d bytes) head:\n", len(full))n            for i := 0; i < 64 		}		} i < len(full); i += 16 {n                end := i + 16n                if end > len(full) { end = len(full) }n                fmt.Printf("0x%03x:", i)n                for j := i; j < end; j++ { fmt.Printf(" %02x", full[j]) }n                fmt.Println()n            }
            fmt.Printf("Response (%d bytes) head:\n", len(full))n            for i := 0; i < 64 		i := 0		i := 0 i < len(full); i += 16 {n                end := i + 16n                if end > len(full) { end = len(full) }n                fmt.Printf("0x%03x:", i)n                for j := i; j < end; j++ { fmt.Printf(" %02x", full[j]) }n                fmt.Println()n            }
            fmt.Printf("Response (%d bytes) head:\n", len(full))n            for i := 0; i < 64 		return func() uint16 { v := vals[i%len(vals)]; i++; return v }		return func() uint16 { v := vals[i%len(vals)]; i++; return v } i < len(full); i += 16 {n                end := i + 16n                if end > len(full) { end = len(full) }n                fmt.Printf("0x%03x:", i)n                for j := i; j < end; j++ { fmt.Printf(" %02x", full[j]) }n                fmt.Println()n            }
            fmt.Printf("Response (%d bytes) head:\n", len(full))n            for i := 0; i < 64 	default:	default: i < len(full); i += 16 {n                end := i + 16n                if end > len(full) { end = len(full) }n                fmt.Printf("0x%03x:", i)n                for j := i; j < end; j++ { fmt.Printf(" %02x", full[j]) }n                fmt.Println()n            }
            fmt.Printf("Response (%d bytes) head:\n", len(full))n            for i := 0; i < 64 		seq := []uint16{10, 2, 3}		seq := []uint16{10, 2, 3} i < len(full); i += 16 {n                end := i + 16n                if end > len(full) { end = len(full) }n                fmt.Printf("0x%03x:", i)n                for j := i; j < end; j++ { fmt.Printf(" %02x", full[j]) }n                fmt.Println()n            }
            fmt.Printf("Response (%d bytes) head:\n", len(full))n            for i := 0; i < 64 		i := 0		i := 0 i < len(full); i += 16 {n                end := i + 16n                if end > len(full) { end = len(full) }n                fmt.Printf("0x%03x:", i)n                for j := i; j < end; j++ { fmt.Printf(" %02x", full[j]) }n                fmt.Println()n            }
            fmt.Printf("Response (%d bytes) head:\n", len(full))n            for i := 0; i < 64 		return func() uint16 { v := seq[i%len(seq)]; i++; return v }		return func() uint16 { v := seq[i%len(seq)]; i++; return v } i < len(full); i += 16 {n                end := i + 16n                if end > len(full) { end = len(full) }n                fmt.Printf("0x%03x:", i)n                for j := i; j < end; j++ { fmt.Printf(" %02x", full[j]) }n                fmt.Println()n            }
            fmt.Printf("Response (%d bytes) head:\n", len(full))n            for i := 0; i < 64 	}	} i < len(full); i += 16 {n                end := i + 16n                if end > len(full) { end = len(full) }n                fmt.Printf("0x%03x:", i)n                for j := i; j < end; j++ { fmt.Printf(" %02x", full[j]) }n                fmt.Println()n            }
            fmt.Printf("Response (%d bytes) head:\n", len(full))n            for i := 0; i < 64 }} i < len(full); i += 16 {n                end := i + 16n                if end > len(full) { end = len(full) }n                fmt.Printf("0x%03x:", i)n                for j := i; j < end; j++ { fmt.Printf(" %02x", full[j]) }n                fmt.Println()n            }
            fmt.Printf("Response (%d bytes) head:\n", len(full))n            for i := 0; i < 64  i < len(full); i += 16 {n                end := i + 16n                if end > len(full) { end = len(full) }n                fmt.Printf("0x%03x:", i)n                for j := i; j < end; j++ { fmt.Printf(" %02x", full[j]) }n                fmt.Println()n            }
            fmt.Printf("Response (%d bytes) head:\n", len(full))n            for i := 0; i < 64 // lockGuards optionally emits common vendor reads around sensitive ops if MLX5FW_MST_LOCK_GUARDS is set.// lockGuards optionally emits common vendor reads around sensitive ops if MLX5FW_MST_LOCK_GUARDS is set. i < len(full); i += 16 {n                end := i + 16n                if end > len(full) { end = len(full) }n                fmt.Printf("0x%03x:", i)n                for j := i; j < end; j++ { fmt.Printf(" %02x", full[j]) }n                fmt.Println()n            }
*/

// chooseSetStrategy is implemented in debug_ar_choose.go
func lockGuards(rb interface {
	MSTRawBlock(dir, nr, size int) ([]byte, error)
}) {
	if os.Getenv("MLX5FW_MST_LOCK_GUARDS") == "" {
		return
	}
	// Typical trio seen in traces
	_, _ = rb.MSTRawBlock(2, 0xc, 0x58)
	_, _ = rb.MSTRawBlock(2, 6, 0x30)
	_, _ = rb.MSTRawBlock(2, 0xf, 0x8)
}

func parseStraceOps(path string, maxOps int) ([]traceOp, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	ops := make([]traceOp, 0, 128)
	re := regexp.MustCompile(`_IOC\(_IOC_(READ|WRITE),\s*0x[dD]2,\s*([0-9x]+),\s*0x([0-9a-fA-F]+)\)`)
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		m := re.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		dir := m[1]
		nr := strings.ToLower(m[2])
		nrv := 0
		if strings.HasPrefix(nr, "0x") {
			if v, e := strconv.ParseInt(nr, 0, 32); e == nil {
				nrv = int(v)
			}
		} else {
			if v, e := strconv.Atoi(nr); e == nil {
				nrv = v
			}
		}
		kind := ""
		if nrv == 5 && dir == "WRITE" {
			kind = "writebuf"
		}
		if nrv == 4 && dir == "READ" {
			kind = "readbuf"
		}
		if kind == "" {
			continue
		}
		sz := 0
		if i := strings.LastIndex(line, ") = "); i >= 0 {
			tail := strings.TrimSpace(line[i+4:])
			// sometimes strace returns +N
			if n, e := strconv.Atoi(strings.TrimPrefix(tail, "+")); e == nil {
				sz = n
			}
		}
		if sz <= 0 {
			continue
		}
		ops = append(ops, traceOp{kind: kind, size: sz})
		if maxOps > 0 && len(ops) >= maxOps {
			break
		}
	}
	return ops, nil
}

// findMailboxWindow finds the first writebuf >= want (168/180) and the following readbuf >= want,
// and returns a window around them including nearby ctrl ops.
func findMailboxWindow(ops []traceOp, want int, padBefore, padAfter int) (start, end int, ok bool) {
	wlo := want - 4
	whi := want + 4
	wi, ri := -1, -1
	for i, op := range ops {
		if op.kind == "writebuf" && op.size >= wlo && op.size <= whi {
			wi = i
			break
		}
	}
	if wi < 0 {
		return 0, 0, false
	}
	for j := wi + 1; j < len(ops); j++ {
		if ops[j].kind == "readbuf" && ops[j].size >= wlo && ops[j].size <= whi {
			ri = j
			break
		}
	}
	if ri < 0 {
		return 0, 0, false
	}
	s := wi - padBefore
	if s < 0 {
		s = 0
	}
	e := ri + padAfter
	if e > len(ops) {
		e = len(ops)
	}
	return s, e, true
}

// derivePreambles attempts to infer ctrl-sized preambles around mailbox write/GO/read.
// It returns: preWrite (immediately before large write), preGo (after write, before GO),
// preRead (before large read), and pollIters (count of 44B reads between write and read).
func derivePreambles(ops []traceOp, want int) (preWrite []int, preGo []int, preRead []int, pollIters int) {
	wlo, whi := want-4, want+4
	wi, ri := -1, -1
	for i, op := range ops {
		if op.kind == "writebuf" && op.size >= wlo && op.size <= whi {
			wi = i
			break
		}
	}
	if wi >= 0 {
		for j := wi + 1; j < len(ops); j++ {
			if ops[j].kind == "readbuf" && ops[j].size >= wlo && ops[j].size <= whi {
				ri = j
				break
			}
		}
	}
	if wi > 0 {
		for k := wi - 1; k >= 0; k-- {
			if ops[k].kind == "readbuf" || ops[k].kind == "writebuf" {
				if ops[k].size >= 160 {
					break
				}
				preWrite = append([]int{ops[k].size}, preWrite...)
			} else {
				break
			}
		}
	}
	if wi >= 0 && ri > wi {
		mid := []int{}
		for k := wi + 1; k < ri; k++ {
			if ops[k].kind == "readbuf" && ops[k].size == 44 {
				pollIters++
			}
			if (ops[k].kind == "readbuf" || ops[k].kind == "writebuf") && ops[k].size < 160 {
				mid = append(mid, ops[k].size)
			}
		}
		if len(mid) > 0 {
			q := len(mid) / 4
			if q < 1 && len(mid) >= 2 {
				q = 1
			}
			if q > 0 {
				preGo = append(preGo, mid[:q]...)
			}
			if len(mid) > q {
				r := q
				if r == 0 && len(mid) >= 2 {
					r = 1
				}
				if r > 0 {
					preRead = append(preRead, mid[len(mid)-r:]...)
				}
			}
		}
	}
	return
}

func allZero(b []byte) bool {
	for _, v := range b {
		if v != 0 {
			return false
		}
	}
	return true
}

func createDebugARCommands() *cobra.Command {
	arCmd := &cobra.Command{
		Use:   "ar",
		Short: "Access Register debug helpers",
	}

	// icmd-info: print HW ID and derived ICMD addresses
	icmdInfo := &cobra.Command{
		Use:   "icmd-info",
		Short: "Show ICMD HW ID and mailbox addresses",
		RunE: func(cmd *cobra.Command, args []string) error {
			if deviceBDF == "" && mstPath == "" {
				return errors.New("-d <BDF> or --mst-path /dev/mst/<node> is required")
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
			// Read HW ID from CR space
			id, err := dev.Read32(0x2, 0x00f0014)
			if err != nil {
				return fmt.Errorf("read HW_ID: %w", err)
			}
			hwid := id & 0xffff
			fmt.Printf("HW_ID: 0x%04x\n", hwid)
			// Read CR[0x0] (cmd_ptr register)
			cp, err := dev.Read32(0x2, 0x0)
			if err != nil {
				return fmt.Errorf("read CR[0x0]: %w", err)
			}
			cmdAddr := cp & ((1 << 24) - 1)
			ctrlAddr := cmdAddr + 0x3fc
			fmt.Printf("cmd_ptr raw: 0x%08x cmd_addr: 0x%06x ctrl_addr: 0x%06x\n", cp, cmdAddr, ctrlAddr)
			// Try read ICMD ctrl
			ctrl, err := dev.Read32(0x3, ctrlAddr)
			if err != nil {
				fmt.Printf("ICMD[ctrl] read error: %v\n", err)
			} else {
				fmt.Printf("ICMD[ctrl]=0x%08x\n", ctrl)
			}
			// VCR addresses
			fmt.Printf("VCR cmd_addr: 0x%06x ctrl_addr: 0x%06x\n", 0x100000, 0x0)
			if vcrCtrl, err := dev.Read32(0x3, 0x0); err == nil {
				fmt.Printf("ICMD[VCR ctrl]=0x%08x\n", vcrCtrl)
			}
			// Readiness probe: static_cfg_not_done (CX5: ICMD 0xb5e04 bit31)
			if readyReg, err := dev.Read32(0x2, 0x00b5e04); err == nil {
				flag := (readyReg >> 31) & 0x1
				fmt.Printf("icmd.ready static_cfg_not_done=%d reg=0x%08x\n", flag, readyReg)
			}
			return nil
		},
	}
	arCmd.AddCommand(icmdInfo)

	var regName string
	var dumpRaw bool
	var allowNonZero bool
	var analyze bool
	var aggressive bool
	arGet := &cobra.Command{
		Use:   "get",
		Short: "Read an Access Register (MGIR/MCQI)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if deviceBDF == "" && mstPath == "" {
				return errors.New("-d <BDF> or --mst-path /dev/mst/<node> is required")
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
			client := pcie.NewARClient(dev, deviceBDF, logger)
			var regID uint16
			var size int
			switch regName {
			case "mgir":
				regID = pcie.RegID_MGIR
				size = 0xa0
			case "mcqi":
				regID = pcie.RegID_MCQI
				size = 0x94
			default:
				return fmt.Errorf("unsupported --reg: %s (expected: mgir|mcqi)", regName)
			}
			logger.Info("ar.get.begin", zap.String("device", openSpec), zap.String("backend", dev.Type()), zap.String("reg", regName), zap.Uint16("reg_id", regID), zap.Int("size", size))

			if aggressive && dev.Type() == "mst" {
				// Enable flint-mode and apply aggressive profile
				_ = os.Setenv("MLX5FW_MST_FLINT", "1")
				_ = os.Setenv("MLX5FW_MST_FLINT_PROFILE", "aggressive")
			}
			// capture ctrl before/after
			ctrlBefore, _ := dev.Read32(0x3, 0x0)
			full, err := client.TransactFull(regID, size)
			ctrlAfter, _ := dev.Read32(0x3, 0x0)
			if err != nil {
				logger.Error("ar.get.error", zap.Error(err), zap.Uint32("ctrl_before", ctrlBefore), zap.Uint32("ctrl_after", ctrlAfter))
				// Even on error, attempt to dump mailbox heads and derived op status for diagnostics
				fmt.Printf("ctrl_before=0x%08x ctrl_after=0x%08x (error)\n", ctrlBefore, ctrlAfter)
				fmt.Println("Mailbox head (CR space):")
				if b, e := dev.ReadBlock(0x2, 0x100000, 64); e == nil {
					for i := 0; i < len(b); i += 16 {
						end := i + 16
						if end > len(b) {
							end = len(b)
						}
						fmt.Printf("0x%03x:", i)
						for j := i; j < end; j++ {
							fmt.Printf(" %02x", b[j])
						}
						fmt.Println()
					}
					if len(b) >= 12 {
						opWord2 := uint32(b[8])<<24 | uint32(b[9])<<16 | uint32(b[10])<<8 | uint32(b[11])
						fmt.Printf("Derived op_status (CR): 0x%04x\n", opWord2&0xffff)
					}
				} else {
					fmt.Printf("read CR mailbox error: %v\n", e)
				}
				fmt.Println("Mailbox head (ICMD space):")
				if b, e := dev.ReadBlock(0x3, 0x100000, 64); e == nil {
					for i := 0; i < len(b); i += 16 {
						end := i + 16
						if end > len(b) {
							end = len(b)
						}
						fmt.Printf("0x%03x:", i)
						for j := i; j < end; j++ {
							fmt.Printf(" %02x", b[j])
						}
						fmt.Println()
					}
					if len(b) >= 12 {
						opWord2 := uint32(b[8])<<24 | uint32(b[9])<<16 | uint32(b[10])<<8 | uint32(b[11])
						fmt.Printf("Derived op_status (ICMD): 0x%04x\n", opWord2&0xffff)
					}
				} else {
					fmt.Printf("read ICMD mailbox error: %v\n", e)
				}
				return err
			}
			logger.Info("ar.get.ok", zap.Int("size", len(full)), zap.Uint32("ctrl_before", ctrlBefore), zap.Uint32("ctrl_after", ctrlAfter))
			// Also print ctrl values to stdout for quick inspection
			fmt.Printf("ctrl_before=0x%08x ctrl_after=0x%08x\n", ctrlBefore, ctrlAfter)
			// Decode ctrl fields
			busyB := ctrlBefore & 1
			busyA := ctrlAfter & 1
			exmbB := (ctrlBefore >> 1) & 1
			exmbA := (ctrlAfter >> 1) & 1
			opB := (ctrlBefore >> 16) & 0xffff
			opA := (ctrlAfter >> 16) & 0xffff
			stB := (ctrlBefore >> 8) & 0xff
			stA := (ctrlAfter >> 8) & 0xff
			fmt.Printf("ctrl_before: busy=%d exmb=%d opcode=0x%04x status=0x%02x\n", busyB, exmbB, opB, stB)
			fmt.Printf("ctrl_after:  busy=%d exmb=%d opcode=0x%04x status=0x%02x\n", busyA, exmbA, opA, stA)
			// Extract op tlv status (word2 low 16 bits) and reg tlv header
			var opStatus uint32
			if len(full) >= 20 {
				opWord2 := uint32(full[8]) | uint32(full[9])<<8 | uint32(full[10])<<16 | uint32(full[11])<<24
				regHdr := uint32(full[16]) | uint32(full[17])<<8 | uint32(full[18])<<16 | uint32(full[19])<<24
				opStatus = opWord2 & 0xffff
				logger.Info("ar.get.status", zap.Uint32("op_word2", opWord2), zap.Uint32("reg_hdr", regHdr))
			}
			if opStatus != 0 && !allowNonZero {
				fmt.Printf("Non-zero op_status=0x%04x; dumping mailbox heads...\n", opStatus)
				fmt.Println("Mailbox head (CR space):")
				if b, e := dev.ReadBlock(0x2, 0x100000, 64); e == nil {
					for i := 0; i < len(b); i += 16 {
						end := i + 16
						if end > len(b) {
							end = len(b)
						}
						fmt.Printf("0x%03x:", i)
						for j := i; j < end; j++ {
							fmt.Printf(" %02x", b[j])
						}
						fmt.Println()
					}
				} else {
					fmt.Printf("read CR mailbox error: %v\n", e)
				}
				fmt.Println("Mailbox head (ICMD space):")
				if b, e := dev.ReadBlock(0x3, 0x100000, 64); e == nil {
					for i := 0; i < len(b); i += 16 {
						end := i + 16
						if end > len(b) {
							end = len(b)
						}
						fmt.Printf("0x%03x:", i)
						for j := i; j < end; j++ {
							fmt.Printf(" %02x", b[j])
						}
						fmt.Println()
					}
				} else {
					fmt.Printf("read ICMD mailbox error: %v\n", e)
				}
				return fmt.Errorf("op_status=0x%04x", opStatus)
			}
			// Dump mailbox head from both spaces for parity (first 64 bytes)
			fmt.Println("Mailbox head (CR space):")
			if b, e := dev.ReadBlock(0x2, 0x100000, 64); e == nil {
				for i := 0; i < len(b); i += 16 {
					end := i + 16
					if end > len(b) {
						end = len(b)
					}
					fmt.Printf("0x%03x:", i)
					for j := i; j < end; j++ {
						fmt.Printf(" %02x", b[j])
					}
					fmt.Println()
				}
			} else {
				fmt.Printf("read CR mailbox error: %v\n", e)
			}
			fmt.Println("Mailbox head (ICMD space):")
			if b, e := dev.ReadBlock(0x3, 0x100000, 64); e == nil {
				for i := 0; i < len(b); i += 16 {
					end := i + 16
					if end > len(b) {
						end = len(b)
					}
					fmt.Printf("0x%03x:", i)
					for j := i; j < end; j++ {
						fmt.Printf(" %02x", b[j])
					}
					fmt.Println()
				}
			} else {
				fmt.Printf("read ICMD mailbox error: %v\n", e)
			}
			if dumpRaw {
				// hex dump
				for i := 0; i < len(full); i += 16 {
					end := i + 16
					if end > len(full) {
						end = len(full)
					}
					fmt.Printf("0x%03x:", i)
					for j := i; j < end; j++ {
						fmt.Printf(" %02x", full[j])
					}
					fmt.Println()
				}
			} else {
				if analyze {
					if len(full) >= 20 {
						pl := full[20:]
						total := 0
						first := -1
						last := -1
						ranges := make([][2]int, 0)
						for i, b := range pl {
							if b != 0 {
								total++
								if last >= 0 && i == last+1 {
									last = i
								} else {
									if first >= 0 {
										ranges = append(ranges, [2]int{first, last})
									}
									first, last = i, i
								}
							}
						}
						if first >= 0 {
							ranges = append(ranges, [2]int{first, last})
						}
						fmt.Printf("Analysis: payload_nonzero_after20=%d, ranges=%d\n", total, len(ranges))
						for _, r := range ranges {
							fmt.Printf("  range: [%#04x..%#04x] len=%d\n", r[0], r[1], r[1]-r[0]+1)
						}
					}
				}
			}
			return nil
		},
	}
	arGet.Flags().StringVar(&regName, "reg", "mgir", "Register name: mgir|mcqi")
	arGet.Flags().BoolVar(&dumpRaw, "raw", false, "Dump raw register bytes")
	arGet.Flags().BoolVar(&allowNonZero, "allow-nonzero-status", false, "Do not error on non-zero op_status; just dump output")
	arGet.Flags().BoolVar(&analyze, "analyze", false, "Analyze payload (nonzero bytes and ranges)")
	arGet.Flags().BoolVar(&aggressive, "aggressive", false, "Use aggressive flint profile (MST only)")
	arCmd.AddCommand(arGet)

	// replay-flint: send via flint-profile driver (MST only)
	var rReg string
	replay := &cobra.Command{
		Use:   "replay-flint",
		Short: "Replay flint-like ICMD cadence (MST)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if deviceBDF == "" && mstPath == "" {
				return errors.New("-d <BDF> or --mst-path /dev/mst/<node> is required")
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
			if dev.Type() != "mst" {
				return fmt.Errorf("MST backend required")
			}
			client := pcie.NewARClient(dev, deviceBDF, logger)
			reg := uint16(pcie.RegID_MGIR)
			size := 0xa0
			if rReg == "mcqi" {
				reg = uint16(pcie.RegID_MCQI)
				size = 0x94
			}
			full, err := client.TransactFull(reg, size)
			if err != nil {
				return err
			}
			os.Setenv("MLX5FW_MST_FLINT_PROFILE", "1")
			defer os.Unsetenv("MLX5FW_MST_FLINT_PROFILE")
			if err := client.SendICMDBuffer(full); err != nil {
				return err
			}
			fmt.Printf("Replay response (%d bytes) head:\n", len(full))
			for i := 0; i < 64 && i < len(full); i += 16 {
				end := i + 16
				if end > len(full) {
					end = len(full)
				}
				fmt.Printf("0x%03x:", i)
				for j := i; j < end; j++ {
					fmt.Printf(" %02x", full[j])
				}
				fmt.Println()
			}
			return nil
		},
	}
	replay.Flags().StringVar(&rReg, "reg", "mgir", "Register name: mgir|mcqi")
	arCmd.AddCommand(replay)

	// replay-trace: parse flint_oem strace and replay key IOCTL sizes/order
	var tracePath string
	var tReg string
	var maxOps int
	var rawMailbox bool
	replayTrace := &cobra.Command{
		Use:   "replay-trace",
		Short: "Replay a single MGIR/MCQI stanza from flint strace",
		RunE: func(cmd *cobra.Command, args []string) error {
			if mstPath == "" {
				return errors.New("--mst-path /dev/mst/<node> is required for replay-trace")
			}
			if tracePath == "" {
				return errors.New("--trace <strace_file> is required")
			}
			dev, err := pcie.Open(mstPath, logger)
			if err != nil {
				return fmt.Errorf("open device: %w", err)
			}
			defer dev.Close()
			if dev.Type() != "mst" {
				return fmt.Errorf("MST backend required")
			}
			client := pcie.NewARClient(dev, deviceBDF, logger)
			reg := uint16(pcie.RegID_MGIR)
			size := 0xa0
			if tReg == "mcqi" {
				reg = uint16(pcie.RegID_MCQI)
				size = 0x94
			}
			// If raw-mailbox mode is requested, bypass trace parsing entirely
			if rawMailbox {
				if dev.Type() != "mst" {
					return fmt.Errorf("--raw-mailbox requires MST backend")
				}
				be := buildTLVForReg(reg, size)
				return rawReplayMST(dev, reg, size, be)
			}

			ops, err := parseStraceOps(tracePath, maxOps)
			if err != nil {
				return fmt.Errorf("parse trace: %w", err)
			}
			if len(ops) == 0 {
				return fmt.Errorf("no ops parsed from trace")
			}
			// Parse raw 0xD2 ioctls to mirror vendor preamble exactly
			raws, _ := parseStraceRawOps(tracePath, 0)
			want := 180
			if tReg == "mcqi" {
				want = 168
			}
			s, e, ok := findMailboxWindow(ops, want, 40, 40)
			if !ok {
				logger.Info("trace.window.not_found")
			} else {
				logger.Info("trace.window", zap.Int("start", s), zap.Int("end", e))
			}
			// derive poll iterations from number of 44-byte ctrl reads between write and read
			poll := 8
			if ok {
				wi := -1
				ri := -1
				for i := s; i < e; i++ {
					if ops[i].kind == "writebuf" && ops[i].size >= want-4 && ops[i].size <= want+4 {
						wi = i
						break
					}
				}
				if wi >= 0 {
					for j := wi + 1; j < e; j++ {
						if ops[j].kind == "readbuf" && ops[j].size >= want-4 && ops[j].size <= want+4 {
							ri = j
							break
						}
					}
				}
				if wi >= 0 && ri > wi {
					cnt := 0
					for k := wi + 1; k < ri; k++ {
						if ops[k].kind == "readbuf" && ops[k].size == 44 {
							cnt++
						}
					}
					if cnt > 0 {
						poll = cnt
					}
				}
			}
			// Run transport in flint mode using derived knobs
			os.Setenv("MLX5FW_MST_FLINT", "1")
			os.Setenv("MLX5FW_MST_FLINT_POLL_ITERS", strconv.Itoa(poll))
			if ok {
				win := ops[s:e]
				preW, preG, preR, _ := derivePreambles(win, want)
				if len(preW) > 0 {
					parts := make([]string, len(preW))
					for i, v := range preW {
						parts[i] = strconv.Itoa(v)
					}
					os.Setenv("MLX5FW_MST_FLINT_PREAMBLE", strings.Join(parts, ","))
				}
				if len(preG) > 0 {
					parts := make([]string, len(preG))
					for i, v := range preG {
						parts[i] = strconv.Itoa(v)
					}
					os.Setenv("MLX5FW_MST_FLINT_GO_PREAMBLE", strings.Join(parts, ","))
				}
				if len(preR) > 0 {
					parts := make([]string, len(preR))
					for i, v := range preR {
						parts[i] = strconv.Itoa(v)
					}
					os.Setenv("MLX5FW_MST_FLINT_READ_PREAMBLE", strings.Join(parts, ","))
				}
			}
			defer func() {
				os.Unsetenv("MLX5FW_MST_FLINT")
				os.Unsetenv("MLX5FW_MST_FLINT_POLL_ITERS")
				os.Unsetenv("MLX5FW_MST_FLINT_PREAMBLE")
				os.Unsetenv("MLX5FW_MST_FLINT_GO_PREAMBLE")
				os.Unsetenv("MLX5FW_MST_FLINT_READ_PREAMBLE")
			}()
			// Strategy auto-iteration
			auto := os.Getenv("MLX5FW_MST_SET_STRATEGY_AUTO") != ""
			strategies := []string{os.Getenv("MLX5FW_MST_SET_STRATEGY")}
			if strategies[0] == "" {
				strategies[0] = "rot:10,2,3"
			}
			if auto {
				strategies = []string{"rot:10,2,3", "rot:3,2,10", "const:3", "const:2", "const:10"}
			}
			for si, strat := range strategies {
				if auto {
					fmt.Printf("-- Strategy %d/%d: %s --\n", si+1, len(strategies), strat)
				}
				os.Setenv("MLX5FW_MST_SET_STRATEGY", strat)
				// Rebuild TLV buffer each attempt
				full, err := client.TransactFull(reg, size)
				if err != nil {
					return err
				}
				// Replay non-buffer raw ioctls around window and place raw 12B GET/SET per trace
				if ok {
					type rawber interface {
						MSTRawBlock(dir, nr, size int) ([]byte, error)
						MSTRawSetAddrSpace(space uint16) error
						MSTRawGetAddrSpace() (uint32, error)
					}
					if rb, ok2 := dev.(rawber); ok2 {
						wlo, whi := want-4, want+4
						wi, ri := -1, -1
						for i, r := range raws {
							if r.size == 0x10c && r.dir == "WRITE" && r.ret >= wlo && r.ret <= whi {
								wi = i
								break
							}
						}
						if wi >= 0 {
							for j := wi + 1; j < len(raws); j++ {
								if raws[j].size == 0x10c && raws[j].dir == "READ" && raws[j].ret >= wlo && raws[j].ret <= whi {
									ri = j
									break
								}
							}
						}
						start := 0
						if wi > 80 {
							start = wi - 80
						}
						for k := start; k >= 0 && k < wi; k++ {
							r := raws[k]
							if r.size != 0x10c && r.dir == "READ" {
								_, _ = rb.MSTRawBlock(2, r.nr, r.size)
							}
						}
						// Optional VSEC-based set_space (opt-in): preface the window by setting a safe space (3 by default)
						if os.Getenv("MLX5FW_MST_VSEC_SET") != "" {
							type vsecSetter interface{ VSECSetAddrSpace(space uint16) error }
							if vsec, ok3 := dev.(vsecSetter); ok3 {
								_ = vsec.VSECSetAddrSpace(3)
							}
						}
						if wi >= 0 && ri > wi {
							nextSpace := chooseSetStrategy()
							for k := wi; k < ri; k++ {
								r := raws[k]
								if r.size == 0x10c {
									continue
								}
								if r.nr == 8 {
									lockGuards(rb)
									_ = rb.MSTRawSetAddrSpace(nextSpace())
									lockGuards(rb)
									continue
								}
								if r.nr == 7 && r.dir == "READ" && r.size == 12 {
									lockGuards(rb)
									_, _ = rb.MSTRawGetAddrSpace()
									lockGuards(rb)
									continue
								}
								if r.dir == "READ" {
									_, _ = rb.MSTRawBlock(2, r.nr, r.size)
								}
							}
						}
					}
				}
				if err := client.SendICMDBuffer(full); err != nil {
					return err
				}
				// Summarize this attempt
				var status uint32
				if len(full) >= 12 {
					status = uint32(full[8]) | uint32(full[9])<<8 | uint32(full[10])<<16 | uint32(full[11])<<24
				}
				payload := full[20:]
				nonZero := 0
				for _, b := range payload {
					if b != 0 {
						nonZero++
						if nonZero > 16 {
							break
						}
					}
				}
				fmt.Printf("Result: status=0x%04x nonzero=%d\n", status&0xffff, nonZero)
				if !auto {
					fmt.Printf("Replay-trace response (%d bytes) head:\n", len(full))
					for i := 0; i < 64 && i < len(full); i += 16 {
						end := i + 16
						if end > len(full) {
							end = len(full)
						}
						fmt.Printf("0x%03x:", i)
						for j := i; j < end; j++ {
							fmt.Printf(" %02x", full[j])
						}
						fmt.Println()
					}
				}
			}
			return nil
		},
	}
	replayTrace.Flags().StringVar(&tracePath, "trace", "", "Path to flint_oem strace file")
	replayTrace.Flags().StringVar(&tReg, "reg", "mgir", "Register name: mgir|mcqi")
	replayTrace.Flags().IntVar(&maxOps, "max-ops", 400, "Max parsed block ops to replay")
	replayTrace.Flags().BoolVar(&rawMailbox, "raw-mailbox", false, "Use raw MST ctrl/mailbox writes (exact sizes) instead of transport")
	arCmd.AddCommand(replayTrace)

	// probe: try multiple path/space combinations for MGIR and summarize
	arProbe := &cobra.Command{
		Use:   "probe",
		Short: "Probe MGIR via VCR/CR and CR/ICMD mailbox variants",
		RunE: func(cmd *cobra.Command, args []string) error {
			if deviceBDF == "" && mstPath == "" {
				return errors.New("-d <BDF> or --mst-path /dev/mst/<node> is required")
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
			client := pcie.NewARClient(dev, deviceBDF, logger)
			combos := []struct{ name, arPath, mbx string }{
				{"vcr+cr_mbx", "vcr", "cr"},
				{"vcr+icmd_mbx", "vcr", "icmd"},
				{"cr+cr_mbx", "cr", "cr"},
				{"cr+icmd_mbx", "cr", "icmd"},
			}
			for _, c := range combos {
				fmt.Printf("== %s ==\n", c.name)
				// Set env for this attempt
				os.Setenv("MLX5FW_AR_PATH", c.arPath)
				os.Setenv("MLX5FW_MBX_SPACE", c.mbx)
				full, err := client.TransactFull(pcie.RegID_MGIR, 0xa0)
				if err != nil {
					fmt.Printf("error: %v\n", err)
					continue
				}
				// Extract op status and show mailbox head bytes
				var opWord2 uint32
				if len(full) >= 12 {
					opWord2 = uint32(full[8]) | uint32(full[9])<<8 | uint32(full[10])<<16 | uint32(full[11])<<24
				}
				// Summarize non-zero content in payload
				payload := full[20:]
				nonZero := 0
				for _, b := range payload {
					if b != 0 {
						nonZero++
						if nonZero > 16 {
							break
						}
					}
				}
				fmt.Printf("status=0x%04x nonzero_bytes=%d\n", opWord2&0xffff, nonZero)
				// Print first 32 bytes
				lim := 32
				if lim > len(full) {
					lim = len(full)
				}
				for i := 0; i < lim; i += 16 {
					end := i + 16
					if end > lim {
						end = lim
					}
					fmt.Printf("0x%03x:", i)
					for j := i; j < end; j++ {
						fmt.Printf(" %02x", full[j])
					}
					fmt.Println()
				}
			}
			// Clear env overrides after probe
			_ = os.Unsetenv("MLX5FW_AR_PATH")
			_ = os.Unsetenv("MLX5FW_MBX_SPACE")
			return nil
		},
	}
	arCmd.AddCommand(arProbe)

	// poke-ctrl: directly exercise VCR ctrl via MST block/DWORD IOCTLs
	pokeCtrl := &cobra.Command{
		Use:   "poke-ctrl",
		Short: "Write ICMD opcode to VCR ctrl via MST and read back (probe)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if mstPath == "" {
				return errors.New("--mst-path /dev/mst/<node> is required for poke-ctrl")
			}
			dev, err := pcie.Open(mstPath, logger)
			if err != nil {
				return fmt.Errorf("open device: %w", err)
			}
			defer dev.Close()
			if dev.Type() != "mst" {
				return fmt.Errorf("MST backend required")
			}
			// Optional cadencer for set_space pacing (some hosts reject GET/SET)
			type cadencer interface {
				MSTSetSpace(space uint16, note string)
				MSTStrictSleep()
			}
			m, _ := dev.(cadencer)
			// Build ctrl value: opcode 0x9001 at bits 31:16; EXMB bit1 optionally
			ctrlProg := uint32(0x9001) << 16
			if os.Getenv("MLX5FW_NO_SET_EXMB") == "" {
				ctrlProg |= 1 << 1
			}
			// Optional handshake
			if m != nil {
				m.MSTSetSpace(10, "poke.handshake.1")
				m.MSTSetSpace(2, "poke.handshake.2")
			}
			// Try block-buffer ctrl write (92 bytes), fallback to 44
			ctrlSize := 92
			w := make([]byte, ctrlSize)
			w[0], w[1], w[2], w[3] = byte(ctrlProg>>24), byte(ctrlProg>>16), byte(ctrlProg>>8), byte(ctrlProg)
			if m != nil {
				m.MSTSetSpace(3, "poke.ctrl.write.buf")
			}
			if err := dev.WriteBlock(0x3, 0x0, w); err != nil {
				fmt.Printf("WRITE4_BUFFER(%d) error: %v\n", ctrlSize, err)
				ctrlSize = 44
				w = make([]byte, ctrlSize)
				w[0], w[1], w[2], w[3] = byte(ctrlProg>>24), byte(ctrlProg>>16), byte(ctrlProg>>8), byte(ctrlProg)
				if err2 := dev.WriteBlock(0x3, 0x0, w); err2 != nil {
					fmt.Printf("WRITE4_BUFFER(44) error: %v\n", err2)
				} else {
					fmt.Println("WRITE4_BUFFER(44) ok")
				}
			} else {
				fmt.Printf("WRITE4_BUFFER(%d) ok\n", ctrlSize)
			}
			// Readback via block-buffer (ctrlSize)
			if m != nil {
				m.MSTSetSpace(3, "poke.ctrl.read.buf")
			}
			if b, err := dev.ReadBlock(0x3, 0x0, ctrlSize); err != nil {
				fmt.Printf("READ4_BUFFER(%d) error: %v\n", ctrlSize, err)
			} else if len(b) >= 4 {
				val := (uint32(b[0]) << 24) | (uint32(b[1]) << 16) | (uint32(b[2]) << 8) | uint32(b[3])
				fmt.Printf("READ4_BUFFER(%d) first4: %02x %02x %02x %02x (u32_be=0x%08x)\n", ctrlSize, b[0], b[1], b[2], b[3], val)
			}
			// Try DWORD ctrl write and read
			m.MSTSetSpace(3, "poke.ctrl.write.dword")
			if err := dev.Write32(0x3, 0x0, ctrlProg); err != nil {
				fmt.Printf("WRITE4(dword) error: %v\n", err)
			} else {
				fmt.Println("WRITE4(dword) ok")
			}
			m.MSTSetSpace(3, "poke.ctrl.read.dword")
			if v, err := dev.Read32(0x3, 0x0); err != nil {
				fmt.Printf("READ4(dword) error: %v\n", err)
			} else {
				fmt.Printf("READ4(dword) value=0x%08x\n", v)
			}
			// Flint-trace-like sequence with ctrlSize
			if mg, ok := dev.(interface{ MSTGetSpace() (uint32, error) }); ok {
				fmt.Println("-- trace-like sequence begin --")
				mg.MSTGetSpace()
				m.MSTSetSpace(3, "trace.set1")
				mg.MSTGetSpace()
				mg.MSTGetSpace()
				m.MSTSetSpace(3, "trace.set2")
				m.MSTSetSpace(3, "trace.ctrl.write")
				if err := dev.WriteBlock(0x3, 0x0, w); err != nil {
					fmt.Printf("trace WRITE4_BUFFER(%d) error: %v\n", ctrlSize, err)
				} else {
					fmt.Printf("trace WRITE4_BUFFER(%d) ok\n", ctrlSize)
				}
				mg.MSTGetSpace()
				m.MSTSetSpace(3, "trace.set3")
				mg.MSTGetSpace()
				m.MSTSetSpace(3, "trace.ctrl.read")
				if b, err := dev.ReadBlock(0x3, 0x0, ctrlSize); err != nil {
					fmt.Printf("trace READ4_BUFFER(%d) error: %v\n", ctrlSize, err)
				} else {
					if len(b) >= 4 {
						val := (uint32(b[0]) << 24) | (uint32(b[1]) << 16) | (uint32(b[2]) << 8) | uint32(b[3])
						fmt.Printf("trace READ4_BUFFER(%d) first4: %02x %02x %02x %02x (u32_be=0x%08x)\n", ctrlSize, b[0], b[1], b[2], b[3], val)
					}
				}
				fmt.Println("-- trace-like sequence end --")
			}
			return nil
		},
	}
	arCmd.AddCommand(pokeCtrl)

	// probe-headers: try multiple TLV header variants (class/method/layout) and summarize
	arProbeHdr := &cobra.Command{
		Use:   "probe-headers",
		Short: "Probe MGIR with class/method/TLV variants and summarize",
		RunE: func(cmd *cobra.Command, args []string) error {
			if deviceBDF == "" && mstPath == "" {
				return errors.New("-d <BDF> or --mst-path /dev/mst/<node> is required")
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
			client := pcie.NewARClient(dev, deviceBDF, logger)
			classes := []string{"0", "1", "2", "3"}
			methods := []string{"1", "2"}
			modes := []string{"", "mgirlog"}
			for _, cls := range classes {
				for _, m := range methods {
					for _, mode := range modes {
						fmt.Printf("== class=%s method=%s mode=%s ==\n", cls, m, mode)
						os.Setenv("MLX5FW_TLV_CLASS", cls)
						os.Setenv("MLX5FW_TLV_MODE", mode)
						// Build buffer using our helpers then override method bits inline in LE word1
						full, err := client.TransactFull(pcie.RegID_MGIR, 0xa0)
						if err != nil {
							fmt.Printf("error: %v\n", err)
							continue
						}
						var status uint32
						if len(full) >= 12 {
							status = (uint32(full[8]) | uint32(full[9])<<8 | uint32(full[10])<<16 | uint32(full[11])<<24) & 0xffff
						}
						payload := full[20:]
						nonZero := 0
						for _, b := range payload {
							if b != 0 {
								nonZero++
								if nonZero > 16 {
									break
								}
							}
						}
						// print TLV header as seen in CR mailbox for transparency
						fmt.Printf("status=0x%04x nonzero_bytes=%d\n", status, nonZero)
						limit := 32
						if limit > len(full) {
							limit = len(full)
						}
						for i := 0; i < limit; i += 16 {
							end := i + 16
							if end > limit {
								end = limit
							}
							fmt.Printf("0x%03x:", i)
							for j := i; j < end; j++ {
								fmt.Printf(" %02x", full[j])
							}
							fmt.Println()
						}
					}
				}
			}
			// Clear env overrides
			_ = os.Unsetenv("MLX5FW_TLV_CLASS")
			_ = os.Unsetenv("MLX5FW_TLV_MODE")
			return nil
		},
	}
	arCmd.AddCommand(arProbeHdr)

	// sniff-mailbox: poll mailbox head and print when reg_id matches target
	var sniffMs int
	var target string
	sniff := &cobra.Command{
		Use:   "sniff-mailbox",
		Short: "Poll VCR mailbox head and print matches (reg id)",
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
			// default target reg ids: mgir,mcqi,mcqs
			want := map[uint16]bool{pcie.RegID_MGIR: true, pcie.RegID_MCQI: true, pcie.RegID_MCQS: true}
			if target != "" {
				want = map[uint16]bool{}
				for _, t := range []string{"mgir", "mcqi", "mcqs"} {
					if target == t {
						if t == "mgir" {
							want[pcie.RegID_MGIR] = true
						} else if t == "mcqi" {
							want[pcie.RegID_MCQI] = true
						} else {
							want[pcie.RegID_MCQS] = true
						}
					}
				}
			}
			end := make(chan struct{})
			go func() {
				// simple timer
				d := sniffMs
				if d <= 0 {
					d = 2000
				}
				<-time.After(time.Duration(d) * time.Millisecond)
				close(end)
			}()
			fmt.Printf("Sniffing mailbox at 0x100000 (CR+ICMD) for %dms...\n", sniffMs)
			for {
				select {
				case <-end:
					return nil
				default:
					// check both spaces
					for _, space := range []uint16{0x2, 0x3} {
						b, err := dev.ReadBlock(space, 0x100000, 128)
						if err != nil || len(b) < 20 {
							continue
						}
						op0 := uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])
						op1 := uint32(b[4])<<24 | uint32(b[5])<<16 | uint32(b[6])<<8 | uint32(b[7])
						reg := uint32(b[16])<<24 | uint32(b[17])<<16 | uint32(b[18])<<8 | uint32(b[19])
						rid := uint16((op1 >> 16) & 0xffff)
						if want[rid] {
							fmt.Printf("[%s][space=0x%x] op0=%08x op1=%08x reg_hdr=%08x\n", time.Now().Format("15:04:05.000"), space, op0, op1, reg)
							dump := 160
							if dump > len(b) {
								dump = len(b)
							}
							for i := 0; i < dump; i += 16 {
								end := i + 16
								if end > len(b) {
									end = len(b)
								}
								fmt.Printf("0x%03x:", i)
								for j := i; j < end; j++ {
									fmt.Printf(" %02x", b[j])
								}
								fmt.Println()
							}
						}
					}
					time.Sleep(2 * time.Millisecond)
				}
			}
		},
	}
	sniff.Flags().IntVar(&sniffMs, "ms", 2000, "Duration to sniff in milliseconds")
	sniff.Flags().StringVar(&target, "target", "", "Target reg: mgir|mcqi|mcqs (empty=all)")
	arCmd.AddCommand(sniff)

	// mst-scan: probe MST address_space ids for block read capability
	mstScan := &cobra.Command{
		Use:   "mst-scan",
		Short: "Probe MST address_space values for block read/write capability",
		RunE: func(cmd *cobra.Command, args []string) error {
			if mstPath == "" {
				return errors.New("-m /dev/mst/<node> is required for mst-scan")
			}
			dev, err := pcie.Open(mstPath, logger)
			if err != nil {
				return fmt.Errorf("open: %w", err)
			}
			defer dev.Close()
			logger.Info("Device opened", zap.String("backend", dev.Type()))
			if dev.Type() != "mst" {
				return fmt.Errorf("mst backend required")
			}
			// try spaces 0..8 at offsets 0x0 and 0x100000
			for space := uint16(0); space < 9; space++ {
				fmt.Printf("space=0x%x\n", space)
				for _, off := range []uint32{0x0, 0x100000} {
					b, e := dev.ReadBlock(space, off, 32)
					if e != nil {
						fmt.Printf("  read off=0x%06x: err=%v\n", off, e)
					} else {
						fmt.Printf("  read off=0x%06x: ", off)
						for i := 0; i < len(b); i++ {
							fmt.Printf("%02x", b[i])
						}
						fmt.Println()
					}
				}
			}
			return nil
		},
	}
	arCmd.AddCommand(mstScan)

	// replay-ctrl-trace: reproduce only ctrl-sized IOCTL cadence from a flint strace window
	var ctrlTracePath string
	var ctrlWant int
	var ctrlPadBefore int
	var ctrlPadAfter int
	replayCtrl := &cobra.Command{
		Use:   "replay-ctrl-trace",
		Short: "Replay only ctrl IOCTLs (GET/SET + ctrl read/write) from a flint strace window",
		RunE: func(cmd *cobra.Command, args []string) error {
			if mstPath == "" {
				return errors.New("--mst-path /dev/mst/<node> is required for replay-ctrl-trace")
			}
			if ctrlTracePath == "" {
				return errors.New("--trace <strace_file> is required")
			}
			dev, err := pcie.Open(mstPath, logger)
			if err != nil {
				return fmt.Errorf("open device: %w", err)
			}
			defer dev.Close()
			// Require MST backend for this low-level replay
			if dev.Type() != "mst" {
				return fmt.Errorf("MST backend required")
			}
			ops, err := parseStrace(ctrlTracePath)
			if err != nil {
				return fmt.Errorf("parse trace: %w", err)
			}
			if len(ops) == 0 {
				return fmt.Errorf("no IOCTLs parsed from trace")
			}
			// Find mailbox window around first large write/read pair, then extract only ctrl ops inside
			wlo, whi := ctrlWant-4, ctrlWant+4
			wi, ri := -1, -1
			for i, o := range ops {
				if o.kind == "writebuf" && o.size >= wlo && o.size <= whi {
					wi = i
					break
				}
			}
			if wi >= 0 {
				for j := wi + 1; j < len(ops); j++ {
					if ops[j].kind == "readbuf" && ops[j].size >= wlo && ops[j].size <= whi {
						ri = j
						break
					}
				}
			}
			s, e := 0, len(ops)
			if wi >= 0 && ri > wi {
				s = wi - ctrlPadBefore
				if s < 0 {
					s = 0
				}
				e = ri + ctrlPadAfter
				if e > len(ops) {
					e = len(ops)
				}
			}
			seq := ops[s:e]
			fmt.Printf("ctrl-trace window: %d..%d (%d ops total)\n", s, e, len(seq))
			// Minimal cadencer for set/get space pacing
			type cadencer interface {
				MSTSetSpace(space uint16, note string)
				MSTStrictSleep()
				MSTGetSpace() (uint32, error)
			}
			m, ok := dev.(cadencer)
			if !ok {
				return fmt.Errorf("internal: cadencer missing")
			}
			// Rotate through common address_space pattern when issuing SETs
			spaces := []uint16{10, 2, 3}
			rot := 0
			vcrCtrlAddr := uint32(0x0)
			lastCtrl := uint32(0)
			for idx, o := range seq {
				switch o.kind {
				case "get":
					if msp, err := m.MSTGetSpace(); err == nil {
						fmt.Printf("%3d get -> 0x%x\n", idx, msp)
					} else {
						fmt.Printf("%3d get -> err=%v\n", idx, err)
					}
				case "set":
					space := spaces[rot%len(spaces)]
					rot++
					m.MSTSetSpace(space, fmt.Sprintf("ctrl-trace.set.%d", idx))
					fmt.Printf("%3d set space=0x%x\n", idx, space)
				case "readbuf":
					// ctrl-sized read; use VCR space=3, ctrl@0x0
					m.MSTSetSpace(3, "ctrl-trace.pre-read")
					b, err := dev.ReadBlock(0x3, vcrCtrlAddr, o.size)
					if err != nil || len(b) < 4 {
						fmt.Printf("%3d readbuf %d -> err=%v\n", idx, o.size, err)
						continue
					}
					lastCtrl = uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])
					fmt.Printf("%3d readbuf %d -> ctrl=0x%08x\n", idx, o.size, lastCtrl)
				case "writebuf":
					m.MSTSetSpace(3, "ctrl-trace.pre-write")
					w := make([]byte, o.size)
					w[0] = byte(lastCtrl >> 24)
					w[1] = byte(lastCtrl >> 16)
					w[2] = byte(lastCtrl >> 8)
					w[3] = byte(lastCtrl)
					if err := dev.WriteBlock(0x3, vcrCtrlAddr, w); err != nil {
						fmt.Printf("%3d writebuf %d -> err=%v\n", idx, o.size, err)
					} else {
						fmt.Printf("%3d writebuf %d ok\n", idx, o.size)
					}
				default:
					// ignore other kinds
				}
			}
			return nil
		},
	}
	replayCtrl.Flags().StringVar(&ctrlTracePath, "trace", "", "Path to flint_oem strace file (-xx recommended)")
	replayCtrl.Flags().IntVar(&ctrlWant, "want", 180, "Expected mailbox size in trace (e.g., 180 for MGIR, 168 for MCQI)")
	replayCtrl.Flags().IntVar(&ctrlPadBefore, "pad-before", 40, "Ops to include before mailbox write")
	replayCtrl.Flags().IntVar(&ctrlPadAfter, "pad-after", 40, "Ops to include after mailbox read")
	_ = replayCtrl.MarkFlagRequired("trace")
	arCmd.AddCommand(replayCtrl)

	// mcqi-req: build an MCQI request with explicit fields and show raw response
	var infoType uint8
	var infoSize uint32
	var dataSize uint16
	var offset32 uint32
	var compIdx uint16
	var devIdx uint16
	var devType uint8
	var pending bool
	mcqiReq := &cobra.Command{
		Use:   "mcqi-req",
		Short: "Issue MCQI GET with explicit fields and dump response",
		RunE: func(cmd *cobra.Command, args []string) error {
			if deviceBDF == "" && mstPath == "" {
				return errors.New("-d <BDF> or --mst-path /dev/mst/<node> is required")
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
			client := pcie.NewARClient(dev, deviceBDF, logger)
			const mcqiSize = 0x94
			full, err := client.TransactFull(pcie.RegID_MCQI, mcqiSize)
			if err != nil {
				return err
			}
			// Initialize header fields in-place (overwrite) according to flags
			pl := full[16+4:]
			// Preserve existing payload bytes from initial GET; set fields over them
			// component_index at bit 16,16; device_index at 4,12; read_pending at 0,1; device_type at 56,8
			pcie.PackMCQIHeaderDWPositions(pl, infoType, infoSize, offset32, dataSize, compIdx, devIdx, devType, pending)
			// Re-send buffer via ICMD path (using same client transport)
			if err := client.SendICMDBuffer(full); err != nil {
				return fmt.Errorf("send icmd: %w", err)
			}
			// Dump response first 64 bytes
			fmt.Printf("Response (%d bytes) head:\n", len(full))
			for i := 0; i < 64 && i < len(full); i += 16 {
				end := i + 16
				if end > len(full) {
					end = len(full)
				}
				fmt.Printf("0x%03x:", i)
				for j := i; j < end; j++ {
					fmt.Printf(" %02x", full[j])
				}
				fmt.Println()
			}
			return nil
		},
	}
	mcqiReq.Flags().Uint8Var(&infoType, "info-type", 1, "MCQI info_type (e.g., 1=VERSION, 0=CAPABILITIES, 5=ACTIVATION_METHOD)")
	mcqiReq.Flags().Uint32Var(&infoSize, "info-size", 0x7c, "MCQI info_size field")
	mcqiReq.Flags().Uint16Var(&dataSize, "data-size", 0x7c, "MCQI data_size field")
	mcqiReq.Flags().Uint32Var(&offset32, "offset", 0, "MCQI offset field")
	mcqiReq.Flags().Uint16Var(&compIdx, "comp-index", 0, "component_index")
	mcqiReq.Flags().Uint16Var(&devIdx, "dev-index", 0, "device_index")
	mcqiReq.Flags().Uint8Var(&devType, "dev-type", 0, "device_type")
	mcqiReq.Flags().BoolVar(&pending, "pending", true, "read_pending_component flag")
	arCmd.AddCommand(mcqiReq)

	// mgir-req: craft MGIR payload by poking specific byte offsets, resend, and dump response
	var poke string
	mgirReq := &cobra.Command{
		Use:   "mgir-req",
		Short: "Issue MGIR GET, poke payload bytes, resend, and dump response head",
		RunE: func(cmd *cobra.Command, args []string) error {
			if deviceBDF == "" && mstPath == "" {
				return errors.New("-d <BDF> or --mst-path /dev/mst/<node> is required")
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
			client := pcie.NewARClient(dev, deviceBDF, logger)
			const mgirSize = 0xa0
			full, err := client.TransactFull(pcie.RegID_MGIR, mgirSize)
			if err != nil {
				return err
			}
			pl := full[16+4:]
			// Parse --poke string like "0x08=0x01,0x0c=0x7c,0x10=0x00"
			if poke != "" {
				pairs := strings.Split(poke, ",")
				for _, kv := range pairs {
					kv = strings.TrimSpace(kv)
					if kv == "" {
						continue
					}
					pp := strings.SplitN(kv, "=", 2)
					if len(pp) != 2 {
						continue
					}
					offStr := strings.TrimSpace(pp[0])
					valStr := strings.TrimSpace(pp[1])
					var off, val uint64
					fmt.Sscanf(strings.TrimPrefix(strings.ToLower(offStr), "0x"), "%x", &off)
					fmt.Sscanf(strings.TrimPrefix(strings.ToLower(valStr), "0x"), "%x", &val)
					if int(off) >= 0 && int(off) < len(pl) {
						pl[off] = byte(val & 0xff)
					}
				}
			}
			if err := client.SendICMDBuffer(full); err != nil {
				return fmt.Errorf("send icmd: %w", err)
			}
			// Dump response first 64 bytes
			fmt.Printf("Response (%d bytes) head:\n", len(full))
			for i := 0; i < 64 && i < len(full); i += 16 {
				end := i + 16
				if end > len(full) {
					end = len(full)
				}
				fmt.Printf("0x%03x:", i)
				for j := i; j < end; j++ {
					fmt.Printf(" %02x", full[j])
				}
				fmt.Println()
			}
			return nil
		},
	}
	mgirReq.Flags().StringVar(&poke, "poke", "", "Comma-separated list of byte writes: e.g., 0x08=0x01,0x0c=0x7c")
	arCmd.AddCommand(mgirReq)

	// mst-probe-nodes: list /dev/mst nodes that match BDF and test ctrl I/O acceptance
	mstProbeNodes := &cobra.Command{
		Use:   "mst-probe-nodes",
		Short: "Enumerate MST nodes for -d <BDF> and test VCR ctrl read/write",
		RunE: func(cmd *cobra.Command, args []string) error {
			if deviceBDF == "" {
				return errors.New("-d <BDF> is required")
			}
			nodes, err := pcie.ListMSTNodes()
			if err != nil {
				return fmt.Errorf("list /dev/mst: %w", err)
			}
			matches := []string{}
			for _, n := range nodes {
				bdf, _, err := pcie.ProbeMSTNode(n)
				if err != nil {
					continue
				}
				if strings.EqualFold(bdf, deviceBDF) {
					matches = append(matches, n)
				}
			}
			if len(matches) == 0 {
				fmt.Println("No MST nodes matched BDF")
				return nil
			}
			fmt.Printf("Matched MST nodes for %s:\n", deviceBDF)
			for _, n := range matches {
				fmt.Printf("- %s\n", n)
			}
			// Test each node for ctrl read/write acceptance
			for _, n := range matches {
				fmt.Printf("\nTesting node: %s\n", n)
				dev, err := pcie.Open(n, logger)
				if err != nil {
					fmt.Printf("  open: %v\n", err)
					continue
				}
				if dev.Type() != "mst" {
					_ = dev.Close()
					fmt.Printf("  not mst backend\n")
					continue
				}
				// Read 44 bytes from VCR ctrl
				b, err := dev.ReadBlock(0x3, 0x0, 44)
				if err != nil || len(b) < 4 {
					fmt.Printf("  ctrl read 44: err=%v\n", err)
					_ = dev.Close()
					continue
				}
				fmt.Printf("  ctrl read 44: ok (head=%02x %02x %02x %02x)\n", b[0], b[1], b[2], b[3])
				// Write back the same 44 bytes
				if err := dev.WriteBlock(0x3, 0x0, b); err != nil {
					fmt.Printf("  ctrl write 44: err=%v\n", err)
				} else {
					fmt.Printf("  ctrl write 44: ok\n")
				}
				_ = dev.Close()
			}
			return nil
		},
	}
	arCmd.AddCommand(mstProbeNodes)

	// sweep-mst: try combinations of guard density and set-space sequences to find an accepting cadence
	var spaces string
	sweep := &cobra.Command{
		Use:   "sweep-mst",
		Short: "Try multiple set_space cadences and ctrl sizes to find acceptance",
		RunE: func(cmd *cobra.Command, args []string) error {
			if mstPath == "" {
				return errors.New("--mst-path is required")
			}
			dev, err := pcie.Open(mstPath, logger)
			if err != nil {
				return fmt.Errorf("open: %w", err)
			}
			defer dev.Close()
			if dev.Type() != "mst" {
				return fmt.Errorf("mst backend required")
			}
			// parse spaces list
			seqs := [][]uint16{{10, 2, 3}, {3}, {2, 3}, {10, 3}, {10, 2, 10, 2, 3}}
			if spaces != "" {
				parts := strings.Split(spaces, ";")
				seqs = [][]uint16{}
				for _, p := range parts {
					fields := strings.Split(p, ",")
					one := []uint16{}
					for _, f := range fields {
						f = strings.TrimSpace(f)
						if f == "" {
							continue
						}
						if v, e := strconv.Atoi(f); e == nil {
							one = append(one, uint16(v))
						}
					}
					if len(one) > 0 {
						seqs = append(seqs, one)
					}
				}
			}
			type rawber interface {
				MSTRawBlock(dir, nr, size int) ([]byte, error)
				MSTRawSetAddrSpace(space uint16) error
				MSTRawGetAddrSpace() (uint32, error)
			}
			rb, ok := dev.(rawber)
			if !ok {
				return fmt.Errorf("missing raw backend")
			}
			// variants of ctrl size
			sizes := []int{92, 44}
			// try runs
			for _, s := range sizes {
				fmt.Printf("== ctrlSize=%d ==\n", s)
				for qi, seq := range seqs {
					fmt.Printf("-- seq[%d]=%v --\n", qi, seq)
					// guards
					_, _ = rb.MSTRawBlock(2, 0xc, 0x58)
					_, _ = rb.MSTRawBlock(2, 6, 0x30)
					_, _ = rb.MSTRawBlock(2, 0xb, 0x200)
					// set sequence
					for _, sp := range seq {
						_ = rb.MSTRawSetAddrSpace(sp)
					}
					cur, _ := rb.MSTRawGetAddrSpace()
					fmt.Printf("space.cur=0x%x\n", cur)
					// ctrl write/read
					buf := make([]byte, s)
					// write opcode only (0x9001<<16)
					val := uint32(0x9001) << 16
					buf[0] = byte(val >> 24)
					buf[1] = byte(val >> 16)
					buf[2] = byte(val >> 8)
					buf[3] = byte(val)
					if err := dev.WriteBlock(0x3, 0x0, buf); err != nil {
						fmt.Printf("ctrl.write err=%v\n", err)
					} else {
						fmt.Printf("ctrl.write ok\n")
					}
					if b, err := dev.ReadBlock(0x3, 0x0, s); err != nil {
						fmt.Printf("ctrl.read err=%v\n", err)
					} else {
						fmt.Printf("ctrl.read ok head=%02x %02x %02x %02x\n", b[0], b[1], b[2], b[3])
					}
				}
			}
			return nil
		},
	}
	sweep.Flags().StringVar(&spaces, "spaces", "", "Semicolon-separated sequences, e.g., '10,2,3;3;2,3'")
	arCmd.AddCommand(sweep)

	// mst-raw: issue a raw MST_BLOCK_ACCESS ioctl (debug)
	var rawNr int
	var rawSize int
	var rawDir string
	mstRaw := &cobra.Command{
		Use:   "mst-raw",
		Short: "Issue a raw MST_BLOCK_ACCESS ioctl (dir=read/write, nr, size)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if mstPath == "" {
				return errors.New("--mst-path /dev/mst/<node> is required for mst-raw")
			}
			if rawNr <= 0 || rawSize < 0 {
				return errors.New("--nr and --size must be set")
			}
			dev, err := pcie.Open(mstPath, logger)
			if err != nil {
				return fmt.Errorf("open device: %w", err)
			}
			defer dev.Close()
			// Access backend method
			type rawber interface {
				MSTRawBlock(dir, nr, size int) ([]byte, error)
			}
			m, ok := dev.(rawber)
			if !ok {
				return fmt.Errorf("MST backend required")
			}
			dir := 2
			if rawDir == "write" {
				dir = 1
			}
			b, err := m.MSTRawBlock(dir, rawNr, rawSize)
			if err != nil {
				return err
			}
			if dir == 2 {
				fmt.Printf("read %d bytes:\n", len(b))
				lim := rawSize
				if lim > len(b) {
					lim = len(b)
				}
				for i := 0; i < lim; i += 16 {
					end := i + 16
					if end > lim {
						end = lim
					}
					fmt.Printf("0x%03x:", i)
					for j := i; j < end; j++ {
						fmt.Printf(" %02x", b[j])
					}
					fmt.Println()
				}
			} else {
				fmt.Println("write ok")
			}
			return nil
		},
	}
	mstRaw.Flags().StringVar(&rawDir, "dir", "read", "Direction: read|write")
	mstRaw.Flags().IntVar(&rawNr, "nr", 12, "MST_BLOCK_ACCESS ioctl number")
	mstRaw.Flags().IntVar(&rawSize, "size", 0x58, "Payload size in bytes")
	arCmd.AddCommand(mstRaw)

	// flash-dump: read flash via MFPA/MFBA access registers
	var outPath string
	var sizeBytes int
	var offsetBytes int
	var chunkSize int
	var useSysfs bool
	flashDump := &cobra.Command{
		Use:   "flash-dump",
		Short: "Dump flash using MFPA/MFBA registers",
		RunE: func(cmd *cobra.Command, args []string) error {
			if deviceBDF == "" && mstPath == "" {
				return errors.New("-d <BDF>|/sys/bus/pci/devices/<BDF> or --mst-path is required")
			}
			// Prefer sysfs when requested or when a sysfs path is provided
			openSpec := deviceBDF
			if useSysfs && deviceBDF != "" && !strings.Contains(deviceBDF, "/sys/bus/pci/devices/") {
				openSpec = "/sys/bus/pci/devices/" + deviceBDF
			}
			if mstPath != "" {
				openSpec = mstPath
			}
			dev, err := pcie.Open(openSpec, logger)
			if err != nil {
				return fmt.Errorf("open device: %w", err)
			}
			defer dev.Close()
			if logger != nil {
				logger.Info("flash.dump.open", zap.String("backend", dev.Type()), zap.String("spec", openSpec))
			}
			// Ensure MST uses LE ctrl as observed for ICMD VCR on this NIC
			if dev.Type() == "mst" && os.Getenv("MLX5FW_CTRL_LE") == "" {
				_ = os.Setenv("MLX5FW_CTRL_LE", "1")
				if logger != nil {
					logger.Info("mst.ctrl_le.enabled")
				}
			}
			ar := pcie.NewARClient(dev, deviceBDF, logger)

			// Read MFPA (optional, best-effort) to log basics
			if payload, err := ar.TransactFull(pcie.RegID_MFPA, 0x20); err == nil && len(payload) == 0x20 {
				if logger != nil {
					logger.Info("flash.mfpa.ok", zap.Int("size", len(payload)))
				}
			} else if err != nil && logger != nil {
				logger.Debug("flash.mfpa.fail", zap.Error(err))
			}

			// Dump MFBA in chunks
			f, err := os.Create(outPath)
			if err != nil {
				return err
			}
			defer f.Close()
			buf := make([]byte, chunkSize)
			const mfbaSize = 0x10c
			start := offsetBytes
			end := offsetBytes + sizeBytes
			for off := start; off < end; off += chunkSize {
				toRead := chunkSize
				if off+toRead > end {
					toRead = end - off
				}
				full, err := ar.TransactFull(pcie.RegID_MFBA, mfbaSize)
				if err != nil || len(full) != 16+4+mfbaSize {
					return fmt.Errorf("mfba transact failed at 0x%x: %v (size=%d)", off, err, len(full))
				}
				// Initialize request payload in-place: set size (9 bits) and address (32 bits)
				pl := full[16+4:]
				for i := range pl {
					pl[i] = 0
				}
				// size at bit offset 32+23, length 9; address at bit offset 64, length 32
				setBitsBE(pl, 32+23, 9, uint32(toRead))
				setBitsBE(pl, 64, 32, uint32(off))
				// Resend with initialized payload
				if err := ar.SendICMDBuffer(full); err != nil {
					return fmt.Errorf("mfba send failed at 0x%x: %v", off, err)
				}
				payload := full[16+4:]
				if len(payload) != mfbaSize {
					return fmt.Errorf("mfba read failed at 0x%x: %v (size=%d)", off, err, len(payload))
				}
				// Data is the last chunkSize bytes of the register payload
				data := payload[mfbaSize-toRead : mfbaSize]
				copy(buf[:toRead], data)
				if _, err := f.Write(buf[:toRead]); err != nil {
					return err
				}
				if logger != nil && ((off-start)&0xFFFF) == 0 {
					logger.Info("flash.dump.progress", zap.Int("offset", off))
				}
			}
			if logger != nil {
				logger.Info("flash.dump.done", zap.String("out", outPath), zap.Int("size", sizeBytes))
			}
			return nil
		},
	}
	flashDump.Flags().StringVar(&outPath, "out", "flash_data.bin", "Output file path")
	flashDump.Flags().IntVar(&sizeBytes, "size", 16*1024*1024, "Total bytes to read")
	flashDump.Flags().IntVar(&offsetBytes, "offset", 0, "Starting offset in flash")
	flashDump.Flags().IntVar(&chunkSize, "chunk", 0x40, "Chunk size in bytes (MFBA data)")
	flashDump.Flags().BoolVar(&useSysfs, "sysfs", true, "Prefer sysfs VSEC backend when -d is a BDF")
	arCmd.AddCommand(flashDump)

	// vsec-set: dump and optionally set VSEC address_space (guarded)
	var vsecSpace int
	var vsecOffset uint32
	vsecSet := &cobra.Command{
		Use:   "vsec-set",
		Short: "Dump VSEC ctrl dword and optionally set space (guarded)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if deviceBDF == "" {
				return errors.New("-d <BDF> is required for vsec-set")
			}
			// Read via sysfs config file for robustness
			cfg := fmt.Sprintf("/sys/bus/pci/devices/%s/config", deviceBDF)
			f, err := os.OpenFile(cfg, os.O_RDWR, 0)
			if err != nil {
				return fmt.Errorf("open pci config: %w", err)
			}
			defer f.Close()
			// Find VSEC ctrl offset by reading MST node if provided, else require offset flag in future (TBD).
			// Use a minimal scan: try reading dword at common VSEC ctrl offset from mst status (functional_vsc_offset)
			// Prefer using MST backend when mstPath provided.
			off := uint32(0)
			if mstPath != "" {
				if dev, e := pcie.Open(mstPath, logger); e == nil {
					defer dev.Close()
					if p, e2 := pcie.MSTParams(dev); e2 == nil {
						off = p.FunctionalVsecOffset + 0x4
					}
				}
			}
			if off == 0 && vsecOffset != 0 {
				off = vsecOffset + 0x4
			}
			if off == 0 {
				// Try to locate first PCIe extended VSEC capability
				buf := make([]byte, 4096)
				if _, err := f.ReadAt(buf, 0); err == nil {
					// Extended caps start at 0x100
					next := uint32(0x100)
					for iter := 0; iter < 64 && next >= 0x100 && int(next+4) <= len(buf); iter++ {
						hdr := uint32(buf[next]) | uint32(buf[next+1])<<8 | uint32(buf[next+2])<<16 | uint32(buf[next+3])<<24
						capID := hdr & 0xffff
						if capID == 0x000b { // PCI_EXT_CAP_ID_VNDR
							// Treat this VSEC as functional VSEC
							off = next + 0x4
							break
						}
						next = (hdr >> 20) & 0xfff
						if next == 0 {
							break
						}
					}
				}
			}
			if off == 0 {
				return fmt.Errorf("unable to resolve functional_vsc_offset; provide --mst-path or --offset <hex>")
			}
			buf := make([]byte, 4)
			if _, err := f.ReadAt(buf, int64(off)); err != nil {
				return fmt.Errorf("read VSEC ctrl@0x%x: %w", off, err)
			}
			val := uint32(buf[0]) | uint32(buf[1])<<8 | uint32(buf[2])<<16 | uint32(buf[3])<<24
			space := val & 0xffff
			status := (val >> 29) & 0x7
			fmt.Printf("Before: ctrl=0x%08x space=0x%04x status=0x%x\n", val, space, status)
			if vsecSpace >= 0 {
				// Guarded write path via sysfs config (only if MLX5FW_MST_VSEC_SET=1)
				type vsecSetter interface{ VSECSetAddrSpace(space uint16) error }
				if mstPath != "" {
					if dev, e := pcie.Open(mstPath, logger); e == nil {
						defer dev.Close()
						if vs, ok := dev.(vsecSetter); ok {
							if err := vs.VSECSetAddrSpace(uint16(vsecSpace)); err != nil {
								return fmt.Errorf("VSEC set: %w", err)
							}
						} else {
							return fmt.Errorf("backend does not expose VSECSetAddrSpace")
						}
					} else {
						return fmt.Errorf("open mst: %w", e)
					}
				} else {
					return fmt.Errorf("--mst-path is required to resolve vsec offset")
				}
				// Re-read
				if _, err := f.ReadAt(buf, int64(off)); err != nil {
					return fmt.Errorf("readback VSEC ctrl: %w", err)
				}
				val2 := uint32(buf[0]) | uint32(buf[1])<<8 | uint32(buf[2])<<16 | uint32(buf[3])<<24
				space2 := val2 & 0xffff
				status2 := (val2 >> 29) & 0x7
				fmt.Printf("After:  ctrl=0x%08x space=0x%04x status=0x%x\n", val2, space2, status2)
			}
			return nil
		},
	}
	vsecSpace = -1
	vsecSet.Flags().IntVar(&vsecSpace, "space", -1, "Space to set (e.g., 3). Negative=read-only dump")
	vsecSet.Flags().Uint32Var(&vsecOffset, "offset", 0, "Functional VSEC offset (hex, e.g., 0xABC) when MST_PARAMS is unavailable")
	arCmd.AddCommand(vsecSet)

	// pci-scan-vsec: scan PCI ext caps for VNDR caps to help locating VSEC offsets
	pciScan := &cobra.Command{
		Use:   "pci-scan-vsec",
		Short: "Scan PCI extended capabilities for VNDR caps and print offsets",
		RunE: func(cmd *cobra.Command, args []string) error {
			if deviceBDF == "" {
				return errors.New("-d <BDF> is required")
			}
			cfg := fmt.Sprintf("/sys/bus/pci/devices/%s/config", deviceBDF)
			f, err := os.Open(cfg)
			if err != nil {
				return fmt.Errorf("open config: %w", err)
			}
			defer f.Close()
			buf := make([]byte, 4096)
			if _, err := f.ReadAt(buf, 0); err != nil {
				return fmt.Errorf("read config: %w", err)
			}
			next := uint32(0x100)
			for i := 0; i < 64 && next >= 0x100 && int(next+4) <= len(buf); i++ {
				hdr := uint32(buf[next]) | uint32(buf[next+1])<<8 | uint32(buf[next+2])<<16 | uint32(buf[next+3])<<24
				capID := hdr & 0xffff
				nxt := (hdr >> 20) & 0xfff
				fmt.Printf("extcap id=0x%04x off=0x%03x next=0x%03x\n", capID, next, nxt)
				next = nxt
				if next == 0 {
					break
				}
			}
			return nil
		},
	}
	arCmd.AddCommand(pciScan)

	// mst-params: print MST_PARAMS for a specific /dev/mst node
	mstParamsCmd := &cobra.Command{
		Use:   "mst-params",
		Short: "Show MST_PARAMS for --mst-path",
		RunE: func(cmd *cobra.Command, args []string) error {
			if mstPath == "" {
				return errors.New("--mst-path /dev/mst/<node> is required")
			}
			dev, err := pcie.Open(mstPath, logger)
			if err != nil {
				return fmt.Errorf("open: %w", err)
			}
			defer dev.Close()
			if dev.Type() != "mst" {
				return fmt.Errorf("mst backend required")
			}
			p, err := pcie.MSTParams(dev)
			if err != nil {
				return err
			}
			fmt.Printf("BDF: %04x:%02x:%02x.%d\n", p.Domain, p.Bus, p.Slot, p.Func)
			fmt.Printf("Vendor: 0x%04x Device: 0x%04x SubVendor: 0x%04x SubDevice: 0x%04x\n", p.Vendor, p.Device, p.SubsystemVendor, p.SubsystemDevice)
			fmt.Printf("FunctionalVsecOffset: 0x%08x\n", p.FunctionalVsecOffset)
			return nil
		},
	}
	arCmd.AddCommand(mstParamsCmd)

	return arCmd
}

// scanVsecExtOffset reads the PCI config space and walks the extended capability list
// to find the first VNDR (0x000b) capability base offset.
// (helpers moved into existing commands above)

// buildTLVForReg constructs a minimal GET TLV buffer (Operation TLV + Reg TLV + zero payload)
// with BE headers suitable for writing to the device mailbox directly.
func buildTLVForReg(regID uint16, regSize int) []byte {
	total := 16 + 4 + regSize
	buf := make([]byte, total)
	// Operation TLV exact (TLV len=4 dwords, Type=1, class=1, method=1, regID)
	// Clear first 16 bytes
	for i := 0; i < 16; i++ {
		buf[i] = 0
	}
	// len: 4 dwords at bit 5..15
	setBitsBE(buf[0:16], 5, 11, 4)
	// Type=1 at bits 0..4
	setBitsBE(buf[0:16], 0, 5, 1)
	// class=1 at bits 32..39
	setBitsBE(buf[0:16], 32, 8, 1)
	// method=1 at bits 49..55
	setBitsBE(buf[0:16], 49, 7, 1)
	// register_id at bits 48..63
	setBitsBE(buf[0:16], 48, 16, uint32(regID))
	// Reg TLV header
	dwords := (4 + regSize) / 4
	hdr := buf[16:20]
	for i := 0; i < 4; i++ {
		hdr[i] = 0
	}
	setBitsBE(hdr, 5, 11, uint32(dwords))
	setBitsBE(hdr, 0, 5, 3) // TLV_REG=3
	// Payload already zeroed by default
	return buf
}

// rawReplayMST performs a strictly raw MST ctrl/mailbox sequence using exact sizes.
// It programs opcode, writes mailbox, pulses GO, polls busy clear, then reads back the mailbox.
func rawReplayMST(dev pcie.Device, reg uint16, regSize int, be []byte) error {
	fmt.Println("[raw] begin raw MST replay")
	// Constants mirrored from transport
	const (
		ctrlSize   = 92
		vcrSpace   = uint16(0x3)
		vcrCtrlOff = uint32(0x0)
		vcrCmdOff  = uint32(0x100000)
		icmdOpcode = uint32(0x9001)
		busyBit    = uint32(1 << 0)
		exmbBit    = uint32(1 << 1)
	)
	// Try to align address space via raw 12B if supported
	type rawber interface {
		MSTRawSetAddrSpace(space uint16) error
		MSTRawGetAddrSpace() (uint32, error)
	}
	if rb, ok := dev.(rawber); ok {
		_ = rb.MSTRawSetAddrSpace(10)
		_ = rb.MSTRawSetAddrSpace(2)
		_ = rb.MSTRawSetAddrSpace(3)
		_, _ = rb.MSTRawGetAddrSpace()
	}
	// Read ctrl head (buffer path) to get current value (best-effort)
	ctrlBefore := uint32(0)
	if b, err := dev.ReadBlock(vcrSpace, vcrCtrlOff, ctrlSize); err == nil && len(b) >= 4 {
		ctrlBefore = (uint32(b[0]) << 24) | (uint32(b[1]) << 16) | (uint32(b[2]) << 8) | uint32(b[3])
	}
	// Program opcode + EXMB
	opcodeMask := uint32(0xffff) << 16
	setEXMB := os.Getenv("MLX5FW_NO_SET_EXMB") == ""
	ctrlProg := (ctrlBefore &^ opcodeMask) | (icmdOpcode << 16)
	if setEXMB {
		ctrlProg |= exmbBit
	}
	wb := make([]byte, ctrlSize)
	if os.Getenv("MLX5FW_CTRL_LE") != "" { // experimental: write ctrl dword in LE byte order
		wb[0], wb[1], wb[2], wb[3] = byte(ctrlProg), byte(ctrlProg>>8), byte(ctrlProg>>16), byte(ctrlProg>>24)
	} else {
		wb[0], wb[1], wb[2], wb[3] = byte(ctrlProg>>24), byte(ctrlProg>>16), byte(ctrlProg>>8), byte(ctrlProg)
	}
	fmt.Printf("[raw] ctrl_before=0x%08x ctrl_prog=0x%08x\n", ctrlBefore, ctrlProg)
	if err := dev.WriteBlock(vcrSpace, vcrCtrlOff, wb); err != nil {
		return fmt.Errorf("raw ctrl write (opcode): %w", err)
	}
	// Write mailbox (exact size = 16+4+regSize)
	mbLen := 16 + 4 + regSize
	if err := dev.WriteBlock(vcrSpace, vcrCmdOff, be[:mbLen]); err != nil {
		return fmt.Errorf("raw mailbox write: %w", err)
	}
	// GO pulse (set busy)
	ctrlGo := ctrlProg | busyBit
	if os.Getenv("MLX5FW_CTRL_LE") != "" {
		wb[0], wb[1], wb[2], wb[3] = byte(ctrlGo), byte(ctrlGo>>8), byte(ctrlGo>>16), byte(ctrlGo>>24)
	} else {
		wb[0], wb[1], wb[2], wb[3] = byte(ctrlGo>>24), byte(ctrlGo>>16), byte(ctrlGo>>8), byte(ctrlGo)
	}
	fmt.Printf("[raw] ctrl_go=0x%08x\n", ctrlGo)
	if err := dev.WriteBlock(vcrSpace, vcrCtrlOff, wb); err != nil {
		return fmt.Errorf("raw ctrl write (go): %w", err)
	}
	// Poll busy clear (bounded)
	deadline := time.Now().Add(3 * time.Second)
	for {
		b, err := dev.ReadBlock(vcrSpace, vcrCtrlOff, ctrlSize)
		if err == nil && len(b) >= 4 {
			cur := (uint32(b[0]) << 24) | (uint32(b[1]) << 16) | (uint32(b[2]) << 8) | uint32(b[3])
			if (cur & busyBit) == 0 {
				break
			}
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("icmd timeout waiting busy clear")
		}
		time.Sleep(100 * time.Microsecond)
	}
	// Read mailbox back (exact)
	data, err := dev.ReadBlock(vcrSpace, vcrCmdOff, mbLen)
	if err != nil {
		return fmt.Errorf("raw mailbox read: %w", err)
	}
	status := uint32(0)
	if len(data) >= 12 {
		status = uint32(data[8]) | uint32(data[9])<<8 | uint32(data[10])<<16 | uint32(data[11])<<24
	}
	fmt.Printf("Result: status=0x%04x len=%d\n", status&0xffff, len(data))
	for i := 0; i < 64 && i < len(data); i += 16 {
		end := i + 16
		if end > len(data) {
			end = len(data)
		}
		fmt.Printf("0x%03x:", i)
		for j := i; j < end; j++ {
			fmt.Printf(" %02x", data[j])
		}
		fmt.Println()
	}
	return nil
}
