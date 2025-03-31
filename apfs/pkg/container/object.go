// File: pkg/container/object.go
// Package container implements low-level APFS object storage and manipulation
package container

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"sync"
	"unsafe"

	"github.com/deploymenttheory/go-apfs/apfs/pkg/checksum"
	"github.com/deploymenttheory/go-apfs/apfs/pkg/types"
)

// ObjectStorage handles low-level object storage operations
type ObjectStorage struct {
	device     types.BlockDevice
	mu         sync.RWMutex
	blockCache sync.Map
}

// NewObjectStorage creates a new ObjectStorage instance
func NewObjectStorage(device types.BlockDevice) *ObjectStorage {
	return &ObjectStorage{
		device: device,
	}
}

// Store writes an object to the specified physical address
func (os *ObjectStorage) Store(addr types.PAddr, data []byte) error {
	os.mu.Lock()
	defer os.mu.Unlock()

	// Verify data fits in a block
	if len(data) > int(os.device.GetBlockSize()) {
		return fmt.Errorf("object data exceeds block size")
	}

	// Prepare data with checksum
	blockData := make([]byte, os.device.GetBlockSize())
	copy(blockData, data)

	// Calculate and set checksum (zero out first 8 bytes for calculation)
	checksumBytes := make([]byte, len(blockData))
	copy(checksumBytes, blockData)
	for i := 0; i < 8; i++ {
		checksumBytes[i] = 0
	}
	checksum := checksum.Fletcher64WithZeroedChecksum(checksumBytes, 0)
	binary.LittleEndian.PutUint64(blockData[0:8], checksum)

	// Write to device
	if err := os.device.WriteBlock(addr, blockData); err != nil {
		return fmt.Errorf("failed to write object: %w", err)
	}

	// Cache the object
	os.blockCache.Store(addr, blockData)

	return nil
}

// Retrieve reads an object from the specified physical address
func (os *ObjectStorage) Retrieve(addr types.PAddr) ([]byte, error) {
	// Check cache first
	if cachedData, ok := os.blockCache.Load(addr); ok {
		return cachedData.([]byte), nil
	}

	// Read from device
	data, err := os.device.ReadBlock(addr)
	if err != nil {
		return nil, fmt.Errorf("failed to read object: %w", err)
	}

	// Verify checksum
	expectedChecksum := binary.LittleEndian.Uint64(data[0:8])
	computedChecksum := checksum.Fletcher64WithZeroedChecksum(data, 0)
	if expectedChecksum != computedChecksum {
		return nil, fmt.Errorf("checksum mismatch for object at address %d", addr)
	}

	// Cache the object
	os.blockCache.Store(addr, data)

	return data, nil
}

// Delete removes an object from the specified physical address
func (os *ObjectStorage) Delete(addr types.PAddr, spaceman types.SpaceManager) error {
	os.mu.Lock()
	defer os.mu.Unlock()

	// Remove from cache
	os.blockCache.Delete(addr)

	// Free the block
	if err := spaceman.FreeBlock(addr); err != nil {
		return fmt.Errorf("failed to free object block: %w", err)
	}

	return nil
}

// ExtractObjectHeader extracts the object header from raw object data
func ExtractObjectHeader(data []byte) (*types.ObjectHeader, error) {
	if len(data) < 32 {
		return nil, fmt.Errorf("insufficient data for object header")
	}

	header := &types.ObjectHeader{
		Cksum:   types.Checksum(data[0:8]),
		OID:     types.OID(binary.LittleEndian.Uint64(data[8:16])),
		XID:     types.XID(binary.LittleEndian.Uint64(data[16:24])),
		Type:    binary.LittleEndian.Uint32(data[24:28]),
		Subtype: binary.LittleEndian.Uint32(data[28:32]),
	}

	return header, nil
}

// ClearCache clears the entire object cache
func (os *ObjectStorage) ClearCache() {
	os.blockCache = sync.Map{}
}

// CacheObject manually adds an object to the cache
func (os *ObjectStorage) CacheObject(addr types.PAddr, data []byte) {
	os.blockCache.Store(addr, data)
}

// SerializeObjectHeader serializes an ObjectHeader to binary data
func SerializeObjectHeader(header *types.ObjectHeader) ([]byte, error) {
	buf := new(bytes.Buffer)
	writer := types.NewBinaryWriter(buf, binary.LittleEndian)

	if err := writer.Write(header.Cksum); err != nil {
		return nil, fmt.Errorf("failed to write checksum: %w", err)
	}

	if err := writer.WriteOID(header.OID); err != nil {
		return nil, fmt.Errorf("failed to write OID: %w", err)
	}

	if err := writer.WriteXID(header.XID); err != nil {
		return nil, fmt.Errorf("failed to write XID: %w", err)
	}

	if err := writer.WriteUint32(header.Type); err != nil {
		return nil, fmt.Errorf("failed to write type: %w", err)
	}

	if err := writer.WriteUint32(header.Subtype); err != nil {
		return nil, fmt.Errorf("failed to write subtype: %w", err)
	}

	return buf.Bytes(), nil
}

// DeserializeObjectHeader deserializes an ObjectHeader from binary data
func DeserializeObjectHeader(data []byte) (*types.ObjectHeader, error) {
	if len(data) < int(unsafe.Sizeof(types.ObjectHeader{})) {
		return nil, types.ErrStructTooShort
	}

	reader := types.NewBinaryReader(bytes.NewReader(data), binary.LittleEndian)
	header := &types.ObjectHeader{}

	if err := reader.Read(&header.Cksum); err != nil {
		return nil, fmt.Errorf("failed to read checksum: %w", err)
	}

	var err error
	header.OID, err = reader.ReadOID()
	if err != nil {
		return nil, fmt.Errorf("failed to read OID: %w", err)
	}

	header.XID, err = reader.ReadXID()
	if err != nil {
		return nil, fmt.Errorf("failed to read XID: %w", err)
	}

	header.Type, err = reader.ReadUint32()
	if err != nil {
		return nil, fmt.Errorf("failed to read type: %w", err)
	}

	header.Subtype, err = reader.ReadUint32()
	if err != nil {
		return nil, fmt.Errorf("failed to read subtype: %w", err)
	}

	return header, nil
}
