package types

// Sealed Volumes (pages 150-158)
// Sealed volumes contain a hash of their file system, which can be compared to their
// current content to determine whether the volume has been modified after it was sealed,
// or compared to a known value to determine whether the volume contains the expected content.

// IntegrityMetaPhysT is integrity metadata for a sealed volume.
// Reference: page 150
type IntegrityMetaPhysT struct {
	// The object's header. (page 150)
	ImO ObjPhysT

	// The version of this data structure. (page 150)
	ImVersion uint32

	// The flags used to describe configuration options. (page 151)
	// For the values used in this bit field, see Integrity Metadata Flags.
	ImFlags uint32

	// The hash algorithm being used. (page 151)
	ImHashType ApfsHashTypeT

	// The offset, in bytes, of the root hash relative to the start of this integrity metadata object. (page 151)
	ImRootHashOffset uint32

	// The identifier of the transaction that unsealed the volume. (page 151)
	// When a sealed volume is modified, breaking its seal, that transaction identifier is recorded
	// in this field and the APFS_SEAL_BROKEN flag is set. Otherwise, the value of this field is zero.
	ImBrokenXid XidT

	// Reserved. (page 151)
	// This field appears in version 2 and later of this data structure.
	ImReserved [9]uint64
}

// Integrity Metadata Version Constants (pages 151-152)

// These constants are used as the value of the im_version field of the integrity_meta_phys_t structure.

const (
	// IntegrityMetaVersionInvalid indicates an invalid version.
	// Reference: page 152
	IntegrityMetaVersionInvalid uint32 = 0

	// IntegrityMetaVersion1 indicates the first version of the structure.
	// Reference: page 152
	IntegrityMetaVersion1 uint32 = 1

	// IntegrityMetaVersion2 indicates the second version of the structure.
	// Reference: page 152
	IntegrityMetaVersion2 uint32 = 2

	// IntegrityMetaVersionHighest indicates the highest valid version number.
	// Reference: page 152
	IntegrityMetaVersionHighest uint32 = IntegrityMetaVersion2
)

// Integrity Metadata Flags (page 152)

// ApfsSealBroken indicates the volume was modified after being sealed, breaking its seal.
// Reference: page 152
// If this flag is set, the im_broken_xid field of integrity_meta_phys_t contains the
// transaction identifier for the modification that broke the seal.
const ApfsSealBroken uint32 = 1 << 0

// ApfsHashTypeT contains constants used to identify hash algorithms.
// Reference: page 152
type ApfsHashTypeT uint32

const (
	// ApfsHashInvalid indicates an invalid hash algorithm.
	// Reference: page 153
	ApfsHashInvalid ApfsHashTypeT = 0

	// ApfsHashSha256 indicates the SHA-256 variant of Secure Hash Algorithm 2.
	// Reference: page 153
	ApfsHashSha256 ApfsHashTypeT = 0x1

	// ApfsHashSha512256 indicates the SHA-512/256 variant of Secure Hash Algorithm 2.
	// Reference: page 153
	ApfsHashSha512256 ApfsHashTypeT = 0x2

	// ApfsHashSha384 indicates the SHA-384 variant of Secure Hash Algorithm 2.
	// Reference: page 153
	ApfsHashSha384 ApfsHashTypeT = 0x3

	// ApfsHashSha512 indicates the SHA-512 variant of Secure Hash Algorithm 2.
	// Reference: page 153
	ApfsHashSha512 ApfsHashTypeT = 0x4

	// ApfsHashMin is the smallest valid value for identifying a hash algorithm.
	// Reference: page 153
	ApfsHashMin ApfsHashTypeT = ApfsHashSha256

	// ApfsHashMax is the largest valid value for identifying a hash algorithm.
	// Reference: page 154
	ApfsHashMax ApfsHashTypeT = ApfsHashSha512

	// ApfsHashDefault is the default hash algorithm.
	// Reference: page 154
	ApfsHashDefault ApfsHashTypeT = ApfsHashSha256
)

// ApfsHashCcsha256Size is the size of a SHA-256 hash.
// Reference: page 154
const ApfsHashCcsha256Size uint32 = 32

// ApfsHashCcsha512256Size is the size of a SHA-512/256 hash.
// Reference: page 154
const ApfsHashCcsha512256Size uint32 = 32

// ApfsHashCcsha384Size is the size of a SHA-384 hash.
// Reference: page 154
const ApfsHashCcsha384Size uint32 = 48

// ApfsHashCcsha512Size is the size of a SHA-512 hash.
// Reference: page 154
const ApfsHashCcsha512Size uint32 = 64

// ApfsHashMaxSize is the maximum valid hash size.
// Reference: page 154
// This value is the same as BTREE_NODE_HASH_SIZE_MAX.
const ApfsHashMaxSize uint32 = 64

// FextTreeKeyT is the key half of a record from a file extent tree.
// Reference: page 154
type FextTreeKeyT struct {
	// The object identifier of the file. (page 155)
	// This value corresponds the object identifier portion of the obj_id_and_type field of j_key_t.
	PrivateId uint64

	// The offset within the file's data, in bytes, for the data stored in this extent. (page 155)
	LogicalAddr uint64
}

// FextTreeValT is the value half of a record from a file extent tree.
// Reference: page 155
type FextTreeValT struct {
	// A bit field that contains the length of the extent and its flags. (page 155)
	// The extent's length is a uint64_t value, accessed as len_and_kind & J_FILE_EXTENT_LEN_MASK,
	// and measured in bytes. The length must be a multiple of the block size defined by
	// the nx_block_size field of nx_superblock_t.
	// The extent's flags are accessed as
	// (len_and_kind & J_FILE_EXTENT_FLAG_MASK) >> J_FILE_EXTENT_FLAG_SHIFT.
	// There are currently no flags defined.
	LenAndFlags uint64

	// The physical block address that the extent starts at. (page 155)
	PhysBlockNum uint64
}

// JFileInfoKeyT is the key half of a file-info record.
// Reference: page 155
type JFileInfoKeyT struct {
	// The record's header. (page 156)
	// The object identifier in the header is the file-system object's identifier.
	// The type in the header is always APFS_TYPE_FILE_INFO.
	Hdr JKeyT

	// A bit field that contains the address and other information. (page 156)
	// The address is a paddr_t value accessed as info_and_lba & J_FILE_INFO_LBA_MASK.
	// The type is a j_obj_file_info_type value accessed as
	// (info_and_lba & J_FILE_INFO_TYPE_MASK) >> J_FILE_INFO_TYPE_SHIFT.
	InfoAndLba uint64
}

// JFileInfoLbaMask is the bit mask used to access file-info addresses.
// Reference: page 156
const JFileInfoLbaMask uint64 = 0x00ffffffffffffff

// JFileInfoTypeMask is the bit mask used to access file-info types.
// Reference: page 156
const JFileInfoTypeMask uint64 = 0xff00000000000000

// JFileInfoTypeShift is the bit shift used to access file-info types.
// Reference: page 156
const JFileInfoTypeShift uint64 = 56

// JFileInfoValT is the value half of a file-info record.
// Reference: page 156
type JFileInfoValT struct {
	// A union containing the different types of file info.
	// Reference: page 156
	// Use the type stored in the j_file_info_key_t half of this record to determine
	// which of the union's fields to use.

	// A hash of the file data.
	Dhash JFileDataHashValT
}

// JObjFileInfoType represents the type of a file-info record.
// Reference: page 157
type JObjFileInfoType uint8

const (
	// ApfsFileInfoDataHash indicates the file-info record contains a hash of file data.
	// Reference: page 157
	ApfsFileInfoDataHash JObjFileInfoType = 1

	// Maximum valid file info type - currently only one type is defined in the spec
	ApfsFileInfoMaxValid JObjFileInfoType = ApfsFileInfoDataHash
)

// JFileDataHashValT contains a hash of file data.
// Reference: page 157
type JFileDataHashValT struct {
	// The length, in blocks, of the data segment that was hashed. (page 157)
	HashedLen uint16

	// The length, in bytes, of the hash data. (page 157)
	// The value of this field must match the constant that corresponds to the hash algorithm specified
	// in the im_hash_type field of integrity_meta_phys_t.
	HashSize uint8

	// The hash data. APFS hash sizes are fixed per algorithm (e.g., 32, 48, 64 bytes). (page 158)
	Hash [ApfsHashMaxSize]byte
}
