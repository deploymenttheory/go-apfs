package types

import "fmt"

// VersionInfo contains version information for APFS features
type VersionInfo struct {
	// Major version
	Major uint16

	// Minor version
	Minor uint16

	// Incompatible features
	IncompatFeatures uint64

	// Read-only compatible features
	ROCompatFeatures uint64

	// Compatible features
	CompatFeatures uint64

	// Formatted by
	FormattedBy string

	// Last modified by
	LastModifiedBy string

	// Last modified transaction ID
	LastModifiedXID XID
}

// APFSVersion represents an APFS version
type APFSVersion uint32

const (
	// APFSVersionUnknown represents an unknown or invalid version
	APFSVersionUnknown APFSVersion = 0

	// APFSVersionPrerelease represents the prerelease version (v1)
	APFSVersionPrerelease APFSVersion = 1

	// APFSVersion2 represents APFS version 2 (macOS 10.13, iOS 10.3)
	APFSVersion2 APFSVersion = 2

	// APFSVersion3 represents APFS version 3
	APFSVersion3 APFSVersion = 3

	// APFSVersionCurrent represents the current supported version
	APFSVersionCurrent = APFSVersion2
)

// String returns a string representation of the APFS version
func (v APFSVersion) String() string {
	switch v {
	case APFSVersionUnknown:
		return "Unknown"
	case APFSVersionPrerelease:
		return "Prerelease (v1)"
	case APFSVersion2:
		return "APFS v2 (macOS 10.13, iOS 10.3)"
	case APFSVersion3:
		return "APFS v3"
	default:
		return fmt.Sprintf("Unknown version %d", v)
	}
}

// DetectContainerVersion detects the APFS container version from a superblock
func DetectContainerVersion(sb *NXSuperblock) APFSVersion {
	if sb.IncompatFeatures&NXIncompatVersion1 != 0 {
		return APFSVersionPrerelease
	}

	if sb.IncompatFeatures&NXIncompatVersion2 != 0 {
		return APFSVersion2
	}

	return APFSVersionUnknown
}

// IsFeatureSupported checks if a feature is supported
func IsFeatureSupported(feature uint64, supportedFeatures uint64) bool {
	return (feature & supportedFeatures) == feature
}

// ContainerCompatibilityError describes a compatibility issue with a container
type ContainerCompatibilityError struct {
	Version            APFSVersion
	UnsupportedFeature uint64
	FeatureType        string
}

// Error implements the error interface
func (e *ContainerCompatibilityError) Error() string {
	return fmt.Sprintf("container has unsupported %s feature 0x%016x (version %s)",
		e.FeatureType, e.UnsupportedFeature, e.Version)
}

// VolumeCompatibilityError describes a compatibility issue with a volume
type VolumeCompatibilityError struct {
	UnsupportedFeature uint64
	FeatureType        string
}

// Error implements the error interface
func (e *VolumeCompatibilityError) Error() string {
	return fmt.Sprintf("volume has unsupported %s feature 0x%016x",
		e.FeatureType, e.UnsupportedFeature)
}

// CheckContainerCompatibility checks if a container can be accessed
func CheckContainerCompatibility(container *NXSuperblock) error {
	// Check container version
	version := DetectContainerVersion(container)
	if version == APFSVersionUnknown {
		return &ContainerCompatibilityError{
			Version:     version,
			FeatureType: "version",
		}
	}

	// Check for unsupported incompatible features
	if unsupported := container.IncompatFeatures &^ NXSupportedIncompatMask; unsupported != 0 {
		return &ContainerCompatibilityError{
			Version:            version,
			UnsupportedFeature: unsupported,
			FeatureType:        "incompatible",
		}
	}

	// If no errors, container is compatible
	return nil
}

// CheckVolumeCompatibility checks if a volume can be accessed
func CheckVolumeCompatibility(volume *APFSSuperblock, readOnly bool) error {
	// Check for unsupported incompatible features
	if unsupported := volume.IncompatFeatures &^ APFSSupportedIncompatMask; unsupported != 0 {
		return &VolumeCompatibilityError{
			UnsupportedFeature: unsupported,
			FeatureType:        "incompatible",
		}
	}

	// Check for unsupported read-only compatible features
	if !readOnly {
		if unsupported := volume.ReadOnlyCompatFeatures &^ APFSSupportedROCompatMask; unsupported != 0 {
			return &VolumeCompatibilityError{
				UnsupportedFeature: unsupported,
				FeatureType:        "read-only compatible",
			}
		}
	}

	// If no errors, volume is compatible
	return nil
}

// IsContainerReadOnly determines if the container should be mounted read-only
func IsContainerReadOnly(container *NXSuperblock) bool {
	// If there are read-only compatible features we don't support, it must be mounted read-only
	if unsupported := container.ReadOnlyCompatFeatures &^ NXSupportedROCompatMask; unsupported != 0 {
		return true
	}
	return false
}
