package types

// B-Trees (pages 122-134)
// The B-trees used in Apple File System are implemented using the btree_node_phys_t structure to represent a node.
// The same structure is used for all nodes in a tree.

// BtreeNodePhysT is a B-tree node.
// Reference: page 123
type BtreeNodePhysT struct {
	// The object's header. (page 124)
	BtnO ObjPhysT

	// The B-tree node's flags. (page 124)
	// For the values used in this bit field, see B-Tree Node Flags.
	BtnFlags uint16

	// The number of child levels below this node. (page 124)
	// For example, the value of this field is zero for a leaf node and one for the immediate parent of a leaf node.
	// Likewise, the height of a tree is one plus the value of this field on the tree's root node.
	BtnLevel uint16

	// The number of keys stored in this node. (page 124)
	BtnNkeys uint32

	// The location of the table of contents. (page 124)
	// The offset for the table of contents is counted from the beginning of the node's btn_data field
	// to the beginning of the table of contents.
	// If the BTNODE_FIXED_KV_SIZE flag is set, the table of contents is an array of instances of kvoff_t;
	// otherwise, it's an array of instances of kvloc_t.
	BtnTableSpace NlocT

	// The location of the shared free space for keys and values. (page 124)
	// The location's offset is counted from the beginning of the key area to the beginning of the free space.
	BtnFreeSpace NlocT

	// A linked list that tracks free key space. (page 125)
	// The offset from the beginning of the key area to the first available space for a key is stored in the off field,
	// and the total amount of free key space is stored in the len field.
	// Each free space stores an instance of nloc_t whose len field indicates the size of that free space
	// and whose off field contains the location of the next free space.
	BtnKeyFreeList NlocT

	// A linked list that tracks free value space. (page 125)
	// The offset from the end of the value area to the first available space for a value is stored in the off field,
	// and the total amount of free value space is stored in the len field.
	// Each free space stores an instance of nloc_t whose len field indicates the size of that free space
	// and whose off field contains the location of the next free space.
	BtnValFreeList NlocT

	// The node's storage area. (page 125)
	// This area contains the table of contents, keys, free space, and values.
	// A root node also has as an instance of btree_info_t at the end of its storage area.
	BtnData []byte
}

// BtreeInfoFixedT contains static information about a B-tree.
// Reference: page 125
type BtreeInfoFixedT struct {
	// The B-tree's flags. (page 125)
	// For the values used in this bit field, see B-Tree Flags.
	BtFlags uint32

	// The on-disk size, in bytes, of a node in this B-tree. (page 126)
	// Leaf nodes, nonleaf nodes, and the root node are all the same size.
	BtNodeSize uint32

	// The size of a key, or zero if the keys have variable size. (page 126)
	// If this field has a value of zero, the btn_flags field of instances of btree_node_phys_t
	// in this tree must not include BTNODE_FIXED_KV_SIZE.
	BtKeySize uint32

	// The size of a value, or zero if the values have variable size. (page 126)
	// If this field has a value of zero, the btn_flags field of instances of btree_node_phys_t
	// for leaf nodes in this tree must not include BTNODE_FIXED_KV_SIZE.
	// Nonleaf nodes in a tree with variable-size values include BTNODE_FIXED_KV_SIZE,
	// because the values stored in those nodes are the object identifiers of their child nodes,
	// and object identifiers have a fixed size.
	BtValSize uint32
}

// BtreeInfoT contains information about a B-tree.
// Reference: page 126
type BtreeInfoT struct {
	// Information about the B-tree that doesn't change over time. (page 126)
	BtFixed BtreeInfoFixedT

	// The length, in bytes, of the longest key that has ever been stored in the B-tree. (page 126)
	BtLongestKey uint32

	// The length, in bytes, of the longest value that has ever been stored in the B-tree. (page 126)
	BtLongestVal uint32

	// The number of keys stored in the B-tree. (page 127)
	BtKeyCount uint64

	// The number of nodes stored in the B-tree. (page 127)
	BtNodeCount uint64
}

// BtnIndexNodeValT is the value used by hashed B-trees for nonleaf nodes.
// Reference: page 127
type BtnIndexNodeValT struct {
	// The object identifier of the child node. (page 127)
	BinvChildOid OidT

	// The hash of the child node. (page 127)
	// The hash algorithm used by this tree determines the length of the hash.
	// To compute the hash, use the entire child node object as the input for the hash algorithm
	// specified for this tree. If the output from that hash algorithm is smaller than the
	// BTREE_NODE_HASH_SIZE_MAX bytes, treat the remaining bytes as padding.
	BinvChildHash [BtreeNodeHashSizeMax]byte
}

// BtreeNodeHashSizeMax is the maximum length of a hash that can be stored in this structure.
// Reference: page 128
// This value is the same as APFS_HASH_MAX_SIZE.
const BtreeNodeHashSizeMax = 64

// NlocT is a location within a B-tree node.
// Reference: page 128
type NlocT struct {
	// The offset, in bytes. (page 128)
	// Depending on the data type that contains this location, the offset is either
	// implicitly positive or negative, and is counted starting at different points in the B-tree node.
	Off uint16

	// The length, in bytes. (page 128)
	Len uint16
}

// BtoffInvalid is an invalid offset.
// Reference: page 128
// This value is stored in the off field of nloc_t to indicate that there's no offset.
// For example, the last entry in a free list has no entry after it, so it uses this value for its off field.
const BtoffInvalid uint16 = 0xffff

// KvlocT is the location, within a B-tree node, of a key and value.
// Reference: page 128
type KvlocT struct {
	// The location of the key. (page 129)
	K NlocT

	// The location of the value. (page 129)
	V NlocT
}

// KvoffT is the location, within a B-tree node, of a fixed-size key and value.
// Reference: page 129
type KvoffT struct {
	// The offset of the key. (page 129)
	K uint16

	// The offset of the value. (page 129)
	V uint16
}

// B-Tree Flags (pages 129-131)

// BtreeUint64Keys indicates code that works with the B-tree should enable optimizations
// to make comparison of keys fast.
// Reference: page 130
const BtreeUint64Keys uint32 = 0x00000001

// BtreeSequentialInsert indicates code that works with the B-tree should enable optimizations
// to keep the B-tree compact during sequential insertion of entries.
// Reference: page 130
const BtreeSequentialInsert uint32 = 0x00000002

// BtreeAllowGhosts indicates the table of contents is allowed to contain keys that have no corresponding value.
// Reference: page 130
const BtreeAllowGhosts uint32 = 0x00000004

// BtreeEphemeral indicates the nodes in the B-tree use ephemeral object identifiers to link to child nodes.
// Reference: page 130
const BtreeEphemeral uint32 = 0x00000008

// BtreePhysical indicates the nodes in the B-tree use physical object identifiers to link to child nodes.
// Reference: page 130
const BtreePhysical uint32 = 0x00000010

// BtreeNonpersistent indicates the B-tree isn't persisted across unmounting.
// Reference: page 131
const BtreeNonpersistent uint32 = 0x00000020

// BtreeKvNonaligned indicates the keys and values in the B-tree aren't required to be
// aligned to eight-byte boundaries.
// Reference: page 131
const BtreeKvNonaligned uint32 = 0x00000040

// BtreeHashed indicates the nonleaf nodes of this B-tree store a hash of their child nodes.
// Reference: page 131
const BtreeHashed uint32 = 0x00000080

// BtreeNoheader indicates the nodes of this B-tree are stored without object headers.
// Reference: page 131
const BtreeNoheader uint32 = 0x00000100

// B-Tree Table of Contents Constants (page 131)

// BtreeTocEntryIncrement is the number of entries that are added or removed
// when changing the size of the table of contents.
// Reference: page 131
const BtreeTocEntryIncrement uint32 = 8

// BtreeTocEntryMaxUnused is the maximum allowed number of unused entries in the table of contents.
// Reference: page 131
const BtreeTocEntryMaxUnused uint32 = 2 * BtreeTocEntryIncrement

// B-Tree Node Flags (pages 132-133)

// BtnodeRoot indicates the B-tree node is a root node.
// Reference: page 132
const BtnodeRoot uint16 = 0x0001

// BtnodeLeaf indicates the B-tree node is a leaf node.
// Reference: page 132
const BtnodeLeaf uint16 = 0x0002

// BtnodeFixedKvSize indicates the B-tree node has keys and values of a fixed size,
// and the table of contents omits their lengths.
// Reference: page 132
const BtnodeFixedKvSize uint16 = 0x0004

// BtnodeHashed indicates the B-tree node contains child hashes.
// Reference: page 132
const BtnodeHashed uint16 = 0x0008

// BtnodeNoheader indicates the B-tree node is stored without an object header.
// Reference: page 133
const BtnodeNoheader uint16 = 0x0010

// BtnodeCheckKoffInval indicates the B-tree node is in a transient state.
// Reference: page 133
const BtnodeCheckKoffInval uint16 = 0x8000

// B-Tree Node Constants (page 133)

// BtreeNodeSizeDefault is the default size, in bytes, of a B-tree node.
// Reference: page 133
const BtreeNodeSizeDefault uint32 = 4096

// BtreeNodeMinEntryCount is the minimum number of entries that must be able to fit
// in a nonleaf B-tree node.
// Reference: page 133
const BtreeNodeMinEntryCount uint32 = 4
