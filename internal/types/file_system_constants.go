package types

// File-System Constants
// Reference: Apple File System Reference, pages 683-744

// JObjType represents the type of a file-system record.
// Used in B-tree keys to identify the type of data stored.
// Reference: page 687
type JObjType uint8

const (
	// JObjTypeAny matches any record type.
	// Used for generic B-tree operations that don't care about the specific type.
	// Reference: page 693
	JObjTypeAny JObjType = 0

	// JObjTypeSnapMetadata marks a snapshot metadata record.
	// Contains metadata about snapshots in the volume.
	// Reference: page 694
	JObjTypeSnapMetadata JObjType = 1

	// JObjTypeExtent marks an extent record.
	// Describes physical extents of file data.
	// Reference: page 695
	JObjTypeExtent JObjType = 2

	// JObjTypeInode marks an inode record.
	// Contains file metadata (permissions, timestamps, size, etc.).
	// Reference: page 696
	JObjTypeInode JObjType = 3

	// JObjTypeXattr marks an extended attribute record.
	// Stores extended file attributes and properties.
	// Reference: page 697
	JObjTypeXattr JObjType = 4

	// JObjTypeSiblingLink marks a sibling link record.
	// Used to link sibling files across snapshots.
	// Reference: page 698
	JObjTypeSiblingLink JObjType = 5

	// JObjTypeDStreamID marks a data stream ID record.
	// Identifies which data stream a block belongs to.
	// Reference: page 699
	JObjTypeDStreamID JObjType = 6

	// JObjTypeCryptoState marks a crypto state record.
	// Stores encryption state for files and volumes.
	// Reference: page 700
	JObjTypeCryptoState JObjType = 7

	// JObjTypeFileExtent marks a file extent record.
	// Describes physical extents specific to file data.
	// Reference: page 701
	JObjTypeFileExtent JObjType = 8

	// JObjTypeDirRec marks a directory record.
	// Contains entries (names and inode references) in a directory.
	// Reference: page 702
	JObjTypeDirRec JObjType = 9

	// JObjTypeDirStats marks a directory stats record.
	// Caches statistics about a directory (child count, etc.).
	// Reference: page 703
	JObjTypeDirStats JObjType = 10

	// JObjTypeSnapName marks a snapshot name record.
	// Maps snapshot names to their metadata.
	// Reference: page 704
	JObjTypeSnapName JObjType = 11

	// JObjTypeSiblingMap marks a sibling map record.
	// Maps inodes to their sibling counterparts across snapshots.
	// Reference: page 705
	JObjTypeSiblingMap JObjType = 12

	// JObjTypeFileInfo marks a file info record.
	// Stores file-specific information like data hashes for integrity checking.
	// Reference: page 706
	JObjTypeFileInfo JObjType = 13

	// JObjTypeMaxValid is the highest valid object type.
	// Use this to validate that a type value is within acceptable range.
	// Reference: page 707
	JObjTypeMaxValid JObjType = 13

	// JObjTypeMax represents the maximum object type value.
	// Used for type validation and range checking.
	// Reference: page 708
	JObjTypeMax JObjType = 15

	// JObjTypeInvalid marks an invalid record type.
	// Indicates a corrupted or unrecognized record type.
	// Reference: page 709
	JObjTypeInvalid JObjType = 15
)

// Inode Numbers
// Inodes whose number is always the same.
// Reference: page 713

// InvalidInoNum is an invalid inode number.
// Reference: page 718
const InvalidInoNum uint64 = 0

// RootDirParent is the inode number for the root directory's parent.
// Reference: page 719
// This is a sentinel value; there's no inode on disk with this inode number.
const RootDirParent uint64 = 1

// RootDirInoNum is the inode number for the root directory of the volume.
// Reference: page 720
const RootDirInoNum uint64 = 2

// PrivDirInoNum is the inode number for the private directory.
// Reference: page 721
// The private directory's filename is "private-dir". When creating a new volume,
// you must create a directory with this name and inode number.
const PrivDirInoNum uint64 = 3

// SnapDirInoNum is the inode number for the directory where snapshot metadata is stored.
// Reference: page 722
// Snapshot inodes are stored in the snapshot metadata tree.
const SnapDirInoNum uint64 = 6

// PurgeableDirInoNum is the inode number used for storing references to purgeable files.
// Reference: page 723
// This inode number and the directory records that use it are reserved.
// Other implementations of the Apple File System must not modify them.
// There isn't an actual directory with this inode number.
const PurgeableDirInoNum uint64 = 7

// MinUserInoNum is the smallest inode number available for user content.
// Reference: page 724
// All inode numbers less than this value are reserved.
const MinUserInoNum uint64 = 16

// UnifiedIDSpaceMark marks a unified ID space.
// Reference: page 725
const UnifiedIDSpaceMark uint64 = 0x0800000000000000

// File Modes
// The values used by the mode field of j_inode_val_t to indicate a file's mode.
// These follow POSIX file type conventions.
// Reference: page 728

// Mode represents file mode bits for inodes.
type Mode uint16

const (
	// ModeIFMT is the bit mask for the file type field.
	// AND this with a mode value to extract just the file type bits.
	// Reference: page 735
	ModeIFMT Mode = 0o170000

	// ModeIFIFO marks a FIFO (named pipe) file.
	// Used for inter-process communication.
	// Reference: page 736
	ModeIFIFO Mode = 0o010000

	// ModeIFCHR marks a character device file.
	// Represents unbuffered I/O devices (terminals, serial ports, etc.).
	// Reference: page 737
	ModeIFCHR Mode = 0o020000

	// ModeIFDIR marks a directory file.
	// Contains entries (files and subdirectories) indexed by name.
	// Reference: page 738
	ModeIFDIR Mode = 0o040000

	// ModeIFBLK marks a block device file.
	// Represents buffered I/O devices (disk drives, etc.).
	// Reference: page 739
	ModeIFBLK Mode = 0o060000

	// ModeIFREG marks a regular file.
	// Contains arbitrary data bytes (text, binary, etc.).
	// Reference: page 740
	ModeIFREG Mode = 0o100000

	// ModeIFLNK marks a symbolic link file.
	// Contains a path to another file (may be to a file on a different device).
	// Reference: page 741
	ModeIFLNK Mode = 0o120000

	// ModeIFSOCK marks a socket file.
	// Used for network communication endpoints.
	// Reference: page 742
	ModeIFSOCK Mode = 0o140000

	// ModeIFWHT marks a whiteout file.
	// Used in Union mounts to mark deleted files from lower layers.
	// Reference: page 743
	ModeIFWHT Mode = 0o160000
)
