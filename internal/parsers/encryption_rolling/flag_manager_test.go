package encryptionrolling

import (
	"testing"

	"github.com/deploymenttheory/go-apfs/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestFlagManager_Methods(t *testing.T) {
	tests := []struct {
		name     string
		flags    uint64
		expected map[string]bool
	}{
		{
			name:  "encrypting flag",
			flags: types.ErsbFlagEncrypting,
			expected: map[string]bool{
				"IsEncrypting": true,
				"IsDecrypting": false,
				"IsKeyRolling": false,
				"IsPaused":     false,
				"HasFailed":    false,
				"IsCIDTweak":   false,
				"IsFromOneKey": false,
			},
		},
		{
			name:  "decrypting flag",
			flags: types.ErsbFlagDecrypting,
			expected: map[string]bool{
				"IsEncrypting": false,
				"IsDecrypting": true,
				"IsKeyRolling": false,
				"IsPaused":     false,
				"HasFailed":    false,
				"IsCIDTweak":   false,
				"IsFromOneKey": false,
			},
		},
		{
			name:  "key rolling flag",
			flags: types.ErsbFlagKeyrolling,
			expected: map[string]bool{
				"IsEncrypting": false,
				"IsDecrypting": false,
				"IsKeyRolling": true,
				"IsPaused":     false,
				"HasFailed":    false,
				"IsCIDTweak":   false,
				"IsFromOneKey": false,
			},
		},
		{
			name:  "paused flag",
			flags: types.ErsbFlagPaused,
			expected: map[string]bool{
				"IsEncrypting": false,
				"IsDecrypting": false,
				"IsKeyRolling": false,
				"IsPaused":     true,
				"HasFailed":    false,
				"IsCIDTweak":   false,
				"IsFromOneKey": false,
			},
		},
		{
			name:  "failed flag",
			flags: types.ErsbFlagFailed,
			expected: map[string]bool{
				"IsEncrypting": false,
				"IsDecrypting": false,
				"IsKeyRolling": false,
				"IsPaused":     false,
				"HasFailed":    true,
				"IsCIDTweak":   false,
				"IsFromOneKey": false,
			},
		},
		{
			name:  "CID tweak flag",
			flags: types.ErsbFlagCidIsTweak,
			expected: map[string]bool{
				"IsEncrypting": false,
				"IsDecrypting": false,
				"IsKeyRolling": false,
				"IsPaused":     false,
				"HasFailed":    false,
				"IsCIDTweak":   true,
				"IsFromOneKey": false,
			},
		},
		{
			name:  "from one key flag",
			flags: types.ErsbFlagFromOnekey,
			expected: map[string]bool{
				"IsEncrypting": false,
				"IsDecrypting": false,
				"IsKeyRolling": false,
				"IsPaused":     false,
				"HasFailed":    false,
				"IsCIDTweak":   false,
				"IsFromOneKey": true,
			},
		},
		{
			name:  "multiple flags",
			flags: types.ErsbFlagEncrypting | types.ErsbFlagPaused | types.ErsbFlagCidIsTweak,
			expected: map[string]bool{
				"IsEncrypting": true,
				"IsDecrypting": false,
				"IsKeyRolling": false,
				"IsPaused":     true,
				"HasFailed":    false,
				"IsCIDTweak":   true,
				"IsFromOneKey": false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewFlagManager(tt.flags)

			assert.Equal(t, tt.expected["IsEncrypting"], manager.IsEncrypting())
			assert.Equal(t, tt.expected["IsDecrypting"], manager.IsDecrypting())
			assert.Equal(t, tt.expected["IsKeyRolling"], manager.IsKeyRolling())
			assert.Equal(t, tt.expected["IsPaused"], manager.IsPaused())
			assert.Equal(t, tt.expected["HasFailed"], manager.HasFailed())
			assert.Equal(t, tt.expected["IsCIDTweak"], manager.IsCIDTweak())
			assert.Equal(t, tt.expected["IsFromOneKey"], manager.IsFromOneKey())
		})
	}
}

func TestFlagManager_GetBlockSize(t *testing.T) {
	tests := []struct {
		name          string
		blockSizeCode uint32
		expectedSize  uint64
	}{
		{"512B block size", types.Er512bBlocksize, 512},
		{"2KiB block size", types.Er2kibBlocksize, 2048},
		{"4KiB block size", types.Er4kibBlocksize, 4096},
		{"8KiB block size", types.Er8kibBlocksize, 8192},
		{"16KiB block size", types.Er16kibBlocksize, 16384},
		{"32KiB block size", types.Er32kibBlocksize, 32768},
		{"64KiB block size", types.Er64kibBlocksize, 65536},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create flags with the block size code
			flags := (uint64(tt.blockSizeCode) << types.ErsbFlagCmBlockSizeShift) & types.ErsbFlagCmBlockSizeMask
			manager := NewFlagManager(flags)

			assert.Equal(t, tt.expectedSize, manager.GetBlockSize())
		})
	}
}

func TestFlagManager_GetPhase(t *testing.T) {
	tests := []struct {
		name          string
		phaseCode     types.ErPhaseT
		expectedPhase types.ErPhaseT
	}{
		{"OMAP roll phase", types.ErPhaseOmapRoll, types.ErPhaseOmapRoll},
		{"Data roll phase", types.ErPhaseDataRoll, types.ErPhaseDataRoll},
		{"Snapshot roll phase", types.ErPhaseSnapRoll, types.ErPhaseSnapRoll},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create flags with the phase code
			flags := (uint64(tt.phaseCode) << types.ErsbFlagErPhaseShift) & types.ErsbFlagErPhaseMask
			manager := NewFlagManager(flags)

			assert.Equal(t, tt.expectedPhase, manager.GetPhase())
		})
	}
}

func TestPhaseManager_Methods(t *testing.T) {
	tests := []struct {
		name                string
		phaseCode           types.ErPhaseT
		expectedPhase       types.ErPhaseT
		expectedDescription string
		isOmap              bool
		isData              bool
		isSnapshot          bool
	}{
		{
			name:                "OMAP roll phase",
			phaseCode:           types.ErPhaseOmapRoll,
			expectedPhase:       types.ErPhaseOmapRoll,
			expectedDescription: "Object Map Rolling",
			isOmap:              true,
			isData:              false,
			isSnapshot:          false,
		},
		{
			name:                "Data roll phase",
			phaseCode:           types.ErPhaseDataRoll,
			expectedPhase:       types.ErPhaseDataRoll,
			expectedDescription: "Data Rolling",
			isOmap:              false,
			isData:              true,
			isSnapshot:          false,
		},
		{
			name:                "Snapshot roll phase",
			phaseCode:           types.ErPhaseSnapRoll,
			expectedPhase:       types.ErPhaseSnapRoll,
			expectedDescription: "Snapshot Rolling",
			isOmap:              false,
			isData:              false,
			isSnapshot:          true,
		},
		{
			name:                "Unknown phase",
			phaseCode:           types.ErPhaseT(99),
			expectedPhase:       types.ErPhaseT(3), // 99 & 0x3 = 3 (ErPhaseSnapRoll)
			expectedDescription: "Snapshot Rolling",
			isOmap:              false,
			isData:              false,
			isSnapshot:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create flags with the phase code
			flags := (uint64(tt.phaseCode) << types.ErsbFlagErPhaseShift) & types.ErsbFlagErPhaseMask
			manager := NewPhaseManager(flags)

			assert.Equal(t, tt.expectedPhase, manager.GetCurrentPhase())
			assert.Equal(t, tt.expectedDescription, manager.GetPhaseDescription())
			assert.Equal(t, tt.isOmap, manager.IsOmapRollPhase())
			assert.Equal(t, tt.isData, manager.IsDataRollPhase())
			assert.Equal(t, tt.isSnapshot, manager.IsSnapshotRollPhase())
		})
	}

	// Test a truly unknown phase by directly setting flags
	t.Run("Truly unknown phase", func(t *testing.T) {
		// Set phase bits to 0 (which is outside valid range 1-3)
		flags := uint64(0) << types.ErsbFlagErPhaseShift
		manager := NewPhaseManager(flags)

		assert.Equal(t, types.ErPhaseT(0), manager.GetCurrentPhase())
		assert.Equal(t, "Unknown Phase", manager.GetPhaseDescription())
		assert.False(t, manager.IsOmapRollPhase())
		assert.False(t, manager.IsDataRollPhase())
		assert.False(t, manager.IsSnapshotRollPhase())
	})
}

func TestBlockSizeResolver_GetBlockSizeValue(t *testing.T) {
	tests := []struct {
		name         string
		constant     uint32
		expectedSize uint32
	}{
		{"512B", types.Er512bBlocksize, 512},
		{"2KiB", types.Er2kibBlocksize, 2048},
		{"4KiB", types.Er4kibBlocksize, 4096},
		{"8KiB", types.Er8kibBlocksize, 8192},
		{"16KiB", types.Er16kibBlocksize, 16384},
		{"32KiB", types.Er32kibBlocksize, 32768},
		{"64KiB", types.Er64kibBlocksize, 65536},
		{"Invalid", 99, 0},
	}

	resolver := NewBlockSizeResolver()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			size := resolver.GetBlockSizeValue(tt.constant)
			assert.Equal(t, tt.expectedSize, size)
		})
	}
}

func TestBlockSizeResolver_GetBlockSizeConstant(t *testing.T) {
	tests := []struct {
		name             string
		sizeInBytes      uint32
		expectedConstant uint32
	}{
		{"512B", 512, types.Er512bBlocksize},
		{"2KiB", 2048, types.Er2kibBlocksize},
		{"4KiB", 4096, types.Er4kibBlocksize},
		{"8KiB", 8192, types.Er8kibBlocksize},
		{"16KiB", 16384, types.Er16kibBlocksize},
		{"32KiB", 32768, types.Er32kibBlocksize},
		{"64KiB", 65536, types.Er64kibBlocksize},
		{"Invalid", 1024, 0xFFFFFFFF},
	}

	resolver := NewBlockSizeResolver()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constant := resolver.GetBlockSizeConstant(tt.sizeInBytes)
			assert.Equal(t, tt.expectedConstant, constant)
		})
	}
}

func TestBlockSizeResolver_GetSupportedBlockSizes(t *testing.T) {
	resolver := NewBlockSizeResolver()
	sizes := resolver.GetSupportedBlockSizes()

	expectedSizes := []uint32{512, 2048, 4096, 8192, 16384, 32768, 65536}
	assert.Equal(t, expectedSizes, sizes)
}

func TestNewFlagManager(t *testing.T) {
	flags := uint64(0x12345678)
	manager := NewFlagManager(flags)
	assert.NotNil(t, manager)
}

func TestNewPhaseManager(t *testing.T) {
	flags := uint64(0x12345678)
	manager := NewPhaseManager(flags)
	assert.NotNil(t, manager)
}

func TestNewBlockSizeResolver(t *testing.T) {
	resolver := NewBlockSizeResolver()
	assert.NotNil(t, resolver)
}
