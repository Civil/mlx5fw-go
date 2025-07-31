package sections

import (
	crcpkg "github.com/Civil/mlx5fw-go/pkg/crc"
	"github.com/Civil/mlx5fw-go/pkg/interfaces"
	"github.com/Civil/mlx5fw-go/pkg/parser"
	"github.com/Civil/mlx5fw-go/pkg/types"
	"github.com/ansel1/merry/v2"
)

// DefaultSectionFactory is the default implementation of SectionFactory
type DefaultSectionFactory struct{
	crcCalculator interfaces.CRCCalculator
}

// NewDefaultSectionFactory creates a new default section factory
func NewDefaultSectionFactory() *DefaultSectionFactory {
	return &DefaultSectionFactory{
		crcCalculator: parser.NewCRCCalculator(),
	}
}

// CreateSection creates a new section instance based on the section type
func (f *DefaultSectionFactory) CreateSection(sectionType uint16, offset uint64, size uint32, 
	crcType types.CRCType, crc uint32, isEncrypted, isDeviceData bool, 
	entry *types.ITOCEntry, isFromHWPointer bool) (interfaces.SectionInterface, error) {
	
	// Create base section
	base := interfaces.NewBaseSection(sectionType, offset, size, crcType, crc, 
		isEncrypted, isDeviceData, entry, isFromHWPointer)
	
	// Determine CRC handler based on CRC type and section type
	var crcHandler interfaces.SectionCRCHandler
	
	// First check if this section type uses hardware or software CRC algorithm
	crcAlgorithm := types.GetSectionCRCAlgorithm(sectionType)
	
	switch crcType {
	case types.CRCInSection:
		// Sections with CRC at the end
		if crcAlgorithm == types.CRCAlgorithmHardware {
			// HW pointers use hardware CRC
			crcHandler = crcpkg.NewHardwareCRC16Handler(f.crcCalculator)
		} else {
			// All other sections use software CRC16
			crcHandler = crcpkg.NewInSectionCRC16Handler(f.crcCalculator)
		}
	case types.CRCInITOCEntry:
		// Sections with CRC in ITOC entry always use software CRC16
		crcHandler = crcpkg.NewSoftwareCRC16Handler(f.crcCalculator)
	case types.CRCNone:
		// No CRC handler needed
		crcHandler = crcpkg.NewNoCRCHandler()
	default:
		// Default to no CRC
		crcHandler = crcpkg.NewNoCRCHandler()
	}
	base.SetCRCHandler(crcHandler)
	
	// Create specific section type based on sectionType
	switch sectionType {
	case types.SectionTypeItoc:
		return NewITOCSection(base), nil
		
	case types.SectionTypeDtoc:
		return NewDTOCSection(base), nil
		
	case types.SectionTypeImageInfo:
		return NewImageInfoSection(base), nil
		
	case types.SectionTypeDevInfo, types.SectionTypeDevInfo1, types.SectionTypeDevInfo2:
		return NewDeviceInfoSection(base), nil
		
	case types.SectionTypeMfgInfo:
		return NewMFGInfoSection(base), nil
		
	case types.SectionTypeHwPtr:
		return NewHWPointerSection(base), nil
		
	case types.SectionTypeHashesTable:
		return NewHashesTableSection(base), nil
		
	case types.SectionTypeBoot2:
		// BOOT2 has special structure
		return NewBoot2Section(base), nil
		
	case types.SectionTypeToolsArea:
		return NewToolsAreaExtendedSection(base), nil
		
	case types.SectionTypeImageSignature256:
		return NewImageSignatureSection(base), nil
		
	case types.SectionTypeImageSignature512:
		return NewImageSignature2Section(base), nil
		
	case types.SectionTypePublicKeys2048:
		return NewPublicKeysSection(base), nil
		
	case types.SectionTypePublicKeys4096:
		return NewPublicKeys2Section(base), nil
		
	case types.SectionTypeResetInfo:
		return NewResetInfoSection(base), nil
		
	case types.SectionTypeForbiddenVersions:
		return NewForbiddenVersionsSection(base), nil
		
	case types.SectionTypeDbgFWINI:
		return NewDBGFwIniSection(base), nil
		
	case types.SectionTypeDbgFWParams:
		return NewDBGFwParamsSection(base), nil
		
	case types.SectionTypeFWAdb:
		return NewFWAdbSection(base), nil
		
	// DTOC sections
	case types.SectionTypeVpdR0:
		return NewVPD_R0Section(base), nil
		
	case types.SectionTypeFwNvLog:
		return NewFWNVLogSection(base), nil
		
	case types.SectionTypeNvData0, types.SectionTypeNvData1, types.SectionTypeNvData2:
		return NewNVDataSection(base), nil
		
	case types.SectionTypeCRDumpMaskData:
		return NewCRDumpMaskDataSection(base), nil
		
	case types.SectionTypeFwInternalUsage:
		return NewFWInternalUsageSection(base), nil
		
	case types.SectionTypeProgrammableHwFw1, types.SectionTypeProgrammableHwFw2:
		return NewProgrammableHWFWSection(base), nil
		
	case types.SectionTypeDigitalCertPtr:
		return NewDigitalCertPtrSection(base), nil
		
	case types.SectionTypeDigitalCertRw:
		return NewDigitalCertRWSection(base), nil
		
	// Add more specific section types as needed
	
	default:
		// Return generic section for unknown types
		return NewGenericSection(base), nil
	}
}

// CreateSectionFromData creates a section and parses its data
func (f *DefaultSectionFactory) CreateSectionFromData(sectionType uint16, offset uint64, 
	size uint32, crcType types.CRCType, crc uint32, isEncrypted, isDeviceData bool, 
	entry *types.ITOCEntry, isFromHWPointer bool, data []byte) (interfaces.SectionInterface, error) {
	
	// Create the section
	section, err := f.CreateSection(sectionType, offset, size, crcType, crc, 
		isEncrypted, isDeviceData, entry, isFromHWPointer)
	if err != nil {
		return nil, err
	}
	
	// Parse the data
	if err := section.Parse(data); err != nil {
		return nil, merry.Wrap(err)
	}
	
	return section, nil
}