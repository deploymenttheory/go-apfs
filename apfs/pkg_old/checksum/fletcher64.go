// File: pkg/checksum/checksum.go
package checksum

import "encoding/binary"

// Fletcher64 computes the Fletcher64 checksum over 64-bit words (Apple spec).
func Fletcher64(data []byte) uint64 {
	var sum1, sum2 uint64

	n := len(data) / 8
	for i := 0; i < n; i++ {
		word := binary.LittleEndian.Uint64(data[i*8:])
		sum1 += word
		sum2 += sum1
	}

	return (sum2 << 32) | (sum1 & 0xffffffff)
}

// ValidateFletcher64 compares the computed checksum to the expected one
func ValidateFletcher64(data []byte, expected uint64) bool {
	return Fletcher64(data) == expected
}

// Fletcher64WithZeroedChecksum zeros the first 8 bytes before computing checksum
func Fletcher64WithZeroedChecksum(data []byte, offset int) uint64 {
	tmp := make([]byte, len(data))
	copy(tmp, data)

	if offset+8 <= len(tmp) {
		for i := 0; i < 8; i++ {
			tmp[offset+i] = 0
		}
	}

	return Fletcher64(tmp)
}
