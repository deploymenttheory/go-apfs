package reaper

import (
	"encoding/binary"
	"testing"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

func TestReaperReader_ValidData(t *testing.T) {
	data := make([]byte, 200) // Extra space for state buffer

	// Set up ObjPhysT header (40 bytes)
	for i := 0; i < 32; i++ {
		data[i] = 0xAB
	}
	binary.LittleEndian.PutUint64(data[32:40], 1000) // OOid
	binary.LittleEndian.PutUint64(data[40:48], 100)  // OXid
	binary.LittleEndian.PutUint32(data[48:52], 0x20) // OType
	binary.LittleEndian.PutUint32(data[52:56], 0)    // OSubtype

	// Set up reaper fields
	binary.LittleEndian.PutUint64(data[56:64], 999)               // NrNextReapId
	binary.LittleEndian.PutUint64(data[64:72], 888)               // NrCompletedId
	binary.LittleEndian.PutUint64(data[72:80], 777)               // NrHead
	binary.LittleEndian.PutUint64(data[80:88], 666)               // NrTail
	binary.LittleEndian.PutUint32(data[88:92], types.NrBhmFlag)   // NrFlags
	binary.LittleEndian.PutUint32(data[92:96], 5)                 // NrRlcount
	binary.LittleEndian.PutUint32(data[96:100], 1)                // NrType
	binary.LittleEndian.PutUint32(data[100:104], 4096)            // NrSize
	binary.LittleEndian.PutUint64(data[104:112], 555)             // NrFsOid
	binary.LittleEndian.PutUint64(data[112:120], 444)             // NrOid
	binary.LittleEndian.PutUint64(data[120:128], 333)             // NrXid
	binary.LittleEndian.PutUint32(data[128:132], types.NrleValid) // NrNrleFlags
	binary.LittleEndian.PutUint32(data[132:136], 10)              // NrStateBufferSize

	// Add state buffer
	copy(data[136:146], []byte("test_state"))

	reader, err := NewReaperReader(data, binary.LittleEndian)
	if err != nil {
		t.Fatalf("NewReaperReader failed: %v", err)
	}

	if reader.NextReapID() != 999 {
		t.Errorf("NextReapID() = %d, want 999", reader.NextReapID())
	}

	if reader.CompletedReapID() != 888 {
		t.Errorf("CompletedReapID() = %d, want 888", reader.CompletedReapID())
	}

	if reader.ReaperListCount() != 5 {
		t.Errorf("ReaperListCount() = %d, want 5", reader.ReaperListCount())
	}
}

func TestReaperListEntryReader_Flags(t *testing.T) {
	data := make([]byte, 40)

	// Set up entry with VALID and CALL flags
	binary.LittleEndian.PutUint32(data[0:4], 0) // NrleNext
	flags := types.NrleValid | types.NrleCall
	binary.LittleEndian.PutUint32(data[4:8], flags) // NrleFlags
	binary.LittleEndian.PutUint32(data[8:12], 1)    // NrleType
	binary.LittleEndian.PutUint32(data[12:16], 100) // NrleSize
	binary.LittleEndian.PutUint64(data[16:24], 200) // NrleFsOid
	binary.LittleEndian.PutUint64(data[24:32], 300) // NrleOid
	binary.LittleEndian.PutUint64(data[32:40], 400) // NrleXid

	reader, err := NewReaperListEntryReader(data, binary.LittleEndian)
	if err != nil {
		t.Fatalf("NewReaperListEntryReader failed: %v", err)
	}

	if !reader.IsValid() {
		t.Error("IsValid() = false, want true")
	}

	if !reader.IsReadyToCall() {
		t.Error("IsReadyToCall() = false, want true")
	}

	if reader.IsReapIDRecord() {
		t.Error("IsReapIDRecord() = true, want false")
	}

	if reader.IsCompletionEntry() {
		t.Error("IsCompletionEntry() = true, want false")
	}

	if reader.IsCleanupEntry() {
		t.Error("IsCleanupEntry() = true, want false")
	}
}

func TestApfsReapStateReader_ValidData(t *testing.T) {
	data := make([]byte, 20)

	binary.LittleEndian.PutUint64(data[0:8], 12345)                         // LastPbn
	binary.LittleEndian.PutUint64(data[8:16], 67890)                        // CurSnapXid
	binary.LittleEndian.PutUint32(data[16:20], types.ApfsReapPhaseActiveFs) // Phase

	reader, err := NewApfsReapStateReader(data, binary.LittleEndian)
	if err != nil {
		t.Fatalf("NewApfsReapStateReader failed: %v", err)
	}

	if reader.LastProcessedBlockNumber() != 12345 {
		t.Errorf("LastProcessedBlockNumber() = %d, want 12345", reader.LastProcessedBlockNumber())
	}

	desc := reader.PhaseDescription()
	if desc != "Active Filesystem" {
		t.Errorf("PhaseDescription() = %s, want 'Active Filesystem'", desc)
	}
}

func TestOmapReapStateReader_ValidData(t *testing.T) {
	data := make([]byte, 28)

	binary.LittleEndian.PutUint32(data[0:4], types.ApfsReapPhaseSnapshots) // OmrPhase
	// OmapKeyT: OkOid(8) + OkXid(8)
	binary.LittleEndian.PutUint64(data[4:12], 1111)  // OkOid
	binary.LittleEndian.PutUint64(data[12:20], 2222) // OkXid
	// Padding/reserved at data[20:28]

	reader, err := NewOmapReapStateReader(data, binary.LittleEndian)
	if err != nil {
		t.Fatalf("NewOmapReapStateReader failed: %v", err)
	}

	if reader.ReapingPhase() != types.ApfsReapPhaseSnapshots {
		t.Errorf("ReapingPhase() = %d, want %d", reader.ReapingPhase(), types.ApfsReapPhaseSnapshots)
	}
}

func TestOmapCleanupStateReader_ValidData(t *testing.T) {
	data := make([]byte, 56)

	binary.LittleEndian.PutUint32(data[0:4], 1)       // OmcCleaning
	binary.LittleEndian.PutUint32(data[4:8], 99)      // OmcOmsflags
	binary.LittleEndian.PutUint64(data[8:16], 11111)  // OmcSxidprev
	binary.LittleEndian.PutUint64(data[16:24], 22222) // OmcSxidstart
	binary.LittleEndian.PutUint64(data[24:32], 33333) // OmcSxidend
	binary.LittleEndian.PutUint64(data[32:40], 44444) // OmcSxidnext
	// OmapKeyT (16 bytes)
	binary.LittleEndian.PutUint64(data[40:48], 5555) // OkOid
	binary.LittleEndian.PutUint64(data[48:56], 6666) // OkXid

	reader, err := NewOmapCleanupStateReader(data, binary.LittleEndian)
	if err != nil {
		t.Fatalf("NewOmapCleanupStateReader failed: %v", err)
	}

	if !reader.IsCleaning() {
		t.Error("IsCleaning() = false, want true")
	}

	if reader.SnapshotFlags() != 99 {
		t.Errorf("SnapshotFlags() = %d, want 99", reader.SnapshotFlags())
	}

	if reader.StartSnapshotXID() != 22222 {
		t.Errorf("StartSnapshotXID() = %d, want 22222", reader.StartSnapshotXID())
	}
}

func TestReaperListReader_ValidData(t *testing.T) {
	data := make([]byte, 200) // Space for header + entries

	// Set up ObjPhysT (40 bytes)
	for i := 0; i < 32; i++ {
		data[i] = 0xCD
	}
	binary.LittleEndian.PutUint64(data[32:40], 5000)

	// Set up list header
	binary.LittleEndian.PutUint64(data[56:64], 6000) // NrlNext
	binary.LittleEndian.PutUint32(data[64:68], 0)    // NrlFlags
	binary.LittleEndian.PutUint32(data[68:72], 10)   // NrlMax
	binary.LittleEndian.PutUint32(data[72:76], 2)    // NrlCount (2 entries)
	binary.LittleEndian.PutUint32(data[76:80], 0)    // NrlFirst
	binary.LittleEndian.PutUint32(data[80:84], 1)    // NrlLast
	binary.LittleEndian.PutUint32(data[84:88], 2)    // NrlFree

	// Add two entries at offset 88 and 128
	for i := 0; i < 2; i++ {
		offset := 88 + (i * 40)
		binary.LittleEndian.PutUint32(data[offset:offset+4], uint32(i))         // NrleNext
		binary.LittleEndian.PutUint32(data[offset+4:offset+8], types.NrleValid) // NrleFlags
	}

	reader, err := NewReaperListReader(data, binary.LittleEndian)
	if err != nil {
		t.Fatalf("NewReaperListReader failed: %v", err)
	}

	if reader.MaxEntries() != 10 {
		t.Errorf("MaxEntries() = %d, want 10", reader.MaxEntries())
	}

	if reader.CurrentEntryCount() != 2 {
		t.Errorf("CurrentEntryCount() = %d, want 2", reader.CurrentEntryCount())
	}

	if len(reader.Entries()) != 2 {
		t.Errorf("Entries() count = %d, want 2", len(reader.Entries()))
	}
}

func TestReaperReader_ErrorCases(t *testing.T) {
	tests := []struct {
		name      string
		dataSize  int
		shouldErr bool
	}{
		{"Valid", 200, false},
		{"Too small", 100, true},
		{"Minimum size", 136, false},
		{"One byte short", 135, true},
		{"Empty", 0, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			data := make([]byte, tc.dataSize)

			if len(data) >= 136 {
				// Fill in minimal valid data
				for i := 0; i < 32; i++ {
					data[i] = 0xFF
				}
				binary.LittleEndian.PutUint32(data[88:92], types.NrBhmFlag)
			}

			_, err := NewReaperReader(data, binary.LittleEndian)
			if tc.shouldErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tc.shouldErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestReaperReader_BigEndian(t *testing.T) {
	data := make([]byte, 150)

	for i := 0; i < 32; i++ {
		data[i] = 0xEE
	}
	binary.BigEndian.PutUint64(data[32:40], 1000)
	binary.BigEndian.PutUint64(data[40:48], 100)
	binary.BigEndian.PutUint32(data[48:52], 0x20)
	binary.BigEndian.PutUint32(data[52:56], 0)

	binary.BigEndian.PutUint64(data[56:64], 555)
	binary.BigEndian.PutUint64(data[64:72], 444)
	binary.BigEndian.PutUint64(data[72:80], 333)
	binary.BigEndian.PutUint64(data[80:88], 222)
	binary.BigEndian.PutUint32(data[88:92], types.NrBhmFlag)
	binary.BigEndian.PutUint32(data[92:96], 3)
	binary.BigEndian.PutUint32(data[96:100], 2)
	binary.BigEndian.PutUint32(data[100:104], 8192)
	binary.BigEndian.PutUint64(data[104:112], 111)
	binary.BigEndian.PutUint64(data[112:120], 222)
	binary.BigEndian.PutUint64(data[120:128], 333)
	binary.BigEndian.PutUint32(data[128:132], 0)
	binary.BigEndian.PutUint32(data[132:136], 0)

	reader, err := NewReaperReader(data, binary.BigEndian)
	if err != nil {
		t.Fatalf("NewReaperReader failed: %v", err)
	}

	if reader.NextReapID() != 555 {
		t.Errorf("NextReapID() = %d, want 555", reader.NextReapID())
	}

	if reader.ReaperListCount() != 3 {
		t.Errorf("ReaperListCount() = %d, want 3", reader.ReaperListCount())
	}
}
