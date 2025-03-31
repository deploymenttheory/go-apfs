// File: pkg/util/checksum_test.go
package util

import (
	"testing"
)

func TestUUIDStringRoundTrip(t *testing.T) {
	original := "00112233-4455-6677-8899-aabbccddeeff"
	uuid, err := ParseUUID(original)
	if err != nil {
		t.Fatalf("ParseUUID failed: %v", err)
	}

	result := uuid.String()
	reparsed, err := ParseUUID(result)
	if err != nil {
		t.Fatalf("ParseUUID failed after round-trip: %v", err)
	}

	if !uuid.Equal(reparsed) {
		t.Errorf("UUID round-trip failed: original %v != reparsed %v", uuid, reparsed)
	}
}

func TestUUIDEquality(t *testing.T) {
	u1, _ := ParseUUID("deadbeef-0001-2345-6789-abcdefabcdef")
	u2, _ := ParseUUID("deadbeef-0001-2345-6789-abcdefabcdef")
	u3, _ := ParseUUID("11112222-3333-4444-5555-666677778888")

	if !u1.Equal(u2) {
		t.Errorf("Expected UUIDs to be equal: %v != %v", u1, u2)
	}
	if u1.Equal(u3) {
		t.Errorf("Expected UUIDs to be different: %v == %v", u1, u3)
	}
}

func TestUUIDParseErrors(t *testing.T) {
	bad := []string{
		"this-is-not-a-uuid",
		"1234",
		"zzzzzzzz-zzzz-zzzz-zzzz-zzzzzzzzzzzz",
		"00112233445566778899aabbccddeeff0011", // 34 chars
	}

	for _, s := range bad {
		_, err := ParseUUID(s)
		if err == nil {
			t.Errorf("Expected error for malformed UUID: %s", s)
		}
	}
}
