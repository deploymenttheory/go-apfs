package spacemanager

import (
	"encoding/binary"
	"fmt"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

// SpacemanDeviceReader provides high-level access to device information
// Wraps spaceman_device_t to provide convenient methods for device statistics
type SpacemanDeviceReader struct {
	device *types.SpacemanDeviceT
	data   []byte
	endian binary.ByteOrder
}

// NewSpacemanDeviceReader creates a new spaceman device reader
// Device structure is 56 bytes: all fields from SpacemanDeviceT including SmCabOid
func NewSpacemanDeviceReader(data []byte, endian binary.ByteOrder) (*SpacemanDeviceReader, error) {
	if len(data) < 56 {
		return nil, fmt.Errorf("data too small for spaceman device: %d bytes, need at least 56", len(data))
	}

	device, err := parseSpacemanDevice(data, endian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse spaceman device: %w", err)
	}

	return &SpacemanDeviceReader{
		device: device,
		data:   data,
		endian: endian,
	}, nil
}

// parseSpacemanDevice parses raw bytes into SpacemanDeviceT
func parseSpacemanDevice(data []byte, endian binary.ByteOrder) (*types.SpacemanDeviceT, error) {
	if len(data) < 56 {
		return nil, fmt.Errorf("insufficient data for spaceman device")
	}

	sd := &types.SpacemanDeviceT{}
	offset := 0

	// Parse block count (uint64)
	sd.SmBlockCount = endian.Uint64(data[offset : offset+8])
	offset += 8

	// Parse chunk count (uint64)
	sd.SmChunkCount = endian.Uint64(data[offset : offset+8])
	offset += 8

	// Parse CIB count (uint32)
	sd.SmCibCount = endian.Uint32(data[offset : offset+4])
	offset += 4

	// Parse CAB count (uint32)
	sd.SmCabCount = endian.Uint32(data[offset : offset+4])
	offset += 4

	// Parse free count (uint64)
	sd.SmFreeCount = endian.Uint64(data[offset : offset+8])
	offset += 8

	// Parse address offset (uint32)
	sd.SmAddrOffset = endian.Uint32(data[offset : offset+4])
	offset += 4

	// Parse reserved (uint32)
	sd.SmReserved = endian.Uint32(data[offset : offset+4])
	offset += 4

	// Parse reserved2 (uint64)
	sd.SmReserved2 = endian.Uint64(data[offset : offset+8])
	offset += 8

	// Parse CAB OID (uint64)
	sd.SmCabOid = types.OidT(endian.Uint64(data[offset : offset+8]))
	offset += 8

	return sd, nil
}

// GetDevice returns the device structure
func (sdr *SpacemanDeviceReader) GetDevice() *types.SpacemanDeviceT {
	return sdr.device
}

// BlockCount returns the total number of blocks on this device
func (sdr *SpacemanDeviceReader) BlockCount() uint64 {
	return sdr.device.SmBlockCount
}

// ChunkCount returns the number of chunks on this device
func (sdr *SpacemanDeviceReader) ChunkCount() uint64 {
	return sdr.device.SmChunkCount
}

// CIBCount returns the number of chunk-info blocks
func (sdr *SpacemanDeviceReader) CIBCount() uint32 {
	return sdr.device.SmCibCount
}

// CABCount returns the number of chunk-info address blocks
func (sdr *SpacemanDeviceReader) CABCount() uint32 {
	return sdr.device.SmCabCount
}

// FreeCount returns the number of free blocks on this device
func (sdr *SpacemanDeviceReader) FreeCount() uint64 {
	return sdr.device.SmFreeCount
}

// UsedCount returns the number of used blocks (total - free)
func (sdr *SpacemanDeviceReader) UsedCount() uint64 {
	if sdr.device.SmBlockCount >= sdr.device.SmFreeCount {
		return sdr.device.SmBlockCount - sdr.device.SmFreeCount
	}
	return 0
}

// AddressOffset returns the address offset for this device
// Used in Fusion Drive configurations for multi-device addressing
func (sdr *SpacemanDeviceReader) AddressOffset() uint32 {
	return sdr.device.SmAddrOffset
}

// CABOid returns the Object ID of the root chunk-info address block
func (sdr *SpacemanDeviceReader) CABOid() types.OidT {
	return sdr.device.SmCabOid
}

// Utilization returns the percentage of blocks that are in use
// Returns a value between 0.0 (empty) and 100.0 (full)
func (sdr *SpacemanDeviceReader) Utilization() float64 {
	if sdr.device.SmBlockCount == 0 {
		return 0.0
	}
	used := sdr.UsedCount()
	return (float64(used) / float64(sdr.device.SmBlockCount)) * 100.0
}

// FreePercentage returns the percentage of blocks that are free
func (sdr *SpacemanDeviceReader) FreePercentage() float64 {
	if sdr.device.SmBlockCount == 0 {
		return 0.0
	}
	return (float64(sdr.device.SmFreeCount) / float64(sdr.device.SmBlockCount)) * 100.0
}

// IsFull returns true if the device has no free blocks
func (sdr *SpacemanDeviceReader) IsFull() bool {
	return sdr.device.SmFreeCount == 0
}

// IsEmpty returns true if the device has no used blocks
func (sdr *SpacemanDeviceReader) IsEmpty() bool {
	return sdr.device.SmFreeCount == sdr.device.SmBlockCount
}

// HasChunks returns true if this device has any chunks
func (sdr *SpacemanDeviceReader) HasChunks() bool {
	return sdr.device.SmChunkCount > 0
}

// HasCABs returns true if this device has chunk-info address blocks
func (sdr *SpacemanDeviceReader) HasCABs() bool {
	return sdr.device.SmCabCount > 0
}

// HasCIBs returns true if this device has chunk-info blocks
func (sdr *SpacemanDeviceReader) HasCIBs() bool {
	return sdr.device.SmCibCount > 0
}

// IsActive returns true if this device is active (has blocks and CAB root)
func (sdr *SpacemanDeviceReader) IsActive() bool {
	return sdr.device.SmBlockCount > 0 && sdr.device.SmCabOid != 0
}

// Summary returns a human-readable summary of device status
func (sdr *SpacemanDeviceReader) Summary() string {
	return fmt.Sprintf("Device{Blocks: %d, Free: %d (%.1f%%), Chunks: %d, CABs: %d, CIBs: %d, CABOid: %d}",
		sdr.device.SmBlockCount,
		sdr.device.SmFreeCount,
		sdr.FreePercentage(),
		sdr.device.SmChunkCount,
		sdr.device.SmCabCount,
		sdr.device.SmCibCount,
		sdr.device.SmCabOid)
}
