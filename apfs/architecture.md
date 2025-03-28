# APFS Project Architecture

## Overview

This architecture follows the layered design of Apple File System as described in the reference documentation, with a clear separation between the container layer and file-system layer.

## Directory Structure

```
apfs/
├── cmd/                     # Command-line interfaces
│   ├── apfs-info/           # Tool to display APFS container/volume info
│   ├── apfs-mount/          # Tool to mount APFS volumes
│   └── apfs-recover/        # Data recovery tool
├── internal/                # Non-exported internal packages
│   └── binary/              # Binary parsing utilities
└── pkg/                     # Exported package code
    ├── checksum/            # Checksum
    │   └── fletcher64.go    # Checksum algorithm
    ├── types/               # Core types and constants
    │   ├── constants.go     # All APFS constants from the spec
    │   ├── errors.go        # Error definitions
    │   ├── types.go         # Common data structures and interfaces
    │   ├── interfaces.go    # Extended interfaces
    │   ├── structs.go       # All APFS on-disk data structures as Go structs
    │   ├── binary.go        # Serialization/deserialization helpers
    │   └── version.go       # Version compatibility checking
    ├── container/           # Container layer
    │   ├── object.go        # Object structures (obj_phys_t)
    │   ├── checkpoint.go    # Checkpoint mechanism + Checkpoint scanning
    │   ├── container.go     # Container manager (nx_superblock_t)
    │   ├── mount.go         # Container mounting procedures
    │   ├── omap.go          # Object maps + Object resolution
    │   ├── resolver.go      # Virtual object resolution
    │   ├── spaceman.go      # Space manager
    │   ├── btree.go         # B-tree structures
    │   ├── navigator.go     # B-tree traversal helpers
    │   └── reaper.go        # Reaper for delayed deletion
    ├── fs/                  # File system layer
    │   ├── volume.go        # Volume structures (apfs_superblock_t)
    │   ├── mount.go         # Volume mounting operations
    │   ├── navigator.go     # File system navigation
    │   ├── inode.go         # Inode structures and operations
    │   ├── dentry.go        # Directory entry structures
    │   ├── lookup.go        # Path lookup and traversal
    │   ├── file.go          # File access implementation
    │   ├── directory.go     # Directory access implementation
    │   ├── xattr.go         # Extended attributes
    │   ├── datastream.go    # File data stream handling
    │   ├── extents.go       # File extent management
    │   ├── extfields.go     # Extended field handling
    │   └── siblings.go      # Hard link management
    ├── crypto/              # Encryption support
    │   ├── keybag.go        # Keybag structures and handling
    │   ├── keys.go          # KEK/VEK key management
    │   └── crypto.go        # Encryption/decryption utilities
    ├── snapshot/            # Snapshot management
    │   ├── snapshot.go      # Snapshot structures
    │   └── operations.go    # Snapshot operations
    ├── fusion/              # Fusion drive support
    │   ├── fusion.go        # Fusion drive structures
    │   └── tier.go          # Tier management
    ├── transaction/         # Transaction handling
    │   ├── transaction.go   # Transaction structures and operations
    │   └── operations.go    # Operation interfaces
    └── util/                # Utilities
        ├── io.go            # I/O utilities
        ├── checksum.go      # Fletcher64 implementation
        ├── bits.go          # Bit manipulation utilities
        └── uuid.go          # UUID handling
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
