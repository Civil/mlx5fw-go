//go:build linux
// +build linux

package pcie

import (
    "encoding/binary"
    "encoding/hex"
    "fmt"
    "os"
    "time"

	"go.uber.org/zap"
)

// Internal HW IDs from mstflint (subset)
const (
	hwID_CX5 = 525
)

// ICMD constants (from mstflint)
const (
	icmdCtrlOffset   = 0x3fc
	icmdBusyBitOff   = 0
	icmdBusyBitMask  = 1 << icmdBusyBitOff
	icmdExmbBitMask  = 1 << 1
	icmdOpcodeBitOff = 16
	icmdOpcodeMask   = 0xffff << icmdOpcodeBitOff

	icmdFlashRegAccess = 0x9001
)

type icmdInfo struct {
	cmdAddr  uint32
	ctrlAddr uint32
	semAddr  uint32
}

// icmdStaticCfgReady checks the device readiness bit exposed by firmware.
// CX5-class devices expose a static_cfg_not_done bit at ICMD offset 0xb5e04, bit 31.
// When set, firmware is not ready to accept ICMD transactions.
func (c *ARClient) icmdStaticCfgReady() error {
	// Best-effort: read CR space at 0xb5e04 and check MSB (static_cfg_not_done)
	const off uint32 = 0x00b5e04
	v, err := c.dev.Read32(0x2, off)
	if err != nil {
		// Don't hard-fail if the register is not accessible on this family
		return nil
	}
	if (v & (1 << 31)) != 0 {
		return fmt.Errorf("icmd not ready: static_cfg_not_done=1 (0x%08x)", v)
	}
	return nil
}

// tryInitICMDCR initializes ICMD addresses for CX5-class devices using CR-space.
func (c *ARClient) tryInitICMDCR() (*icmdInfo, error) {
	// Read internal HW ID from CR space
	id, err := c.dev.Read32(0x2, 0x00f0014)
	if err != nil {
		return nil, fmt.Errorf("read HW_ID: %w", err)
	}
	hwid := int(id & 0xffff)
	var cmdPtrAddr uint32
	var semAddr uint32
	switch hwid {
	case hwID_CX5:
		// CX5 path uses CIB addrs
		cmdPtrAddr = 0x0
		semAddr = 0x0e74e0
	default:
		// Fallback: attempt CX5 defaults (many NICs alias CIB here)
		cmdPtrAddr = 0x0
		semAddr = 0x0e74e0
	}
	// Read CR at cmdPtrAddr to fetch pointer and compute ctrl
	reg, err := c.dev.Read32(0x2, cmdPtrAddr)
	if err != nil {
		return nil, fmt.Errorf("read cmd_ptr: %w", err)
	}
	cmdAddr := reg & ((1 << 24) - 1) // 24-bit pointer
	ctrlAddr := cmdAddr + icmdCtrlOffset
	return &icmdInfo{cmdAddr: cmdAddr, ctrlAddr: ctrlAddr, semAddr: semAddr}, nil
}

// sendICMD sends buffer via ICMD mailbox and reads back into the same buffer.
func (c *ARClient) sendICMD(buf []byte) error {
    // Optional path control for debugging/stability: MLX5FW_AR_PATH={vcr,cr}
    // Default remains: try VCR, then CR fallback.
    arPath := os.Getenv("MLX5FW_AR_PATH")

    // Readiness check (best-effort)
    if err := c.icmdStaticCfgReady(); err != nil {
        return err
    }

    // Build on-wire buffer: headers BE, payload raw
    be := make([]byte, len(buf))
    headerBytes := 16 + 4
    fromBEHeader := func(dst, src []byte, hdr int) {
        n := hdr / 4
        for i := 0; i < n; i++ {
            off := i * 4
            v := uint32(src[off])<<24 | uint32(src[off+1])<<16 | uint32(src[off+2])<<8 | uint32(src[off+3])
            binary.LittleEndian.PutUint32(dst[off:off+4], v)
        }
    }
    copy(be, buf)
    toBE(be[:headerBytes], buf[:headerBytes])

    if c.logger != nil {
        op0le := binary.LittleEndian.Uint32(buf[0:4])
        op1le := binary.LittleEndian.Uint32(buf[4:8])
        regle := binary.LittleEndian.Uint32(buf[16:20])
        dump := 64
        if dump > len(be) {
            dump = len(be)
        }
        c.logger.Debug("ar.tlv.header",
            zap.Uint32("op_word0_le", op0le),
            zap.Uint32("op_word1_le", op1le),
            zap.Uint32("reg_hdr_le", regle),
            zap.String("op_hdr_be_bytes", hex.EncodeToString(be[:16])),
            zap.String("reg_hdr_be_bytes", hex.EncodeToString(be[16:20])),
        )
        c.logger.Debug("icmd.mailbox.pre_go", zap.String("bytes", hex.EncodeToString(be[:dump])))
    }

    // VCR attempt (unless forced CR)
    if arPath != "cr" {
        vcrCmdAddr := uint32(0x100000)
        vcrCtrlAddr := uint32(0x0)
        if c.logger != nil {
            c.logger.Info("icmd.begin", zap.Int("size", len(be)), zap.String("ar_path", arPath), zap.Uint32("vcr_cmd_addr", vcrCmdAddr), zap.Uint32("vcr_ctrl_addr", vcrCtrlAddr))
        }
        if _, err := c.dev.Read32(0x3, vcrCtrlAddr); err == nil {
            _ = icmdLockSemaphore(c.dev, 0x0)
            defer icmdUnlockSemaphore(c.dev, 0x0)
            if err := icmdWaitBusyClear(c.dev, vcrCtrlAddr, 3, 5*time.Second); err != nil {
                return err
            }
            ctrlBefore, err := c.dev.Read32(0x3, vcrCtrlAddr)
            if err != nil { return fmt.Errorf("icmd vcr read ctrl: %w", err) }
            ctrlProg := (ctrlBefore &^ icmdOpcodeMask) | (uint32(icmdFlashRegAccess) << icmdOpcodeBitOff)
            if err := c.dev.Write32(0x3, vcrCtrlAddr, ctrlProg); err != nil {
                return fmt.Errorf("icmd vcr write ctrl (opcode): %w (before=0x%08x prog=0x%08x)", err, ctrlBefore, ctrlProg)
            }
            if c.logger != nil {
                if chk, err := c.dev.Read32(0x3, vcrCtrlAddr); err == nil {
            c.logger.Debug("icmd.vcr.ctrl.check_go", zap.Uint32("ctrl", chk))
                }
            }
            if err := c.dev.WriteBlock(0x2, vcrCmdAddr, be); err != nil {
                if err2 := c.dev.WriteBlock(0x3, vcrCmdAddr, be); err2 != nil {
                    return fmt.Errorf("icmd write vcr mailbox: %w", err)
                }
            }
            ctrl := ctrlProg | icmdBusyBitMask
            if err := c.dev.Write32(0x3, vcrCtrlAddr, ctrl); err != nil {
                return fmt.Errorf("icmd vcr write ctrl (go): %w (go=0x%08x)", err, ctrl)
            }
            if c.logger != nil {
            c.logger.Debug("icmd.vcr.ctrl",
                zap.Uint32("before", ctrlBefore),
                zap.Uint32("prog", ctrlProg),
                zap.Uint32("go", ctrl),
                zap.Uint32("opcode", (ctrlProg>>icmdOpcodeBitOff)&0xffff),
                zap.Uint32("exmb", (ctrlProg>>1)&1),
                zap.Uint32("busy", ctrlProg&1),
            )
            }
            if err := c.waitBusyClear(vcrCtrlAddr, 3, 5*time.Second); err == nil {
                ctrlAfter, _ := c.dev.Read32(0x3, vcrCtrlAddr)
                if c.logger != nil {
                c.logger.Debug("icmd.vcr.ctrl.after",
                    zap.Uint32("after", ctrlAfter),
                    zap.Uint32("busy", ctrlAfter&1),
                    zap.Uint32("exmb", (ctrlAfter>>1)&1),
                    zap.Uint32("opcode", (ctrlAfter>>icmdOpcodeBitOff)&0xffff),
                )
            }
                if ctrlAfter != ctrlBefore {
                    // Read mailbox
                    data, err := c.dev.ReadBlock(0x2, vcrCmdAddr, len(buf))
                    if err != nil {
                        if data2, err2 := c.dev.ReadBlock(0x3, vcrCmdAddr, len(buf)); err2 == nil { data = data2 } else { return fmt.Errorf("icmd vcr read mailbox: %w", err) }
                    }
                    fromBEHeader(buf, data, headerBytes)
                    copy(buf[headerBytes:], data[headerBytes:])
                    if len(buf) >= 12 && c.logger != nil {
                        opWord2 := uint32(buf[8]) | uint32(buf[9])<<8 | uint32(buf[10])<<16 | uint32(buf[11])<<24
                        c.logger.Debug("icmd.vcr.op.status", zap.Uint32("op_word2", opWord2), zap.Uint32("status", opWord2&0xffff))
                        // Dump mailbox head after completion for parity
                        dump := 64; if dump > len(buf) { dump = len(buf) }
                        c.logger.Debug("icmd.vcr.mailbox.head", zap.String("bytes", hex.EncodeToString(buf[:dump])))
                    }
                    return nil
                } else {
                    if data, err := c.dev.ReadBlock(0x2, vcrCmdAddr, len(buf)); err == nil {
                        fromBEHeader(buf, data, headerBytes)
                        copy(buf[headerBytes:], data[headerBytes:])
                        if len(buf) >= 12 {
                            opWord2 := uint32(buf[8]) | uint32(buf[9])<<8 | uint32(buf[10])<<16 | uint32(buf[11])<<24
                            if c.logger != nil { c.logger.Debug("icmd.vcr.op.status.fallback", zap.Uint32("op_word2", opWord2), zap.Uint32("status", opWord2&0xffff)) }
                            if (opWord2 & 0xffff) == 0 { return nil }
                        }
                    }
                }
            }
        }
    }

    if arPath == "vcr" {
        return fmt.Errorf("icmd: no successful path (ar_path=%s)", arPath)
    }

    // CR path
    icmd, err := c.tryInitICMDCR()
    if err != nil { return err }
    if c.logger != nil { c.logger.Debug("icmd.cr.addrs", zap.Uint32("cmd_addr", icmd.cmdAddr), zap.Uint32("ctrl_addr", icmd.ctrlAddr), zap.Uint32("sem_addr", icmd.semAddr)) }
    _ = icmdLockSemaphore(c.dev, icmd.semAddr)
    defer icmdUnlockSemaphore(c.dev, icmd.semAddr)
    if err := c.waitBusyClear(icmd.ctrlAddr, 3, 5*time.Second); err != nil { return err }
    ctrlBefore, err := c.dev.Read32(0x3, icmd.ctrlAddr)
    if err != nil { return fmt.Errorf("icmd read ctrl: %w", err) }
    ctrlProg := (ctrlBefore &^ icmdOpcodeMask) | (uint32(icmdFlashRegAccess) << icmdOpcodeBitOff)
    if err := c.dev.Write32(0x3, icmd.ctrlAddr, ctrlProg); err != nil {
        return fmt.Errorf("icmd write ctrl (opcode): %w (before=0x%08x prog=0x%08x)", err, ctrlBefore, ctrlProg)
    }
    if err := c.dev.WriteBlock(0x2, icmd.cmdAddr, be); err != nil {
        if err2 := c.dev.WriteBlock(0x3, icmd.cmdAddr, be); err2 != nil { return fmt.Errorf("icmd write mailbox: %w", err) }
    }
    ctrl := ctrlProg | icmdBusyBitMask
    if err := c.dev.Write32(0x3, icmd.ctrlAddr, ctrl); err != nil { return fmt.Errorf("icmd write ctrl (go): %w (go=0x%08x)", err, ctrl) }
    if c.logger != nil {
        c.logger.Debug("icmd.cr.ctrl",
            zap.Uint32("before", ctrlBefore),
            zap.Uint32("prog", ctrlProg),
            zap.Uint32("go", ctrl),
            zap.Uint32("opcode", (ctrlProg>>icmdOpcodeBitOff)&0xffff),
            zap.Uint32("exmb", (ctrlProg>>1)&1),
            zap.Uint32("busy", ctrlProg&1),
        )
    }
    if err := icmdWaitBusyClear(c.dev, icmd.ctrlAddr, 3, 5*time.Second); err != nil { return err }
    if c.logger != nil {
        if after, err := c.dev.Read32(0x3, icmd.ctrlAddr); err == nil {
            c.logger.Debug("icmd.cr.ctrl.after",
                zap.Uint32("after", after),
                zap.Uint32("busy", after&1),
                zap.Uint32("exmb", (after>>1)&1),
                zap.Uint32("opcode", (after>>icmdOpcodeBitOff)&0xffff),
            )
        }
    }
    data, err := c.dev.ReadBlock(0x2, icmd.cmdAddr, len(buf))
    if err != nil {
        if data2, err2 := c.dev.ReadBlock(0x3, icmd.cmdAddr, len(buf)); err2 == nil { data = data2 } else { return fmt.Errorf("icmd read mailbox: %w", err) }
    }
    fromBEHeader(buf, data, headerBytes)
    copy(buf[headerBytes:], data[headerBytes:])
    if len(buf) >= 12 && c.logger != nil {
        opWord2 := uint32(buf[8]) | uint32(buf[9])<<8 | uint32(buf[10])<<16 | uint32(buf[11])<<24
        c.logger.Debug("icmd.cr.op.status", zap.Uint32("op_word2", opWord2), zap.Uint32("status", opWord2&0xffff))
        dump := 64; if dump > len(buf) { dump = len(buf) }
        c.logger.Debug("icmd.cr.mailbox.head", zap.String("bytes", hex.EncodeToString(buf[:dump])))
    }
    return nil
}

func icmdWaitBusyClear(dev Device, ctrlAddr uint32, space uint16, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		cur, err := dev.Read32(space, ctrlAddr)
		if err != nil {
			return fmt.Errorf("icmd poll ctrl: %w", err)
		}
		if (cur & icmdBusyBitMask) == 0 {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("icmd timeout waiting for busy clear")
		}
		time.Sleep(2 * time.Millisecond)
	}
}

// waitBusyClear polls ctrl busy bit with instrumentation when logger is present.
func (c *ARClient) waitBusyClear(ctrlAddr uint32, space uint16, timeout time.Duration) error {
	if c.logger == nil {
		return icmdWaitBusyClear(c.dev, ctrlAddr, space, timeout)
	}
	deadline := time.Now().Add(timeout)
	var first, last uint32
	const logEvery = 64
	iter := 0
	for {
		cur, err := c.dev.Read32(space, ctrlAddr)
		if err != nil {
			return fmt.Errorf("icmd poll ctrl: %w", err)
		}
		if iter == 0 {
			first = cur
		}
		last = cur
		if (cur & icmdBusyBitMask) == 0 {
			c.logger.Info("icmd.poll.clear",
				zap.Uint32("first", first),
				zap.Uint32("last", last),
				zap.Uint32("busy", last&1),
				zap.Uint32("status", (last>>8)&0xff),
			)
			return nil
		}
		if time.Now().After(deadline) {
			c.logger.Info("icmd.poll.timeout",
				zap.Uint32("first", first),
				zap.Uint32("last", last),
				zap.Uint32("busy", last&1),
				zap.Uint32("status", (last>>8)&0xff),
			)
			return fmt.Errorf("icmd timeout waiting for busy clear")
		}
		if (iter % logEvery) == 0 {
			c.logger.Info("icmd.poll.sample",
				zap.Int("iter", iter),
				zap.Uint32("ctrl", cur),
				zap.Uint32("busy", cur&1),
				zap.Uint32("status", (cur>>8)&0xff),
			)
		}
		iter++
		time.Sleep(2 * time.Millisecond)
	}
}

// toBE converts slice of bytes interpreted as consecutive u32 (host endian)
// into big-endian bytes in dst.
func toBE(dst, src []byte) {
	n := len(src) / 4
	for i := 0; i < n; i++ {
		off := i * 4
		v := uint32(src[off]) | uint32(src[off+1])<<8 | uint32(src[off+2])<<16 | uint32(src[off+3])<<24
		dst[off] = byte(v >> 24)
		dst[off+1] = byte(v >> 16)
		dst[off+2] = byte(v >> 8)
		dst[off+3] = byte(v)
	}
}

// fromBE converts big-endian bytes src into host-endian layout in dst (u32-wise).
func fromBE(dst, src []byte) {
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

// Attempt to lock/unlock ICMD semaphore via semaphore space (best-effort)
func icmdLockSemaphore(dev Device, semOff uint32) error {
	const semSpace = 0x0a
	ticket := uint32(os.Getpid())
	for retries := 0; retries < 256; retries++ {
		_ = dev.Write32(semSpace, semOff, ticket)
		v, err := dev.Read32(semSpace, semOff)
		if err == nil && v == ticket {
			return nil
		}
		time.Sleep(1 * time.Millisecond)
	}
	return fmt.Errorf("icmd semaphore acquire timeout")
}

func icmdUnlockSemaphore(dev Device, semOff uint32) {
	const semSpace = 0x0a
	_ = dev.Write32(semSpace, semOff, 0)
}

// transactAR builds a TLV for a GET access register and returns the raw register payload.
func (c *ARClient) transactAR(regID uint16, regSize int) ([]byte, error) {
	const opTLVSize = 16
	const regHdr = 4
	total := opTLVSize + regHdr + regSize
	// Build TLV per mstflint (packets_layout.c):
	// Operation TLV (16 bytes):
	// word0: [Type:5][len:11][dr:1=0][status:7=0][reserved:8=0]
	// word1: [register_id:16][r:1=0][method:7][class:8=1]
	// word2+3: tid (64b) = 0
	word0 := (uint32(1) & 0x1f) | (uint32(4) << 5)
	word1 := (uint32(regID) & 0xffff) | (uint32(RegMethodGet) << 17) | (uint32(1) << 24)
	// Reg TLV header (4 bytes): [Type:5=3][len:11=dwords][reserved:16=0]
	dwords := uint16((regHdr + regSize) / 4)
	regHdrWord := (uint32(3) & 0x1f) | (uint32(dwords) << 5)
	buf := make([]byte, total)
	// Write opTLV words in little-endian; transport flips header to BE on wire
	binary.LittleEndian.PutUint32(buf[0:4], word0)
	binary.LittleEndian.PutUint32(buf[4:8], word1)
	// tid = 0 for bytes 8..15
	for i := 8; i < 16; i++ {
		buf[i] = 0
	}
	// Reg TLV header at bytes 16..19 (little-endian, flipped by transport)
	binary.LittleEndian.PutUint32(buf[16:20], regHdrWord)
	// Zero reg payload for GET is fine; device fills it
	// Send via ICMD
	if err := c.sendICMD(buf); err != nil {
		return nil, err
	}
	// Extract reg payload
	return buf[opTLVSize+regHdr:], nil
}

// transactARWithInit builds a GET TLV for regID and allows caller to initialize
// the register payload bytes before sending (used for parameterized regs like MCQI).
func (c *ARClient) transactARWithInit(regID uint16, regSize int, init func(payload []byte)) ([]byte, error) {
	const opTLVSize = 16
	const regHdr = 4
	total := opTLVSize + regHdr + regSize
	// Operation TLV header and Reg TLV header (same packing as transactAR)
	word0 := (uint32(1) & 0x1f) | (uint32(4) << 5)
	word1 := (uint32(regID) & 0xffff) | (uint32(RegMethodGet) << 17) | (uint32(1) << 24)
	dwords := uint16((regHdr + regSize) / 4)
	regHdrWord := (uint32(3) & 0x1f) | (uint32(dwords) << 5)
	buf := make([]byte, total)
	// op tlv (written LE, flipped by transport)
	binary.LittleEndian.PutUint32(buf[0:4], word0)
	binary.LittleEndian.PutUint32(buf[4:8], word1)
	// tid = 0
	for i := 8; i < 16; i++ {
		buf[i] = 0
	}
	// reg tlv header (write LE, flipped in transport)
	binary.LittleEndian.PutUint32(buf[16:20], regHdrWord)
	// Initialize payload if requested
	if init != nil {
		init(buf[opTLVSize+regHdr:])
	}
	if err := c.sendICMD(buf); err != nil {
		return nil, err
	}
	return buf[opTLVSize+regHdr:], nil
}
