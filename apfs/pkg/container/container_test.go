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
