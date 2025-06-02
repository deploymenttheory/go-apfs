package encryptionrolling

import (
	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// flagManager implements the EncryptionRollingFlagManager interface
type flagManager struct {
	flags uint64
}

// phaseManager implements the EncryptionRollingPhaseManager interface
type phaseManager struct {
	flags uint64
}

// blockSizeResolver implements the BlockSizeResolver interface
type blockSizeResolver struct{}

// Ensure interface compliance
var _ interfaces.EncryptionRollingFlagManager = (*flagManager)(nil)
var _ interfaces.EncryptionRollingPhaseManager = (*phaseManager)(nil)
var _ interfaces.BlockSizeResolver = (*blockSizeResolver)(nil)

// NewFlagManager creates a new EncryptionRollingFlagManager
func NewFlagManager(flags uint64) interfaces.EncryptionRollingFlagManager {
	return &flagManager{
		flags: flags,
	}
}

// NewPhaseManager creates a new EncryptionRollingPhaseManager
func NewPhaseManager(flags uint64) interfaces.EncryptionRollingPhaseManager {
	return &phaseManager{
		flags: flags,
	}
}

// NewBlockSizeResolver creates a new BlockSizeResolver
func NewBlockSizeResolver() interfaces.BlockSizeResolver {
	return &blockSizeResolver{}
}

// Implementation of EncryptionRollingFlagManager interface

func (f *flagManager) IsEncrypting() bool {
	return (f.flags & types.ErsbFlagEncrypting) != 0
}

func (f *flagManager) IsDecrypting() bool {
	return (f.flags & types.ErsbFlagDecrypting) != 0
}

func (f *flagManager) IsKeyRolling() bool {
	return (f.flags & types.ErsbFlagKeyrolling) != 0
}

func (f *flagManager) IsPaused() bool {
	return (f.flags & types.ErsbFlagPaused) != 0
}

func (f *flagManager) HasFailed() bool {
	return (f.flags & types.ErsbFlagFailed) != 0
}

func (f *flagManager) IsCIDTweak() bool {
	return (f.flags & types.ErsbFlagCidIsTweak) != 0
}

func (f *flagManager) GetBlockSize() uint64 {
	blockSizeCode := (f.flags & types.ErsbFlagCmBlockSizeMask) >> types.ErsbFlagCmBlockSizeShift
	resolver := NewBlockSizeResolver()
	return uint64(resolver.GetBlockSizeValue(uint32(blockSizeCode)))
}

func (f *flagManager) GetPhase() types.ErPhaseT {
	return types.ErPhaseT((f.flags & types.ErsbFlagErPhaseMask) >> types.ErsbFlagErPhaseShift)
}

func (f *flagManager) IsFromOneKey() bool {
	return (f.flags & types.ErsbFlagFromOnekey) != 0
}

// Implementation of EncryptionRollingPhaseManager interface

func (p *phaseManager) GetCurrentPhase() types.ErPhaseT {
	return types.ErPhaseT((p.flags & types.ErsbFlagErPhaseMask) >> types.ErsbFlagErPhaseShift)
}

func (p *phaseManager) GetPhaseDescription() string {
	phase := p.GetCurrentPhase()
	switch phase {
	case types.ErPhaseOmapRoll:
		return "Object Map Rolling"
	case types.ErPhaseDataRoll:
		return "Data Rolling"
	case types.ErPhaseSnapRoll:
		return "Snapshot Rolling"
	default:
		return "Unknown Phase"
	}
}

func (p *phaseManager) IsOmapRollPhase() bool {
	return p.GetCurrentPhase() == types.ErPhaseOmapRoll
}

func (p *phaseManager) IsDataRollPhase() bool {
	return p.GetCurrentPhase() == types.ErPhaseDataRoll
}

func (p *phaseManager) IsSnapshotRollPhase() bool {
	return p.GetCurrentPhase() == types.ErPhaseSnapRoll
}

// Implementation of BlockSizeResolver interface

func (b *blockSizeResolver) GetBlockSizeValue(blockSizeConstant uint32) uint32 {
	switch blockSizeConstant {
	case types.Er512bBlocksize:
		return 512
	case types.Er2kibBlocksize:
		return 2048 // 2 KiB
	case types.Er4kibBlocksize:
		return 4096 // 4 KiB
	case types.Er8kibBlocksize:
		return 8192 // 8 KiB
	case types.Er16kibBlocksize:
		return 16384 // 16 KiB
	case types.Er32kibBlocksize:
		return 32768 // 32 KiB
	case types.Er64kibBlocksize:
		return 65536 // 64 KiB
	default:
		return 0 // Invalid block size
	}
}

func (b *blockSizeResolver) GetBlockSizeConstant(sizeInBytes uint32) uint32 {
	switch sizeInBytes {
	case 512:
		return types.Er512bBlocksize
	case 2048:
		return types.Er2kibBlocksize
	case 4096:
		return types.Er4kibBlocksize
	case 8192:
		return types.Er8kibBlocksize
	case 16384:
		return types.Er16kibBlocksize
	case 32768:
		return types.Er32kibBlocksize
	case 65536:
		return types.Er64kibBlocksize
	default:
		return 0xFFFFFFFF // Invalid constant
	}
}

func (b *blockSizeResolver) GetSupportedBlockSizes() []uint32 {
	return []uint32{512, 2048, 4096, 8192, 16384, 32768, 65536}
}
