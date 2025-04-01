package types

// Data Streams (pages 102-107)
// Short pieces of information like a file's name are stored inside the data structures that contain metadata.
// Data that's too large to store inline is stored separately, in a data stream.
// This includes the contents of files, and the value of some attributes.

// JPhysExtKeyT is the key half of a physical extent record.
// Reference: page 102
type JPhysExtKeyT struct {
	// The record's header. (page 102)
	// The object identifier in the header is the physical block address of the start of the extent.
	// The type in the header is always APFS_TYPE_EXTENT.
	Hdr JKeyT
}

// JPhysExtValT is the value half of a physical extent record.
// Reference: page 102
type JPhysExtValT struct {
	// A bit field that contains the length of the extent and its kind. (page 102)
	// The extent's length is a uint64_t value, accessed as len_and_kind & PEXT_LEN_MASK,
	// and measured in blocks. The extent's kind is a j_obj_kinds value,
	// accessed as (len_and_kind & PEXT_KIND_MASK) >> PEXT_KIND_SHIFT.
	// For a volume that has no snapshots, the kind is always APFS_KIND_NEW.
	LenAndKind uint64

	// The identifier of the file system record that's using this extent. (page 103)
	// If the owning record is an inode, this field contains the inode's private identifier
	// (the private_id field of j_inode_val_t). If the owning record is an extended attribute,
	// this field contains the extended attribute's record identifier
	// (the identifier from the hdr field of j_xattr_key_t).
	OwningObjId uint64

	// The reference count. (page 103)
	// The extent can be deleted when its reference count reaches zero.
	Refcnt int32
}

// PextLenMask is the bit mask used to access the extent length.
// Reference: page 103
const PextLenMask uint64 = 0x0fffffffffffffff

// PextKindMask is the bit mask used to access the extent kind.
// Reference: page 103
const PextKindMask uint64 = 0xf000000000000000

// PextKindShift is the bit shift used to access the extent kind.
// Reference: page 103
const PextKindShift uint64 = 60

// JFileExtentKeyT is the key half of a file extent record.
// Reference: page 103
type JFileExtentKeyT struct {
	// The record's header. (page 103)
	// The object identifier in the header is the file-system object's identifier.
	// The type in the header is always APFS_TYPE_FILE_EXTENT.
	Hdr JKeyT

	// The offset within the file's data, in bytes, for the data stored in this extent. (page 104)
	LogicalAddr uint64
}

// JFileExtentValT is the value half of a file extent record.
// Reference: page 104
type JFileExtentValT struct {
	// A bit field that contains the length of the extent and its flags. (page 104)
	// The extent's length is a uint64_t value, accessed as len_and_kind & J_FILE_EXTENT_LEN_MASK,
	// and measured in bytes. The length must be a multiple of the block size defined
	// by the nx_block_size field of nx_superblock_t.
	// The extent's flags are accessed as (len_and_kind & J_FILE_EXTENT_FLAG_MASK) >> J_FILE_EXTENT_FLAG_SHIFT.
	// There are currently no flags defined.
	LenAndFlags uint64

	// The physical block address that the extent starts at. (page 104)
	PhysBlockNum uint64

	// The encryption key or the encryption tweak used in this extent. (page 104)
	// If the APFS_FS_ONEKEY flag is set on the volume, this field contains the AES-XTS tweak value.
	// Otherwise, this value matches the obj_id field of the j_crypto_key_t record that contains information
	// about how this file extent is encrypted, including the per-file encryption key.
	// The default value for this field is the value of the default_crypto_id field of the j_dstream_t
	// for the data stream that this extent is part of.
	CryptoId uint64
}

// JFileExtentLenMask is the bit mask used to access the extent length.
// Reference: page 105
const JFileExtentLenMask uint64 = 0x00ffffffffffffff

// JFileExtentFlagMask is the bit mask used to access the flags.
// Reference: page 105
const JFileExtentFlagMask uint64 = 0xff00000000000000

// JFileExtentFlagShift is the bit shift used to access the flags.
// Reference: page 105
const JFileExtentFlagShift uint64 = 56

// JDstreamIdKeyT is the key half of a directory-information record.
// Reference: page 105
type JDstreamIdKeyT struct {
	// The record's header. (page 105)
	// The object identifier in the header is the file-system object's identifier.
	// The type in the header is always APFS_TYPE_DSTREAM_ID.
	Hdr JKeyT
}

// JDstreamIdValT is the value half of a data stream record.
// Reference: page 105
type JDstreamIdValT struct {
	// The reference count. (page 105)
	// The data stream record can be deleted when its reference count reaches zero.
	Refcnt uint32
}

// JXattrDstreamT is a data stream for extended attributes.
// Reference: page 106
type JXattrDstreamT struct {
	// The identifier for the data stream. (page 106)
	// This field contains the record identifier of the data stream that owns this record.
	XattrObjId uint64

	// Information about the data stream. (page 106)
	Dstream JDstreamT
}

// JDstreamT contains information about a data stream.
// Reference: page 106
type JDstreamT struct {
	// The size, in bytes, of the data. (page 106)
	Size uint64

	// The total space allocated for the data stream, including any unused space. (page 106)
	AllocedSize uint64

	// The default encryption key or encryption tweak used in this data stream. (page 107)
	// This value matches the obj_id field in the j_key_t key that corresponds to a j_crypto_val_t value.
	// For a volume that uses software encryption, the value of this field is always CRYPTO_SW_ID.
	// This value is used as the default value by file extents (j_file_extent_val_t)
	// that make up this data stream.
	DefaultCryptoId uint64

	// The total number of bytes that have been written to this data stream. (page 107)
	// The value of this field increases every time a write operation occurs.
	// This value is allowed to overflow and restart from zero.
	TotalBytesWritten uint64

	// The total number of bytes that have been read from this data stream. (page 107)
	// The value of this field increases every time a read operation occurs.
	// This value is allowed to overflow and restart from zero.
	TotalBytesRead uint64
}

// FextCryptoIdIsTweak indicates that the crypto_id field of a file extent record contains
// an encryption tweak value.
// Reference: page 98
const FextCryptoIdIsTweak uint32 = 0x01
