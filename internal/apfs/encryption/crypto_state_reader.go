package encryption

import (
	"encoding/binary"
	"fmt"

	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// cryptoStateReader implements the CryptoStateReader interface
type cryptoStateReader struct {
	cryptoKey   *types.JCryptoKeyT
	cryptoValue *types.JCryptoValT
	endian      binary.ByteOrder
}

// Ensure cryptoStateReader implements the CryptoStateReader interface
var _ interfaces.CryptoStateReader = (*cryptoStateReader)(nil)

// NewCryptoStateReader creates a new CryptoStateReader from raw key and value data
func NewCryptoStateReader(keyData, valueData []byte, endian binary.ByteOrder) (interfaces.CryptoStateReader, error) {
	if endian == nil {
		endian = binary.LittleEndian
	}

	// Parse the crypto key
	cryptoKey, err := parseCryptoKey(keyData, endian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse crypto key: %w", err)
	}

	// Parse the crypto value
	cryptoValue, err := parseCryptoValue(valueData, endian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse crypto value: %w", err)
	}

	return &cryptoStateReader{
		cryptoKey:   cryptoKey,
		cryptoValue: cryptoValue,
		endian:      endian,
	}, nil
}

// parseCryptoKey parses raw bytes into a JCryptoKeyT structure
func parseCryptoKey(data []byte, endian binary.ByteOrder) (*types.JCryptoKeyT, error) {
	if len(data) < 8 { // JKeyT is 8 bytes
		return nil, fmt.Errorf("insufficient data for crypto key: need at least 8 bytes, got %d", len(data))
	}

	key := &types.JCryptoKeyT{}
	key.Hdr.ObjIdAndType = endian.Uint64(data[0:8])

	return key, nil
}

// parseCryptoValue parses raw bytes into a JCryptoValT structure
func parseCryptoValue(data []byte, endian binary.ByteOrder) (*types.JCryptoValT, error) {
	if len(data) < 4 { // At minimum need refcnt (4 bytes)
		return nil, fmt.Errorf("insufficient data for crypto value: need at least 4 bytes, got %d", len(data))
	}

	value := &types.JCryptoValT{}

	// Parse reference count
	value.Refcnt = endian.Uint32(data[0:4])

	// Parse the wrapped crypto state if we have enough data
	// Minimum wrapped state size: versions(4) + flags(4) + class(4) + os_version(4) + revision(2) + keylen(2) = 20 bytes
	if len(data) >= 4+20 { // refcnt + wrapped state minimum
		state, err := parseWrappedCryptoState(data[4:], endian)
		if err != nil {
			return nil, fmt.Errorf("failed to parse wrapped crypto state: %w", err)
		}
		value.State = *state
	}

	return value, nil
}

// parseWrappedCryptoState parses raw bytes into a WrappedCryptoStateT structure
func parseWrappedCryptoState(data []byte, endian binary.ByteOrder) (*types.WrappedCryptoStateT, error) {
	// Minimum size: versions(4) + flags(4) + class(4) + os_version(4) + revision(2) + keylen(2) = 20 bytes
	if len(data) < 20 {
		return nil, fmt.Errorf("insufficient data for wrapped crypto state: need at least 20 bytes, got %d", len(data))
	}

	state := &types.WrappedCryptoStateT{}
	offset := 0

	// Parse versions
	state.MajorVersion = endian.Uint16(data[offset : offset+2])
	offset += 2
	state.MinorVersion = endian.Uint16(data[offset : offset+2])
	offset += 2

	// Parse flags and class
	state.Cpflags = types.CryptoFlagsT(endian.Uint32(data[offset : offset+4]))
	offset += 4
	state.PersistentClass = types.CpKeyClassT(endian.Uint32(data[offset : offset+4]))
	offset += 4

	// Parse OS version and key revision
	state.KeyOsVersion = types.CpKeyOsVersionT(endian.Uint32(data[offset : offset+4]))
	offset += 4
	state.KeyRevision = types.CpKeyRevisionT(endian.Uint16(data[offset : offset+2]))
	offset += 2

	// Parse key length
	state.KeyLen = endian.Uint16(data[offset : offset+2])
	offset += 2

	// Validate key length
	if state.KeyLen > types.CpMaxWrappedkeysize {
		return nil, fmt.Errorf("key length %d exceeds maximum %d", state.KeyLen, types.CpMaxWrappedkeysize)
	}

	// Parse wrapped key data if available
	if len(data) >= offset+int(state.KeyLen) {
		copy(state.PersistentKey[:state.KeyLen], data[offset:offset+int(state.KeyLen)])
	}

	return state, nil
}

// ReferenceCount returns the reference count for the encryption state
func (csr *cryptoStateReader) ReferenceCount() uint32 {
	return csr.cryptoValue.Refcnt
}

// ProtectionClass returns the protection class of the encryption state
func (csr *cryptoStateReader) ProtectionClass() types.CpKeyClassT {
	return csr.cryptoValue.State.PersistentClass
}

// KeyVersion returns the version of the encryption key
func (csr *cryptoStateReader) KeyVersion() types.CpKeyRevisionT {
	return csr.cryptoValue.State.KeyRevision
}

// IsValid checks if the encryption state is valid
func (csr *cryptoStateReader) IsValid() bool {
	// Check if we have valid reference count and key length
	if csr.cryptoValue.Refcnt == 0 {
		return false
	}

	// Check if the protection class is valid
	protectionClass := csr.cryptoValue.State.PersistentClass & types.CpEffectiveClassmask
	switch protectionClass {
	case types.ProtectionClassDirNone, types.ProtectionClassA, types.ProtectionClassB,
		types.ProtectionClassC, types.ProtectionClassD, types.ProtectionClassF, types.ProtectionClassM:
		return true
	default:
		return false
	}
}

// MajorVersion returns the major version of the wrapped crypto state
func (csr *cryptoStateReader) MajorVersion() uint16 {
	return csr.cryptoValue.State.MajorVersion
}

// MinorVersion returns the minor version of the wrapped crypto state
func (csr *cryptoStateReader) MinorVersion() uint16 {
	return csr.cryptoValue.State.MinorVersion
}

// KeyLength returns the length of the wrapped key data
func (csr *cryptoStateReader) KeyLength() uint16 {
	return csr.cryptoValue.State.KeyLen
}

// WrappedKeyData returns the encrypted key data
func (csr *cryptoStateReader) WrappedKeyData() []byte {
	keyLen := csr.cryptoValue.State.KeyLen
	if keyLen == 0 || keyLen > types.CpMaxWrappedkeysize {
		return nil
	}

	result := make([]byte, keyLen)
	copy(result, csr.cryptoValue.State.PersistentKey[:keyLen])
	return result
}

// CryptoFlags returns the encryption flags
func (csr *cryptoStateReader) CryptoFlags() types.CryptoFlagsT {
	return csr.cryptoValue.State.Cpflags
}

// OSVersion returns the OS version that created this encryption state
func (csr *cryptoStateReader) OSVersion() types.CpKeyOsVersionT {
	return csr.cryptoValue.State.KeyOsVersion
}

// ObjectID returns the object identifier from the crypto key header
func (csr *cryptoStateReader) ObjectID() types.OidT {
	return types.OidT(csr.cryptoKey.Hdr.ObjIdAndType & types.ObjIdMask)
}

// ObjectType returns the object type from the crypto key header
func (csr *cryptoStateReader) ObjectType() uint32 {
	// For 64-bit ObjIdAndType, the type field is 4 bits (60-63), not 16 bits
	objTypeFieldMask := uint64(0xF) << types.ObjTypeShift
	return uint32((csr.cryptoKey.Hdr.ObjIdAndType & objTypeFieldMask) >> types.ObjTypeShift)
}
