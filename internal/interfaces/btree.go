// File: internal/interfaces/btrees.go
package interfaces

import (
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// BTreeNodeReader provides methods for reading information from a B-tree node
type BTreeNodeReader interface {
	// Flags returns the B-tree node's flags
	Flags() uint16

	// Level returns the number of child levels below this node
	Level() uint16

	// KeyCount returns the number of keys stored in this node
	KeyCount() uint32

	// TableSpace returns the location of the table of contents
	TableSpace() types.NlocT

	// FreeSpace returns the location of the shared free space for keys and values
	FreeSpace() types.NlocT

	// KeyFreeList returns the linked list that tracks free key space
	KeyFreeList() types.NlocT

	// ValueFreeList returns the linked list that tracks free value space
	ValueFreeList() types.NlocT

	// Data returns the node's storage area
	Data() []byte

	// IsRoot checks if the node is a root node
	IsRoot() bool

	// IsLeaf checks if the node is a leaf node
	IsLeaf() bool

	// HasFixedKVSize checks if the node has keys and values of fixed size
	HasFixedKVSize() bool

	// IsHashed checks if the node contains child hashes
	IsHashed() bool

	// HasHeader checks if the node is stored with an object header
	HasHeader() bool
}

// BTreeInfoReader provides methods for reading information about a B-tree
type BTreeInfoReader interface {
	// Flags returns the B-tree's flags
	Flags() uint32

	// NodeSize returns the on-disk size in bytes of a node in this B-tree
	NodeSize() uint32

	// KeySize returns the size of a key, or zero if keys have variable size
	KeySize() uint32

	// ValueSize returns the size of a value, or zero if values have variable size
	ValueSize() uint32

	// LongestKey returns the length in bytes of the longest key ever stored in the B-tree
	LongestKey() uint32

	// LongestValue returns the length in bytes of the longest value ever stored in the B-tree
	LongestValue() uint32

	// KeyCount returns the number of keys stored in the B-tree
	KeyCount() uint64

	// NodeCount returns the number of nodes stored in the B-tree
	NodeCount() uint64

	// HasUint64Keys checks if the B-tree uses 64-bit unsigned integer keys
	HasUint64Keys() bool

	// SupportsSequentialInsert checks if the B-tree is optimized for sequential insertions
	SupportsSequentialInsert() bool

	// AllowsGhosts checks if the table of contents can contain keys with no corresponding value
	AllowsGhosts() bool

	// IsEphemeral checks if the nodes use ephemeral object identifiers
	IsEphemeral() bool

	// IsPhysical checks if the nodes use physical object identifiers
	IsPhysical() bool

	// IsPersistent checks if the B-tree is persisted across unmounting
	IsPersistent() bool

	// HasAlignedKV checks if keys and values are aligned to eight-byte boundaries
	HasAlignedKV() bool

	// IsHashed checks if nonleaf nodes store a hash of their child nodes
	IsHashed() bool

	// HasHeaderlessNodes checks if nodes are stored without object headers
	HasHeaderlessNodes() bool
}

// BTreeIndexNodeValueReader provides methods for reading hashed B-tree nonleaf node values
type BTreeIndexNodeValueReader interface {
	// ChildObjectID returns the object identifier of the child node
	ChildObjectID() types.OidT

	// ChildHash returns the hash of the child node
	ChildHash() [types.BtreeNodeHashSizeMax]byte
}

// BTreeLocationReader provides methods for reading locations within a B-tree node
type BTreeLocationReader interface {
	// Offset returns the offset in bytes
	Offset() uint16

	// Length returns the length in bytes
	Length() uint16

	// IsValid checks if the offset is valid
	IsValid() bool
}

// BTreeKVLocationReader provides methods for reading the location of a key and value
type BTreeKVLocationReader interface {
	// KeyLocation returns the location of the key
	KeyLocation() types.NlocT

	// ValueLocation returns the location of the value
	ValueLocation() types.NlocT
}

// BTreeKVOffsetReader provides methods for reading fixed-size key and value offsets
type BTreeKVOffsetReader interface {
	// KeyOffset returns the offset of the key
	KeyOffset() uint16

	// ValueOffset returns the offset of the value
	ValueOffset() uint16
}

// BTreeNavigator provides methods for navigating a B-tree
type BTreeNavigator interface {
	// GetRootNode returns the root node of the B-tree
	GetRootNode() (BTreeNodeReader, error)

	// GetChildNode returns a child node of the given parent node at the specified index
	GetChildNode(parent BTreeNodeReader, index int) (BTreeNodeReader, error)

	// GetNodeByObjectID returns a node with the specified object identifier
	GetNodeByObjectID(objectID types.OidT) (BTreeNodeReader, error)

	// GetHeight returns the height of the B-tree
	GetHeight() (uint16, error)
}

// BTreeSearcher provides methods for searching a B-tree
type BTreeSearcher interface {
	// Find looks for a key in the B-tree and returns its associated value
	Find(key []byte) ([]byte, error)

	// FindRange returns all key-value pairs within a given key range
	FindRange(startKey []byte, endKey []byte) ([]KeyValuePair, error)

	// ContainsKey checks if a key exists in the B-tree
	ContainsKey(key []byte) (bool, error)
}

// KeyValuePair represents a key-value pair in a B-tree
type KeyValuePair struct {
	// The key data
	Key []byte

	// The value data
	Value []byte
}

// BTreeTraverser provides methods for traversing a B-tree
type BTreeTraverser interface {
	// PreOrderTraversal performs a pre-order traversal of the B-tree
	PreOrderTraversal(visitor NodeVisitor) error

	// InOrderTraversal performs an in-order traversal of the B-tree
	InOrderTraversal(visitor NodeVisitor) error

	// PostOrderTraversal performs a post-order traversal of the B-tree
	PostOrderTraversal(visitor NodeVisitor) error

	// LevelOrderTraversal performs a level-order traversal of the B-tree
	LevelOrderTraversal(visitor NodeVisitor) error
}

// NodeVisitor defines a function to be called for each node during traversal
type NodeVisitor func(node BTreeNodeReader, depth int) (bool, error)

// BTreeAnalyzer provides methods for analyzing a B-tree
type BTreeAnalyzer interface {
	// GetNodeDistribution returns information about the distribution of nodes at each level
	GetNodeDistribution() ([]LevelInfo, error)

	// CalculateFillFactor returns the average fill factor of the B-tree
	CalculateFillFactor() (float64, error)

	// CalculateHeight returns the height of the B-tree
	CalculateHeight() (int, error)

	// AnalyzeStructure performs a comprehensive analysis of the B-tree structure
	AnalyzeStructure() (BTreeAnalysis, error)
}

// LevelInfo provides information about B-tree nodes at a specific level
type LevelInfo struct {
	// The level number (0 for leaf nodes)
	Level int

	// The number of nodes at this level
	NodeCount int

	// The average number of keys per node at this level
	AverageKeyCount float64

	// The minimum number of keys in any node at this level
	MinKeyCount int

	// The maximum number of keys in any node at this level
	MaxKeyCount int
}

// BTreeAnalysis provides comprehensive analysis of a B-tree
type BTreeAnalysis struct {
	// The height of the tree
	Height int

	// The total number of nodes
	TotalNodes int

	// The total number of keys
	TotalKeys int

	// The average fill factor as a percentage
	FillFactor float64

	// Information about each level in the tree
	Levels []LevelInfo

	// True if the tree is balanced
	IsBalanced bool

	// Information about the largest and smallest node in the tree
	LargestNode BTreeNodeInfo

	// Information about the largest and smallest node in the tree
	SmallestNode BTreeNodeInfo
}

// BTreeNodeInfo provides detailed information about a specific B-tree node
type BTreeNodeInfo struct {
	// The object identifier of the node
	ObjectID types.OidT

	// The level of the node
	Level uint16

	// The number of keys in the node
	KeyCount uint32

	// The flags for the node
	Flags uint16

	// The size of the node in bytes
	SizeInBytes uint32
}
