// File: internal/interfaces/extended_fields.go
package interfaces

import (
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// ExtendedFieldsReader provides methods for reading extended fields
type ExtendedFieldsReader interface {
	// NumberOfExtendedFields returns the total number of extended fields
	NumberOfExtendedFields() uint16

	// TotalUsedDataSize returns the total bytes used by extended fields
	TotalUsedDataSize() uint16

	// ListExtendedFields returns all extended fields
	ListExtendedFields() ([]ExtendedField, error)
}

// ExtendedField represents a single extended field
type ExtendedField interface {
	// Type returns the extended field's data type
	Type() uint8

	// Flags returns the extended field's flags
	Flags() uint8

	// Size returns the size of the data in bytes
	Size() uint16

	// Data returns the raw data of the extended field
	Data() []byte

	// IsDataDependent checks if the field depends on file data
	IsDataDependent() bool

	// ShouldCopy checks if the field should be copied
	ShouldCopy() bool

	// IsUserField checks if the field was added by a user-space program
	IsUserField() bool

	// IsSystemField checks if the field was added by the kernel or APFS implementation
	IsSystemField() bool
}

// ExtendedFieldTypeResolver provides methods for resolving extended field types
type ExtendedFieldTypeResolver interface {
	// ResolveName returns a human-readable name for an extended field type
	ResolveName(fieldType uint8) string

	// ResolveDescription provides a detailed description of an extended field type
	ResolveDescription(fieldType uint8) string

	// ListSupportedFieldTypes returns all supported extended field types
	ListSupportedFieldTypes() []uint8
}

// InodeExtendedFieldReader provides specialized methods for reading inode-specific extended fields
type InodeExtendedFieldReader interface {
	// SnapshotTransactionID returns the snapshot transaction identifier
	SnapshotTransactionID() (uint64, bool)

	// DeltaTreeOID returns the virtual object identifier for snapshot extent delta list
	DeltaTreeOID() (types.OidT, bool)

	// DocumentID returns the document identifier
	DocumentID() (uint32, bool)

	// PreviousFileSize returns the file's previous size (used for crash recovery)
	PreviousFileSize() (uint64, bool)

	// FinderInfo returns the Finder-specific opaque data
	FinderInfo() ([]byte, bool)

	// SparseByteCount returns the number of sparse bytes in the data stream
	SparseByteCount() (uint64, bool)

	// DeviceIdentifier returns the device identifier for special files
	DeviceIdentifier() (uint32, bool)

	// OriginalSyncRootID returns the original sync-root hierarchy inode number
	OriginalSyncRootID() (uint64, bool)
}

// DirectoryExtendedFieldReader provides specialized methods for reading directory-specific extended fields
type DirectoryExtendedFieldReader interface {
	// SiblingID returns the sibling identifier for hard links
	SiblingID() (uint64, bool)

	// FileSystemUUID returns the UUID of the automatically mounted file system
	FileSystemUUID() (types.UUID, bool)
}
