package efijumpstart

import (
	"unsafe"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

// Jumpstart wraps types.NxEfiJumpstartT to provide helper methods.
type Jumpstart struct {
	*types.NxEfiJumpstartT
}

// TotalSize returns the total size in bytes of the NxEfiJumpstartT structure,
// including the variable-length array of Prange extent entries.
//
// This calculation includes:
//   - ObjPhysT: 8 bytes (assumed fixed, double-check if yours is larger)
//   - NejMagic: 4 bytes
//   - NejVersion: 4 bytes
//   - NejEfiFileLen: 4 bytes
//   - NejNumExtents: 4 bytes
//   - NejReserved: 16 * 8 bytes = 128 bytes
//   - NejRecExtents: num_extents * sizeof(Prange)
//
// Note: This is only valid if the structure was read from disk or constructed
// as a contiguous block. In Go memory, sizes may vary based on alignment.
//
// Reference: APFS Reference, page 24–25.
func (j *Jumpstart) TotalSize() int {
	return 8 + 4 + 4 + 4 + 4 + (16 * 8) + int(j.NejNumExtents)*int(unsafe.Sizeof(types.Prange{}))
}
