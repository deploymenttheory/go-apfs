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
	cache     *ObjectMapBTreeCache
}

// NewBTreeService creates a new B-tree service
func NewBTreeService(container *ContainerReader) *BTreeService {
	return &BTreeService{
		container: container,
		resolver:  NewBTreeObjectResolver(container),
		cache:     NewObjectMapBTreeCache(DefaultCacheConfig()),
	}
}

// NewBTreeServiceWithCache creates a new B-tree service with custom cache settings
func NewBTreeServiceWithCache(container *ContainerReader, config CacheConfig) *BTreeService {
	return &BTreeService{
		container: container,
		resolver:  NewBTreeObjectResolver(container),
		cache:     NewObjectMapBTreeCache(config),
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
	// Get root node using cache
	rootNode, err := bt.GetRootNode(rootTreeOID, maxXID)
	if err != nil {
		return nil, fmt.Errorf("failed to get root B-tree node: %w", err)
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
		VirtualOID:   virtualOID,
		PhysicalAddr: physAddr,
		XID:          maxXID,
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

			keyOffset := binary.LittleEndian.Uint16(nodeData[offset : offset+2])
			valueOffset := binary.LittleEndian.Uint16(nodeData[offset+2 : offset+4])

			// Extract key and check if it matches our criteria
			keyStart := btnDataStart + int(keyOffset)
			if keyStart+8 <= len(nodeData) {
				objIdAndType := binary.LittleEndian.Uint64(nodeData[keyStart : keyStart+8])
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
			nextKeyOffset := binary.LittleEndian.Uint16(nodeData[nextOffset : nextOffset+2])
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

			keyOffset := binary.LittleEndian.Uint16(nodeData[offset : offset+2])
			valueOffset := binary.LittleEndian.Uint16(nodeData[offset+2 : offset+4])

			// Extract key to check OID range
			keyStart := btnDataStart + int(keyOffset)
			if keyStart+8 <= len(nodeData) {
				objIdAndType := binary.LittleEndian.Uint64(nodeData[keyStart : keyStart+8])
				objID := objIdAndType & types.ObjIdMask

				// If this key's OID is >= our target, this child might contain our target
				if objID >= uint64(targetOID) {
					valueStart := btnDataStart + int(valueOffset)
					if valueStart+8 <= len(nodeData) {
						childOID := types.OidT(binary.LittleEndian.Uint64(nodeData[valueStart : valueStart+8]))
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
	// Get child node using cache
	childNode, err := bt.GetChildNode(childOID, maxXID)
	if err != nil {
		return nil, fmt.Errorf("failed to get child B-tree node with OID %d: %w", childOID, err)
	}

	// Recursively search this child node
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

// GetRootNode retrieves the root B-tree node with caching
func (bt *BTreeService) GetRootNode(rootTreeOID types.OidT, maxXID types.XidT) (interfaces.BTreeNodeReader, error) {
	// Try to get from cache first
	if cachedNode, found := bt.cache.GetNode(rootTreeOID); found {
		return cachedNode, nil
	}

	// Resolve the root tree OID to get physical address
	physAddr, err := bt.resolver.ResolveVirtualObject(rootTreeOID, maxXID)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve root tree OID %d: %w", rootTreeOID, err)
	}

	// Try to get block from cache first
	var blockData []byte
	if cached, found := bt.cache.GetBlock(uint64(physAddr)); found {
		blockData = cached
	} else {
		// Read from container
		var readErr error
		blockData, readErr = bt.container.ReadBlock(uint64(physAddr))
		if readErr != nil {
			return nil, fmt.Errorf("failed to read root B-tree node at address %d: %w", physAddr, readErr)
		}
		// Cache the block
		bt.cache.PutBlock(uint64(physAddr), blockData)
	}

	// Parse the B-tree node
	node, err := btrees.NewBTreeNodeReader(blockData, binary.LittleEndian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse root B-tree node: %w", err)
	}

	// Cache the parsed node
	bt.cache.PutNode(rootTreeOID, node)

	return node, nil
}

// GetChildNode retrieves a child B-tree node with caching
func (bt *BTreeService) GetChildNode(childOID types.OidT, maxXID types.XidT) (interfaces.BTreeNodeReader, error) {
	// Try to get from node cache first
	if cachedNode, found := bt.cache.GetNode(childOID); found {
		return cachedNode, nil
	}

	// Resolve the child OID to get physical address
	physAddr, err := bt.resolver.ResolveVirtualObject(childOID, maxXID)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve child OID %d: %w", childOID, err)
	}

	// Try to get block from cache
	var blockData []byte
	if cached, found := bt.cache.GetBlock(uint64(physAddr)); found {
		blockData = cached
	} else {
		// Read from container
		var readErr error
		blockData, readErr = bt.container.ReadBlock(uint64(physAddr))
		if readErr != nil {
			return nil, fmt.Errorf("failed to read child B-tree node at address %d: %w", physAddr, readErr)
		}
		// Cache the block
		bt.cache.PutBlock(uint64(physAddr), blockData)
	}

	// Parse the B-tree node
	node, err := btrees.NewBTreeNodeReader(blockData, binary.LittleEndian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse child B-tree node: %w", err)
	}

	// Cache the parsed node
	bt.cache.PutNode(childOID, node)

	return node, nil
}

// GetCacheStats returns current cache statistics
func (bt *BTreeService) GetCacheStats() CacheStatistics {
	stats := bt.cache.GetStats()
	return CacheStatistics{
		NodeCachedCount:  stats.NodeCachedCount,
		NodeHits:         stats.NodeHits,
		NodeMisses:       stats.NodeMisses,
		NodeHitRate:      stats.NodeHitRate,
		NodeEvictions:    stats.NodeEvictions,
		BlockCachedCount: stats.BlockCachedCount,
		BlockCachedSize:  stats.BlockCachedSize,
		BlockHits:        stats.BlockHits,
		BlockMisses:      stats.BlockMisses,
		BlockHitRate:     stats.BlockHitRate,
		BlockEvictions:   stats.BlockEvictions,
	}
}

// InvalidateCacheEntry invalidates a specific node from cache
func (bt *BTreeService) InvalidateCacheEntry(oid types.OidT) {
	bt.cache.InvalidateNode(oid)
}

// ClearCache clears all cached data
func (bt *BTreeService) ClearCache() {
	bt.cache.ClearNodeCache()
	bt.cache.ClearBlockCache()
}

// CacheStatistics represents cache statistics
type CacheStatistics struct {
	NodeCachedCount  int
	NodeHits         int64
	NodeMisses       int64
	NodeHitRate      float64
	NodeEvictions    int64
	BlockCachedCount int
	BlockCachedSize  int64
	BlockHits        int64
	BlockMisses      int64
	BlockHitRate     float64
	BlockEvictions   int64
}
