//go:build linux
// +build linux

package pcie

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"go.uber.org/zap"
)

// Access register client (Phase 2 scaffold)
// Encodes a minimal interface to support MGIR/MCQI sequences for device query.

// The Access Register mechanism is device-family specific. Here we lay down
// the basic helpers and placeholders for the concrete register IDs and payloads.

// Address space identifiers (placeholder values, to be finalized)
const (
	// Access register window address space (CAP9 VSEC) — placeholder
	SpaceAccessReg uint16 = 0x0009
)

// ARClient wraps a Device and provides helpers for Access Register operations.
type ARClient struct {
	dev    Device
	bdf    string      // optional: BDF for sysfs fallback even when backend is MST
	logger *zap.Logger // optional: structured debug logs
}

func NewARClient(dev Device, bdf string, logger *zap.Logger) *ARClient {
	return &ARClient{dev: dev, bdf: bdf, logger: logger}
}

// writeARAddr selects the access register address space and writes the target offset.
func (c *ARClient) writeARAddr(offset uint32) error {
	// In a real implementation we may need to write both space and offset, or pack space into the address field.
	// Here we assume Device.Write32 handles space selection via the first parameter.
	return c.dev.Write32(SpaceAccessReg, 0x00000000, offset)
}

// read32 reads a 32-bit word from the AR data port at the given offset.
func (c *ARClient) read32(offset uint32) (uint32, error) {
	return c.dev.Read32(SpaceAccessReg, offset)
}

// write32 writes a 32-bit word to the AR data port at the given offset.
func (c *ARClient) write32(offset uint32, v uint32) error {
	return c.dev.Write32(SpaceAccessReg, offset, v)
}

// findBootComponent iterates MCQS to discover the BOOT_IMG component index
func (c *ARClient) findBootComponent() (idx uint16, devType uint8, ok bool) {
	const mcqsSize = 0x10
	last := false
	for comp := uint16(0); !last && comp < 512; comp++ {
		payload, err := c.transactARWithInit(RegID_MCQS, mcqsSize, func(pl []byte) {
			if len(pl) < mcqsSize {
				return
			}
			packMCQSHeader(pl[:mcqsSize], comp, 0, 0)
		})
		if err != nil || len(payload) != mcqsSize {
			if c.logger != nil {
				c.logger.Info("ar.mcqs.fail", zap.Uint16("comp", comp), zap.Error(err))
			}
			break
		}
		// Decode minimal fields: identifier (bits at offset 48, 16), last_index_flag (bit 0), device_type (offset 120,8)
		ident := uint16(getBitsBE(payload, 48, 16))
		lastFlag := getBitsBE(payload, 0, 1) != 0
		status := uint8(getBitsBE(payload, 87, 5)) // component_status: 0 not present, 1 present, 2 in use
		dtype := uint8(getBitsBE(payload, 120, 8))
		if c.logger != nil {
			c.logger.Info("ar.mcqs.entry", zap.Uint16("comp", comp), zap.Uint16("identifier", ident), zap.Uint8("status", status), zap.Bool("last", lastFlag), zap.Uint8("dev_type", dtype))
		}
		if ident == 1 && status != 0 {
			return comp, dtype, true
		}
		last = lastFlag
	}
	return 0, 0, false
}

// decodeBCD decodes a single packed BCD byte (00-99) to integer; returns -1 if invalid
func decodeBCD(b byte) int {
	hi := int((b >> 4) & 0xF)
	lo := int(b & 0xF)
	if hi > 9 || lo > 9 {
		return -1
	}
	return hi*10 + lo
}

// decodeBCD16 decodes a 2-byte packed BCD (yyyy) to integer year; returns -1 if invalid
func decodeBCD16(hi, lo byte) int {
	d1 := decodeBCD(hi)
	d2 := decodeBCD(lo)
	if d1 < 0 || d2 < 0 {
		return -1
	}
	return d1*100 + d2
}

// transact is a high-level helper that would write a command structure and read back a response.
// Placeholder for concrete MGIR/MCQI encodings.
func (c *ARClient) transact(cmd []byte, resp []byte) error {
	// Send/receive via ICMD mailbox
	if len(resp) < len(cmd) {
		return fmt.Errorf("resp buffer too small")
	}
	tmp := make([]byte, len(cmd))
	copy(tmp, cmd)
	if err := c.sendICMD(tmp); err != nil {
		return err
	}
	copy(resp, tmp)
	return nil
}

// QueryFirmwareInfo retrieves running/pending FW versions, PSID, and security bits via MGIR/MCQI.
// TODO: implement encoding/decoding per device family.
func (c *ARClient) QueryFirmwareInfo() (map[string]interface{}, error) {
	// Build a result map seeded with sysfs-derived fields when possible,
	// then enrich via MGIR/MCQI/ROM where supported. Do not early-return
	// so we can populate fw_release_date and rom_info for sysfs backend too.
	res := make(map[string]interface{})
	// Always attempt to populate sysfs-derived fields first
	if sb, ok := c.dev.(*sysfsBackend); ok || c.bdf != "" {
		bdf := c.bdf
		if ok {
			bdf = sb.bdf
		}
		if bdf != "" {
			readText := func(p string) string {
				b, err := os.ReadFile(p)
				if err != nil {
					return ""
				}
				return strings.TrimSpace(string(b))
			}
			root := filepath.Join("/sys/bus/pci/devices", bdf)
			res["VendorID"] = readText(filepath.Join(root, "vendor"))
			res["DeviceID"] = readText(filepath.Join(root, "device"))
			// Set a conservative default image type for NIC families; MGIR may override later
			res["ImageType"] = "FS4"
			if entries, err := os.ReadDir(filepath.Join(root, "infiniband")); err == nil {
				for _, e := range entries {
					fw := readText(filepath.Join(root, "infiniband", e.Name(), "fw_ver"))
					if fw != "" {
						res["FWVersion"] = fw
						// Try to extract release date from the fw_ver string if present (dd.mm.yyyy, with sane ranges)
						if _, ok := res["FWReleaseDate"]; !ok || res["FWReleaseDate"] == "" {
							if re := regexp.MustCompile(`\b(\d{2})\.(\d{2})\.(\d{4})\b`); re != nil {
								if m := re.FindStringSubmatch(fw); len(m) == 4 {
									// Validate plausible date: day 1..31, month 1..12
									var dd, mm int
									fmt.Sscanf(m[1], "%d", &dd)
									fmt.Sscanf(m[2], "%d", &mm)
									if dd >= 1 && dd <= 31 && mm >= 1 && mm <= 12 {
										res["FWReleaseDate"] = m[0]
									}
								}
							}
						}
						break
					}
				}
				for _, e := range entries {
					ps := readText(filepath.Join(root, "infiniband", e.Name(), "board_id"))
					if ps != "" {
						res["PSID"] = ps
						break
					}
				}
				// Base GUID
				for _, e := range entries {
					ng := readText(filepath.Join(root, "infiniband", e.Name(), "node_guid"))
					if ng != "" {
						s := strings.ToLower(ng)
						s = strings.TrimPrefix(s, "0x")
						s = strings.ReplaceAll(s, ":", "")
						if len(s) >= 16 {
							var v uint64
							fmt.Sscanf(s[:16], "%16x", &v)
							res["BaseGUID"] = v
							if pentries, err := os.ReadDir(filepath.Join(root, "infiniband", e.Name(), "ports")); err == nil {
								res["BaseGUIDNum"] = len(pentries)
							}
						}
						break
					}
				}
			}
			// Base MAC
			if nentries, err := os.ReadDir(filepath.Join(root, "net")); err == nil {
				base := ""
				count := 0
				for _, ne := range nentries {
					addr := readText(filepath.Join(root, "net", ne.Name(), "address"))
					if addr == "" {
						continue
					}
					count++
					if base == "" || strings.Compare(addr, base) < 0 {
						base = addr
					}
				}
				if base != "" {
					s := strings.ReplaceAll(base, ":", "")
					var mv uint64
					fmt.Sscanf(s, "%12x", &mv)
					res["BaseMAC"] = mv
					res["BaseMACNum"] = count
				}
			}
		}
	}
	// Attempt MGIR access register to fill richer data
	// MGIR ext size is 0xa0 bytes (from reg_access_hca_layouts.h)
	const mgirSize = 0xa0
	payload, err := c.transactAR(RegID_MGIR, mgirSize)
	if err == nil && len(payload) == mgirSize {
		// Log a small window regardless to debug availability
		if c.logger != nil {
			dumpLen := 64
			if dumpLen > len(payload) {
				dumpLen = len(payload)
			}
			start := 0x20
			end := 0x50
			if end > len(payload) {
				end = len(payload)
			}
			c.logger.Info("ar.mgir.raw",
				zap.Int("size", len(payload)),
				zap.String("bytes", hex.EncodeToString(payload[:dumpLen])),
			)
			if end > start {
				c.logger.Info("ar.mgir.fw_info.window", zap.String("0x20_0x50", hex.EncodeToString(payload[start:end])))
			}
		}
		// Only trust MGIR if payload is non-zero
		nonZero := false
		for _, bb := range payload {
			if bb != 0 {
				nonZero = true
				break
			}
		}
		if nonZero {
			info := decodeMGIRFWInfo(payload)
			// Only accept MGIR version when it matches NIC-style X.Y.ZZZZ
			if info.FWVersion != "" && validFWVersionString(info.FWVersion) {
				res["FWVersion"] = info.FWVersion
				res["ProductVersion"] = info.FWVersion
				if info.SecurityAttrs != "" {
					res["SecurityAttrs"] = info.SecurityAttrs
				}
			}
			res["SecurityVer"] = info.SecurityVer
			if info.FWReleaseDate != "" {
				res["FWReleaseDate"] = info.FWReleaseDate
			}
			if info.PSID != "" {
				res["PSID"] = info.PSID
			}
			// Heuristic: try to extract ASCII date dd.mm.yyyy from MGIR payload if present
			if _, ok := res["FWReleaseDate"]; !ok || res["FWReleaseDate"] == "" {
				for i := 0; i+10 <= len(payload); i++ {
					b := payload[i:]
					if i+10 <= len(payload) &&
						b[0] >= '0' && b[0] <= '9' && b[1] >= '0' && b[1] <= '9' &&
						b[2] == '.' &&
						b[3] >= '0' && b[3] <= '9' && b[4] >= '0' && b[4] <= '9' &&
						b[5] == '.' &&
						b[6] >= '0' && b[6] <= '9' && b[7] >= '0' && b[7] <= '9' && b[8] >= '0' && b[8] <= '9' && b[9] >= '0' && b[9] <= '9' {
						cand := string(payload[i : i+10])
						var dd, mm int
						fmt.Sscanf(cand[0:2], "%d", &dd)
						fmt.Sscanf(cand[3:5], "%d", &mm)
						if dd >= 1 && dd <= 31 && mm >= 1 && mm <= 12 {
							res["FWReleaseDate"] = cand
							break
						}
					}
				}
			}
		}
	} else {
		// Fallback: try MCQI to retrieve at least basic version context (decode TBD)
		const mcqiSize = 0x94
		if c.logger != nil {
			c.logger.Info("ar.query.mcqi.attempt")
		}
		if p2, err2 := c.transactAR(RegID_MCQI, mcqiSize); err2 == nil && len(p2) == mcqiSize {
			if c.logger != nil {
				c.logger.Info("ar.query.mcqi.ok", zap.Int("size", len(p2)))
			}
			mcqi := decodeMCQI(p2)
			if mcqi.ActivationMethod != "" {
				res["ActivationMethod"] = mcqi.ActivationMethod
			}
		} else if err2 != nil && c.logger != nil {
			c.logger.Info("ar.query.mcqi.fail", zap.Error(err2))
		}
	}

	// Attempt ROM info extraction from expansion ROM for device-mode parity
	if entries, err := extractROMInfo(c.dev); err == nil && len(entries) > 0 {
		// Attach as a generic array; caller can map to FirmwareInfo
		rom := make([]map[string]string, 0, len(entries))
		for _, e := range entries {
			rom = append(rom, map[string]string{"type": e.Type, "version": e.Version, "cpu": e.CPU})
		}
		res["ROMInfo"] = rom
	}

	// Discover boot component index via MCQS, then query MCQI for VERSION/ACTIVATION
	bootIndex, devType, haveBoot := c.findBootComponent()
	if c.logger != nil {
		c.logger.Info("ar.mcqs.boot_index", zap.Bool("ok", haveBoot), zap.Uint16("index", bootIndex), zap.Uint8("dev_type", devType))
	}
	// Explicit MCQI pulls for activation method and (optionally) version
	const mcqiSize = 0x94
	// Helper to build a parameterized MCQI request
	mcqiRequest := func(infoType byte, dataSize uint16) []byte {
		p, e := c.transactARWithInit(RegID_MCQI, mcqiSize, func(pl []byte) {
			if len(pl) < mcqiSize {
				return
			}
			// Pack MCQI header using bit-accurate mapping (big-endian bit order)
			packMCQIHeader(pl[:mcqiSize], infoType, dataSize)
			if haveBoot {
				// Override component/device indices if discovered
				pushBitsBE(pl, 16, 16, uint32(bootIndex))
				pushBitsBE(pl, 120, 8, uint32(devType))
			}
		})
		if e != nil {
			return nil
		}
		return p
	}
	// ACTIVATION_METHOD (info_type=0x5), ext size 0x7c
	if p := mcqiRequest(0x5, 0x7c); len(p) == mcqiSize {
		// Record header fields for JSON
		if len(p) >= 0x16 {
			res["MCQI_Activation_InfoType"] = p[0x8]
			res["MCQI_Activation_InfoSize"] = binary.LittleEndian.Uint32(p[0x0c:0x10])
			res["MCQI_Activation_DataSize"] = binary.LittleEndian.Uint16(p[0x14:0x16])
		}
		mcqi := decodeMCQI(p)
		if mcqi.ActivationMethod != "" {
			res["ActivationMethod"] = mcqi.ActivationMethod
			res["MCQI_ActivationMethod"] = mcqi.ActivationMethod
			if c.logger != nil {
				c.logger.Info("ar.query.mcqi.activation", zap.String("method", mcqi.ActivationMethod))
			}
		}
	}
	// Always fetch MCQI VERSION (info_type=0x1) to capture version string; do not override MGIR FWVersion
	if p := mcqiRequest(0x1, 0x7c); len(p) == mcqiSize {
		if len(p) >= 0x16 {
			res["MCQI_Version_InfoType"] = p[0x8]
			res["MCQI_Version_InfoSize"] = binary.LittleEndian.Uint32(p[0x0c:0x10])
			res["MCQI_Version_DataSize"] = binary.LittleEndian.Uint16(p[0x14:0x16])
		}
		if c.logger != nil {
			dumpLen := 64
			if dumpLen > len(p) {
				dumpLen = len(p)
			}
			c.logger.Info("ar.mcqi.version.raw",
				zap.Int("size", len(p)),
				zap.String("bytes", hex.EncodeToString(p[:dumpLen])),
				zap.Uint8("info_type", p[0x8]),
				zap.Uint32("info_size_le", binary.LittleEndian.Uint32(p[0x0c:0x10])),
				zap.Uint16("data_size_le", binary.LittleEndian.Uint16(p[0x14:0x16])),
				zap.Uint32("info_size_be", binary.BigEndian.Uint32(p[0x0c:0x10])),
				zap.Uint16("data_size_be", binary.BigEndian.Uint16(p[0x14:0x16])),
			)
			if len(p) >= 0x28 {
				c.logger.Info("ar.mcqi.version.build_time.bytes",
					zap.String("bt_0x20_0x28", hex.EncodeToString(p[0x20:0x28])),
				)
			}
		}
		// Extract build_time from MCQI VERSION data union using date_time_layout_ext (packed BCD)
		if len(p) >= 0x20+0x8 {
			base := 0x20 // 0x18 (union start) + 0x8 (build_time)
			// dword[0]: [unused][hours][minutes][seconds]
			secB := p[base+3]
			minB := p[base+2]
			hrB := p[base+1]
			// dword[1]: [day][month][year_hi][year_lo] in LE byte order within the dword
			dayB := p[base+4]
			monthB := p[base+5]
			yearHi := p[base+6]
			yearLo := p[base+7]
			sec := decodeBCD(secB)
			min := decodeBCD(minB)
			hr := decodeBCD(hrB)
			year := decodeBCD16(yearHi, yearLo)
			day := decodeBCD(dayB)
			month := decodeBCD(monthB)
			if c.logger != nil {
				c.logger.Info("ar.mcqi.version.build_time.fields",
					zap.Int("hr", hr), zap.Int("min", min), zap.Int("sec", sec),
					zap.Int("day", day), zap.Int("month", month), zap.Int("year", year))
			}
			if day >= 1 && day <= 31 && month >= 1 && month <= 12 && year >= 2000 && year <= 2099 {
				res["FWReleaseDate"] = fmt.Sprintf("%02d.%02d.%04d", day, month, year)
			}
		}
		// Extract user_defined_time (often used for release date) — also BCD
		if len(p) >= 0x28+0x8 {
			base := 0x28 // 0x18 + 0x10 (user_defined_time)
			if c.logger != nil {
				c.logger.Info("ar.mcqi.version.user_time.bytes",
					zap.String("ut_0x28_0x30", hex.EncodeToString(p[base:base+8])),
				)
			}
			sec := decodeBCD(p[base+3])
			min := decodeBCD(p[base+2])
			hr := decodeBCD(p[base+1])
			day := decodeBCD(p[base+4])
			month := decodeBCD(p[base+5])
			year := decodeBCD16(p[base+6], p[base+7])
			if c.logger != nil {
				c.logger.Info("ar.mcqi.version.user_time.fields",
					zap.Int("hr", hr), zap.Int("min", min), zap.Int("sec", sec),
					zap.Int("day", day), zap.Int("month", month), zap.Int("year", year))
			}
			if day >= 1 && day <= 31 && month >= 1 && month <= 12 && year >= 2000 && year <= 2099 {
				// Prefer user-defined time as release date if valid
				res["FWReleaseDate"] = fmt.Sprintf("%02d.%02d.%04d", day, month, year)
			}
		}
		if len(p) > 0x38 {
			vs := p[0x38:]
			end := 0
			for end < len(vs) && vs[end] != 0 {
				end++
			}
			if end > 0 {
				verStr := string(vs[:end])
				res["MCQI_VersionString"] = verStr
				// Heuristic: extract release date like dd.mm.yyyy if present; validate ranges to avoid mistaking versions
				if re := regexp.MustCompile(`\b(\d{2})\.(\d{2})\.(\d{4})\b`); re != nil {
					if m := re.FindStringSubmatch(verStr); len(m) == 4 {
						var dd, mm int
						fmt.Sscanf(m[1], "%d", &dd)
						fmt.Sscanf(m[2], "%d", &mm)
						if dd >= 1 && dd <= 31 && mm >= 1 && mm <= 12 {
							res["FWReleaseDate"] = m[0]
						}
					}
				}
				// Heuristic: prefer product/fw version in form X.Y.ZZZZ (last part 4 digits)
				// If FWVersion is missing or not in NIC-style pattern, prefer version from MCQI string
				cur, has := res["FWVersion"].(string)
				if !has || !validFWVersionString(cur) {
					if v := extractFWVersionFromString(verStr); v != "" {
						res["FWVersion"] = v
						res["ProductVersion"] = v
					}
				}
			}
		}
	}
	// CAPABILITIES (info_type=0x0)
	if p := mcqiRequest(0x0, 0x7c); len(p) == mcqiSize {
		if len(p) >= 0x24 {
			res["MCQI_Cap_InfoType"] = p[0x8]
			res["MCQI_Cap_InfoSize"] = binary.LittleEndian.Uint32(p[0x0c:0x10])
			res["MCQI_Cap_DataSize"] = binary.LittleEndian.Uint16(p[0x14:0x16])
			supp := binary.LittleEndian.Uint32(p[0x18:0x1c])
			res["MCQI_SupportedInfoMask"] = supp
			compSize := binary.LittleEndian.Uint32(p[0x1c:0x20])
			res["MCQI_ComponentSize"] = compSize
			maxComp := binary.LittleEndian.Uint32(p[0x20:0x24])
			res["MCQI_MaxComponentSize"] = maxComp
		}
	}
	return res, nil
}

// TransactRaw issues a GET for regID and returns raw payload of regSize bytes.
func (c *ARClient) TransactRaw(regID uint16, regSize int) ([]byte, error) {
	return c.transactAR(regID, regSize)
}

// TransactFull issues a GET and returns the entire TLV buffer (OP TLV + REG TLV + payload).
func (c *ARClient) TransactFull(regID uint16, regSize int) ([]byte, error) {
	const opTLVSize = 16
	const regHdr = 4
	total := opTLVSize + regHdr + regSize
	// Pack TLV in the same LE form used by transactAR; sendICMD flips headers to BE
	word0 := (uint32(1) & 0x1f) | (uint32(4) << 5)
	word1 := (uint32(regID) & 0xffff) | (uint32(RegMethodGet) << 17) | (uint32(1) << 24)
	dwords := uint16((regHdr + regSize) / 4)
	regHdrWord := (uint32(3) & 0x1f) | (uint32(dwords) << 5)
	buf := make([]byte, total)
	// op tlv (LE)
	binary.LittleEndian.PutUint32(buf[0:4], word0)
	binary.LittleEndian.PutUint32(buf[4:8], word1)
	// tid = 0
	for i := 8; i < 16; i++ {
		buf[i] = 0
	}
	// reg hdr (LE)
	binary.LittleEndian.PutUint32(buf[16:20], regHdrWord)
	if err := c.sendICMD(buf); err != nil {
		return nil, err
	}
	return buf, nil
}

// validFWVersionString returns true if s matches X.Y.ZZZZ (last part 4 digits)
func validFWVersionString(s string) bool {
	re := regexp.MustCompile(`^\d{1,2}\.\d{1,2}\.\d{4}$`)
	return re.MatchString(s)
}

// extractFWVersionFromString tries to find and return a NIC-style version X.Y.ZZZZ in text
func extractFWVersionFromString(text string) string {
	re := regexp.MustCompile(`\b\d{1,2}\.\d{1,2}\.\d{4}\b`)
	if m := re.FindString(text); m != "" {
		return m
	}
	return ""
}
