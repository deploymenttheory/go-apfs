package objects

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

// generateFletcher64Checksum generates a correct Fletcher-64 checksum for APFS
// This must match the algorithm in object_checksum_verifier.go
func generateFletcher64Checksum(data []byte) [types.MaxCksumSize]byte {
	const maxUint32 = uint64(0xFFFFFFFF)
	const chunkSize = 1024 // Process 1024 words (4096 bytes) at a time

	var sum1, sum2 uint64

	// Process data in chunks of 32-bit words
	for offset := 0; offset < len(data); offset += chunkSize * 4 {
		// Determine the end of this chunk
		chunkEnd := offset + chunkSize*4
		if chunkEnd > len(data) {
			chunkEnd = len(data)
		}

		// Process 32-bit words in this chunk
		for i := offset; i < chunkEnd; i += 4 {
			// Ensure we don't read past the end
			if i+4 > len(data) {
				break
			}

			word := binary.LittleEndian.Uint32(data[i : i+4])
			sum1 += uint64(word)
			sum2 += sum1
		}

		// Apply modulo operations after each chunk
		sum1 %= maxUint32
		sum2 %= maxUint32
	}

	// Calculate the value needed to be able to get a checksum of zero
	ckLow := maxUint32 - ((sum1 + sum2) % maxUint32)
	ckHigh := maxUint32 - ((sum1 + ckLow) % maxUint32)

	// Combine into final 64-bit checksum
	result := ckLow | (ckHigh << 32)

	var checksum [types.MaxCksumSize]byte
	binary.LittleEndian.PutUint64(checksum[:], result)
	return checksum
}

func TestChecksumAndVerify(t *testing.T) {
	// Create a fake 64-byte payload (multiple of 8)
	payload := make([]byte, 64)
	for i := range payload {
		payload[i] = byte(i)
	}

	// Zero out first 8 bytes for checksum position
	copy(payload[0:8], make([]byte, 8))

	checksum := generateFletcher64Checksum(payload)
	copy(payload[0:8], checksum[:])

	// Parse the object from payload
	var obj types.ObjPhysT
	reader := bytes.NewReader(payload)
	if err := binary.Read(reader, binary.LittleEndian, &obj); err != nil {
		t.Fatalf("Failed to read ObjPhysT from payload: %v", err)
	}

	inspector := NewChecksumInspector(&obj, payload)

	if got := inspector.Checksum(); got != checksum {
		t.Errorf("Checksum() mismatch. Got %v, expected %v", got, checksum)
	}

	if !inspector.VerifyChecksum() {
		t.Error("Expected VerifyChecksum() to return true")
	}

	// Corrupt the payload
	payload[16] ^= 0xFF
	inspector = NewChecksumInspector(&obj, payload)

	if inspector.VerifyChecksum() {
		t.Error("Expected VerifyChecksum() to return false after corruption")
	}
}
