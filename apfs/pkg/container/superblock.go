// File: pkg/container/container.go
package container

import (
	"encoding/binary"
	"fmt"

	"github.com/deploymenttheory/go-apfs/apfs/pkg/checksum"
	"github.com/deploymenttheory/go-apfs/apfs/pkg/types"
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

	offset := 184
	for i := 0; i < 100; i++ {
		sb.FSOID[i] = types.OID(reader.Uint64(data[offset : offset+8]))
		offset += 8
	}

	for i := 0; i < 32; i++ {
		sb.Counters[i] = reader.Uint64(data[offset : offset+8])
		offset += 8
	}

	// Checksum validation
	computedChecksum := checksum.Fletcher64WithZeroedChecksum(data, 0)
	expectedChecksum := reader.Uint64(sb.Header.Cksum[:])
	if computedChecksum != expectedChecksum {
		return nil, fmt.Errorf("checksum mismatch: computed 0x%x, expected 0x%x", computedChecksum, expectedChecksum)
	}

	// Minimal validation
	if err := ValidateNXSuperblock(&sb); err != nil {
		return nil, err
	}

	return &sb, nil
}

// WriteVolumeSuperblock writes a volume superblock to disk
func WriteVolumeSuperblock(device types.BlockDevice, volume *types.APFSSuperblock) error {
	// Serialize the volume superblock
	data, err := serializeAPFSSuperblock(volume)
	if err != nil {
		return err
	}

	// Write to the volume's superblock address
	// Assuming this is stored in the volume structure or is a known location
	return device.WriteBlock(types.PAddr(volume.BlockNum), data)
}

// serializeAPFSSuperblock serializes an APFSSuperblock to binary data
func serializeAPFSSuperblock(volume *types.APFSSuperblock) ([]byte, error) {
	// Implementation similar to serializeERStatePhys
	// This would serialize all fields of the APFSSuperblock structure

	// For brevity, this is a simplified implementation
	// In a real implementation, you would serialize all fields

	// Calculate the size based on APFSSuperblock structure
	dataSize := 1024 // Example size, actual size depends on the structure

	data := make([]byte, dataSize)
	r := binary.LittleEndian

	// Write header (leave checksum as 0 for now)
	offset := 8 // Skip checksum
	r.PutUint64(data[offset:], uint64(volume.Header.OID))
	offset += 8
	r.PutUint64(data[offset:], uint64(volume.Header.XID))
	offset += 8
	r.PutUint32(data[offset:], volume.Header.Type)
	offset += 4
	r.PutUint32(data[offset:], volume.Header.Subtype)
	offset += 4

	// Write volume-specific fields
	r.PutUint32(data[offset:], volume.Magic)
	offset += 4

	// ... Write all other fields ...

	// Include the ERStateOID
	r.PutUint64(data[offset:], uint64(volume.ERStateOID))
	offset += 8

	// ... Write remaining fields ...

	// Calculate and set checksum
	checksum := checksum.Fletcher64WithZeroedChecksum(data, 0)
	r.PutUint64(data[0:8], checksum)

	return data, nil
}

// ValidateNXSuperblock performs thorough validation checks on the superblock structure.
func ValidateNXSuperblock(sb *types.NXSuperblock) error {
	if sb.Magic != types.NXMagic {
		return fmt.Errorf("invalid NXSuperblock magic: 0x%x (expected 0x%x)", sb.Magic, types.NXMagic)
	}
	if sb.BlockSize < types.MinBlockSize || sb.BlockSize > types.MaxBlockSize {
		return fmt.Errorf("unsupported block size: %d (must be between %d and %d)",
			sb.BlockSize, types.MinBlockSize, types.MaxBlockSize)
	}
	if sb.BlockCount == 0 {
		return fmt.Errorf("block count must be non-zero")
	}
	if sb.MaxFileSystems == 0 || sb.MaxFileSystems > types.NXMaxFileSystems {
		return fmt.Errorf("invalid max file systems: %d (maximum allowed is %d)", sb.MaxFileSystems, types.NXMaxFileSystems)
	}
	if sb.SpacemanOID == types.OIDInvalid {
		return fmt.Errorf("invalid Spaceman OID (cannot be OIDInvalid)")
	}
	if sb.OMapOID == types.OIDInvalid {
		return fmt.Errorf("invalid OMap OID (cannot be OIDInvalid)")
	}
	if sb.ReaperOID == types.OIDInvalid {
		return fmt.Errorf("invalid Reaper OID (cannot be OIDInvalid)")
	}
	if sb.Features&types.UnsupportedFeaturesMask != 0 {
		return fmt.Errorf("unsupported features detected: 0x%x", sb.Features&types.UnsupportedFeaturesMask)
	}
	if sb.IncompatFeatures&types.UnsupportedIncompatFeaturesMask != 0 {
		return fmt.Errorf("unsupported incompatible features detected: 0x%x", sb.IncompatFeatures&types.UnsupportedIncompatFeaturesMask)
	}
	// Additional optional checks:
	if sb.NextOID <= types.OIDReservedCount {
		return fmt.Errorf("NextOID (%d) is within reserved range (must exceed %d)", sb.NextOID, types.OIDReservedCount)
	}
	return nil
}
