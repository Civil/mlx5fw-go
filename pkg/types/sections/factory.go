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