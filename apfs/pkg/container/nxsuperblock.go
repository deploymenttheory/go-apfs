// File: pkg/container/container.go
package container

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"unsafe"

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

// WriteNXSuperblock writes the NXSuperblock to the given BlockDevice at the specified address.
func WriteNXSuperblock(device types.BlockDevice, addr types.PAddr, sb *types.NXSuperblock) error {
	if device == nil || sb == nil {
		return fmt.Errorf("device and superblock must not be nil")
	}

	data := make([]byte, NXSuperblockSize)
	writer := binary.LittleEndian

	// Write Header
	copy(data[0:8], make([]byte, 8)) // Placeholder for checksum
	writer.PutUint64(data[8:16], uint64(sb.Header.OID))
	writer.PutUint64(data[16:24], uint64(sb.Header.XID))
	writer.PutUint32(data[24:28], sb.Header.Type)
	writer.PutUint32(data[28:32], sb.Header.Subtype)

	writer.PutUint32(data[32:36], sb.Magic)
	writer.PutUint32(data[36:40], sb.BlockSize)
	writer.PutUint64(data[40:48], sb.BlockCount)
	writer.PutUint64(data[48:56], sb.Features)
	writer.PutUint64(data[56:64], sb.ReadOnlyCompatFeatures)
	writer.PutUint64(data[64:72], sb.IncompatFeatures)
	copy(data[72:88], sb.UUID[:])
	writer.PutUint64(data[88:96], uint64(sb.NextOID))
	writer.PutUint64(data[96:104], uint64(sb.NextXID))

	writer.PutUint32(data[104:108], sb.XPDescBlocks)
	writer.PutUint32(data[108:112], sb.XPDataBlocks)
	writer.PutUint64(data[112:120], uint64(sb.XPDescBase))
	writer.PutUint64(data[120:128], uint64(sb.XPDataBase))
	writer.PutUint32(data[128:132], sb.XPDescNext)
	writer.PutUint32(data[132:136], sb.XPDataNext)
	writer.PutUint32(data[136:140], sb.XPDescIndex)
	writer.PutUint32(data[140:144], sb.XPDescLen)
	writer.PutUint32(data[144:148], sb.XPDataIndex)
	writer.PutUint32(data[148:152], sb.XPDataLen)

	writer.PutUint64(data[152:160], uint64(sb.SpacemanOID))
	writer.PutUint64(data[160:168], uint64(sb.OMapOID))
	writer.PutUint64(data[168:176], uint64(sb.ReaperOID))

	writer.PutUint32(data[176:180], sb.TestType)
	writer.PutUint32(data[180:184], sb.MaxFileSystems)

	offset := 184
	for i := 0; i < 100; i++ {
		writer.PutUint64(data[offset:offset+8], uint64(sb.FSOID[i]))
		offset += 8
	}

	for i := 0; i < 32; i++ {
		writer.PutUint64(data[offset:offset+8], sb.Counters[i])
		offset += 8
	}

	// Compute and write checksum
	checksum := checksum.Fletcher64WithZeroedChecksum(data, 0)
	writer.PutUint64(data[0:8], checksum)

	// Write block
	if err := device.WriteBlock(addr, data); err != nil {
		return fmt.Errorf("failed to write NXSuperblock: %w", err)
	}

	return nil
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

// DeserializeNXSuperblock deserializes an NXSuperblock from binary data
func DeserializeNXSuperblock(data []byte) (*NXSuperblock, error) {
	// Validate input size
	if len(data) < int(unsafe.Sizeof(NXSuperblock{})) {
		return nil, ErrStructTooShort
	}

	// Create a new binary reader for the full data slice
	sb := &NXSuperblock{}

	// Read object header first
	header, err := DeserializeObjectHeader(data)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize object header: %w", err)
	}
	sb.Header = *header

	// Reinitialize binary reader to skip over the header and resume reading
	headerSize := int(unsafe.Sizeof(ObjectHeader{}))
	br := NewBinaryReader(bytes.NewReader(data[headerSize:]), binary.LittleEndian)

	// Begin reading all subsequent NXSuperblock fields
	if sb.Magic, err = br.ReadUint32(); err != nil {
		return nil, fmt.Errorf("failed to read magic: %w", err)
	}
	if sb.Magic != NXMagic {
		return nil, ErrInvalidMagic
	}
	if sb.BlockSize, err = br.ReadUint32(); err != nil {
		return nil, err
	}
	if sb.BlockCount, err = br.ReadUint64(); err != nil {
		return nil, err
	}
	if sb.Features, err = br.ReadUint64(); err != nil {
		return nil, err
	}
	if sb.ReadOnlyCompatFeatures, err = br.ReadUint64(); err != nil {
		return nil, err
	}
	if sb.IncompatFeatures, err = br.ReadUint64(); err != nil {
		return nil, err
	}
	if sb.UUID, err = br.ReadUUID(); err != nil {
		return nil, err
	}
	if sb.NextOID, err = br.ReadOID(); err != nil {
		return nil, err
	}
	if sb.NextXID, err = br.ReadXID(); err != nil {
		return nil, err
	}
	if sb.XPDescBlocks, err = br.ReadUint32(); err != nil {
		return nil, err
	}
	if sb.XPDataBlocks, err = br.ReadUint32(); err != nil {
		return nil, err
	}
	if sb.XPDescBase, err = br.ReadPAddr(); err != nil {
		return nil, err
	}
	if sb.XPDataBase, err = br.ReadPAddr(); err != nil {
		return nil, err
	}
	if sb.XPDescNext, err = br.ReadUint32(); err != nil {
		return nil, err
	}
	if sb.XPDataNext, err = br.ReadUint32(); err != nil {
		return nil, err
	}
	if sb.XPDescIndex, err = br.ReadUint32(); err != nil {
		return nil, err
	}
	if sb.XPDescLen, err = br.ReadUint32(); err != nil {
		return nil, err
	}
	if sb.XPDataIndex, err = br.ReadUint32(); err != nil {
		return nil, err
	}
	if sb.XPDataLen, err = br.ReadUint32(); err != nil {
		return nil, err
	}
	if sb.SpacemanOID, err = br.ReadOID(); err != nil {
		return nil, err
	}
	if sb.OMapOID, err = br.ReadOID(); err != nil {
		return nil, err
	}
	if sb.ReaperOID, err = br.ReadOID(); err != nil {
		return nil, err
	}
	if sb.TestType, err = br.ReadUint32(); err != nil {
		return nil, err
	}
	if sb.MaxFileSystems, err = br.ReadUint32(); err != nil {
		return nil, err
	}

	// Arrays
	fsOIDs, err := br.ReadOIDArray(NXMaxFileSystems)
	if err != nil {
		return nil, fmt.Errorf("failed to read FSOIDs: %w", err)
	}
	copy(sb.FSOID[:], fsOIDs)

	counters, err := br.ReadUint64Array(NXNumCounters)
	if err != nil {
		return nil, fmt.Errorf("failed to read counters: %w", err)
	}
	copy(sb.Counters[:], counters)

	ephemeralInfo, err := br.ReadUint64Array(NXEphemeralInfoCount)
	if err != nil {
		return nil, fmt.Errorf("failed to read ephemeral info: %w", err)
	}
	copy(sb.EphemeralInfo[:], ephemeralInfo)

	// Nested structs
	if sb.BlockedOutPRange.StartAddr, err = br.ReadPAddr(); err != nil {
		return nil, err
	}
	if sb.BlockedOutPRange.BlockCount, err = br.ReadUint64(); err != nil {
		return nil, err
	}
	if sb.EvictMappingTreeOID, err = br.ReadOID(); err != nil {
		return nil, err
	}
	if sb.Flags, err = br.ReadUint64(); err != nil {
		return nil, err
	}
	if sb.EFIJumpstart, err = br.ReadPAddr(); err != nil {
		return nil, err
	}
	if sb.FusionUUID, err = br.ReadUUID(); err != nil {
		return nil, err
	}
	if sb.KeyLocker.StartAddr, err = br.ReadPAddr(); err != nil {
		return nil, err
	}
	if sb.KeyLocker.BlockCount, err = br.ReadUint64(); err != nil {
		return nil, err
	}

	// Final optional Fusion fields
	if sb.TestOID, err = br.ReadOID(); err != nil {
		return nil, err
	}
	if sb.FusionMtOID, err = br.ReadOID(); err != nil {
		return nil, err
	}
	if sb.FusionWbcOID, err = br.ReadOID(); err != nil {
		return nil, err
	}
	if sb.FusionWbc.StartAddr, err = br.ReadPAddr(); err != nil {
		return nil, err
	}
	if sb.FusionWbc.BlockCount, err = br.ReadUint64(); err != nil {
		return nil, err
	}
	if sb.NewestMountedVersion, err = br.ReadUint64(); err != nil {
		return nil, err
	}
	if sb.MkbLocker.StartAddr, err = br.ReadPAddr(); err != nil {
		return nil, err
	}
	if sb.MkbLocker.BlockCount, err = br.ReadUint64(); err != nil {
		return nil, err
	}

	return sb, nil
}

// SerializeNXSuperblock serializes an NXSuperblock to binary data
func SerializeNXSuperblock(sb *NXSuperblock) ([]byte, error) {
	buf := new(bytes.Buffer)
	writer := NewBinaryWriter(buf, binary.LittleEndian)

	// Write object header
	headerBytes, err := SerializeObjectHeader(&sb.Header)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize object header: %w", err)
	}
	if err := writer.WriteBytes(headerBytes); err != nil {
		return nil, fmt.Errorf("failed to write object header: %w", err)
	}

	// Begin writing NXSuperblock fields
	if err := writer.WriteUint32(sb.Magic); err != nil {
		return nil, err
	}
	if err := writer.WriteUint32(sb.BlockSize); err != nil {
		return nil, err
	}
	if err := writer.WriteUint64(sb.BlockCount); err != nil {
		return nil, err
	}
	if err := writer.WriteUint64(sb.Features); err != nil {
		return nil, err
	}
	if err := writer.WriteUint64(sb.ReadOnlyCompatFeatures); err != nil {
		return nil, err
	}
	if err := writer.WriteUint64(sb.IncompatFeatures); err != nil {
		return nil, err
	}
	if err := writer.WriteUUID(sb.UUID); err != nil {
		return nil, err
	}
	if err := writer.WriteOID(sb.NextOID); err != nil {
		return nil, err
	}
	if err := writer.WriteXID(sb.NextXID); err != nil {
		return nil, err
	}
	if err := writer.WriteUint32(sb.XPDescBlocks); err != nil {
		return nil, err
	}
	if err := writer.WriteUint32(sb.XPDataBlocks); err != nil {
		return nil, err
	}
	if err := writer.WritePAddr(sb.XPDescBase); err != nil {
		return nil, err
	}
	if err := writer.WritePAddr(sb.XPDataBase); err != nil {
		return nil, err
	}
	if err := writer.WriteUint32(sb.XPDescNext); err != nil {
		return nil, err
	}
	if err := writer.WriteUint32(sb.XPDataNext); err != nil {
		return nil, err
	}
	if err := writer.WriteUint32(sb.XPDescIndex); err != nil {
		return nil, err
	}
	if err := writer.WriteUint32(sb.XPDescLen); err != nil {
		return nil, err
	}
	if err := writer.WriteUint32(sb.XPDataIndex); err != nil {
		return nil, err
	}
	if err := writer.WriteUint32(sb.XPDataLen); err != nil {
		return nil, err
	}
	if err := writer.WriteOID(sb.SpacemanOID); err != nil {
		return nil, err
	}
	if err := writer.WriteOID(sb.OMapOID); err != nil {
		return nil, err
	}
	if err := writer.WriteOID(sb.ReaperOID); err != nil {
		return nil, err
	}
	if err := writer.WriteUint32(sb.TestType); err != nil {
		return nil, err
	}
	if err := writer.WriteUint32(sb.MaxFileSystems); err != nil {
		return nil, err
	}

	// Write fixed-size arrays
	for _, oid := range sb.FSOID {
		if err := writer.WriteOID(oid); err != nil {
			return nil, fmt.Errorf("failed to write FSOID: %w", err)
		}
	}
	for _, counter := range sb.Counters {
		if err := writer.WriteUint64(counter); err != nil {
			return nil, fmt.Errorf("failed to write counter: %w", err)
		}
	}
	for _, ephem := range sb.EphemeralInfo {
		if err := writer.WriteUint64(ephem); err != nil {
			return nil, fmt.Errorf("failed to write ephemeral info: %w", err)
		}
	}

	// Write nested structs
	if err := writer.WritePAddr(sb.BlockedOutPRange.StartAddr); err != nil {
		return nil, err
	}
	if err := writer.WriteUint64(sb.BlockedOutPRange.BlockCount); err != nil {
		return nil, err
	}
	if err := writer.WriteOID(sb.EvictMappingTreeOID); err != nil {
		return nil, err
	}
	if err := writer.WriteUint64(sb.Flags); err != nil {
		return nil, err
	}
	if err := writer.WritePAddr(sb.EFIJumpstart); err != nil {
		return nil, err
	}
	if err := writer.WriteUUID(sb.FusionUUID); err != nil {
		return nil, err
	}
	if err := writer.WritePAddr(sb.KeyLocker.StartAddr); err != nil {
		return nil, err
	}
	if err := writer.WriteUint64(sb.KeyLocker.BlockCount); err != nil {
		return nil, err
	}

	// Write fusion metadata
	if err := writer.WriteOID(sb.TestOID); err != nil {
		return nil, err
	}
	if err := writer.WriteOID(sb.FusionMtOID); err != nil {
		return nil, err
	}
	if err := writer.WriteOID(sb.FusionWbcOID); err != nil {
		return nil, err
	}
	if err := writer.WritePAddr(sb.FusionWbc.StartAddr); err != nil {
		return nil, err
	}
	if err := writer.WriteUint64(sb.FusionWbc.BlockCount); err != nil {
		return nil, err
	}
	if err := writer.WriteUint64(sb.NewestMountedVersion); err != nil {
		return nil, err
	}
	if err := writer.WritePAddr(sb.MkbLocker.StartAddr); err != nil {
		return nil, err
	}
	if err := writer.WriteUint64(sb.MkbLocker.BlockCount); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// DeserializeAPFSSuperblock deserializes an APFSSuperblock from binary data
func DeserializeAPFSSuperblock(data []byte) (*APFSSuperblock, error) {
	if len(data) < int(unsafe.Sizeof(APFSSuperblock{})) {
		return nil, ErrStructTooShort
	}

	sb := &APFSSuperblock{}

	header, err := DeserializeObjectHeader(data)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize object header: %w", err)
	}
	sb.Header = *header

	headerSize := int(unsafe.Sizeof(ObjectHeader{}))
	br := NewBinaryReader(bytes.NewReader(data[headerSize:]), binary.LittleEndian)

	if sb.Magic, err = br.ReadUint32(); err != nil {
		return nil, fmt.Errorf("failed to read magic: %w", err)
	}
	if sb.Magic != APFSMagic {
		return nil, ErrInvalidMagic
	}
	if sb.FSIndex, err = br.ReadUint32(); err != nil {
		return nil, fmt.Errorf("failed to read FSIndex: %w", err)
	}
	if sb.Features, err = br.ReadUint64(); err != nil {
		return nil, fmt.Errorf("failed to read features: %w", err)
	}
	if sb.ReadOnlyCompatFeatures, err = br.ReadUint64(); err != nil {
		return nil, fmt.Errorf("failed to read read-only compatible features: %w", err)
	}
	if sb.IncompatFeatures, err = br.ReadUint64(); err != nil {
		return nil, fmt.Errorf("failed to read incompatible features: %w", err)
	}
	if sb.UnmountTime, err = br.ReadUint64(); err != nil {
		return nil, fmt.Errorf("failed to read unmount time: %w", err)
	}
	if sb.ReserveBlockCount, err = br.ReadUint64(); err != nil {
		return nil, fmt.Errorf("failed to read reserve block count: %w", err)
	}
	if sb.QuotaBlockCount, err = br.ReadUint64(); err != nil {
		return nil, fmt.Errorf("failed to read quota block count: %w", err)
	}
	if sb.AllocCount, err = br.ReadUint64(); err != nil {
		return nil, fmt.Errorf("failed to read alloc count: %w", err)
	}
	if err = br.Read(&sb.MetaCrypto); err != nil {
		return nil, fmt.Errorf("failed to read meta crypto state: %w", err)
	}
	if sb.RootTreeType, err = br.ReadUint32(); err != nil {
		return nil, fmt.Errorf("failed to read root tree type: %w", err)
	}
	if sb.ExtentrefTreeType, err = br.ReadUint32(); err != nil {
		return nil, fmt.Errorf("failed to read extentref tree type: %w", err)
	}
	if sb.SnapMetaTreeType, err = br.ReadUint32(); err != nil {
		return nil, fmt.Errorf("failed to read snap meta tree type: %w", err)
	}
	if sb.OMapOID, err = br.ReadOID(); err != nil {
		return nil, fmt.Errorf("failed to read OMapOID: %w", err)
	}
	if sb.RootTreeOID, err = br.ReadOID(); err != nil {
		return nil, fmt.Errorf("failed to read root tree OID: %w", err)
	}
	if sb.ExtentrefTreeOID, err = br.ReadOID(); err != nil {
		return nil, fmt.Errorf("failed to read extentref tree OID: %w", err)
	}
	if sb.SnapMetaTreeOID, err = br.ReadOID(); err != nil {
		return nil, fmt.Errorf("failed to read snap meta tree OID: %w", err)
	}
	if sb.RevertToXID, err = br.ReadXID(); err != nil {
		return nil, fmt.Errorf("failed to read revert to XID: %w", err)
	}
	if sb.RevertToSblockOID, err = br.ReadOID(); err != nil {
		return nil, fmt.Errorf("failed to read revert to sblock OID: %w", err)
	}
	if sb.NextObjID, err = br.ReadUint64(); err != nil {
		return nil, fmt.Errorf("failed to read next object ID: %w", err)
	}
	if sb.NumFiles, err = br.ReadUint64(); err != nil {
		return nil, fmt.Errorf("failed to read num files: %w", err)
	}
	if sb.NumDirectories, err = br.ReadUint64(); err != nil {
		return nil, fmt.Errorf("failed to read num directories: %w", err)
	}
	if sb.NumSymlinks, err = br.ReadUint64(); err != nil {
		return nil, fmt.Errorf("failed to read num symlinks: %w", err)
	}
	if sb.NumOtherFSObjects, err = br.ReadUint64(); err != nil {
		return nil, fmt.Errorf("failed to read num other fs objects: %w", err)
	}
	if sb.NumSnapshots, err = br.ReadUint64(); err != nil {
		return nil, fmt.Errorf("failed to read num snapshots: %w", err)
	}
	if sb.TotalBlocksAlloced, err = br.ReadUint64(); err != nil {
		return nil, fmt.Errorf("failed to read total blocks alloced: %w", err)
	}
	if sb.TotalBlocksFreed, err = br.ReadUint64(); err != nil {
		return nil, fmt.Errorf("failed to read total blocks freed: %w", err)
	}
	if sb.UUID, err = br.ReadUUID(); err != nil {
		return nil, fmt.Errorf("failed to read volume UUID: %w", err)
	}
	if sb.LastModTime, err = br.ReadUint64(); err != nil {
		return nil, fmt.Errorf("failed to read last mod time: %w", err)
	}
	if sb.FSFlags, err = br.ReadUint64(); err != nil {
		return nil, fmt.Errorf("failed to read fs flags: %w", err)
	}
	if err = br.Read(&sb.FormattedBy); err != nil {
		return nil, fmt.Errorf("failed to read formatted by: %w", err)
	}
	if err = br.Read(&sb.ModifiedBy); err != nil {
		return nil, fmt.Errorf("failed to read modified by: %w", err)
	}
	if err = br.Read(&sb.VolName); err != nil {
		return nil, fmt.Errorf("failed to read volume name: %w", err)
	}
	if sb.NextDocID, err = br.ReadUint32(); err != nil {
		return nil, fmt.Errorf("failed to read next doc ID: %w", err)
	}
	if sb.Role, err = br.ReadUint16(); err != nil {
		return nil, fmt.Errorf("failed to read role: %w", err)
	}
	if sb.Reserved, err = br.ReadUint16(); err != nil {
		return nil, fmt.Errorf("failed to read reserved: %w", err)
	}
	if sb.RootToXID, err = br.ReadXID(); err != nil {
		return nil, fmt.Errorf("failed to read root to XID: %w", err)
	}
	if sb.ERStateOID, err = br.ReadOID(); err != nil {
		return nil, fmt.Errorf("failed to read ER state OID: %w", err)
	}
	if sb.CloneinfoIDEpoch, err = br.ReadUint64(); err != nil {
		return nil, fmt.Errorf("failed to read cloneinfo ID epoch: %w", err)
	}
	if sb.CloneinfoXID, err = br.ReadUint64(); err != nil {
		return nil, fmt.Errorf("failed to read cloneinfo XID: %w", err)
	}
	if sb.SnapMetaExtOID, err = br.ReadOID(); err != nil {
		return nil, fmt.Errorf("failed to read snap meta ext OID: %w", err)
	}
	if sb.VolumeGroupID, err = br.ReadUUID(); err != nil {
		return nil, fmt.Errorf("failed to read volume group ID: %w", err)
	}
	if sb.IntegrityMetaOID, err = br.ReadOID(); err != nil {
		return nil, fmt.Errorf("failed to read integrity meta OID: %w", err)
	}
	if sb.FextTreeOID, err = br.ReadOID(); err != nil {
		return nil, fmt.Errorf("failed to read fext tree OID: %w", err)
	}
	if sb.FextTreeType, err = br.ReadUint32(); err != nil {
		return nil, fmt.Errorf("failed to read fext tree type: %w", err)
	}
	if sb.ReservedType, err = br.ReadUint32(); err != nil {
		return nil, fmt.Errorf("failed to read reserved type: %w", err)
	}
	if sb.ReservedOID, err = br.ReadOID(); err != nil {
		return nil, fmt.Errorf("failed to read reserved OID: %w", err)
	}

	return sb, nil
}
