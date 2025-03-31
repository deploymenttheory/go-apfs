// File: pkg/container/omap.go
package container

import (
	"encoding/binary"
	"fmt"

	"github.com/deploymenttheory/go-apfs/apfs/pkg/checksum"
	"github.com/deploymenttheory/go-apfs/apfs/pkg/types"
)

// OMapPhysSize defines the fixed size of an OMapPhys structure.
const OMapPhysSize = 56

type OMap struct {
	Phys     *types.OMapPhys
	Device   types.BlockDevice
	Spaceman types.SpaceManager
}

func (o *OMap) Resolve(oid types.OID) (types.PAddr, error) {
	return ResolveOID(o.Device, o.Phys, oid)
}

func (o *OMap) Set(oid types.OID, addr types.PAddr) error {
	return InsertOIDMapping(o.Device, o.Phys, oid, addr, o.Spaceman)
}

func (o *OMap) Delete(oid types.OID) error {
	return DeleteOIDMapping(o.Device, o.Phys, oid, o.Spaceman)
}

// ReadOMapPhys reads and parses an OMapPhys structure from a block device.
func ReadOMapPhys(device types.BlockDevice, addr types.PAddr) (*types.OMapPhys, error) {
	blockSize := device.GetBlockSize()
	if OMapPhysSize > int(blockSize) {
		return nil, fmt.Errorf("OMapPhys size (%d) exceeds device block size (%d)", OMapPhysSize, blockSize)
	}

	data, err := device.ReadBlock(addr)
	if err != nil {
		return nil, fmt.Errorf("failed to read block at addr %d: %w", addr, err)
	}

	if len(data) < OMapPhysSize {
		return nil, fmt.Errorf("data too short: expected %d bytes, got %d", OMapPhysSize, len(data))
	}

	var omap types.OMapPhys
	reader := binary.LittleEndian

	omap.Header.Cksum = types.Checksum(data[0:8])
	omap.Header.OID = types.OID(reader.Uint64(data[8:16]))
	omap.Header.XID = types.XID(reader.Uint64(data[16:24]))
	omap.Header.Type = reader.Uint32(data[24:28])
	omap.Header.Subtype = reader.Uint32(data[28:32])

	omap.Flags = reader.Uint32(data[32:36])
	omap.SnapCount = reader.Uint32(data[36:40])
	omap.TreeType = reader.Uint32(data[40:44])
	omap.SnapshotTreeType = reader.Uint32(data[44:48])
	omap.TreeOID = types.OID(reader.Uint64(data[48:56]))

	return &omap, nil
}

func WriteOMapPhys(device types.BlockDevice, addr types.PAddr, omap *types.OMapPhys) error {
	data := make([]byte, OMapPhysSize)
	r := binary.LittleEndian
	r.PutUint64(data[8:16], uint64(omap.Header.OID))
	r.PutUint64(data[16:24], uint64(omap.Header.XID))
	r.PutUint32(data[24:28], omap.Header.Type)
	r.PutUint32(data[28:32], omap.Header.Subtype)
	r.PutUint32(data[32:36], omap.Flags)
	r.PutUint32(data[36:40], omap.SnapCount)
	r.PutUint32(data[40:44], omap.TreeType)
	r.PutUint32(data[44:48], omap.SnapshotTreeType)
	r.PutUint64(data[48:56], uint64(omap.TreeOID))

	// Calculate and write checksum
	cks := checksum.Fletcher64WithZeroedChecksum(data, 0)
	binary.LittleEndian.PutUint64(data[0:8], cks)

	return device.WriteBlock(addr, data)
}

func ResolveOID(device types.BlockDevice, omap *types.OMapPhys, oid types.OID) (types.PAddr, error) {
	rootAddr := types.PAddr(omap.TreeOID)
	key := make([]byte, 8)
	binary.LittleEndian.PutUint64(key, uint64(oid))

	value, found, err := LookupBTree(device, rootAddr, key, bytesCompareUint64)
	if err != nil || !found {
		return 0, fmt.Errorf("object not found: %v", err)
	}
	if len(value) < 8 {
		return 0, fmt.Errorf("invalid value length: %d", len(value))
	}
	return types.PAddr(binary.LittleEndian.Uint64(value[:8])), nil
}

func InsertOIDMapping(device types.BlockDevice, omap *types.OMapPhys, oid types.OID, addr types.PAddr, spaceman types.SpaceManager) error {
	rootAddr := types.PAddr(omap.TreeOID)
	key := make([]byte, 8)
	binary.LittleEndian.PutUint64(key, uint64(oid))
	val := make([]byte, 8)
	binary.LittleEndian.PutUint64(val, uint64(addr))

	newRoot, err := InsertBTree(device, rootAddr, key, val, bytesCompareUint64, spaceman)
	if err != nil {
		return err
	}
	omap.TreeOID = types.OID(newRoot)
	return nil
}

func DeleteOIDMapping(device types.BlockDevice, omap *types.OMapPhys, oid types.OID, spaceman types.SpaceManager) error {
	rootAddr := types.PAddr(omap.TreeOID)
	key := make([]byte, 8)
	binary.LittleEndian.PutUint64(key, uint64(oid))

	newRoot, err := DeleteBTree(device, rootAddr, key, bytesCompareUint64, spaceman)
	if err != nil {
		return err
	}
	omap.TreeOID = types.OID(newRoot)
	return nil
}

func DumpOMap(device types.BlockDevice, omap *types.OMapPhys) error {
	rootAddr := types.PAddr(omap.TreeOID)
	entries, err := RangeBTree(device, rootAddr, nil, nil, bytesCompareUint64)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if len(entry[0]) >= 8 && len(entry[1]) >= 8 {
			oid := binary.LittleEndian.Uint64(entry[0])
			paddr := binary.LittleEndian.Uint64(entry[1])
			fmt.Printf("OID: 0x%x -> PAddr: 0x%x\n", oid, paddr)
		}
	}
	return nil
}

func bytesCompareUint64(a, b []byte) int {
	return int(binary.LittleEndian.Uint64(a) - binary.LittleEndian.Uint64(b))
}

// Validate performs basic validation on the OMapPhys structure.
func ValidateOMapPhys(omap *types.OMapPhys) error {
	if omap.TreeOID == types.OIDInvalid {
		return fmt.Errorf("invalid TreeOID: cannot be OIDInvalid (0)")
	}
	return nil
}
