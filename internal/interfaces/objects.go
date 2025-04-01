package interfaces

import (
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// ObjectIdentifier provides methods for working with object identifiers
type ObjectIdentifier interface {
	// ID returns the object's unique identifier
	ID() types.OidT

	// TransactionID returns the transaction identifier of the most recent modification
	TransactionID() types.XidT

	// IsValid checks if the object identifier is valid
	IsValid() bool
}

// ObjectTypeInspector provides methods for inspecting object types and characteristics
type ObjectTypeInspector interface {
	// Type returns the base object type
	Type() uint32

	// Subtype returns the object's subtype
	Subtype() uint32

	// TypeName returns a human-readable name for the object type
	TypeName() string

	// IsVirtual checks if the object is a virtual object
	IsVirtual() bool

	// IsEphemeral checks if the object is an ephemeral object
	IsEphemeral() bool

	// IsPhysical checks if the object is a physical object
	IsPhysical() bool

	// IsEncrypted checks if the object is encrypted
	IsEncrypted() bool

	// IsNonpersistent checks if the object is non-persistent
	IsNonpersistent() bool

	// HasHeader checks if the object has a standard header
	HasHeader() bool
}

// ObjectChecksumVerifier provides methods for verifying object integrity
type ObjectChecksumVerifier interface {
	// Checksum returns the object's Fletcher 64 checksum
	Checksum() [types.MaxCksumSize]byte

	// VerifyChecksum checks the integrity of the object's checksum
	VerifyChecksum() bool
}

// ObjectTypeResolver provides methods for resolving object type details
type ObjectTypeResolver interface {
	// ResolveType converts a raw object type to a human-readable description
	ResolveType(objectType uint32) string

	// SupportedObjectTypes returns a list of all supported object types
	SupportedObjectTypes() []uint32

	// GetObjectTypeCategory categorizes the object type (e.g., metadata, file system, container)
	GetObjectTypeCategory(objectType uint32) string
}

// ObjectRegistry provides a comprehensive registry of APFS object types
type ObjectRegistry interface {
	// LookupType provides detailed information about a specific object type
	LookupType(objectType uint32) (ObjectTypeInfo, bool)

	// ListObjectTypes returns all known object types with their descriptions
	ListObjectTypes() []ObjectTypeInfo
}

// ObjectTypeInfo contains detailed information about an APFS object type
type ObjectTypeInfo struct {
	// Numeric type identifier
	Type uint32

	// Human-readable name
	Name string

	// Detailed description
	Description string

	// Categorization (e.g., metadata, file system, container)
	Category string
}

// ObjectStorageTypeResolver provides methods for resolving object storage characteristics
type ObjectStorageTypeResolver interface {
	// DetermineStorageType resolves the storage type (virtual, ephemeral, physical)
	DetermineStorageType(objectType uint32) string

	// IsStorageTypeSupported checks if a specific storage type is supported
	IsStorageTypeSupported(storageType string) bool
}
