package btrees

import (
	"fmt"

	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// BTreeValidator provides validation for B-tree node structures
type BTreeValidator struct{}

// NewBTreeValidator creates a new B-tree validator
func NewBTreeValidator() *BTreeValidator {
	return &BTreeValidator{}
}

// ValidationResult contains the result of validation
type ValidationResult struct {
	Valid    bool
	Errors   []string
	Warnings []string
}

// ValidateNode performs comprehensive validation on a B-tree node
func (btv *BTreeValidator) ValidateNode(node interfaces.BTreeNodeReader) *ValidationResult {
	result := &ValidationResult{
		Valid:    true,
		Errors:   []string{},
		Warnings: []string{},
	}

	// Perform all validation checks
	btv.checkKeyCount(node, result)
	btv.checkTableOfContents(node, result)
	btv.checkEntryBounds(node, result)
	btv.checkFreeLists(node, result)
	btv.checkNodeConsistency(node, result)
	btv.checkFooterIfRoot(node, result)

	// If any errors were found, mark as invalid
	if len(result.Errors) > 0 {
		result.Valid = false
	}

	return result
}

// checkKeyCount validates the key count is within reasonable bounds
func (btv *BTreeValidator) checkKeyCount(node interfaces.BTreeNodeReader, result *ValidationResult) {
	keyCount := node.KeyCount()

	// Check for zero keys in root or internal nodes
	if keyCount == 0 && node.IsRoot() {
		result.Errors = append(result.Errors, "root node has zero keys")
		return
	}

	// Check for excessively large key count (likely corruption)
	// APFS default node size is 4096 bytes. Minimum entry size is 4 bytes (kvoff_t)
	// So theoretical maximum is around 1000 entries, but realistically much less
	if keyCount > 10000 {
		result.Errors = append(result.Errors, fmt.Sprintf("excessive key count: %d (likely corruption)", keyCount))
		return
	}

	// Warn if suspiciously high but not impossible
	if keyCount > 1000 {
		result.Warnings = append(result.Warnings, fmt.Sprintf("unusually high key count: %d", keyCount))
	}
}

// checkTableOfContents validates the table of contents location and size
func (btv *BTreeValidator) checkTableOfContents(node interfaces.BTreeNodeReader, result *ValidationResult) {
	tableSpace := node.TableSpace()
	nodeData := node.Data()

	// Check table offset is within node data
	if int(tableSpace.Off)+int(tableSpace.Len) > len(nodeData) {
		result.Errors = append(result.Errors, fmt.Sprintf(
			"table of contents out of bounds: offset=%d size=%d available=%d",
			tableSpace.Off, tableSpace.Len, len(nodeData)))
		return
	}

	// Calculate expected TOC size based on entry count and key size
	keyCount := node.KeyCount()
	var expectedTocSize uint16

	if node.HasFixedKVSize() {
		// kvoff_t: 4 bytes per entry
		expectedTocSize = uint16(keyCount) * 4
	} else {
		// kvloc_t: 8 bytes per entry
		expectedTocSize = uint16(keyCount) * 8
	}

	if tableSpace.Len != expectedTocSize {
		result.Warnings = append(result.Warnings, fmt.Sprintf(
			"table of contents size mismatch: expected=%d actual=%d",
			expectedTocSize, tableSpace.Len))
	}
}

// checkEntryBounds validates individual entry offsets
func (btv *BTreeValidator) checkEntryBounds(node interfaces.BTreeNodeReader, result *ValidationResult) {
	nodeData := node.Data()
	if len(nodeData) == 0 {
		return
	}

	tableSpace := node.TableSpace()
	tableOffset := int(tableSpace.Off)
	keyCount := node.KeyCount()

	// Determine entry size
	var entrySize int
	if node.HasFixedKVSize() {
		entrySize = 4 // kvoff_t
	} else {
		entrySize = 8 // kvloc_t
	}

	// Check each entry's offsets
	errorCount := 0
	for i := uint32(0); i < keyCount && errorCount < 5; i++ {
		entryOffset := tableOffset + int(i)*entrySize

		// Check entry is within TOC
		if entryOffset+entrySize > len(nodeData) {
			result.Errors = append(result.Errors, fmt.Sprintf(
				"entry %d offset out of bounds: offset=%d size=%d available=%d",
				i, entryOffset, entrySize, len(nodeData)))
			errorCount++
			continue
		}

		// Extract key and value offsets
		if node.HasFixedKVSize() {
			keyOffset := btv.readUint16LE(nodeData, entryOffset)
			valueOffset := btv.readUint16LE(nodeData, entryOffset+2)

			// Check key offset is reasonable (must be within 4096 byte node)
			if keyOffset > 4096 {
				result.Errors = append(result.Errors, fmt.Sprintf(
					"entry %d: invalid key offset %d (exceeds node size)",
					i, keyOffset))
				errorCount++
			}

			// Check value offset is reasonable
			if valueOffset > 4096 {
				result.Errors = append(result.Errors, fmt.Sprintf(
					"entry %d: invalid value offset %d (exceeds node size)",
					i, valueOffset))
				errorCount++
			}

			// For internal nodes, value should be 8 bytes (child OID)
			if !node.IsLeaf() && valueOffset > 0 {
				childOidOffset := 56 + int(valueOffset)
				if childOidOffset+8 > len(nodeData) {
					result.Warnings = append(result.Warnings, fmt.Sprintf(
						"entry %d: child OID extends beyond node data",
						i))
				}
			}
		} else {
			// Variable-size entry
			keyOffset := btv.readUint16LE(nodeData, entryOffset)
			keySize := btv.readUint16LE(nodeData, entryOffset+2)
			valueOffset := btv.readUint16LE(nodeData, entryOffset+4)
			valueSize := btv.readUint16LE(nodeData, entryOffset+6)

			// Check key bounds
			if keyOffset+keySize > 4096 {
				result.Errors = append(result.Errors, fmt.Sprintf(
					"entry %d: key bounds exceeded (offset=%d size=%d)",
					i, keyOffset, keySize))
				errorCount++
			}

			// Check value bounds
			if valueOffset+valueSize > 4096 {
				result.Errors = append(result.Errors, fmt.Sprintf(
					"entry %d: value bounds exceeded (offset=%d size=%d)",
					i, valueOffset, valueSize))
				errorCount++
			}
		}
	}

	if errorCount >= 5 {
		result.Errors = append(result.Errors, fmt.Sprintf(
			"... and %d more entry boundary errors",
			int(keyCount)-errorCount))
	}
}

// checkFreeLists validates free list pointers
func (btv *BTreeValidator) checkFreeLists(node interfaces.BTreeNodeReader, result *ValidationResult) {
	nodeData := node.Data()

	// Check key free list
	keyFreeList := node.KeyFreeList()
	if keyFreeList.Off != types.BtoffInvalid && keyFreeList.Off < 0xf000 {
		if int(keyFreeList.Off) >= len(nodeData) {
			result.Warnings = append(result.Warnings, fmt.Sprintf(
				"key free list offset out of bounds: %d",
				keyFreeList.Off))
		}
	}

	// Check value free list
	valFreeList := node.ValueFreeList()
	if valFreeList.Off != types.BtoffInvalid && valFreeList.Off < 0xf000 {
		if int(valFreeList.Off) >= len(nodeData) {
			result.Warnings = append(result.Warnings, fmt.Sprintf(
				"value free list offset out of bounds: %d",
				valFreeList.Off))
		}
	}

	// Check free space
	freeSpace := node.FreeSpace()
	if freeSpace.Off > 0 && int(freeSpace.Off) >= len(nodeData) {
		result.Warnings = append(result.Warnings, fmt.Sprintf(
			"free space offset out of bounds: %d",
			freeSpace.Off))
	}
}

// checkNodeConsistency checks internal consistency of node structure
func (btv *BTreeValidator) checkNodeConsistency(node interfaces.BTreeNodeReader, result *ValidationResult) {
	// Check level consistency with flags
	if node.IsLeaf() && node.Level() != 0 {
		result.Warnings = append(result.Warnings, fmt.Sprintf(
			"leaf node has non-zero level: %d",
			node.Level()))
	}

	if !node.IsLeaf() && node.Level() == 0 {
		result.Warnings = append(result.Warnings,
			"internal node has level 0 (should be > 0)")
	}

	// Check flags consistency
	flags := node.Flags()

	// Both FIXED_KV_SIZE and variable can't be true at same time
	hasFixedKV := (flags & types.BtnodeFixedKvSize) != 0

	// For now, we don't have all flags defined, so just check the ones we know
	// Root, Leaf, and FixedKvSize should be valid combinations

	// If both Root and Leaf are set, this is a single-node tree (valid but unusual)
	if node.IsRoot() && node.IsLeaf() && node.KeyCount() == 0 {
		result.Warnings = append(result.Warnings,
			"empty root-leaf node (single-node tree with no keys)")
	}

	// Check that fixed KV nodes have reasonable key count
	if hasFixedKV && node.KeyCount() > 256 && node.IsLeaf() {
		// 256+ keys in a leaf with fixed-size entries (minimum 4 bytes TOC + 16 bytes keys/values)
		// would need (256 * 4) + (256 * 16) = 5KB just for keys/values, exceeding typical node size
		result.Warnings = append(result.Warnings, fmt.Sprintf(
			"unusually large fixed-KV leaf node: %d keys",
			node.KeyCount()))
	}
}

// checkFooterIfRoot validates B-tree footer in root nodes
func (btv *BTreeValidator) checkFooterIfRoot(node interfaces.BTreeNodeReader, result *ValidationResult) {
	// Root nodes should have a B-tree footer at the end of the node (last 56 bytes)
	// This is difficult to validate without the full node data including header,
	// but we can check that the node has sufficient size
	if node.IsRoot() && len(node.Data()) < 56 {
		result.Warnings = append(result.Warnings,
			"root node data too small to contain footer")
	}
}

// readUint16LE reads a little-endian uint16 from a byte slice
func (btv *BTreeValidator) readUint16LE(data []byte, offset int) uint16 {
	if offset+2 > len(data) {
		return 0
	}
	return uint16(data[offset]) | (uint16(data[offset+1]) << 8)
}

// IsValid checks if a validation result indicates a valid node
func (vr *ValidationResult) IsValid() bool {
	return vr.Valid && len(vr.Errors) == 0
}

// ErrorString returns a formatted string of all errors
func (vr *ValidationResult) ErrorString() string {
	if len(vr.Errors) == 0 {
		return ""
	}
	result := "Validation errors:\n"
	for _, err := range vr.Errors {
		result += fmt.Sprintf("  - %s\n", err)
	}
	return result
}

// WarningString returns a formatted string of all warnings
func (vr *ValidationResult) WarningString() string {
	if len(vr.Warnings) == 0 {
		return ""
	}
	result := "Validation warnings:\n"
	for _, warn := range vr.Warnings {
		result += fmt.Sprintf("  - %s\n", warn)
	}
	return result
}
