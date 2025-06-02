package encryption

import (
	"fmt"

	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// protectionClassResolver implements the ProtectionClassResolver interface
type protectionClassResolver struct{}

// Ensure protectionClassResolver implements the ProtectionClassResolver interface
var _ interfaces.ProtectionClassResolver = (*protectionClassResolver)(nil)

// NewProtectionClassResolver creates a new ProtectionClassResolver
func NewProtectionClassResolver() interfaces.ProtectionClassResolver {
	return &protectionClassResolver{}
}

// ResolveName returns a human-readable name for a protection class
func (pcr *protectionClassResolver) ResolveName(class types.CpKeyClassT) string {
	// Apply the effective class mask to get the actual protection class
	effectiveClass := class & types.CpEffectiveClassmask

	switch effectiveClass {
	case types.ProtectionClassDirNone:
		return "Directory Default"
	case types.ProtectionClassA:
		return "Complete Protection"
	case types.ProtectionClassB:
		return "Protected Unless Open"
	case types.ProtectionClassC:
		return "Protected Until First User Authentication"
	case types.ProtectionClassD:
		return "No Protection"
	case types.ProtectionClassF:
		return "No Protection (Non-persistent Key)"
	case types.ProtectionClassM:
		return "Class M"
	default:
		return fmt.Sprintf("Unknown Protection Class (0x%08X)", uint32(effectiveClass))
	}
}

// ResolveDescription provides a detailed description of a protection class
func (pcr *protectionClassResolver) ResolveDescription(class types.CpKeyClassT) string {
	// Apply the effective class mask to get the actual protection class
	effectiveClass := class & types.CpEffectiveClassmask

	switch effectiveClass {
	case types.ProtectionClassDirNone:
		return "Files with this protection class use their containing directory's default protection class. Used only on iOS devices."

	case types.ProtectionClassA:
		return "Files are encrypted and inaccessible until the user unlocks the device for the first time after restart. Highest level of protection."

	case types.ProtectionClassB:
		return "Files are encrypted and accessible only while the device is unlocked or the file is open. Files close when device locks."

	case types.ProtectionClassC:
		return "Files are encrypted but become accessible after the user unlocks the device for the first time after restart. Remain accessible until next restart."

	case types.ProtectionClassD:
		return "Files are encrypted with a key derived from the device hardware. Accessible at all times, even when device is locked."

	case types.ProtectionClassF:
		return "Same behavior as Class D, but the key is not stored persistently. Suitable for temporary files that don't need to survive device restarts."

	case types.ProtectionClassM:
		return "Protection class M - specific behavior not documented in public APFS specification."

	default:
		return fmt.Sprintf("Unknown protection class with value 0x%08X. This may indicate a newer APFS version or corrupted data.", uint32(effectiveClass))
	}
}

// ListSupportedProtectionClasses returns all supported protection classes
func (pcr *protectionClassResolver) ListSupportedProtectionClasses() []types.CpKeyClassT {
	return []types.CpKeyClassT{
		types.ProtectionClassDirNone,
		types.ProtectionClassA,
		types.ProtectionClassB,
		types.ProtectionClassC,
		types.ProtectionClassD,
		types.ProtectionClassF,
		types.ProtectionClassM,
	}
}

// IsValidProtectionClass checks if a protection class is known and valid
func (pcr *protectionClassResolver) IsValidProtectionClass(class types.CpKeyClassT) bool {
	effectiveClass := class & types.CpEffectiveClassmask
	supportedClasses := pcr.ListSupportedProtectionClasses()

	for _, supported := range supportedClasses {
		if effectiveClass == supported {
			return true
		}
	}
	return false
}

// GetEffectiveClass returns the effective protection class after applying the mask
func (pcr *protectionClassResolver) GetEffectiveClass(class types.CpKeyClassT) types.CpKeyClassT {
	return class & types.CpEffectiveClassmask
}

// IsiOSOnly returns true if the protection class is used only on iOS devices
func (pcr *protectionClassResolver) IsiOSOnly(class types.CpKeyClassT) bool {
	effectiveClass := class & types.CpEffectiveClassmask
	switch effectiveClass {
	case types.ProtectionClassDirNone, types.ProtectionClassF:
		return true
	default:
		return false
	}
}

// IsmacOSOnly returns true if the protection class is used only on macOS devices
func (pcr *protectionClassResolver) IsmacOSOnly(class types.CpKeyClassT) bool {
	// Currently no protection classes are exclusive to macOS in the public specification
	return false
}

// GetSecurityLevel returns a numeric security level (higher = more secure)
func (pcr *protectionClassResolver) GetSecurityLevel(class types.CpKeyClassT) int {
	effectiveClass := class & types.CpEffectiveClassmask

	switch effectiveClass {
	case types.ProtectionClassA:
		return 5 // Highest security
	case types.ProtectionClassB:
		return 4
	case types.ProtectionClassC:
		return 3
	case types.ProtectionClassD:
		return 2
	case types.ProtectionClassF:
		return 1 // Non-persistent, still some protection
	case types.ProtectionClassDirNone:
		return 0 // Depends on directory
	case types.ProtectionClassM:
		return 0 // Unknown behavior
	default:
		return -1 // Unknown/invalid
	}
}
