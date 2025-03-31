package types

import (
	"encoding/hex"
	"fmt"
	"strings"
)

// UUIDFromString parses a UUID string (e.g., "EBC1...") into a UUID struct
func UUIDFromString(s string) (UUID, error) {
	var uuid UUID

	s = strings.ReplaceAll(s, "-", "")
	if len(s) != 32 {
		return uuid, fmt.Errorf("invalid UUID string length: %d", len(s))
	}

	bytes, err := hex.DecodeString(s)
	if err != nil {
		return uuid, fmt.Errorf("failed to decode UUID: %w", err)
	}

	copy(uuid[:], bytes)
	return uuid, nil
}

// String returns the standard UUID string representation (with hyphens)
func (u UUID) String() string {
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		u[0:4], u[4:6], u[6:8], u[8:10], u[10:16])
}
