package container

import (
	"encoding/binary"
	"fmt"

	"github.com/deploymenttheory/go-apfs/pkg/checksum"
	"github.com/deploymenttheory/go-apfs/pkg/types"
)

// NXSuperblockSize defines the exact size of NXSuperblock on-disk structure (fixed size portion)
const NXSuperblockSize = 1376

// ReadNXSuperblock reads the NXSuperblock from the given BlockDevice at the specified address.
func ReadNXSuperblock(device types.BlockDevice, addr types.PAddr) (*types.NXSuperblock, error) {
	blockSize := device.GetBlockSize()
	if NXSuperblockSize > int(blockSize) {
		return nil, fmt.Errorf("superblock size (%d) exceeds device block size (%d)", NXSuperblockSize, blockSize)
	}

	// Read the raw block containing NXSuperblock
	data, err := device.ReadBlock(addr)
	if err != nil {
		return nil, fmt.Errorf("failed to read block at addr %d: %w", addr, err)
	}

	if len(data) < NXSuperblockSize {
		return nil, fmt.Errorf("data too short: expected %d bytes, got %d", NXSuperblockSize, len(data))
	}

	// Deserialize into NXSuperblock struct
	var sb types.NXSuperblock
	reader := binary.LittleEndian

	sb.Header.Cksum = types.Checksum(data[0:8])
	sb.Header.OID = types.OID(reader.Uint64(data[8:16]))
	sb.Header.XID = types.XID(reader.Uint64(data[16:24]))
	sb.Header.Type = reader.Uint32(data[24:28])
	sb.Header.Subtype = reader.Uint32(data[28:32])

	sb.Magic = reader.Uint32(data[32:36])
	sb.BlockSize = reader.Uint32(data[36:40])
	sb.BlockCount = reader.Uint64(data[40:48])
	sb.Features = reader.Uint64(data[48:56])
	sb.ReadOnlyCompatFeatures = reader.Uint64(data[56:64])
	sb.IncompatFeatures = reader.Uint64(data[64:72])
	copy(sb.UUID[:], data[72:88])
	sb.NextOID = types.OID(reader.Uint64(data[88:96]))
	sb.NextXID = types.XID(reader.Uint64(data[96:104]))

	sb.XPDescBlocks = reader.Uint32(data[104:108])
	sb.XPDataBlocks = reader.Uint32(data[108:112])
	sb.XPDescBase = types.PAddr(reader.Uint64(data[112:120]))
	sb.XPDataBase = types.PAddr(reader.Uint64(data[120:128]))
	sb.XPDescNext = reader.Uint32(data[128:132])
	sb.XPDataNext = reader.Uint32(data[132:136])
	sb.XPDescIndex = reader.Uint32(data[136:140])
	sb.XPDescLen = reader.Uint32(data[140:144])
	sb.XPDataIndex = reader.Uint32(data[144:148])
	sb.XPDataLen = reader.Uint32(data[148:152])

	sb.SpacemanOID = types.OID(reader.Uint64(data[152:160]))
	sb.OMapOID = types.OID(reader.Uint64(data[160:168]))
	sb.ReaperOID = types.OID(reader.Uint64(data[168:176]))

	sb.TestType = reader.Uint32(data[176:180])
	sb.MaxFileSystems = reader.Uint32(data[180:184])

	// Deserialize volume OIDs
	offset := 184
	for i := 0; i < 100; i++ {
		sb.FSOID[i] = types.OID(reader.Uint64(data[offset : offset+8]))
		offset += 8
	}

	// Deserialize counters
	for i := 0; i < 32; i++ {
		sb.Counters[i] = reader.Uint64(data[offset : offset+8])
		offset += 8
	}

	// Validate checksum
	computedChecksum := checksum.Fletcher64WithZeroedChecksum(data, 0)
	expectedChecksum := reader.Uint64(sb.Header.Cksum[:])
	if computedChecksum != expectedChecksum {
		return nil, fmt.Errorf("checksum validation failed: computed 0x%x, expected 0x%x", computedChecksum, expectedChecksum)
	}

	// Minimal validation
	if sb.Magic != types.NXMagic {
		return nil, fmt.Errorf("invalid NXSuperblock magic number: got 0x%x, want 0x%x", sb.Magic, types.NXMagic)
	}

	return &sb, nil
}

// Validate performs basic validation checks on the superblock structure.
func (sb *types.NXSuperblock) Validate() error {
	if sb.BlockSize < types.MinBlockSize || sb.BlockSize > types.MaxBlockSize {
		return fmt.Errorf("unsupported block size: %d", sb.BlockSize)
	}
	if sb.BlockCount == 0 {
		return fmt.Errorf("invalid block count: %d", sb.BlockCount)
	}
	return nil
}
