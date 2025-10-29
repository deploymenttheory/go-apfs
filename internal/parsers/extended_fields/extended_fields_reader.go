package extendedfields

import (
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

type ExtendedFieldsReader struct {
	blob types.XfBlobT
}

func NewExtendedFieldsReader(blob types.XfBlobT) *ExtendedFieldsReader {
	return &ExtendedFieldsReader{blob: blob}
}

func (r *ExtendedFieldsReader) NumberOfExtendedFields() uint16 {
	return r.blob.XfNumExts
}

func (r *ExtendedFieldsReader) TotalUsedDataSize() uint16 {
	return r.blob.XfUsedData
}

func (r *ExtendedFieldsReader) ListExtendedFields() ([]interfaces.ExtendedField, error) {
	if r.blob.XfNumExts == 0 || r.blob.XfUsedData == 0 {
		return nil, nil
	}

	offset := 0
	fields := make([]interfaces.ExtendedField, 0, r.blob.XfNumExts)

	for i := 0; i < int(r.blob.XfNumExts); i++ {
		if offset+4 > len(r.blob.XfData) {
			return nil, fmt.Errorf("insufficient data for field header at index %d", i)
		}

		// Decode header
		hdr := types.XFieldT{
			XType:  r.blob.XfData[offset],
			XFlags: r.blob.XfData[offset+1],
			XSize:  binary.LittleEndian.Uint16(r.blob.XfData[offset+2 : offset+4]),
		}
		offset += 4

		// Validate that data exists
		if offset+int(hdr.XSize) > len(r.blob.XfData) {
			return nil, errors.New("insufficient data for extended field payload")
		}

		data := r.blob.XfData[offset : offset+int(hdr.XSize)]
		offset += int(hdr.XSize)

		// Extended fields must be 8-byte aligned per APFS specification
		// Round up to next 8-byte boundary
		offset = (offset + 7) & ^7

		fields = append(fields, &ExtendedField{header: hdr, data: data})
	}

	return fields, nil
}
