package types

// Encryption Rolling (pages 169-171)
// Encryption rolling refers to the process of changing the encryption settings for a volume,
// such as enabling or disabling encryption, or changing the encryption key.

// ErStatePhysT represents the encryption-rolling state.
// Reference: page 169
type ErStatePhysT struct {
	// The header for the encryption-rolling state.
	ErsbHeader ErStatePhysHeaderT

	// The flags for the encryption-rolling state.
	ErsbFlags uint64

	// The snapshot transaction identifier.
	ErsbSnapXid uint64

	// The current file extent object identifier.
	ErsbCurrentFextObjId uint64

	// The file offset.
	ErsbFileOffset uint64

	// The progress of the encryption rolling.
	ErsbProgress uint64

	// The total number of blocks to encrypt.
	ErsbTotalBlkToEncrypt uint64

	// The object identifier of the block map.
	ErsbBlockmapOid OidT

	// The tidemark object identifier.
	ErsbTidemarkObjId uint64

	// The count of recovery extents.
	ErsbRecoveryExtentsCount uint64

	// The object identifier of the recovery list.
	ErsbRecoveryListOid OidT

	// The length of the recovery.
	ErsbRecoveryLength uint64
}

// ErStatePhysV1T represents version 1 of the encryption-rolling state.
// Reference: page 169
type ErStatePhysV1T struct {
	// The header for the encryption-rolling state.
	ErsbHeader ErStatePhysHeaderT

	// The flags for the encryption-rolling state.
	ErsbFlags uint64

	// The snapshot transaction identifier.
	ErsbSnapXid uint64

	// The current file extent object identifier.
	ErsbCurrentFextObjId uint64

	// The file offset.
	ErsbFileOffset uint64

	// The file extent physical block number.
	ErsbFextPbn uint64

	// The physical address.
	ErsbPaddr uint64

	// The progress of the encryption rolling.
	ErsbProgress uint64

	// The total number of blocks to encrypt.
	ErsbTotalBlkToEncrypt uint64

	// The object identifier of the block map.
	ErsbBlockmapOid uint64

	// The count of checksums.
	ErsbChecksumCount uint32

	// Reserved.
	ErsbReserved uint32

	// The file extent crypto identifier.
	ErsbFextCid uint64

	// The checksums for the file extents.
	ErsbChecksum []byte
}

// ErStatePhysHeaderT is the header for the encryption-rolling state.
// Reference: page 169
type ErStatePhysHeaderT struct {
	// The object's header.
	ErsbO ObjPhysT

	// The magic number for the encryption-rolling state.
	ErsbMagic uint32

	// The version of the encryption-rolling state.
	ErsbVersion uint32
}

// ErPhaseT represents the phase of encryption rolling.
// Reference: page 170
type ErPhaseT uint32

const (
	// ErPhaseOmapRoll is the phase where the object map is being rolled.
	// Reference: page 170
	ErPhaseOmapRoll ErPhaseT = 1

	// ErPhaseDataRoll is the phase where the data is being rolled.
	// Reference: page 170
	ErPhaseDataRoll ErPhaseT = 2

	// ErPhaseSnapRoll is the phase where the snapshots are being rolled.
	// Reference: page 170
	ErPhaseSnapRoll ErPhaseT = 3
)

// ErRecoveryBlockPhysT represents a recovery block for encryption rolling.
// Reference: page 170
type ErRecoveryBlockPhysT struct {
	// The object's header.
	ErbO ObjPhysT

	// The offset of the recovery block.
	ErbOffset uint64

	// The object identifier of the next recovery block.
	ErbNextOid OidT

	// The data in the recovery block.
	ErbData []byte
}

// GbitmapBlockPhysT represents a general bitmap block.
// Reference: page 170
type GbitmapBlockPhysT struct {
	// The object's header.
	BmbO ObjPhysT

	// The bitmap field.
	BmbField []uint64
}

// GbitmapPhysT represents a general bitmap.
// Reference: page 170
type GbitmapPhysT struct {
	// The object's header.
	BmO ObjPhysT

	// The object identifier of the bitmap tree.
	BmTreeOid OidT

	// The number of bits in the bitmap.
	BmBitCount uint64

	// The flags for the bitmap.
	BmFlags uint64
}

// Encryption-Rolling Checksum Block Sizes (page 170)

const (
	// Er512bBlocksize is the 512-byte block size.
	// Reference: page 171
	Er512bBlocksize uint32 = 0

	// Er2kibBlocksize is the 2 KiB block size.
	// Reference: page 171
	Er2kibBlocksize uint32 = 1

	// Er4kibBlocksize is the 4 KiB block size.
	// Reference: page 171
	Er4kibBlocksize uint32 = 2

	// Er8kibBlocksize is the 8 KiB block size.
	// Reference: page 171
	Er8kibBlocksize uint32 = 3

	// Er16kibBlocksize is the 16 KiB block size.
	// Reference: page 171
	Er16kibBlocksize uint32 = 4

	// Er32kibBlocksize is the 32 KiB block size.
	// Reference: page 171
	Er32kibBlocksize uint32 = 5

	// Er64kibBlocksize is the 64 KiB block size.
	// Reference: page 171
	Er64kibBlocksize uint32 = 6
)

// Encryption Rolling Flags (page 171)

// ErsbFlagEncrypting indicates that encryption is in progress.
// Reference: page 171
const ErsbFlagEncrypting uint64 = 0x00000001

// ErsbFlagDecrypting indicates that decryption is in progress.
// Reference: page 171
const ErsbFlagDecrypting uint64 = 0x00000002

// ErsbFlagKeyrolling indicates that key rolling is in progress.
// Reference: page 171
const ErsbFlagKeyrolling uint64 = 0x00000004

// ErsbFlagPaused indicates that encryption rolling is paused.
// Reference: page 171
const ErsbFlagPaused uint64 = 0x00000008

// ErsbFlagFailed indicates that encryption rolling has failed.
// Reference: page 171
const ErsbFlagFailed uint64 = 0x00000010

// ErsbFlagCidIsTweak indicates that the crypto ID is a tweak.
// Reference: page 171
const ErsbFlagCidIsTweak uint64 = 0x00000020

// ErsbFlagFree1 is a free flag that can be used by implementations.
// Reference: page 171
const ErsbFlagFree1 uint64 = 0x00000040

// ErsbFlagFree2 is a free flag that can be used by implementations.
// Reference: page 171
const ErsbFlagFree2 uint64 = 0x00000080

// ErsbFlagCmBlockSizeMask is the mask for the block size flags.
// Reference: page 171
const ErsbFlagCmBlockSizeMask uint64 = 0x00000F00

// ErsbFlagCmBlockSizeShift is the shift for the block size flags.
// Reference: page 171
const ErsbFlagCmBlockSizeShift uint64 = 8

// ErsbFlagErPhaseMask is the mask for the phase flags.
// Reference: page 171
const ErsbFlagErPhaseMask uint64 = 0x00003000

// ErsbFlagErPhaseShift is the shift for the phase flags.
// Reference: page 171
const ErsbFlagErPhaseShift uint64 = 12

// ErsbFlagFromOnekey indicates that encryption rolling is from a one-key system.
// Reference: page 171
const ErsbFlagFromOnekey uint64 = 0x00004000

// Encryption-Rolling Constants (page 171)

// ErChecksumLength is the length of a checksum for encryption rolling.
// Reference: page 171
const ErChecksumLength uint32 = 8

// ErMagic is the magic number for encryption rolling.
// Reference: page 171
const ErMagic uint32 = 'F' | 'L'<<8 | 'A'<<16 | 'B'<<24 // 'FLAB'

// ErVersion is the version number for encryption rolling.
// Reference: page 171
const ErVersion uint32 = 1

// ErMaxChecksumCountShift is the shift for the maximum checksum count.
// Reference: page 171
const ErMaxChecksumCountShift uint32 = 16

// ErCurChecksumCountMask is the mask for the current checksum count.
// Reference: page 171
const ErCurChecksumCountMask uint32 = 0x0000FFFF
