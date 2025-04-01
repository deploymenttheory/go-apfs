// File: internal/interfaces/encryption_enrollment.go
package interfaces

import (
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// EncryptionRollingStateReader provides methods for reading the encryption rolling state
type EncryptionRollingStateReader interface {
	// Version returns the version number of the encryption rolling state
	Version() uint32

	// Magic returns the magic number for validating the encryption rolling state
	Magic() uint32

	// Flags returns the encryption rolling state flags
	Flags() uint64

	// SnapshotXID returns the snapshot transaction identifier
	SnapshotXID() types.XidT

	// CurrentFileExtentObjectID returns the current file extent object identifier
	CurrentFileExtentObjectID() uint64

	// FileOffset returns the file offset where encryption rolling is currently at
	FileOffset() uint64

	// Progress returns the current progress of encryption rolling
	Progress() uint64

	// TotalBlocksToEncrypt returns the total number of blocks to encrypt
	TotalBlocksToEncrypt() uint64

	// BlockmapOID returns the object identifier of the block map
	BlockmapOID() types.OidT

	// TidemarkObjectID returns the tidemark object identifier
	TidemarkObjectID() uint64

	// RecoveryExtentsCount returns the count of recovery extents
	RecoveryExtentsCount() uint64

	// RecoveryListOID returns the object identifier of the recovery list
	RecoveryListOID() types.OidT

	// RecoveryLength returns the length of the recovery
	RecoveryLength() uint64
}

// EncryptionRollingV1StateReader provides methods for reading version 1 of the encryption rolling state
type EncryptionRollingV1StateReader interface {
	// Version returns the version number of the encryption rolling state
	Version() uint32

	// Magic returns the magic number for validating the encryption rolling state
	Magic() uint32

	// Flags returns the encryption rolling state flags
	Flags() uint64

	// SnapshotXID returns the snapshot transaction identifier
	SnapshotXID() types.XidT

	// CurrentFileExtentObjectID returns the current file extent object identifier
	CurrentFileExtentObjectID() uint64

	// FileOffset returns the file offset where encryption rolling is currently at
	FileOffset() uint64

	// FileExtentPhysicalBlockNumber returns the file extent physical block number
	FileExtentPhysicalBlockNumber() uint64

	// PhysicalAddress returns the physical address
	PhysicalAddress() uint64

	// Progress returns the current progress of encryption rolling
	Progress() uint64

	// TotalBlocksToEncrypt returns the total number of blocks to encrypt
	TotalBlocksToEncrypt() uint64

	// BlockmapOID returns the object identifier of the block map
	BlockmapOID() uint64

	// ChecksumCount returns the count of checksums
	ChecksumCount() uint32

	// FileExtentCryptoID returns the file extent crypto identifier
	FileExtentCryptoID() uint64

	// Checksums returns the checksums for the file extents
	Checksums() []byte
}

// EncryptionRollingFlagManager provides methods for working with encryption rolling flags
type EncryptionRollingFlagManager interface {
	// IsEncrypting checks if encryption is in progress
	IsEncrypting() bool

	// IsDecrypting checks if decryption is in progress
	IsDecrypting() bool

	// IsKeyRolling checks if key rolling is in progress
	IsKeyRolling() bool

	// IsPaused checks if encryption rolling is paused
	IsPaused() bool

	// HasFailed checks if encryption rolling has failed
	HasFailed() bool

	// IsCIDTweak checks if the crypto ID is a tweak
	IsCIDTweak() bool

	// GetBlockSize returns the block size used for encryption
	GetBlockSize() uint64

	// GetPhase returns the current phase of encryption rolling
	GetPhase() types.ErPhaseT

	// IsFromOneKey checks if encryption rolling is from a one-key system
	IsFromOneKey() bool
}

// EncryptionRollingPhaseManager provides methods for managing encryption rolling phases
type EncryptionRollingPhaseManager interface {
	// GetCurrentPhase returns the current phase of encryption rolling
	GetCurrentPhase() types.ErPhaseT

	// GetPhaseDescription returns a human-readable description of the current phase
	GetPhaseDescription() string

	// IsOmapRollPhase checks if currently in the object map roll phase
	IsOmapRollPhase() bool

	// IsDataRollPhase checks if currently in the data roll phase
	IsDataRollPhase() bool

	// IsSnapshotRollPhase checks if currently in the snapshot roll phase
	IsSnapshotRollPhase() bool
}

// EncryptionRollingRecoveryBlockReader provides methods for reading recovery blocks
type EncryptionRollingRecoveryBlockReader interface {
	// Offset returns the offset of the recovery block
	Offset() uint64

	// NextObjectID returns the object identifier of the next recovery block
	NextObjectID() types.OidT

	// Data returns the data in the recovery block
	Data() []byte
}

// GeneralBitmapReader provides methods for reading general bitmaps used in encryption rolling
type GeneralBitmapReader interface {
	// TreeObjectID returns the object identifier of the bitmap tree
	TreeObjectID() types.OidT

	// BitCount returns the number of bits in the bitmap
	BitCount() uint64

	// Flags returns the flags for the bitmap
	Flags() uint64
}

// GeneralBitmapBlockReader provides methods for reading bitmap blocks
type GeneralBitmapBlockReader interface {
	// BitmapField returns the bitmap data as an array of uint64
	BitmapField() []uint64

	// IsBitSet checks if a specific bit is set in the bitmap
	IsBitSet(bitIndex uint64) bool

	// SetBit sets a specific bit in the bitmap
	SetBit(bitIndex uint64)

	// ClearBit clears a specific bit in the bitmap
	ClearBit(bitIndex uint64)
}

// EncryptionRollingManager provides methods for managing the encryption rolling process
type EncryptionRollingManager interface {
	// Start initiates the encryption rolling process
	Start() error

	// Pause pauses the encryption rolling process
	Pause() error

	// Resume resumes the encryption rolling process after a pause
	Resume() error

	// Cancel cancels the encryption rolling process
	Cancel() error

	// Status returns the current status of encryption rolling
	Status() (EncryptionRollingStatus, error)
}

// EncryptionRollingStatus represents the current status of an encryption rolling operation
type EncryptionRollingStatus struct {
	// The current state (encrypting, decrypting, key rolling, paused, failed)
	State string

	// The current phase of encryption rolling
	Phase string

	// The percentage of completion (0-100)
	ProgressPercentage float64

	// The block size being used
	BlockSize uint64

	// Remaining blocks to be processed
	RemainingBlocks uint64

	// Total blocks to be processed
	TotalBlocks uint64

	// Error message if the process has failed
	ErrorMessage string
}

// EncryptionRollingLogReader provides methods for reading encryption rolling events
type EncryptionRollingLogReader interface {
	// GetHistory returns the history of encryption rolling events
	GetHistory() ([]EncryptionRollingEvent, error)

	// GetLastError returns the last error that occurred during encryption rolling
	GetLastError() (string, error)
}

// EncryptionRollingEvent represents a single event in the encryption rolling process
type EncryptionRollingEvent struct {
	// The timestamp of the event
	Timestamp uint64

	// The type of event (start, pause, resume, progress, complete, error)
	EventType string

	// Additional details about the event
	Details string

	// The progress percentage at the time of the event
	ProgressPercentage float64
}

// BlockSizeResolver provides methods for resolving encryption rolling block sizes
type BlockSizeResolver interface {
	// GetBlockSizeValue returns the actual size in bytes for a given block size constant
	GetBlockSizeValue(blockSizeConstant uint32) uint32

	// GetBlockSizeConstant returns the block size constant for a given size in bytes
	GetBlockSizeConstant(sizeInBytes uint32) uint32

	// GetSupportedBlockSizes returns all supported block sizes in bytes
	GetSupportedBlockSizes() []uint32
}
