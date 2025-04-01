// File: internal/efijumpstart/efi_jumpstart_reader.go
package efijumpstart

import (
	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// EFIJumpstartReader implements the EFIJumpstartReader interface.
// It holds the parsed EFI jumpstart data structure.
type EFIJumpstartReader struct {
	data types.NxEfiJumpstartT
}

// Compile-time check to ensure EFIJumpstartReader implements EFIJumpstartReader
var _ interfaces.EFIJumpstartReader = (*EFIJumpstartReader)(nil)

// NewEFIJumpstartReader creates a new reader instance from the parsed jumpstart data.
func NewEFIJumpstartReader(data types.NxEfiJumpstartT) interfaces.EFIJumpstartReader {
	return &EFIJumpstartReader{data: data}
}

// Magic returns the magic number for validating the EFI jumpstart structure.
func (r *EFIJumpstartReader) Magic() uint32 {
	return r.data.NejMagic
}

// Version returns the version number of the EFI jumpstart structure.
func (r *EFIJumpstartReader) Version() uint32 {
	return r.data.NejVersion
}

// EFIFileLength returns the size in bytes of the embedded EFI driver.
func (r *EFIJumpstartReader) EFIFileLength() uint32 {
	return r.data.NejEfiFileLen
}

// ExtentCount returns the number of extents where the EFI driver is stored.
func (r *EFIJumpstartReader) ExtentCount() uint32 {
	return r.data.NejNumExtents
}

// Extents returns a copy of the locations where the EFI driver is stored.
// Returning a copy prevents external modification of the internal slice.
func (r *EFIJumpstartReader) Extents() []types.Prange {
	if r.data.NejRecExtents == nil {
		return nil
	}
	extentsCopy := make([]types.Prange, len(r.data.NejRecExtents))
	copy(extentsCopy, r.data.NejRecExtents)
	return extentsCopy
}

// IsValid checks if the EFI jumpstart structure is valid based on magic number and version.
func (r *EFIJumpstartReader) IsValid() bool {
	return r.Magic() == types.NxEfiJumpstartMagic && r.Version() == types.NxEfiJumpstartVersion
}
