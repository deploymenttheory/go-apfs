# File System Objects Parsers

This package contains low-level parsers for APFS file system objects. These parsers are responsible for reading and interpreting raw binary data from APFS B-tree records into structured Go types.

## Overview

The parsers in this package implement the APFS file system object specifications as defined in the Apple File System Reference (pages 71-101). Each parser handles a specific type of file system object record.

## Parsers

### 1. J-Key Reader (`j_key_reader.go`)
**Purpose**: Parses the common header used by all file system object keys.

**Key Features**:
- Extracts object identifier (60-bit) and object type (4-bit) from combined field
- Supports both little-endian and big-endian byte orders
- Handles system object IDs with proper bit masking

**Methods**:
- `ObjectIdentifier() uint64` - Returns the object's unique identifier
- `ObjectType() types.JObjTypes` - Returns the file system object type
- `RawObjIdAndType() uint64` - Returns the raw combined field

### 2. Inode Reader (`inode_reader.go`)
**Purpose**: Parses inode records (APFS_TYPE_INODE) containing file and directory metadata.

**Key Features**:
- Parses all inode fields including timestamps, ownership, permissions
- Handles union field (nchildren for directories, nlink for files)
- Supports extended fields parsing
- Implements `InodeReader` interface

**Methods**:
- `ParentID() uint64` - Parent directory identifier
- `PrivateID() uint64` - Data stream identifier
- `CreationTime() time.Time` - File creation timestamp
- `ModificationTime() time.Time` - Last modification timestamp
- `ChangeTime() time.Time` - Attribute change timestamp
- `AccessTime() time.Time` - Last access timestamp
- `Owner() types.UidT` - File owner user ID
- `Group() types.GidT` - File group ID
- `Mode() types.ModeT` - File permissions and type
- `IsDirectory() bool` - Check if inode represents a directory
- `NumberOfChildren() int32` - Number of directory entries (directories only)
- `NumberOfHardLinks() int32` - Number of hard links (files only)

### 3. Directory Entry Reader (`directory_entry_reader.go`)
**Purpose**: Parses directory entry records (APFS_TYPE_DIR_REC) that map filenames to inode IDs.

**Key Features**:
- Supports both regular and hashed directory keys
- Handles null-terminated UTF-8 filenames
- Parses directory entry flags and timestamps
- Implements `DirectoryEntryReader` interface

**Methods**:
- `FileName() string` - Directory entry filename (null-terminated strings handled)
- `FileID() uint64` - Target inode identifier
- `DateAdded() time.Time` - When entry was added to directory
- `FileType() uint16` - File type flags

### 4. Directory Statistics Reader (`directory_stats_reader.go`)
**Purpose**: Parses directory statistics records (APFS_TYPE_DIR_STATS) containing directory metadata.

**Key Features**:
- Tracks directory size and child count statistics
- Maintains generation counters for change tracking
- Links to parent directory via chained key
- Implements `DirectoryStatsReader` interface

**Methods**:
- `NumChildren() uint64` - Number of files and folders in directory
- `TotalSize() uint64` - Total size of all files in directory tree
- `ChainedKey() uint64` - Parent directory's object identifier
- `GenCount() uint64` - Generation counter for change tracking

### 5. Extended Attribute Reader (`extended_attribute_reader.go`)
**Purpose**: Parses extended attribute records (APFS_TYPE_XATTR) containing file metadata.

**Key Features**:
- Handles both embedded data and data stream references
- Supports file system owned attributes
- Parses attribute names and flags
- Implements `ExtendedAttributeReader` interface

**Methods**:
- `AttributeName() string` - Extended attribute name
- `IsDataEmbedded() bool` - Check if data is embedded in record
- `IsDataStream() bool` - Check if data is stored in separate stream
- `IsFileSystemOwned() bool` - Check if attribute is system-owned
- `Data() []byte` - Attribute data or stream identifier

### 6. File System Object Type Resolver (`file_system_object_type_resolver.go`)
**Purpose**: Provides human-readable descriptions for object types and kinds.

**Key Features**:
- Maps numeric object types to descriptive strings
- Lists all supported object types
- Implements `FileSystemObjectTypeResolver` interface

**Methods**:
- `ResolveObjectType(types.JObjTypes) string` - Convert type to description
- `ResolveObjectKind(types.JObjKinds) string` - Convert kind to description
- `ListSupportedObjectTypes() []types.JObjTypes` - List all supported types

## Design Principles

### SSOT (Single Source of Truth)
- Each parser is the authoritative source for its specific record type
- Type definitions are centralized in the `types` package
- Interface definitions are centralized in the `interfaces` package

### YAGNI (You Aren't Gonna Need It)
- Only implements methods required by the interfaces
- No speculative features or unused functionality
- Focused on core parsing requirements

### KISS (Keep It Simple, Stupid)
- Straightforward parsing logic without unnecessary complexity
- Clear separation of concerns between parsers
- Simple error handling with descriptive messages

### DRY (Don't Repeat Yourself)
- Common parsing patterns are reused across parsers
- Shared test helper functions for data generation
- Consistent error handling patterns

### SOLID Principles
- **Single Responsibility**: Each parser handles one record type
- **Open/Closed**: Extensible through interfaces, closed for modification
- **Liskov Substitution**: All parsers implement their interfaces correctly
- **Interface Segregation**: Focused interfaces with specific responsibilities
- **Dependency Inversion**: Depends on interfaces, not concrete types

### Fail-Fast Principle
- Input validation at parser creation time
- Immediate error reporting for malformed data
- No silent failures or data corruption

## Error Handling

All parsers implement robust error handling:
- **Input Validation**: Check data size requirements before parsing
- **Bounds Checking**: Verify field lengths don't exceed available data
- **Type Safety**: Ensure proper type conversions and bit masking
- **Descriptive Errors**: Clear error messages indicating the specific failure

## Testing

Comprehensive test coverage includes:
- **Happy Path Testing**: Valid data parsing scenarios
- **Error Case Testing**: Invalid input handling
- **Edge Case Testing**: Boundary conditions and special values
- **Interface Compliance**: Verification of interface implementations
- **Data Integrity**: Bit masking and endianness handling

### Test Coverage
- Directory Entry Reader: 100% coverage with hashed/non-hashed variants
- Directory Stats Reader: 100% coverage with various data sizes
- Extended Attribute Reader: 100% coverage with embedded/stream data
- File System Object Type Resolver: 100% coverage of all types/kinds
- Inode Reader: 100% coverage with directory/file variants
- J-Key Reader: 100% coverage with endianness and bit masking tests

## Usage Example

```go
// Parse a directory entry record
keyData := []byte{...}    // Raw key data from B-tree
valueData := []byte{...}  // Raw value data from B-tree

reader, err := NewDirectoryEntryReader(keyData, valueData, binary.LittleEndian, false)
if err != nil {
    return fmt.Errorf("failed to parse directory entry: %w", err)
}

// Access parsed data
filename := reader.FileName()
fileID := reader.FileID()
objectType := reader.ObjectType()
```

## Dependencies

- `github.com/deploymenttheory/go-apfs/internal/types` - APFS type definitions
- `github.com/deploymenttheory/go-apfs/internal/interfaces` - Parser interfaces
- `encoding/binary` - Binary data parsing
- `time` - Timestamp handling
- `strings` - String manipulation

## Compliance

These parsers implement the APFS specification as documented in:
- Apple File System Reference, pages 71-101
- File-System Objects section
- B-tree record format specifications

All parsers handle:
- Little-endian and big-endian byte orders
- Variable-length fields and extended data
- Null-terminated UTF-8 strings
- Bit field extraction and masking
- Timestamp conversion (nanoseconds since Unix epoch) 