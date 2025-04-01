package efijumpstart

import (
	"testing"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

func TestEFIJumpstartReader(t *testing.T) {
	t.Run("fields are returned correctly", func(t *testing.T) {
		extents := []types.Prange{
			{PrStartPaddr: 0x1000, PrBlockCount: 4},
			{PrStartPaddr: 0x2000, PrBlockCount: 2},
		}

		jump := types.NxEfiJumpstartT{
			NejMagic:      types.NxEfiJumpstartMagic,
			NejVersion:    types.NxEfiJumpstartVersion,
			NejEfiFileLen: 8192,
			NejNumExtents: uint32(len(extents)),
			NejRecExtents: extents,
		}

		reader := NewEFIJumpstartReader(jump)

		if reader.Magic() != types.NxEfiJumpstartMagic {
			t.Errorf("Magic() = %x; want %x", reader.Magic(), types.NxEfiJumpstartMagic)
		}
		if reader.Version() != types.NxEfiJumpstartVersion {
			t.Errorf("Version() = %d; want %d", reader.Version(), types.NxEfiJumpstartVersion)
		}
		if reader.EFIFileLength() != 8192 {
			t.Errorf("EFIFileLength() = %d; want 8192", reader.EFIFileLength())
		}
		if reader.ExtentCount() != 2 {
			t.Errorf("ExtentCount() = %d; want 2", reader.ExtentCount())
		}

		gotExtents := reader.Extents()
		if len(gotExtents) != 2 || gotExtents[0] != extents[0] || gotExtents[1] != extents[1] {
			t.Errorf("Extents() = %+v; want %+v", gotExtents, extents)
		}
	})

	t.Run("IsValid returns true with correct magic and version", func(t *testing.T) {
		jump := types.NxEfiJumpstartT{
			NejMagic:   types.NxEfiJumpstartMagic,
			NejVersion: types.NxEfiJumpstartVersion,
		}
		reader := NewEFIJumpstartReader(jump)
		if !reader.IsValid() {
			t.Error("IsValid() = false; want true")
		}
	})

	t.Run("IsValid returns false with incorrect magic", func(t *testing.T) {
		jump := types.NxEfiJumpstartT{
			NejMagic:   0xDEADBEEF,
			NejVersion: types.NxEfiJumpstartVersion,
		}
		reader := NewEFIJumpstartReader(jump)
		if reader.IsValid() {
			t.Error("IsValid() = true; want false")
		}
	})

	t.Run("IsValid returns false with incorrect version", func(t *testing.T) {
		jump := types.NxEfiJumpstartT{
			NejMagic:   types.NxEfiJumpstartMagic,
			NejVersion: 0x12345678,
		}
		reader := NewEFIJumpstartReader(jump)
		if reader.IsValid() {
			t.Error("IsValid() = true; want false")
		}
	})
}
