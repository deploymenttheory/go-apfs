// File: internal/interfaces/object_map.go
package interfaces

import (
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// ObjectMapReader provides methods for reading object map information
type ObjectMapReader interface {
	// Flags returns the object map's flags
	Flags() uint32

	// SnapshotCount returns the number of snapshots in this object map
	SnapshotCount() uint32

	// TreeType returns the type of tree used for object mappings
	TreeType() uint32

	// SnapshotTreeType returns the type of tree used for snapshots
	SnapshotTreeType() uint32

	// TreeOID returns the virtual object identifier of the object mapping tree
	TreeOID() types.OidT

	// SnapshotTreeOID returns the virtual object identifier of the snapshot tree
	SnapshotTreeOID() types.OidT

	// MostRecentSnapshotXID returns the transaction ID of the most recent snapshot
	MostRecentSnapshotXID() types.XidT
}

// ObjectMapTransactionManager provides methods for managing object map transactions
type ObjectMapTransactionManager interface {
	// PendingRevertMinXID returns the smallest transaction ID for an in-progress revert
	PendingRevertMinXID() types.XidT

	// PendingRevertMaxXID returns the largest transaction ID for an in-progress revert
	PendingRevertMaxXID() types.XidT

	// IsRevertInProgress checks if a revert operation is currently happening
	IsRevertInProgress() bool
}

// ObjectMapEntryReader provides methods for reading individual object map entries
type ObjectMapEntryReader interface {
	// ObjectID returns the object's identifier
	ObjectID() types.OidT

	// TransactionID returns the transaction identifier
	TransactionID() types.XidT

	// Flags returns the entry's flags
	Flags() uint32

	// Size returns the size of the object
	Size() uint32

	// PhysicalAddress returns the physical address of the object
	PhysicalAddress() types.Paddr

	// IsDeleted checks if the object is marked as deleted
	IsDeleted() bool

	// IsEncrypted checks if the object is encrypted
	IsEncrypted() bool

	// HasHeader checks if the object has a physical header
	HasHeader() bool
}

// ObjectMapSnapshotManager provides methods for managing object map snapshots
type ObjectMapSnapshotManager interface {
	// ListSnapshots returns all snapshots in the object map
	ListSnapshots() ([]ObjectMapSnapshotInfo, error)

	// FindSnapshotByXID finds a snapshot by its transaction ID
	FindSnapshotByXID(xid types.XidT) (ObjectMapSnapshotInfo, error)
}

// ObjectMapSnapshotInfo represents information about an object map snapshot
type ObjectMapSnapshotInfo struct {
	// Transaction ID of the snapshot
	XID types.XidT

	// Flags associated with the snapshot
	Flags uint32

	// Object identifier associated with the snapshot
	OID types.OidT
}

// ObjectMapInspector provides methods for comprehensive object map inspection
type ObjectMapInspector interface {
	// ListObjects returns all objects in the object map
	ListObjects() ([]ObjectMapEntryReader, error)

	// FindObjectByID finds an object by its identifier and optional transaction ID
	FindObjectByID(objectID types.OidT, txID ...types.XidT) (ObjectMapEntryReader, error)

	// CountObjects returns the total number of objects in the object map
	CountObjects() (int, error)

	// FindDeletedObjects returns all deleted objects
	FindDeletedObjects() ([]ObjectMapEntryReader, error)

	// FindEncryptedObjects returns all encrypted objects
	FindEncryptedObjects() ([]ObjectMapEntryReader, error)
}
