package services

import (
	"fmt"
	"unicode"
)

// NameHashingService provides name hashing utilities for APFS directory operations
type NameHashingService struct{}

// NewNameHashingService creates a new name hashing service
func NewNameHashingService() *NameHashingService {
	return &NameHashingService{}
}

// HashUTF8 calculates the APFS name hash for a UTF-8 string
func (nhs *NameHashingService) HashUTF8(name []byte, useCaseFolding bool) (uint32, error) {
	if len(name) == 0 {
		return 0, nil
	}

	var hash uint32 = 5381

	if useCaseFolding {
		// Decode UTF-8 and apply case folding
		runes := []rune{}
		for len(name) > 0 {
			r, size := nhs.decodeUTF8Rune(name)
			if r == unicode.ReplacementChar && size == 1 {
				return 0, fmt.Errorf("invalid UTF-8 sequence")
			}
			runes = append(runes, unicode.ToLower(r))
			name = name[size:]
		}

		// Hash the case-folded runes
		for _, r := range runes {
			encoded := nhs.encodeUTF8Rune(r)
			for _, b := range encoded {
				if b == 0 {
					break
				}
				hash = ((hash << 5) + hash) + uint32(b)
			}
		}
	} else {
		// Hash UTF-8 bytes directly
		for _, b := range name {
			hash = ((hash << 5) + hash) + uint32(b)
		}
	}

	return hash, nil
}

// HashUTF16 calculates the APFS name hash for a UTF-16 string
func (nhs *NameHashingService) HashUTF16(name []uint16, useCaseFolding bool) uint32 {
	if len(name) == 0 {
		return 0
	}

	var hash uint32 = 5381

	if useCaseFolding {
		// Decode UTF-16 and apply case folding
		runes := nhs.decodeUTF16(name)
		for _, r := range runes {
			r = unicode.ToLower(r)
			encoded := nhs.encodeUTF8Rune(r)
			for _, b := range encoded {
				if b == 0 {
					break
				}
				hash = ((hash << 5) + hash) + uint32(b)
			}
		}
	} else {
		// Hash UTF-16 code units
		for _, codeUnit := range name {
			hash = ((hash << 5) + hash) + uint32(codeUnit&0xFF)
			hash = ((hash << 5) + hash) + uint32((codeUnit>>8)&0xFF)
		}
	}

	return hash
}

// Private helper methods

func (nhs *NameHashingService) decodeUTF8Rune(b []byte) (rune, int) {
	if len(b) == 0 {
		return unicode.ReplacementChar, 0
	}

	c0 := b[0]
	if c0 < 0x80 {
		return rune(c0), 1
	}

	if c0&0xE0 == 0xC0 && len(b) >= 2 {
		c1 := b[1]
		if nhs.isUTF8Continuation(c1) {
			r := ((rune(c0) & 0x1F) << 6) | rune(c1&0x3F)
			return r, 2
		}
	}

	if c0&0xF0 == 0xE0 && len(b) >= 3 {
		c1, c2 := b[1], b[2]
		if nhs.isUTF8Continuation(c1) && nhs.isUTF8Continuation(c2) {
			r := ((rune(c0) & 0x0F) << 12) | ((rune(c1) & 0x3F) << 6) | rune(c2&0x3F)
			return r, 3
		}
	}

	if c0&0xF8 == 0xF0 && len(b) >= 4 {
		c1, c2, c3 := b[1], b[2], b[3]
		if nhs.isUTF8Continuation(c1) && nhs.isUTF8Continuation(c2) && nhs.isUTF8Continuation(c3) {
			r := ((rune(c0) & 0x07) << 18) | ((rune(c1) & 0x3F) << 12) | ((rune(c2) & 0x3F) << 6) | rune(c3&0x3F)
			return r, 4
		}
	}

	return unicode.ReplacementChar, 1
}

func (nhs *NameHashingService) isUTF8Continuation(b byte) bool {
	return (b & 0xC0) == 0x80
}

func (nhs *NameHashingService) encodeUTF8Rune(r rune) [4]byte {
	var result [4]byte

	if r < 0x80 {
		result[0] = byte(r)
	} else if r < 0x800 {
		result[0] = byte(0xC0 | (r >> 6))
		result[1] = byte(0x80 | (r & 0x3F))
	} else if r < 0x10000 {
		result[0] = byte(0xE0 | (r >> 12))
		result[1] = byte(0x80 | ((r >> 6) & 0x3F))
		result[2] = byte(0x80 | (r & 0x3F))
	} else if r < 0x110000 {
		result[0] = byte(0xF0 | (r >> 18))
		result[1] = byte(0x80 | ((r >> 12) & 0x3F))
		result[2] = byte(0x80 | ((r >> 6) & 0x3F))
		result[3] = byte(0x80 | (r & 0x3F))
	}

	return result
}

func (nhs *NameHashingService) decodeUTF16(utf16String []uint16) []rune {
	var runes []rune
	i := 0

	for i < len(utf16String) {
		c := utf16String[i]

		// Check for surrogate pair
		if c >= 0xD800 && c <= 0xDBFF && i+1 < len(utf16String) {
			low := utf16String[i+1]
			if low >= 0xDC00 && low <= 0xDFFF {
				// Valid surrogate pair
				high := c - 0xD800
				low -= 0xDC00
				r := 0x10000 + (rune(high)<<10 | rune(low))
				runes = append(runes, r)
				i += 2
				continue
			}
		}

		// Single UTF-16 code unit
		runes = append(runes, rune(c))
		i++
	}

	return runes
}
