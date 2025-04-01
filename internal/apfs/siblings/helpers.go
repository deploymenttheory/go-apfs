package siblings

import (
	"strings"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

// Sibling wraps types.JSiblingValT to provide helper methods.
type Sibling struct {
	*types.JSiblingValT
}

// NameString returns the sibling name as a Go string.
//
// It trims any null terminator (`\x00`) commonly present in null-terminated UTF-8 strings.
//
// Reference: APFS Reference, page 116.
func (s *Sibling) NameString() string {
	return strings.TrimRight(string(s.Name), "\x00")
}

// HasName returns true if the sibling's name (after trimming) matches the given string.
//
// Comparison is case-sensitive and assumes UTF-8 encoding.
func (s *Sibling) HasName(name string) bool {
	return s.NameString() == name
}

// ParentID returns the identifier of the parent inode.
//
// Reference: APFS Reference, page 116.
func (s *Sibling) ParentID() uint64 {
	return s.ParentId
}

// NameLength returns the stored name length, including the null terminator.
//
// This is the raw `NameLen` field and not the length of the string returned by NameString().
func (s *Sibling) NameLength() uint16 {
	return s.NameLen
}
