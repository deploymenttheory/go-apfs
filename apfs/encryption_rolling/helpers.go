package encryptionrolling

import (
	"github.com/deploymenttheory/go-apfs/apfs/types"
)

// State wraps types.ErStatePhysT to provide local method definitions.
type State struct {
	*types.ErStatePhysT
}

// BlockSize returns the block size code used for encryption checksums.
//
// This value is extracted from the ErsbFlags field using the
// ErsbFlagCmBlockSizeMask and ErsbFlagCmBlockSizeShift defined in the APFS spec.
//
// The returned value is an index into the following:
//
//	0 = 512 B
//	1 = 2 KiB
//	2 = 4 KiB
//	3 = 8 KiB
//	4 = 16 KiB
//	5 = 32 KiB
//	6 = 64 KiB
//
// Reference: APFS Reference, page 171.
func (s *State) BlockSize() uint32 {
	return uint32((s.ErsbFlags & types.ErsbFlagCmBlockSizeMask) >> types.ErsbFlagCmBlockSizeShift)
}

// Phase returns the current encryption rolling phase.
//
// This value is extracted from the ErsbFlags field using the
// ErsbFlagErPhaseMask and ErsbFlagErPhaseShift defined in the APFS spec.
//
// Valid phases include:
//
//	1 = OMAP rolling
//	2 = Data rolling
//	3 = Snapshot rolling
//
// Reference: APFS Reference, page 170.
func (s *State) Phase() uint32 {
	return uint32((s.ErsbFlags & types.ErsbFlagErPhaseMask) >> types.ErsbFlagErPhaseShift)
}
