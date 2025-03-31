// File: pkg/util/uuid.go
package util

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
)

// UUID represents a 16-byte universally unique identifier as used in APFS
// APFS stores UUIDs in little-endian byte order
// The fields are grouped for parsing purposes: time_low, time_mid, time_hi, clock_seq, node
// Reference: Apple File System Reference (UUID fields in volume/snapshot structures)
type UUID [16]byte

// String returns the UUID in canonical 8-4-4-4-12 format
func (u UUID) String() string {
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		binary.BigEndian.Uint32(u[0:4]),
		binary.BigEndian.Uint16(u[4:6]),
		binary.BigEndian.Uint16(u[6:8]),
		binary.BigEndian.Uint16(u[8:10]),
		u[10:16],
	)
}

// Equal reports whether two UUIDs are equal
func (u UUID) Equal(other UUID) bool {
	return u == other
}

// ParseUUID parses a 16-byte UUID from a string in canonical format (with or without dashes)
func ParseUUID(s string) (UUID, error) {
	var uuid UUID
	s = removeDashes(s)
	if len(s) != 32 {
		return uuid, fmt.Errorf("invalid UUID string length: %d", len(s))
	}
	decoded, err := hex.DecodeString(s)
	if err != nil {
		return uuid, err
	}
	copy(uuid[:], decoded[:16])
	return uuid, nil
}

// removeDashes strips dashes from a UUID string
func removeDashes(s string) string {
	buf := make([]byte, 0, 32)
	for i := 0; i < len(s); i++ {
		if s[i] != '-' {
			buf = append(buf, s[i])
		}
	}
	return string(buf)
}
