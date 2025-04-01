package extendedfields

import (
	"testing"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

func TestExtendedField(t *testing.T) {
	header := types.XFieldT{
		XType:  42,
		XFlags: types.XfUserField | types.XfSystemField | types.XfDataDependent, // multiple flags
		XSize:  5,
	}
	data := []byte{0xDE, 0xAD, 0xBE, 0xEF, 0x00}

	field := NewExtendedField(header, data)

	t.Run("Type returns correct field type", func(t *testing.T) {
		if field.Type() != 42 {
			t.Errorf("Expected Type() to return 42, got %d", field.Type())
		}
	})

	t.Run("Flags returns correct flags byte", func(t *testing.T) {
		if field.Flags() != header.XFlags {
			t.Errorf("Expected Flags() = %08b, got %08b", header.XFlags, field.Flags())
		}
	})

	t.Run("Size returns header size", func(t *testing.T) {
		if field.Size() != 5 {
			t.Errorf("Expected Size() = 5, got %d", field.Size())
		}
	})

	t.Run("Data returns exact data slice", func(t *testing.T) {
		got := field.Data()
		if len(got) != len(data) {
			t.Errorf("Expected data length %d, got %d", len(data), len(got))
		}
		for i := range got {
			if got[i] != data[i] {
				t.Errorf("Data mismatch at index %d: got 0x%X, want 0x%X", i, got[i], data[i])
			}
		}
	})

	t.Run("IsDataDependent returns true when XfDataDependent is set", func(t *testing.T) {
		if !field.IsDataDependent() {
			t.Error("Expected IsDataDependent() = true")
		}
	})

	t.Run("ShouldCopy returns true when XfDoNotCopy is not set", func(t *testing.T) {
		if !field.ShouldCopy() {
			t.Error("Expected ShouldCopy() = true")
		}
	})

	t.Run("IsUserField returns true when XfUserField is set", func(t *testing.T) {
		if !field.IsUserField() {
			t.Error("Expected IsUserField() = true")
		}
	})

	t.Run("IsSystemField returns true when XfSystemField is set", func(t *testing.T) {
		if !field.IsSystemField() {
			t.Error("Expected IsSystemField() = true")
		}
	})

	t.Run("ShouldCopy returns false if XfDoNotCopy is set", func(t *testing.T) {
		headerWithDoNotCopy := types.XFieldT{
			XFlags: types.XfDoNotCopy,
		}
		copyField := NewExtendedField(headerWithDoNotCopy, nil)
		if copyField.ShouldCopy() {
			t.Error("Expected ShouldCopy() = false")
		}
	})
}
