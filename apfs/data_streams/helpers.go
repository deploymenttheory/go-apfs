package datastreams

import (
	"github.com/deploymenttheory/go-apfs/apfs/types"
)

// PhysExtent wraps types.JPhysExtValT to provide local helper methods.
type PhysExtent struct {
	*types.JPhysExtValT
}

// ExtentLength returns the length of the physical extent in blocks.
//
// This value is extracted from the lower 60 bits of the LenAndKind field,
// using the PextLenMask defined in the APFS specification.
//
// Reference: APFS Reference, page 103.
func (v *PhysExtent) ExtentLength() uint64 {
	return v.LenAndKind & types.PextLenMask
}

// ExtentKind returns the kind of the physical extent.
//
// This value is extracted from the high 4 bits of the LenAndKind field,
// using the PextKindMask and PextKindShift constants defined in the APFS specification.
//
// The kind determines the classification of the extent, such as NEW, OLD, etc.
//
// Reference: APFS Reference, page 103.
func (v *PhysExtent) ExtentKind() uint64 {
	return (v.LenAndKind & types.PextKindMask) >> types.PextKindShift
}
