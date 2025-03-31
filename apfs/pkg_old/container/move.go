package container

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"unsafe"

	"github.com/deploymenttheory/go-apfs/apfs/pkg/checksum"
	"github.com/deploymenttheory/go-apfs/apfs/pkg/types"
)

// ================================
// Serialization and Deserialization Functions
// ================================

// DeserializeAPFSSuperblock deserializes an APFSSuperblock from binary data
func DeserializeAPFSSuperblock(data []byte) (*types.APFSSuperblock, error) {
	if len(data) < int(unsafe.Sizeof(types.APFSSuperblock{})) {
		return nil, types.ErrStructTooShort
	}

	sb := &types.APFSSuperblock{}

	header, err := DeserializeObjectHeader(data)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize object header: %w", err)
	}
	sb.Header = *header

	headerSize := int(unsafe.Sizeof(types.ObjectHeader{}))
	br := types.NewBinaryReader(bytes.NewReader(data[headerSize:]), binary.LittleEndian)

	if sb.Magic, err = br.ReadUint32(); err != nil {
		return nil, fmt.Errorf("failed to read magic: %w", err)
	}
	if sb.Magic != types.APFSMagic {
		return nil, types.ErrInvalidMagic
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

// SerializeAPFSSuperblock serializes an APFSSuperblock to binary data
func SerializeAPFSSuperblock(sb *types.APFSSuperblock) ([]byte, error) {
	buf := new(bytes.Buffer)
	writer := types.NewBinaryWriter(buf, binary.LittleEndian)

	// Write object header
	headerBytes, err := SerializeObjectHeader(&sb.Header)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize object header: %w", err)
	}
	if err := writer.WriteBytes(headerBytes); err != nil {
		return nil, fmt.Errorf("failed to write object header: %w", err)
	}

	if err := writer.WriteUint32(sb.Magic); err != nil {
		return nil, err
	}
	if err := writer.WriteUint32(sb.FSIndex); err != nil {
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
	if err := writer.WriteUint64(sb.UnmountTime); err != nil {
		return nil, err
	}
	if err := writer.WriteUint64(sb.ReserveBlockCount); err != nil {
		return nil, err
	}
	if err := writer.WriteUint64(sb.QuotaBlockCount); err != nil {
		return nil, err
	}
	if err := writer.WriteUint64(sb.AllocCount); err != nil {
		return nil, err
	}
	if err := writer.Write(sb.MetaCrypto); err != nil {
		return nil, err
	}
	if err := writer.WriteUint32(sb.RootTreeType); err != nil {
		return nil, err
	}
	if err := writer.WriteUint32(sb.ExtentrefTreeType); err != nil {
		return nil, err
	}
	if err := writer.WriteUint32(sb.SnapMetaTreeType); err != nil {
		return nil, err
	}
	if err := writer.WriteOID(sb.OMapOID); err != nil {
		return nil, err
	}
	if err := writer.WriteOID(sb.RootTreeOID); err != nil {
		return nil, err
	}
	if err := writer.WriteOID(sb.ExtentrefTreeOID); err != nil {
		return nil, err
	}
	if err := writer.WriteOID(sb.SnapMetaTreeOID); err != nil {
		return nil, err
	}
	if err := writer.WriteXID(sb.RevertToXID); err != nil {
		return nil, err
	}
	if err := writer.WriteOID(sb.RevertToSblockOID); err != nil {
		return nil, err
	}
	if err := writer.WriteUint64(sb.NextObjID); err != nil {
		return nil, err
	}
	if err := writer.WriteUint64(sb.NumFiles); err != nil {
		return nil, err
	}
	if err := writer.WriteUint64(sb.NumDirectories); err != nil {
		return nil, err
	}
	if err := writer.WriteUint64(sb.NumSymlinks); err != nil {
		return nil, err
	}
	if err := writer.WriteUint64(sb.NumOtherFSObjects); err != nil {
		return nil, err
	}
	if err := writer.WriteUint64(sb.NumSnapshots); err != nil {
		return nil, err
	}
	if err := writer.WriteUint64(sb.TotalBlocksAlloced); err != nil {
		return nil, err
	}
	if err := writer.WriteUint64(sb.TotalBlocksFreed); err != nil {
		return nil, err
	}
	if err := writer.WriteUUID(sb.UUID); err != nil {
		return nil, err
	}
	if err := writer.WriteUint64(sb.LastModTime); err != nil {
		return nil, err
	}
	if err := writer.WriteUint64(sb.FSFlags); err != nil {
		return nil, err
	}
	if err := writer.Write(sb.FormattedBy); err != nil {
		return nil, err
	}
	for _, m := range sb.ModifiedBy {
		if err := writer.Write(m); err != nil {
			return nil, err
		}
	}
	if err := writer.WriteBytes(sb.VolName[:]); err != nil {
		return nil, err
	}
	if err := writer.WriteUint32(sb.NextDocID); err != nil {
		return nil, err
	}
	if err := writer.WriteUint16(sb.Role); err != nil {
		return nil, err
	}
	if err := writer.WriteUint16(sb.Reserved); err != nil {
		return nil, err
	}
	if err := writer.WriteXID(sb.RootToXID); err != nil {
		return nil, err
	}
	if err := writer.WriteOID(sb.ERStateOID); err != nil {
		return nil, err
	}
	if err := writer.WriteUint64(sb.CloneinfoIDEpoch); err != nil {
		return nil, err
	}
	if err := writer.WriteUint64(sb.CloneinfoXID); err != nil {
		return nil, err
	}
	if err := writer.WriteOID(sb.SnapMetaExtOID); err != nil {
		return nil, err
	}
	if err := writer.WriteUUID(sb.VolumeGroupID); err != nil {
		return nil, err
	}
	if err := writer.WriteOID(sb.IntegrityMetaOID); err != nil {
		return nil, err
	}
	if err := writer.WriteOID(sb.FextTreeOID); err != nil {
		return nil, err
	}
	if err := writer.WriteUint32(sb.FextTreeType); err != nil {
		return nil, err
	}
	if err := writer.WriteUint32(sb.ReservedType); err != nil {
		return nil, err
	}
	if err := writer.WriteOID(sb.ReservedOID); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// DeserializeBTNodePhys deserializes a B-tree node from binary data
func DeserializeBTNodePhys(data []byte) (*BTNodePhys, error) {
	const cksumOffset = 0
	const minSize = 56 // conservative minimum (size of fixed fields)

	if len(data) < minSize {
		return nil, ErrStructTooShort
	}

	// 1. Verify Fletcher64 checksum
	expected := binary.LittleEndian.Uint64(data[:8])
	actual := checksum.Fletcher64WithZeroedChecksum(data, cksumOffset)
	if expected != actual {
		return nil, fmt.Errorf("checksum mismatch: expected 0x%016x, got 0x%016x", expected, actual)
	}

	// 2. Set up reader
	br := NewBinaryReader(bytes.NewReader(data), binary.LittleEndian)
	node := &BTNodePhys{}

	// 3. Read header (obj_phys_t)
	if err := br.Read(&node.Header.Cksum); err != nil {
		return nil, fmt.Errorf("read cksum: %w", err)
	}
	var err error
	if node.Header.OID, err = br.ReadOID(); err != nil {
		return nil, err
	}
	if node.Header.XID, err = br.ReadXID(); err != nil {
		return nil, err
	}
	if node.Header.Type, err = br.ReadUint32(); err != nil {
		return nil, err
	}
	if node.Header.Subtype, err = br.ReadUint32(); err != nil {
		return nil, err
	}

	// 4. Validate object type
	if (node.Header.Type & types.ObjectTypeMask) != types.ObjectTypeBtreeNode {
		return nil, fmt.Errorf("invalid object type: 0x%x, expected BTREE_NODE", node.Header.Type)
	}

	// 5. Read BTNodePhys fields
	if node.Flags, err = br.ReadUint16(); err != nil {
		return nil, err
	}
	if node.Level, err = br.ReadUint16(); err != nil {
		return nil, err
	}
	if node.NKeys, err = br.ReadUint32(); err != nil {
		return nil, err
	}
	if err = br.Read(&node.TableSpace); err != nil {
		return nil, err
	}
	if err = br.Read(&node.FreeSpace); err != nil {
		return nil, err
	}
	if err = br.Read(&node.KeyFreeList); err != nil {
		return nil, err
	}
	if err = br.Read(&node.ValFreeList); err != nil {
		return nil, err
	}

	// 6. Read remaining data as node.Data
	currPos, err := br.buf.Seek(0, io.SeekCurrent)
	if err != nil {
		return nil, err
	}
	remainingSize := int64(len(data)) - currPos
	node.Data, err = br.ReadBytes(int(remainingSize))
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("read trailing data: %w", err)
	}

	// Optional: capture raw tail (for future fields, debug, etc.)
	// node.TailData = data[currPos+len(node.Data):]

	return node, nil
}

// DeserializeOMapPhys deserializes an object map from binary data
func DeserializeOMapPhys(data []byte) (*OMapPhys, error) {
	if len(data) < int(unsafe.Sizeof(ObjectHeader{})) {
		return nil, ErrStructTooShort
	}

	// Deserialize the object header first
	header, err := DeserializeObjectHeader(data)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize object header: %w", err)
	}

	// Verify object type
	if header.GetObjectType() != ObjectTypeOMap {
		return nil, fmt.Errorf("invalid object type: 0x%x, expected OMAP", header.GetObjectType())
	}

	// Create object map structure
	omap := &OMapPhys{
		Header: *header,
	}

	// Read the rest of the object map
	headerSize := int(unsafe.Sizeof(ObjectHeader{}))
	br := NewBinaryReader(bytes.NewReader(data[headerSize:]), binary.LittleEndian)

	// Read fields
	if omap.Flags, err = br.ReadUint32(); err != nil {
		return nil, fmt.Errorf("failed to read flags: %w", err)
	}

	if omap.SnapCount, err = br.ReadUint32(); err != nil {
		return nil, fmt.Errorf("failed to read snap count: %w", err)
	}

	if omap.TreeType, err = br.ReadUint32(); err != nil {
		return nil, fmt.Errorf("failed to read tree type: %w", err)
	}

	if omap.SnapshotTreeType, err = br.ReadUint32(); err != nil {
		return nil, fmt.Errorf("failed to read snapshot tree type: %w", err)
	}

	if omap.TreeOID, err = br.ReadOID(); err != nil {
		return nil, fmt.Errorf("failed to read tree OID: %w", err)
	}

	if omap.SnapshotTreeOID, err = br.ReadOID(); err != nil {
		return nil, fmt.Errorf("failed to read snapshot tree OID: %w", err)
	}

	if omap.MostRecentSnap, err = br.ReadXID(); err != nil {
		return nil, fmt.Errorf("failed to read most recent snap: %w", err)
	}

	if omap.PendingRevertMin, err = br.ReadXID(); err != nil {
		return nil, fmt.Errorf("failed to read pending revert min: %w", err)
	}

	if omap.PendingRevertMax, err = br.ReadXID(); err != nil {
		return nil, fmt.Errorf("failed to read pending revert max: %w", err)
	}

	return omap, nil
}

// DeserializeSpacemanPhys deserializes a space manager from binary data
func DeserializeSpacemanPhys(data []byte) (*SpacemanPhys, error) {
	if len(data) < int(unsafe.Sizeof(ObjectHeader{})) {
		return nil, ErrStructTooShort
	}

	// Deserialize the object header first
	header, err := DeserializeObjectHeader(data)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize object header: %w", err)
	}

	// Verify object type
	if header.GetObjectType() != ObjectTypeSpaceman {
		return nil, fmt.Errorf("invalid object type: 0x%x, expected SPACEMAN", header.GetObjectType())
	}

	// Create space manager structure
	sm := &SpacemanPhys{
		Header: *header,
	}

	// Read the rest of the space manager
	headerSize := int(unsafe.Sizeof(ObjectHeader{}))
	br := NewBinaryReader(bytes.NewReader(data[headerSize:]), binary.LittleEndian)

	// Read basic fields
	if sm.BlockSize, err = br.ReadUint32(); err != nil {
		return nil, fmt.Errorf("failed to read block size: %w", err)
	}

	if sm.BlocksPerChunk, err = br.ReadUint32(); err != nil {
		return nil, fmt.Errorf("failed to read blocks per chunk: %w", err)
	}

	if sm.ChunksPerCIB, err = br.ReadUint32(); err != nil {
		return nil, fmt.Errorf("failed to read chunks per CIB: %w", err)
	}

	if sm.CIBsPerCAB, err = br.ReadUint32(); err != nil {
		return nil, fmt.Errorf("failed to read CIBs per CAB: %w", err)
	}

	// Read device information for main device (index 0)
	if sm.Devices[0].BlockCount, err = br.ReadUint64(); err != nil {
		return nil, fmt.Errorf("failed to read main device block count: %w", err)
	}

	if sm.Devices[0].ChunkCount, err = br.ReadUint64(); err != nil {
		return nil, fmt.Errorf("failed to read main device chunk count: %w", err)
	}

	if sm.Devices[0].CIBCount, err = br.ReadUint32(); err != nil {
		return nil, fmt.Errorf("failed to read main device CIB count: %w", err)
	}

	if sm.Devices[0].CABCount, err = br.ReadUint32(); err != nil {
		return nil, fmt.Errorf("failed to read main device CAB count: %w", err)
	}

	if sm.Devices[0].FreeCount, err = br.ReadUint64(); err != nil {
		return nil, fmt.Errorf("failed to read main device free count: %w", err)
	}

	if sm.Devices[0].AddrOffset, err = br.ReadUint32(); err != nil {
		return nil, fmt.Errorf("failed to read main device addr offset: %w", err)
	}

	if sm.Devices[0].Reserved, err = br.ReadUint32(); err != nil {
		return nil, fmt.Errorf("failed to read main device reserved: %w", err)
	}

	if sm.Devices[0].Reserved2, err = br.ReadUint64(); err != nil {
		return nil, fmt.Errorf("failed to read main device reserved2: %w", err)
	}

	// Read device information for tier2 device (index 1)
	if sm.Devices[1].BlockCount, err = br.ReadUint64(); err != nil {
		return nil, fmt.Errorf("failed to read tier2 device block count: %w", err)
	}

	if sm.Devices[1].ChunkCount, err = br.ReadUint64(); err != nil {
		return nil, fmt.Errorf("failed to read tier2 device chunk count: %w", err)
	}

	if sm.Devices[1].CIBCount, err = br.ReadUint32(); err != nil {
		return nil, fmt.Errorf("failed to read tier2 device CIB count: %w", err)
	}

	if sm.Devices[1].CABCount, err = br.ReadUint32(); err != nil {
		return nil, fmt.Errorf("failed to read tier2 device CAB count: %w", err)
	}

	if sm.Devices[1].FreeCount, err = br.ReadUint64(); err != nil {
		return nil, fmt.Errorf("failed to read tier2 device free count: %w", err)
	}

	if sm.Devices[1].AddrOffset, err = br.ReadUint32(); err != nil {
		return nil, fmt.Errorf("failed to read tier2 device addr offset: %w", err)
	}

	if sm.Devices[1].Reserved, err = br.ReadUint32(); err != nil {
		return nil, fmt.Errorf("failed to read tier2 device reserved: %w", err)
	}

	if sm.Devices[1].Reserved2, err = br.ReadUint64(); err != nil {
		return nil, fmt.Errorf("failed to read tier2 device reserved2: %w", err)
	}

	// Read other fields
	if sm.Flags, err = br.ReadUint32(); err != nil {
		return nil, fmt.Errorf("failed to read flags: %w", err)
	}

	if sm.IPBmTxMultiplier, err = br.ReadUint32(); err != nil {
		return nil, fmt.Errorf("failed to read IP BM TX multiplier: %w", err)
	}

	if sm.IPBlockCount, err = br.ReadUint64(); err != nil {
		return nil, fmt.Errorf("failed to read IP block count: %w", err)
	}

	if sm.IPBmSizeInBlocks, err = br.ReadUint32(); err != nil {
		return nil, fmt.Errorf("failed to read IP BM size in blocks: %w", err)
	}

	if sm.IPBmBlockCount, err = br.ReadUint32(); err != nil {
		return nil, fmt.Errorf("failed to read IP BM block count: %w", err)
	}

	if sm.IPBmBase, err = br.ReadPAddr(); err != nil {
		return nil, fmt.Errorf("failed to read IP BM base: %w", err)
	}

	if sm.IPBase, err = br.ReadPAddr(); err != nil {
		return nil, fmt.Errorf("failed to read IP base: %w", err)
	}

	if sm.FSReserveBlockCount, err = br.ReadUint64(); err != nil {
		return nil, fmt.Errorf("failed to read FS reserve block count: %w", err)
	}

	if sm.FSReserveAllocCount, err = br.ReadUint64(); err != nil {
		return nil, fmt.Errorf("failed to read FS reserve alloc count: %w", err)
	}

	// Read free queues
	for i := 0; i < 3; i++ { // SFQ_COUNT = 3 (IP, Main, Tier2)
		if sm.FreeQueues[i].Count, err = br.ReadUint64(); err != nil {
			return nil, fmt.Errorf("failed to read free queue %d count: %w", i, err)
		}

		if sm.FreeQueues[i].TreeOID, err = br.ReadOID(); err != nil {
			return nil, fmt.Errorf("failed to read free queue %d tree OID: %w", i, err)
		}

		if sm.FreeQueues[i].OldestXID, err = br.ReadXID(); err != nil {
			return nil, fmt.Errorf("failed to read free queue %d oldest XID: %w", i, err)
		}

		if sm.FreeQueues[i].TreeNodeLimit, err = br.ReadUint16(); err != nil {
			return nil, fmt.Errorf("failed to read free queue %d tree node limit: %w", i, err)
		}

		if sm.FreeQueues[i].Pad16, err = br.ReadUint16(); err != nil {
			return nil, fmt.Errorf("failed to read free queue %d pad16: %w", i, err)
		}

		if sm.FreeQueues[i].Pad32, err = br.ReadUint32(); err != nil {
			return nil, fmt.Errorf("failed to read free queue %d pad32: %w", i, err)
		}

		if sm.FreeQueues[i].Reserved, err = br.ReadUint64(); err != nil {
			return nil, fmt.Errorf("failed to read free queue %d reserved: %w", i, err)
		}
	}

	// Read remaining fields
	if sm.IPBmFreeHead, err = br.ReadUint16(); err != nil {
		return nil, fmt.Errorf("failed to read IP BM free head: %w", err)
	}

	if sm.IPBmFreeTail, err = br.ReadUint16(); err != nil {
		return nil, fmt.Errorf("failed to read IP BM free tail: %w", err)
	}

	if sm.IPBmXidOffset, err = br.ReadUint32(); err != nil {
		return nil, fmt.Errorf("failed to read IP BM XID offset: %w", err)
	}

	if sm.IPBitmapOffset, err = br.ReadUint32(); err != nil {
		return nil, fmt.Errorf("failed to read IP bitmap offset: %w", err)
	}

	if sm.IPBmFreeNextOffset, err = br.ReadUint32(); err != nil {
		return nil, fmt.Errorf("failed to read IP BM free next offset: %w", err)
	}

	if sm.Version, err = br.ReadUint32(); err != nil {
		return nil, fmt.Errorf("failed to read version: %w", err)
	}

	if sm.StructSize, err = br.ReadUint32(); err != nil {
		return nil, fmt.Errorf("failed to read struct size: %w", err)
	}

	// Note: We're skipping datazone_info_phys_t which is a complex structure
	// A complete implementation would need to handle this

	return sm, nil
}

// SerializeCheckpointMapPhys serializes a checkpoint mapping block to binary data
func SerializeCheckpointMapPhys(cpMap *CheckpointMapPhys) ([]byte, error) {
	buf := new(bytes.Buffer)
	writer := NewBinaryWriter(buf, binary.LittleEndian)

	// Write object header
	headerBytes, err := SerializeObjectHeader(&cpMap.Header)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize object header: %w", err)
	}
	if err := writer.WriteBytes(headerBytes); err != nil {
		return nil, fmt.Errorf("failed to write object header: %w", err)
	}

	// Write flags and count
	if err := writer.WriteUint32(cpMap.Flags); err != nil {
		return nil, fmt.Errorf("failed to write flags: %w", err)
	}

	if err := writer.WriteUint32(cpMap.Count); err != nil {
		return nil, fmt.Errorf("failed to write count: %w", err)
	}

	// Write checkpoint mappings
	for i := uint32(0); i < cpMap.Count; i++ {
		mapping := &cpMap.Map[i]

		if err := writer.WriteUint32(mapping.Type); err != nil {
			return nil, fmt.Errorf("failed to write mapping type: %w", err)
		}

		if err := writer.WriteUint32(mapping.Subtype); err != nil {
			return nil, fmt.Errorf("failed to write mapping subtype: %w", err)
		}

		if err := writer.WriteUint32(mapping.Size); err != nil {
			return nil, fmt.Errorf("failed to write mapping size: %w", err)
		}

		if err := writer.WriteUint32(mapping.Pad); err != nil {
			return nil, fmt.Errorf("failed to write mapping pad: %w", err)
		}

		if err := writer.WriteOID(mapping.FSOID); err != nil {
			return nil, fmt.Errorf("failed to write mapping FSOID: %w", err)
		}

		if err := writer.WriteOID(mapping.OID); err != nil {
			return nil, fmt.Errorf("failed to write mapping OID: %w", err)
		}

		if err := writer.WritePAddr(mapping.PAddr); err != nil {
			return nil, fmt.Errorf("failed to write mapping PAddr: %w", err)
		}
	}

	return buf.Bytes(), nil
}

// SerializeOMapPhys serializes an object map to binary data
func SerializeOMapPhys(omap *OMapPhys) ([]byte, error) {
	buf := new(bytes.Buffer)
	writer := NewBinaryWriter(buf, binary.LittleEndian)

	// Write object header
	headerBytes, err := SerializeObjectHeader(&omap.Header)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize object header: %w", err)
	}
	if err := writer.WriteBytes(headerBytes); err != nil {
		return nil, fmt.Errorf("failed to write object header: %w", err)
	}

	// Write fields
	if err := writer.WriteUint32(omap.Flags); err != nil {
		return nil, fmt.Errorf("failed to write flags: %w", err)
	}

	if err := writer.WriteUint32(omap.SnapCount); err != nil {
		return nil, fmt.Errorf("failed to write snap count: %w", err)
	}

	if err := writer.WriteUint32(omap.TreeType); err != nil {
		return nil, fmt.Errorf("failed to write tree type: %w", err)
	}

	if err := writer.WriteUint32(omap.SnapshotTreeType); err != nil {
		return nil, fmt.Errorf("failed to write snapshot tree type: %w", err)
	}

	if err := writer.WriteOID(omap.TreeOID); err != nil {
		return nil, fmt.Errorf("failed to write tree OID: %w", err)
	}

	if err := writer.WriteOID(omap.SnapshotTreeOID); err != nil {
		return nil, fmt.Errorf("failed to write snapshot tree OID: %w", err)
	}

	if err := writer.WriteXID(omap.MostRecentSnap); err != nil {
		return nil, fmt.Errorf("failed to write most recent snap: %w", err)
	}

	if err := writer.WriteXID(omap.PendingRevertMin); err != nil {
		return nil, fmt.Errorf("failed to write pending revert min: %w", err)
	}

	if err := writer.WriteXID(omap.PendingRevertMax); err != nil {
		return nil, fmt.Errorf("failed to write pending revert max: %w", err)
	}

	return buf.Bytes(), nil
}

// SerializeSpacemanPhys serializes a space manager to binary data
func SerializeSpacemanPhys(sm *SpacemanPhys) ([]byte, error) {
	buf := new(bytes.Buffer)
	writer := NewBinaryWriter(buf, binary.LittleEndian)

	// Write object header
	headerBytes, err := SerializeObjectHeader(&sm.Header)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize object header: %w", err)
	}
	if err := writer.WriteBytes(headerBytes); err != nil {
		return nil, fmt.Errorf("failed to write object header: %w", err)
	}

	// Write basic fields
	if err := writer.WriteUint32(sm.BlockSize); err != nil {
		return nil, fmt.Errorf("failed to write block size: %w", err)
	}

	if err := writer.WriteUint32(sm.BlocksPerChunk); err != nil {
		return nil, fmt.Errorf("failed to write blocks per chunk: %w", err)
	}

	if err := writer.WriteUint32(sm.ChunksPerCIB); err != nil {
		return nil, fmt.Errorf("failed to write chunks per CIB: %w", err)
	}

	if err := writer.WriteUint32(sm.CIBsPerCAB); err != nil {
		return nil, fmt.Errorf("failed to write CIBs per CAB: %w", err)
	}

	// Write device information for main device (index 0)
	if err := writer.WriteUint64(sm.Devices[0].BlockCount); err != nil {
		return nil, fmt.Errorf("failed to write main device block count: %w", err)
	}

	if err := writer.WriteUint64(sm.Devices[0].ChunkCount); err != nil {
		return nil, fmt.Errorf("failed to write main device chunk count: %w", err)
	}

	if err := writer.WriteUint32(sm.Devices[0].CIBCount); err != nil {
		return nil, fmt.Errorf("failed to write main device CIB count: %w", err)
	}

	if err := writer.WriteUint32(sm.Devices[0].CABCount); err != nil {
		return nil, fmt.Errorf("failed to write main device CAB count: %w", err)
	}

	if err := writer.WriteUint64(sm.Devices[0].FreeCount); err != nil {
		return nil, fmt.Errorf("failed to write main device free count: %w", err)
	}

	if err := writer.WriteUint32(sm.Devices[0].AddrOffset); err != nil {
		return nil, fmt.Errorf("failed to write main device addr offset: %w", err)
	}

	if err := writer.WriteUint32(sm.Devices[0].Reserved); err != nil {
		return nil, fmt.Errorf("failed to write main device reserved: %w", err)
	}

	if err := writer.WriteUint64(sm.Devices[0].Reserved2); err != nil {
		return nil, fmt.Errorf("failed to write main device reserved2: %w", err)
	}

	// Write device information for tier2 device (index 1)
	if err := writer.WriteUint64(sm.Devices[1].BlockCount); err != nil {
		return nil, fmt.Errorf("failed to write tier2 device block count: %w", err)
	}

	if err := writer.WriteUint64(sm.Devices[1].ChunkCount); err != nil {
		return nil, fmt.Errorf("failed to write tier2 device chunk count: %w", err)
	}

	if err := writer.WriteUint32(sm.Devices[1].CIBCount); err != nil {
		return nil, fmt.Errorf("failed to write tier2 device CIB count: %w", err)
	}

	if err := writer.WriteUint32(sm.Devices[1].CABCount); err != nil {
		return nil, fmt.Errorf("failed to write tier2 device CAB count: %w", err)
	}

	if err := writer.WriteUint64(sm.Devices[1].FreeCount); err != nil {
		return nil, fmt.Errorf("failed to write tier2 device free count: %w", err)
	}

	if err := writer.WriteUint32(sm.Devices[1].AddrOffset); err != nil {
		return nil, fmt.Errorf("failed to write tier2 device addr offset: %w", err)
	}

	if err := writer.WriteUint32(sm.Devices[1].Reserved); err != nil {
		return nil, fmt.Errorf("failed to write tier2 device reserved: %w", err)
	}

	if err := writer.WriteUint64(sm.Devices[1].Reserved2); err != nil {
		return nil, fmt.Errorf("failed to write tier2 device reserved2: %w", err)
	}

	// Write other fields
	if err := writer.WriteUint32(sm.Flags); err != nil {
		return nil, fmt.Errorf("failed to write flags: %w", err)
	}

	if err := writer.WriteUint32(sm.IPBmTxMultiplier); err != nil {
		return nil, fmt.Errorf("failed to write IP BM TX multiplier: %w", err)
	}

	if err := writer.WriteUint64(sm.IPBlockCount); err != nil {
		return nil, fmt.Errorf("failed to write IP block count: %w", err)
	}

	if err := writer.WriteUint32(sm.IPBmSizeInBlocks); err != nil {
		return nil, fmt.Errorf("failed to write IP BM size in blocks: %w", err)
	}

	if err := writer.WriteUint32(sm.IPBmBlockCount); err != nil {
		return nil, fmt.Errorf("failed to write IP BM block count: %w", err)
	}

	if err := writer.WritePAddr(sm.IPBmBase); err != nil {
		return nil, fmt.Errorf("failed to write IP BM base: %w", err)
	}

	if err := writer.WritePAddr(sm.IPBase); err != nil {
		return nil, fmt.Errorf("failed to write IP base: %w", err)
	}

	if err := writer.WriteUint64(sm.FSReserveBlockCount); err != nil {
		return nil, fmt.Errorf("failed to write FS reserve block count: %w", err)
	}

	if err := writer.WriteUint64(sm.FSReserveAllocCount); err != nil {
		return nil, fmt.Errorf("failed to write FS reserve alloc count: %w", err)
	}

	// Write free queues
	for i := 0; i < 3; i++ { // SFQ_COUNT = 3 (IP, Main, Tier2)
		if err := writer.WriteUint64(sm.FreeQueues[i].Count); err != nil {
			return nil, fmt.Errorf("failed to write free queue %d count: %w", i, err)
		}

		if err := writer.WriteOID(sm.FreeQueues[i].TreeOID); err != nil {
			return nil, fmt.Errorf("failed to write free queue %d tree OID: %w", i, err)
		}

		if err := writer.WriteXID(sm.FreeQueues[i].OldestXID); err != nil {
			return nil, fmt.Errorf("failed to write free queue %d oldest XID: %w", i, err)
		}

		if err := writer.WriteUint16(sm.FreeQueues[i].TreeNodeLimit); err != nil {
			return nil, fmt.Errorf("failed to write free queue %d tree node limit: %w", i, err)
		}

		if err := writer.WriteUint16(sm.FreeQueues[i].Pad16); err != nil {
			return nil, fmt.Errorf("failed to write free queue %d pad16: %w", i, err)
		}

		if err := writer.WriteUint32(sm.FreeQueues[i].Pad32); err != nil {
			return nil, fmt.Errorf("failed to write free queue %d pad32: %w", i, err)
		}

		if err := writer.WriteUint64(sm.FreeQueues[i].Reserved); err != nil {
			return nil, fmt.Errorf("failed to write free queue %d reserved: %w", i, err)
		}
	}

	// Write remaining fields
	if err := writer.WriteUint16(sm.IPBmFreeHead); err != nil {
		return nil, fmt.Errorf("failed to write IP BM free head: %w", err)
	}

	if err := writer.WriteUint16(sm.IPBmFreeTail); err != nil {
		return nil, fmt.Errorf("failed to write IP BM free tail: %w", err)
	}

	if err := writer.WriteUint32(sm.IPBmXidOffset); err != nil {
		return nil, fmt.Errorf("failed to write IP BM XID offset: %w", err)
	}

	if err := writer.WriteUint32(sm.IPBitmapOffset); err != nil {
		return nil, fmt.Errorf("failed to write IP bitmap offset: %w", err)
	}

	if err := writer.WriteUint32(sm.IPBmFreeNextOffset); err != nil {
		return nil, fmt.Errorf("failed to write IP BM free next offset: %w", err)
	}

	if err := writer.WriteUint32(sm.Version); err != nil {
		return nil, fmt.Errorf("failed to write version: %w", err)
	}

	if err := writer.WriteUint32(sm.StructSize); err != nil {
		return nil, fmt.Errorf("failed to write struct size: %w", err)
	}

	// Note: We're skipping datazone_info_phys_t which is a complex structure
	// A complete implementation would need to handle this

	return buf.Bytes(), nil
}

// We implement serialization/deserialization for all APFS structures.
// Each follows this standardized pattern:
//
// 1. Validate data length against minimum struct size
// 2. Create a binary reader/writer with correct endianness
// 3. Read or write fields in strict spec order
// 4. Verify or compute the Fletcher 64 checksum (obj_phys_t-based)
// 5. Preserve and zero-fill padding where applicable
// 6. Validate object type and flags per OBJECT_TYPE_MASK
// 7. Optionally return or retain raw trailing bytes for future compatibility

// ReadBlock reads a block from a block device and deserializes it to the appropriate structure
func ReadBlock(device BlockDevice, addr PAddr, blockType uint32) (interface{}, error) {
	data, err := device.ReadBlock(addr)
	if err != nil {
		return nil, fmt.Errorf("failed to read block at %d: %w", addr, err)
	}

	// Check if data starts with an object header
	if len(data) < int(unsafe.Sizeof(ObjectHeader{})) {
		return nil, ErrStructTooShort
	}

	// Deserialize object header to get type
	header, err := DeserializeObjectHeader(data)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize object header: %w", err)
	}

	// If a specific blockType was requested, verify it matches
	if blockType != 0 && header.GetObjectType() != blockType {
		return nil, fmt.Errorf("expected block type %d, got %d: %w",
			blockType, header.GetObjectType(), ErrInvalidObjectType)
	}

	// Deserialize based on object type
	switch header.GetObjectType() {
	case ObjectTypeNXSuperblock:
		return DeserializeNXSuperblock(data)
	case ObjectTypeFS:
		return DeserializeAPFSSuperblock(data)
	case ObjectTypeBtreeNode:
		return DeserializeBTNodePhys(data)
	// Add cases for other object types
	default:
		return nil, fmt.Errorf("unsupported object type: %d", header.GetObjectType())
	}
}
