package services

import (
	"encoding/binary"
	"fmt"
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

// GetFileMetadata returns metadata for a file at the given path
func (fs *FileSystemServiceImpl) GetFileMetadata(path string) (*FileEntry, error) {
	path = filepath.Clean(path)
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	inode, err := fs.getInodeByPath(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get inode: %w", err)
	}

	inodeData, err := fs.loadInodeData(inode)
	if err != nil {
		return nil, fmt.Errorf("failed to load inode data: %w", err)
	}

	inodeReader, err := file_system_objects.NewInodeReader(inodeData.key, inodeData.value, binary.LittleEndian)
	if err != nil {
		return nil, fmt.Errorf("failed to parse inode: %w", err)
	}

	return &FileEntry{
		Inode:    uint64(inode),
		Name:     filepath.Base(path),
		Path:     path,
		IsDir:    inodeReader.IsDirectory(),
		Size:     0, // Size would need to be calculated from file extents
		Mode:     uint16(inodeReader.Mode()),
		Modified: uint64(inodeReader.ModificationTime().Unix()),
	}, nil
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
			
			keyOffset := binary.LittleEndian.Uint16(nodeData[offset:offset+2])
			valueOffset := binary.LittleEndian.Uint16(nodeData[offset+2:offset+4])
			
			// Extract and parse the key
			keyStart := btnDataStart + int(keyOffset)
			if keyStart+8 <= len(nodeData) {
				objIdAndType := binary.LittleEndian.Uint64(nodeData[keyStart:keyStart+8])
				objID := objIdAndType & types.ObjIdMask
				
				// If this key's object ID is >= our target, use this child
				if objID >= uint64(targetOID) {
					valueStart := btnDataStart + int(valueOffset)
					if valueStart+8 <= len(nodeData) {
						targetChildOID = types.OidT(binary.LittleEndian.Uint64(nodeData[valueStart:valueStart+8]))
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
			
			keyOffset := binary.LittleEndian.Uint16(nodeData[offset:offset+2])
			valueOffset := binary.LittleEndian.Uint16(nodeData[offset+2:offset+4])
			
			// Extract key and check if it matches our target inode
			keyStart := btnDataStart + int(keyOffset)
			if keyStart+8 <= len(nodeData) {
				objIdAndType := binary.LittleEndian.Uint64(nodeData[keyStart:keyStart+8])
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
						keyData = nodeData[keyStart:keyStart+keySize]
					}
					
					// Value size is variable, read what we can
					if valueStart < len(nodeData) {
						// Find the next entry to determine value size, or use remaining data
						var valueSize int
						if i+1 < keyCount {
							// Get next entry's offset to calculate this value's size
							nextOffset := tableOffset + int(i+1)*entrySize
							if nextOffset+2 <= len(nodeData) {
								nextKeyOffset := binary.LittleEndian.Uint16(nodeData[nextOffset:nextOffset+2])
								nextKeyStart := btnDataStart + int(nextKeyOffset)
								valueSize = nextKeyStart - valueStart
							}
						} else {
							// Last entry, use remaining data
							valueSize = len(nodeData) - valueStart
						}
						
						if valueSize > 0 && valueStart+valueSize <= len(nodeData) {
							valueData = nodeData[valueStart:valueStart+valueSize]
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
	entry, err := fs.GetFileMetadata(path)
	if err != nil {
		return false, err
	}
	return entry.IsDir, nil
}

// Exists checks if a path exists
func (fs *FileSystemServiceImpl) Exists(path string) (bool, error) {
	_, err := fs.GetFileMetadata(path)
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
			
			keyOffset := binary.LittleEndian.Uint16(nodeData[offset:offset+2])
			valueOffset := binary.LittleEndian.Uint16(nodeData[offset+2:offset+4])
			
			// Extract and parse the key
			keyStart := btnDataStart + int(keyOffset)
			if keyStart+16 <= len(nodeData) {
				// Parse the key as JKeyT to check object ID and type
				objIdAndType := binary.LittleEndian.Uint64(nodeData[keyStart:keyStart+8])
				objID := objIdAndType & types.ObjIdMask
				
				// If this key's object ID is >= our target, use this child
				if objID >= privateID {
					// Extract child OID from value
					valueStart := btnDataStart + int(valueOffset)
					if valueStart+8 <= len(nodeData) {
						targetChildOID = types.OidT(binary.LittleEndian.Uint64(nodeData[valueStart:valueStart+8]))
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
			
			keyOffset := binary.LittleEndian.Uint16(nodeData[offset:offset+2])
			valueOffset := binary.LittleEndian.Uint16(nodeData[offset+2:offset+4])
			
			// Extract key and check if it's a file extent record
			keyStart := btnDataStart + int(keyOffset)
			if keyStart+16 <= len(nodeData) {
				// Parse JFileExtentKeyT
				objIdAndType := binary.LittleEndian.Uint64(nodeData[keyStart:keyStart+8])
				objID := objIdAndType & types.ObjIdMask
				objType := (objIdAndType & types.ObjTypeMask) >> types.ObjTypeShift
				
				// Check if this is a file extent record for our private ID
				if objID == privateID && objType == uint64(types.ApfsTypeFileExtent) {
					// Extract logical address from key
					logicalAddr := binary.LittleEndian.Uint64(nodeData[keyStart+8:keyStart+16])
					
					// Extract value (JFileExtentValT)
					valueStart := btnDataStart + int(valueOffset)
					if valueStart+24 <= len(nodeData) {
						lenAndFlags := binary.LittleEndian.Uint64(nodeData[valueStart:valueStart+8])
						physBlockNum := binary.LittleEndian.Uint64(nodeData[valueStart+8:valueStart+16])
						
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
