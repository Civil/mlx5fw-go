package crc

import (
	"errors"
	"testing"

	pkgerrors "github.com/Civil/mlx5fw-go/pkg/errors"
	"github.com/Civil/mlx5fw-go/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestHandlerStrategies(t *testing.T) {
	mockCalc := &MockCRCCalculator{result: 0x1234}
	
	t.Run("SoftwareCRC16Strategy", func(t *testing.T) {
		strategy := &SoftwareCRC16Strategy{}
		
		// Test CalculateCRC
		result, err := strategy.CalculateCRC(mockCalc, []byte{1, 2, 3, 4}, types.CRCInITOCEntry)
		assert.NoError(t, err)
		assert.Equal(t, uint32(0x1234), result)
		
		// Test ValidCRCTypes
		validTypes := strategy.ValidCRCTypes()
		assert.Contains(t, validTypes, types.CRCInITOCEntry)
		assert.Contains(t, validTypes, types.CRCInSection)
	})
	
	t.Run("HardwareCRC16Strategy", func(t *testing.T) {
		strategy := &HardwareCRC16Strategy{}
		
		// Test CalculateCRC
		result, err := strategy.CalculateCRC(mockCalc, []byte{1, 2, 3, 4}, types.CRCInSection)
		assert.NoError(t, err)
		assert.Equal(t, uint32(0x1234), result)
		
		// Test ValidCRCTypes
		validTypes := strategy.ValidCRCTypes()
		assert.Contains(t, validTypes, types.CRCInSection)
		assert.Len(t, validTypes, 1)
	})
	
	t.Run("InSectionCRC16Strategy", func(t *testing.T) {
		strategy := &InSectionCRC16Strategy{}
		
		// Test with sufficient data
		data := []byte{1, 2, 3, 4, 5, 6, 7, 8}
		result, err := strategy.CalculateCRC(mockCalc, data, types.CRCInSection)
		assert.NoError(t, err)
		assert.Equal(t, uint32(0x1234), result)
		
		// Test with insufficient data
		shortData := []byte{1, 2, 3}
		_, err = strategy.CalculateCRC(mockCalc, shortData, types.CRCInSection)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, pkgerrors.ErrDataTooShort))
		
		// Test ValidCRCTypes
		validTypes := strategy.ValidCRCTypes()
		assert.Contains(t, validTypes, types.CRCInSection)
		assert.Len(t, validTypes, 1)
	})
}

func TestUnifiedCRCHandler(t *testing.T) {
	mockCalc := &MockCRCCalculator{result: 0x1234}
	
	t.Run("NewUnifiedCRCHandler", func(t *testing.T) {
		strategy := &SoftwareCRC16Strategy{}
		handler := NewUnifiedCRCHandler(mockCalc, strategy, -4, true)
		
		assert.NotNil(t, handler)
		assert.Equal(t, -4, handler.GetCRCOffset())
		assert.True(t, handler.HasEmbeddedCRC())
	})
	
	t.Run("CalculateCRC", func(t *testing.T) {
		strategy := &SoftwareCRC16Strategy{}
		handler := NewUnifiedCRCHandler(mockCalc, strategy, -4, false)
		
		// Test valid CRC type
		result, err := handler.CalculateCRC([]byte{1, 2, 3, 4}, types.CRCInITOCEntry)
		assert.NoError(t, err)
		assert.Equal(t, uint32(0x1234), result)
		
		// Test invalid CRC type
		_, err = handler.CalculateCRC([]byte{1, 2, 3, 4}, types.CRCNone)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported CRC type")
	})
	
	t.Run("VerifyCRC", func(t *testing.T) {
		// Test with non-embedded CRC (uses VerifyCRC)
		strategy := &SoftwareCRC16Strategy{}
		handler := NewUnifiedCRCHandler(mockCalc, strategy, -1, false)
		
		err := handler.VerifyCRC([]byte{1, 2, 3, 4}, 0x1234, types.CRCInITOCEntry)
		assert.NoError(t, err)
		
		// Test with embedded CRC (uses VerifyCRC16)
		handler = NewUnifiedCRCHandler(mockCalc, strategy, -4, true)
		
		err = handler.VerifyCRC([]byte{1, 2, 3, 4}, 0xFFFF1234, types.CRCInSection)
		assert.NoError(t, err)
		
		// Test failed verification
		err = handler.VerifyCRC([]byte{1, 2, 3, 4}, 0x5678, types.CRCInSection)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, pkgerrors.ErrCRCMismatch))
	})
}

func TestFactoryFunctionsForHandlers(t *testing.T) {
	mockCalc := &MockCRCCalculator{result: 0x1234}
	
	t.Run("NewSoftwareCRC16Handler", func(t *testing.T) {
		handler := NewSoftwareCRC16Handler(mockCalc)
		assert.NotNil(t, handler)
		assert.Equal(t, -1, handler.GetCRCOffset())
		assert.False(t, handler.HasEmbeddedCRC())
	})
	
	t.Run("NewHardwareCRC16Handler", func(t *testing.T) {
		handler := NewHardwareCRC16Handler(mockCalc)
		assert.NotNil(t, handler)
		assert.Equal(t, -4, handler.GetCRCOffset())
		assert.True(t, handler.HasEmbeddedCRC())
	})
	
	t.Run("NewInSectionCRC16Handler", func(t *testing.T) {
		handler := NewInSectionCRC16Handler(mockCalc)
		assert.NotNil(t, handler)
		assert.Equal(t, -4, handler.GetCRCOffset())
		assert.True(t, handler.HasEmbeddedCRC())
	})
}