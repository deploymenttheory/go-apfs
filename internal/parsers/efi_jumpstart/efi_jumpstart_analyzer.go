// File: internal/efijumpstart/efi_jumpstart_analyzer.go
package efijumpstart

import (
	"fmt"
	"math"

	"github.com/deploymenttheory/go-apfs/internal/interfaces" // Adjust import path
	"github.com/deploymenttheory/go-apfs/internal/types"      // Adjust import path
)

// JumpstartAnalyzer implements the EFIJumpstartAnalyzer interface.
type JumpstartAnalyzer struct {
	jumpstartData   types.NxEfiJumpstartT
	blockSize       uint32
	driverExtractor interfaces.EFIDriverExtractor // Optional: Used to analyze driver content
}

// Compile-time check to ensure JumpstartAnalyzer implements the interface
var _ interfaces.EFIJumpstartAnalyzer = (*JumpstartAnalyzer)(nil)

// NewEFIJumpstartAnalyzer creates a new analyzer instance.
// The driverExtractor is optional; if nil, driver content analysis will be skipped.
func NewEFIJumpstartAnalyzer(
	data types.NxEfiJumpstartT,
	blockSize uint32,
	extractor interfaces.EFIDriverExtractor, // Can be nil
) (interfaces.EFIJumpstartAnalyzer, error) {

	if blockSize == 0 {
		return nil, fmt.Errorf("block size cannot be zero")
	}
	// Basic consistency check
	if data.NejNumExtents != uint32(len(data.NejRecExtents)) {
		return nil, fmt.Errorf("NejNumExtents (%d) mismatch with NejRecExtents length (%d)", data.NejNumExtents, len(data.NejRecExtents))
	}

	return &JumpstartAnalyzer{
		jumpstartData:   data,
		blockSize:       blockSize,
		driverExtractor: extractor, // Store nil if not provided
	}, nil
}

// VerifyEFIJumpstart checks basic validity (magic, version) and readability via the extractor if available.
func (a *JumpstartAnalyzer) VerifyEFIJumpstart() error {
	// 1. Check Magic
	if a.jumpstartData.NejMagic != types.NxEfiJumpstartMagic {
		return fmt.Errorf("invalid magic number: expected %X, got %X", types.NxEfiJumpstartMagic, a.jumpstartData.NejMagic)
	}

	// 2. Check Version
	if a.jumpstartData.NejVersion != types.NxEfiJumpstartVersion {
		return fmt.Errorf("invalid version: expected %d, got %d", types.NxEfiJumpstartVersion, a.jumpstartData.NejVersion)
	}

	// 3. Check Extent Consistency (already done in constructor, maybe redundant)
	if a.jumpstartData.NejNumExtents != uint32(len(a.jumpstartData.NejRecExtents)) {
		return fmt.Errorf("internal inconsistency: NejNumExtents (%d) mismatch with NejRecExtents length (%d)", a.jumpstartData.NejNumExtents, len(a.jumpstartData.NejRecExtents))
	}

	// 4. Check if driver data is readable (if extractor provided)
	if a.driverExtractor != nil {
		err := a.driverExtractor.ValidateEFIDriver()
		if err != nil {
			return fmt.Errorf("driver validation/readability check failed: %w", err)
		}
	}

	return nil // All checks passed
}

// AnalyzeEFIJumpstart performs a detailed analysis and populates the EFIJumpstartAnalysis struct.
func (a *JumpstartAnalyzer) AnalyzeEFIJumpstart() (interfaces.EFIJumpstartAnalysis, error) {
	analysis := interfaces.EFIJumpstartAnalysis{
		// Initialize map to avoid nil map errors later
		DriverInfo: make(map[string]string),
	}

	// --- Populate basic fields ---
	analysis.IsValid = (a.jumpstartData.NejMagic == types.NxEfiJumpstartMagic &&
		a.jumpstartData.NejVersion == types.NxEfiJumpstartVersion)
	analysis.Version = a.jumpstartData.NejVersion
	analysis.DriverSize = a.jumpstartData.NejEfiFileLen
	analysis.ExtentCount = a.jumpstartData.NejNumExtents

	// --- Populate Extent Details ---
	analysis.ExtentDetails = make([]interfaces.EFIExtentDetail, 0, analysis.ExtentCount)
	var calculatedTotalSizeFromExtents uint64 = 0
	blockSizeF := float64(a.blockSize) // For potential overflow check

	for _, extent := range a.jumpstartData.NejRecExtents {
		// Check for potential overflow when calculating size in bytes
		// Very unlikely with uint64 unless block count is astronomically large
		sizeInBytesF := float64(extent.PrBlockCount) * blockSizeF
		if sizeInBytesF > float64(math.MaxUint64) {
			// Handle potential overflow - maybe skip this extent detail or log warning?
			// For now, let's assume it fits or truncates if uint64() conversion overflows.
			// A more robust solution might involve big.Int if truly huge values are possible.
		}
		sizeInBytes := extent.PrBlockCount * uint64(a.blockSize)
		calculatedTotalSizeFromExtents += sizeInBytes

		detail := interfaces.EFIExtentDetail{
			StartAddress: extent.PrStartPaddr,
			BlockCount:   extent.PrBlockCount,
			SizeInBytes:  sizeInBytes,
		}
		analysis.ExtentDetails = append(analysis.ExtentDetails, detail)
	}

	// Add consistency check info to DriverInfo
	if calculatedTotalSizeFromExtents < uint64(analysis.DriverSize) {
		analysis.DriverInfo["ConsistencyWarning"] = fmt.Sprintf("Declared DriverSize (%d) is LARGER than total size calculated from extents (%d)", analysis.DriverSize, calculatedTotalSizeFromExtents)
	} else if calculatedTotalSizeFromExtents > uint64(analysis.DriverSize) && analysis.DriverSize > 0 {
		// It's normal for extents to cover more bytes than the actual file length
		analysis.DriverInfo["ExtentCoverage"] = fmt.Sprintf("Extents cover %d bytes, DriverSize is %d bytes", calculatedTotalSizeFromExtents, analysis.DriverSize)
	}

	// --- Analyze Driver Content (if extractor available) ---
	if a.driverExtractor != nil && analysis.DriverSize > 0 {
		driverData, err := a.driverExtractor.GetEFIDriverData()
		if err != nil {
			analysis.DriverInfo["DriverReadable"] = "No"
			analysis.DriverInfo["ReadError"] = err.Error()
		} else {
			analysis.DriverInfo["DriverReadable"] = "Yes"
			// Basic PE/COFF check (DOS MZ header)
			if len(driverData) >= 2 && driverData[0] == 'M' && driverData[1] == 'Z' {
				analysis.DriverInfo["MZHeaderFound"] = "Yes"
				// Could add more checks here (e.g., find PE signature offset, check PE signature)
			} else {
				analysis.DriverInfo["MZHeaderFound"] = "No"
			}
			// Add length check consistency
			if len(driverData) != int(analysis.DriverSize) {
				analysis.DriverInfo["ReadLengthConsistency"] = fmt.Sprintf("Warning: Read %d bytes, but DriverSize field was %d", len(driverData), analysis.DriverSize)
			}
		}
	} else if analysis.DriverSize == 0 {
		analysis.DriverInfo["DriverContent"] = "N/A (DriverSize is 0)"
	} else {
		analysis.DriverInfo["DriverContent"] = "N/A (Extractor not provided)"
	}

	return analysis, nil // Assume analysis itself doesn't fail fundamentally
}
