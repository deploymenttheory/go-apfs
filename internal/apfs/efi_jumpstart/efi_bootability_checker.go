// File: internal/efijumpstart/bootability_checker.go
package efijumpstart

import (
	"fmt"

	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	// "github.com/deploymenttheory/go-apfs/internal/types" // Not directly needed here
)

// SimpleBootabilityChecker implements BootabilityChecker using other interfaces.
type SimpleBootabilityChecker struct {
	reader   interfaces.EFIJumpstartReader   // Required for basic check
	analyzer interfaces.EFIJumpstartAnalyzer // Optional, used for VerifyBootConfiguration
}

var _ interfaces.BootabilityChecker = (*SimpleBootabilityChecker)(nil)

// NewSimpleBootabilityChecker creates a new bootability checker.
// The reader is required, the analyzer is optional.
func NewSimpleBootabilityChecker(reader interfaces.EFIJumpstartReader, analyzer interfaces.EFIJumpstartAnalyzer) (*SimpleBootabilityChecker, error) {
	if reader == nil {
		return nil, fmt.Errorf("EFIJumpstartReader cannot be nil for bootability checks")
	}
	return &SimpleBootabilityChecker{
		reader:   reader,
		analyzer: analyzer, // Can be nil
	}, nil
}

// IsBootable checks if a valid EFI jumpstart structure exists.
func (bc *SimpleBootabilityChecker) IsBootable() bool {
	// Basic check relies solely on the reader reporting validity (magic/version).
	return bc.reader.IsValid()
}

// GetBootRequirements returns a static list of general requirements.
func (bc *SimpleBootabilityChecker) GetBootRequirements() ([]string, error) {
	// These are general requirements for APFS EFI booting.
	return []string{
		"Valid APFS Container Superblock/Checkpoint",
		"Valid EFI Jumpstart structure (Correct Magic & Version)",
		"Valid EFI Driver file referenced by Jumpstart extents",
		"EFI Firmware support for APFS booting",
	}, nil
}

// VerifyBootConfiguration performs a more detailed check using the analyzer if available.
func (bc *SimpleBootabilityChecker) VerifyBootConfiguration() error {
	// First, check basic validity via the reader.
	if !bc.reader.IsValid() {
		return fmt.Errorf("basic jumpstart validation failed (magic/version mismatch)")
	}

	if bc.analyzer != nil {
		err := bc.analyzer.VerifyEFIJumpstart()
		if err != nil {

			return fmt.Errorf("detailed boot configuration verification failed: %w", err)
		}

		return nil // Analyzer verification passed
	}

	return nil
}
