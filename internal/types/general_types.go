// Package apfs implements data structures for the Apple File System.
// This package is based on the official Apple File System Reference (June 2020).
package types

// General-Purpose Types (page 9)
// Basic types that are used in a variety of contexts, and aren't associated with
// any particular functionality.

// Paddr represents a physical address of an on-disk block.
// Negative numbers aren't valid addresses.
// This value is modeled as a signed integer to match IOKit.
// Reference: page 9
type Paddr int64

// Validate checks if the physical address is valid.
func (p Paddr) Validate() bool {
	return p >= 0
}

// Prange represents a range of physical addresses.
// Reference: page 9
type Prange struct {
	// The first block in the range. (page 9)
	PrStartPaddr Paddr
	// The number of blocks in the range. (page 9)
	PrBlockCount uint64
}

// UUID represents a universally unique identifier.
// Reference: page 9
type UUID [16]byte
