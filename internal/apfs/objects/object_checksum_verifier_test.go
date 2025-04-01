package objects

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

func generateFletcher64Checksum(data []byte) [types.MaxCksumSize]byte {
	var sum1, sum2 uint64
	for i := 0; i < len(data); i += 8 {
		word := binary.LittleEndian.Uint64(data[i : i+8])
		sum1 += word
		sum2 += sum1
	}
	var result [types.MaxCksumSize]byte
	binary.LittleEndian.PutUint64(result[:], sum2)
	return result
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
