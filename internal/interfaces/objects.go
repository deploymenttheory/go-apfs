// File: internal/interfaces/objects.go
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

// ObjectDataReader provides methods for reading object data
type ObjectDataReader interface {
	// ReadObjectData reads raw data from an object
	ReadObjectData(id types.OidT, txID types.XidT) ([]byte, error)
}

// ObjectDataWriter provides methods for modifying object data
type ObjectDataWriter interface {
	// WriteObjectData writes raw data to an object
	WriteObjectData(id types.OidT, txID types.XidT, data []byte) error
}

// ObjectAddressResolver provides methods for working with object physical addresses
type ObjectAddressResolver interface {
	// PhysicalAddress returns the physical address where the object is stored
	PhysicalAddress(id types.OidT) (types.Paddr, error)

	// AddressToObjectID converts a physical address to an object identifier if possible
	AddressToObjectID(addr types.Paddr) (types.OidT, error)
}

// ObjectTierClassifier provides methods for determining storage tier placement
type ObjectTierClassifier interface {
	// GetStorageTier determines which storage tier an object is located on
	GetStorageTier(id types.OidT) (string, error)

	// IsOnMainTier checks if an object is on the main (SSD) tier
	IsOnMainTier(id types.OidT) (bool, error)

	// IsOnSecondaryTier checks if an object is on the secondary (HDD) tier
	IsOnSecondaryTier(id types.OidT) (bool, error)
}

// ObjectRelocationManager provides methods for moving objects between tiers
type ObjectRelocationManager interface {
	// PrepareRelocation prepares an object to be moved between tiers
	PrepareRelocation(id types.OidT, targetTier string) error

	// IsBeingRelocated checks if an object is in the process of being relocated
	IsBeingRelocated(id types.OidT) (bool, error)

	// ConfirmRelocation marks an object relocation as complete
	ConfirmRelocation(id types.OidT) error
}

// ObjectExtentReader provides methods for reading object data from extents
type ObjectExtentReader interface {
	// ReadExtent reads data from a specified extent
	ReadExtent(prange types.Prange) ([]byte, error)

	// ValidateExtent verifies if an extent is valid and accessible
	ValidateExtent(prange types.Prange) (bool, error)
}

// ObjectMagicVerifier provides methods for verifying object magic numbers
type ObjectMagicVerifier interface {
	// VerifyMagic checks if the object has the expected magic number
	VerifyMagic(expectedMagic uint32) bool

	// GetMagic returns the object's magic number
	GetMagic() uint32
}

// ObjectVersionVerifier provides methods for verifying object versions
type ObjectVersionVerifier interface {
	// VerifyVersion checks if the object has a supported version number
	VerifyVersion(expectedVersion uint32) bool

	// GetVersion returns the object's version number
	GetVersion() uint32
}

// ObjectLifecycleManager provides methods for managing object lifecycle
type ObjectLifecycleManager interface {
	// MarkForDeletion flags an object for deletion
	MarkForDeletion(id types.OidT, xid types.XidT) error

	// IsMarkedForDeletion checks if an object is marked for deletion
	IsMarkedForDeletion(id types.OidT) (bool, error)

	// CompletePhysicalDeletion performs the actual removal of an object
	CompletePhysicalDeletion(id types.OidT) error
}

// ObjectReferenceCounted provides methods for tracking object references
type ObjectReferenceCounted interface {
	// ReferenceCount returns the number of references to this object
	ReferenceCount() (uint32, error)

	// IncrementReferenceCount increases the reference count
	IncrementReferenceCount() error

	// DecrementReferenceCount decreases the reference count
	DecrementReferenceCount() error
}

// ObjectStateAccessor provides methods for accessing object state
type ObjectStateAccessor interface {
	// GetState retrieves state information for an object
	GetState(id types.OidT) ([]byte, error)

	// SetState updates state information for an object
	SetState(id types.OidT, state []byte) error
}

// ObjectTransactionManager provides methods for transaction-based operations
type ObjectTransactionManager interface {
	// BeginTransaction starts a new transaction
	BeginTransaction() (types.XidT, error)

	// CommitTransaction commits a transaction
	CommitTransaction(xid types.XidT) error

	// RollbackTransaction rolls back a transaction
	RollbackTransaction(xid types.XidT) error

	// GetObjectAtTransaction gets an object as it existed at a specific transaction
	GetObjectAtTransaction(id types.OidT, xid types.XidT) (any, error)
}

// ObjectListManager provides methods for managing ordered collections of objects
type ObjectListManager interface {
	// CreateList creates a new list of objects
	CreateList() (types.OidT, error)

	// AddToList adds an object to a list
	AddToList(listID types.OidT, objectID types.OidT) error

	// RemoveFromList removes an object from a list
	RemoveFromList(listID types.OidT, objectID types.OidT) error

	// GetListHead gets the first object in a list
	GetListHead(listID types.OidT) (types.OidT, error)

	// GetListTail gets the last object in a list
	GetListTail(listID types.OidT) (types.OidT, error)

	// GetNextInList gets the next object in a list
	GetNextInList(listID types.OidT, currentID types.OidT) (types.OidT, error)
}

// ObjectDependencyManager provides methods for handling object relationships
type ObjectDependencyManager interface {
	// GetDependencies returns objects that depend on the specified object
	GetDependencies(id types.OidT) ([]types.OidT, error)

	// GetDependents returns objects that the specified object depends on
	GetDependents(id types.OidT) ([]types.OidT, error)

	// HasDependencies checks if an object has any dependencies
	HasDependencies(id types.OidT) (bool, error)

	// ResolveDependencies ensures all dependencies are properly resolved
	ResolveDependencies(id types.OidT) error
}

// ObjectPhaseTracker provides methods for tracking multi-phase operations
type ObjectPhaseTracker interface {
	// GetCurrentPhase gets the current phase of an operation
	GetCurrentPhase(operationID uint64) (uint32, error)

	// SetCurrentPhase updates the current phase of an operation
	SetCurrentPhase(operationID uint64, phase uint32) error

	// IsPhaseComplete checks if a phase is complete
	IsPhaseComplete(operationID uint64, phase uint32) (bool, error)

	// GetPhaseProgress gets the progress of the current phase
	GetPhaseProgress(operationID uint64) (float64, error)
}

// ObjectBlockAllocator provides methods for allocating and freeing blocks for objects
type ObjectBlockAllocator interface {
	// AllocateBlocks allocates a specific number of blocks for an object
	AllocateBlocks(id types.OidT, count uint32) (types.Prange, error)

	// FreeBlocks releases allocated blocks back to the free pool
	FreeBlocks(id types.OidT, blockRange types.Prange) error

	// GetAllocatedBlocks retrieves the blocks allocated to an object
	GetAllocatedBlocks(id types.OidT) ([]types.Prange, error)
}

// ObjectSizeCalculator provides methods for calculating object size requirements
type ObjectSizeCalculator interface {
	// CalculateRequiredBlocks determines how many blocks an object needs
	CalculateRequiredBlocks(objectType uint32, dataSize uint64) (uint32, error)

	// GetObjectSize returns the size of an object in bytes
	GetObjectSize(id types.OidT) (uint64, error)
}

// ObjectStorageStatistics provides methods for gathering storage statistics
type ObjectStorageStatistics interface {
	// GetTotalBlocksAllocated returns the total number of blocks allocated
	GetTotalBlocksAllocated() (uint64, error)

	// GetTotalBlocksFreed returns the total number of blocks freed
	GetTotalBlocksFreed() (uint64, error)

	// GetObjectAllocationCount returns the number of blocks allocated to an object
	GetObjectAllocationCount(id types.OidT) (uint32, error)
}

type ObjectLocator interface {
	// Locate an object by its object identifier
	LocateObject(objectID types.OidT) ([]byte, error)

	// Check if an object exists
	ObjectExists(objectID types.OidT) bool
}

type ObjectHeaderReader interface {
	// Read the header of an object
	ReadObjectHeader(objectID types.OidT) (types.ObjPhysT, error)

	// Verify object header integrity
	VerifyObjectHeader(objectID types.OidT) bool
}

type ObjectTreeNavigator interface {
	// Get child nodes for a given node
	GetChildNodes(nodeID types.OidT) ([]types.OidT, error)

	// Get parent node for a given node
	GetParentNode(nodeID types.OidT) (types.OidT, error)

	// Determine tree height
	GetTreeHeight(rootNodeID types.OidT) (int, error)
}
