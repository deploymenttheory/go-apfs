package container

import (
	"encoding/binary"
	"fmt"

	"github.com/deploymenttheory/go-apfs/apfs/pkg/checksum"
	"github.com/deploymenttheory/go-apfs/apfs/pkg/types"
)

// WriteVolumeSuperblock writes a volume superblock to the given physical address on disk.
func WriteVolumeSuperblock(device types.BlockDevice, addr types.PAddr, volume *types.APFSSuperblock) error {
	if device == nil || volume == nil {
		return fmt.Errorf("device and volume must not be nil")
	}

	// Serialize the volume superblock
	data, err := serializeAPFSSuperblock(volume)
	if err != nil {
		return fmt.Errorf("failed to serialize volume superblock: %w", err)
	}

	// Write to the specified block address
	return device.WriteBlock(addr, data)
}

func serializeAPFSSuperblock(volume *types.APFSSuperblock) ([]byte, error) {
	const size = 1024 // The on-disk size of apfs_superblock_t is 1024 bytes (0x400).

	data := make([]byte, size)
	writer := binary.LittleEndian
	offset := 8 // Leave space for checksum

	writer.PutUint64(data[offset:], uint64(volume.Header.OID))
	offset += 8
	writer.PutUint64(data[offset:], uint64(volume.Header.XID))
	offset += 8
	writer.PutUint32(data[offset:], volume.Header.Type)
	offset += 4
	writer.PutUint32(data[offset:], volume.Header.Subtype)
	offset += 4

	writer.PutUint32(data[offset:], volume.Magic)
	offset += 4
	writer.PutUint32(data[offset:], volume.FSIndex)
	offset += 4
	writer.PutUint64(data[offset:], volume.Features)
	offset += 8
	writer.PutUint64(data[offset:], volume.ReadOnlyCompatFeatures)
	offset += 8
	writer.PutUint64(data[offset:], volume.IncompatFeatures)
	offset += 8
	writer.PutUint64(data[offset:], volume.UnmountTime)
	offset += 8
	writer.PutUint64(data[offset:], volume.ReserveBlockCount)
	offset += 8
	writer.PutUint64(data[offset:], volume.QuotaBlockCount)
	offset += 8
	writer.PutUint64(data[offset:], volume.AllocCount)
	offset += 8

	// MetaCryptoState is a nested struct — manually marshal it per APFS layout
	binary.LittleEndian.PutUint16(data[offset:], volume.MetaCrypto.MajorVersion)
	offset += 2
	binary.LittleEndian.PutUint16(data[offset:], volume.MetaCrypto.MinorVersion)
	offset += 2
	binary.LittleEndian.PutUint32(data[offset:], volume.MetaCrypto.Flags)
	offset += 4
	binary.LittleEndian.PutUint32(data[offset:], volume.MetaCrypto.PersistentClass)
	offset += 4
	binary.LittleEndian.PutUint32(data[offset:], volume.MetaCrypto.KeyOSVersion)
	offset += 4
	binary.LittleEndian.PutUint16(data[offset:], volume.MetaCrypto.KeyRevision)
	offset += 2
	binary.LittleEndian.PutUint16(data[offset:], volume.MetaCrypto.Unused)
	offset += 2

	writer.PutUint32(data[offset:], volume.RootTreeType)
	offset += 4
	writer.PutUint32(data[offset:], volume.ExtentrefTreeType)
	offset += 4
	writer.PutUint32(data[offset:], volume.SnapMetaTreeType)
	offset += 4

	writer.PutUint64(data[offset:], uint64(volume.OMapOID))
	offset += 8
	writer.PutUint64(data[offset:], uint64(volume.RootTreeOID))
	offset += 8
	writer.PutUint64(data[offset:], uint64(volume.ExtentrefTreeOID))
	offset += 8
	writer.PutUint64(data[offset:], uint64(volume.SnapMetaTreeOID))
	offset += 8
	writer.PutUint64(data[offset:], uint64(volume.RevertToXID))
	offset += 8
	writer.PutUint64(data[offset:], uint64(volume.RevertToSblockOID))
	offset += 8
	writer.PutUint64(data[offset:], volume.NextObjID)
	offset += 8
	writer.PutUint64(data[offset:], volume.NumFiles)
	offset += 8
	writer.PutUint64(data[offset:], volume.NumDirectories)
	offset += 8
	writer.PutUint64(data[offset:], volume.NumSymlinks)
	offset += 8
	writer.PutUint64(data[offset:], volume.NumOtherFSObjects)
	offset += 8
	writer.PutUint64(data[offset:], volume.NumSnapshots)
	offset += 8
	writer.PutUint64(data[offset:], volume.TotalBlocksAlloced)
	offset += 8
	writer.PutUint64(data[offset:], volume.TotalBlocksFreed)
	offset += 8

	copy(data[offset:], volume.UUID[:])
	offset += 16
	binary.LittleEndian.PutUint64(data[offset:], volume.LastModTime)
	offset += 8
	binary.LittleEndian.PutUint64(data[offset:], volume.FSFlags)
	offset += 8

	// FormattedBy (struct apfs_modified_by_t: 32 + 8 + 8 = 48 bytes)
	copy(data[offset:], volume.FormattedBy.ID[:])
	offset += 32
	binary.LittleEndian.PutUint64(data[offset:], volume.FormattedBy.Timestamp)
	offset += 8
	binary.LittleEndian.PutUint64(data[offset:], uint64(volume.FormattedBy.LastXID))
	offset += 8

	// ModifiedBy[8]
	for i := 0; i < 8; i++ {
		copy(data[offset:], volume.ModifiedBy[i].ID[:])
		offset += 32
		binary.LittleEndian.PutUint64(data[offset:], volume.ModifiedBy[i].Timestamp)
		offset += 8
		binary.LittleEndian.PutUint64(data[offset:], uint64(volume.ModifiedBy[i].LastXID))
		offset += 8
	}

	// Volume name (UTF-8, 256 bytes, null-terminated)
	copy(data[offset:], volume.VolName[:])
	offset += 256

	writer.PutUint32(data[offset:], volume.NextDocID)
	offset += 4
	writer.PutUint16(data[offset:], volume.Role)
	offset += 2
	writer.PutUint16(data[offset:], volume.Reserved)
	offset += 2
	writer.PutUint64(data[offset:], uint64(volume.RootToXID))
	offset += 8
	writer.PutUint64(data[offset:], uint64(volume.ERStateOID))
	offset += 8
	writer.PutUint64(data[offset:], volume.CloneinfoIDEpoch)
	offset += 8
	writer.PutUint64(data[offset:], volume.CloneinfoXID)
	offset += 8
	writer.PutUint64(data[offset:], uint64(volume.SnapMetaExtOID))
	offset += 8
	copy(data[offset:], volume.VolumeGroupID[:])
	offset += 16
	writer.PutUint64(data[offset:], uint64(volume.IntegrityMetaOID))
	offset += 8
	writer.PutUint64(data[offset:], uint64(volume.FextTreeOID))
	offset += 8
	writer.PutUint32(data[offset:], volume.FextTreeType)
	offset += 4
	writer.PutUint32(data[offset:], volume.ReservedType)
	offset += 4
	writer.PutUint64(data[offset:], uint64(volume.ReservedOID))
	offset += 8

	// Checksum
	checksum := checksum.Fletcher64WithZeroedChecksum(data, 0)
	writer.PutUint64(data[0:8], checksum)

	return data[:offset], nil
}
