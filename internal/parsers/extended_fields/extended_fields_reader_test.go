package extendedfields

import (
	"encoding/binary"
	"testing"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

func encodeFieldHeader(xType, xFlags uint8, xSize uint16) []byte {
	header := make([]byte, 4)
	header[0] = xType
	header[1] = xFlags
	binary.LittleEndian.PutUint16(header[2:], xSize)
	return header
}

func TestExtendedFieldsReader(t *testing.T) {
	t.Run("Empty blob returns no fields", func(t *testing.T) {
		blob := types.XfBlobT{
			XfNumExts:  0,
			XfUsedData: 0,
			XfData:     nil,
		}
		reader := NewExtendedFieldsReader(blob)

		if got := reader.NumberOfExtendedFields(); got != 0 {
			t.Errorf("Expected 0 fields, got %d", got)
		}
		if got := reader.TotalUsedDataSize(); got != 0 {
			t.Errorf("Expected 0 used bytes, got %d", got)
		}

		fields, err := reader.ListExtendedFields()
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if len(fields) != 0 {
			t.Errorf("Expected 0 fields, got %d", len(fields))
		}
	})

	t.Run("Single field is decoded correctly", func(t *testing.T) {
		payload := []byte{0xCA, 0xFE, 0xBA, 0xBE}
		header := encodeFieldHeader(1, types.XfUserField, uint16(len(payload)))
		data := append(header, payload...)

		blob := types.XfBlobT{
			XfNumExts:  1,
			XfUsedData: uint16(len(data)),
			XfData:     data,
		}
		reader := NewExtendedFieldsReader(blob)

		fields, err := reader.ListExtendedFields()
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if len(fields) != 1 {
			t.Fatalf("Expected 1 field, got %d", len(fields))
		}

		field := fields[0]
		if field.Type() != 1 {
			t.Errorf("Expected Type=1, got %d", field.Type())
		}
		if !field.IsUserField() {
			t.Error("Expected IsUserField() = true")
		}
		if field.Size() != uint16(len(payload)) {
			t.Errorf("Expected size %d, got %d", len(payload), field.Size())
		}
		if got := field.Data(); string(got) != string(payload) {
			t.Errorf("Data mismatch. Got %v, want %v", got, payload)
		}
	})

	t.Run("Multiple fields are decoded in order", func(t *testing.T) {
		field1 := []byte{0xAA, 0xBB}
		field2 := []byte{0x11, 0x22, 0x33}
		h1 := encodeFieldHeader(5, types.XfSystemField, uint16(len(field1)))
		h2 := encodeFieldHeader(6, types.XfUserField, uint16(len(field2)))

		// Build data with proper 8-byte alignment between fields
		data := append(h1, field1...)
		// Field 1 size: 4 (header) + 2 (data) = 6 bytes
		// Align to 8 bytes by adding 2 padding bytes
		data = append(data, 0x00, 0x00)
		data = append(data, h2...)
		data = append(data, field2...)

		blob := types.XfBlobT{
			XfNumExts:  2,
			XfUsedData: uint16(len(data)),
			XfData:     data,
		}
		reader := NewExtendedFieldsReader(blob)

		fields, err := reader.ListExtendedFields()
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if len(fields) != 2 {
			t.Fatalf("Expected 2 fields, got %d", len(fields))
		}

		if fields[0].Type() != 5 || !fields[0].IsSystemField() {
			t.Error("First field incorrect")
		}
		if fields[1].Type() != 6 || !fields[1].IsUserField() {
			t.Error("Second field incorrect")
		}
	})

	t.Run("Fails on insufficient data for header", func(t *testing.T) {
		blob := types.XfBlobT{
			XfNumExts:  1,
			XfUsedData: 3,
			XfData:     []byte{1, 2, 3}, // too short for header (needs 4 bytes)
		}
		reader := NewExtendedFieldsReader(blob)
		_, err := reader.ListExtendedFields()
		if err == nil {
			t.Error("Expected error due to short header, got nil")
		}
	})

	t.Run("Fails on insufficient data for payload", func(t *testing.T) {
		header := encodeFieldHeader(9, 0, 10)         // claims 10-byte payload
		data := append(header, []byte{0x01, 0x02}...) // only 2 bytes
		blob := types.XfBlobT{
			XfNumExts:  1,
			XfUsedData: uint16(len(data)),
			XfData:     data,
		}
		reader := NewExtendedFieldsReader(blob)
		_, err := reader.ListExtendedFields()
		if err == nil {
			t.Error("Expected error due to short payload, got nil")
		}
	})
}
