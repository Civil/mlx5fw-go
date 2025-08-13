package crc

import (
	"testing"

	"github.com/Civil/mlx5fw-go/pkg/parser"
	"github.com/Civil/mlx5fw-go/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestCRCStrategies(t *testing.T) {
	testData := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}

	t.Run("SoftwareCRCStrategy", func(t *testing.T) {
		strategy := &SoftwareCRCStrategy{}
		calc := parser.NewCRCCalculator()

		// Calculate CRC
		result := strategy.Calculate(calc, testData)
		assert.NotZero(t, result)

		// Check type
		assert.Equal(t, types.CRCInITOCEntry, strategy.GetType())
	})

	t.Run("HardwareCRCStrategy", func(t *testing.T) {
		strategy := &HardwareCRCStrategy{}
		calc := parser.NewCRCCalculator()

		// Calculate CRC
		result := strategy.Calculate(calc, testData)
		assert.NotZero(t, result)

		// Check type
		assert.Equal(t, types.CRCInSection, strategy.GetType())
	})
}

func TestGenericCRCCalculator(t *testing.T) {
	testData := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}

	t.Run("WithSoftwareStrategy", func(t *testing.T) {
		calc := NewGenericCRCCalculator(&SoftwareCRCStrategy{})

		// Test Calculate
		result := calc.Calculate(testData)
		assert.NotZero(t, result)

		// Test GetType
		assert.Equal(t, types.CRCInITOCEntry, calc.GetType())

		// Test CalculateImageCRC
		imageResult := calc.CalculateImageCRC(testData, 2)
		assert.NotZero(t, imageResult)

		// Test CalculateWithParams (should use default calculation)
		paramsResult := calc.CalculateWithParams(testData, 0, 0, 0)
		assert.Equal(t, result, paramsResult)
	})

	t.Run("WithHardwareStrategy", func(t *testing.T) {
		calc := NewGenericCRCCalculator(&HardwareCRCStrategy{})

		// Test Calculate
		result := calc.Calculate(testData)
		assert.NotZero(t, result)

		// Test GetType
		assert.Equal(t, types.CRCInSection, calc.GetType())
	})
}

func TestFactoryFunctions(t *testing.T) {
	t.Run("NewSoftwareCRCCalculator", func(t *testing.T) {
		calc := NewSoftwareCRCCalculator()
		assert.NotNil(t, calc)
		assert.Equal(t, types.CRCInITOCEntry, calc.GetType())
	})

	t.Run("NewHardwareCRCCalculator", func(t *testing.T) {
		calc := NewHardwareCRCCalculator()
		assert.NotNil(t, calc)
		assert.Equal(t, types.CRCInSection, calc.GetType())
	})
}

func TestDefaultCRCHandler(t *testing.T) {
	handler := NewDefaultCRCHandler()
	assert.NotNil(t, handler)

	testData := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}

	t.Run("CalculateCRC", func(t *testing.T) {
		// Test CRCInITOCEntry
		result, err := handler.CalculateCRC(testData, types.CRCInITOCEntry)
		assert.NoError(t, err)
		assert.NotZero(t, result)

		// Test CRCInSection (should exclude last 4 bytes)
		result, err = handler.CalculateCRC(testData, types.CRCInSection)
		assert.NoError(t, err)
		assert.NotZero(t, result)

		// Test CRCNone
		result, err = handler.CalculateCRC(testData, types.CRCNone)
		assert.NoError(t, err)
		assert.Zero(t, result)

		// Test with data too short for CRCInSection
		shortData := []byte{0x01, 0x02}
		result, err = handler.CalculateCRC(shortData, types.CRCInSection)
		assert.NoError(t, err)
		assert.Zero(t, result)
	})

	t.Run("VerifyCRC", func(t *testing.T) {
		// Calculate expected CRC
		expectedCRC, _ := handler.CalculateCRC(testData, types.CRCInITOCEntry)

		// Test successful verification
		err := handler.VerifyCRC(testData, expectedCRC, types.CRCInITOCEntry)
		assert.NoError(t, err)

		// Test failed verification
		err = handler.VerifyCRC(testData, 0x1234, types.CRCInITOCEntry)
		assert.Error(t, err)
	})

	t.Run("GetCRCOffset", func(t *testing.T) {
		assert.Equal(t, -4, handler.GetCRCOffset())
	})

	t.Run("HasEmbeddedCRC", func(t *testing.T) {
		assert.False(t, handler.HasEmbeddedCRC())
	})
}
