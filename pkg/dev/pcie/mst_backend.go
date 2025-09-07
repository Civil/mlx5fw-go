//go:build pcie_enabled
// +build pcie_enabled

package pcie

import (
	"fmt"
	"os"
	"strings"

	"go.uber.org/zap"
)

// mstBackend implements Device via /dev/mst pciconf IOCTLs.
type mstBackend struct {
	logger *zap.Logger
	path   string
	fd     *os.File
}

func openMST(path string, logger *zap.Logger) (Device, error) {
	// Allow specifying directory or device name; pick first pciconf entry if dir
	p := path
	if fi, err := os.Stat(path); err == nil && fi.IsDir() {
		entries, _ := os.ReadDir(path)
		for _, e := range entries {
			if strings.Contains(e.Name(), "pciconf") || strings.Contains(e.Name(), "_mstconf") {
				p = path + "/" + e.Name()
				break
			}
		}
	}
	f, err := os.OpenFile(p, os.O_RDWR, 0)
	if err != nil {
		return nil, fmt.Errorf("open MST device: %w", err)
	}
	return &mstBackend{logger: logger, path: p, fd: f}, nil
}

func (m *mstBackend) Type() string { return "mst" }
func (m *mstBackend) Close() error { return m.fd.Close() }

// ioctl encoding helpers (mirrors linux/ioctl.h macros)
const (
	iocNRBits   = 8
	iocTypeBits = 8
	iocSizeBits = 14
	iocDirBits  = 2

	iocNRShift   = 0
	iocTypeShift = iocNRShift + iocNRBits
	iocSizeShift = iocTypeShift + iocTypeBits
	iocDirShift  = iocSizeShift + iocSizeBits

	iocNone  = 0
	iocWrite = 1
	iocRead  = 2
)

func _IOC(dir, typ, nr, size uint32) uint32 {
	return (dir << iocDirShift) | (typ << iocTypeShift) | (nr << iocNRShift) | (size << iocSizeShift)
}
func _IOR(typ, nr, size uint32) uint32 { return _IOC(iocRead, typ, nr, size) }
func _IOW(typ, nr, size uint32) uint32 { return _IOC(iocWrite, typ, nr, size) }

// Magic numbers from reference/mstflint/kernel/mst.h
const (
	mstByteAccessMagic  = 0xD1
	mstBlockAccessMagic = 0xD2
	mstVpdMagic         = 0xD6
)

// Structs must match the kernel layout
type mstRead4 struct {
	AddressSpace uint32
	Offset       uint32
	Data         uint32 // OUT
}
type mstWrite4 struct {
	AddressSpace uint32
	Offset       uint32
	Data         uint32
}

// IOCTL numbers
var (
	ioctlPCICONFRead4  = _IOR(mstByteAccessMagic, 1, 12)
	ioctlPCICONFWrite4 = _IOW(mstByteAccessMagic, 2, 12)
)

// Extra IOCTLs and structs for debug/inspection
const (
	mstParamsMagic  = 0xD0
	mstPCICONFMagic = 0xD3
)

type mstParams struct {
	Domain               uint32
	Bus                  uint32
	Slot                 uint32
	Func                 uint32
	Bar                  uint32
	Device               uint32
	Vendor               uint32
	SubsystemDevice      uint32
	SubsystemVendor      uint32
	FunctionalVsecOffset uint32
}

type mstCfgReadDword struct {
	Offset uint32
	Data   uint32
}

var (
	ioctlMSTParams           = _IOR(mstParamsMagic, 1, 40)
	ioctlPCICONFReadCfgDword = _IOR(mstPCICONFMagic, 15, 8)
)

// MSTParams returns kernel-provided device parameters if backend is MST.
func MSTParams(dev Device) (*mstParams, error) {
	m, ok := dev.(*mstBackend)
	if !ok {
		return nil, fmt.Errorf("MST backend required")
	}
	return cgoMSTParams(m.fd.Fd())
}

// ReadPCIConfigDword reads a dword from PCI config space via MST PCICONF ioctl.
func ReadPCIConfigDword(dev Device, offset uint32) (uint32, error) {
	m, ok := dev.(*mstBackend)
	if !ok {
		return 0, fmt.Errorf("MST backend required")
	}
	return cgoReadPCIConfigDword(m.fd.Fd(), offset)
}

// mstParamsOnPath opens a raw MST device path and fetches MST_PARAMS.
func mstParamsOnPath(path string) (*mstParams, error) {
	f, err := os.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return cgoMSTParams(f.Fd())
}

// ProbeMSTNode opens a node, fetches MST params, and returns derived BDF.
func ProbeMSTNode(path string) (string, *mstParams, error) {
	f, err := os.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		return "", nil, err
	}
	defer f.Close()
	p, err := cgoMSTParams(f.Fd())
	if err != nil {
		return "", nil, err
	}
	bdf := fmt.Sprintf("%04x:%02x:%02x.%d", p.Domain, p.Bus, p.Slot, p.Func)
	return bdf, p, nil
}

func (m *mstBackend) Read32(space uint16, offset uint32) (uint32, error) {
	return cgoPCICONFRead4(m.fd.Fd(), space, offset)
}

func (m *mstBackend) Write32(space uint16, offset uint32, value uint32) error {
	return cgoPCICONFWrite4(m.fd.Fd(), space, offset, value)
}

func (m *mstBackend) ReadBlock(space uint16, offset uint32, size int) ([]byte, error) {
	// Simple loop using READ4 until we wire buffered IOCTLs
	out := make([]byte, 0, size)
	for i := 0; i < size; i += 4 {
		v, err := m.Read32(space, offset+uint32(i))
		if err != nil {
			return nil, err
		}
		out = append(out, byte(v>>24), byte(v>>16), byte(v>>8), byte(v))
	}
	return out[:size], nil
}

func (m *mstBackend) WriteBlock(space uint16, offset uint32, data []byte) error {
	for i := 0; i < len(data); i += 4 {
		var v uint32
		end := i + 4
		if end > len(data) {
			end = len(data)
		}
		for j := i; j < end; j++ {
			v |= uint32(data[j]) << (8 * (3 - (j - i)))
		}
		if err := m.Write32(space, offset+uint32(i), v); err != nil {
			return err
		}
	}
	return nil
}
