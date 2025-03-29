# APFS Project Architecture

## Overview

This architecture follows the layered design of Apple File System as described in the reference documentation, with a clear separation between the container layer and file-system layer.

### Refined Architecture

```
apfs/
├── cmd/                     # Command-line interfaces
│   ├── apfs-info/           # Tool to display APFS container/volume info
│   ├── apfs-mount/          # Tool to mount APFS volumes
│   └── apfs-recover/        # Data recovery tool
├── internal/                # Non-exported internal packages
│   └── binary/              # Binary parsing utilities
└── pkg/                     # Exported package code
    ├── checksum/            # Checksum implementation
    │   └── fletcher64.go    # Fletcher64 checksum algorithm
    ├── types/               # Core types and constants
    │   ├── constants.go     # All APFS constants from the spec
    │   ├── container_types.go # Container layer data structures
    │   ├── fs_types.go      # File system layer data structures
    │   ├── interfaces.go    # Core interfaces (BlockDevice, etc.)
    │   └── common.go        # Common types like UUID, PAddr, etc.
    ├── container/           # Container layer
    │   ├── object.go        # Object handling and common operations
    │   ├── superblock.go    # Container superblock (nx_superblock_t)
    │   ├── checkpoint.go    # Checkpoint management
    │   ├── omap.go          # Object map operations
    │   ├── btree.go         # B-tree operations (search, insert, etc.)
    │   ├── btnode.go        # B-tree node implementation details
    │   ├── spaceman.go      # Space manager implementation
    │   ├── allocation.go    # Block allocation routines
    │   ├── reaper.go        # Reaper implementation
    │   └── encryption.go    # Container-level encryption (keybags)
    ├── fs/                  # File system layer
    │   ├── volume.go        # Volume superblock (apfs_superblock_t)
    │   ├── tree.go          # File system tree operations
    │   ├── inode.go         # Inode operations
    │   ├── dentry.go        # Directory entry operations
    │   ├── extattr.go       # Extended attributes handling
    │   ├── datastream.go    # Data stream management
    │   ├── extent.go        # File extent handling
    │   ├── sibling.go       # Hard link management
    │   └── crypto.go        # File-level encryption
    ├── io/                  # I/O package
    │   ├── blockdevice.go   # Block device implementations
    │   ├── cache.go         # Caching layer for block reads
    │   └── transaction.go   # Transaction handling and journaling
    ├── crypto/              # Encryption support
    │   ├── keybag.go        # Keybag structures and operations
    │   ├── key.go           # Key management (KEK/VEK)
    │   ├── aes.go           # AES-XTS implementation
    │   └── wrappers.go      # Key wrapping utilities
    ├── snapshot/            # Snapshot management
    │   ├── metadata.go      # Snapshot metadata handling
    │   └── operations.go    # Snapshot creation/deletion/mounting
    └── seal/                # Sealed volume support
        ├── integrity.go     # Integrity metadata
        └── hash.go          # Hash algorithm implementations
```

### key decisions

1. **Layer-Specific Type Files**: Split types into `container_types.go` and `fs_types.go` to better reflect the layered architecture described in the spec.

2. **B-Tree Structure**: Added `btnode.go` to handle the complex B-tree node structure described in pages 122-133 of the spec, separate from the higher-level operations.

3. **Sealed Volumes**: Added a `seal` package to handle the sealed volumes functionality described on pages 150-158 of the spec.

4. **I/O Package**: Separated I/O operations (including transaction handling) into their own package, which better reflects the different roles described in the spec.

5. **Allocation Operations**: Added dedicated `allocation.go` for the space allocation functions, which are complex enough to warrant their own file based on pages 159-163.

6. **Crypto Package Structure**: Reorganized the crypto package to better match the structures and operations described in pages 135-149 of the spec.

### Implementation Considerations

1. **Object Storage Methods**: The spec (p.10) describes three storage methods for objects: ephemeral, physical, and virtual. Your `object.go` should handle these distinctions explicitly.

2. **Checkpoints**: The checkpoint mechanism (p.26-27) is central to crash protection, so `checkpoint.go` should implement both reading and writing of checkpoints.

3. **Copy-on-Write**: Throughout the spec, it emphasizes that "objects on disk are never modified in place" (p.7). Ensure your implementation maintains this principle, especially in the transaction and B-tree code.

4. **Object Map Design**: The spec (p.44-49) details how object maps use B-trees to map from virtual object identifiers to physical addresses. Your `omap.go` should implement this lookup mechanism carefully.

5. **B-Tree Complex Layout**: The B-tree implementation (p.122-133) has unique characteristics including separate storage areas for keys and values growing from opposite ends. Ensure `btnode.go` captures these details.

This refined architecture closely follows the Apple File System specification while maintaining a clean, logical structure that will be maintainable and extensible as you implement more advanced features.

Looking at the APFS reference document and your Go project architecture, I'd recommend implementing the system in the following logical order:

1. **Core Types and Constants** (types/constants.go, types/types.go)
   - Define fundamental types like OID, XID, PAddr, etc.
   - Set up essential constants and error types
   - Implement basic interfaces (BlockDevice, Object, etc.)

2. **IO and Utilities** (util/io.go, util/checksum.go)
   - Implement the block device interface for reading/writing blocks
   - Build Fletcher64 checksum implementation
   - Add UUID handling utilities

3. **Container Layer Core**
   - Object handling (container/object.go)
   - Container superblock parsing (container/container.go)
   - Checkpoint mechanism (container/checkpoint.go)

4. **Basic Reading Infrastructure**
   - Object map implementation (container/omap.go)
   - B-tree structures (container/btree.go)
   - Space manager basics (container/spaceman.go)

5. **File System Layer Basics**
   - Volume superblock (fs/volume.go)
   - Inode structures (fs/inode.go)
   - Directory entries (fs/dentry.go)

6. **Data Access**
   - Data streams (fs/datastream.go)
   - File extents (fs/extents.go)
   - Extended fields (fs/extfields.go)

7. **Extended Functionality**
   - Extended attributes (fs/xattr.go)
   - Hard link handling (fs/siblings.go)
   - Snapshot support (snapshot/)

8. **Encryption Support** (crypto/)
   - Keybag handling
   - Encryption/decryption utilities

9. **Advanced Features**
   - Transaction handling (transaction/)
   - Reaper implementation (container/reaper.go)
   - Fusion drive support (fusion/)

10. **Command-Line Tools**
    - Info tool (cmd/apfs-info)
    - Mount tool with FUSE (cmd/apfs-mount)
    - Recovery tool (cmd/apfs-recover)

This order follows a logical progression from the lowest-level building blocks to more advanced features, allowing you to build and test a basic read-only implementation before expanding to more complex features. I'd recommend implementing a minimal read-only file system first, then adding support for snapshots, encryption, and finally write capabilities.
