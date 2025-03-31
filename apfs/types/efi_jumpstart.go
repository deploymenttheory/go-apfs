package types

// EFI Jumpstart (pages 22-25)
// A partition formatted using the Apple File System contains an embedded
// EFI driver that's used to boot a machine from that partition.

// NxEfiJumpstartT represents information about the embedded EFI driver that's
// used to boot from an Apple File System partition.
// Reference: page 24
type NxEfiJumpstartT struct {
	// The object's header. (page 24)
	NejO ObjPhysT
	// A number that can be used to verify that you're reading an instance of nx_efi_jumpstart_t. (page 24)
	// The value of this field is always NxEfiJumpstartMagic.
	NejMagic uint32
	// The version of this data structure. (page 24)
	// The value of this field is always NxEfiJumpstartVersion.
	NejVersion uint32
	// The size, in bytes, of the embedded EFI driver. (page 24)
	NejEfiFileLen uint32
	// The number of extents in the array. (page 25)
	NejNumExtents uint32
	// Reserved. (page 25)
	// Populate this field with zero when you create a new instance, and preserve its value when you modify an existing
	// instance.
	NejReserved [16]uint64
	// The locations where the EFI driver is stored. (page 25)
	// This is a Go slice representing the C flexible array member.
	NejRecExtents []Prange
}

// NxEfiJumpstartMagic is the value of the nej_magic field.
// This magic number was chosen because in hex dumps it appears as "JSDR",
// which is an abbreviated form of jumpstart driver record.
// Reference: page 25
const NxEfiJumpstartMagic uint32 = 'R' | 'D'<<8 | 'S'<<16 | 'J'<<24 // 'RDSJ'

// NxEfiJumpstartVersion is the version number for the EFI jumpstart.
// Reference: page 25
const NxEfiJumpstartVersion uint32 = 1

// ApfsGptPartitionUUID is the partition type for a partition that contains an Apple File System container.
// Reference: page 25
const ApfsGptPartitionUUID string = "7C3457EF-0000-11AA-AA11-00306543ECAC"
