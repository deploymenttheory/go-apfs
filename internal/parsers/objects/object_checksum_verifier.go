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
	if len(c.Payload)%8 != 0 {
		// Fletcher64 operates on 64-bit words
		return false
	}

	var sum1, sum2 uint64

	// The checksum field is at the beginning, so zero it for computation
	payloadCopy := make([]byte, len(c.Payload))
	copy(payloadCopy, c.Payload)
	for i := 0; i < types.MaxCksumSize; i++ {
		payloadCopy[i] = 0
	}

	for i := 0; i < len(payloadCopy); i += 8 {
		word := binary.LittleEndian.Uint64(payloadCopy[i : i+8])
		sum1 += word
		sum2 += sum1
	}

	calculated := [types.MaxCksumSize]byte{}
	binary.LittleEndian.PutUint64(calculated[:], sum2)

	return calculated == c.Obj.OChecksum
}
