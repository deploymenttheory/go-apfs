package types

// DMG and GPT partition table constants
const (
	// GPT (GUID Partition Table) header and partition entry offsets
	// Reference: UEFI Specification Part 1, Chapter 5

	GPTHeaderOffset       = 512  // LBA 1: Primary GPT header location (byte offset)
	GPTEntrySize          = 128  // Size of each GPT partition entry (bytes)
	GPTEntriesStartOffset = 2048 // LBA 4: Standard partition entries location (byte offset)

	// APFS-specific DMG offsets
	APFSMagicOffset = 32    // Offset of NXSB magic within nx_superblock_t
	GPTAPFSOffset   = 20480 // Standard APFS offset after GPT (LBA 40 × 512 bytes)

	// Note: APFS GPT partition type UUID is defined in efi_jumpstart.go as ApfsGptPartitionUUID
)
