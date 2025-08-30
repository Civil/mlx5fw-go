//go:build ignore
// +build linux,cgo

package pcie

/*
#include <stdlib.h>
#include <stdint.h>
#include <sys/ioctl.h>
#include <errno.h>
#include "../../../reference/mstflint/kernel/mst.h"

static int go_mst_params(int fd, struct mst_params *p) {
    int rc = ioctl(fd, MST_PARAMS, p);
    if (rc != 0) return errno;
    return 0;
}

static int go_pciconf_read4(int fd, struct mst_read4_st *st) {
    int rc = ioctl(fd, PCICONF_READ4, st);
    if (rc != 0) return errno;
    return 0;
}

static int go_pciconf_write4(int fd, struct mst_write4_st *st) {
    int rc = ioctl(fd, PCICONF_WRITE4, st);
    if (rc != 0) return errno;
    return 0;
}

static int go_pciconf_read_cfg_dword(int fd, struct read_dword_from_config_space *st) {
    int rc = ioctl(fd, PCICONF_READ_DWORD_FROM_CONFIG_SPACE, st);
    if (rc != 0) return errno;
    return 0;
}

static int go_get_mst_params_flat(int fd, uint32_t out[10]) {
    struct mst_params p;
    int rc = ioctl(fd, MST_PARAMS, &p);
    if (rc != 0) { return rc; }
    out[0] = p.domain;
    out[1] = p.bus;
    out[2] = p.slot;
    out[3] = p.func;
    out[4] = p.bar;
    out[5] = p.device;
    out[6] = p.vendor;
    out[7] = p.subsystem_device;
    out[8] = p.subsystem_vendor;
    out[9] = p.functional_vsc_offset;
    return 0;
}
*/
import "C"

import (
	"fmt"
	"unsafe"
)

func cgoMSTParams(fd uintptr) (*mstParams, error) {
	var out [10]C.uint32_t
	if rc := C.go_get_mst_params_flat(C.int(fd), &out[0]); rc != 0 {
		return nil, fmt.Errorf("MST_PARAMS ioctl errno=%d", int(rc))
	}
	st := &mstParams{
		Domain:               uint32(out[0]),
		Bus:                  uint32(out[1]),
		Slot:                 uint32(out[2]),
		Func:                 uint32(out[3]),
		Bar:                  uint32(out[4]),
		Device:               uint32(out[5]),
		Vendor:               uint32(out[6]),
		SubsystemDevice:      uint32(out[7]),
		SubsystemVendor:      uint32(out[8]),
		FunctionalVsecOffset: uint32(out[9]),
	}
	return st, nil
}

func cgoPCICONFRead4(fd uintptr, space uint16, offset uint32) (uint32, error) {
	var st C.struct_mst_read4_st
	st.address_space = C.uint(space)
	st.offset = C.uint(offset)
	if rc := C.go_pciconf_read4(C.int(fd), &st); rc != 0 {
		return 0, fmt.Errorf("PCICONF_READ4 ioctl errno=%d", int(rc))
	}
	return uint32(st.data), nil
}

func cgoPCICONFWrite4(fd uintptr, space uint16, offset uint32, value uint32) error {
	var st C.struct_mst_write4_st
	st.address_space = C.uint(space)
	st.offset = C.uint(offset)
	st.data = C.uint(value)
	if rc := C.go_pciconf_write4(C.int(fd), &st); rc != 0 {
		return fmt.Errorf("PCICONF_WRITE4 ioctl errno=%d", int(rc))
	}
	return nil
}

func cgoReadPCIConfigDword(fd uintptr, offset uint32) (uint32, error) {
	var st C.struct_read_dword_from_config_space
	st.offset = C.uint(offset)
	if rc := C.go_pciconf_read_cfg_dword(C.int(fd), &st); rc != 0 {
		return 0, fmt.Errorf("PCICONF_READ_DWORD_FROM_CONFIG_SPACE ioctl errno=%d", int(rc))
	}
	return uint32(st.data), nil
}

var _ = unsafe.Sizeof(C.struct_mst_read4_st{})
