// File: internal/efijumpstart/efi_jumpstart_analyzer_test.go
package efijumpstart // Should likely be jumpstartanalyzer

import (
	"bytes"
	"testing"

	"github.com/deploymenttheory/go-apfs/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEFIJumpstartAnalyzer(t *testing.T) {
	validData := types.NxEfiJumpstartT{NejNumExtents: 0, NejRecExtents: []types.Prange{}}
	t.Run("Success_NilExtractor", func(t *testing.T) {
		analyzer, err := NewEFIJumpstartAnalyzer(validData, 4096, nil)
		require.NoError(t, err)
		require.NotNil(t, analyzer)
	})
	t.Run("Success_WithExtractor", func(t *testing.T) {
		mockReader := bytes.NewReader([]byte{})
		realExtractor, _ := NewEFIDriverExtractor(validData, mockReader, 4096)
		analyzer, err := NewEFIJumpstartAnalyzer(validData, 4096, realExtractor)
		require.NoError(t, err)
		require.NotNil(t, analyzer)
	})
	t.Run("Fail_ZeroBlockSize", func(t *testing.T) {
		analyzer, err := NewEFIJumpstartAnalyzer(validData, 0, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "block size cannot be zero")
		assert.Nil(t, analyzer)
	})
	t.Run("Fail_ExtentCountMismatch", func(t *testing.T) {
		mismatchData := types.NxEfiJumpstartT{NejNumExtents: 1, NejRecExtents: []types.Prange{}}
		analyzer, err := NewEFIJumpstartAnalyzer(mismatchData, 4096, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "NejNumExtents (1) mismatch with NejRecExtents length (0)")
		assert.Nil(t, analyzer)
	})
}

func TestJumpstartAnalyzerImpl_VerifyEFIJumpstart(t *testing.T) {
	blockSize := uint32(4096)
	validData := types.NxEfiJumpstartT{
		NejMagic: types.NxEfiJumpstartMagic, NejVersion: types.NxEfiJumpstartVersion,
		NejEfiFileLen: 0, NejNumExtents: 0, NejRecExtents: []types.Prange{},
	}
	mockReader := bytes.NewReader([]byte{})
	validExtractor, _ := NewEFIDriverExtractor(validData, mockReader, blockSize)

	t.Run("Success", func(t *testing.T) {
		analyzer, err := NewEFIJumpstartAnalyzer(validData, blockSize, validExtractor)
		require.NoError(t, err)
		assert.NoError(t, analyzer.VerifyEFIJumpstart())
	})
	t.Run("Success_NilExtractor", func(t *testing.T) {
		analyzer, err := NewEFIJumpstartAnalyzer(validData, blockSize, nil)
		require.NoError(t, err)
		assert.NoError(t, analyzer.VerifyEFIJumpstart())
	})
	t.Run("Fail_InvalidMagic", func(t *testing.T) {
		invalidMagic := validData
		invalidMagic.NejMagic = 0xDEADBEEF
		analyzer, err := NewEFIJumpstartAnalyzer(invalidMagic, blockSize, nil)
		require.NoError(t, err)
		verifyErr := analyzer.VerifyEFIJumpstart()
		require.Error(t, verifyErr)
		assert.Contains(t, verifyErr.Error(), "invalid magic number")
	})
	t.Run("Fail_InvalidVersion", func(t *testing.T) {
		invalidVersion := validData
		invalidVersion.NejVersion = 99
		analyzer, err := NewEFIJumpstartAnalyzer(invalidVersion, blockSize, nil)
		require.NoError(t, err)
		verifyErr := analyzer.VerifyEFIJumpstart()
		require.Error(t, verifyErr)
		assert.Contains(t, verifyErr.Error(), "invalid version")
	})
	t.Run("Fail_ExtractorValidationFails", func(t *testing.T) {
		failingExtractor := &mockFailingExtractor{}
		analyzer, err := NewEFIJumpstartAnalyzer(validData, blockSize, failingExtractor)
		require.NoError(t, err)
		verifyErr := analyzer.VerifyEFIJumpstart()
		require.Error(t, verifyErr)
		assert.Contains(t, verifyErr.Error(), "driver validation/readability check failed")
		assert.Contains(t, verifyErr.Error(), "mock validate error")
	})
}

func TestJumpstartAnalyzerImpl_AnalyzeEFIJumpstart(t *testing.T) {
	blockSize := uint32(4096)
	paddr1 := types.Paddr(100)
	paddr2 := types.Paddr(200)
	extents := []types.Prange{
		{PrStartPaddr: paddr1, PrBlockCount: 2},
		{PrStartPaddr: paddr2, PrBlockCount: 1},
	}
	driverLen := uint32(calculatedSize(extents, blockSize))
	validData := types.NxEfiJumpstartT{
		NejMagic:      types.NxEfiJumpstartMagic,
		NejVersion:    types.NxEfiJumpstartVersion,
		NejEfiFileLen: driverLen,
		NejNumExtents: uint32(len(extents)),
		NejRecExtents: extents,
	}
	mockDriverContentWithMZ := make([]byte, driverLen)
	mockDriverContentWithMZ[0], mockDriverContentWithMZ[1] = 'M', 'Z'
	for i := 2; i < len(mockDriverContentWithMZ); i++ {
		mockDriverContentWithMZ[i] = byte(i % 256)
	}
	mockDriverContentNoMZ := bytes.Repeat([]byte{0xFF}, int(driverLen))
	maxPaddr := paddr2 + 1
	mockDiskSize := uint64(maxPaddr) * uint64(blockSize)
	mockDiskWithMZ := make([]byte, mockDiskSize)
	mockDiskNoMZ := make([]byte, mockDiskSize)
	offset1 := uint64(paddr1) * uint64(blockSize)
	offset2 := uint64(paddr2) * uint64(blockSize)
	copy(mockDiskWithMZ[offset1:offset1+8192], mockDriverContentWithMZ[0:8192])
	copy(mockDiskWithMZ[offset2:offset2+4096], mockDriverContentWithMZ[8192:12288])
	copy(mockDiskNoMZ[offset1:offset1+8192], mockDriverContentNoMZ[0:8192])
	copy(mockDiskNoMZ[offset2:offset2+4096], mockDriverContentNoMZ[8192:12288])
	readerWithMZ := bytes.NewReader(mockDiskWithMZ)
	readerNoMZ := bytes.NewReader(mockDiskNoMZ)
	extractorWithMZ, _ := NewEFIDriverExtractor(validData, readerWithMZ, blockSize)
	extractorNoMZ, _ := NewEFIDriverExtractor(validData, readerNoMZ, blockSize)

	t.Run("Success_FullAnalysis", func(t *testing.T) {
		analyzer, err := NewEFIJumpstartAnalyzer(validData, blockSize, extractorWithMZ)
		require.NoError(t, err)
		analysis, analyzeErr := analyzer.AnalyzeEFIJumpstart()
		require.NoError(t, analyzeErr)
		assert.True(t, analysis.IsValid)
		assert.Equal(t, types.NxEfiJumpstartVersion, analysis.Version)
		assert.Equal(t, driverLen, analysis.DriverSize)
		assert.Equal(t, uint32(len(extents)), analysis.ExtentCount)
		require.Len(t, analysis.ExtentDetails, 2)
		assert.Equal(t, paddr1, analysis.ExtentDetails[0].StartAddress)
		assert.Equal(t, uint64(2), analysis.ExtentDetails[0].BlockCount)
		assert.Equal(t, uint64(2*blockSize), analysis.ExtentDetails[0].SizeInBytes)
		assert.Equal(t, paddr2, analysis.ExtentDetails[1].StartAddress)
		assert.Equal(t, uint64(1), analysis.ExtentDetails[1].BlockCount)
		assert.Equal(t, uint64(1*blockSize), analysis.ExtentDetails[1].SizeInBytes)
		require.NotNil(t, analysis.DriverInfo)
		assert.Equal(t, "Yes", analysis.DriverInfo["DriverReadable"])
		assert.Equal(t, "Yes", analysis.DriverInfo["MZHeaderFound"])
		assert.NotContains(t, analysis.DriverInfo, "ConsistencyWarning")
		assert.NotContains(t, analysis.DriverInfo, "ReadLengthConsistency")
		_, hasCoverage := analysis.DriverInfo["ExtentCoverage"]
		assert.False(t, hasCoverage)
	})

	t.Run("Success_NoMZHeader", func(t *testing.T) {
		analyzer, err := NewEFIJumpstartAnalyzer(validData, blockSize, extractorNoMZ)
		require.NoError(t, err)
		analysis, analyzeErr := analyzer.AnalyzeEFIJumpstart()
		require.NoError(t, analyzeErr)
		assert.True(t, analysis.IsValid)
		require.NotNil(t, analysis.DriverInfo)
		assert.Equal(t, "Yes", analysis.DriverInfo["DriverReadable"])
		assert.Equal(t, "No", analysis.DriverInfo["MZHeaderFound"])
	})

	t.Run("Analysis_InvalidJumpstart", func(t *testing.T) {
		invalidMagic := validData
		invalidMagic.NejMagic = 0xDEADBEEF
		analyzer, err := NewEFIJumpstartAnalyzer(invalidMagic, blockSize, nil)
		require.NoError(t, err)
		analysis, analyzeErr := analyzer.AnalyzeEFIJumpstart()
		require.NoError(t, analyzeErr)
		assert.False(t, analysis.IsValid)
		assert.Equal(t, invalidMagic.NejVersion, analysis.Version)
		assert.Equal(t, invalidMagic.NejEfiFileLen, analysis.DriverSize)
		assert.Equal(t, invalidMagic.NejNumExtents, analysis.ExtentCount)
		assert.Len(t, analysis.ExtentDetails, len(invalidMagic.NejRecExtents))
		assert.Equal(t, "N/A (Extractor not provided)", analysis.DriverInfo["DriverContent"])
	})

	t.Run("Analysis_ExtractorFailsRead", func(t *testing.T) {
		failingExtractor := &mockFailingExtractor{}
		analyzer, err := NewEFIJumpstartAnalyzer(validData, blockSize, failingExtractor)
		require.NoError(t, err)
		analysis, analyzeErr := analyzer.AnalyzeEFIJumpstart()
		require.NoError(t, analyzeErr)
		assert.True(t, analysis.IsValid)
		require.NotNil(t, analysis.DriverInfo)
		assert.Equal(t, "No", analysis.DriverInfo["DriverReadable"])
		assert.Equal(t, "mock get data error", analysis.DriverInfo["ReadError"])
		_, mzFound := analysis.DriverInfo["MZHeaderFound"]
		assert.False(t, mzFound)
	})

	t.Run("Analysis_NilExtractor", func(t *testing.T) {
		analyzer, err := NewEFIJumpstartAnalyzer(validData, blockSize, nil)
		require.NoError(t, err)
		analysis, analyzeErr := analyzer.AnalyzeEFIJumpstart()
		require.NoError(t, analyzeErr)
		assert.True(t, analysis.IsValid)
		assert.Len(t, analysis.ExtentDetails, 2)
		require.NotNil(t, analysis.DriverInfo)
		assert.Equal(t, "N/A (Extractor not provided)", analysis.DriverInfo["DriverContent"])
		_, readable := analysis.DriverInfo["DriverReadable"]
		assert.False(t, readable)
		_, mzFound := analysis.DriverInfo["MZHeaderFound"]
		assert.False(t, mzFound)
	})

	t.Run("Analysis_SizeConsistencyWarning", func(t *testing.T) {
		inconsistentData := validData
		inconsistentData.NejEfiFileLen = uint32(calculatedSize(extents, blockSize) + 100)
		analyzer, err := NewEFIJumpstartAnalyzer(inconsistentData, blockSize, nil)
		require.NoError(t, err)
		analysis, analyzeErr := analyzer.AnalyzeEFIJumpstart()
		require.NoError(t, analyzeErr)
		require.NotNil(t, analysis.DriverInfo)
		assert.Contains(t, analysis.DriverInfo["ConsistencyWarning"], "LARGER than total size")
	})

	t.Run("Analysis_ExtentCoverageInfo", func(t *testing.T) {
		inconsistentData := validData
		inconsistentData.NejEfiFileLen = uint32(calculatedSize(extents, blockSize) - 100)
		analyzer, err := NewEFIJumpstartAnalyzer(inconsistentData, blockSize, nil)
		require.NoError(t, err)
		analysis, analyzeErr := analyzer.AnalyzeEFIJumpstart()
		require.NoError(t, analyzeErr)
		require.NotNil(t, analysis.DriverInfo)
		assert.Contains(t, analysis.DriverInfo["ExtentCoverage"], "Extents cover")
		_, hasWarning := analysis.DriverInfo["ConsistencyWarning"]
		assert.False(t, hasWarning)
	})
}

func calculatedSize(extents []types.Prange, blockSize uint32) uint64 {
	var total uint64
	for _, e := range extents {
		total += e.PrBlockCount * uint64(blockSize)
	}
	return total
}
