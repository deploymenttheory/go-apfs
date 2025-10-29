package services

import (
	"testing"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

func TestDeriveContainerKeybagKey(t *testing.T) {
	eh := NewEncryptionHelper()

	// Test with a sample container UUID
	containerUUID := types.UUID{
		0x01, 0x02, 0x03, 0x04,
		0x05, 0x06, 0x07, 0x08,
		0x09, 0x0A, 0x0B, 0x0C,
		0x0D, 0x0E, 0x0F, 0x10,
	}

	key := eh.DeriveContainerKeybagKey(containerUUID)

	// Verify key length
	if len(key) != 32 {
		t.Errorf("Expected key length 32, got %d", len(key))
	}

	// Verify that the key is UUID + UUID
	// First 16 bytes should match UUID
	for i := 0; i < 16; i++ {
		if key[i] != containerUUID[i] {
			t.Errorf("First half of key at byte %d: expected 0x%02X, got 0x%02X", i, containerUUID[i], key[i])
		}
	}

	// Second 16 bytes should also match UUID
	for i := 0; i < 16; i++ {
		if key[16+i] != containerUUID[i] {
			t.Errorf("Second half of key at byte %d: expected 0x%02X, got 0x%02X", i, containerUUID[i], key[16+i])
		}
	}
}

func TestDeriveVolumeKeybagKey(t *testing.T) {
	eh := NewEncryptionHelper()

	// Test with a sample volume UUID
	volumeUUID := types.UUID{
		0xAA, 0xBB, 0xCC, 0xDD,
		0xEE, 0xFF, 0x11, 0x22,
		0x33, 0x44, 0x55, 0x66,
		0x77, 0x88, 0x99, 0x00,
	}

	key := eh.DeriveVolumeKeybagKey(volumeUUID)

	// Verify key length
	if len(key) != 32 {
		t.Errorf("Expected key length 32, got %d", len(key))
	}

	// Verify that the key is UUID + UUID
	// First 16 bytes should match UUID
	for i := 0; i < 16; i++ {
		if key[i] != volumeUUID[i] {
			t.Errorf("First half of key at byte %d: expected 0x%02X, got 0x%02X", i, volumeUUID[i], key[i])
		}
	}

	// Second 16 bytes should also match UUID
	for i := 0; i < 16; i++ {
		if key[16+i] != volumeUUID[i] {
			t.Errorf("Second half of key at byte %d: expected 0x%02X, got 0x%02X", i, volumeUUID[i], key[16+i])
		}
	}
}

func TestCalculateFSTreeNodeTweak(t *testing.T) {
	eh := NewEncryptionHelper()

	tests := []struct {
		name            string
		physicalAddress types.Paddr
		blockSize       uint32
		expectedTweak   uint64
	}{
		{
			name:            "Block at address 0, 4KB blocks",
			physicalAddress: 0,
			blockSize:       4096,
			expectedTweak:   0, // (0 * 4096) / 512 = 0
		},
		{
			name:            "Block at address 1, 4KB blocks",
			physicalAddress: 1,
			blockSize:       4096,
			expectedTweak:   8, // (1 * 4096) / 512 = 8
		},
		{
			name:            "Block at address 100, 4KB blocks",
			physicalAddress: 100,
			blockSize:       4096,
			expectedTweak:   800, // (100 * 4096) / 512 = 800
		},
		{
			name:            "Block at address 1000, 8KB blocks",
			physicalAddress: 1000,
			blockSize:       8192,
			expectedTweak:   16000, // (1000 * 8192) / 512 = 16000
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tweak := eh.CalculateFSTreeNodeTweak(tt.physicalAddress, tt.blockSize)
			if tweak != tt.expectedTweak {
				t.Errorf("CalculateFSTreeNodeTweak() = %d, expected %d", tweak, tt.expectedTweak)
			}
		})
	}
}

func TestCalculateExtentTweak(t *testing.T) {
	eh := NewEncryptionHelper()

	tests := []struct {
		name          string
		cryptoID      uint64
		expectedTweak uint64
	}{
		{
			name:          "Crypto ID 0",
			cryptoID:      0,
			expectedTweak: 0,
		},
		{
			name:          "Crypto ID 12345",
			cryptoID:      12345,
			expectedTweak: 12345,
		},
		{
			name:          "Crypto ID max uint64",
			cryptoID:      ^uint64(0),
			expectedTweak: ^uint64(0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tweak := eh.CalculateExtentTweak(tt.cryptoID)
			if tweak != tt.expectedTweak {
				t.Errorf("CalculateExtentTweak() = %d, expected %d", tweak, tt.expectedTweak)
			}
		})
	}
}

func TestIncrementTweak(t *testing.T) {
	eh := NewEncryptionHelper()

	tests := []struct {
		name          string
		currentTweak  uint64
		expectedTweak uint64
	}{
		{
			name:          "Increment from 0",
			currentTweak:  0,
			expectedTweak: 1,
		},
		{
			name:          "Increment from 100",
			currentTweak:  100,
			expectedTweak: 101,
		},
		{
			name:          "Increment from max-1",
			currentTweak:  ^uint64(0) - 1,
			expectedTweak: ^uint64(0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tweak := eh.IncrementTweak(tt.currentTweak)
			if tweak != tt.expectedTweak {
				t.Errorf("IncrementTweak() = %d, expected %d", tweak, tt.expectedTweak)
			}
		})
	}
}

func TestIsObjectMapValueEncrypted(t *testing.T) {
	eh := NewEncryptionHelper()

	tests := []struct {
		name      string
		flags     uint32
		encrypted bool
	}{
		{
			name:      "No flags set",
			flags:     0x00000000,
			encrypted: false,
		},
		{
			name:      "Encrypted flag set",
			flags:     0x00000001,
			encrypted: true,
		},
		{
			name:      "Other flags set, not encrypted",
			flags:     0x00000002,
			encrypted: false,
		},
		{
			name:      "Encrypted flag and other flags set",
			flags:     0x00000003,
			encrypted: true,
		},
		{
			name:      "Multiple flags including encrypted",
			flags:     0x000000FF,
			encrypted: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encrypted := eh.IsObjectMapValueEncrypted(tt.flags)
			if encrypted != tt.encrypted {
				t.Errorf("IsObjectMapValueEncrypted() = %v, expected %v", encrypted, tt.encrypted)
			}
		})
	}
}

func TestValidateEncryptionKey(t *testing.T) {
	eh := NewEncryptionHelper()

	tests := []struct {
		name      string
		key       []byte
		expectErr bool
	}{
		{
			name:      "Valid 32-byte key",
			key:       make([]byte, 32),
			expectErr: false,
		},
		{
			name:      "Invalid 16-byte key",
			key:       make([]byte, 16),
			expectErr: true,
		},
		{
			name:      "Invalid 64-byte key",
			key:       make([]byte, 64),
			expectErr: true,
		},
		{
			name:      "Invalid empty key",
			key:       []byte{},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := eh.ValidateEncryptionKey(tt.key)
			if (err != nil) != tt.expectErr {
				t.Errorf("ValidateEncryptionKey() error = %v, expectErr %v", err, tt.expectErr)
			}
		})
	}
}

func TestGetBlockSizeForTweak(t *testing.T) {
	eh := NewEncryptionHelper()

	blockSize := eh.GetBlockSizeForTweak()
	if blockSize != 512 {
		t.Errorf("GetBlockSizeForTweak() = %d, expected 512", blockSize)
	}
}

func TestKeybagKeyDerivationConsistency(t *testing.T) {
	eh := NewEncryptionHelper()

	// Test that deriving the same UUID twice produces the same key
	uuid := types.UUID{
		0x12, 0x34, 0x56, 0x78,
		0x9A, 0xBC, 0xDE, 0xF0,
		0x11, 0x22, 0x33, 0x44,
		0x55, 0x66, 0x77, 0x88,
	}

	key1 := eh.DeriveContainerKeybagKey(uuid)
	key2 := eh.DeriveContainerKeybagKey(uuid)

	if key1 != key2 {
		t.Error("Deriving key from same UUID should produce identical results")
	}

	// Test volume keybag key derivation consistency
	volKey1 := eh.DeriveVolumeKeybagKey(uuid)
	volKey2 := eh.DeriveVolumeKeybagKey(uuid)

	if volKey1 != volKey2 {
		t.Error("Deriving volume key from same UUID should produce identical results")
	}
}

func TestTweakCalculationForMultipleBlocks(t *testing.T) {
	eh := NewEncryptionHelper()

	// Test calculating tweaks for multiple 512-byte blocks in a 4KB APFS block
	physicalAddress := types.Paddr(100)
	blockSize := uint32(4096)

	initialTweak := eh.CalculateFSTreeNodeTweak(physicalAddress, blockSize)

	// A 4KB block contains 8 x 512-byte blocks
	// Verify the tweak progression
	expectedTweaks := []uint64{
		initialTweak,
		initialTweak + 1,
		initialTweak + 2,
		initialTweak + 3,
		initialTweak + 4,
		initialTweak + 5,
		initialTweak + 6,
		initialTweak + 7,
	}

	currentTweak := initialTweak
	for i, expected := range expectedTweaks {
		if currentTweak != expected {
			t.Errorf("Block %d: tweak = %d, expected %d", i, currentTweak, expected)
		}
		if i < len(expectedTweaks)-1 {
			currentTweak = eh.IncrementTweak(currentTweak)
		}
	}
}
