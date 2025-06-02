package volumes

import (
	"fmt"
	"testing"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

// Helper function to create a test superblock
func createTestSuperblock() *types.ApfsSuperblockT {
	return &types.ApfsSuperblockT{
		ApfsVolUuid: [16]byte{
			0x01, 0x23, 0x45, 0x67, 0x89, 0xAB, 0xCD, 0xEF,
			0x01, 0x23, 0x45, 0x67, 0x89, 0xAB, 0xCD, 0xEF,
		},
		ApfsVolname: [256]byte{
			'T', 'e', 's', 't', ' ', 'V', 'o', 'l', 'u', 'm', 'e', 0,
		},
		ApfsRole:    types.ApfsVolRoleUser,
		ApfsFsIndex: 42,
	}
}

// Test UUID extraction
func TestVolumeIdentity_UUID(t *testing.T) {
	sb := createTestSuperblock()
	vi := NewVolumeIdentity(sb)

	uuid := vi.UUID()
	expectedUUID := sb.ApfsVolUuid

	for i := range uuid {
		if uuid[i] != expectedUUID[i] {
			t.Errorf("UUID mismatch at index %d: got %x, want %x", i, uuid[i], expectedUUID[i])
		}
	}
}

// Test Volume Name Extraction
func TestVolumeIdentity_Name(t *testing.T) {
	testCases := []struct {
		name     string
		volname  [256]byte
		expected string
	}{
		{
			name: "Normal Name",
			volname: func() [256]byte {
				var n [256]byte
				copy(n[:], "Test Volume")
				return n
			}(),
			expected: "Test Volume",
		},
		{
			name: "Name with Null Terminator",
			volname: func() [256]byte {
				var n [256]byte
				copy(n[:], "Valid Name\x00\x00\x00")
				return n
			}(),
			expected: "Valid Name",
		},
		{
			name: "Name with Trailing Spaces",
			volname: func() [256]byte {
				var n [256]byte
				copy(n[:], "Spaced Name   \x00")
				return n
			}(),
			expected: "Spaced Name",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sb := createTestSuperblock()
			sb.ApfsVolname = tc.volname

			vi := NewVolumeIdentity(sb)

			if got := vi.Name(); got != tc.expected {
				t.Errorf("Name() = %q, want %q", got, tc.expected)
			}
		})
	}
}

// Test Role Extraction
func TestVolumeIdentity_Role(t *testing.T) {
	testCases := []struct {
		role     uint16
		expected uint16
	}{
		{types.ApfsVolRoleNone, types.ApfsVolRoleNone},
		{types.ApfsVolRoleSystem, types.ApfsVolRoleSystem},
		{types.ApfsVolRoleUser, types.ApfsVolRoleUser},
		{types.ApfsVolRoleRecovery, types.ApfsVolRoleRecovery},
		{types.ApfsVolRoleVm, types.ApfsVolRoleVm},
		{types.ApfsVolRolePreboot, types.ApfsVolRolePreboot},
		{types.ApfsVolRoleInstaller, types.ApfsVolRoleInstaller},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Role %d", tc.role), func(t *testing.T) {
			sb := createTestSuperblock()
			sb.ApfsRole = tc.role

			vi := NewVolumeIdentity(sb)

			if got := vi.Role(); got != tc.expected {
				t.Errorf("Role() = %d, want %d", got, tc.expected)
			}
		})

	}
}

// Test Role Name Conversion
func TestVolumeIdentity_RoleName(t *testing.T) {
	testCases := []struct {
		role     uint16
		expected string
	}{
		{types.ApfsVolRoleNone, "None"},
		{types.ApfsVolRoleSystem, "System"},
		{types.ApfsVolRoleUser, "User"},
		{types.ApfsVolRoleRecovery, "Recovery"},
		{types.ApfsVolRoleVm, "Virtual Memory"},
		{types.ApfsVolRolePreboot, "Preboot"},
		{types.ApfsVolRoleInstaller, "Installer"},
		{types.ApfsVolRoleData, "Data"},
		{types.ApfsVolRoleBaseband, "Baseband"},
		{types.ApfsVolRoleUpdate, "Update"},
		{types.ApfsVolRoleXart, "XART (Secure User Data)"},
		{types.ApfsVolRoleHardware, "Hardware"},
		{types.ApfsVolRoleBackup, "Backup"},
		{types.ApfsVolRoleEnterprise, "Enterprise"},
		{types.ApfsVolRolePrelogin, "Prelogin"},
		{0xFFFF, "Unknown"},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			sb := createTestSuperblock()
			sb.ApfsRole = tc.role

			vi := NewVolumeIdentity(sb)

			if got := vi.RoleName(); got != tc.expected {
				t.Errorf("RoleName() = %q, want %q", got, tc.expected)
			}
		})
	}
}

// Test Index Extraction
func TestVolumeIdentity_Index(t *testing.T) {
	sb := createTestSuperblock()
	vi := NewVolumeIdentity(sb)

	if got := vi.Index(); got != sb.ApfsFsIndex {
		t.Errorf("Index() = %d, want %d", got, sb.ApfsFsIndex)
	}
}

// Benchmark UUID extraction
func BenchmarkVolumeIdentity_UUID(b *testing.B) {
	sb := createTestSuperblock()
	vi := NewVolumeIdentity(sb)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = vi.UUID()
	}
}

// Benchmark Name extraction
func BenchmarkVolumeIdentity_Name(b *testing.B) {
	sb := createTestSuperblock()
	vi := NewVolumeIdentity(sb)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = vi.Name()
	}
}
