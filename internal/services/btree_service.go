package services

import (
	"encoding/binary"
	"fmt"

	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/parsers/btrees"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// BTreeService provides core B-tree navigation and record extraction
type BTreeService struct {
	container *ContainerReader
	resolver  *BTreeObjectResolver
}

// NewBTreeService creates a new B-tree service
func NewBTreeService(container *ContainerReader) *BTreeService {
	return &BTreeService{
		container: container,
		resolver:  NewBTreeObjectResolver(container),
	}
}

// FSRecord represents a filesystem record from the B-tree
type FSRecord struct {
	OID       uint64
	XID       uint64
	Type      types.JObjTypes
	KeyData   []byte
	ValueData []byte
}

// GetFSRecordsForOID gets all filesystem records for a given object ID
func (bt *BTreeService) GetFSRecordsForOID(rootTreeOID types.OidT, targetOID types.OidT, maxXID types.XidT) ([]FSRecord, error) {
	// Resolve the root tree OID to get the filesystem B-tree root
	physAddr, err := bt.resolver.ResolveVirtualObject(rootTreeOID, maxXID)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve root tree OID %d: %w", rootTreeOID, err)
	}

	// Read the root B-tree node
	rootData, err := bt.container.ReadBlock(uint64(physAddr))
	if err != nil {
		return nil, fmt.Errorf("failed to read root B-tree node: %w", err)
	}

	// Parse the root node
	rootNode, err := btrees.NewBTreeNodeReader(rootData, binary.LittleEndian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse root B-tree node: %w", err)
	}

	// Search the B-tree for records matching the target OID
	return bt.searchBTreeForFSRecords(rootNode, targetOID, maxXID)
}

// GetOMapEntry gets an object map entry for a virtual OID
func (bt *BTreeService) GetOMapEntry(omapTreeOID types.OidT, virtualOID types.OidT, maxXID types.XidT) (*OMapEntry, error) {
	// This is essentially what our BTreeObjectResolver does, but we'll expose it as a service method
	physAddr, err := bt.resolver.ResolveVirtualObject(virtualOID, maxXID)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve virtual OID %d: %w", virtualOID, err)
	}

	return &OMapEntry{
		VirtualOID:  virtualOID,
		PhysicalAddr: physAddr,
		XID:         maxXID,
	}, nil
}

// OMapEntry represents an object map entry
type OMapEntry struct {
	VirtualOID   types.OidT
	PhysicalAddr types.Paddr
	XID          types.XidT
}

// searchBTreeForFSRecords recursively searches a B-tree for filesystem records
func (bt *BTreeService) searchBTreeForFSRecords(node interfaces.BTreeNodeReader, targetOID types.OidT, maxXID types.XidT) ([]FSRecord, error) {
	var records []FSRecord

	if node.IsLeaf() {
		// Search leaf node for matching records
		leafRecords, err := bt.extractFSRecordsFromLeaf(node, targetOID, maxXID)
		if err != nil {
			return nil, err
		}
		records = append(records, leafRecords...)
	} else {
		// Search internal node - need to traverse children
		children, err := bt.findChildrenForOID(node, targetOID)
		if err != nil {
			return nil, err
		}

		// Recursively search each relevant child
		for _, childOID := range children {
			childRecords, err := bt.searchChildNodeForFSRecords(childOID, targetOID, maxXID)
			if err != nil {
				continue // Skip failed children
			}
			records = append(records, childRecords...)
		}
	}

	return records, nil
}

// extractFSRecordsFromLeaf extracts filesystem records from a leaf node
func (bt *BTreeService) extractFSRecordsFromLeaf(node interfaces.BTreeNodeReader, targetOID types.OidT, maxXID types.XidT) ([]FSRecord, error) {
	var records []FSRecord

	tableSpace := node.TableSpace()
	nodeData := node.Data()
	keyCount := node.KeyCount()

	btnDataStart := 56
	tableOffset := btnDataStart + int(tableSpace.Off)

	if tableOffset >= len(nodeData) {
		return records, nil
	}

	if node.HasFixedKVSize() {
		entrySize := 4
		for i := uint32(0); i < keyCount; i++ {
			offset := tableOffset + int(i)*entrySize
			if offset+entrySize > len(nodeData) {
				break
			}

			keyOffset := binary.LittleEndian.Uint16(nodeData[offset:offset+2])
			valueOffset := binary.LittleEndian.Uint16(nodeData[offset+2:offset+4])

			// Extract key and check if it matches our criteria
			keyStart := btnDataStart + int(keyOffset)
			if keyStart+8 <= len(nodeData) {
				objIdAndType := binary.LittleEndian.Uint64(nodeData[keyStart:keyStart+8])
				objID := objIdAndType & types.ObjIdMask
				objType := types.JObjTypes((objIdAndType & types.ObjTypeMask) >> types.ObjTypeShift)

				// Check if this record is for our target OID
				if objID == uint64(targetOID) {
					// Extract the full key and value
					keyData, valueData, err := bt.extractKeyValueData(nodeData, keyStart, btnDataStart+int(valueOffset), i, keyCount, tableOffset, entrySize)
					if err != nil {
						continue
					}

					record := FSRecord{
						OID:       objID,
						XID:       0, // Would need to extract from key if present
						Type:      objType,
						KeyData:   keyData,
						ValueData: valueData,
					}
					records = append(records, record)
				}
			}
		}
	}

	return records, nil
}

// extractKeyValueData extracts key and value data with proper size calculation
func (bt *BTreeService) extractKeyValueData(nodeData []byte, keyStart, valueStart int, entryIndex, keyCount uint32, tableOffset, entrySize int) ([]byte, []byte, error) {
	var keyData, valueData []byte

	// For keys, we need to determine the size by looking at the value offset or next key
	var keySize int
	if valueStart > keyStart {
		keySize = valueStart - keyStart
	} else {
		keySize = 8 // Minimum size for most APFS keys
	}

	if keyStart+keySize <= len(nodeData) {
		keyData = nodeData[keyStart : keyStart+keySize]
	}

	// For values, calculate size using next entry or end of data
	var valueSize int
	if entryIndex+1 < keyCount {
		// Get next entry's key offset to calculate this value's size
		nextOffset := tableOffset + int(entryIndex+1)*entrySize
		if nextOffset+2 <= len(nodeData) {
			nextKeyOffset := binary.LittleEndian.Uint16(nodeData[nextOffset:nextOffset+2])
			nextKeyStart := 56 + int(nextKeyOffset) // btn_data start + offset
			valueSize = nextKeyStart - valueStart
		}
	} else {
		// Last entry, use remaining data
		valueSize = len(nodeData) - valueStart
	}

	if valueSize > 0 && valueStart+valueSize <= len(nodeData) {
		valueData = nodeData[valueStart : valueStart+valueSize]
	}

	return keyData, valueData, nil
}

// findChildrenForOID finds child node OIDs that might contain the target OID
func (bt *BTreeService) findChildrenForOID(node interfaces.BTreeNodeReader, targetOID types.OidT) ([]types.OidT, error) {
	var children []types.OidT

	tableSpace := node.TableSpace()
	nodeData := node.Data()
	keyCount := node.KeyCount()

	btnDataStart := 56
	tableOffset := btnDataStart + int(tableSpace.Off)

	if tableOffset >= len(nodeData) {
		return children, nil
	}

	if node.HasFixedKVSize() {
		entrySize := 4
		for i := uint32(0); i < keyCount; i++ {
			offset := tableOffset + int(i)*entrySize
			if offset+entrySize > len(nodeData) {
				break
			}

			keyOffset := binary.LittleEndian.Uint16(nodeData[offset:offset+2])
			valueOffset := binary.LittleEndian.Uint16(nodeData[offset+2:offset+4])

			// Extract key to check OID range
			keyStart := btnDataStart + int(keyOffset)
			if keyStart+8 <= len(nodeData) {
				objIdAndType := binary.LittleEndian.Uint64(nodeData[keyStart:keyStart+8])
				objID := objIdAndType & types.ObjIdMask

				// If this key's OID is >= our target, this child might contain our target
				if objID >= uint64(targetOID) {
					valueStart := btnDataStart + int(valueOffset)
					if valueStart+8 <= len(nodeData) {
						childOID := types.OidT(binary.LittleEndian.Uint64(nodeData[valueStart:valueStart+8]))
						children = append(children, childOID)
					}
				}
			}
		}
	}

	return children, nil
}

// searchChildNodeForFSRecords searches a child node for filesystem records
func (bt *BTreeService) searchChildNodeForFSRecords(childOID types.OidT, targetOID types.OidT, maxXID types.XidT) ([]FSRecord, error) {
	// Resolve child OID to physical address
	physAddr, err := bt.resolver.ResolveVirtualObject(childOID, maxXID)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve child OID %d: %w", childOID, err)
	}

	// Read child node
	childData, err := bt.container.ReadBlock(uint64(physAddr))
	if err != nil {
		return nil, fmt.Errorf("failed to read child node: %w", err)
	}

	// Parse child node
	childNode, err := btrees.NewBTreeNodeReader(childData, binary.LittleEndian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse child node: %w", err)
	}

	// Recursively search the child
	return bt.searchBTreeForFSRecords(childNode, targetOID, maxXID)
}

// ParseDirectoryRecord parses a directory record from filesystem record data
func (bt *BTreeService) ParseDirectoryRecord(record FSRecord) (*DirectoryRecord, error) {
	if record.Type != types.ApfsTypeDirRec {
		return nil, fmt.Errorf("record type %v is not a directory record", record.Type)
	}

	// Parse the directory record from the value data
	// This would implement parsing of JDrecVal structure
	if len(record.ValueData) < 8 {
		return nil, fmt.Errorf("directory record value too short")
	}

	fileID := binary.LittleEndian.Uint64(record.ValueData[0:8])
	
	// Extract name from key data (simplified)
	name := "unknown" // Would parse from JDrecHashedKeyT structure
	if len(record.KeyData) > 16 {
		// Name parsing would go here
	}

	return &DirectoryRecord{
		InodeNumber: fileID,
		Name:        name,
		FileType:    0, // Would extract from flags
	}, nil
}

// ParseInodeRecord parses an inode record from filesystem record data
func (bt *BTreeService) ParseInodeRecord(record FSRecord) (*InodeRecord, error) {
	if record.Type != types.ApfsTypeInode {
		return nil, fmt.Errorf("record type %v is not an inode record", record.Type)
	}

	// Parse the inode record from the value data
	// This would implement parsing of JInodeVal structure
	if len(record.ValueData) < 98 {
		return nil, fmt.Errorf("inode record value too short")
	}

	// Basic parsing of JInodeVal fields
	parentID := binary.LittleEndian.Uint64(record.ValueData[0:8])
	privateID := binary.LittleEndian.Uint64(record.ValueData[8:16])
	createTime := binary.LittleEndian.Uint64(record.ValueData[16:24])
	modTime := binary.LittleEndian.Uint64(record.ValueData[24:32])
	mode := binary.LittleEndian.Uint16(record.ValueData[96:98])

	return &InodeRecord{
		OID:        record.OID,
		ParentID:   parentID,
		PrivateID:  privateID,
		CreateTime: createTime,
		ModTime:    modTime,
		Mode:       mode,
	}, nil
}

// InodeRecord represents a parsed inode record
type InodeRecord struct {
	OID        uint64
	ParentID   uint64
	PrivateID  uint64
	CreateTime uint64
	ModTime    uint64
	Mode       uint16
}