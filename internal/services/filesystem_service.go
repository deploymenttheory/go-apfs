package services

import (
	"encoding/binary"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/parsers/btrees"
	"github.com/deploymenttheory/go-apfs/internal/parsers/file_system_objects"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// FileSystemServiceImpl implements filesystem traversal and directory listing
type FileSystemServiceImpl struct {
	container    *ContainerReader
	resolver     *BTreeObjectResolver
	volumeOID    types.OidT
	volumeSB     *types.ApfsSuperblockT
	rootInodeOID types.OidT
}

// FileEntry represents a file or directory entry
type FileEntry struct {
	Inode    uint64
	Name     string
	Path     string
	IsDir    bool
	Size     uint64
	Mode     uint16
	Modified uint64
}

// NewFileSystemService creates a new FileSystemService instance
func NewFileSystemService(container *ContainerReader, volumeOID types.OidT, volumeSB *types.ApfsSuperblockT) (*FileSystemServiceImpl, error) {
	if container == nil {
		return nil, fmt.Errorf("container reader cannot be nil")
	}
	if volumeOID == 0 {
		return nil, fmt.Errorf("invalid volume OID: 0")
	}
	if volumeSB == nil {
		return nil, fmt.Errorf("volume superblock cannot be nil")
	}

	fs := &FileSystemServiceImpl{
		container:    container,
		resolver:     NewBTreeObjectResolver(container),
		volumeOID:    volumeOID,
		volumeSB:     volumeSB,
		rootInodeOID: types.OidT(volumeSB.ApfsRootTreeOid),
	}

	return fs, nil
}

// ListDirectory lists all entries in a directory by path
func (fs *FileSystemServiceImpl) ListDirectory(path string) ([]FileEntry, error) {
	// Normalize and validate path
	path = filepath.Clean(path)
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	// Get inode for the path
	inode, err := fs.getInodeByPath(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get inode for path %s: %w", path, err)
	}

	// Load the inode data
	inodeData, err := fs.loadInodeData(inode)
	if err != nil {
		return nil, fmt.Errorf("failed to load inode data: %w", err)
	}

	// Parse inode
	inodeReader, err := file_system_objects.NewInodeReader(inodeData.key, inodeData.value, binary.LittleEndian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse inode: %w", err)
	}

	// Check if it's a directory
	if !inodeReader.IsDirectory() {
		return nil, fmt.Errorf("path %s is not a directory", path)
	}

	// List directory contents
	entries, err := fs.listDirectoryContents(inode, path)
	if err != nil {
		return nil, fmt.Errorf("failed to list directory contents: %w", err)
	}

	return entries, nil
}

// WalkTree recursively walks the filesystem tree from a given path
func (fs *FileSystemServiceImpl) WalkTree(startPath string, callback func(*FileEntry) error) error {
	startPath = filepath.Clean(startPath)
	if !strings.HasPrefix(startPath, "/") {
		startPath = "/" + startPath
	}

	return fs.walkTreeRecursive(startPath, callback)
}

// Private helper functions

type inodeData struct {
	key   []byte
	value []byte
}

func (fs *FileSystemServiceImpl) getInodeByPath(path string) (types.OidT, error) {
	if path == "/" {
		return types.OidT(fs.volumeSB.ApfsRootTreeOid), nil
	}

	parts := strings.Split(strings.Trim(path, "/"), "/")
	currentInode := types.OidT(fs.volumeSB.ApfsRootTreeOid)

	for _, part := range parts {
		if part == "" {
			continue
		}

		// List current directory
		entries, err := fs.listDirectoryContents(currentInode, "/")
		if err != nil {
			return 0, fmt.Errorf("failed to list directory: %w", err)
		}

		// Find matching entry
		found := false
		for _, entry := range entries {
			if entry.Name == part {
				currentInode = types.OidT(entry.Inode)
				found = true
				break
			}
		}

		if !found {
			return 0, fmt.Errorf("path component not found: %s", part)
		}
	}

	return currentInode, nil
}

func (fs *FileSystemServiceImpl) loadInodeData(oid types.OidT) (*inodeData, error) {
	// The OID we receive is actually a virtual object ID that points to a B-tree node
	// We need to resolve it and then parse the B-tree node to find the actual inode record

	// Resolve the virtual object ID to physical address
	physAddr, err := fs.resolver.ResolveVirtualObject(oid, fs.container.GetSuperblock().NxNextXid-1)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve inode OID %d: %w", oid, err)
	}

	// Read the block containing the B-tree node
	blockData, err := fs.container.ReadBlock(uint64(physAddr))
	if err != nil {
		return nil, fmt.Errorf("failed to read block at %d: %w", physAddr, err)
	}

	// Parse the block as a B-tree node
	nodeReader, err := btrees.NewBTreeNodeReader(blockData, binary.LittleEndian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse B-tree node: %w", err)
	}

	// Search the B-tree node for the inode record
	return fs.searchBTreeNodeForInode(nodeReader, oid)
}

// searchBTreeNodeForInode searches a B-tree node for an inode record with the given OID
func (fs *FileSystemServiceImpl) searchBTreeNodeForInode(nodeReader interfaces.BTreeNodeReader, targetOID types.OidT) (*inodeData, error) {
	if !nodeReader.IsLeaf() {
		// For internal nodes, we need to traverse to the correct child
		return fs.traverseToInodeLeaf(nodeReader, targetOID)
	}

	// For leaf nodes, search for the inode record
	return fs.extractInodeFromLeaf(nodeReader, targetOID)
}

// traverseToInodeLeaf traverses an internal B-tree node to find the leaf containing the inode
func (fs *FileSystemServiceImpl) traverseToInodeLeaf(nodeReader interfaces.BTreeNodeReader, targetOID types.OidT) (*inodeData, error) {
	tableSpace := nodeReader.TableSpace()
	nodeData := nodeReader.Data()
	keyCount := nodeReader.KeyCount()

	btnDataStart := 56
	tableOffset := btnDataStart + int(tableSpace.Off)

	if tableOffset >= len(nodeData) {
		return nil, fmt.Errorf("table offset exceeds node data")
	}

	var targetChildOID types.OidT

	if nodeReader.HasFixedKVSize() {
		entrySize := 4
		for i := uint32(0); i < keyCount; i++ {
			offset := tableOffset + int(i)*entrySize
			if offset+entrySize > len(nodeData) {
				break
			}

			keyOffset := binary.LittleEndian.Uint16(nodeData[offset : offset+2])
			valueOffset := binary.LittleEndian.Uint16(nodeData[offset+2 : offset+4])

			// Extract and parse the key
			keyStart := btnDataStart + int(keyOffset)
			if keyStart+8 <= len(nodeData) {
				objIdAndType := binary.LittleEndian.Uint64(nodeData[keyStart : keyStart+8])
				objID := objIdAndType & types.ObjIdMask

				// If this key's object ID is >= our target, use this child
				if objID >= uint64(targetOID) {
					valueStart := btnDataStart + int(valueOffset)
					if valueStart+8 <= len(nodeData) {
						targetChildOID = types.OidT(binary.LittleEndian.Uint64(nodeData[valueStart : valueStart+8]))
						break
					}
				}
			}
		}
	}

	if targetChildOID == 0 {
		return nil, fmt.Errorf("no suitable child node found for inode %d", targetOID)
	}

	// Resolve and read the child node
	physAddr, err := fs.resolver.ResolveVirtualObject(targetChildOID, fs.container.GetSuperblock().NxNextXid-1)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve child node: %w", err)
	}

	childData, err := fs.container.ReadBlock(uint64(physAddr))
	if err != nil {
		return nil, fmt.Errorf("failed to read child node: %w", err)
	}

	childReader, err := btrees.NewBTreeNodeReader(childData, binary.LittleEndian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse child node: %w", err)
	}

	// Recursively search the child
	return fs.searchBTreeNodeForInode(childReader, targetOID)
}

// extractInodeFromLeaf extracts an inode record from a leaf B-tree node
func (fs *FileSystemServiceImpl) extractInodeFromLeaf(nodeReader interfaces.BTreeNodeReader, targetOID types.OidT) (*inodeData, error) {
	tableSpace := nodeReader.TableSpace()
	nodeData := nodeReader.Data()
	keyCount := nodeReader.KeyCount()

	btnDataStart := 56
	tableOffset := btnDataStart + int(tableSpace.Off)

	if tableOffset >= len(nodeData) {
		return nil, fmt.Errorf("table offset exceeds node data")
	}

	if nodeReader.HasFixedKVSize() {
		entrySize := 4
		for i := uint32(0); i < keyCount; i++ {
			offset := tableOffset + int(i)*entrySize
			if offset+entrySize > len(nodeData) {
				break
			}

			keyOffset := binary.LittleEndian.Uint16(nodeData[offset : offset+2])
			valueOffset := binary.LittleEndian.Uint16(nodeData[offset+2 : offset+4])

			// Extract key and check if it matches our target inode
			keyStart := btnDataStart + int(keyOffset)
			if keyStart+8 <= len(nodeData) {
				objIdAndType := binary.LittleEndian.Uint64(nodeData[keyStart : keyStart+8])
				objID := objIdAndType & types.ObjIdMask
				objType := (objIdAndType & types.ObjTypeMask) >> types.ObjTypeShift

				// Check if this is an inode record for our target OID
				if objID == uint64(targetOID) && objType == uint64(types.ApfsTypeInode) {
					// Extract the key and value data
					valueStart := btnDataStart + int(valueOffset)

					// Determine key and value sizes
					var keyData, valueData []byte

					// For fixed-size entries, we need to determine the actual sizes
					// This is simplified - in reality we'd need to look at the B-tree info
					keySize := 8 // Minimum size for JInodeKeyT
					if keyStart+keySize <= len(nodeData) {
						keyData = nodeData[keyStart : keyStart+keySize]
					}

					// Value size is variable, read what we can
					if valueStart < len(nodeData) {
						// Find the next entry to determine value size, or use remaining data
						var valueSize int
						if i+1 < keyCount {
							// Get next entry's offset to calculate this value's size
							nextOffset := tableOffset + int(i+1)*entrySize
							if nextOffset+2 <= len(nodeData) {
								nextKeyOffset := binary.LittleEndian.Uint16(nodeData[nextOffset : nextOffset+2])
								nextKeyStart := btnDataStart + int(nextKeyOffset)
								valueSize = nextKeyStart - valueStart
							}
						} else {
							// Last entry, use remaining data
							valueSize = len(nodeData) - valueStart
						}

						if valueSize > 0 && valueStart+valueSize <= len(nodeData) {
							valueData = nodeData[valueStart : valueStart+valueSize]
						}
					}

					if len(keyData) > 0 && len(valueData) > 0 {
						return &inodeData{
							key:   keyData,
							value: valueData,
						}, nil
					}
				}
			}
		}
	}

	return nil, fmt.Errorf("inode %d not found in leaf node", targetOID)
}

func (fs *FileSystemServiceImpl) listDirectoryContents(dirInode types.OidT, dirPath string) ([]FileEntry, error) {
	var entries []FileEntry

	// Load directory inode
	inodeData, err := fs.loadInodeData(dirInode)
	if err != nil {
		return nil, fmt.Errorf("failed to load directory inode: %w", err)
	}

	// Parse inode to get directory extents
	inodeReader, err := file_system_objects.NewInodeReader(inodeData.key, inodeData.value, binary.LittleEndian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse directory inode: %w", err)
	}

	if !inodeReader.IsDirectory() {
		return nil, fmt.Errorf("inode is not a directory")
	}

	// Get file extents from directory inode
	extents, err := fs.getFileExtents(inodeReader)
	if err != nil {
		return nil, fmt.Errorf("failed to get file extents: %w", err)
	}
	if len(extents) == 0 {
		return entries, nil // Empty directory
	}

	// Read directory data from extents
	for _, extent := range extents {
		startBlock := extent.PhysicalBlock
		blockCount := extent.PhysicalSize / uint64(fs.container.GetBlockSize())

		// Read all blocks in this extent
		for i := uint64(0); i < blockCount; i++ {
			blockData, err := fs.container.ReadBlock(startBlock + i)
			if err != nil {
				continue // Skip corrupted blocks
			}

			// Parse directory entries from this block
			blockEntries, err := fs.parseDirectoryBlock(blockData, dirPath)
			if err != nil {
				continue // Skip unparseable blocks
			}

			entries = append(entries, blockEntries...)
		}
	}

	return entries, nil
}

func (fs *FileSystemServiceImpl) walkTreeRecursive(dirPath string, callback func(*FileEntry) error) error {
	entries, err := fs.ListDirectory(dirPath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if err := callback(&entry); err != nil {
			return err
		}

		if entry.IsDir {
			fullPath := filepath.Join(dirPath, entry.Name)
			if err := fs.walkTreeRecursive(fullPath, callback); err != nil {
				return err
			}
		}
	}

	return nil
}

// parseDirectoryBlock parses directory entries from a block of data
func (fs *FileSystemServiceImpl) parseDirectoryBlock(blockData []byte, dirPath string) ([]FileEntry, error) {
	var entries []FileEntry

	// Directory blocks contain B-tree nodes with directory records
	// Try to parse as B-tree node first
	nodeReader, err := fs.parseDirectoryBTreeNode(blockData)
	if err != nil {
		// If B-tree parsing fails, try direct directory record parsing
		return fs.parseDirectoryRecords(blockData, dirPath)
	}

	// If it's a B-tree node, extract directory records from it
	if nodeReader != nil && nodeReader.IsLeaf() {
		entries = fs.extractDirectoryEntriesFromBTreeNode(nodeReader, dirPath)
	}

	return entries, nil
}

// parseDirectoryBTreeNode attempts to parse block data as a B-tree node
func (fs *FileSystemServiceImpl) parseDirectoryBTreeNode(blockData []byte) (interfaces.BTreeNodeReader, error) {
	// Check if this looks like a B-tree node by examining the header
	if len(blockData) < 56 {
		return nil, fmt.Errorf("block too small for B-tree node")
	}

	// Check object type in header
	objType := binary.LittleEndian.Uint32(blockData[24:28])
	if objType&0x0000FFFF != types.ObjectTypeBtreeNode {
		return nil, fmt.Errorf("not a B-tree node")
	}

	// Parse as B-tree node using the existing btrees package
	return fs.parseBTreeNodeForDirectory(blockData)
}

// parseDirectoryRecords parses directory records directly from block data
func (fs *FileSystemServiceImpl) parseDirectoryRecords(blockData []byte, dirPath string) ([]FileEntry, error) {
	var entries []FileEntry
	offset := 0

	// Parse directory records sequentially
	for offset+16 < len(blockData) { // Minimum record size
		// Try to parse a directory record at this offset
		record, recordSize, err := fs.parseDirectoryRecord(blockData[offset:])
		if err != nil {
			break // End of valid records
		}

		if record != nil {
			entry := FileEntry{
				Inode:    record.InodeNumber,
				Name:     record.Name,
				Path:     filepath.Join(dirPath, record.Name),
				IsDir:    uint16(record.FileType) == types.DtDir,
				Size:     0, // Size not available in directory record
				Mode:     0, // Mode not available in directory record
				Modified: 0, // Modified time not available in directory record
			}
			entries = append(entries, entry)
		}

		offset += recordSize
		if recordSize == 0 {
			break // Prevent infinite loop
		}
	}

	return entries, nil
}

// DirectoryRecord represents a parsed directory record
type DirectoryRecord struct {
	InodeNumber uint64
	Name        string
	FileType    uint8
}

// parseDirectoryRecord parses a single directory record from data
func (fs *FileSystemServiceImpl) parseDirectoryRecord(data []byte) (*DirectoryRecord, int, error) {
	if len(data) < 16 {
		return nil, 0, fmt.Errorf("insufficient data for directory record")
	}

	// Parse APFS directory record structure
	// This is a simplified version - actual APFS dir records are more complex
	inodeNumber := binary.LittleEndian.Uint64(data[0:8])
	nameLength := binary.LittleEndian.Uint16(data[8:10])
	fileType := data[10]

	if nameLength == 0 || int(nameLength) > len(data)-16 {
		return nil, 0, fmt.Errorf("invalid name length")
	}

	name := string(data[16 : 16+nameLength])
	recordSize := 16 + int(nameLength)

	// Align to 8-byte boundary
	if recordSize%8 != 0 {
		recordSize += 8 - (recordSize % 8)
	}

	record := &DirectoryRecord{
		InodeNumber: inodeNumber,
		Name:        name,
		FileType:    fileType,
	}

	return record, recordSize, nil
}

// parseBTreeNodeForDirectory parses a B-tree node containing directory entries
func (fs *FileSystemServiceImpl) parseBTreeNodeForDirectory(blockData []byte) (interfaces.BTreeNodeReader, error) {
	// Use the existing btrees package to parse the node
	return btrees.NewBTreeNodeReader(blockData, binary.LittleEndian)
}

// extractDirectoryEntriesFromBTreeNode extracts directory entries from a B-tree leaf node
func (fs *FileSystemServiceImpl) extractDirectoryEntriesFromBTreeNode(nodeReader interfaces.BTreeNodeReader, dirPath string) []FileEntry {
	// This would extract actual directory entries from the B-tree node
	// For now, return empty slice
	return []FileEntry{}
}

// IsDirectory checks if a path is a directory
func (fs *FileSystemServiceImpl) IsDirectory(path string) (bool, error) {
	node, err := fs.GetInodeByPath(path)
	if err != nil {
		return false, err
	}
	return node.IsDirectory, nil
}

// Exists checks if a path exists
func (fs *FileSystemServiceImpl) Exists(path string) (bool, error) {
	_, err := fs.GetInodeByPath(path)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// getFileExtents extracts file extents from an inode by looking up extent records
func (fs *FileSystemServiceImpl) getFileExtents(inodeReader interfaces.InodeReader) ([]ExtentMapping, error) {
	var extents []ExtentMapping

	// Get the private ID from the inode - this is used to look up extent records
	privateID := inodeReader.PrivateID()
	if privateID == 0 {
		return extents, nil // No data stream
	}

	// In APFS, file extents are stored as separate records in the volume's B-tree
	// We need to search for records of type APFS_TYPE_FILE_EXTENT with our private ID
	// This requires B-tree traversal of the volume's filesystem records

	// For now, implement a simplified version that would need to be expanded
	// with proper B-tree searching for extent records

	// The actual implementation would:
	// 1. Search the volume's filesystem B-tree for records with:
	//    - Object type: APFS_TYPE_FILE_EXTENT
	//    - Object ID: privateID
	// 2. Parse each JFileExtentKeyT/JFileExtentValT pair
	// 3. Convert to ExtentMapping structs

	extent, err := fs.findFileExtentForPrivateID(privateID)
	if err != nil {
		return extents, nil // No extents found
	}

	if extent != nil {
		extents = append(extents, *extent)
	}

	return extents, nil
}

// findFileExtentForPrivateID searches for file extent records for a given private ID
func (fs *FileSystemServiceImpl) findFileExtentForPrivateID(privateID uint64) (*ExtentMapping, error) {
	// Get the volume's filesystem tree OID
	fsTreeOID := types.OidT(fs.volumeSB.ApfsRootTreeOid)
	if fsTreeOID == 0 {
		return nil, fmt.Errorf("volume filesystem tree OID is zero")
	}

	// Resolve the filesystem tree OID to physical address
	physAddr, err := fs.resolver.ResolveVirtualObject(fsTreeOID, fs.container.GetSuperblock().NxNextXid-1)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve filesystem tree OID: %w", err)
	}

	// Read the filesystem tree root node
	treeData, err := fs.container.ReadBlock(uint64(physAddr))
	if err != nil {
		return nil, fmt.Errorf("failed to read filesystem tree root: %w", err)
	}

	// Parse as B-tree node
	nodeReader, err := btrees.NewBTreeNodeReader(treeData, binary.LittleEndian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse filesystem tree node: %w", err)
	}

	// Search for file extent records with matching private ID
	return fs.searchFileSystemTreeForExtents(nodeReader, privateID)
}

// searchFileSystemTreeForExtents searches the filesystem B-tree for file extent records
func (fs *FileSystemServiceImpl) searchFileSystemTreeForExtents(nodeReader interfaces.BTreeNodeReader, privateID uint64) (*ExtentMapping, error) {
	if !nodeReader.IsLeaf() {
		// For internal nodes, we need to traverse to children
		return fs.traverseInternalNode(nodeReader, privateID)
	}

	// For leaf nodes, parse the records looking for file extent records
	return fs.parseLeafNodeForExtents(nodeReader, privateID)
}

// traverseInternalNode traverses an internal B-tree node to find the correct child
func (fs *FileSystemServiceImpl) traverseInternalNode(nodeReader interfaces.BTreeNodeReader, privateID uint64) (*ExtentMapping, error) {
	// Get table space and parse entries
	tableSpace := nodeReader.TableSpace()
	nodeData := nodeReader.Data()
	keyCount := nodeReader.KeyCount()

	// Calculate table offset - table is relative to btn_data which starts at offset 56
	btnDataStart := 56
	tableOffset := btnDataStart + int(tableSpace.Off)

	if tableOffset >= len(nodeData) {
		return nil, fmt.Errorf("table offset exceeds node data")
	}

	// Parse table entries to find the appropriate child node
	var targetChildOID types.OidT

	if nodeReader.HasFixedKVSize() {
		// Fixed-size entries (kvoff_t)
		entrySize := 4
		for i := uint32(0); i < keyCount; i++ {
			offset := tableOffset + int(i)*entrySize
			if offset+entrySize > len(nodeData) {
				break
			}

			keyOffset := binary.LittleEndian.Uint16(nodeData[offset : offset+2])
			valueOffset := binary.LittleEndian.Uint16(nodeData[offset+2 : offset+4])

			// Extract and parse the key
			keyStart := btnDataStart + int(keyOffset)
			if keyStart+16 <= len(nodeData) {
				// Parse the key as JKeyT to check object ID and type
				objIdAndType := binary.LittleEndian.Uint64(nodeData[keyStart : keyStart+8])
				objID := objIdAndType & types.ObjIdMask

				// If this key's object ID is >= our target, use this child
				if objID >= privateID {
					// Extract child OID from value
					valueStart := btnDataStart + int(valueOffset)
					if valueStart+8 <= len(nodeData) {
						targetChildOID = types.OidT(binary.LittleEndian.Uint64(nodeData[valueStart : valueStart+8]))
						break
					}
				}
			}
		}
	}

	if targetChildOID == 0 {
		return nil, fmt.Errorf("no suitable child node found")
	}

	// Resolve and read the child node
	physAddr, err := fs.resolver.ResolveVirtualObject(targetChildOID, fs.container.GetSuperblock().NxNextXid-1)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve child node: %w", err)
	}

	childData, err := fs.container.ReadBlock(uint64(physAddr))
	if err != nil {
		return nil, fmt.Errorf("failed to read child node: %w", err)
	}

	childReader, err := btrees.NewBTreeNodeReader(childData, binary.LittleEndian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse child node: %w", err)
	}

	// Recursively search the child
	return fs.searchFileSystemTreeForExtents(childReader, privateID)
}

// parseLeafNodeForExtents parses a leaf node looking for file extent records
func (fs *FileSystemServiceImpl) parseLeafNodeForExtents(nodeReader interfaces.BTreeNodeReader, privateID uint64) (*ExtentMapping, error) {
	tableSpace := nodeReader.TableSpace()
	nodeData := nodeReader.Data()
	keyCount := nodeReader.KeyCount()

	btnDataStart := 56
	tableOffset := btnDataStart + int(tableSpace.Off)

	if tableOffset >= len(nodeData) {
		return nil, fmt.Errorf("table offset exceeds node data")
	}

	// Parse table entries looking for file extent records
	if nodeReader.HasFixedKVSize() {
		entrySize := 4
		for i := uint32(0); i < keyCount; i++ {
			offset := tableOffset + int(i)*entrySize
			if offset+entrySize > len(nodeData) {
				break
			}

			keyOffset := binary.LittleEndian.Uint16(nodeData[offset : offset+2])
			valueOffset := binary.LittleEndian.Uint16(nodeData[offset+2 : offset+4])

			// Extract key and check if it's a file extent record
			keyStart := btnDataStart + int(keyOffset)
			if keyStart+16 <= len(nodeData) {
				// Parse JFileExtentKeyT
				objIdAndType := binary.LittleEndian.Uint64(nodeData[keyStart : keyStart+8])
				objID := objIdAndType & types.ObjIdMask
				objType := (objIdAndType & types.ObjTypeMask) >> types.ObjTypeShift

				// Check if this is a file extent record for our private ID
				if objID == privateID && objType == uint64(types.ApfsTypeFileExtent) {
					// Extract logical address from key
					logicalAddr := binary.LittleEndian.Uint64(nodeData[keyStart+8 : keyStart+16])

					// Extract value (JFileExtentValT)
					valueStart := btnDataStart + int(valueOffset)
					if valueStart+24 <= len(nodeData) {
						lenAndFlags := binary.LittleEndian.Uint64(nodeData[valueStart : valueStart+8])
						physBlockNum := binary.LittleEndian.Uint64(nodeData[valueStart+8 : valueStart+16])

						// Extract length from flags field
						physicalSize := lenAndFlags & types.JFileExtentLenMask

						return &ExtentMapping{
							LogicalOffset:   logicalAddr,
							LogicalSize:     physicalSize,
							PhysicalBlock:   physBlockNum,
							PhysicalSize:    physicalSize,
							IsCompressed:    false, // Would need to check flags
							CompressionType: "",
							IsEncrypted:     false, // Would need to check encryption fields
						}, nil
					}
				}
			}
		}
	}

	return nil, fmt.Errorf("no file extent found for private ID %d", privateID)
}

// GetInodeByPath gets the inode for a given path and returns FileNode metadata
func (fs *FileSystemServiceImpl) GetInodeByPath(path string) (*FileNode, error) {
	path = filepath.Clean(path)
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	// Get the inode OID
	inodeOID, err := fs.getInodeByPath(path)
	if err != nil {
		// If we can't get inode by path, try alternative approaches
		// This handles forensic scenarios where object maps might be incomplete
		if path == "/" {
			// For root, try to find it by scanning
			inodeOID, err = fs.findRootTreeByScanning()
			if err != nil {
				return nil, fmt.Errorf("failed to get inode for path %s: %w", path, err)
			}
		} else {
			return nil, fmt.Errorf("failed to get inode for path %s: %w", path, err)
		}
	}

	// Load inode data
	inodeData, err := fs.loadInodeData(inodeOID)
	if err != nil {
		return nil, fmt.Errorf("failed to load inode data: %w", err)
	}

	// Parse inode
	inodeReader, err := file_system_objects.NewInodeReader(inodeData.key, inodeData.value, binary.LittleEndian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse inode: %w", err)
	}

	// Extract metadata and convert to FileNode
	return fs.inodeToFileNode(path, inodeOID, inodeReader)
}

// findRootTreeByScanning scans the container for the root tree B-tree structure
// This is used when object map lookup fails (forensic/recovery scenarios)
func (fs *FileSystemServiceImpl) findRootTreeByScanning() (types.OidT, error) {
	// The root tree OID is stored in volumeSB.ApfsRootTreeOid
	// In normal scenarios, this is a virtual OID that needs resolving through the object map
	// In forensic scenarios (empty object maps), we try the OID as a physical address
	rootTreeOID := types.OidT(fs.volumeSB.ApfsRootTreeOid)

	// Try treating it as a direct physical block address first
	blockData, err := fs.container.ReadBlock(uint64(rootTreeOID))
	if err == nil && isValidBTreeNode(blockData) {
		return rootTreeOID, nil
	}

	// Try scanning nearby blocks for B-tree node structure
	for scanOffset := uint64(0); scanOffset < 100; scanOffset++ {
		blockData, err := fs.container.ReadBlock(uint64(rootTreeOID) + scanOffset)
		if err == nil && isValidBTreeNode(blockData) {
			return types.OidT(uint64(rootTreeOID) + scanOffset), nil
		}
	}

	return 0, fmt.Errorf("could not locate root tree (OID=%d) via object map or scanning", rootTreeOID)
}

// isValidBTreeNode checks if a block contains a valid B-tree node structure
func isValidBTreeNode(data []byte) bool {
	if len(data) < 64 {
		return false
	}

	// B-tree nodes start with obj_phys_t header
	// Check for valid object header pattern
	// The flags field at offset 8 should have reasonable values
	flags := uint32(data[8]) | uint32(data[9])<<8 | uint32(data[10])<<16 | uint32(data[11])<<24

	// Check if this looks like a B-tree node (flags shouldn't be all zeros or all ones)
	if flags == 0 || flags == 0xFFFFFFFF {
		return false
	}

	// Check B-tree specific fields
	// btn_level at offset 76 should be reasonable (0-20 for most trees)
	if len(data) >= 78 {
		level := uint16(data[76]) | uint16(data[77])<<8
		if level > 100 {
			return false
		}
	}

	return true
}

// ListDirectoryContents lists all entries in a directory by inode ID
func (fs *FileSystemServiceImpl) ListDirectoryContents(inodeID uint64) ([]*FileNode, error) {
	// Load inode data
	inodeData, err := fs.loadInodeData(types.OidT(inodeID))
	if err != nil {
		return nil, fmt.Errorf("failed to load inode data: %w", err)
	}

	// Parse inode
	inodeReader, err := file_system_objects.NewInodeReader(inodeData.key, inodeData.value, binary.LittleEndian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse inode: %w", err)
	}

	// Check if it's a directory
	if !inodeReader.IsDirectory() {
		return nil, fmt.Errorf("inode %d is not a directory", inodeID)
	}

	// Get directory contents as FileEntry
	entries, err := fs.listDirectoryContents(types.OidT(inodeID), "/")
	if err != nil {
		return nil, fmt.Errorf("failed to list directory contents: %w", err)
	}

	// Convert FileEntry to FileNode
	var fileNodes []*FileNode
	for _, entry := range entries {
		// Load the entry's inode data to get full metadata
		entryInodeData, err := fs.loadInodeData(types.OidT(entry.Inode))
		if err != nil {
			continue // Skip entries we can't load
		}

		entryInodeReader, err := file_system_objects.NewInodeReader(entryInodeData.key, entryInodeData.value, binary.LittleEndian)
		if err != nil {
			continue // Skip entries we can't parse
		}

		fileNode, err := fs.inodeToFileNode(entry.Path, types.OidT(entry.Inode), entryInodeReader)
		if err != nil {
			continue // Skip entries with errors
		}

		fileNodes = append(fileNodes, fileNode)
	}

	return fileNodes, nil
}

// GetFileExtents gets the physical extent mappings for a file by inode ID
func (fs *FileSystemServiceImpl) GetFileExtents(inodeID uint64) ([]ExtentMapping, error) {
	// Load inode data
	inodeData, err := fs.loadInodeData(types.OidT(inodeID))
	if err != nil {
		return nil, fmt.Errorf("failed to load inode data: %w", err)
	}

	// Parse inode
	inodeReader, err := file_system_objects.NewInodeReader(inodeData.key, inodeData.value, binary.LittleEndian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse inode: %w", err)
	}

	// Get extents
	extents, err := fs.getFileExtents(inodeReader)
	if err != nil {
		return nil, fmt.Errorf("failed to get file extents: %w", err)
	}

	return extents, nil
}

// GetFileMetadata gets metadata for a file by inode ID (interface implementation)
func (fs *FileSystemServiceImpl) GetFileMetadata(inodeID uint64) (*FileNode, error) {
	// Load inode data
	inodeData, err := fs.loadInodeData(types.OidT(inodeID))
	if err != nil {
		return nil, fmt.Errorf("failed to load inode data: %w", err)
	}

	// Parse inode
	inodeReader, err := file_system_objects.NewInodeReader(inodeData.key, inodeData.value, binary.LittleEndian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse inode: %w", err)
	}

	// Convert to FileNode and return
	return fs.inodeToFileNode("", types.OidT(inodeID), inodeReader)
}

// FindFilesByName searches for files matching a name pattern
func (fs *FileSystemServiceImpl) FindFilesByName(pattern string, maxResults int) ([]*FileNode, error) {
	var results []*FileNode

	// Walk the tree from root and search for matching names
	err := fs.WalkTree("/", func(entry *FileEntry) error {
		if len(results) >= maxResults {
			return fmt.Errorf("max results reached")
		}

		// Simple pattern matching (supports * wildcards)
		if fs.matchPattern(entry.Name, pattern) {
			// Load inode data for full metadata
			inodeData, err := fs.loadInodeData(types.OidT(entry.Inode))
			if err != nil {
				return nil // Skip this entry
			}

			inodeReader, err := file_system_objects.NewInodeReader(inodeData.key, inodeData.value, binary.LittleEndian)
			if err != nil {
				return nil // Skip this entry
			}

			fileNode, err := fs.inodeToFileNode(entry.Path, types.OidT(entry.Inode), inodeReader)
			if err != nil {
				return nil // Skip this entry
			}

			results = append(results, fileNode)
		}

		return nil
	})

	// If we hit max results, that's not an error
	if err != nil && err.Error() == "max results reached" {
		return results, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to search for files: %w", err)
	}

	return results, nil
}

// GetParentDirectory gets the parent directory of an inode
func (fs *FileSystemServiceImpl) GetParentDirectory(inodeID uint64) (*FileNode, error) {
	// Load inode data to get parent reference
	inodeData, err := fs.loadInodeData(types.OidT(inodeID))
	if err != nil {
		return nil, fmt.Errorf("failed to load inode data: %w", err)
	}

	// Parse inode
	inodeReader, err := file_system_objects.NewInodeReader(inodeData.key, inodeData.value, binary.LittleEndian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse inode: %w", err)
	}

	// Get parent inode ID from inode - use ParentID() method
	parentInodeID := inodeReader.ParentID()
	if parentInodeID == 0 {
		return nil, fmt.Errorf("inode %d has no parent", inodeID)
	}

	// Load parent inode data
	parentInodeData, err := fs.loadInodeData(types.OidT(parentInodeID))
	if err != nil {
		return nil, fmt.Errorf("failed to load parent inode data: %w", err)
	}

	// Parse parent inode
	parentInodeReader, err := file_system_objects.NewInodeReader(parentInodeData.key, parentInodeData.value, binary.LittleEndian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse parent inode: %w", err)
	}

	// Convert to FileNode and return
	return fs.inodeToFileNode("", types.OidT(parentInodeID), parentInodeReader)
}

// IsPathAccessible checks if a path is accessible
func (fs *FileSystemServiceImpl) IsPathAccessible(path string) (bool, error) {
	path = filepath.Clean(path)
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	_, err := fs.getInodeByPath(path)
	if err != nil {
		return false, nil
	}

	return true, nil
}

// Helper method to convert inode data to FileNode
func (fs *FileSystemServiceImpl) inodeToFileNode(path string, inodeID types.OidT, inodeReader interfaces.InodeReader) (*FileNode, error) {
	// Extract inode metadata using correct InodeReader API methods
	mode := inodeReader.Mode()
	// Get file size from inode - this is calculated from the data stream size
	size := inodeReader.Size()
	parentID := inodeReader.ParentID()
	mtime := inodeReader.ModificationTime()
	ctime := inodeReader.ChangeTime()
	atime := inodeReader.AccessTime()
	btime := inodeReader.CreationTime()
	uid := uint32(inodeReader.Owner())
	gid := uint32(inodeReader.Group())
	hardLinkCount := uint32(inodeReader.NumberOfHardLinks())
	flags := uint32(inodeReader.Flags())

	isDir := inodeReader.IsDirectory()
	// Mode bits can indicate symlink: check S_ISLNK (120000 octal)
	isSymlink := (uint16(mode) & 0o170000) == 0o120000
	// Check encryption status from inode flags (INODE_IS_ENCRYPTED flag would indicate encryption)
	isEncrypted := (flags & uint32(types.InodeProtClassExplicit)) != 0

	return &FileNode{
		Inode:         uint64(inodeID),
		Path:          path,
		Name:          filepath.Base(path),
		Mode:          uint16(mode),
		Size:          size,
		CreatedTime:   btime,
		ModifiedTime:  mtime,
		ChangedTime:   ctime,
		AccessedTime:  atime,
		UID:           uid,
		GID:           gid,
		IsDirectory:   isDir,
		IsSymlink:     isSymlink,
		IsEncrypted:   isEncrypted,
		ParentInode:   parentID,
		HardLinkCount: hardLinkCount,
		Flags:         flags,
	}, nil
}

// Helper method for pattern matching with * wildcard support
func (fs *FileSystemServiceImpl) matchPattern(name, pattern string) bool {
	if pattern == "*" {
		return true
	}
	if strings.Contains(pattern, "*") {
		parts := strings.Split(pattern, "*")
		if len(parts) == 2 {
			if parts[0] == "" && parts[1] == "" {
				return true
			}
			if parts[0] != "" && !strings.HasPrefix(name, parts[0]) {
				return false
			}
			if parts[1] != "" && !strings.HasSuffix(name, parts[1]) {
				return false
			}
			return true
		}
	}
	return name == pattern
}

// ReadFile reads the entire content of a file by inode ID
func (fs *FileSystemServiceImpl) ReadFile(inodeID uint64) ([]byte, error) {
	// Get file size first
	fileSize, err := fs.GetFileSize(inodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get file size: %w", err)
	}

	if fileSize == 0 {
		return []byte{}, nil
	}

	// Read entire file
	return fs.ReadFileRange(inodeID, 0, fileSize)
}

// ReadFileRange reads a specific range of bytes from a file
func (fs *FileSystemServiceImpl) ReadFileRange(inodeID uint64, offset, length uint64) ([]byte, error) {
	// Get file extents
	extents, err := fs.GetFileExtents(inodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get file extents: %w", err)
	}

	if len(extents) == 0 {
		return nil, fmt.Errorf("file has no extents")
	}

	// Build a map of logical to physical extents for efficient access
	var data []byte
	currentLogicalOffset := uint64(0)

	for _, extent := range extents {
		extentLogicalEnd := currentLogicalOffset + extent.LogicalSize

		// Check if this extent overlaps with our requested range
		if extentLogicalEnd > offset && currentLogicalOffset < offset+length {
			// Calculate the actual read range within this extent
			readStartInExtent := uint64(0)
			if offset > currentLogicalOffset {
				readStartInExtent = offset - currentLogicalOffset
			}

			readEndInExtent := extent.LogicalSize
			if offset+length < extentLogicalEnd {
				readEndInExtent = offset + length - currentLogicalOffset
			}

			bytesToRead := readEndInExtent - readStartInExtent

			// Check for sparse extent (PhysicalBlock == 0 means this is a "hole" in the file)
			if extent.PhysicalBlock == 0 {
				// Sparse extent - return zeros without reading from disk
				zeroData := make([]byte, bytesToRead)
				data = append(data, zeroData...)
			} else if extent.IsCompressed {
				// Handle compressed extent
				extentData, err := fs.readCompressedExtent(extent, readStartInExtent, bytesToRead)
				if err != nil {
					return nil, fmt.Errorf("failed to read compressed extent: %w", err)
				}
				data = append(data, extentData...)
			} else {
				// Read from the physical location
				physicalOffset := extent.PhysicalBlock*uint64(fs.container.GetBlockSize()) + readStartInExtent
				blockOffset := physicalOffset / uint64(fs.container.GetBlockSize())
				blockInternalOffset := physicalOffset % uint64(fs.container.GetBlockSize())

				// Read blocks
				var extentData []byte
				remainingBytes := bytesToRead
				currentBlock := blockOffset
				blockInternalPos := blockInternalOffset

				for remainingBytes > 0 {
					blockData, err := fs.container.ReadBlock(currentBlock)
					if err != nil {
						return nil, fmt.Errorf("failed to read block %d: %w", currentBlock, err)
					}

					// Calculate how many bytes to read from this block
					bytesInBlock := uint64(len(blockData)) - blockInternalPos
					if bytesInBlock > remainingBytes {
						bytesInBlock = remainingBytes
					}

					if bytesInBlock > 0 {
						extentData = append(extentData, blockData[blockInternalPos:blockInternalPos+bytesInBlock]...)
						remainingBytes -= bytesInBlock
					}

					// Move to next block
					currentBlock++
					blockInternalPos = 0
				}

				data = append(data, extentData...)
			}
		}

		currentLogicalOffset = extentLogicalEnd
	}

	// Verify we read the expected amount
	if uint64(len(data)) < length {
		// This might be OK if we're reading beyond EOF
		return data, nil
	}

	return data, nil
}

// readCompressedExtent reads and decompresses a compressed extent
func (fs *FileSystemServiceImpl) readCompressedExtent(extent ExtentMapping, offsetInExtent, bytesToRead uint64) ([]byte, error) {
	// Read the compressed data from the physical location
	blockSize := fs.container.GetBlockSize()
	physicalOffset := extent.PhysicalBlock * uint64(blockSize)

	// Read the entire compressed block(s)
	blockOffset := physicalOffset / uint64(blockSize)
	compressedData, err := fs.container.ReadBlock(blockOffset)
	if err != nil {
		return nil, fmt.Errorf("failed to read compressed block: %w", err)
	}

	// Parse compression method from the type string or infer from data
	var compressionMethod types.CompressionMethodType
	switch extent.CompressionType {
	case "DEFLATE":
		compressionMethod = types.CompressionMethodDeflate
	case "LZFSE":
		compressionMethod = types.CompressionMethodLzfse
	case "LZVN":
		compressionMethod = types.CompressionMethodLzvn
	case "LZ4":
		compressionMethod = types.CompressionMethodLz4
	case "ZSTD":
		compressionMethod = types.CompressionMethodZstd
	default:
		// Try to infer from compressed data header
		if len(compressedData) >= 4 {
			signature := binary.LittleEndian.Uint32(compressedData[0:4])
			if signature == types.CompressionSignature {
				// This is a compressed data block with header
				if len(compressedData) < 16 {
					return nil, fmt.Errorf("insufficient data for compression header")
				}
				compressionMethod = types.CompressionMethodType(binary.LittleEndian.Uint32(compressedData[4:8]))
			} else {
				compressionMethod = types.CompressionMethodDeflate // Default assumption
			}
		} else {
			return nil, fmt.Errorf("unknown compression method: %s", extent.CompressionType)
		}
	}

	// Decompress the data
	compressionService := NewCompressionService()
	decompressed, err := compressionService.Decompress(compressedData, compressionMethod)
	if err != nil {
		return nil, fmt.Errorf("decompression failed: %w", err)
	}

	// Extract the requested range from decompressed data
	if offsetInExtent+bytesToRead > uint64(len(decompressed)) {
		bytesToRead = uint64(len(decompressed)) - offsetInExtent
	}

	if offsetInExtent >= uint64(len(decompressed)) {
		return nil, fmt.Errorf("offset beyond decompressed data size")
	}

	return decompressed[offsetInExtent : offsetInExtent+bytesToRead], nil
}

// GetFileSize returns the size of a file in bytes
func (fs *FileSystemServiceImpl) GetFileSize(inodeID uint64) (uint64, error) {
	// Load inode data
	inodeData, err := fs.loadInodeData(types.OidT(inodeID))
	if err != nil {
		return 0, fmt.Errorf("failed to load inode data: %w", err)
	}

	// Parse inode
	inodeReader, err := file_system_objects.NewInodeReader(inodeData.key, inodeData.value, binary.LittleEndian)
	if err != nil {
		return 0, fmt.Errorf("failed to parse inode: %w", err)
	}

	// Return size - for regular files, this is the size field
	// For compressed files, we need to use UncompressedSize
	size := inodeReader.Size()
	return size, nil
}

// CreateFileReader creates an io.Reader for streaming file content
func (fs *FileSystemServiceImpl) CreateFileReader(inodeID uint64) (io.Reader, error) {
	fileSize, err := fs.GetFileSize(inodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get file size: %w", err)
	}

	return &FileReaderAdapter{
		fs:      fs,
		inodeID: inodeID,
		size:    fileSize,
		offset:  0,
	}, nil
}

// CreateFileSeeker creates an io.ReadSeeker for random access to file content
func (fs *FileSystemServiceImpl) CreateFileSeeker(inodeID uint64) (io.ReadSeeker, error) {
	fileSize, err := fs.GetFileSize(inodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get file size: %w", err)
	}

	return &FileSeekerAdapter{
		fs:      fs,
		inodeID: inodeID,
		size:    fileSize,
		offset:  0,
	}, nil
}

// GetExtendedAttributes retrieves all extended attributes for an inode
func (fs *FileSystemServiceImpl) GetExtendedAttributes(inodeID uint64) (map[string][]byte, error) {
	// Load inode data
	inodeData, err := fs.loadInodeData(types.OidT(inodeID))
	if err != nil {
		return nil, fmt.Errorf("failed to load inode data: %w", err)
	}

	// Parse inode
	inodeReader, err := file_system_objects.NewInodeReader(inodeData.key, inodeData.value, binary.LittleEndian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse inode: %w", err)
	}

	// Query the filesystem tree for extended attributes
	attributes := make(map[string][]byte)

	// Scan the file system B-tree for extended attribute records
	// Extended attributes are stored as separate B-tree entries
	// This would require traversing the B-tree looking for EA records related to this inode
	// For now, return empty map as a placeholder
	// Full implementation would require:
	// 1. Search B-tree for EA keys matching this inode OID
	// 2. Parse each EA value
	// 3. Aggregate into the map

	_ = inodeReader // Use the reader
	return attributes, nil
}

// GetFileExtents returns all extents for a file (already implemented, needed for interface)
// This is a duplicate but required by the interface

// VerifyFileChecksum verifies the integrity of a file's metadata
//
// IMPORTANT: APFS does NOT provide checksums for file data content.
// According to the APFS specification, file data integrity relies on hardware ECC,
// not filesystem-level checksums. APFS only checksums metadata structures.
//
// This method verifies:
//  1. The B-tree node containing the inode has a valid Fletcher-64 checksum (verified in NewBTreeNodeReader)
//  2. The inode structure can be successfully parsed
//  3. All metadata structures are intact
//
// File data integrity cannot be verified at the filesystem level in APFS.
//
// Fletcher-64 verification was added to btrees/btree_node_reader.go:NewBTreeNodeReader()
// All B-tree nodes are now automatically verified when loaded.
func (fs *FileSystemServiceImpl) VerifyFileChecksum(inodeID uint64) (bool, error) {
	// Load inode data - this automatically verifies B-tree node checksums
	// via the Fletcher-64 verification in NewBTreeNodeReader
	inodeData, err := fs.loadInodeData(types.OidT(inodeID))
	if err != nil {
		return false, fmt.Errorf("failed to load inode metadata: %w", err)
	}

	// Parse inode to verify structure integrity
	inodeReader, err := file_system_objects.NewInodeReader(inodeData.key, inodeData.value, binary.LittleEndian)
	if err != nil {
		return false, fmt.Errorf("failed to parse inode structure: %w", err)
	}

	// Verify the inode has valid basic properties
	objID := inodeReader.ObjectIdentifier()
	if objID == 0 {
		return false, fmt.Errorf("inode has invalid object identifier")
	}

	// Verify the object type is correct
	if inodeReader.ObjectType() != types.ApfsTypeInode {
		return false, fmt.Errorf("object is not an inode (type: %v)", inodeReader.ObjectType())
	}

	// If we successfully loaded and parsed the inode, the metadata is valid
	// The B-tree infrastructure already verified Fletcher-64 checksums
	return true, nil
}
