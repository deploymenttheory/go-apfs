// File: pkg/container/container_test.go
package container

import (
	"encoding/binary"
	"testing"

	"github.com/deploymenttheory/go-apfs/apfs/pkg/checksum"
	"github.com/deploymenttheory/go-apfs/apfs/pkg/types"
)

func TestReadNXSuperblock(t *testing.T) {
	device := &MockBlockDevice{
		BlockSize: 4096,
		Blocks:    make(map[types.PAddr][]byte),
	}

	addr := types.PAddr(0x10)
	data := make([]byte, NXSuperblockSize)

	binary.LittleEndian.PutUint32(data[32:36], types.NXMagic)
	binary.LittleEndian.PutUint32(data[36:40], types.DefaultBlockSize)
	binary.LittleEndian.PutUint64(data[40:48], 1024)
	binary.LittleEndian.PutUint64(data[48:56], types.NXSupportedFeaturesMask)
	binary.LittleEndian.PutUint64(data[64:72], types.NXSupportedIncompatMask)

	// explicitly set MaxFileSystems (required for validation)
	binary.LittleEndian.PutUint32(data[180:184], types.NXMaxFileSystems)

	// valid OIDs
	binary.LittleEndian.PutUint64(data[152:160], 1234) // SpacemanOID
	binary.LittleEndian.PutUint64(data[160:168], 5678) // OMapOID
	binary.LittleEndian.PutUint64(data[168:176], 9012) // ReaperOID

	// NextOID
	binary.LittleEndian.PutUint64(data[88:96], uint64(types.OIDReservedCount+1))

	// Valid checksum
	csum := checksum.Fletcher64WithZeroedChecksum(data, 0)
	binary.LittleEndian.PutUint64(data[0:8], csum)

	device.Blocks[addr] = data

	sb, err := ReadNXSuperblock(device, addr)
	if err != nil {
		t.Fatalf("ReadNXSuperblock failed: %v", err)
	}

	if sb.Magic != types.NXMagic {
		t.Errorf("expected NXMagic 0x%x, got 0x%x", types.NXMagic, sb.Magic)
	}
	if sb.BlockSize != types.DefaultBlockSize {
		t.Errorf("expected BlockSize %d, got %d", types.DefaultBlockSize, sb.BlockSize)
	}
}

func TestWriteNXSuperblock(t *testing.T) {
	device := &MockBlockDevice{
		BlockSize: 4096,
		Blocks:    make(map[types.PAddr][]byte),
	}

	addr := types.PAddr(0x20)
	sb := &types.NXSuperblock{
		Header: types.ObjectHeader{
			OID:     123,
			XID:     456,
			Type:    1,
			Subtype: 2,
		},
		Magic:            types.NXMagic,
		BlockSize:        4096,
		BlockCount:       10000,
		Features:         types.NXSupportedFeaturesMask,
		IncompatFeatures: types.NXSupportedIncompatMask,
		MaxFileSystems:   types.NXMaxFileSystems,
		SpacemanOID:      11,
		OMapOID:          22,
		ReaperOID:        33,
		NextOID:          types.OIDReservedCount + 1,
	}

	err := WriteNXSuperblock(device, addr, sb)
	if err != nil {
		t.Fatalf("WriteNXSuperblock failed: %v", err)
	}

	block := device.Blocks[addr]
	if len(block) == 0 {
		t.Fatal("No data written to block")
	}

	// Check checksum was written
	writtenChecksum := binary.LittleEndian.Uint64(block[:8])
	if writtenChecksum == 0 {
		t.Error("Expected non-zero checksum")
	}
}

func TestWriteVolumeSuperblock(t *testing.T) {
	device := &MockBlockDevice{
		BlockSize: 4096,
		Blocks:    make(map[types.PAddr][]byte),
	}

	addr := types.PAddr(0x30)

	volume := &types.APFSSuperblock{
		Header: types.ObjectHeader{
			OID:     100,
			XID:     200,
			Type:    1,
			Subtype: 2,
		},
		Magic:                  0x42535041, // 'APSB'
		FSIndex:                1,
		Features:               0x12345678,
		ReadOnlyCompatFeatures: 0,
		IncompatFeatures:       0x9abcdef0,
		UnmountTime:            1234567890,
		ReserveBlockCount:      2,
		QuotaBlockCount:        2,
		AllocCount:             10,
		MetaCrypto: types.MetaCryptoState{
			MajorVersion:    1,
			MinorVersion:    0,
			Flags:           0,
			PersistentClass: 1,
			KeyOSVersion:    1234,
			KeyRevision:     1,
			Unused:          0,
		},
		RootTreeType:       1,
		ExtentrefTreeType:  2,
		SnapMetaTreeType:   3,
		OMapOID:            101,
		RootTreeOID:        102,
		ExtentrefTreeOID:   103,
		SnapMetaTreeOID:    104,
		RevertToXID:        300,
		RevertToSblockOID:  105,
		NextObjID:          500,
		NumFiles:           10,
		NumDirectories:     5,
		NumSymlinks:        2,
		NumOtherFSObjects:  1,
		NumSnapshots:       3,
		TotalBlocksAlloced: 100,
		TotalBlocksFreed:   50,
		UUID:               types.UUID{0xaa, 0xbb, 0xcc, 0xdd},
		LastModTime:        987654321,
		FSFlags:            0xdeadbeef,
		FormattedBy: types.ModifiedBy{
			ID:        [32]byte{'T', 'E', 'S', 'T'},
			Timestamp: 1680000000000000000,
			LastXID:   123,
		},
		ModifiedBy: [8]types.ModifiedBy{
			{ID: [32]byte{'T', 'E', 'S', 'T'}, Timestamp: 1680000000, LastXID: 1},
			{ID: [32]byte{'A', 'P', 'F', 'S'}, Timestamp: 1680000001, LastXID: 2},
			{}, {}, {}, {}, {}, {}, // Remaining entries to fill all 8
		},
		VolName:          [256]byte{'V', 'o', 'l', '1'},
		NextDocID:        777,
		Role:             1,
		Reserved:         0,
		RootToXID:        600,
		ERStateOID:       106,
		CloneinfoIDEpoch: 1000,
		CloneinfoXID:     2000,
		SnapMetaExtOID:   107,
		VolumeGroupID:    types.UUID{0x01, 0x02, 0x03},
		IntegrityMetaOID: 108,
		FextTreeOID:      109,
		FextTreeType:     5,
		ReservedType:     0,
		ReservedOID:      0,
	}

	err := WriteVolumeSuperblock(device, addr, volume)
	if err != nil {
		t.Fatalf("WriteVolumeSuperblock failed: %v", err)
	}

	block := device.Blocks[addr]
	if len(block) == 0 {
		t.Fatal("No data written to volume block")
	}

	if string(block[40:44]) != "APSB" && binary.LittleEndian.Uint32(block[40:44]) != volume.Magic {
		t.Errorf("Magic mismatch: expected 0x%x", volume.Magic)
	}
}

func TestValidateNXSuperblock(t *testing.T) {
	validSuperblock := &types.NXSuperblock{
		Magic:            types.NXMagic,
		BlockSize:        types.DefaultBlockSize,
		BlockCount:       1000,
		MaxFileSystems:   types.NXMaxFileSystems,
		SpacemanOID:      1,
		OMapOID:          2,
		ReaperOID:        3,
		Features:         types.NXSupportedFeaturesMask,
		IncompatFeatures: types.NXSupportedIncompatMask,
		NextOID:          types.OIDReservedCount + 1,
	}

	if err := ValidateNXSuperblock(validSuperblock); err != nil {
		t.Errorf("expected no validation errors, got: %v", err)
	}

	// Test invalid magic
	invalidMagic := *validSuperblock
	invalidMagic.Magic = 0
	if err := ValidateNXSuperblock(&invalidMagic); err == nil {
		t.Errorf("expected error for invalid magic, got nil")
	}

	// Test invalid block size
	invalidBlockSize := *validSuperblock
	invalidBlockSize.BlockSize = types.MinBlockSize - 1
	if err := ValidateNXSuperblock(&invalidBlockSize); err == nil {
		t.Errorf("expected error for invalid block size, got nil")
	}

	// Test invalid OIDs
	invalidOID := *validSuperblock
	invalidOID.SpacemanOID = types.OIDInvalid
	if err := ValidateNXSuperblock(&invalidOID); err == nil {
		t.Errorf("expected error for invalid SpacemanOID, got nil")
	}

	// Test unsupported features
	unsupportedFeatures := *validSuperblock
	unsupportedFeatures.Features = ^types.NXSupportedFeaturesMask
	if err := ValidateNXSuperblock(&unsupportedFeatures); err == nil {
		t.Errorf("expected error for unsupported features, got nil")
	}
}
