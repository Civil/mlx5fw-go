package crc

import (
	"errors"
	"testing"

	pkgerrors "github.com/Civil/mlx5fw-go/pkg/errors"
	"github.com/Civil/mlx5fw-go/pkg/types"
	"github.com/stretchr/testify/assert"
)

// MockCRCCalculator for testing
type MockCRCCalculator struct {
	result uint32
}

func (m *MockCRCCalculator) Calculate(data []byte) uint32 {
	return m.result
}

func (m *MockCRCCalculator) CalculateWithParams(data []byte, polynomial, initial, xorOut uint32) uint32 {
	return m.result
}

func (m *MockCRCCalculator) CalculateImageCRC(data []byte, sizeInDwords int) uint16 {
	return uint16(m.result)
}

func (m *MockCRCCalculator) GetType() types.CRCType {
	return types.CRCInITOCEntry
}

func TestBaseCRCHandler(t *testing.T) {
	mockCalc := &MockCRCCalculator{result: 0x1234}
	
	t.Run("NewBaseCRCHandler", func(t *testing.T) {
		handler := NewBaseCRCHandler(mockCalc, -4, true)
		assert.NotNil(t, handler)
		assert.Equal(t, -4, handler.GetCRCOffset())
		assert.True(t, handler.HasEmbeddedCRC())
		assert.Equal(t, mockCalc, handler.GetCalculator())
	})
	
	t.Run("VerifyCRC", func(t *testing.T) {
		handler := NewBaseCRCHandler(mockCalc, -4, true)
		
		// Mock calculate function
		calculateFunc := func(data []byte, crcType types.CRCType) (uint32, error) {
			return 0x1234, nil
		}
		
		// Test successful verification
		err := handler.VerifyCRC([]byte{}, 0x1234, types.CRCInITOCEntry, calculateFunc)
		assert.NoError(t, err)
		
		// Test failed verification
		err = handler.VerifyCRC([]byte{}, 0x5678, types.CRCInITOCEntry, calculateFunc)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, pkgerrors.ErrCRCMismatch))
		
		// Verify error contains correct data
		crcData, ok := pkgerrors.GetCRCMismatchData(err)
		assert.True(t, ok)
		assert.Equal(t, uint32(0x5678), crcData.Expected)
		assert.Equal(t, uint32(0x1234), crcData.Actual)
	})
	
	t.Run("VerifyCRC16", func(t *testing.T) {
		handler := NewBaseCRCHandler(mockCalc, -4, true)
		
		// Mock calculate function that returns a 32-bit value
		calculateFunc := func(data []byte, crcType types.CRCType) (uint32, error) {
			return 0xABCD1234, nil // Only lower 16 bits (0x1234) should be used
		}
		
		// Test successful verification (only lower 16 bits compared)
		err := handler.VerifyCRC16([]byte{}, 0xFFFF1234, types.CRCInSection, calculateFunc)
		assert.NoError(t, err)
		
		// Test failed verification
		err = handler.VerifyCRC16([]byte{}, 0x5678, types.CRCInSection, calculateFunc)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, pkgerrors.ErrCRCMismatch))
		
		// Verify error contains correct 16-bit values
		crcData, ok := pkgerrors.GetCRCMismatchData(err)
		assert.True(t, ok)
		assert.Equal(t, uint32(0x5678), crcData.Expected)
		assert.Equal(t, uint32(0x1234), crcData.Actual)
	})
	
	t.Run("ValidateCRCType", func(t *testing.T) {
		handler := NewBaseCRCHandler(mockCalc, -4, true)
		
		// Test valid CRC type
		err := handler.ValidateCRCType(types.CRCInSection, types.CRCInITOCEntry, types.CRCInSection)
		assert.NoError(t, err)
		
		// Test invalid CRC type
		err = handler.ValidateCRCType(types.CRCNone, types.CRCInITOCEntry, types.CRCInSection)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported CRC type")
	})
}