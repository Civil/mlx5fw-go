package types

// SecurityModeMask defines the security mode flags
// Based on mstflint's security_mode_mask enum
type SecurityModeMask struct {
	MCC_EN                uint32
	DEBUG_FW              uint32
	SIGNED_FW             uint32
	SECURE_FW             uint32
	DEV_FW                uint32
	CS_TOKEN              uint32
	DBG_TOKEN             uint32
	CRYPTO_TO_COMMISSIONING uint32
	RMCS_TOKEN            uint32
	RMDT_TOKEN            uint32
}

// SMMFlags contains the security mode mask values
var SMMFlags = SecurityModeMask{
	MCC_EN:                0x1,
	DEBUG_FW:              0x1 << 1,
	SIGNED_FW:             0x1 << 2,
	SECURE_FW:             0x1 << 3,
	DEV_FW:                0x1 << 4,
	CS_TOKEN:              0x1 << 5,
	DBG_TOKEN:             0x1 << 6,
	CRYPTO_TO_COMMISSIONING: 0x1 << 7,
	RMCS_TOKEN:            0x1 << 8,
	RMDT_TOKEN:            0x1 << 9,
}