package encryption

import (
	"fmt"

	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// encryptionInspector implements the EncryptionInspector interface
type encryptionInspector struct {
	cryptoStateReader  interfaces.CryptoStateReader
	protectionResolver interfaces.ProtectionClassResolver
}

// Ensure encryptionInspector implements the EncryptionInspector interface
var _ interfaces.EncryptionInspector = (*encryptionInspector)(nil)

// NewEncryptionInspector creates a new EncryptionInspector
func NewEncryptionInspector(cryptoStateReader interfaces.CryptoStateReader) interfaces.EncryptionInspector {
	return &encryptionInspector{
		cryptoStateReader:  cryptoStateReader,
		protectionResolver: NewProtectionClassResolver(),
	}
}

// IsEncryptionEnabled checks if encryption is currently enabled
func (ei *encryptionInspector) IsEncryptionEnabled() bool {
	if ei.cryptoStateReader == nil {
		return false
	}

	// Check if the crypto state is valid and has a valid protection class
	return ei.cryptoStateReader.IsValid() &&
		ei.protectionResolver.IsValidProtectionClass(ei.cryptoStateReader.ProtectionClass())
}

// GetCryptoIdentifier returns the unique identifier for the encryption state
func (ei *encryptionInspector) GetCryptoIdentifier() uint64 {
	if ei.cryptoStateReader == nil {
		return 0
	}

	// Return the object ID from the crypto state reader
	return uint64(ei.cryptoStateReader.ObjectID())
}

// AnalyzeEncryptionState provides a detailed analysis of the current encryption configuration
func (ei *encryptionInspector) AnalyzeEncryptionState() (interfaces.EncryptionStateAnalysis, error) {
	analysis := interfaces.EncryptionStateAnalysis{
		Metadata: make(map[string]string),
	}

	if ei.cryptoStateReader == nil {
		return analysis, fmt.Errorf("no crypto state reader available")
	}

	// Basic encryption status
	analysis.IsEncrypted = ei.IsEncryptionEnabled()
	analysis.IsValid = ei.cryptoStateReader.IsValid()
	analysis.KeyVersion = uint16(ei.cryptoStateReader.KeyVersion())

	// Protection class information
	protectionClass := ei.cryptoStateReader.ProtectionClass()
	analysis.ProtectionClass = ei.protectionResolver.ResolveName(protectionClass)

	// Additional metadata
	analysis.Metadata["protection_class_description"] = ei.protectionResolver.ResolveDescription(protectionClass)
	analysis.Metadata["effective_class"] = fmt.Sprintf("0x%08X", uint32(ei.protectionResolver.GetEffectiveClass(protectionClass)))
	analysis.Metadata["security_level"] = fmt.Sprintf("%d", ei.protectionResolver.GetSecurityLevel(protectionClass))
	analysis.Metadata["ios_only"] = fmt.Sprintf("%t", ei.protectionResolver.IsiOSOnly(protectionClass))
	analysis.Metadata["macos_only"] = fmt.Sprintf("%t", ei.protectionResolver.IsmacOSOnly(protectionClass))

	// Crypto state details
	analysis.Metadata["reference_count"] = fmt.Sprintf("%d", ei.cryptoStateReader.ReferenceCount())
	analysis.Metadata["major_version"] = fmt.Sprintf("%d", ei.cryptoStateReader.MajorVersion())
	analysis.Metadata["minor_version"] = fmt.Sprintf("%d", ei.cryptoStateReader.MinorVersion())
	analysis.Metadata["key_length"] = fmt.Sprintf("%d", ei.cryptoStateReader.KeyLength())
	analysis.Metadata["object_id"] = fmt.Sprintf("0x%016X", uint64(ei.cryptoStateReader.ObjectID()))
	analysis.Metadata["object_type"] = fmt.Sprintf("0x%08X", ei.cryptoStateReader.ObjectType())

	// OS version information
	osVersion := ei.cryptoStateReader.OSVersion()
	majorVersion, minorLetter, buildNumber := UnpackOsVersion(osVersion)
	analysis.Metadata["os_major_version"] = fmt.Sprintf("%d", majorVersion)
	analysis.Metadata["os_minor_letter"] = fmt.Sprintf("%c", minorLetter)
	analysis.Metadata["os_build_number"] = fmt.Sprintf("%d", buildNumber)
	analysis.Metadata["os_version_raw"] = fmt.Sprintf("0x%08X", uint32(osVersion))

	// Crypto flags
	cryptoFlags := ei.cryptoStateReader.CryptoFlags()
	analysis.Metadata["crypto_flags"] = fmt.Sprintf("0x%08X", uint32(cryptoFlags))

	return analysis, nil
}

// keyRollingManager implements the KeyRollingManager interface
type keyRollingManager struct {
	currentCryptoState  interfaces.CryptoStateReader
	previousCryptoState interfaces.CryptoStateReader
}

// Ensure keyRollingManager implements the KeyRollingManager interface
var _ interfaces.KeyRollingManager = (*keyRollingManager)(nil)

// NewKeyRollingManager creates a new KeyRollingManager
func NewKeyRollingManager(currentState, previousState interfaces.CryptoStateReader) interfaces.KeyRollingManager {
	return &keyRollingManager{
		currentCryptoState:  currentState,
		previousCryptoState: previousState,
	}
}

// IsKeyRollingInProgress checks if a key rotation is currently happening
func (krm *keyRollingManager) IsKeyRollingInProgress() bool {
	if krm.currentCryptoState == nil || krm.previousCryptoState == nil {
		return false
	}

	// Key rolling is in progress if we have different key versions
	return krm.currentCryptoState.KeyVersion() != krm.previousCryptoState.KeyVersion()
}

// GetPreviousKeyVersion returns the version of the previous encryption key
func (krm *keyRollingManager) GetPreviousKeyVersion() types.CpKeyRevisionT {
	if krm.previousCryptoState == nil {
		return 0
	}
	return krm.previousCryptoState.KeyVersion()
}

// GetCurrentKeyVersion returns the version of the current encryption key
func (krm *keyRollingManager) GetCurrentKeyVersion() types.CpKeyRevisionT {
	if krm.currentCryptoState == nil {
		return 0
	}
	return krm.currentCryptoState.KeyVersion()
}
