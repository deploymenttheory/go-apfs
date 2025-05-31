package container

import (
	"testing"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

// createTestSuperblockForCheckpoints creates a test superblock with checkpoint data
func createTestSuperblockForCheckpoints(
	descBlocks, dataBlocks uint32,
	descBase, dataBase types.Paddr,
	descNext, dataNext, descIndex, descLen, dataIndex, dataLen uint32,
) *types.NxSuperblockT {
	return &types.NxSuperblockT{
		NxXpDescBlocks: descBlocks,
		NxXpDataBlocks: dataBlocks,
		NxXpDescBase:   descBase,
		NxXpDataBase:   dataBase,
		NxXpDescNext:   descNext,
		NxXpDataNext:   dataNext,
		NxXpDescIndex:  descIndex,
		NxXpDescLen:    descLen,
		NxXpDataIndex:  dataIndex,
		NxXpDataLen:    dataLen,
	}
}

func TestContainerCheckpointManager(t *testing.T) {
	tests := []struct {
		name       string
		descBlocks uint32
		dataBlocks uint32
		descBase   types.Paddr
		dataBase   types.Paddr
		descNext   uint32
		dataNext   uint32
		descIndex  uint32
		descLen    uint32
		dataIndex  uint32
		dataLen    uint32
	}{
		{
			name:       "Basic checkpoint configuration",
			descBlocks: 100,
			dataBlocks: 500,
			descBase:   types.Paddr(0x1000),
			dataBase:   types.Paddr(0x5000),
			descNext:   10,
			dataNext:   50,
			descIndex:  5,
			descLen:    20,
			dataIndex:  25,
			dataLen:    100,
		},
		{
			name:       "Zero values",
			descBlocks: 0,
			dataBlocks: 0,
			descBase:   types.Paddr(0),
			dataBase:   types.Paddr(0),
			descNext:   0,
			dataNext:   0,
			descIndex:  0,
			descLen:    0,
			dataIndex:  0,
			dataLen:    0,
		},
		{
			name:       "Large checkpoint",
			descBlocks: 0x10000,
			dataBlocks: 0x100000,
			descBase:   types.Paddr(0x100000000),
			dataBase:   types.Paddr(0x200000000),
			descNext:   0x1000,
			dataNext:   0x10000,
			descIndex:  0x500,
			descLen:    0x2000,
			dataIndex:  0x5000,
			dataLen:    0x20000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			superblock := createTestSuperblockForCheckpoints(
				tt.descBlocks, tt.dataBlocks,
				tt.descBase, tt.dataBase,
				tt.descNext, tt.dataNext,
				tt.descIndex, tt.descLen,
				tt.dataIndex, tt.dataLen,
			)
			manager := NewContainerCheckpointManager(superblock)

			// Test block counts
			if descBlocks := manager.CheckpointDescriptorBlockCount(); descBlocks != tt.descBlocks {
				t.Errorf("CheckpointDescriptorBlockCount() = %d, want %d", descBlocks, tt.descBlocks)
			}

			if dataBlocks := manager.CheckpointDataBlockCount(); dataBlocks != tt.dataBlocks {
				t.Errorf("CheckpointDataBlockCount() = %d, want %d", dataBlocks, tt.dataBlocks)
			}

			// Test base addresses
			if descBase := manager.CheckpointDescriptorBase(); descBase != tt.descBase {
				t.Errorf("CheckpointDescriptorBase() = 0x%X, want 0x%X", descBase, tt.descBase)
			}

			if dataBase := manager.CheckpointDataBase(); dataBase != tt.dataBase {
				t.Errorf("CheckpointDataBase() = 0x%X, want 0x%X", dataBase, tt.dataBase)
			}

			// Test next indices
			if descNext := manager.CheckpointDescriptorNext(); descNext != tt.descNext {
				t.Errorf("CheckpointDescriptorNext() = %d, want %d", descNext, tt.descNext)
			}

			if dataNext := manager.CheckpointDataNext(); dataNext != tt.dataNext {
				t.Errorf("CheckpointDataNext() = %d, want %d", dataNext, tt.dataNext)
			}

			// Test indices
			if descIndex := manager.CheckpointDescriptorIndex(); descIndex != tt.descIndex {
				t.Errorf("CheckpointDescriptorIndex() = %d, want %d", descIndex, tt.descIndex)
			}

			if dataIndex := manager.CheckpointDataIndex(); dataIndex != tt.dataIndex {
				t.Errorf("CheckpointDataIndex() = %d, want %d", dataIndex, tt.dataIndex)
			}

			// Test lengths
			if descLen := manager.CheckpointDescriptorLength(); descLen != tt.descLen {
				t.Errorf("CheckpointDescriptorLength() = %d, want %d", descLen, tt.descLen)
			}

			if dataLen := manager.CheckpointDataLength(); dataLen != tt.dataLen {
				t.Errorf("CheckpointDataLength() = %d, want %d", dataLen, tt.dataLen)
			}
		})
	}
}

func TestContainerCheckpointManager_FlagMasking(t *testing.T) {
	// Test that the highest bit is properly masked out for block counts
	tests := []struct {
		name               string
		descBlocks         uint32
		dataBlocks         uint32
		expectedDescBlocks uint32
		expectedDataBlocks uint32
	}{
		{
			name:               "No high bit set",
			descBlocks:         100,
			dataBlocks:         200,
			expectedDescBlocks: 100,
			expectedDataBlocks: 200,
		},
		{
			name:               "High bit set in desc blocks",
			descBlocks:         0x80000064, // High bit set + 100
			dataBlocks:         200,
			expectedDescBlocks: 100, // High bit masked out
			expectedDataBlocks: 200,
		},
		{
			name:               "High bit set in data blocks",
			descBlocks:         100,
			dataBlocks:         0x800000C8, // High bit set + 200
			expectedDescBlocks: 100,
			expectedDataBlocks: 200, // High bit masked out
		},
		{
			name:               "High bit set in both",
			descBlocks:         0x80000064, // High bit set + 100
			dataBlocks:         0x800000C8, // High bit set + 200
			expectedDescBlocks: 100,        // High bit masked out
			expectedDataBlocks: 200,        // High bit masked out
		},
		{
			name:               "Maximum values with high bit",
			descBlocks:         0xFFFFFFFF, // All bits set
			dataBlocks:         0xFFFFFFFF, // All bits set
			expectedDescBlocks: 0x7FFFFFFF, // High bit masked out
			expectedDataBlocks: 0x7FFFFFFF, // High bit masked out
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			superblock := &types.NxSuperblockT{
				NxXpDescBlocks: tt.descBlocks,
				NxXpDataBlocks: tt.dataBlocks,
			}
			manager := NewContainerCheckpointManager(superblock)

			if descBlocks := manager.CheckpointDescriptorBlockCount(); descBlocks != tt.expectedDescBlocks {
				t.Errorf("CheckpointDescriptorBlockCount() = %d, want %d", descBlocks, tt.expectedDescBlocks)
			}

			if dataBlocks := manager.CheckpointDataBlockCount(); dataBlocks != tt.expectedDataBlocks {
				t.Errorf("CheckpointDataBlockCount() = %d, want %d", dataBlocks, tt.expectedDataBlocks)
			}
		})
	}
}

func TestContainerCheckpointManager_EdgeCases(t *testing.T) {
	// Test edge cases with maximum and minimum values
	superblock := createTestSuperblockForCheckpoints(
		0xFFFFFFFF, 0xFFFFFFFF,
		types.Paddr(0x7FFFFFFFFFFFFFFF), types.Paddr(0x7FFFFFFFFFFFFFFF), // Max signed int64
		0xFFFFFFFF, 0xFFFFFFFF,
		0xFFFFFFFF, 0xFFFFFFFF,
		0xFFFFFFFF, 0xFFFFFFFF,
	)
	manager := NewContainerCheckpointManager(superblock)

	// Test that all methods return values without panicking
	_ = manager.CheckpointDescriptorBlockCount()
	_ = manager.CheckpointDataBlockCount()
	_ = manager.CheckpointDescriptorBase()
	_ = manager.CheckpointDataBase()
	_ = manager.CheckpointDescriptorNext()
	_ = manager.CheckpointDataNext()
	_ = manager.CheckpointDescriptorIndex()
	_ = manager.CheckpointDescriptorLength()
	_ = manager.CheckpointDataIndex()
	_ = manager.CheckpointDataLength()
}
