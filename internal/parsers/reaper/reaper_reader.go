package reaper

import (
	"encoding/binary"
	"fmt"

	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// reaperReader implements ReaperReader and ReaperObjectInfo interfaces
type reaperReader struct {
	reaper *types.NxReaperPhysT
	endian binary.ByteOrder
}

// NewReaperReader creates a new reader for reaper structures
func NewReaperReader(data []byte, endian binary.ByteOrder) (interfaces.ReaperReader, error) {
	// ObjPhysT(40) + NrNextReapId(8) + NrCompletedId(8) + NrHead(8) + NrTail(8) + NrFlags(4) + NrRlcount(4) +
	// NrType(4) + NrSize(4) + NrFsOid(8) + NrOid(8) + NrXid(8) + NrNrleFlags(4) + NrStateBufferSize(4) = 136 bytes minimum
	if len(data) < 136 {
		return nil, fmt.Errorf("data too small for reaper: %d bytes", len(data))
	}

	reaper, err := parseReaper(data, endian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse reaper: %w", err)
	}

	return &reaperReader{
		reaper: reaper,
		endian: endian,
	}, nil
}

// parseReaper parses raw bytes into a NxReaperPhysT structure
func parseReaper(data []byte, endian binary.ByteOrder) (*types.NxReaperPhysT, error) {
	if len(data) < 136 {
		return nil, fmt.Errorf("insufficient data for reaper: need 136 bytes, got %d", len(data))
	}

	reaper := &types.NxReaperPhysT{}

	// Parse ObjPhysT header (40 bytes)
	copy(reaper.NrO.OChecksum[:], data[0:32])
	reaper.NrO.OOid = types.OidT(endian.Uint64(data[32:40]))
	reaper.NrO.OXid = types.XidT(endian.Uint64(data[40:48]))
	reaper.NrO.OType = endian.Uint32(data[48:52])
	reaper.NrO.OSubtype = endian.Uint32(data[52:56])

	// Parse reaper fields
	reaper.NrNextReapId = endian.Uint64(data[56:64])
	reaper.NrCompletedId = endian.Uint64(data[64:72])
	reaper.NrHead = types.OidT(endian.Uint64(data[72:80]))
	reaper.NrTail = types.OidT(endian.Uint64(data[80:88]))
	reaper.NrFlags = endian.Uint32(data[88:92])
	reaper.NrRlcount = endian.Uint32(data[92:96])
	reaper.NrType = endian.Uint32(data[96:100])
	reaper.NrSize = endian.Uint32(data[100:104])
	reaper.NrFsOid = types.OidT(endian.Uint64(data[104:112]))
	reaper.NrOid = types.OidT(endian.Uint64(data[112:120]))
	reaper.NrXid = types.XidT(endian.Uint64(data[120:128]))
	reaper.NrNrleFlags = endian.Uint32(data[128:132])
	reaper.NrStateBufferSize = endian.Uint32(data[132:136])

	// Parse state buffer (variable-length)
	stateBufferEnd := 136 + reaper.NrStateBufferSize
	if stateBufferEnd > uint32(len(data)) {
		stateBufferEnd = uint32(len(data))
	}

	if stateBufferEnd > 136 {
		reaper.NrStateBuffer = make([]byte, stateBufferEnd-136)
		copy(reaper.NrStateBuffer, data[136:stateBufferEnd])
	} else {
		reaper.NrStateBuffer = make([]byte, 0)
	}

	return reaper, nil
}

// NextReapID returns the next reap identifier to be assigned
func (rr *reaperReader) NextReapID() uint64 {
	return rr.reaper.NrNextReapId
}

// CompletedReapID returns the identifier of the last completed reap
func (rr *reaperReader) CompletedReapID() uint64 {
	return rr.reaper.NrCompletedId
}

// HeadOID returns the object identifier of the head of the reaper list
func (rr *reaperReader) HeadOID() types.OidT {
	return rr.reaper.NrHead
}

// TailOID returns the object identifier of the tail of the reaper list
func (rr *reaperReader) TailOID() types.OidT {
	return rr.reaper.NrTail
}

// Flags returns the reaper flags
func (rr *reaperReader) Flags() uint32 {
	return rr.reaper.NrFlags
}

// ReaperListCount returns the count of reaper lists
func (rr *reaperReader) ReaperListCount() uint32 {
	return rr.reaper.NrRlcount
}

// Type returns the type of the object being reaped (ReaperObjectInfo)
func (rr *reaperReader) Type() uint32 {
	return rr.reaper.NrType
}

// Size returns the size of the object being reaped (ReaperObjectInfo)
func (rr *reaperReader) Size() uint32 {
	return rr.reaper.NrSize
}

// FileSystemOID returns the filesystem object identifier (ReaperObjectInfo)
func (rr *reaperReader) FileSystemOID() types.OidT {
	return rr.reaper.NrFsOid
}

// ObjectID returns the object identifier (ReaperObjectInfo)
func (rr *reaperReader) ObjectID() types.OidT {
	return rr.reaper.NrOid
}

// TransactionID returns the transaction identifier (ReaperObjectInfo)
func (rr *reaperReader) TransactionID() types.XidT {
	return rr.reaper.NrXid
}

// IsBhmFlagSet checks if the BHM flag is set
func (rr *reaperReader) IsBhmFlagSet() bool {
	return (rr.reaper.NrFlags & types.NrBhmFlag) != 0
}

// IsContinueSet checks if the CONTINUE flag is set
func (rr *reaperReader) IsContinueSet() bool {
	return (rr.reaper.NrFlags & types.NrContinue) != 0
}

// StateBuffer returns the state buffer for the reaper
func (rr *reaperReader) StateBuffer() []byte {
	return rr.reaper.NrStateBuffer
}
