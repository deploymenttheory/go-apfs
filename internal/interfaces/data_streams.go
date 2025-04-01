package interfaces

// PhysicalExtentReader provides methods for reading information about physical extents
type PhysicalExtentReader interface {
	// Length returns the length of the extent in blocks
	Length() uint64

	// Kind returns the kind of the extent (e.g., APFS_KIND_NEW)
	Kind() uint8

	// OwningObjectID returns the identifier of the file system record using this extent
	OwningObjectID() uint64

	// ReferenceCount returns the reference count for this extent
	ReferenceCount() int32

	// PhysicalBlockAddress returns the physical block address of the start of the extent
	PhysicalBlockAddress() uint64
}

// FileExtentReader provides methods for reading information about file extents
type FileExtentReader interface {
	// Length returns the length of the extent in bytes
	Length() uint64

	// Flags returns the file extent flags
	Flags() uint64

	// PhysicalBlockNumber returns the physical block address that the extent starts at
	PhysicalBlockNumber() uint64

	// CryptoID returns the encryption key or encryption tweak used in this extent
	CryptoID() uint64

	// LogicalAddress returns the offset within the file's data where the data is stored
	LogicalAddress() uint64

	// IsCryptoIDTweak checks if the crypto_id field contains an encryption tweak value
	IsCryptoIDTweak() bool
}

// DataStreamIDReader provides methods for reading information about data stream IDs
type DataStreamIDReader interface {
	// ReferenceCount returns the reference count for the data stream record
	ReferenceCount() uint32

	// ObjectID returns the object identifier for the data stream
	ObjectID() uint64
}

// DataStreamReader provides methods for reading information about data streams
type DataStreamReader interface {
	// Size returns the size of the data in bytes
	Size() uint64

	// AllocatedSize returns the total space allocated for the data stream
	AllocatedSize() uint64

	// DefaultCryptoID returns the default encryption key or tweak used in this data stream
	DefaultCryptoID() uint64

	// TotalBytesWritten returns the total bytes written to this data stream
	TotalBytesWritten() uint64

	// TotalBytesRead returns the total bytes read from this data stream
	TotalBytesRead() uint64
}

// ExtendedAttributeDataStreamReader provides methods for reading extended attribute data streams
type ExtendedAttributeDataStreamReader interface {
	// AttributeObjectID returns the identifier for the data stream
	AttributeObjectID() uint64

	// DataStream returns the data stream information
	DataStream() DataStreamReader
}

// DataStreamManager provides methods for managing data streams
type DataStreamManager interface {
	// GetDataStream retrieves the data stream for a given file or extended attribute
	GetDataStream(objectID uint64) (DataStreamReader, error)

	// ListExtents returns all extents for a given data stream
	ListExtents(dataStreamID uint64) ([]FileExtentReader, error)

	// GetTotalAllocatedSize returns the total allocated size for a given object
	GetTotalAllocatedSize(objectID uint64) (uint64, error)
}

// PhysicalExtentManager provides methods for managing physical extents
type PhysicalExtentManager interface {
	// FindExtent finds a physical extent at a given address
	FindExtent(blockAddress uint64) (PhysicalExtentReader, error)

	// ListExtentsByOwner returns all physical extents for a given owner
	ListExtentsByOwner(ownerID uint64) ([]PhysicalExtentReader, error)

	// GetExtentReferenceCount returns the reference count for a specific extent
	GetExtentReferenceCount(blockAddress uint64) (int32, error)
}

// FileExtentManager provides methods for managing file extents
type FileExtentManager interface {
	// FindExtent finds a file extent at a given logical address for a file
	FindExtent(fileID uint64, logicalAddress uint64) (FileExtentReader, error)

	// ListExtents returns all file extents for a given file
	ListExtents(fileID uint64) ([]FileExtentReader, error)

	// GetContiguousExtents returns contiguous file extents for a given logical address range
	GetContiguousExtents(fileID uint64, startAddress uint64, length uint64) ([]FileExtentReader, error)
}

// DataStreamAnalyzer provides methods for analyzing data streams
type DataStreamAnalyzer interface {
	// AnalyzeFragmentation analyzes the fragmentation level of a data stream
	AnalyzeFragmentation(dataStreamID uint64) (FragmentationAnalysis, error)

	// GetDataLayout returns the physical layout of data for a given stream
	GetDataLayout(dataStreamID uint64) (DataLayout, error)

	// VerifyDataIntegrity checks if all the extents for a data stream are valid
	VerifyDataIntegrity(dataStreamID uint64) error
}

// FragmentationAnalysis contains information about the fragmentation of a data stream
type FragmentationAnalysis struct {
	// The total number of extents in the data stream
	ExtentCount int

	// The average size of extents in blocks
	AverageExtentSize float64

	// The fragmentation level as a percentage (0-100)
	FragmentationLevel float64

	// The largest contiguous extent size in blocks
	LargestContiguousExtent uint64

	// The smallest extent size in blocks
	SmallestExtent uint64
}

// DataLayout provides information about the physical layout of data
type DataLayout struct {
	// The total size of the data in bytes
	TotalSize uint64

	// The total allocated size in bytes
	AllocatedSize uint64

	// Information about each extent in the data stream
	Extents []ExtentLayoutInfo
}

// ExtentLayoutInfo contains detailed information about an extent's layout
type ExtentLayoutInfo struct {
	// The logical address (file offset) in bytes
	LogicalAddress uint64

	// The physical block number where the extent starts
	PhysicalBlockNumber uint64

	// The length of the extent in bytes
	Length uint64

	// The encryption method or key used for this extent
	CryptoID uint64
}

// CryptoIDResolver provides methods for resolving crypto IDs
type CryptoIDResolver interface {
	// IsTweak checks if a crypto ID is a tweak value
	IsTweak(cryptoID uint64) bool

	// IsDefaultValue checks if a crypto ID is using the default value
	IsDefaultValue(cryptoID uint64, defaultCryptoID uint64) bool

	// ResolveDescription provides a description of what a crypto ID represents
	ResolveDescription(cryptoID uint64) string
}
