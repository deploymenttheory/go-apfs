// File: pkg/checksum/checksum_test.go
package checksum

import "testing"

func TestFletcher64(t *testing.T) {
	data := make([]byte, 4096)
	for i := 0; i < len(data); i++ {
		data[i] = byte(i % 256)
	}

	sum := Fletcher64(data)
	if sum == 0 {
		t.Errorf("Fletcher64 returned zero checksum unexpectedly")
	}

	if !ValidateFletcher64(data, sum) {
		t.Errorf("ValidateFletcher64 failed for known-good checksum")
	}

	// Corrupt one byte and verify it fails validation
	data[0] ^= 0xFF
	if ValidateFletcher64(data, sum) {
		t.Errorf("ValidateFletcher64 passed for corrupted data")
	}
}
