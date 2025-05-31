package encryption

import (
	"testing"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

func TestProtectionClassResolverResolveName(t *testing.T) {
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
			name:         "Unknown class",
			class:        types.CpKeyClassT(0xFF),
			expectedName: "Unknown Protection Class (0x0000001F)", // Masked value
		},
		{
			name:         "Class with extra bits",
			class:        types.ProtectionClassC | 0xFFE0,             // Set reserved bits
			expectedName: "Protected Until First User Authentication", // Should be masked
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			name := resolver.ResolveName(tc.class)
			if name != tc.expectedName {
				t.Errorf("ResolveName(%d) = '%s', want '%s'", tc.class, name, tc.expectedName)
			}
		})
	}
}

func TestProtectionClassResolverResolveDescription(t *testing.T) {
	resolver := NewProtectionClassResolver()

	tests := []struct {
		name                string
		class               types.CpKeyClassT
		expectedDescription string
	}{
		{
			name:                "Directory Default description",
			class:               types.ProtectionClassDirNone,
			expectedDescription: "Files with this protection class use their containing directory's default protection class. Used only on iOS devices.",
		},
		{
			name:                "Complete Protection description",
			class:               types.ProtectionClassA,
			expectedDescription: "Files are encrypted and inaccessible until the user unlocks the device for the first time after restart. Highest level of protection.",
		},
		{
			name:                "Protected Unless Open description",
			class:               types.ProtectionClassB,
			expectedDescription: "Files are encrypted and accessible only while the device is unlocked or the file is open. Files close when device locks.",
		},
		{
			name:                "Protected Until First User Authentication description",
			class:               types.ProtectionClassC,
			expectedDescription: "Files are encrypted but become accessible after the user unlocks the device for the first time after restart. Remain accessible until next restart.",
		},
		{
			name:                "No Protection description",
			class:               types.ProtectionClassD,
			expectedDescription: "Files are encrypted with a key derived from the device hardware. Accessible at all times, even when device is locked.",
		},
		{
			name:                "No Protection (Non-persistent Key) description",
			class:               types.ProtectionClassF,
			expectedDescription: "Same behavior as Class D, but the key is not stored persistently. Suitable for temporary files that don't need to survive device restarts.",
		},
		{
			name:                "Class M description",
			class:               types.ProtectionClassM,
			expectedDescription: "Protection class M - specific behavior not documented in public APFS specification.",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			description := resolver.ResolveDescription(tc.class)
			if description != tc.expectedDescription {
				t.Errorf("ResolveDescription(%d) = '%s', want '%s'", tc.class, description, tc.expectedDescription)
			}
		})
	}
}

func TestProtectionClassResolverListSupportedProtectionClasses(t *testing.T) {
	resolver := NewProtectionClassResolver()

	classes := resolver.ListSupportedProtectionClasses()

	expectedClasses := []types.CpKeyClassT{
		types.ProtectionClassDirNone,
		types.ProtectionClassA,
		types.ProtectionClassB,
		types.ProtectionClassC,
		types.ProtectionClassD,
		types.ProtectionClassF,
		types.ProtectionClassM,
	}

	if len(classes) != len(expectedClasses) {
		t.Errorf("ListSupportedProtectionClasses() returned %d classes, want %d", len(classes), len(expectedClasses))
	}

	for i, expected := range expectedClasses {
		if i >= len(classes) {
			t.Errorf("Missing expected class at index %d: %d", i, expected)
			continue
		}
		if classes[i] != expected {
			t.Errorf("ListSupportedProtectionClasses()[%d] = %d, want %d", i, classes[i], expected)
		}
	}
}

func TestProtectionClassResolverIsValidProtectionClass(t *testing.T) {
	resolver := NewProtectionClassResolver()

	tests := []struct {
		name          string
		class         types.CpKeyClassT
		expectedValid bool
	}{
		{"Directory Default", types.ProtectionClassDirNone, true},
		{"Complete Protection", types.ProtectionClassA, true},
		{"Protected Unless Open", types.ProtectionClassB, true},
		{"Protected Until First User Authentication", types.ProtectionClassC, true},
		{"No Protection", types.ProtectionClassD, true},
		{"No Protection (Non-persistent Key)", types.ProtectionClassF, true},
		{"Class M", types.ProtectionClassM, true},
		{"Unknown class", types.CpKeyClassT(0xFF), false},
		{"Reserved class", types.CpKeyClassT(5), false},
		{"Class with extra bits but valid base", types.ProtectionClassC | 0xFFE0, true}, // Should mask to valid class
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			valid := resolver.IsValidProtectionClass(tc.class)
			if valid != tc.expectedValid {
				t.Errorf("IsValidProtectionClass(%d) = %t, want %t", tc.class, valid, tc.expectedValid)
			}
		})
	}
}

func TestProtectionClassResolverGetEffectiveClass(t *testing.T) {
	resolver := NewProtectionClassResolver()

	tests := []struct {
		name              string
		class             types.CpKeyClassT
		expectedEffective types.CpKeyClassT
	}{
		{
			name:              "No extra bits",
			class:             types.ProtectionClassC,
			expectedEffective: types.ProtectionClassC,
		},
		{
			name:              "With reserved bits set",
			class:             types.ProtectionClassC | 0xFFE0,
			expectedEffective: types.ProtectionClassC,
		},
		{
			name:              "All reserved bits set",
			class:             types.ProtectionClassA | 0xFFFFFFE0,
			expectedEffective: types.ProtectionClassA,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			effective := resolver.GetEffectiveClass(tc.class)
			if effective != tc.expectedEffective {
				t.Errorf("GetEffectiveClass(0x%08X) = 0x%08X, want 0x%08X",
					uint32(tc.class), uint32(effective), uint32(tc.expectedEffective))
			}
		})
	}
}

func TestProtectionClassResolverIsiOSOnly(t *testing.T) {
	resolver := NewProtectionClassResolver()

	tests := []struct {
		name          string
		class         types.CpKeyClassT
		expectedIsiOS bool
	}{
		{"Directory Default", types.ProtectionClassDirNone, true},
		{"Complete Protection", types.ProtectionClassA, false},
		{"Protected Unless Open", types.ProtectionClassB, false},
		{"Protected Until First User Authentication", types.ProtectionClassC, false},
		{"No Protection", types.ProtectionClassD, false},
		{"No Protection (Non-persistent Key)", types.ProtectionClassF, true},
		{"Class M", types.ProtectionClassM, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			isiOS := resolver.IsiOSOnly(tc.class)
			if isiOS != tc.expectedIsiOS {
				t.Errorf("IsiOSOnly(%d) = %t, want %t", tc.class, isiOS, tc.expectedIsiOS)
			}
		})
	}
}

func TestProtectionClassResolverIsmacOSOnly(t *testing.T) {
	resolver := NewProtectionClassResolver()

	// Currently no protection classes are exclusive to macOS
	allClasses := resolver.ListSupportedProtectionClasses()

	for _, class := range allClasses {
		t.Run(resolver.ResolveName(class), func(t *testing.T) {
			ismacOS := resolver.IsmacOSOnly(class)
			if ismacOS {
				t.Errorf("IsmacOSOnly(%d) = true, but no classes should be macOS-only according to specification", class)
			}
		})
	}
}

func TestProtectionClassResolverGetSecurityLevel(t *testing.T) {
	resolver := NewProtectionClassResolver()

	tests := []struct {
		name          string
		class         types.CpKeyClassT
		expectedLevel int
	}{
		{"Directory Default", types.ProtectionClassDirNone, 0},
		{"Complete Protection", types.ProtectionClassA, 5},
		{"Protected Unless Open", types.ProtectionClassB, 4},
		{"Protected Until First User Authentication", types.ProtectionClassC, 3},
		{"No Protection", types.ProtectionClassD, 2},
		{"No Protection (Non-persistent Key)", types.ProtectionClassF, 1},
		{"Class M", types.ProtectionClassM, 0},
		{"Unknown class", types.CpKeyClassT(0xFF), -1},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			level := resolver.GetSecurityLevel(tc.class)
			if level != tc.expectedLevel {
				t.Errorf("GetSecurityLevel(%d) = %d, want %d", tc.class, level, tc.expectedLevel)
			}
		})
	}
}

func TestProtectionClassResolverSecurityLevelOrdering(t *testing.T) {
	resolver := NewProtectionClassResolver()

	// Test that security levels are properly ordered
	levelA := resolver.GetSecurityLevel(types.ProtectionClassA)
	levelB := resolver.GetSecurityLevel(types.ProtectionClassB)
	levelC := resolver.GetSecurityLevel(types.ProtectionClassC)
	levelD := resolver.GetSecurityLevel(types.ProtectionClassD)
	levelF := resolver.GetSecurityLevel(types.ProtectionClassF)

	if !(levelA > levelB && levelB > levelC && levelC > levelD && levelD > levelF) {
		t.Errorf("Security levels are not properly ordered: A=%d, B=%d, C=%d, D=%d, F=%d",
			levelA, levelB, levelC, levelD, levelF)
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
