// File: internal/partitionmanager/efi_partition_manager.go
package efijumpstart

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log" // Keep if you want the log.Printf for Sync error
	"strings"
	"unicode/utf16"

	"github.com/deploymenttheory/go-apfs/internal/interfaces" // Adjust import path
	"github.com/deploymenttheory/go-apfs/internal/types"      // Adjust import path
)

const (
	// Standard GPT Header signature "EFI PART"
	gptHeaderSignature = 0x5452415020494645
	// Standard EFI System Partition (ESP) Type GUID
	efiSystemPartitionGUIDString = "C12A7328-F81F-11D2-BA4B-00A0C93EC93B"
	// Size of a standard GPT Header struct (up to Header CRC32)
	gptHeaderSize = 92
	// Size of a standard GPT Partition Entry
	gptPartitionEntrySize = 128 // <<< Keep this constant
)

// gptHeader represents the relevant fields from the GPT Header.
// Based on UEFI Specification 2.10, Section 5.3.2
type gptHeader struct {
	Signature                uint64   // Offset 0
	Revision                 uint32   // Offset 8
	HeaderSize               uint32   // Offset 12
	HeaderCRC32              uint32   // Offset 16 (unverified in this impl)
	Reserved                 uint32   // Offset 20
	MyLBA                    uint64   // Offset 24
	AlternateLBA             uint64   // Offset 32
	FirstUsableLBA           uint64   // Offset 40
	LastUsableLBA            uint64   // Offset 48
	DiskGUID                 [16]byte // Offset 56
	PartitionEntryLBA        uint64   // Offset 72: Starting LBA of partition entries array
	NumberOfPartitionEntries uint32   // Offset 80: Number of entries in the array
	SizeOfPartitionEntry     uint32   // Offset 84: Size of each entry (usually 128)
	PartitionEntryArrayCRC32 uint32   // Offset 88 (unverified in this impl)
	// Remainder is reserved padding
}

// gptPartitionEntry represents the relevant fields from a GPT Partition Entry.
// Based on UEFI Specification 2.10, Section 5.3.3
type gptPartitionEntry struct {
	PartitionTypeGUID   [16]byte // Offset 0
	UniquePartitionGUID [16]byte // Offset 16
	FirstLBA            uint64   // Offset 32
	LastLBA             uint64   // Offset 40
	Attributes          uint64   // Offset 48
	PartitionName       [72]byte // Offset 56: UTF-16LE (36 characters)
}

// GPTPartitionManager implements the EFIPartitionManager interface using GPT.
// It reads GPT data to provide information about EFI-relevant partitions.
type GPTPartitionManager struct {
	device            io.ReaderAt // Reader for the raw disk/image
	logicalBlockSize  uint64      // Usually 512
	apfsPartitionUUID string      // The specific APFS partition this instance might be tied to (for GetPartitionUUID)
}

// Compile-time check
var _ interfaces.EFIPartitionManager = (*GPTPartitionManager)(nil)

// NewGPTPartitionManager creates a new GPT-based partition manager.
func NewGPTPartitionManager(device io.ReaderAt, logicalBlockSize uint64, associatedAPFSPartitionUUID string) (*GPTPartitionManager, error) {
	if device == nil {
		return nil, fmt.Errorf("device reader cannot be nil")
	}
	if logicalBlockSize == 0 {
		return nil, fmt.Errorf("logical block size cannot be zero")
	}
	return &GPTPartitionManager{
		device:            device,
		logicalBlockSize:  logicalBlockSize,
		apfsPartitionUUID: associatedAPFSPartitionUUID,
	}, nil
}

// GetPartitionUUID returns the UUID of the APFS partition associated with this manager instance.
func (m *GPTPartitionManager) GetPartitionUUID() string {
	// Return the specific APFS partition UUID this instance was configured with, if any.
	// This assumes the caller provided the relevant UUID during creation.
	// A more complex implementation might scan the GPT to find *an* APFS partition if one wasn't provided.
	return m.apfsPartitionUUID
}

// IsAPFSPartition checks if a given UUID string matches the known APFS Partition Type GUID.
// Comparison is case-insensitive.
func (m *GPTPartitionManager) IsAPFSPartition(partitionUUID string) bool {
	return strings.EqualFold(partitionUUID, types.ApfsGptPartitionUUID)
}

// ListEFIPartitions scans the primary GPT and returns information about all found EFI System Partitions.
func (m *GPTPartitionManager) ListEFIPartitions() ([]interfaces.EFIPartitionInfo, error) {
	header, err := m.readGPTHeader()
	if err != nil {
		return nil, fmt.Errorf("failed to read GPT header: %w", err)
	}

	// Basic validation
	if header.Signature != gptHeaderSignature {
		return nil, fmt.Errorf("invalid GPT signature: expected %X, got %X", gptHeaderSignature, header.Signature)
	}
	if header.SizeOfPartitionEntry == 0 || header.NumberOfPartitionEntries == 0 {
		return []interfaces.EFIPartitionInfo{}, nil // No entries to process
	}
	// *** USE THE CONSTANT FOR VALIDATION ***
	if header.SizeOfPartitionEntry != gptPartitionEntrySize {
		// While technically allowed by spec to vary, it's almost always 128.
		// Log a warning or return an error if it differs, as our parsing assumes 128.
		log.Printf("Warning: GPT header reports SizeOfPartitionEntry as %d, expected %d. Parsing might be incorrect.", header.SizeOfPartitionEntry, gptPartitionEntrySize)
		// Decide whether to proceed or error out. Let's proceed with caution for now.
		// return nil, fmt.Errorf("unsupported GPT SizeOfPartitionEntry: got %d, expected %d", header.SizeOfPartitionEntry, gptPartitionEntrySize)
	}

	// Read partition entries array
	entriesOffset := header.PartitionEntryLBA * m.logicalBlockSize
	// Use the size reported by the header for reading, but parse based on the constant size
	bytesToReadForEntries := uint64(header.NumberOfPartitionEntries) * uint64(header.SizeOfPartitionEntry)
	entriesData := make([]byte, bytesToReadForEntries)

	n, err := m.device.ReadAt(entriesData, int64(entriesOffset))
	if err != nil && err != io.EOF { // EOF might be okay if disk ends exactly after entries
		return nil, fmt.Errorf("failed to read partition entry array at offset %d: %w", entriesOffset, err)
	}
	// Check if we read enough data based on the number of entries * expected size
	expectedDataSizeFromConstant := uint64(header.NumberOfPartitionEntries) * uint64(gptPartitionEntrySize)
	if uint64(n) < expectedDataSizeFromConstant {
		return nil, fmt.Errorf("short read for partition entry array: read %d bytes, expected at least %d bytes for %d entries", n, expectedDataSizeFromConstant, header.NumberOfPartitionEntries)
	}

	var efiPartitions []interfaces.EFIPartitionInfo
	entryReader := bytes.NewReader(entriesData)

	// Parse each entry using the constant size for safety
	for i := uint32(0); i < header.NumberOfPartitionEntries; i++ {
		entryBytes := make([]byte, gptPartitionEntrySize) // Use constant here
		// Ensure we don't read past the end of our buffer 'entriesData'
		if _, err := io.ReadFull(entryReader, entryBytes); err != nil {
			// Check if it's EOF because we correctly read all expected entries based on 'n'
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				log.Printf("Warning: Reached end of partition entry data buffer unexpectedly after entry %d.", i)
				break // Stop processing entries
			}
			return nil, fmt.Errorf("failed to read bytes for partition entry %d: %w", i, err)
		}

		var entry gptPartitionEntry
		entryParseReader := bytes.NewReader(entryBytes)
		// Parse based on the structure which implicitly uses the size of gptPartitionEntry
		if err := binary.Read(entryParseReader, binary.LittleEndian, &entry); err != nil {
			return nil, fmt.Errorf("failed to parse partition entry %d: %w", i, err)
		}

		// Check for empty entry (all zeros Type GUID)
		isEmpty := true
		for _, b := range entry.PartitionTypeGUID {
			if b != 0 {
				isEmpty = false
				break
			}
		}
		if isEmpty {
			continue // Skip unused entries
		}

		// Check if it's an EFI System Partition by Type GUID
		entryTypeGUIDStr := formatGUID(entry.PartitionTypeGUID)
		if strings.EqualFold(entryTypeGUIDStr, efiSystemPartitionGUIDString) {
			partInfo := interfaces.EFIPartitionInfo{
				UUID:   entryTypeGUIDStr,
				Name:   decodeUTF16LE(entry.PartitionName[:]),
				Offset: entry.FirstLBA * m.logicalBlockSize,
				Size:   (entry.LastLBA - entry.FirstLBA + 1) * m.logicalBlockSize,
			}
			efiPartitions = append(efiPartitions, partInfo)
		}
	}

	return efiPartitions, nil
}

// readGPTHeader reads and parses the primary GPT header (LBA 1).
func (m *GPTPartitionManager) readGPTHeader() (*gptHeader, error) {
	headerOffset := int64(1 * m.logicalBlockSize) // GPT Header is at LBA 1
	headerData := make([]byte, gptHeaderSize)

	n, err := m.device.ReadAt(headerData, headerOffset)
	if err != nil {
		return nil, fmt.Errorf("failed to read GPT header data at offset %d: %w", headerOffset, err)
	}
	if n != gptHeaderSize {
		return nil, fmt.Errorf("short read for GPT header: read %d bytes, expected %d", n, gptHeaderSize)
	}

	var header gptHeader
	reader := bytes.NewReader(headerData)
	if err := binary.Read(reader, binary.LittleEndian, &header); err != nil {
		return nil, fmt.Errorf("failed to parse GPT header binary data: %w", err)
	}

	return &header, nil
}

// formatGUID converts a 16-byte slice (as stored in GPT) to a standard UUID string.
func formatGUID(g [16]byte) string {
	return fmt.Sprintf("%02X%02X%02X%02X-%02X%02X-%02X%02X-%02X%02X-%02X%02X%02X%02X%02X%02X",
		g[3], g[2], g[1], g[0], g[5], g[4], g[7], g[6], g[8], g[9], g[10], g[11], g[12], g[13], g[14], g[15])
}

// decodeUTF16LE decodes a UTF-16 Little Endian byte slice into a Go string, stopping at the first null character (0x0000).
func decodeUTF16LE(b []byte) string {
	if len(b)%2 != 0 {
		return ""
	}
	u16s := make([]uint16, 0, len(b)/2)
	for i := 0; i+1 < len(b); i += 2 {
		val := binary.LittleEndian.Uint16(b[i : i+2])
		if val == 0 {
			break // Stop at first null rune
		}
		u16s = append(u16s, val)
	}
	return string(utf16.Decode(u16s))
}
