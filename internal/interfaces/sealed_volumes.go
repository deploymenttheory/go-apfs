package interfaces

import (
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// IntegrityReader provides methods for reading integrity metadata of a sealed volume
type IntegrityReader interface {
	// Version returns the version of the integrity metadata structure
	Version() uint32

	// Flags returns the integrity metadata flags
	Flags() uint32

	// HashType returns the hash algorithm being used
	HashType() types.ApfsHashTypeT

	// RootHashOffset returns the offset of the root hash
	RootHashOffset() uint32
}

// SealedVolumeChecker provides methods for checking the integrity of a sealed volume
type SealedVolumeChecker interface {
	// IsSealed checks if the volume is currently sealed
	IsSealed() bool

	// IsSealBroken checks if the volume's seal has been broken
	IsSealBroken() bool

	// SealBreakTransactionID returns the transaction ID that broke the seal
	// Returns 0 if the seal is intact
	SealBreakTransactionID() types.XidT

	// VerifyIntegrity performs an integrity check of the volume
	VerifyIntegrity() (IntegrityVerificationResult, error)
}

// FileIntegrityReader provides methods for reading file-specific integrity information
type FileIntegrityReader interface {
	// DataHash returns the hash of the file's data
	DataHash() []byte

	// HashType returns the type of hash used
	HashType() types.ApfsHashTypeT

	// HashedLength returns the length of the data segment that was hashed
	HashedLength() uint16
}

// IntegrityVerificationResult represents the result of an integrity verification
type IntegrityVerificationResult struct {
	// Indicates whether the integrity check passed
	Passed bool

	// Detailed information about the integrity check
	Details string

	// List of any files that failed integrity verification
	FailedFiles []string
}

// HashAlgorithmInfo provides information about hash algorithms
type HashAlgorithmInfo interface {
	// Name returns the name of the hash algorithm
	Name() string

	// Size returns the size of the hash in bytes
	Size() uint32

	// IsSupported checks if the hash algorithm is supported
	IsSupported() bool
}

// IntegrityManager provides methods for managing volume integrity
type IntegrityManager interface {
	// ListHashAlgorithms returns all supported hash algorithms
	ListHashAlgorithms() []HashAlgorithmInfo

	// GetDefaultHashAlgorithm returns the default hash algorithm
	GetDefaultHashAlgorithm() HashAlgorithmInfo

	// ComputeVolumeHash computes the hash of the entire volume
	ComputeVolumeHash(hashType types.ApfsHashTypeT) ([]byte, error)
}
