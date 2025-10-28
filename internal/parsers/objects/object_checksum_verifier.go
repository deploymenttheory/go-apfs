package objects

import (
	"encoding/binary"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

// ChecksumInspector implements ObjectChecksumVerifier for ObjPhysT
type ChecksumInspector struct {
	Obj     *types.ObjPhysT
	Payload []byte // full raw object data including header
}

func NewChecksumInspector(obj *types.ObjPhysT, payload []byte) *ChecksumInspector {
	return &ChecksumInspector{Obj: obj, Payload: payload}
}

func (c *ChecksumInspector) Checksum() [types.MaxCksumSize]byte {
	return c.Obj.OChecksum
}

func (c *ChecksumInspector) VerifyChecksum() bool {
	if len(c.Payload)%4 != 0 {
		// Fletcher64 operates on 32-bit words
		return false
	}

	// The checksum field is at the beginning, so zero it for computation
	payloadCopy := make([]byte, len(c.Payload))
	copy(payloadCopy, c.Payload)
	for i := 0; i < types.MaxCksumSize; i++ {
		payloadCopy[i] = 0
	}

	calculated := fletcher64(payloadCopy)
	return calculated == c.Obj.OChecksum
}

// fletcher64 implements the correct Fletcher-64 checksum algorithm for APFS
// Based on the blog analysis: uses 32-bit words, modulo operations, and proper chunking
func fletcher64(data []byte) [types.MaxCksumSize]byte {
	const maxUint32 = uint64(0xFFFFFFFF)
	const chunkSize = 1024 // Process 1024 words (4096 bytes) at a time for modulo

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

		// Apply modulo operations after each chunk to prevent overflow
		sum1 %= maxUint32
		sum2 %= maxUint32
	}

	// Final result: combine sum2 and sum1 into 64-bit checksum
	// sum2 in high 32 bits, sum1 in low 32 bits
	result := (sum2 << 32) | sum1

	var checksum [types.MaxCksumSize]byte
	binary.LittleEndian.PutUint64(checksum[:], result)
	return checksum
}
