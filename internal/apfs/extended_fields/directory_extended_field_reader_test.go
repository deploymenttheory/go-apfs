package extendedfields

import (
	"encoding/binary"
	"testing"

	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

type fakeExtendedField struct {
	xtype uint8
	data  []byte
}

func (f *fakeExtendedField) Type() uint8           { return f.xtype }
func (f *fakeExtendedField) Flags() uint8          { return 0 }
func (f *fakeExtendedField) Size() uint16          { return uint16(len(f.data)) }
func (f *fakeExtendedField) Data() []byte          { return f.data }
func (f *fakeExtendedField) IsDataDependent() bool { return false }
func (f *fakeExtendedField) ShouldCopy() bool      { return true }
func (f *fakeExtendedField) IsUserField() bool     { return false }
func (f *fakeExtendedField) IsSystemField() bool   { return false }

func encodeUint64LE(value uint64) []byte {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, value)
	return buf
}

func encodeUint32LE(v uint32) []byte {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, v)
	return buf
}

func TestDirectoryExtendedFieldReader_SiblingID(t *testing.T) {
	t.Run("returns value when field type 9 with 8 bytes", func(t *testing.T) {
		field := &fakeExtendedField{
			xtype: 9,
			data:  encodeUint64LE(0x123456789ABCDEF0),
		}
		reader := NewDirectoryExtendedFieldReader([]interfaces.ExtendedField{field})

		val, ok := reader.SiblingID()
		if !ok {
			t.Fatal("Expected true for SiblingID, got false")
		}
		if val != 0x123456789ABCDEF0 {
			t.Errorf("Expected 0x123456789ABCDEF0, got 0x%X", val)
		}
	})

	t.Run("returns false if data is too short", func(t *testing.T) {
		field := &fakeExtendedField{
			xtype: 9,
			data:  []byte{0x01, 0x02}, // too short
		}
		reader := NewDirectoryExtendedFieldReader([]interfaces.ExtendedField{field})

		_, ok := reader.SiblingID()
		if ok {
			t.Error("Expected false for short SiblingID data, got true")
		}
	})

	t.Run("returns false if field type does not match", func(t *testing.T) {
		field := &fakeExtendedField{
			xtype: 99, // wrong type
			data:  encodeUint64LE(0x123456789ABCDEF0),
		}
		reader := NewDirectoryExtendedFieldReader([]interfaces.ExtendedField{field})

		_, ok := reader.SiblingID()
		if ok {
			t.Error("Expected false for mismatched field type, got true")
		}
	})
}

func TestDirectoryExtendedFieldReader_FileSystemUUID(t *testing.T) {
	t.Run("returns UUID when field type 10 with 16 bytes", func(t *testing.T) {
		expected := types.UUID{0xA1, 0xB2, 0xC3, 0xD4, 0xE5, 0xF6, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A}
		field := &fakeExtendedField{
			xtype: 10,
			data:  expected[:],
		}
		reader := NewDirectoryExtendedFieldReader([]interfaces.ExtendedField{field})

		uuid, ok := reader.FileSystemUUID()
		if !ok {
			t.Fatal("Expected true for FileSystemUUID, got false")
		}
		if uuid != expected {
			t.Errorf("UUID mismatch.\nExpected: %v\nGot:      %v", expected, uuid)
		}
	})

	t.Run("returns false if data is not 16 bytes", func(t *testing.T) {
		field := &fakeExtendedField{
			xtype: 10,
			data:  make([]byte, 8), // too short
		}
		reader := NewDirectoryExtendedFieldReader([]interfaces.ExtendedField{field})

		_, ok := reader.FileSystemUUID()
		if ok {
			t.Error("Expected false for short UUID data, got true")
		}
	})

	t.Run("returns false if field type does not match", func(t *testing.T) {
		field := &fakeExtendedField{
			xtype: 11,
			data:  make([]byte, 16),
		}
		reader := NewDirectoryExtendedFieldReader([]interfaces.ExtendedField{field})

		_, ok := reader.FileSystemUUID()
		if ok {
			t.Error("Expected false for incorrect field type, got true")
		}
	})
}
