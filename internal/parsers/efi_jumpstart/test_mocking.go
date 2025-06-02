// File: internal/efijumpstart/test_mocking.go
package efijumpstart

import (
	"errors"
	"fmt"
	"io"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

// --- Test Suite Setup ---
const (
	testBlockSize      = 4096
	testBlock0Offset   = 0
	testBlock1Offset   = testBlockSize
	testBlock2Offset   = 2 * testBlockSize
	testJumpstartPaddr = types.Paddr(42)
	testNxMagic        = 0x4253584E // "NXSB" in little-endian (matches nxMagic in the main code)
	// Define the correct magic number for APFS container superblocks
	// NXSB in hex: 4E 58 53 42 (big endian)
	// or 42 53 58 4E (little endian, as stored on disk)
	containerMagic = 0x4E585342
)

// mockReaderAt simulates reading from a source, allowing for errors.
// Ensure this mock correctly handles reads within its data bounds.
type mockReaderAt struct {
	data        []byte
	failAfterN  int
	failOnRead  bool
	readCounter int
	errToReturn error
}

func (m *mockReaderAt) ReadAt(p []byte, off int64) (n int, err error) {
	m.readCounter++
	if m.failOnRead && m.readCounter > m.failAfterN {
		return 0, m.errToReturn
	}
	if off < 0 {
		return 0, fmt.Errorf("negative offset %d", off) // Use fmt
	}
	if off >= int64(len(m.data)) {
		return 0, io.EOF // Use io
	}
	end := off + int64(len(p))
	if end > int64(len(m.data)) {
		end = int64(len(m.data))
	}
	n = copy(p, m.data[off:end])
	// ReadAt contract: return error if n < len(p), unless EOF was hit exactly at the end.
	if n < len(p) && int64(n) != int64(len(m.data))-off { // Check if it wasn't just reading the last few bytes
		return n, io.ErrUnexpectedEOF // Use io
	}
	if n < len(p) && int64(n) == int64(len(m.data))-off { // Reached end exactly
		return n, io.EOF // Return EOF if less than requested *and* hit end of data
	}
	return n, nil
}

// mock Failing Extractor
type mockFailingExtractor struct{}

func (m *mockFailingExtractor) ExtractEFIDriver(outputPath string) error {
	return errors.New("mock extract error")
}
func (m *mockFailingExtractor) GetEFIDriverData() ([]byte, error) {
	return nil, errors.New("mock get data error")
}
func (m *mockFailingExtractor) ValidateEFIDriver() error { return errors.New("mock validate error") }
