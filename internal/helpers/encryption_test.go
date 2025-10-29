package helpers

import (
	"testing"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

// PackOsVersion creates a CpKeyOsVersionT from major version, minor letter, and build number
func PackOsVersion(majorVersion uint16, minorLetter byte, buildNumber uint32) types.CpKeyOsVersionT {
	return types.CpKeyOsVersionT(uint32(majorVersion)<<24 | uint32(minorLetter)<<16 | (buildNumber & 0xFFFF))
}

// UnpackOsVersion extracts the components from a CpKeyOsVersionT
func UnpackOsVersion(version types.CpKeyOsVersionT) (majorVersion uint16, minorLetter byte, buildNumber uint32) {
	majorVersion = uint16((version >> 24) & 0xFF)
	minorLetter = byte((version >> 16) & 0xFF)
	buildNumber = uint32(version & 0xFFFF)
	return
}

func TestPackOsVersion(t *testing.T) {
	tests := []struct {
		name         string
		majorVersion uint16
		minorLetter  byte
		buildNumber  uint32
		expected     types.CpKeyOsVersionT
	}{
		{
			name:         "macOS 10.15 build 19A583",
			majorVersion: 10,
			minorLetter:  15,     // 'O' for Catalina, but using numeric for test
			buildNumber:  0x4A83, // 19A583 simplified
			expected:     types.CpKeyOsVersionT(0x0A0F4A83),
		},
		{
			name:         "macOS 11.0 build 20A5384c",
			majorVersion: 11,
			minorLetter:  0, // Big Sur
			buildNumber:  0x5384,
			expected:     types.CpKeyOsVersionT(0x0B005384),
		},
		{
			name:         "iOS 14.0 build 18A373",
			majorVersion: 14,
			minorLetter:  0,
			buildNumber:  0xA373,
			expected:     types.CpKeyOsVersionT(0x0E00A373),
		},
		{
			name:         "Maximum values",
			majorVersion: 255,
			minorLetter:  255,
			buildNumber:  0xFFFF,
			expected:     types.CpKeyOsVersionT(0xFFFFFFFF),
		},
		{
			name:         "Minimum values",
			majorVersion: 0,
			minorLetter:  0,
			buildNumber:  0,
			expected:     types.CpKeyOsVersionT(0x00000000),
		},
		{
			name:         "Build number overflow",
			majorVersion: 12,
			minorLetter:  5,
			buildNumber:  0x12345678, // Should be masked to 0x5678
			expected:     types.CpKeyOsVersionT(0x0C055678),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := PackOsVersion(tc.majorVersion, tc.minorLetter, tc.buildNumber)
			if result != tc.expected {
				t.Errorf("PackOsVersion(%d, %d, 0x%08X) = 0x%08X, want 0x%08X",
					tc.majorVersion, tc.minorLetter, tc.buildNumber, uint32(result), uint32(tc.expected))
			}
		})
	}
}

func TestUnpackOsVersion(t *testing.T) {
	tests := []struct {
		name                 string
		version              types.CpKeyOsVersionT
		expectedMajorVersion uint16
		expectedMinorLetter  byte
		expectedBuildNumber  uint32
	}{
		{
			name:                 "macOS 10.15 build 19A583",
			version:              types.CpKeyOsVersionT(0x0A0F4A83),
			expectedMajorVersion: 10,
			expectedMinorLetter:  15,
			expectedBuildNumber:  0x4A83,
		},
		{
			name:                 "macOS 11.0 build 20A5384c",
			version:              types.CpKeyOsVersionT(0x0B005384),
			expectedMajorVersion: 11,
			expectedMinorLetter:  0,
			expectedBuildNumber:  0x5384,
		},
		{
			name:                 "iOS 14.0 build 18A373",
			version:              types.CpKeyOsVersionT(0x0E00A373),
			expectedMajorVersion: 14,
			expectedMinorLetter:  0,
			expectedBuildNumber:  0xA373,
		},
		{
			name:                 "Maximum values",
			version:              types.CpKeyOsVersionT(0xFFFFFFFF),
			expectedMajorVersion: 255,
			expectedMinorLetter:  255,
			expectedBuildNumber:  0xFFFF,
		},
		{
			name:                 "Minimum values",
			version:              types.CpKeyOsVersionT(0x00000000),
			expectedMajorVersion: 0,
			expectedMinorLetter:  0,
			expectedBuildNumber:  0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			majorVersion, minorLetter, buildNumber := UnpackOsVersion(tc.version)

			if majorVersion != tc.expectedMajorVersion {
				t.Errorf("UnpackOsVersion(0x%08X) majorVersion = %d, want %d",
					uint32(tc.version), majorVersion, tc.expectedMajorVersion)
			}

			if minorLetter != tc.expectedMinorLetter {
				t.Errorf("UnpackOsVersion(0x%08X) minorLetter = %d, want %d",
					uint32(tc.version), minorLetter, tc.expectedMinorLetter)
			}

			if buildNumber != tc.expectedBuildNumber {
				t.Errorf("UnpackOsVersion(0x%08X) buildNumber = 0x%08X, want 0x%08X",
					uint32(tc.version), buildNumber, tc.expectedBuildNumber)
			}
		})
	}
}

func TestPackUnpackOsVersionRoundTrip(t *testing.T) {
	tests := []struct {
		name         string
		majorVersion uint16
		minorLetter  byte
		buildNumber  uint32
	}{
		{"Standard case", 10, 15, 0x4A83},
		{"Zero values", 0, 0, 0},
		{"Maximum major", 255, 0, 0},
		{"Maximum minor", 0, 255, 0},
		{"Maximum build", 0, 0, 0xFFFF},
		{"All maximum", 255, 255, 0xFFFF},
		{"Build overflow", 12, 5, 0x12345678}, // Should be truncated
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Pack the version
			packed := PackOsVersion(tc.majorVersion, tc.minorLetter, tc.buildNumber)

			// Unpack it back
			unpackedMajor, unpackedMinor, unpackedBuild := UnpackOsVersion(packed)

			// For build number, we expect truncation to 16 bits
			expectedBuild := tc.buildNumber & 0xFFFF

			if unpackedMajor != tc.majorVersion {
				t.Errorf("Round trip failed for majorVersion: got %d, want %d", unpackedMajor, tc.majorVersion)
			}

			if unpackedMinor != tc.minorLetter {
				t.Errorf("Round trip failed for minorLetter: got %d, want %d", unpackedMinor, tc.minorLetter)
			}

			if unpackedBuild != expectedBuild {
				t.Errorf("Round trip failed for buildNumber: got 0x%08X, want 0x%08X", unpackedBuild, expectedBuild)
			}
		})
	}
}

func TestOsVersionBitLayout(t *testing.T) {
	// Test that the bit layout is correct
	majorVersion := uint16(0xAB)
	minorLetter := byte(0xCD)
	buildNumber := uint32(0xEF12)

	packed := PackOsVersion(majorVersion, minorLetter, buildNumber)
	expected := types.CpKeyOsVersionT(0xABCDEF12)

	if packed != expected {
		t.Errorf("Bit layout incorrect: got 0x%08X, want 0x%08X", uint32(packed), uint32(expected))
	}

	// Verify each component is in the right position
	if (uint32(packed)>>24)&0xFF != uint32(majorVersion) {
		t.Errorf("Major version not in bits 24-31: got 0x%02X, want 0x%02X",
			(uint32(packed)>>24)&0xFF, majorVersion)
	}

	if (uint32(packed)>>16)&0xFF != uint32(minorLetter) {
		t.Errorf("Minor letter not in bits 16-23: got 0x%02X, want 0x%02X",
			(uint32(packed)>>16)&0xFF, minorLetter)
	}

	if uint32(packed)&0xFFFF != buildNumber&0xFFFF {
		t.Errorf("Build number not in bits 0-15: got 0x%04X, want 0x%04X",
			uint32(packed)&0xFFFF, buildNumber&0xFFFF)
	}
}
