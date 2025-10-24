# APFS Services Implementation Plan

## Current Status

### Implemented Services (2)
1. **ContainerReader** - Low-level block I/O with caching
2. **VolumeService** - Partial implementation (metadata, space stats)

### Service Interfaces Defined (6)
1. FileSystemService - File/directory traversal
2. ObjectLocatorService - Object discovery
3. SnapshotService - Snapshot management
4. VolumeService - Volume operations (partially implemented)
5. DataRecoveryService - Deleted file recovery
6. EncryptionService - Encryption analysis

## Available Parser Infrastructure

### Core Parsing Layers
```
Parsers Available:
├── Container (5 readers)
│   ├── ContainerSuperblockReader
│   ├── CheckpointMapReader
│   ├── SpaceManagerReader
│   └── EvictMappingReader
│
├── File System Objects (5 readers)
│   ├── InodeReader (file/dir metadata)
│   ├── DirectoryEntryReader (directory contents)
│   ├── DirectoryStatsReader
│   ├── ExtendedAttributeReader
│   └── JKeyReader
│
├── Volumes (25 readers)
│   ├── VolumeSuperblockReader
│   ├── VolumeSpaceManagement
│   ├── VolumeSnapshotMetadata
│   ├── VolumeIdentity
│   ├── VolumeEncryptionMetadata
│   └── More metadata readers
│
├── Snapshots (6 readers)
│   ├── SnapMetadataReader
│   └── Snapshot traversal
│
├── Encryption (3 readers)
│   ├── CryptoStateReader
│   ├── KeybagReader
│   └── MediaKeybagReader
│
├── B-Trees (6 readers)
│   ├── BTreeNodeReader
│   ├── BTreeInfoReader
│   └── Key-value extraction
│
├── Data Streams (5 readers)
│   ├── FileExtentReader
│   ├── PhysicalExtentReader
│   ├── DataStreamReader
│   └── Extended attribute streams
│
├── Extended Fields (4 readers)
│   ├── ExtendedFieldReader
│   └── Field type resolution
│
├── Space Manager (10 readers)
│   ├── SpaceManagerReader (comprehensive)
│   ├── CibAddrBlockReader (allocation tracking)
│   ├── ChunkInfoReader
│   └── Free queue management
│
└── More specialized parsers
    ├── EFI Jumpstart (10 readers)
    ├── Encryption Rolling (3 readers)
    ├── Reaper (5 readers)
    ├── Sealed Volumes (3 readers)
    └── Siblings (2 readers)
```

### Key Types Available
```
types/ package provides:
- JInodeKeyT / JInodeValT (inode metadata)
- JDirentT (directory entries)
- SpacemanPhysT (space allocation)
- JSnapMetadataValT (snapshot metadata)
- JCryptoKeyT / JCryptoValT (encryption)
- ApfsSuperblockT (volume superblock)
- NxSuperblockT (container superblock)
- And 20+ other APFS structure types
```

## Recommended Implementation Priority

### Phase 1: Foundation (High Value, Medium Effort)
**1. SnapshotService** ⭐ 
- **Why**: Snapshot parsers already exist; high utility for versioning
- **Depends on**: VolumeService, BTreeObjectResolver
- **Methods to implement**:
  - `ListAllSnapshots()` - Read from volume superblock snapshot metadata
  - `GetSnapshotMetadata(xid)` - Parse snapshot metadata
  - `GetSnapshotSize(xid)` - Calculate from extents
  - `FindSnapshotByName()` - Search by name
  - `GetSnapshotFileCount()` - Count inodes in snapshot
  - `CompareSnapshots()` - Diff two snapshots (advanced)
  - `GetChangedFiles()` - List changes between snapshots (advanced)

**2. EncryptionService** ⭐
- **Why**: Crypto parsers available; essential for security analysis
- **Depends on**: VolumeService, CryptoStateReader
- **Methods to implement**:
  - `GetEncryptionStatus()` - Read from volume superblock
  - `VerifyEncryptionConsistency()` - Check keybag integrity
  - `CheckProtectionClasses()` - Parse protection metadata
  - `ValidateEncryptionMetadata()` - Validate structures
  - `IsFileEncrypted(inode)` - Check file flags
  - `GetEncryptionKeys()` - List available keys (count only)

### Phase 2: Core Filesystem (High Value, High Effort)
**3. FileSystemService** ⭐⭐
- **Why**: Central to all filesystem operations
- **Depends on**: BTreeObjectResolver, InodeReader, DirectoryEntryReader
- **Challenge**: Must traverse B-tree for directory listings
- **Methods to implement**:
  - `GetFileMetadata(inodeID)` - Parse inode via resolver
  - `ListDirectoryContents(inodeID)` - Parse directory entries from inode
  - `GetFileExtents(inodeID)` - Extract extent info
  - `GetInodeByPath(path)` - Path traversal (recursive directory lookup)
  - `FindFilesByName(pattern)` - Scan filesystem tree
  - `GetParentDirectory(inodeID)` - Use parent_id from inode
  - `IsPathAccessible(path)` - Check path existence

**4. ObjectLocatorService** ⭐⭐
- **Why**: Object dependency analysis for integrity checking
- **Depends on**: BTreeObjectResolver, ObjectMapReader
- **Methods to implement**:
  - `FindObjectByID(oid)` - Use B-tree object resolver
  - `GetObjectType(oid)` - Parse object header type
  - `IsObjectValid(oid)` - Verify checksum using ChecksumVerifier
  - `ResolveObjectPath(oid)` - Build reference chain
  - `GetObjectDependencies(oid)` - Trace references
  - `AnalyzeObjectReferences()` - Build reference graph
  - `FindOrphanedObjects()` - Identify unreferenced objects

### Phase 3: Data Recovery (Medium Value, High Effort)
**5. DataRecoveryService** ⭐
- **Why**: Recovery potential analysis for unlinked files
- **Depends on**: FileSystemService, SpaceManager, Reaper
- **Challenge**: Must scan for deleted inodes and extents
- **Methods to implement**:
  - `FindUnlinkedInodes()` - Scan space manager free space
  - `EstimateRecoveryPotential()` - Analyze recoverable data
  - `FindOrphanedExtents()` - Find unlinked extents
  - `ScanForDeletedFiles(pattern)` - Pattern matching on deleted files
  - `BuildRecoveryMap()` - Map recoverable files
  - `GetRecoverableFileCount()` - Count recoverable files
  - `GetRecoverableDataSize()` - Size estimate

---

## Architecture Decisions

### Design Pattern
- **Service pattern**: Each service wraps parsers into high-level APIs
- **Dependency injection**: Services depend on ContainerReader and BTreeObjectResolver
- **Error handling**: Graceful degradation for corrupted/missing data
- **Caching**: Leverage ContainerReader block cache

### Key Dependencies
```
ContainerReader
    ├── BTreeObjectResolver (virtual object resolution)
    │   ├── SnapshotService
    │   ├── EncryptionService  
    │   ├── FileSystemService
    │   ├── ObjectLocatorService
    │   └── DataRecoveryService
    │
    └── VolumeService (existing - needs completion)
        ├── SpaceManagerReader
        ├── VolumeSuperblockReader
        └── VolumeSpaceManagement
```

### Implementation Characteristics

#### SnapshotService - Straightforward
- Snapshot metadata already parsed in volumes package
- Limited B-tree traversal required
- Mostly data extraction and filtering
- **Estimated LOC**: 300-400

#### EncryptionService - Straightforward
- Crypto parsers fully implemented
- Validation logic already exists
- Limited state management
- **Estimated LOC**: 250-350

#### FileSystemService - Complex
- Requires recursive directory traversal
- B-tree navigation for each directory
- Path parsing and validation
- Pattern matching for file search
- **Estimated LOC**: 800-1200

#### ObjectLocatorService - Complex
- Graph building from object references
- Checksum verification for all objects
- Orphan detection algorithm
- **Estimated LOC**: 600-900

#### DataRecoveryService - Very Complex
- Extent scanning algorithms
- Deleted inode pattern matching
- Recovery confidence scoring
- Space bitmap analysis
- **Estimated LOC**: 1000-1500

---

## Testing Strategy

### Test Data Requirements
Each service needs:
1. Valid APFS containers (different configurations)
2. Snapshots for snapshot service
3. Encrypted volumes for encryption service
4. Complex directory trees for filesystem service
5. Corrupted/partial data for recovery service

### Test Approach
```
Unit Tests:
├── Parser integration tests
├── Error handling tests
├── Edge case tests
└── Mock object tests

Integration Tests:
├── Multi-service workflows
├── Large dataset handling
└── Performance benchmarks
```

---

## Implementation Order Recommendation

```
Week 1:  SnapshotService + EncryptionService (quick wins)
         └─> Build confidence with simpler services

Week 2:  FileSystemService (core capability)
         └─> Essential foundation for other services

Week 3:  ObjectLocatorService (validation tooling)
         └─> Build reference graph and corruption detection

Week 4:  DataRecoveryService (advanced feature)
         └─> Most complex, but builds on previous services
```

---

## Next Steps

1. **Complete VolumeService** - Add missing methods (corruption detection, fragmentation)
2. **Implement SnapshotService** - Start with simpler service
3. **Build test data generator** - Create containers with snapshots/encryption
4. **Implement FileSystemService** - Build directory traversal
5. **Add EncryptionService** - Leverage existing parsers
6. **Build DataRecoveryService** - Most complex, iterative development

---

## Notes

- BTreeObjectResolver is **critical** to all services - handles virtual object resolution
- ContainerReader caching is essential for performance
- Services should follow **single responsibility principle** - each handles one domain
- Tests should use real APFS data or high-fidelity mocks
- Documentation should include common APFS quirks (e.g., snapshots, encryption rolling)
