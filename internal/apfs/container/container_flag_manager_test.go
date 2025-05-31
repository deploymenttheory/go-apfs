package container

import (
	"testing"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

// createTestSuperblockForFlags creates a test superblock with specific flags
func createTestSuperblockForFlags(flags uint64) *types.NxSuperblockT {
	return &types.NxSuperblockT{
		NxFlags: flags,
	}
}

func TestContainerFlagManager(t *testing.T) {
	tests := []struct {
		name             string
		flags            uint64
		expectSoftCrypto bool
	}{
		{
			name:             "No flags",
			flags:            0,
			expectSoftCrypto: false,
		},
		{
			name:             "Software crypto enabled",
			flags:            types.NxCryptoSw,
			expectSoftCrypto: true,
		},
		{
			name:             "Multiple flags including crypto",
			flags:            types.NxCryptoSw | types.NxReserved1,
			expectSoftCrypto: true,
		},
		{
			name:             "Reserved flags only",
			flags:            types.NxReserved1 | types.NxReserved2,
			expectSoftCrypto: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			superblock := createTestSuperblockForFlags(tt.flags)
			manager := NewContainerFlagManager(superblock)

			// Test basic flag getter
			if flags := manager.Flags(); flags != tt.flags {
				t.Errorf("Flags() = 0x%X, want 0x%X", flags, tt.flags)
			}

			// Test software cryptography detection
			if softCrypto := manager.UsesSoftwareCryptography(); softCrypto != tt.expectSoftCrypto {
				t.Errorf("UsesSoftwareCryptography() = %v, want %v", softCrypto, tt.expectSoftCrypto)
			}
		})
	}
}

func TestContainerFlagManager_HasFlag(t *testing.T) {
	flags := types.NxCryptoSw | types.NxReserved1
	superblock := createTestSuperblockForFlags(flags)
	manager := NewContainerFlagManager(superblock)

	tests := []struct {
		name     string
		flag     uint64
		expected bool
	}{
		{
			name:     "Has crypto software flag",
			flag:     types.NxCryptoSw,
			expected: true,
		},
		{
			name:     "Has reserved 1 flag",
			flag:     types.NxReserved1,
			expected: true,
		},
		{
			name:     "Does not have reserved 2 flag",
			flag:     types.NxReserved2,
			expected: false,
		},
		{
			name:     "Custom flag not set",
			flag:     0x1000,
			expected: false,
		},
		{
			name:     "Multiple flags check",
			flag:     types.NxCryptoSw | types.NxReserved1,
			expected: true,
		},
		{
			name:     "Partial multiple flags check",
			flag:     types.NxCryptoSw | types.NxReserved2,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result := manager.HasFlag(tt.flag); result != tt.expected {
				t.Errorf("HasFlag(0x%X) = %v, want %v", tt.flag, result, tt.expected)
			}
		})
	}
}

func TestContainerFlagManager_EdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		flags uint64
	}{
		{
			name:  "Zero flags",
			flags: 0,
		},
		{
			name:  "Maximum flags",
			flags: 0xFFFFFFFFFFFFFFFF,
		},
		{
			name:  "Single bit flag",
			flags: 0x1,
		},
		{
			name:  "High bit flag",
			flags: 0x8000000000000000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			superblock := createTestSuperblockForFlags(tt.flags)
			manager := NewContainerFlagManager(superblock)

			// Test that flags are preserved correctly
			if flags := manager.Flags(); flags != tt.flags {
				t.Errorf("Flags() = 0x%X, want 0x%X", flags, tt.flags)
			}

			// Test HasFlag with the exact flags
			if result := manager.HasFlag(tt.flags); tt.flags != 0 && !result {
				t.Errorf("HasFlag(0x%X) = %v, want true", tt.flags, result)
			}
		})
	}
}
