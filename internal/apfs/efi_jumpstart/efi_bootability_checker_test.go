// File: internal/efijumpstart/efi_bootability_checker_test.go
package efijumpstart

import (
	"errors"
	"strings" // Import strings
	"testing"

	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Mocks (remain the same) ---
type mockJumpstartReader struct {
	isValidResult bool
}

func (m *mockJumpstartReader) IsValid() bool           { return m.isValidResult }
func (m *mockJumpstartReader) Magic() uint32           { return 0 }
func (m *mockJumpstartReader) Version() uint32         { return 0 }
func (m *mockJumpstartReader) EFIFileLength() uint32   { return 0 }
func (m *mockJumpstartReader) ExtentCount() uint32     { return 0 }
func (m *mockJumpstartReader) Extents() []types.Prange { return nil }

type mockJumpstartAnalyzer struct {
	verifyError error
}

func (m *mockJumpstartAnalyzer) VerifyEFIJumpstart() error { return m.verifyError }
func (m *mockJumpstartAnalyzer) AnalyzeEFIJumpstart() (interfaces.EFIJumpstartAnalysis, error) {
	return interfaces.EFIJumpstartAnalysis{}, errors.New("AnalyzeEFIJumpstart not implemented in mock")
}

// --- Test New (remains the same) ---
func TestNewSimpleBootabilityChecker(t *testing.T) {
	validReader := &mockJumpstartReader{isValidResult: true}
	validAnalyzer := &mockJumpstartAnalyzer{verifyError: nil}

	t.Run("Success_AllDeps", func(t *testing.T) {
		checker, err := NewSimpleBootabilityChecker(validReader, validAnalyzer)
		require.NoError(t, err)
		require.NotNil(t, checker)
		assert.Equal(t, validReader, checker.reader)
		assert.Equal(t, validAnalyzer, checker.analyzer)
	})
	t.Run("Success_NilAnalyzer", func(t *testing.T) {
		checker, err := NewSimpleBootabilityChecker(validReader, nil)
		require.NoError(t, err)
		require.NotNil(t, checker)
		assert.Equal(t, validReader, checker.reader)
		assert.Nil(t, checker.analyzer)
	})
	t.Run("Fail_NilReader", func(t *testing.T) {
		checker, err := NewSimpleBootabilityChecker(nil, validAnalyzer)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "EFIJumpstartReader cannot be nil")
		assert.Nil(t, checker)
	})
}

// --- Test IsBootable (remains the same) ---
func TestSimpleBootabilityChecker_IsBootable(t *testing.T) {
	t.Run("ReaderIsValid", func(t *testing.T) {
		reader := &mockJumpstartReader{isValidResult: true}
		checker, _ := NewSimpleBootabilityChecker(reader, nil)
		assert.True(t, checker.IsBootable())
	})
	t.Run("ReaderIsNotValid", func(t *testing.T) {
		reader := &mockJumpstartReader{isValidResult: false}
		checker, _ := NewSimpleBootabilityChecker(reader, nil)
		assert.False(t, checker.IsBootable())
	})
}

func TestSimpleBootabilityChecker_GetBootRequirements(t *testing.T) {
	reader := &mockJumpstartReader{isValidResult: true} // Reader validity doesn't affect this method
	checker, _ := NewSimpleBootabilityChecker(reader, nil)
	reqs, err := checker.GetBootRequirements()
	require.NoError(t, err)
	assert.NotEmpty(t, reqs, "Should return a non-empty list of requirements")
	assert.GreaterOrEqual(t, len(reqs), 3, "Should list at least a few requirements")

	// *** CORRECTED CHECK ***
	// Check if the word "Jumpstart" appears in any of the requirements
	foundJumpstart := false
	for _, req := range reqs {
		if strings.Contains(req, "Jumpstart") {
			foundJumpstart = true
			break
		}
	}
	assert.True(t, foundJumpstart, "The requirements list should mention 'Jumpstart'")
	// *** END CORRECTION ***
}

// --- Test VerifyBootConfiguration (remains the same) ---
func TestSimpleBootabilityChecker_VerifyBootConfiguration(t *testing.T) {
	validReader := &mockJumpstartReader{isValidResult: true}
	invalidReader := &mockJumpstartReader{isValidResult: false}
	analyzerSuccess := &mockJumpstartAnalyzer{verifyError: nil}
	analyzerFail := &mockJumpstartAnalyzer{verifyError: errors.New("analyzer verify failed")}

	t.Run("ValidReader_NilAnalyzer", func(t *testing.T) {
		checker, _ := NewSimpleBootabilityChecker(validReader, nil)
		err := checker.VerifyBootConfiguration()
		assert.NoError(t, err)
	})
	t.Run("ValidReader_AnalyzerSuccess", func(t *testing.T) {
		checker, _ := NewSimpleBootabilityChecker(validReader, analyzerSuccess)
		err := checker.VerifyBootConfiguration()
		assert.NoError(t, err)
	})
	t.Run("ValidReader_AnalyzerFail", func(t *testing.T) {
		checker, _ := NewSimpleBootabilityChecker(validReader, analyzerFail)
		err := checker.VerifyBootConfiguration()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "detailed boot configuration verification failed")
		assert.ErrorIs(t, err, analyzerFail.verifyError)
	})
	t.Run("InvalidReader_NilAnalyzer", func(t *testing.T) {
		checker, _ := NewSimpleBootabilityChecker(invalidReader, nil)
		err := checker.VerifyBootConfiguration()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "basic jumpstart validation failed")
	})
	t.Run("InvalidReader_AnalyzerSuccess", func(t *testing.T) {
		checker, _ := NewSimpleBootabilityChecker(invalidReader, analyzerSuccess)
		err := checker.VerifyBootConfiguration()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "basic jumpstart validation failed")
	})
}
