package services

import (
	"fmt"
	"io"
)

// Read implements io.Reader for FileReaderAdapter
func (fra *FileReaderAdapter) Read(p []byte) (n int, err error) {
	if fra.offset >= fra.size {
		return 0, io.EOF
	}

	// Determine how much to read
	toRead := uint64(len(p))
	if fra.offset+toRead > fra.size {
		toRead = fra.size - fra.offset
	}

	// Read from the file
	data, err := fra.fs.ReadFileRange(fra.inodeID, fra.offset, toRead)
	if err != nil {
		return 0, fmt.Errorf("failed to read file range: %w", err)
	}

	// Copy to output buffer
	n = copy(p, data)
	fra.offset += uint64(n)

	if fra.offset >= fra.size && n == 0 {
		err = io.EOF
	}

	return n, err
}

// Read implements io.Reader for FileSeekerAdapter
func (fsa *FileSeekerAdapter) Read(p []byte) (n int, err error) {
	if fsa.offset >= fsa.size {
		return 0, io.EOF
	}

	// Determine how much to read
	toRead := uint64(len(p))
	if fsa.offset+toRead > fsa.size {
		toRead = fsa.size - fsa.offset
	}

	// Read from the file
	data, err := fsa.fs.ReadFileRange(fsa.inodeID, fsa.offset, toRead)
	if err != nil {
		return 0, fmt.Errorf("failed to read file range: %w", err)
	}

	// Copy to output buffer
	n = copy(p, data)
	fsa.offset += uint64(n)

	if fsa.offset >= fsa.size && n == 0 {
		err = io.EOF
	}

	return n, err
}

// Seek implements io.Seeker for FileSeekerAdapter
func (fsa *FileSeekerAdapter) Seek(offset int64, whence int) (int64, error) {
	var newOffset int64

	switch whence {
	case io.SeekStart:
		newOffset = offset
	case io.SeekCurrent:
		newOffset = int64(fsa.offset) + offset
	case io.SeekEnd:
		newOffset = int64(fsa.size) + offset
	default:
		return 0, fmt.Errorf("invalid whence: %d", whence)
	}

	if newOffset < 0 {
		return 0, fmt.Errorf("negative offset: %d", newOffset)
	}

	fsa.offset = uint64(newOffset)
	return newOffset, nil
}
