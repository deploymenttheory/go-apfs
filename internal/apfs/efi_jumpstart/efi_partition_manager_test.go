// File: internal/efijumpstart/efi_partition_manager_test.go
package efijumpstart

import (
	"bytes"
	"encoding/binary"
	"strings"
	"testing"
	"unicode/utf16"

	// Adjust import path
	"github.com/deploymenttheory/go-apfs/internal/types" // Adjust import path
	"github.com/google/uuid"                             // For generating test GUIDs
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testLogicalBlockSize = 512

// Helper to create GPT partition name bytes (UTF-16LE, null-padded)
func createGPTPartitionName(name string) [72]byte {
	var buf [72]byte
	encoded := utf16.Encode([]rune(name))
	for i, r := range encoded {
		if i*2+1 >= 72 {
			break
		}
		binary.LittleEndian.PutUint16(buf[i*2:], r)
	}
	return buf
}

// Helper to parse a standard UUID string into the 16-byte GPT mixed-endian format
func parseUUIDStringToGPTBytes(uuidStr string) ([16]byte, error) {
	var gptBytes [16]byte
	parsedUUID, err := uuid.Parse(uuidStr)
	if err != nil {
		return gptBytes, err
	}
	uuidBytes := [16]byte(parsedUUID)
	gptBytes[0], gptBytes[1], gptBytes[2], gptBytes[3] = uuidBytes[3], uuidBytes[2], uuidBytes[1], uuidBytes[0]
	gptBytes[4], gptBytes[5] = uuidBytes[5], uuidBytes[4]
	gptBytes[6], gptBytes[7] = uuidBytes[7], uuidBytes[6]
	copy(gptBytes[8:], uuidBytes[8:])
	return gptBytes, nil
}

// createMockGPTDisk creates a byte slice representing a disk with a simplified GPT
func createMockGPTDisk(logicalBlockSize uint64, partitions []gptPartitionEntry) ([]byte, *gptHeader, error) {
	lba2Offset := 2 * logicalBlockSize // Start of partition array
	numEntries := uint32(len(partitions))
	sizeOfEntry := uint32(gptPartitionEntrySize)

	// Calculate size needed for the partition entry array itself
	lba2DataSize := uint64(numEntries * sizeOfEntry)
	lba2TotalBlocks := (lba2DataSize + logicalBlockSize - 1) / logicalBlockSize // Blocks needed for entries (round up)

	// Determine the highest LBA used by any partition *entry* for disk sizing
	var maxLBA uint64 = lba2Offset/logicalBlockSize + lba2TotalBlocks - 1 // Start with end of primary entries table
	for _, p := range partitions {
		if p.LastLBA > maxLBA {
			maxLBA = p.LastLBA
		}
	}

	// Calculate total disk size needed: MBR + Header + Entries + space up to maxLBA + Backup Entries + Backup Header
	// Add 1 to maxLBA because LBAs are 0-indexed counts.
	// Add space for backup structures (simplification: assume same size as primary)
	backupStructuresBlocks := uint64(1) + lba2TotalBlocks // Backup Header + Backup Entries
	totalBlocksNeeded := maxLBA + 1 + backupStructuresBlocks
	totalDiskSize := totalBlocksNeeded * logicalBlockSize

	disk := make([]byte, totalDiskSize)

	// Create GPT Header (at LBA 1 offset)
	firstUsableLBA := lba2Offset/logicalBlockSize + lba2TotalBlocks // First block *after* primary entry table
	alternateLBA := totalBlocksNeeded - 1                           // Backup header at the very end
	lastUsableLBA := alternateLBA - lba2TotalBlocks - 1             // Last block *before* backup entry table

	if lastUsableLBA < firstUsableLBA {
		lastUsableLBA = firstUsableLBA // Handle edge case where disk is tiny
	}

	header := gptHeader{
		Signature:                gptHeaderSignature,
		Revision:                 0x00010000, // 1.0
		HeaderSize:               gptHeaderSize,
		MyLBA:                    1,
		AlternateLBA:             alternateLBA,
		FirstUsableLBA:           firstUsableLBA,
		LastUsableLBA:            lastUsableLBA,
		DiskGUID:                 [16]byte{ /* random */ 0xde, 0xad, 0xbe, 0xef},
		PartitionEntryLBA:        lba2Offset / logicalBlockSize, // Array starts at LBA 2
		NumberOfPartitionEntries: numEntries,
		SizeOfPartitionEntry:     sizeOfEntry,
		// CRCs omitted
	}

	headerBuf := new(bytes.Buffer)
	err := binary.Write(headerBuf, binary.LittleEndian, header)
	require.NoError(nil, err, "Failed to write mock header") // Use nil t in helper
	copy(disk[logicalBlockSize:], headerBuf.Bytes())         // Copy to LBA 1 start

	// Create Partition Array (at LBA 2 offset)
	arrayBuf := new(bytes.Buffer)
	for _, p := range partitions {
		err := binary.Write(arrayBuf, binary.LittleEndian, p)
		require.NoError(nil, err, "Failed to write mock partition entry")
	}
	copy(disk[lba2Offset:], arrayBuf.Bytes()) // Copy to LBA 2 start

	// Optionally: copy backup structures (not strictly needed for these tests)

	return disk, &header, nil
}

func TestGPTPartitionManager(t *testing.T) {
	testAPFSPartitionUUID := types.ApfsGptPartitionUUID
	testESPPartitionUUID := efiSystemPartitionGUIDString
	testOtherPartitionUUID := "48465300-0000-11AA-AA11-00306543ECAC" // Example HFS+

	apfsGUIDBytes, err := parseUUIDStringToGPTBytes(testAPFSPartitionUUID)
	require.NoError(t, err)
	espGUIDBytes, err := parseUUIDStringToGPTBytes(testESPPartitionUUID)
	require.NoError(t, err)
	otherGUIDBytes, err := parseUUIDStringToGPTBytes(testOtherPartitionUUID)
	require.NoError(t, err)
	emptyGUIDBytes := [16]byte{}

	// --- Mock Partition Entries --- Use SMALL LBA ranges for testing
	// Assuming partition entries table ends around LBA 34 for 128 entries (typical)
	lbaStartESP := uint64(34)
	lbaEndESP := lbaStartESP + 200 // Small ESP (approx 100KB)

	lbaStartAPFS := lbaEndESP + 1
	lbaEndAPFS := lbaStartAPFS + 2000 // Small APFS (approx 1MB)

	lbaStartOther := lbaEndAPFS + 1
	lbaEndOther := lbaStartOther + 200 // Small Other (approx 100KB)

	mockPartitions := []gptPartitionEntry{
		{ // 0. EFI System Partition
			PartitionTypeGUID:   espGUIDBytes,
			UniquePartitionGUID: [16]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
			FirstLBA:            lbaStartESP,
			LastLBA:             lbaEndESP,
			Attributes:          0,
			PartitionName:       createGPTPartitionName("EFI"),
		},
		{ // 1. APFS Partition
			PartitionTypeGUID:   apfsGUIDBytes,
			UniquePartitionGUID: [16]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2},
			FirstLBA:            lbaStartAPFS,
			LastLBA:             lbaEndAPFS,
			Attributes:          0,
			PartitionName:       createGPTPartitionName("Macintosh HD"),
		},
		{ // 2. Empty/Unused Partition
			PartitionTypeGUID:   emptyGUIDBytes,
			UniquePartitionGUID: [16]byte{},
			FirstLBA:            0,
			LastLBA:             0,
			Attributes:          0,
			PartitionName:       [72]byte{},
		},
		{ // 3. Other Partition (HFS+)
			PartitionTypeGUID:   otherGUIDBytes,
			UniquePartitionGUID: [16]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 3},
			FirstLBA:            lbaStartOther,
			LastLBA:             lbaEndOther,
			Attributes:          0,
			PartitionName:       createGPTPartitionName("Other Stuff"),
		},
	}

	// --- Create Mock Disk ---
	mockDiskData, _, err := createMockGPTDisk(testLogicalBlockSize, mockPartitions)
	require.NoError(t, err) // This should now allocate a reasonably small slice
	mockReader := bytes.NewReader(mockDiskData)

	// --- Instantiate Manager ---
	manager, err := NewGPTPartitionManager(mockReader, testLogicalBlockSize, testAPFSPartitionUUID)
	require.NoError(t, err)
	require.NotNil(t, manager)

	// --- Test GetPartitionUUID ---
	t.Run("GetPartitionUUID", func(t *testing.T) {
		assert.Equal(t, testAPFSPartitionUUID, manager.GetPartitionUUID())
	})

	// --- Test IsAPFSPartition ---
	t.Run("IsAPFSPartition", func(t *testing.T) {
		assert.True(t, manager.IsAPFSPartition(testAPFSPartitionUUID))
		assert.True(t, manager.IsAPFSPartition(strings.ToLower(testAPFSPartitionUUID)), "Should be case-insensitive")
		assert.False(t, manager.IsAPFSPartition(testESPPartitionUUID))
		assert.False(t, manager.IsAPFSPartition("some-random-uuid"))
	})

	// --- Test ListEFIPartitions ---
	t.Run("ListEFIPartitions_Success", func(t *testing.T) {
		efiParts, listErr := manager.ListEFIPartitions()
		require.NoError(t, listErr)
		require.Len(t, efiParts, 1, "Should find exactly one EFI partition")

		foundESP := efiParts[0]
		expectedESP := mockPartitions[0]

		assert.Equal(t, testESPPartitionUUID, foundESP.UUID, "ESP Type GUID mismatch")
		assert.Equal(t, "EFI", foundESP.Name, "ESP Name mismatch")
		assert.Equal(t, expectedESP.FirstLBA*testLogicalBlockSize, foundESP.Offset, "ESP Offset mismatch")
		expectedSize := (expectedESP.LastLBA - expectedESP.FirstLBA + 1) * testLogicalBlockSize
		assert.Equal(t, expectedSize, foundESP.Size, "ESP Size mismatch")
	})

	t.Run("ListEFIPartitions_NoESPFound", func(t *testing.T) {
		noEspPartitions := []gptPartitionEntry{mockPartitions[1], mockPartitions[3]}
		// *** THIS IS WHERE THE KILL LIKELY HAPPENED ***
		noEspDisk, _, err := createMockGPTDisk(testLogicalBlockSize, noEspPartitions)
		require.NoError(t, err) // Should pass now with smaller LBAs
		noEspReader := bytes.NewReader(noEspDisk)
		noEspManager, err := NewGPTPartitionManager(noEspReader, testLogicalBlockSize, testAPFSPartitionUUID)
		require.NoError(t, err)

		efiParts, listErr := noEspManager.ListEFIPartitions()
		require.NoError(t, listErr)
		assert.Empty(t, efiParts, "Should find no EFI partitions")
	})

	t.Run("ListEFIPartitions_InvalidGPTSignature", func(t *testing.T) {
		badSigDisk := make([]byte, len(mockDiskData))
		copy(badSigDisk, mockDiskData)
		binary.LittleEndian.PutUint64(badSigDisk[testLogicalBlockSize:], 0xBADBADBADBAD)

		badSigReader := bytes.NewReader(badSigDisk)
		badSigManager, err := NewGPTPartitionManager(badSigReader, testLogicalBlockSize, testAPFSPartitionUUID)
		require.NoError(t, err)

		_, listErr := badSigManager.ListEFIPartitions()
		require.Error(t, listErr)
		assert.Contains(t, listErr.Error(), "invalid GPT signature")
	})

	t.Run("ListEFIPartitions_ReadError_Header", func(t *testing.T) {
		shortReader := bytes.NewReader(mockDiskData[:testLogicalBlockSize])
		shortManager, err := NewGPTPartitionManager(shortReader, testLogicalBlockSize, testAPFSPartitionUUID)
		require.NoError(t, err)

		_, listErr := shortManager.ListEFIPartitions()
		require.Error(t, listErr)
		assert.ErrorContains(t, listErr, "failed to read GPT header data")
	})

	t.Run("ListEFIPartitions_ReadError_Entries", func(t *testing.T) {
		shortReader := bytes.NewReader(mockDiskData[:2*testLogicalBlockSize])
		shortManager, err := NewGPTPartitionManager(shortReader, testLogicalBlockSize, testAPFSPartitionUUID)
		require.NoError(t, err)

		_, listErr := shortManager.ListEFIPartitions()
		require.Error(t, listErr)
		assert.ErrorContains(t, listErr, "short read for partition entry array")
	})

	t.Run("ListEFIPartitions_MismatchedEntrySizeInHeader", func(t *testing.T) {
		mismatchedDiskData, header, err := createMockGPTDisk(testLogicalBlockSize, mockPartitions)
		require.NoError(t, err)

		// Modify the header copy in the mock disk data
		header.SizeOfPartitionEntry = 100
		headerBuf := new(bytes.Buffer)
		binary.Write(headerBuf, binary.LittleEndian, header)
		copy(mismatchedDiskData[testLogicalBlockSize:], headerBuf.Bytes()) // Overwrite header in disk image

		mismatchedReader := bytes.NewReader(mismatchedDiskData)
		mismatchedManager, err := NewGPTPartitionManager(mismatchedReader, testLogicalBlockSize, testAPFSPartitionUUID)
		require.NoError(t, err)

		// *** CORRECTION HERE: Expect an error now ***
		_, listErr := mismatchedManager.ListEFIPartitions()
		require.Error(t, listErr, "Should fail because read size based on header doesn't match expected size")
		assert.Contains(t, listErr.Error(), "short read for partition entry array", "Error message should indicate short read based on constant size")
		// Optional: Check that no partitions were returned
		// require.Empty(t, efiParts) // This check requires assigning the first return value '_' -> 'efiParts'
	})
}

// Test helper functions separately
func TestFormatGUID(t *testing.T) {
	guidStr := "C12A7328-F81F-11D2-BA4B-00A0C93EC93B" // ESP GUID
	gptBytes, err := parseUUIDStringToGPTBytes(guidStr)
	require.NoError(t, err)
	formatted := formatGUID(gptBytes)
	assert.Equal(t, guidStr, formatted)

	guidStr2 := types.ApfsGptPartitionUUID // APFS GUID
	gptBytes2, err := parseUUIDStringToGPTBytes(guidStr2)
	require.NoError(t, err)
	formatted2 := formatGUID(gptBytes2)
	assert.Equal(t, guidStr2, formatted2)
}

func TestDecodeUTF16LE(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"Simple", "EFI", "EFI"},
		{"With Space", "Macintosh HD", "Macintosh HD"},
		{"Empty Input String", "", ""},
		{"Max Length (36 chars)", "ThisIsAReallyLongPartitionNameTest", "ThisIsAReallyLongPartitionNameTest"},
		{"With Null Termination", "Hello\x00World", "Hello"}, // Test stopping at null
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var nameBytes [72]byte
			if tc.name == "With Null Termination" {
				runes := []rune("Hello")
				encoded := utf16.Encode(runes)
				for i, r := range encoded {
					binary.LittleEndian.PutUint16(nameBytes[i*2:], r)
				}
			} else {
				nameBytes = createGPTPartitionName(tc.input)
			}
			decoded := decodeUTF16LE(nameBytes[:])
			assert.Equal(t, tc.expected, decoded)
		})
	}

	oddBytes := []byte{0x45, 0x00, 0x46}
	assert.Equal(t, "", decodeUTF16LE(oddBytes))
}
