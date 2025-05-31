// Package encryption provides low-level APFS encryption parsing functionality.
// This package implements read-only parsing of APFS encryption structures including
// keybags, crypto states, and protection classes.
package encryption

import (
	"encoding/binary"

	"github.com/deploymenttheory/go-apfs/internal/interfaces"
)

// Low-level constructors for parsing APFS encryption structures

// NewKeybagReaderFromData creates a new KeybagReader from raw keybag data.
func NewKeybagReaderFromData(data []byte, endian binary.ByteOrder) (interfaces.KeybagReader, error) {
	return NewKeybagReader(data, endian)
}

// NewMediaKeybagReaderFromData creates a new KeybagReader from raw media keybag data.
func NewMediaKeybagReaderFromData(data []byte, endian binary.ByteOrder) (interfaces.KeybagReader, error) {
	return NewMediaKeybagReader(data, endian)
}

// NewCryptoStateReaderFromData creates a new CryptoStateReader from raw key and value data.
func NewCryptoStateReaderFromData(keyData, valueData []byte, endian binary.ByteOrder) (interfaces.CryptoStateReader, error) {
	return NewCryptoStateReader(keyData, valueData, endian)
}

// NewProtectionClassResolverInstance creates a new ProtectionClassResolver.
func NewProtectionClassResolverInstance() interfaces.ProtectionClassResolver {
	return NewProtectionClassResolver()
}
