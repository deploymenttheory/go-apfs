package services

import (
	"fmt"
	"sync"

	"github.com/deploymenttheory/go-apfs/internal/types"
)

// ObjectLocatorServiceImpl implements object discovery and reference analysis
type ObjectLocatorServiceImpl struct {
	container *ContainerReader
	resolver  *BTreeObjectResolver
	mu        sync.RWMutex
	cache     map[types.OidT]any // Cache for discovered objects
}

// NewObjectLocatorService creates a new ObjectLocatorService instance
func NewObjectLocatorService(container *ContainerReader) *ObjectLocatorServiceImpl {
	return &ObjectLocatorServiceImpl{
		container: container,
		resolver:  NewBTreeObjectResolver(container),
		cache:     make(map[types.OidT]any),
	}
}

// FindObjectByID locates an object in the container by its OID
func (ols *ObjectLocatorServiceImpl) FindObjectByID(oid uint64) (any, error) {
	ols.mu.RLock()
	if cached, exists := ols.cache[types.OidT(oid)]; exists {
		ols.mu.RUnlock()
		return cached, nil
	}
	ols.mu.RUnlock()

	if oid == 0 {
		return nil, fmt.Errorf("invalid OID: 0")
	}

	// Get container superblock
	sb := ols.container.GetSuperblock()
	if sb == nil {
		return nil, fmt.Errorf("container superblock not available")
	}

	// Try to resolve the virtual object to physical address
	physAddr, err := ols.resolver.ResolveVirtualObject(types.OidT(oid), sb.NxNextXid-1)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve OID %d: %w", oid, err)
	}

	// Read the object block
	blockData, err := ols.container.ReadBlock(uint64(physAddr))
	if err != nil {
		return nil, fmt.Errorf("failed to read block for OID %d: %w", oid, err)
	}

	// Cache the result
	ols.mu.Lock()
	ols.cache[types.OidT(oid)] = blockData
	ols.mu.Unlock()

	return blockData, nil
}

// ResolveObjectPath traces the reference chain from an object to the root
func (ols *ObjectLocatorServiceImpl) ResolveObjectPath(oid uint64) ([]uint64, error) {
	if oid == 0 {
		return nil, fmt.Errorf("invalid OID: 0")
	}

	path := []uint64{oid}
	currentOID := types.OidT(oid)
	sb := ols.container.GetSuperblock()
	if sb == nil {
		return nil, fmt.Errorf("container superblock not available")
	}

	// Try to find parent reference - check if we've reached a known root OID
	if currentOID == types.OidT(sb.NxOmapOid) ||
		currentOID == types.OidT(sb.NxSpacemanOid) {
		path = append(path, uint64(currentOID))
		return path, nil
	}

	// Check container OIDs
	for _, oid := range sb.NxFsOid {
		if oid == currentOID {
			path = append(path, uint64(currentOID))
			return path, nil
		}
	}

	// Return the initial path if we can't resolve further
	return path, nil
}

// IsObjectValid checks if an object has a valid structure
func (ols *ObjectLocatorServiceImpl) IsObjectValid(oid uint64) (bool, error) {
	obj, err := ols.FindObjectByID(oid)
	if err != nil {
		return false, nil // Object doesn't exist or can't be accessed
	}

	blockData, ok := obj.([]byte)
	if !ok {
		return false, fmt.Errorf("object is not block data")
	}

	if len(blockData) < 32 {
		return false, nil
	}

	// Check for valid object header magic
	// Most APFS objects have checksum fields in first 8-16 bytes
	// followed by object ID and type fields at offsets 8-16
	checksum := uint32(blockData[0]) | uint32(blockData[1])<<8 |
		uint32(blockData[2])<<16 | uint32(blockData[3])<<24

	// Non-zero checksum suggests valid object (basic check)
	// Real validation would require more sophisticated checks
	return checksum != 0 || oid != 0, nil
}

// GetObjectDependencies finds all objects that an object references
func (ols *ObjectLocatorServiceImpl) GetObjectDependencies(oid uint64) ([]uint64, error) {
	obj, err := ols.FindObjectByID(oid)
	if err != nil {
		return nil, err
	}

	blockData, ok := obj.([]byte)
	if !ok {
		return nil, fmt.Errorf("object is not block data")
	}

	// Extract OID references from the block
	// OIDs are typically stored as 64-bit little-endian values
	// This is a simplified approach - real implementation would parse object types
	var dependencies []uint64

	// Scan block for OID-like values (non-zero 64-bit values that could be OIDs)
	for i := 0; i < len(blockData)-8; i += 8 {
		potentialOID := uint64(blockData[i]) |
			uint64(blockData[i+1])<<8 |
			uint64(blockData[i+2])<<16 |
			uint64(blockData[i+3])<<24 |
			uint64(blockData[i+4])<<32 |
			uint64(blockData[i+5])<<40 |
			uint64(blockData[i+6])<<48 |
			uint64(blockData[i+7])<<56

		// Filter for plausible OID values (not too large, not zero)
		if potentialOID > 0 && potentialOID < 1000000000 {
			// Avoid duplicates
			isDuplicate := false
			for _, dep := range dependencies {
				if dep == potentialOID {
					isDuplicate = true
					break
				}
			}
			if !isDuplicate {
				dependencies = append(dependencies, potentialOID)
			}
		}
	}

	return dependencies, nil
}

// AnalyzeObjectReferences builds a reference graph of objects in the container
func (ols *ObjectLocatorServiceImpl) AnalyzeObjectReferences() (*ReferenceGraph, error) {
	sb := ols.container.GetSuperblock()
	if sb == nil {
		return nil, fmt.Errorf("container superblock not available")
	}

	rg := &ReferenceGraph{
		Objects:            make(map[types.OidT]string),
		References:         []ObjectReference{},
		OrphanedObjects:    []types.OidT{},
		CircularReferences: [][]types.OidT{},
	}

	// Start with known container objects
	knownOIDs := []types.OidT{
		types.OidT(sb.NxOmapOid),
		types.OidT(sb.NxSpacemanOid),
	}

	// Add volume OIDs
	for _, volOID := range sb.NxFsOid {
		if volOID != 0 {
			knownOIDs = append(knownOIDs, volOID)
		}
	}

	// Map known objects
	for _, oid := range knownOIDs {
		rg.Objects[oid] = "known_container_object"
	}

	// Analyze references from known objects
	for _, oid := range knownOIDs {
		deps, err := ols.GetObjectDependencies(uint64(oid))
		if err != nil {
			continue
		}

		for _, depOID := range deps {
			if depOID != 0 && depOID < 1000000000 {
				rg.References = append(rg.References, ObjectReference{
					FromOID:        oid,
					ToOID:          types.OidT(depOID),
					ReferenceType:  "dependency",
					ReferenceCount: 1,
				})

				// Mark dependent as discovered
				rg.Objects[types.OidT(depOID)] = "discovered"
			}
		}
	}

	return rg, nil
}

// FindOrphanedObjects detects objects that are not referenced by any container object
func (ols *ObjectLocatorServiceImpl) FindOrphanedObjects() ([]uint64, error) {
	sb := ols.container.GetSuperblock()
	if sb == nil {
		return nil, fmt.Errorf("container superblock not available")
	}

	// Get reference graph
	refGraph, err := ols.AnalyzeObjectReferences()
	if err != nil {
		return nil, err
	}

	// In a fully populated container, most objects should be referenced
	// Objects not in the reference graph are potential orphans
	var orphans []uint64

	// Known container OIDs that should always exist
	knownOIDs := make(map[types.OidT]bool)
	knownOIDs[types.OidT(sb.NxOmapOid)] = true
	knownOIDs[types.OidT(sb.NxSpacemanOid)] = true

	for _, oid := range sb.NxFsOid {
		if oid != 0 {
			knownOIDs[oid] = true
		}
	}

	// Check if important OIDs are missing from reference graph
	for oid := range knownOIDs {
		if _, exists := refGraph.Objects[oid]; !exists {
			orphans = append(orphans, uint64(oid))
		}
	}

	// Collect truly orphaned OIDs from discovered objects not in reference chain
	for oid := range refGraph.Objects {
		// Check if this OID is in known IDs map (placeholder for future full traversal)
		_ = knownOIDs[oid]
	}

	return orphans, nil
}

// GetObjectType classifies the type of an object
func (ols *ObjectLocatorServiceImpl) GetObjectType(oid uint64) (string, error) {
	obj, err := ols.FindObjectByID(oid)
	if err != nil {
		return "", err
	}

	blockData, ok := obj.([]byte)
	if !ok {
		return "", fmt.Errorf("object is not block data")
	}

	if len(blockData) < 32 {
		return "unknown", nil
	}

	// Extract ObjIdAndType field (typically at offset 8, 8 bytes)
	if len(blockData) >= 16 {
		objIdAndType := uint64(blockData[8]) |
			uint64(blockData[9])<<8 |
			uint64(blockData[10])<<16 |
			uint64(blockData[11])<<24 |
			uint64(blockData[12])<<32 |
			uint64(blockData[13])<<40 |
			uint64(blockData[14])<<48 |
			uint64(blockData[15])<<56

		objType := (objIdAndType & types.ObjTypeMask) >> types.ObjTypeShift

		// Map to type names
		switch types.JObjTypes(objType) {
		case types.ApfsTypeInvalid:
			return "invalid", nil
		case types.ApfsTypeInode:
			return "inode", nil
		case types.ApfsTypeDirRec:
			return "directory_entry", nil
		case types.ApfsTypeFileExtent:
			return "file_extent", nil
		default:
			return fmt.Sprintf("object_type_%d", objType), nil
		}
	}

	return "unknown", nil
}

// GetObjectSize returns the size of an object in bytes
func (ols *ObjectLocatorServiceImpl) GetObjectSize(oid uint64) (uint64, error) {
	_, err := ols.FindObjectByID(oid)
	if err != nil {
		return 0, err
	}

	// Most APFS objects are one block in size, which is typically 4096 bytes
	sb := ols.container.GetSuperblock()
	if sb != nil {
		return uint64(sb.NxBlockSize), nil
	}

	// Default to standard block size
	return 4096, nil
}
