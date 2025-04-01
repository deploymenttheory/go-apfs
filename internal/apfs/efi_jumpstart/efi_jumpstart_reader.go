// File: efi_jumpstart_reader.go
package efijumpstart

import (
	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

type EFIJumpstart struct {
	data types.NxEfiJumpstartT
}

func NewEFIJumpstartReader(data types.NxEfiJumpstartT) interfaces.EFIJumpstartReader {
	return &EFIJumpstart{data: data}
}

func (e *EFIJumpstart) Magic() uint32 {
	return e.data.NejMagic
}

func (e *EFIJumpstart) Version() uint32 {
	return e.data.NejVersion
}

func (e *EFIJumpstart) EFIFileLength() uint32 {
	return e.data.NejEfiFileLen
}

func (e *EFIJumpstart) ExtentCount() uint32 {
	return e.data.NejNumExtents
}

func (e *EFIJumpstart) Extents() []types.Prange {
	return e.data.NejRecExtents
}

func (e *EFIJumpstart) IsValid() bool {
	return e.Magic() == types.NxEfiJumpstartMagic && e.Version() == types.NxEfiJumpstartVersion
}
