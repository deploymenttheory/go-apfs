package encryption

import (
	"testing"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

func TestProtectionClassResolver_ResolveName(t *testing.T) {
	resolver := NewProtectionClassResolver()

	tests := []struct {
		name         string
		class        types.CpKeyClassT
		expectedName string
	}{
		{
			name:         "Directory Default",
			class:        types.ProtectionClassDirNone,
			expectedName: "Directory Default",
		},
		{
			name:         "Complete Protection",
			class:        types.ProtectionClassA,
			expectedName: "Complete Protection",
		},
		{
			name:         "Protected Unless Open",
			class:        types.ProtectionClassB,
			expectedName: "Protected Unless Open",
		},
		{
			name:         "Protected Until First User Authentication",
			class:        types.ProtectionClassC,
			expectedName: "Protected Until First User Authentication",
		},
		{
			name:         "No Protection",
			class:        types.ProtectionClassD,
			expectedName: "No Protection",
		},
		{
			name:         "No Protection (Non-persistent Key)",
			class:        types.ProtectionClassF,
			expectedName: "No Protection (Non-persistent Key)",
		},
		{
			name:         "Class M",
			class:        types.ProtectionClassM,
			expectedName: "Class M",
		},
		{
			name:         "Unknown Protection Class",
			class:        types.CpKeyClassT(999),
			expectedName: "Unknown Protection Class (0x000003E7)",
		},
		{
			name:         "Class with additional flags",
			class:        types.ProtectionClassC | 0x80000000, // High bits set
			expectedName: "Protected Until First User Authentication",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := resolver.ResolveName(tc.class)
			if result != tc.expectedName {
				t.Errorf("ResolveName(%d) = %q, want %q", tc.class, result, tc.expectedName)
			}
		})
	}
}

func TestProtectionClassResolver_ResolveDescription(t *testing.T) {
	resolver := NewProtectionClassResolver()

	tests := []struct {
		name                string
		class               types.CpKeyClassT
		expectedDescription string
	}{
		{
			name:                "Directory Default Description",
			class:               types.ProtectionClassDirNone,
			expectedDescription: "Files with this protection class use their containing directory's default protection class. Used only on iOS devices.",
		},
		{
			name:                "Complete Protection Description",
			class:               types.ProtectionClassA,
			expectedDescription: "Files are encrypted and inaccessible until the user unlocks the device for the first time after restart. Highest level of protection.",
		},
		{
			name:                "Protected Unless Open Description",
			class:               types.ProtectionClassB,
			expectedDescription: "Files are encrypted and accessible only while the device is unlocked or the file is open. Files close when device locks.",
		},
		{
			name:                "Protected Until First User Authentication Description",
			class:               types.ProtectionClassC,
			expectedDescription: "Files are encrypted but become accessible after the user unlocks the device for the first time after restart. Remain accessible until next restart.",
		},
		{
			name:                "No Protection Description",
			class:               types.ProtectionClassD,
			expectedDescription: "Files are encrypted with a key derived from the device hardware. Accessible at all times, even when device is locked.",
		},
		{
			name:                "No Protection (Non-persistent Key) Description",
			class:               types.ProtectionClassF,
			expectedDescription: "Same behavior as Class D, but the key is not stored persistently. Suitable for temporary files that don't need to survive device restarts.",
		},
		{
			name:                "Class M Description",
			class:               types.ProtectionClassM,
			expectedDescription: "Protection class M - specific behavior not documented in public APFS specification.",
		},
		{
			name:                "Unknown Protection Class Description",
			class:               types.CpKeyClassT(999),
			expectedDescription: "Unknown protection class with value 0x000003E7. This may indicate a newer APFS version or corrupted data.",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := resolver.ResolveDescription(tc.class)
			if result != tc.expectedDescription {
				t.Errorf("ResolveDescription(%d) = %q, want %q", tc.class, result, tc.expectedDescription)
			}
		})
	}
}

func TestProtectionClassResolver_ListSupportedProtectionClasses(t *testing.T) {
	resolver := NewProtectionClassResolver()

	expectedClasses := []types.CpKeyClassT{
		types.ProtectionClassDirNone,
		types.ProtectionClassA,
		types.ProtectionClassB,
		types.ProtectionClassC,
		types.ProtectionClassD,
		types.ProtectionClassF,
		types.ProtectionClassM,
	}

	result := resolver.ListSupportedProtectionClasses()

	if len(result) != len(expectedClasses) {
		t.Errorf("ListSupportedProtectionClasses() count = %d, want %d", len(result), len(expectedClasses))
	}

	// Convert to map for easier checking
	resultMap := make(map[types.CpKeyClassT]bool)
	for _, class := range result {
		resultMap[class] = true
	}

	for _, expectedClass := range expectedClasses {
		if !resultMap[expectedClass] {
			t.Errorf("ListSupportedProtectionClasses() missing class %d", expectedClass)
		}
	}
}

func TestProtectionClassResolver_IsValidProtectionClass(t *testing.T) {
	resolver := NewProtectionClassResolver()

	tests := []struct {
		name          string
		class         types.CpKeyClassT
		expectedValid bool
	}{
		{
			name:          "Valid - Directory Default",
			class:         types.ProtectionClassDirNone,
			expectedValid: true,
		},
		{
			name:          "Valid - Class A",
			class:         types.ProtectionClassA,
			expectedValid: true,
		},
		{
			name:          "Valid - Class B",
			class:         types.ProtectionClassB,
			expectedValid: true,
		},
		{
			name:          "Valid - Class C",
			class:         types.ProtectionClassC,
			expectedValid: true,
		},
		{
			name:          "Valid - Class D",
			class:         types.ProtectionClassD,
			expectedValid: true,
		},
		{
			name:          "Valid - Class F",
			class:         types.ProtectionClassF,
			expectedValid: true,
		},
		{
			name:          "Valid - Class M",
			class:         types.ProtectionClassM,
			expectedValid: true,
		},
		{
			name:          "Valid - Class C with additional flags",
			class:         types.ProtectionClassC | 0x80000000,
			expectedValid: true,
		},
		{
			name:          "Invalid - Unknown class",
			class:         types.CpKeyClassT(999),
			expectedValid: false,
		},
		{
			name:          "Invalid - Unknown class with flags",
			class:         types.CpKeyClassT(999) | 0x80000000,
			expectedValid: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := resolver.IsValidProtectionClass(tc.class)
			if result != tc.expectedValid {
				t.Errorf("IsValidProtectionClass(%d) = %t, want %t", tc.class, result, tc.expectedValid)
			}
		})
	}
}

func TestProtectionClassResolver_GetEffectiveClass(t *testing.T) {
	resolver := NewProtectionClassResolver()

	tests := []struct {
		name              string
		class             types.CpKeyClassT
		expectedEffective types.CpKeyClassT
	}{
		{
			name:              "Class C without flags",
			class:             types.ProtectionClassC,
			expectedEffective: types.ProtectionClassC,
		},
		{
			name:              "Class C with high flags",
			class:             types.ProtectionClassC | 0x80000000,
			expectedEffective: types.ProtectionClassC,
		},
		{
			name:              "Class A with multiple flags",
			class:             types.ProtectionClassA | 0xFFFF0000,
			expectedEffective: types.ProtectionClassA,
		},
		{
			name:              "Unknown class with flags",
			class:             types.CpKeyClassT(999) | 0x12340000,
			expectedEffective: types.CpKeyClassT(999 & 0x1F), // Apply mask
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := resolver.GetEffectiveClass(tc.class)
			if result != tc.expectedEffective {
				t.Errorf("GetEffectiveClass(%d) = %d, want %d", tc.class, result, tc.expectedEffective)
			}
		})
	}
}

func TestProtectionClassResolver_IsiOSOnly(t *testing.T) {
	resolver := NewProtectionClassResolver()

	tests := []struct {
		name      string
		class     types.CpKeyClassT
		expectiOS bool
	}{
		{
			name:      "Directory Default - iOS only",
			class:     types.ProtectionClassDirNone,
			expectiOS: true,
		},
		{
			name:      "Class F - iOS only",
			class:     types.ProtectionClassF,
			expectiOS: true,
		},
		{
			name:      "Class A - Not iOS only",
			class:     types.ProtectionClassA,
			expectiOS: false,
		},
		{
			name:      "Class C - Not iOS only",
			class:     types.ProtectionClassC,
			expectiOS: false,
		},
		{
			name:      "Class D - Not iOS only",
			class:     types.ProtectionClassD,
			expectiOS: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := resolver.IsiOSOnly(tc.class)
			if result != tc.expectiOS {
				t.Errorf("IsiOSOnly(%d) = %t, want %t", tc.class, result, tc.expectiOS)
			}
		})
	}
}

func TestProtectionClassResolver_IsmacOSOnly(t *testing.T) {
	resolver := NewProtectionClassResolver()

	// Currently no protection classes are exclusive to macOS
	allClasses := resolver.ListSupportedProtectionClasses()

	for _, class := range allClasses {
		t.Run("Class "+resolver.ResolveName(class), func(t *testing.T) {
			result := resolver.IsmacOSOnly(class)
			if result != false {
				t.Errorf("IsmacOSOnly(%d) = %t, want false", class, result)
			}
		})
	}
}

func TestProtectionClassResolver_GetSecurityLevel(t *testing.T) {
	resolver := NewProtectionClassResolver()

	tests := []struct {
		name          string
		class         types.CpKeyClassT
		expectedLevel int
	}{
		{
			name:          "Class A - Highest security",
			class:         types.ProtectionClassA,
			expectedLevel: 5,
		},
		{
			name:          "Class B - High security",
			class:         types.ProtectionClassB,
			expectedLevel: 4,
		},
		{
			name:          "Class C - Medium security",
			class:         types.ProtectionClassC,
			expectedLevel: 3,
		},
		{
			name:          "Class D - Low security",
			class:         types.ProtectionClassD,
			expectedLevel: 2,
		},
		{
			name:          "Class F - Minimal security",
			class:         types.ProtectionClassF,
			expectedLevel: 1,
		},
		{
			name:          "Directory Default - Depends on directory",
			class:         types.ProtectionClassDirNone,
			expectedLevel: 0,
		},
		{
			name:          "Class M - Unknown behavior",
			class:         types.ProtectionClassM,
			expectedLevel: 0,
		},
		{
			name:          "Unknown class - Invalid",
			class:         types.CpKeyClassT(999),
			expectedLevel: -1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := resolver.GetSecurityLevel(tc.class)
			if result != tc.expectedLevel {
				t.Errorf("GetSecurityLevel(%d) = %d, want %d", tc.class, result, tc.expectedLevel)
			}
		})
	}
}

func TestProtectionClassResolver_EffectiveClassMask(t *testing.T) {
	resolver := NewProtectionClassResolver()

	// Test that the effective class mask correctly isolates the protection class
	tests := []struct {
		name     string
		input    types.CpKeyClassT
		expected types.CpKeyClassT
	}{
		{
			name:     "Clean class value",
			input:    types.ProtectionClassC,
			expected: types.ProtectionClassC,
		},
		{
			name:     "Class with high bits set",
			input:    types.ProtectionClassC | 0x80000000,
			expected: types.ProtectionClassC,
		},
		{
			name:     "Class with all upper bits set",
			input:    types.ProtectionClassA | 0xFFFFFFE0,
			expected: types.ProtectionClassA,
		},
		{
			name:     "Value larger than mask",
			input:    types.CpKeyClassT(0x12345678),
			expected: types.CpKeyClassT(0x12345678 & 0x1F), // Apply 5-bit mask
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := resolver.GetEffectiveClass(tc.input)
			if result != tc.expected {
				t.Errorf("GetEffectiveClass(0x%08X) = 0x%08X, want 0x%08X",
					uint32(tc.input), uint32(result), uint32(tc.expected))
			}
		})
	}
}

func BenchmarkProtectionClassResolver_ResolveName(b *testing.B) {
	resolver := NewProtectionClassResolver()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = resolver.ResolveName(types.ProtectionClassC)
	}
}

func BenchmarkProtectionClassResolver_ResolveDescription(b *testing.B) {
	resolver := NewProtectionClassResolver()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = resolver.ResolveDescription(types.ProtectionClassC)
	}
}

func BenchmarkProtectionClassResolver_IsValidProtectionClass(b *testing.B) {
	resolver := NewProtectionClassResolver()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = resolver.IsValidProtectionClass(types.ProtectionClassC)
	}
}

func BenchmarkProtectionClassResolver_GetSecurityLevel(b *testing.B) {
	resolver := NewProtectionClassResolver()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = resolver.GetSecurityLevel(types.ProtectionClassC)
	}
}
