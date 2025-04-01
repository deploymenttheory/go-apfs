package interfaces

import (
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// CryptoStateReader provides methods for reading encryption state
type CryptoStateReader interface {
	// ReferenceCount returns the reference count for the encryption state
	ReferenceCount() uint32

	// ProtectionClass returns the protection class of the encryption state
	ProtectionClass() types.CpKeyClassT

	// KeyVersion returns the version of the encryption key
	KeyVersion() types.CpKeyRevisionT

	// IsValid checks if the encryption state is valid
	IsValid() bool
}

// EncryptionKeyReader provides methods for reading encryption key information
type EncryptionKeyReader interface {
	// MajorVersion returns the major version of the key structure
	MajorVersion() uint16

	// MinorVersion returns the minor version of the key structure
	MinorVersion() uint16

	// KeyLength returns the length of the wrapped key data
	KeyLength() uint16

	// WrappedKeyData returns the encrypted key data
	WrappedKeyData() []byte
}

// KeybagReader provides methods for reading keybag information
type KeybagReader interface {
	// Version returns the keybag version
	Version() uint16

	// EntryCount returns the number of entries in the keybag
	EntryCount() uint16

	// TotalDataSize returns the total size of keybag entries in bytes
	TotalDataSize() uint32

	// ListEntries returns all keybag entries
	ListEntries() []KeybagEntryReader
}

// KeybagEntryReader provides methods for reading individual keybag entries
type KeybagEntryReader interface {
	// UUID returns the UUID associated with the entry
	UUID() types.UUID

	// Tag returns the keybag entry tag
	Tag() types.KbTag

	// TagDescription returns a human-readable description of the tag
	TagDescription() string

	// KeyLength returns the length of the entry's key data
	KeyLength() uint16

	// KeyData returns the raw key data
	KeyData() []byte
}

// ProtectionClassResolver provides methods for resolving protection class information
type ProtectionClassResolver interface {
	// ResolveName returns a human-readable name for a protection class
	ResolveName(class types.CpKeyClassT) string

	// ResolveDescription provides a detailed description of a protection class
	ResolveDescription(class types.CpKeyClassT) string

	// ListSupportedProtectionClasses returns all supported protection classes
	ListSupportedProtectionClasses() []types.CpKeyClassT
}

// EncryptionInspector provides methods for comprehensive encryption analysis
type EncryptionInspector interface {
	// IsEncryptionEnabled checks if encryption is currently enabled
	IsEncryptionEnabled() bool

	// GetCryptoIdentifier returns the unique identifier for the encryption state
	GetCryptoIdentifier() uint64

	// AnalyzeEncryptionState provides a detailed analysis of the current encryption configuration
	AnalyzeEncryptionState() (EncryptionStateAnalysis, error)
}

// EncryptionStateAnalysis contains detailed information about an encryption state
type EncryptionStateAnalysis struct {
	// Indicates if the volume is encrypted
	IsEncrypted bool

	// Protection class in use
	ProtectionClass string

	// Key version information
	KeyVersion uint16

	// Indicates if the encryption state is valid
	IsValid bool

	// Additional metadata about the encryption state
	Metadata map[string]string
}

// KeyRollingManager provides methods for managing encryption key transitions
type KeyRollingManager interface {
	// IsKeyRollingInProgress checks if a key rotation is currently happening
	IsKeyRollingInProgress() bool

	// GetPreviousKeyVersion returns the version of the previous encryption key
	GetPreviousKeyVersion() types.CpKeyRevisionT

	// GetCurrentKeyVersion returns the version of the current encryption key
	GetCurrentKeyVersion() types.CpKeyRevisionT
}

type DecryptionManager interface {
	// Authenticate attempts to unlock encryption using a provided passphrase
	Authenticate(passphrase string) (bool, error)

	// DecryptFile decrypts a specific file to a destination
	DecryptFile(sourcePath, destinationPath string) error

	// DecryptVolume provides methods to decrypt an entire volume
	DecryptVolume(outputPath string) error

	// IsDecryptionPossible checks if decryption can be performed
	IsDecryptionPossible() bool
}
