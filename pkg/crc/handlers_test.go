package crc

import (
	"testing"

	"github.com/Civil/mlx5fw-go/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestNoCRCHandler(t *testing.T) {
	handler := NewNoCRCHandler()
	assert.NotNil(t, handler)
	
	testData := []byte{0x01, 0x02, 0x03, 0x04}
	
	t.Run("CalculateCRC", func(t *testing.T) {
		// Should always return 0
		result, err := handler.CalculateCRC(testData, types.CRCNone)
		assert.NoError(t, err)
		assert.Zero(t, result)
		
		// Should return 0 for any CRC type
		result, err = handler.CalculateCRC(testData, types.CRCInSection)
		assert.NoError(t, err)
		assert.Zero(t, result)
	})
	
	t.Run("VerifyCRC", func(t *testing.T) {
		// Should always succeed regardless of expected CRC
		err := handler.VerifyCRC(testData, 0x1234, types.CRCNone)
		assert.NoError(t, err)
		
		err = handler.VerifyCRC(testData, 0, types.CRCInSection)
		assert.NoError(t, err)
	})
	
	t.Run("GetCRCOffset", func(t *testing.T) {
		assert.Equal(t, -1, handler.GetCRCOffset())
	})
	
	t.Run("HasEmbeddedCRC", func(t *testing.T) {
		assert.False(t, handler.HasEmbeddedCRC())
	})
}