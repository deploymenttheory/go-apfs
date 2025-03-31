# APFS Project Architecture

## Overview

This architecture follows the layered design of Apple File System as described in the reference documentation, with a clear separation between the container layer and file-system layer.

## Directory Structure

```bash
apfs/
├── cmd/                     # Command-line interfaces
│   ├── apfs-info/           # Tool to display APFS container/volume info
│   ├── apfs-mount/          # Tool to mount APFS volumes
│   └── apfs-recover/        # Data recovery tool
├── internal/                # Non-exported internal packages
│   └── binary/              # Binary parsing utilities
└── pkg/                     # Exported package code
    ├── service/
    │   ├── container_manager.go
    │   ├── volume_manager.go
    │   ├── space_manager.go
    │   ├── transaction_manager.go
    │   ├── crypto_manager.go
    │   └── object_manager.go
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
