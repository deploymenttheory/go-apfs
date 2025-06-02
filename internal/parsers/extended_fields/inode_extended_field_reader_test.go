package extendedfields

import (
	"bytes"
	"testing"

	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

func TestInodeExtendedFieldReader(t *testing.T) {
	tests := []struct {
		name      string
		xtype     uint8
		data      []byte
		call      func(reader *InodeExtendedFieldReader) (any, bool)
		expectVal any
		expectOK  bool
	}{
		{
			name:      "SnapshotTransactionID present",
			xtype:     5,
			data:      encodeUint64LE(0xAABBCCDDEEFF0011),
			call:      func(r *InodeExtendedFieldReader) (any, bool) { return r.SnapshotTransactionID() },
			expectVal: uint64(0xAABBCCDDEEFF0011),
			expectOK:  true,
		},
		{
			name:      "DeltaTreeOID present",
			xtype:     6,
			data:      encodeUint64LE(0xCAFEBABE12345678),
			call:      func(r *InodeExtendedFieldReader) (any, bool) { return r.DeltaTreeOID() },
			expectVal: types.OidT(0xCAFEBABE12345678),
			expectOK:  true,
		},
		{
			name:      "DocumentID present",
			xtype:     1,
			data:      encodeUint32LE(0x11223344),
			call:      func(r *InodeExtendedFieldReader) (any, bool) { return r.DocumentID() },
			expectVal: uint32(0x11223344),
			expectOK:  true,
		},
		{
			name:      "PreviousFileSize present",
			xtype:     4,
			data:      encodeUint64LE(0x5566778899AABBCC),
			call:      func(r *InodeExtendedFieldReader) (any, bool) { return r.PreviousFileSize() },
			expectVal: uint64(0x5566778899AABBCC),
			expectOK:  true,
		},
		{
			name:      "SparseByteCount present",
			xtype:     3,
			data:      encodeUint64LE(0x1234000012340000),
			call:      func(r *InodeExtendedFieldReader) (any, bool) { return r.SparseByteCount() },
			expectVal: uint64(0x1234000012340000),
			expectOK:  true,
		},
		{
			name:      "DeviceIdentifier present",
			xtype:     7,
			data:      encodeUint32LE(0xFEEDBEEF),
			call:      func(r *InodeExtendedFieldReader) (any, bool) { return r.DeviceIdentifier() },
			expectVal: uint32(0xFEEDBEEF),
			expectOK:  true,
		},
		{
			name:      "OriginalSyncRootID present",
			xtype:     8,
			data:      encodeUint64LE(0xABABABABABABABAB),
			call:      func(r *InodeExtendedFieldReader) (any, bool) { return r.OriginalSyncRootID() },
			expectVal: uint64(0xABABABABABABABAB),
			expectOK:  true,
		},
		{
			name:      "FinderInfo present",
			xtype:     2,
			data:      []byte{0xDE, 0xAD, 0xBE, 0xEF},
			call:      func(r *InodeExtendedFieldReader) (any, bool) { return r.FinderInfo() },
			expectVal: []byte{0xDE, 0xAD, 0xBE, 0xEF},
			expectOK:  true,
		},
		{
			name:      "SnapshotTransactionID missing",
			xtype:     1,
			data:      encodeUint32LE(0xDEADBEEF),
			call:      func(r *InodeExtendedFieldReader) (any, bool) { return r.SnapshotTransactionID() },
			expectVal: uint64(0),
			expectOK:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := &fakeExtendedField{xtype: tt.xtype, data: tt.data}
			reader := NewInodeExtendedFieldReader([]interfaces.ExtendedField{field})

			gotVal, ok := tt.call(reader)
			if ok != tt.expectOK {
				t.Errorf("Expected ok=%v, got %v", tt.expectOK, ok)
				return
			}

			if ok {
				switch expected := tt.expectVal.(type) {
				case []byte:
					gotBytes, ok := gotVal.([]byte)
					if !ok || !bytes.Equal(gotBytes, expected) {
						t.Errorf("Expected byte slice %x, got %x", expected, gotVal)
					}
				default:
					if gotVal != expected {
						t.Errorf("Expected value %v, got %v", expected, gotVal)
					}
				}
			}
		})
	}
}
