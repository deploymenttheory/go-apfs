package encryption

import (
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
