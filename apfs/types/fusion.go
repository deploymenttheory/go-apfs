package types

// Fusion (pages 172-173)
// Fusion refers to a storage configuration that combines a solid-state drive with a hard disk drive,
// using the solid-state drive to cache frequently accessed data.

// FusionWbcPhysT represents the Fusion write-back cache state.
// Reference: page 172
type FusionWbcPhysT struct {
	// The object's header.
	FwpObjHdr ObjPhysT

	// The version of the write-back cache.
	FwpVersion uint64

	// The object identifier of the head of the write-back cache list.
	FwpListHeadOid OidT

	// The object identifier of the tail of the write-back cache list.
	FwpListTailOid OidT

	// The offset of the stable head of the write-back cache.
	FwpStableHeadOffset uint64

	// The offset of the stable tail of the write-back cache.
	FwpStableTailOffset uint64

	// The number of blocks in the write-back cache list.
	FwpListBlocksCount uint32

	// Reserved.
	FwpReserved uint32

	// The amount of space used by the read cache.
	FwpUsedByRC uint64

	// The location of the read cache stash.
	FwpRcStash Prange
}

// FusionWbcListEntryT represents an entry in the Fusion write-back cache list.
// Reference: page 172
type FusionWbcListEntryT struct {
	// The logical block address in the write-back cache.
	FwleWbcLba Paddr

	// The target logical block address.
	FwleTargetLba Paddr

	// The length of the entry.
	FwleLength uint64
}

// FusionWbcListPhysT represents the Fusion write-back cache list.
// Reference: page 172
type FusionWbcListPhysT struct {
	// The object's header.
	FwlpObjHdr ObjPhysT

	// The version of the write-back cache list.
	FwlpVersion uint64

	// The offset of the tail of the write-back cache list.
	FwlpTailOffset uint64

	// The beginning index of the write-back cache list.
	FwlpIndexBegin uint32

	// The ending index of the write-back cache list.
	FwlpIndexEnd uint32

	// The maximum index of the write-back cache list.
	FwlpIndexMax uint32

	// Reserved.
	FwlpReserved uint32

	// The entries in the write-back cache list.
	FwlpListEntries []FusionWbcListEntryT
}

// FusionTier2DeviceByteAddr is the address marker for the Tier 2 device.
// Reference: page 173
const FusionTier2DeviceByteAddr uint64 = 0x4000000000000000

// FusionMtKeyT is a key for the Fusion middle tree.
// Reference: page 173
type FusionMtKeyT Paddr

// FusionMtValT is a value in the Fusion middle tree.
// Reference: page 173
type FusionMtValT struct {
	// The logical block address.
	FmvLba Paddr

	// The length of the extent.
	FmvLength uint32

	// The flags for the extent.
	FmvFlags uint32
}

// Fusion Middle-Tree Flags (page 173)

// FusionMtDirty indicates that the extent is dirty.
// Reference: page 173
const FusionMtDirty uint32 = 1 << 0

// FusionMtTenant indicates that the extent is a tenant.
// Reference: page 173
const FusionMtTenant uint32 = 1 << 1

// FusionMtAllflags is a bit mask of all the flags for a Fusion middle tree.
// Reference: page 173
const FusionMtAllflags uint32 = FusionMtDirty | FusionMtTenant
