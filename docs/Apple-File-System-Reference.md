# Apple File System Reference

**Developer Documentation**  
**2020-06-22 | Copyright © 2020 Apple Inc. All Rights Reserved.**

## Contents

- [About Apple File System](#about-apple-file-system)
- [General-Purpose Types](#general-purpose-types)
- [Objects](#objects)
- [EFI Jumpstart](#efi-jumpstart)
- [Container](#container)
- [Object Maps](#object-maps)
- [Volumes](#volumes)
- [File-System Objects](#file-system-objects)
- [File-System Constants](#file-system-constants)
- [Data Streams](#data-streams)
- [Extended Fields](#extended-fields)
- [Siblings](#siblings)
- [Snapshot Metadata](#snapshot-metadata)
- [B-Trees](#b-trees)
- [Encryption](#encryption)
- [Sealed Volumes](#sealed-volumes)
- [Space Manager](#space-manager)
- [Reaper](#reaper)
- [Encryption Rolling](#encryption-rolling)
- [Fusion](#fusion)
- [Symbol Index](#symbol-index)
- [Revision History](#revision-history)

## About Apple File System

Apple File System is the default file format used on Apple platforms. Apple File System is the successor to HFS Plus, so some aspects of its design intentionally follow HFS Plus to enable data migration from HFS Plus to Apple File System. Other aspects of its design address limitations with HFS Plus and enable features like cloning files, snapshots, encryption, and sharing free space between volumes.

Most apps interact with the file system using high-level interfaces provided by Foundation, which means most developers don't need to read this document. This document is for developers of software that interacts with the file system directly, without using any frameworks or the operating system — for example, a disk recovery utility or an implementation of Apple File System on another platform. The on-disk data structures described in this document make up the file system; software that interacts with them defines corresponding in-memory data structures.

> **Note**: If you need to boot from an Apple File System volume, but don't need to mount the volume or interact with the file system directly, read [Booting from an Apple File System Partition](#booting-from-an-apple-file-system-partition).

### Layered Design

The Apple File System is conceptually divided into two layers, the container layer and the file-system layer. The container layer organizes file-system layer information and stores higher level information, like volume metadata, snapshots of the volume, and encryption state. The file-system layer is made up of the data structures that store information, like directory structures, file metadata, and file content. Many types are prefixed with `nx_` or `j_`, which indicates that they're part of the container layer or the file-system layer, respectively. The abbreviated prefixes don't have a meaningful long form; they're an artifact of how Apple's implementation was developed.

There are several design differences between the layers. Container objects are larger, with a typical size measured in blocks, and contain padding fields that keep data aligned on 64-bit boundaries, to avoid the performance penalty of unaligned memory access. File-system objects are smaller, with a typical size measured in bytes, and are almost always packed to minimize space used.

Numbers in both layers are stored on disk in little-endian order. Objects in both layers begin with a common header that enables object-oriented design patterns in implementations of Apple File System, although the layers have different headers. Container layer objects begin with an instance of `obj_phys_t` and file-system objects begin with an instance of `j_key_t`.

#### Container Layer

Container objects have an object identifier that you use to locate the object; the steps vary depending on how the object is stored:

- **Physical objects** are stored on disk at a particular physical block address.
- **Ephemeral objects** are stored in memory while the container is mounted and in a checkpoint when the container isn't mounted.
- **Virtual objects** are stored on disk at a location that you look up in an object map (an instance of `omap_phys_t`).

The object map includes a B-tree whose keys contain a transaction identifier and an object identifier and whose values contain a physical block address where the object is stored.

An Apple File System partition has a single container, which provides space management and crash protection. A container can contain multiple volumes (also known as file systems), each of which contains a directory structure for files and folders.

Although there's only one container, there are several copies of the container superblock (an instance of `nx_superblock_t`) stored on disk. These copies hold the state of the container at past points in time. Block zero contains a copy of the container superblock that's used as part of the mounting process to find the checkpoints. Block zero is typically a copy of the latest container superblock, assuming the device was properly unmounted and was last modified by a correct Apple File System implementation. However, in practice, you use the block zero copy only to find the checkpoints and use the latest version from the checkpoint for everything else.

Within a container, the checkpoint mechanism and the copy-on-write approach to modifying objects enable crash protection. In-memory state is periodically written to disk in checkpoints, followed by a copy of the container superblock at that point in time. Checkpoint information is stored in two regions: The checkpoint descriptor area contains instances of `checkpoint_map_phys_t` and `nx_superblock_t`, and the checkpoint data area contains ephemeral objects that represent the in-memory state at the point in time when the checkpoint was written to disk.

When mounting a device, you use the most recent checkpoint information that's valid, as discussed in [Mounting an Apple File System Partition](#mounting-an-apple-file-system-partition). If the process of writing a checkpoint is interrupted, that checkpoint is invalid and therefore is ignored the next time the device is mounted, rolling the file system back to the last valid state. Because the checkpoint stores in-memory state, mounting an Apple File System partition includes reading the ephemeral objects from the checkpoint back into memory, re-creating that state in memory.

#### File-System Layer

File-system objects are made up of several records, and each record is stored as a key and value in a B-tree (an instance of `btree_node_phys_t`). For example, a typical directory object is made up of an inode record, several directory entry records, and an extended attributes record. A record contains an object identifier that's used to find it within the B-tree that contains it.

## General-Purpose Types

Basic types that are used in a variety of contexts, and aren't associated with any particular functionality.

### paddr_t

A physical address of an on-disk block.

```c
typedef int64_t paddr_t;
```

Negative numbers aren't valid addresses. This value is modeled as a signed integer to match IOKit.

### prange_t

A range of physical addresses.

```c
struct prange {
    paddr_t pr_start_paddr;
    uint64_t pr_block_count;
};
typedef struct prange prange_t;
```

#### Fields

- **pr_start_paddr**: The first block in the range.
- **pr_block_count**: The number of blocks in the range.

### uuid_t

A universally unique identifier.

```c
typedef unsigned char uuid_t[16];
```

## Objects

Depending on how they're stored, objects have some differences, the most important of which is the way you use an object identifier to find an object. At the container level, there are three storage methods for objects:

- **Ephemeral objects** are stored in memory for a mounted container, and are persisted across unmounts in a checkpoint. Ephemeral objects for a mounted partition can be modified in place while they're in memory, but they're always written back to disk as part of a new checkpoint. They're used for information that's frequently updated because of the performance benefits of in-place, in-memory changes.

- **Physical objects** are stored at a known block address on the disk, and are modified by writing the copy to a new location on disk. Because the object identifier for a physical object is its physical address, this copy-on-write behavior means that the modified copy has a different object identifier.

- **Virtual objects** are stored on disk at a block address that you look up using an object map. Virtual objects are also copied when they are modified; however, both the original and the modified copy have the same object identifier. When you look up a virtual object in an object map, you use a transaction identifier, in addition to the object identifier, to specify the point in time that you want.

Regardless of their storage, objects on disk are never modified in place, and modified copies of an object are always written to a new location on disk. To access an object, you need to know its storage and its identifier. For virtual objects, you also need a transaction identifier. The storage for an object is almost always implicit from the context in which that identifier appears.

Object identifiers are unique inside the entire container, within their storage method. For example, no two virtual objects can have the same identifier — even when stored in different object maps — because their storage methods are the same. However, a virtual object and a physical object can have the same identifier because their storage methods are different.

When writing a new object to disk, fill all unused space in the block with zeros. Future versions of Apple File System add new fields at the end of a structure; zeroing out the uninitialized bytes makes it possible to determine whether data has been stored in a field that was added later.

### obj_phys_t

A header used at the beginning of all objects.

```c
struct obj_phys {
    uint8_t o_cksum[MAX_CKSUM_SIZE];
    oid_t o_oid;
    xid_t o_xid;
    uint32_t o_type;
    uint32_t o_subtype;
};
typedef struct obj_phys obj_phys_t;

#define MAX_CKSUM_SIZE 8
```

#### Fields

- **o_cksum**: The Fletcher 64 checksum of the object.
- **o_oid**: The object's identifier.
- **o_xid**: The identifier of the most recent transaction that this object was modified in.
- **o_type**: The object's type and flags.
- **o_subtype**: The object's subtype.

### Supporting Data Types

Types used as unique identifiers within an object.

```c
typedef uint64_t oid_t;
typedef uint64_t xid_t;
```

#### oid_t

An object identifier.

Objects are identified by this number as follows:

- For a physical object, its identifier is the logical block address on disk where the object is stored.
- For an ephemeral object, its identifier is a number.
- For a virtual object, its identifier is a number.

To determine the identifier for a new physical object, find a free block using the space manager, and use that block's address. To determine the identifier for a new ephemeral or virtual object, check the value of `nx_superblock_t.nx_next_oid`. New ephemeral and virtual object identifiers must be monotonically increasing.

#### xid_t

A transaction identifier.

Transactions are uniquely identified by a monotonically increasing number.

The number zero isn't a valid transaction identifier. Implementations of Apple File System can use it as a sentinel value in memory — for example, to refer to the current transaction — but must not let it appear on disk.

### Object Identifier Constants

Constants used for virtual objects that always have a given identifier.

```c
#define OID_NX_SUPERBLOCK 1
#define OID_INVALID 0ULL
#define OID_RESERVED_COUNT 1024
```

### Object Type Masks

Bit masks used to access specific portions of an object type.

```c
#define OBJECT_TYPE_MASK 0x0000ffff
#define OBJECT_TYPE_FLAGS_MASK 0xffff0000
#define OBJ_STORAGETYPE_MASK 0xc0000000
#define OBJECT_TYPE_FLAGS_DEFINED_MASK 0xf8000000
```

### Object Types

Values used as types and subtypes by the `obj_phys_t` structure.

```c
#define OBJECT_TYPE_NX_SUPERBLOCK 0x00000001
#define OBJECT_TYPE_BTREE 0x00000002
#define OBJECT_TYPE_BTREE_NODE 0x00000003
#define OBJECT_TYPE_SPACEMAN 0x00000005
#define OBJECT_TYPE_SPACEMAN_CAB 0x00000006
#define OBJECT_TYPE_SPACEMAN_CIB 0x00000007
#define OBJECT_TYPE_SPACEMAN_BITMAP 0x00000008
#define OBJECT_TYPE_SPACEMAN_FREE_QUEUE 0x00000009
#define OBJECT_TYPE_EXTENT_LIST_TREE 0x0000000a
#define OBJECT_TYPE_OMAP 0x0000000b
#define OBJECT_TYPE_CHECKPOINT_MAP 0x0000000c
#define OBJECT_TYPE_FS 0x0000000d
#define OBJECT_TYPE_FSTREE 0x0000000e
#define OBJECT_TYPE_BLOCKREFTREE 0x0000000f
#define OBJECT_TYPE_SNAPMETATREE 0x00000010
#define OBJECT_TYPE_NX_REAPER 0x00000011
#define OBJECT_TYPE_NX_REAP_LIST 0x00000012
#define OBJECT_TYPE_OMAP_SNAPSHOT 0x00000013
#define OBJECT_TYPE_EFI_JUMPSTART 0x00000014
#define OBJECT_TYPE_FUSION_MIDDLE_TREE 0x00000015
#define OBJECT_TYPE_NX_FUSION_WBC 0x00000016
#define OBJECT_TYPE_NX_FUSION_WBC_LIST 0x00000017
#define OBJECT_TYPE_ER_STATE 0x00000018
#define OBJECT_TYPE_GBITMAP 0x00000019
#define OBJECT_TYPE_GBITMAP_TREE 0x0000001a
#define OBJECT_TYPE_GBITMAP_BLOCK 0x0000001b
#define OBJECT_TYPE_ER_RECOVERY_BLOCK 0x0000001c
#define OBJECT_TYPE_SNAP_META_EXT 0x0000001d
#define OBJECT_TYPE_INTEGRITY_META 0x0000001e
#define OBJECT_TYPE_FEXT_TREE 0x0000001f
#define OBJECT_TYPE_RESERVED_20 0x00000020
#define OBJECT_TYPE_INVALID 0x00000000
#define OBJECT_TYPE_TEST 0x000000ff
#define OBJECT_TYPE_CONTAINER_KEYBAG 'keys'
#define OBJECT_TYPE_VOLUME_KEYBAG 'recs'
#define OBJECT_TYPE_MEDIA_KEYBAG 'mkey'
```

### Object Type Flags

The flags used in the object type to provide additional information.

```c
#define OBJ_VIRTUAL 0x00000000
#define OBJ_EPHEMERAL 0x80000000
#define OBJ_PHYSICAL 0x40000000
#define OBJ_NOHEADER 0x20000000
#define OBJ_ENCRYPTED 0x10000000
#define OBJ_NONPERSISTENT 0x08000000
```

## EFI Jumpstart

A partition formatted using the Apple File System contains an embedded EFI driver that's used to boot a machine from that partition.

### Booting from an Apple File System Partition

You can locate the EFI driver by reading a few data structures, starting at a known physical address on disk. You don't need any support for reading or mounting Apple File System to locate the EFI driver. This design intentionally simplifies the steps needed to boot, which means the code needed to boot a piece of hardware or virtualization software can likewise be simpler. To boot using the embedded EFI driver, do the following:

1. Read physical block zero from the partition. This block contains a copy of the container superblock, which is an instance of `nx_superblock_t`.

2. Read the `nx_o` field of the superblock, which is an instance of `obj_phys_t`. Then read the `o_cksum` field of the `nx_o` field of the superblock, which contains the Fletcher 64 checksum of the object. Verify that the checksum is correct.

3. Read the `nx_magic` field of the superblock. Verify that the field's value is `NX_MAGIC` (the four-character code 'BSXN').

4. Read the `nx_efi_jumpstart` field of the superblock. This field contains the physical block address (also referred to as the physical object identifier) for the EFI jumpstart information, which is an instance of `nx_efi_jumpstart_t`.

5. Read the `nej_magic` field of the EFI jumpstart information. Verify that the field's value is `NX_EFI_JUMPSTART_MAGIC` (the four-character code 'RDSJ').

6. Read the `nej_o` field of the EFI jumpstart information, which is an instance of `obj_phys_t`. Then read the `o_cksum` field of the `nej_o` field of the jumpstart information, which contains the Fletcher 64 checksum of the object. Verify that the checksum is correct.

7. Read the `nej_version` field of the EFI jumpstart information. This field contains the EFI jumpstart version number. Verify that the field's value is `NX_EFI_JUMPSTART_VERSION` (the number one).

8. Read the `nej_efi_file_len` field of the jumpstart information. This field contains the length, in bytes, of the embedded EFI driver. Allocate a contiguous block of memory of at least that size, which you'll later use to store the EFI driver.

9. Read the `nej_num_extents` field of the jumpstart information, and then read that number of `prange_t` records from the `nej_rec_extents` field.

10. Read each extent of the EFI driver into memory, contiguously, in the order they're listed.

11. Load the EFI driver and start executing it.

### nx_efi_jumpstart_t

Information about the embedded EFI driver that's used to boot from an Apple File System partition.

```c
struct nx_efi_jumpstart {
    obj_phys_t nej_o;
    uint32_t nej_magic;
    uint32_t nej_version;
    uint32_t nej_efi_file_len;
    uint32_t nej_num_extents;
    uint64_t nej_reserved[16];
    prange_t nej_rec_extents[];
};
typedef struct nx_efi_jumpstart nx_efi_jumpstart_t;

#define NX_EFI_JUMPSTART_MAGIC 'RDSJ'
#define NX_EFI_JUMPSTART_VERSION 1
```

#### Fields

- **nej_o**: The object's header.
- **nej_magic**: A number that can be used to verify that you're reading an instance of `nx_efi_jumpstart_t`. The value of this field is always `NX_EFI_JUMPSTART_MAGIC`.
- **nej_version**: The version of this data structure. The value of this field is always `NX_EFI_JUMPSTART_VERSION`.
- **nej_efi_file_len**: The size, in bytes, of the embedded EFI driver.
- **nej_num_extents**: The number of extents in the array.
- **nej_reserved**: Reserved. Populate this field with zero when you create a new instance, and preserve its value when you modify an existing instance.
- **nej_rec_extents**: The locations where the EFI driver is stored.

### Partition UUIDs

Partition types used in GUID partition table entries.

```c
#define APFS_GPT_PARTITION_UUID "7C3457EF-0000-11AA-AA11-00306543ECAC"
```

## Container

The container includes several top-level objects that are shared by all of the container's volumes:

- **Checkpoint description and data areas** store ephemeral objects in a way that provides crash protection. At the end of each transaction, new state is saved by writing a checkpoint.
- **The space manager** keeps track of available space within the container and is used to allocate and free blocks that store objects and file data.
- **The reaper** manages the deletion of objects that are too large to be deleted in the time between transactions. It keeps track of the deletion state so these objects can be deleted across multiple transactions.

The container superblock describes the location of all of these objects.

Because a single container can have multiple volumes, configurations that would require multiple partitions under other file systems can usually share a single partition with Apple File System. For example, a drive can be configured with two bootable volumes — one with a shipping version of macOS and one with a beta version — as well as a user data volume. All three of these volumes share free space, meaning you don't have to decide ahead of time how to divide space between them.

### Mounting an Apple File System Partition

To mount the volumes of a partition that's formatted using the Apple File System, do the following:

1. Read block zero of the partition. This block contains a copy of the container superblock (an instance of `nx_superblock_t`). It might be a copy of the latest version or an old version, depending on whether the drive was unmounted cleanly.

2. Use the block-zero copy of the container superblock to locate the checkpoint descriptor area by reading the `nx_xp_desc_base` field.

3. Read the entries in the checkpoint descriptor area, which are instances of `checkpoint_map_phys_t` or `nx_superblock_t`.

4. Find the container superblock that has the largest transaction identifier and isn't malformed. For example, confirm that its magic number and checksum are valid. That superblock and its checkpoint-mapping blocks comprise the latest valid checkpoint.

5. Read the ephemeral objects listed in the checkpoint from the checkpoint data area into memory. If any of the ephemeral objects is malformed, the checkpoint that contains that object is malformed; go back to the previous step and mount from an older checkpoint.

6. Locate the container object map using the `nx_omap_oid` field of the container superblock.

7. Read the list of volumes from the `nx_fs_oid` field of the container superblock.

8. For each volume, look up the specified virtual object identifier in the container object map to locate the volume superblock (an instance of `apfs_superblock_t`).

9. For each volume, read the root file system tree's virtual object identifier from the `apfs_root_tree_oid` field, and then look it up in the volume object map indicated by the `apfs_omap_oid` field.

10. Walk the root file system tree as needed by your implementation to mount the file system.

### nx_superblock_t

A container superblock.

```c
struct nx_superblock {
    obj_phys_t nx_o;
    uint32_t nx_magic;
    uint32_t nx_block_size;
    uint64_t nx_block_count;
    uint64_t nx_features;
    uint64_t nx_readonly_compatible_features;
    uint64_t nx_incompatible_features;
    uuid_t nx_uuid;
    oid_t nx_next_oid;
    xid_t nx_next_xid;
    uint32_t nx_xp_desc_blocks;
    uint32_t nx_xp_data_blocks;
    paddr_t nx_xp_desc_base;
    paddr_t nx_xp_data_base;
    uint32_t nx_xp_desc_next;
    uint32_t nx_xp_data_next;
    uint32_t nx_xp_desc_index;
    uint32_t nx_xp_desc_len;
    uint32_t nx_xp_data_index;
    uint32_t nx_xp_data_len;
    oid_t nx_spaceman_oid;
    oid_t nx_omap_oid;
    oid_t nx_reaper_oid;
    uint32_t nx_test_type;
    uint32_t nx_max_file_systems;
    oid_t nx_fs_oid[NX_MAX_FILE_SYSTEMS];
    uint64_t nx_counters[NX_NUM_COUNTERS];
    prange_t nx_blocked_out_prange;
    oid_t nx_evict_mapping_tree_oid;
    uint64_t nx_flags;
    paddr_t nx_efi_jumpstart;
    uuid_t nx_fusion_uuid;
    prange_t nx_keylocker;
    uint64_t nx_ephemeral_info[NX_EPH_INFO_COUNT];
    oid_t nx_test_oid;
    oid_t nx_fusion_mt_oid;
    oid_t nx_fusion_wbc_oid;
    prange_t nx_fusion_wbc;
    uint64_t nx_newest_mounted_version;
    prange_t nx_mkb_locker;
};
typedef struct nx_superblock nx_superblock_t;

#define NX_MAGIC 'BSXN'
#define NX_MAX_FILE_SYSTEMS 100
#define NX_EPH_INFO_COUNT 4
#define NX_EPH_MIN_BLOCK_COUNT 8
#define NX_MAX_FILE_SYSTEM_EPH_STRUCTS 4
#define NX_TX_MIN_CHECKPOINT_COUNT 4
#define NX_EPH_INFO_VERSION_1 1
```

Note that all fields are 64-bit aligned.

### Container Flags

The flags used for general information about a container.

```c
#define NX_RESERVED_1 0x00000001LL
#define NX_RESERVED_2 0x00000002LL
#define NX_CRYPTO_SW 0x00000004LL
```

### Optional Container Feature Flags

The flags used to describe optional features of an Apple File System container.

```c
#define NX_FEATURE_DEFRAG 0x0000000000000001ULL
#define NX_FEATURE_LCFD 0x0000000000000002ULL
#define NX_SUPPORTED_FEATURES_MASK (NX_FEATURE_DEFRAG | NX_FEATURE_LCFD)
```

### Read-Only Compatible Container Feature Flags

The flags used to describe read-only compatible features of an Apple File System container.

```c
#define NX_SUPPORTED_ROCOMPAT_MASK (0x0ULL)
```

### Incompatible Container Feature Flags

The flags used to describe backward-incompatible features of an Apple File System container.

```c
#define NX_INCOMPAT_VERSION1 0x0000000000000001ULL
#define NX_INCOMPAT_VERSION2 0x0000000000000002ULL
#define NX_INCOMPAT_FUSION 0x0000000000000100ULL
#define NX_SUPPORTED_INCOMPAT_MASK (NX_INCOMPAT_VERSION2 | NX_INCOMPAT_FUSION)
```

## Object Maps

An object map uses a B-tree to store a mapping from virtual object identifiers and transaction identifiers to the physical addresses where those objects are stored. The keys in the B-tree are instances of `omap_key_t` and the values are instances of `paddr_t`.

To access a virtual object using the object map, perform the following operations:

1. Determine which object map to use. Objects that are within a volume use that volume's object map, and all other objects use the container's object map.

2. Locate the object map for the volume by reading the `apfs_omap_oid` field of `apfs_superblock_t` or the `nx_omap_oid` field of `nx_superblock_t`.

3. Locate the B-tree for the object map by reading the `om_tree_oid` field of `omap_phys_t`.

4. Search the B-tree for a key whose object identifier is the same as the desired object identifier, and whose transaction identifier is less than or equal to the desired transaction identifier. If there are multiple keys that satisfy this test, use the key with the largest transaction identifier.

5. Using the table of contents entry, read the corresponding value for the key you found, which contains a physical address.

6. Read the object from disk at that address.

### omap_phys_t

An object map.

```c
struct omap_phys {
    obj_phys_t om_o;
    uint32_t om_flags;
    uint32_t om_snap_count;
    uint32_t om_tree_type;
    uint32_t om_snapshot_tree_type;
    oid_t om_tree_oid;
    oid_t om_snapshot_tree_oid;
    xid_t om_most_recent_snap;
    xid_t om_pending_revert_min;
    xid_t om_pending_revert_max;
};
typedef struct omap_phys omap_phys_t;
```

### omap_key_t

A key used to access an entry in the object map.

```c
struct omap_key {
    oid_t ok_oid;
    xid_t ok_xid;
};
typedef struct omap_key omap_key_t;
```

### omap_val_t

A value in the object map.

```c
struct omap_val {
    uint32_t ov_flags;
    uint32_t ov_size;
    paddr_t ov_paddr;
};
typedef struct omap_val omap_val_t;
```

## Volumes

A volume contains a file system, the files and metadata that make up that file system, and various supporting data structures like an object map.

### apfs_superblock_t

A volume superblock.

```c
struct apfs_superblock {
    obj_phys_t apfs_o;
    uint32_t apfs_magic;
    uint32_t apfs_fs_index;
    uint64_t apfs_features;
    uint64_t apfs_readonly_compatible_features;
    uint64_t apfs_incompatible_features;
    uint64_t apfs_unmount_time;
    uint64_t apfs_fs_reserve_block_count;
    uint64_t apfs_fs_quota_block_count;
    uint64_t apfs_fs_alloc_count;
    wrapped_meta_crypto_state_t apfs_meta_crypto;
    uint32_t apfs_root_tree_type;
    uint32_t apfs_extentref_tree_type;
    uint32_t apfs_snap_meta_tree_type;
    oid_t apfs_omap_oid;
    oid_t apfs_root_tree_oid;
    oid_t apfs_extentref_tree_oid;
    oid_t apfs_snap_meta_tree_oid;
    xid_t apfs_revert_to_xid;
    oid_t apfs_revert_to_sblock_oid;
    uint64_t apfs_next_obj_id;
    uint64_t apfs_num_files;
    uint64_t apfs_num_directories;
    uint64_t apfs_num_symlinks;
    uint64_t apfs_num_other_fsobjects;
    uint64_t apfs_num_snapshots;
    uint64_t apfs_total_blocks_alloced;
    uint64_t apfs_total_blocks_freed;
    uuid_t apfs_vol_uuid;
    uint64_t apfs_last_mod_time;
    uint64_t apfs_fs_flags;
    apfs_modified_by_t apfs_formatted_by;
    apfs_modified_by_t apfs_modified_by[APFS_MAX_HIST];
    uint8_t apfs_volname[APFS_VOLNAME_LEN];
    uint32_t apfs_next_doc_id;
    uint16_t apfs_role;
    uint16_t reserved;
    xid_t apfs_root_to_xid;
    oid_t apfs_er_state_oid;
    uint64_t apfs_cloneinfo_id_epoch;
    uint64_t apfs_cloneinfo_xid;
    oid_t apfs_snap_meta_ext_oid;
    uuid_t apfs_volume_group_id;
    oid_t apfs_integrity_meta_oid;
    oid_t apfs_fext_tree_oid;
    uint32_t apfs_fext_tree_type;
    uint32_t reserved_type;
    oid_t reserved_oid;
};

#define APFS_MAGIC 'BSPA'
#define APFS_MAX_HIST 8
#define APFS_VOLNAME_LEN 256
```

### Volume Flags

The flags used to indicate volume status.

```c
#define APFS_FS_UNENCRYPTED 0x00000001LL
#define APFS_FS_RESERVED_2 0x00000002LL
#define APFS_FS_RESERVED_4 0x00000004LL
#define APFS_FS_ONEKEY 0x00000008LL
#define APFS_FS_SPILLEDOVER 0x00000010LL
#define APFS_FS_RUN_SPILLOVER_CLEANER 0x00000020LL
#define APFS_FS_ALWAYS_CHECK_EXTENTREF 0x00000040LL
#define APFS_FS_RESERVED_80 0x00000080LL
#define APFS_FS_RESERVED_100 0x00000100LL
```

### Volume Roles

The values used to indicate a volume's roles.

```c
#define APFS_VOL_ROLE_NONE 0x0000
#define APFS_VOL_ROLE_SYSTEM 0x0001
#define APFS_VOL_ROLE_USER 0x0002
#define APFS_VOL_ROLE_RECOVERY 0x0004
#define APFS_VOL_ROLE_VM 0x0008
#define APFS_VOL_ROLE_PREBOOT 0x0010
#define APFS_VOL_ROLE_INSTALLER 0x0020
#define APFS_VOL_ROLE_DATA (1 << APFS_VOLUME_ENUM_SHIFT)
#define APFS_VOL_ROLE_BASEBAND (2 << APFS_VOLUME_ENUM_SHIFT)
#define APFS_VOL_ROLE_UPDATE (3 << APFS_VOLUME_ENUM_SHIFT)
#define APFS_VOL_ROLE_XART (4 << APFS_VOLUME_ENUM_SHIFT)
#define APFS_VOL_ROLE_HARDWARE (5 << APFS_VOLUME_ENUM_SHIFT)
#define APFS_VOL_ROLE_BACKUP (6 << APFS_VOLUME_ENUM_SHIFT)
#define APFS_VOL_ROLE_RESERVED_7 (7 << APFS_VOLUME_ENUM_SHIFT)
#define APFS_VOL_ROLE_RESERVED_8 (8 << APFS_VOLUME_ENUM_SHIFT)
#define APFS_VOL_ROLE_ENTERPRISE (9 << APFS_VOLUME_ENUM_SHIFT)
#define APFS_VOL_ROLE_RESERVED_10 (10 << APFS_VOLUME_ENUM_SHIFT)
#define APFS_VOL_ROLE_PRELOGIN (11 << APFS_VOLUME_ENUM_SHIFT)
#define APFS_VOLUME_ENUM_SHIFT 6
```

## File-System Objects

A file-system object stores information about a part of the file system, like a directory or a file on disk. These objects are stored as one or more records. For example, the file-system object for a directory that contains two files is stored as three records: a record of type `APFS_TYPE_INODE` for the inode, and two records of type `APFS_TYPE_DIR_REC` for the directory entries.

File-system records are stored as key/value pairs in a B-tree. The key contains information, like the object identifier and the record type, used to look up a record. Keys begin with an instance of `j_key_t`, and many records use `j_key_t` as their entire key.

### j_key_t

A header used at the beginning of all file-system keys.

```c
struct j_key {
    uint64_t obj_id_and_type;
} __attribute__((packed));
typedef struct j_key j_key_t;

#define OBJ_ID_MASK 0x0fffffffffffffffULL
#define OBJ_TYPE_MASK 0xf000000000000000ULL
#define OBJ_TYPE_SHIFT 60
#define SYSTEM_OBJ_ID_MARK 0x0fffffff00000000ULL
```

### j_inode_val_t

The value half of an inode record.

```c
struct j_inode_val {
    uint64_t parent_id;
    uint64_t private_id;
    uint64_t create_time;
    uint64_t mod_time;
    uint64_t change_time;
    uint64_t access_time;
    uint64_t internal_flags;
    union {
        int32_t nchildren;
        int32_t nlink;
    };
    cp_key_class_t default_protection_class;
    uint32_t write_generation_counter;
    uint32_t bsd_flags;
    uid_t owner;
    gid_t group;
    mode_t mode;
    uint16_t pad1;
    uint64_t uncompressed_size;
    uint8_t xfields[];
} __attribute__((packed));
typedef struct j_inode_val j_inode_val_t;

typedef uint32_t uid_t;
typedef uint32_t gid_t;
```

## File-System Constants

File-system objects use several groups of constants to define values for record types, reserved inode numbers, and flags and bit masks used in bit fields.

### j_obj_types

The type of a file-system record.

```c
typedef enum {
    APFS_TYPE_ANY = 0,
    APFS_TYPE_SNAP_METADATA = 1,
    APFS_TYPE_EXTENT = 2,
    APFS_TYPE_INODE = 3,
    APFS_TYPE_XATTR = 4,
    APFS_TYPE_SIBLING_LINK = 5,
    APFS_TYPE_DSTREAM_ID = 6,
    APFS_TYPE_CRYPTO_STATE = 7,
    APFS_TYPE_FILE_EXTENT = 8,
    APFS_TYPE_DIR_REC = 9,
    APFS_TYPE_DIR_STATS = 10,
    APFS_TYPE_SNAP_NAME = 11,
    APFS_TYPE_SIBLING_MAP = 12,
    APFS_TYPE_FILE_INFO = 13,
    APFS_TYPE_MAX_VALID = 13,
    APFS_TYPE_MAX = 15,
    APFS_TYPE_INVALID = 15,
} j_obj_types;
```

### Inode Numbers

Inodes whose number is always the same.

```c
#define INVALID_INO_NUM 0
#define ROOT_DIR_PARENT 1
#define ROOT_DIR_INO_NUM 2
#define PRIV_DIR_INO_NUM 3
#define SNAP_DIR_INO_NUM 6
#define PURGEABLE_DIR_INO_NUM 7
#define MIN_USER_INO_NUM 16
#define UNIFIED_ID_SPACE_MARK 0x0800000000000000ULL
```

### File Modes

The values used by the mode field of `j_inode_val_t` to indicate a file's mode.

```c
typedef uint16_t mode_t;

#define S_IFMT 0170000
#define S_IFIFO 0010000
#define S_IFCHR 0020000
#define S_IFDIR 0040000
#define S_IFBLK 0060000
#define S_IFREG 0100000
#define S_IFLNK 0120000
#define S_IFSOCK 0140000
#define S_IFWHT 0160000
```

## Data Streams

Short pieces of information like a file's name are stored inside the data structures that contain metadata. Data that's too large to store inline is stored separately, in a data stream. This includes the contents of files, and the value of some attributes.

### j_phys_ext_key_t

The key half of a physical extent record.

```c
struct j_phys_ext_key {
    j_key_t hdr;
} __attribute__((packed));
typedef struct j_phys_ext_key j_phys_ext_key_t;
```

### j_phys_ext_val_t

The value half of a physical extent record.

```c
struct j_phys_ext_val {
    uint64_t len_and_kind;
    uint64_t owning_obj_id;
    int32_t refcnt;
} __attribute__((packed));
typedef struct j_phys_ext_val j_phys_ext_val_t;

#define PEXT_LEN_MASK 0x0fffffffffffffffULL
#define PEXT_KIND_MASK 0xf000000000000000ULL
#define PEXT_KIND_SHIFT 60
```

### j_file_extent_key_t

The key half of a file extent record.

```c
struct j_file_extent_key {
    j_key_t hdr;
    uint64_t logical_addr;
} __attribute__((packed));
typedef struct j_file_extent_key j_file_extent_key_t;
```

### j_file_extent_val_t

The value half of a file extent record.

```c
struct j_file_extent_val {
    uint64_t len_and_flags;
    uint64_t phys_block_num;
    uint64_t crypto_id;
} __attribute__((packed));
typedef struct j_file_extent_val j_file_extent_val_t;

#define J_FILE_EXTENT_LEN_MASK 0x00ffffffffffffffULL
#define J_FILE_EXTENT_FLAG_MASK 0xff00000000000000ULL
#define J_FILE_EXTENT_FLAG_SHIFT 56
```

# Extended Fields

Directory entries and inodes use extended fields to store a dynamically extensible set of member fields.

To determine whether a directory entry or an inode has any extended fields, find the table of contents entry for the file-system record, and then compare the recorded size to the size of the structure. For example:

```c
kvloc_t toc_entry = /* assume this exists */
if (toc_entry.v.len == sizeof(j_drec_val_t)) {
    // no extended fields
} else {
    // at least one extended field
}
```

Both `j_drec_val_t` and `j_inode_val_t` have an `xfields` field that contains several kinds of data, stored one after another, ordered as follows:

1. An instance of `xf_blob_t`, which tells you how many extended fields there are, and how many bytes they take up on disk.
2. An array of instances of `x_field_t`, one for each extended field, which tells you the field's type and size.
3. An array of extended-field data, aligned to eight-byte boundaries.

The arrays of extended-field metadata (`x_field_t`) and extended-field data are stored in the same order. The extended-field data's type depends on the field. For a list of field types, see [Extended-Field Types](#extended-field-types).

## xf_blob_t

A collection of extended attributes.

```c
struct xf_blob {
    uint16_t xf_num_exts;
    uint16_t xf_used_data;
    uint8_t xf_data[];
};
typedef struct xf_blob xf_blob_t;
```

Directory entries (`j_drec_val_t`) and inodes (`j_inode_val_t`) use this data type to store their extended fields.

### Fields

#### xf_num_exts
The number of extended attributes.
```c
uint16_t xf_num_exts;
```

#### xf_used_data
The amount of space, in bytes, used to store the extended attributes.
```c
uint16_t xf_used_data;
```
This total includes both the space used to store metadata, as instances of `x_field_t`, and values.

#### xf_data
The extended fields.
```c
uint8_t xf_data[];
```
This field contains an array of instances of `x_field_t`, followed by the extended field data.

## x_field_t

An extended field's metadata.

```c
struct x_field {
    uint8_t x_type;
    uint8_t x_flags;
    uint16_t x_size;
};
typedef struct x_field x_field_t;
```

This type is used by `xf_blob_t` to store an array of extended fields. Within the array, each extended field must have a unique type.

The extended field's data is stored outside of this structure, as part of the space set aside by the directory entry or inode.

### Fields

#### x_type
The extended field's data type.
```c
uint8_t x_type;
```
For possible values, see [Extended-Field Types](#extended-field-types).

#### x_flags
The extended field's flags.
```c
uint8_t x_flags;
```
For the values used in this bit field, see [Extended-Field Flags](#extended-field-flags).

#### x_size
The size, in bytes, of the data stored in the extended field.
```c
uint16_t x_size;
```

## Extended-Field Types

Values used by the `x_type` field of `x_field_t` to indicate an extended field's type.

```c
#define DREC_EXT_TYPE_SIBLING_ID 1
#define INO_EXT_TYPE_SNAP_XID 1
#define INO_EXT_TYPE_DELTA_TREE_OID 2
#define INO_EXT_TYPE_DOCUMENT_ID 3
#define INO_EXT_TYPE_NAME 4
#define INO_EXT_TYPE_PREV_FSIZE 5
#define INO_EXT_TYPE_RESERVED_6 6
#define INO_EXT_TYPE_FINDER_INFO 7
#define INO_EXT_TYPE_DSTREAM 8
#define INO_EXT_TYPE_RESERVED_9 9
#define INO_EXT_TYPE_DIR_STATS_KEY 10
#define INO_EXT_TYPE_FS_UUID 11
#define INO_EXT_TYPE_RESERVED_12 12
#define INO_EXT_TYPE_SPARSE_BYTES 13
#define INO_EXT_TYPE_RDEV 14
#define INO_EXT_TYPE_PURGEABLE_FLAGS 15
#define INO_EXT_TYPE_ORIG_SYNC_ROOT_ID 16
```

### DREC_EXT_TYPE_SIBLING_ID
The sibling identifier for a directory record (uint64_t).
```c
#define DREC_EXT_TYPE_SIBLING_ID 1
```
The corresponding sibling-link record has the same identifier in the `sibling_id` field of `j_sibling_key_t`.

This extended field is used only for hard links.

### INO_EXT_TYPE_SNAP_XID
The transaction identifier for a snapshot (xid_t).
```c
#define INO_EXT_TYPE_SNAP_XID 1
```

### INO_EXT_TYPE_DELTA_TREE_OID
The virtual object identifier of the file-system tree that corresponds to a snapshot's extent delta list (oid_t).
```c
#define INO_EXT_TYPE_DELTA_TREE_OID 2
```
The tree object's subtype is always `OBJECT_TYPE_FSTREE`.

### INO_EXT_TYPE_DOCUMENT_ID
The file's document identifier (uint32_t).
```c
#define INO_EXT_TYPE_DOCUMENT_ID 3
```
The document identifier lets applications keep track of the document during operations like atomic save, where one folder replaces another. The document identifier remains associated with the full path, not just with the inode that's currently at that path. Implementations of Apple File System must preserve the document identifier when the inode at that path is replaced.

Both documents that are stored as a bundle and documents that are stored as a single file can have a document identifier assigned.

Valid document identifiers are greater than `MIN_DOC_ID` and less than `UINT32_MAX - 1`. For the next document identifier that will be assigned, see the `apfs_next_doc_id` field of `apfs_superblock_t`.

### INO_EXT_TYPE_NAME
The name of the file, represented as a null-terminated UTF-8 string.
```c
#define INO_EXT_TYPE_NAME 4
```
This extended field is used only for hard links: The name stored in the inode is the name of the primary link to the file, and the name of the hard link is stored in this extended field.

### INO_EXT_TYPE_PREV_FSIZE
The file's previous size (uint64_t).
```c
#define INO_EXT_TYPE_PREV_FSIZE 5
```
This extended field is used for recovering after a crash. If it's set on an inode, truncate the file back to the size contained in this field.

### INO_EXT_TYPE_RESERVED_6
Reserved.
```c
#define INO_EXT_TYPE_RESERVED_6 6
```
Don't create extended fields of this type in your own code. Preserve the value of any extended fields of this type.

### INO_EXT_TYPE_FINDER_INFO
Opaque data stored and used by Finder (32 bytes).
```c
#define INO_EXT_TYPE_FINDER_INFO 7
```

### INO_EXT_TYPE_DSTREAM
A data stream (j_dstream_t).
```c
#define INO_EXT_TYPE_DSTREAM 8
```

### INO_EXT_TYPE_RESERVED_9
Reserved.
```c
#define INO_EXT_TYPE_RESERVED_9 9
```
Don't create extended fields of this type. When you modify an existing volume, preserve the contents of any extended fields of this type.

### INO_EXT_TYPE_DIR_STATS_KEY
Statistics about a directory (j_dir_stats_val_t).
```c
#define INO_EXT_TYPE_DIR_STATS_KEY 10
```

### INO_EXT_TYPE_FS_UUID
The UUID of a file system that's automatically mounted in this directory (uuid_t).
```c
#define INO_EXT_TYPE_FS_UUID 11
```
This value matches the value of the `apfs_vol_uuid` field of `apfs_superblock_t`.

### INO_EXT_TYPE_RESERVED_12
Reserved.
```c
#define INO_EXT_TYPE_RESERVED_12 12
```
Don't create extended fields of this type. If you find an object of this type in production, file a bug against the Apple File System implementation.

### INO_EXT_TYPE_SPARSE_BYTES
The number of sparse bytes in the data stream (uint64_t).
```c
#define INO_EXT_TYPE_SPARSE_BYTES 13
```

### INO_EXT_TYPE_RDEV
The device identifier for a block- or character-special device (uint32_t).
```c
#define INO_EXT_TYPE_RDEV 14
```
This extended field stores the same information as the `st_rdev` field of the stat structure defined in `<sys/stat.h>`.

### INO_EXT_TYPE_PURGEABLE_FLAGS
Information about a purgeable file.
```c
#define INO_EXT_TYPE_PURGEABLE_FLAGS 15
```
The value of this extended field is reserved. Don't create new extended fields of this type. When duplicating a file or directory, omit this extended field from the new copy.

Purgeable files have the `INODE_IS_PURGEABLE` flag set on the `internal_flags` field of `j_inode_val_t`.

### INO_EXT_TYPE_ORIG_SYNC_ROOT_ID
The inode number of the sync-root hierarchy that this file originally belonged to.
```c
#define INO_EXT_TYPE_ORIG_SYNC_ROOT_ID 16
```
The specified inode always has the `INODE_IS_SYNC_ROOT` flag set.

## Extended-Field Flags

The flags used by an extended field's metadata.

```c
#define XF_DATA_DEPENDENT 0x0001
#define XF_DO_NOT_COPY 0x0002
#define XF_RESERVED_4 0x0004
#define XF_CHILDREN_INHERIT 0x0008
#define XF_USER_FIELD 0x0010
#define XF_SYSTEM_FIELD 0x0020
#define XF_RESERVED_40 0x0040
#define XF_RESERVED_80 0x0080
```

These flags are used by the `x_flags` field of `x_field_t`.

### XF_DATA_DEPENDENT
The data in this extended field depends on the file's data.
```c
#define XF_DATA_DEPENDENT 0x0001
```
When the file data changes, this extended field must be updated to match the new data. If it's not possible to update the field — for example, because the Apple File System implementation doesn't recognize the field's type — the field must be removed.

### XF_DO_NOT_COPY
When copying this file, omit this extended field from the copy.
```c
#define XF_DO_NOT_COPY 0x0002
```

### XF_RESERVED_4
Reserved.
```c
#define XF_RESERVED_4 0x0004
```
Don't set this flag, but preserve it if it's already set.

### XF_CHILDREN_INHERIT
When creating a new entry in this directory, copy this extended field to the new directory entry.
```c
#define XF_CHILDREN_INHERIT 0x0008
```

### XF_USER_FIELD
This extended field was added by a user-space program.
```c
#define XF_USER_FIELD 0x0010
```

### XF_SYSTEM_FIELD
This extended field was added by the kernel, by the implementation of Apple File System, or by another system component.
```c
#define XF_SYSTEM_FIELD 0x0020
```
Extended fields with this flag set can't be removed or modified by a program running in user space.

### XF_RESERVED_40
Reserved.
```c
#define XF_RESERVED_40 0x0040
```
Don't set this flag, but preserve it if it's already set.

### XF_RESERVED_80
Reserved.
```c
#define XF_RESERVED_80 0x0080
```
Don't set this flag, but preserve it if it's already set.

# Siblings

Hard links that all refer to the same inode are called siblings. Each sibling has its own identifier that's used instead of the shared inode number when siblings need to be distinguished. For example, some Carbon APIs in macOS use sibling identifiers.

The sibling whose identifier is the lowest number is called the primary link. The other siblings copy various properties of the primary link, as discussed in `j_inode_val_t`.

You use sibling links and sibling maps to convert between sibling identifiers and inode numbers. Sibling-link records let you find all the hard links whose target is a given inode. Sibling-map records let you find the target inode of a given hard link.

## j_sibling_key_t

The key half of a sibling-link record.

```c
struct j_sibling_key {
    j_key_t hdr;
    uint64_t sibling_id;
} __attribute__((packed));
typedef struct j_sibling_key j_sibling_key_t;
```

### Fields

#### hdr
The record's header.
```c
j_key_t hdr;
```
The object identifier in the header is the file-system object's identifier, that is, its inode number. The type in the header is always `APFS_TYPE_SIBLING_LINK`.

#### sibling_id
The sibling's unique identifier.
```c
uint64_t sibling_id;
```
This value matches the object identifier for the sibling map record (`j_sibling_key_t`).

## j_sibling_val_t

The value half of a sibling-link record.

```c
struct j_sibling_val {
    uint64_t parent_id;
    uint16_t name_len;
    uint8_t name[0];
} __attribute__((packed));
typedef struct j_sibling_val j_sibling_val_t;
```

### Fields

#### parent_id
The object identifier for the inode that's the parent directory.
```c
uint64_t parent_id;
```

#### name_len
The length of the name, including the final null character (U+0000).
```c
uint16_t name_len;
```

#### name
The name, represented as a null-terminated UTF-8 string.
```c
uint8_t name[0];
```

## j_sibling_map_key_t

The key half of a sibling-map record.

```c
struct j_sibling_map_key {
    j_key_t hdr;
} __attribute__((packed));
typedef struct j_sibling_map_key j_sibling_map_key_t;
```

### Fields

#### hdr
The record's header.
```c
j_key_t hdr;
```
The object identifier in the header is the sibling's unique identifier, which matches the `sibling_id` field of `j_sibling_key_t`. The type in the header is always `APFS_TYPE_SIBLING_MAP`.

## j_sibling_map_val_t

The value half of a sibling-map record.

```c
struct j_sibling_map_val {
    uint64_t file_id;
} __attribute__((packed));
typedef struct j_sibling_map_val j_sibling_map_val_t;
```

### Fields

#### file_id
The inode number of the underlying file.
```c
uint64_t file_id;
```

# Extended Fields

Directory entries and inodes use extended fields to store a dynamically extensible set of member fields.

To determine whether a directory entry or an inode has any extended fields, find the table of contents entry for the file-system record, and then compare the recorded size to the size of the structure. For example:

```c
kvloc_t toc_entry = /* assume this exists */
if (toc_entry.v.len == sizeof(j_drec_val_t)) {
    // no extended fields
} else {
    // at least one extended field
}
```

Both `j_drec_val_t` and `j_inode_val_t` have an `xfields` field that contains several kinds of data, stored one after another, ordered as follows:

1. An instance of `xf_blob_t`, which tells you how many extended fields there are, and how many bytes they take up on disk.
2. An array of instances of `x_field_t`, one for each extended field, which tells you the field's type and size.
3. An array of extended-field data, aligned to eight-byte boundaries.

The arrays of extended-field metadata (`x_field_t`) and extended-field data are stored in the same order. The extended-field data's type depends on the field. For a list of field types, see [Extended-Field Types](#extended-field-types).

## xf_blob_t

A collection of extended attributes.

```c
struct xf_blob {
    uint16_t xf_num_exts;
    uint16_t xf_used_data;
    uint8_t xf_data[];
};
typedef struct xf_blob xf_blob_t;
```

Directory entries (`j_drec_val_t`) and inodes (`j_inode_val_t`) use this data type to store their extended fields.

### Fields

#### xf_num_exts
The number of extended attributes.
```c
uint16_t xf_num_exts;
```

#### xf_used_data
The amount of space, in bytes, used to store the extended attributes.
```c
uint16_t xf_used_data;
```
This total includes both the space used to store metadata, as instances of `x_field_t`, and values.

#### xf_data
The extended fields.
```c
uint8_t xf_data[];
```
This field contains an array of instances of `x_field_t`, followed by the extended field data.

## x_field_t

An extended field's metadata.

```c
struct x_field {
    uint8_t x_type;
    uint8_t x_flags;
    uint16_t x_size;
};
typedef struct x_field x_field_t;
```

This type is used by `xf_blob_t` to store an array of extended fields. Within the array, each extended field must have a unique type.

The extended field's data is stored outside of this structure, as part of the space set aside by the directory entry or inode.

### Fields

#### x_type
The extended field's data type.
```c
uint8_t x_type;
```
For possible values, see [Extended-Field Types](#extended-field-types).

#### x_flags
The extended field's flags.
```c
uint8_t x_flags;
```
For the values used in this bit field, see [Extended-Field Flags](#extended-field-flags).

#### x_size
The size, in bytes, of the data stored in the extended field.
```c
uint16_t x_size;
```

## Extended-Field Types

Values used by the `x_type` field of `x_field_t` to indicate an extended field's type.

```c
#define DREC_EXT_TYPE_SIBLING_ID 1
#define INO_EXT_TYPE_SNAP_XID 1
#define INO_EXT_TYPE_DELTA_TREE_OID 2
#define INO_EXT_TYPE_DOCUMENT_ID 3
#define INO_EXT_TYPE_NAME 4
#define INO_EXT_TYPE_PREV_FSIZE 5
#define INO_EXT_TYPE_RESERVED_6 6
#define INO_EXT_TYPE_FINDER_INFO 7
#define INO_EXT_TYPE_DSTREAM 8
#define INO_EXT_TYPE_RESERVED_9 9
#define INO_EXT_TYPE_DIR_STATS_KEY 10
#define INO_EXT_TYPE_FS_UUID 11
#define INO_EXT_TYPE_RESERVED_12 12
#define INO_EXT_TYPE_SPARSE_BYTES 13
#define INO_EXT_TYPE_RDEV 14
#define INO_EXT_TYPE_PURGEABLE_FLAGS 15
#define INO_EXT_TYPE_ORIG_SYNC_ROOT_ID 16
```

### DREC_EXT_TYPE_SIBLING_ID
The sibling identifier for a directory record (uint64_t).
```c
#define DREC_EXT_TYPE_SIBLING_ID 1
```
The corresponding sibling-link record has the same identifier in the `sibling_id` field of `j_sibling_key_t`.

This extended field is used only for hard links.

### INO_EXT_TYPE_SNAP_XID
The transaction identifier for a snapshot (xid_t).
```c
#define INO_EXT_TYPE_SNAP_XID 1
```

### INO_EXT_TYPE_DELTA_TREE_OID
The virtual object identifier of the file-system tree that corresponds to a snapshot's extent delta list (oid_t).
```c
#define INO_EXT_TYPE_DELTA_TREE_OID 2
```
The tree object's subtype is always `OBJECT_TYPE_FSTREE`.

### INO_EXT_TYPE_DOCUMENT_ID
The file's document identifier (uint32_t).
```c
#define INO_EXT_TYPE_DOCUMENT_ID 3
```
The document identifier lets applications keep track of the document during operations like atomic save, where one folder replaces another. The document identifier remains associated with the full path, not just with the inode that's currently at that path. Implementations of Apple File System must preserve the document identifier when the inode at that path is replaced.

Both documents that are stored as a bundle and documents that are stored as a single file can have a document identifier assigned.

Valid document identifiers are greater than `MIN_DOC_ID` and less than `UINT32_MAX - 1`. For the next document identifier that will be assigned, see the `apfs_next_doc_id` field of `apfs_superblock_t`.

### INO_EXT_TYPE_NAME
The name of the file, represented as a null-terminated UTF-8 string.
```c
#define INO_EXT_TYPE_NAME 4
```
This extended field is used only for hard links: The name stored in the inode is the name of the primary link to the file, and the name of the hard link is stored in this extended field.

### INO_EXT_TYPE_PREV_FSIZE
The file's previous size (uint64_t).
```c
#define INO_EXT_TYPE_PREV_FSIZE 5
```
This extended field is used for recovering after a crash. If it's set on an inode, truncate the file back to the size contained in this field.

### INO_EXT_TYPE_RESERVED_6
Reserved.
```c
#define INO_EXT_TYPE_RESERVED_6 6
```
Don't create extended fields of this type in your own code. Preserve the value of any extended fields of this type.

### INO_EXT_TYPE_FINDER_INFO
Opaque data stored and used by Finder (32 bytes).
```c
#define INO_EXT_TYPE_FINDER_INFO 7
```

### INO_EXT_TYPE_DSTREAM
A data stream (j_dstream_t).
```c
#define INO_EXT_TYPE_DSTREAM 8
```

### INO_EXT_TYPE_RESERVED_9
Reserved.
```c
#define INO_EXT_TYPE_RESERVED_9 9
```
Don't create extended fields of this type. When you modify an existing volume, preserve the contents of any extended fields of this type.

### INO_EXT_TYPE_DIR_STATS_KEY
Statistics about a directory (j_dir_stats_val_t).
```c
#define INO_EXT_TYPE_DIR_STATS_KEY 10
```

### INO_EXT_TYPE_FS_UUID
The UUID of a file system that's automatically mounted in this directory (uuid_t).
```c
#define INO_EXT_TYPE_FS_UUID 11
```
This value matches the value of the `apfs_vol_uuid` field of `apfs_superblock_t`.

### INO_EXT_TYPE_RESERVED_12
Reserved.
```c
#define INO_EXT_TYPE_RESERVED_12 12
```
Don't create extended fields of this type. If you find an object of this type in production, file a bug against the Apple File System implementation.

### INO_EXT_TYPE_SPARSE_BYTES
The number of sparse bytes in the data stream (uint64_t).
```c
#define INO_EXT_TYPE_SPARSE_BYTES 13
```

### INO_EXT_TYPE_RDEV
The device identifier for a block- or character-special device (uint32_t).
```c
#define INO_EXT_TYPE_RDEV 14
```
This extended field stores the same information as the `st_rdev` field of the stat structure defined in `<sys/stat.h>`.

### INO_EXT_TYPE_PURGEABLE_FLAGS
Information about a purgeable file.
```c
#define INO_EXT_TYPE_PURGEABLE_FLAGS 15
```
The value of this extended field is reserved. Don't create new extended fields of this type. When duplicating a file or directory, omit this extended field from the new copy.

Purgeable files have the `INODE_IS_PURGEABLE` flag set on the `internal_flags` field of `j_inode_val_t`.

### INO_EXT_TYPE_ORIG_SYNC_ROOT_ID
The inode number of the sync-root hierarchy that this file originally belonged to.
```c
#define INO_EXT_TYPE_ORIG_SYNC_ROOT_ID 16
```
The specified inode always has the `INODE_IS_SYNC_ROOT` flag set.

## Extended-Field Flags

The flags used by an extended field's metadata.

```c
#define XF_DATA_DEPENDENT 0x0001
#define XF_DO_NOT_COPY 0x0002
#define XF_RESERVED_4 0x0004
#define XF_CHILDREN_INHERIT 0x0008
#define XF_USER_FIELD 0x0010
#define XF_SYSTEM_FIELD 0x0020
#define XF_RESERVED_40 0x0040
#define XF_RESERVED_80 0x0080
```

These flags are used by the `x_flags` field of `x_field_t`.

### XF_DATA_DEPENDENT
The data in this extended field depends on the file's data.
```c
#define XF_DATA_DEPENDENT 0x0001
```
When the file data changes, this extended field must be updated to match the new data. If it's not possible to update the field — for example, because the Apple File System implementation doesn't recognize the field's type — the field must be removed.

### XF_DO_NOT_COPY
When copying this file, omit this extended field from the copy.
```c
#define XF_DO_NOT_COPY 0x0002
```

### XF_RESERVED_4
Reserved.
```c
#define XF_RESERVED_4 0x0004
```
Don't set this flag, but preserve it if it's already set.

### XF_CHILDREN_INHERIT
When creating a new entry in this directory, copy this extended field to the new directory entry.
```c
#define XF_CHILDREN_INHERIT 0x0008
```

### XF_USER_FIELD
This extended field was added by a user-space program.
```c
#define XF_USER_FIELD 0x0010
```

### XF_SYSTEM_FIELD
This extended field was added by the kernel, by the implementation of Apple File System, or by another system component.
```c
#define XF_SYSTEM_FIELD 0x0020
```
Extended fields with this flag set can't be removed or modified by a program running in user space.

### XF_RESERVED_40
Reserved.
```c
#define XF_RESERVED_40 0x0040
```
Don't set this flag, but preserve it if it's already set.

### XF_RESERVED_80
Reserved.
```c
#define XF_RESERVED_80 0x0080
```
Don't set this flag, but preserve it if it's already set.

# B-Trees

The B-trees used in Apple File System are implemented using the `btree_node_phys_t` structure to represent a node. The same structure is used for all nodes in a tree. Within a node, storage is divided into several areas:

- Information about the node
- The table of contents, which lists the location of keys and values
- Storage for the keys
- Storage for the values
- Information about the entire tree

The figure below shows the storage areas of a typical root node.

The instance of `btree_node_phys_t` stores information about this B-tree node, like its flags and the location of its keys, and is always located at the beginning of the block. For a root node, an instance of `btree_info_t` is located at the end of the block, and contains information like the sizes of keys and values, the total number of keys in the tree, and the number of nodes in the tree. Nonroot nodes omit `btree_info_t`. The rest of the block (the `btn_data` field of `btree_node_phys_t`) is organized dynamically.

Compared to other B-tree implementations, this data structure has some unique characteristics. Traversal is always done from the root node because nodes don't have parent or sibling pointers. All values are stored in leaf nodes, which makes these B+ trees, and the values in nonleaf nodes are object identifiers of child nodes. The keys, values, or both can be of variable size; if the keys and values of a node are both fixed in size, some optimizations for the table of contents are possible.

## Keys and Values

The keys and values are stored starting at opposite ends of the B-tree node's storage area, with free space that's available for new keys or values in the available portion of the storage area between them. The key and value areas grow toward each other into their shared free space. Free space within the key area and within the value area is organized using a free list. For example, free space appears outside the shared free space when an entry is removed from a B-tree. The figure below shows free space for keys and values in a typical nonroot node.

The locations of keys and values are stored as offsets, which uses less on-disk space than storing the full location. The offset to a key is counted from the beginning of the key area to the beginning of the key. The offset to a value is counted from the end of the value area to the beginning of the value.

Keys and value are normally aligned to eight-byte boundaries when stored. The length recorded for a key or value in the table of contents omits any padding needed for alignment. If the `BTREE_KV_NONALIGNED` flag is set, keys and values are stored without padding.

If the `BTREE_ALLOW_GHOSTS` flag is set on the B-tree, the tree can contain keys that have no value.

## Table of Contents

The table of contents stores the location of each key and value that form a key-value pair.

If the `BTNODE_FIXED_KV_SIZE` flag is set, the table of contents stores only the offsets for keys and values. Otherwise, it stores both their offsets and lengths.

Free space within the table of contents is located at the end. If there's no free space remaining, but a new entry is needed, the table of contents area must be expanded. The entire key area is shifted to make space available, using some of the shared free space for key space, and some space from the beginning of the key space for the table of contents. Because the offset to a key is counted relative to the beginning of the key area, moving the entire key area doesn't invalidate any of these offsets. Likewise, when the table of contents has too much unused space, it shrinks, and the key area is shifted into the space from the table of contents. Apple's implementation uses `BTREE_TOC_ENTRY_INCREMENT` and `BTREE_TOC_ENTRY_MAX_UNUSED` to determine when to expand or shrink the table of contents.

> **Note**: When the `BTNODE_FIXED_KV_SIZE` flag is set, Apple's implementation allocates enough space for the table of contents to avoid the need to expand it later. This is possible because the maximum number of entries is known, as well as the size of an entry. However, if the `BTREE_ALLOW_GHOSTS` flag is also set, the table of contents might still need to expand.

## Key Comparison

The entries in the table of contents are sorted by key. The comparison function used for sorting depends on the key's type. Object map B-trees are sorted by object identifier and then by transaction identifier. Free queue B-trees are sorted by transaction identifier and then by physical address. File-system records are sorted according to the rules listed in File-System Objects.

## btree_node_phys_t

A B-tree node.

```c
struct btree_node_phys {
    obj_phys_t btn_o;
    uint16_t btn_flags;
    uint16_t btn_level;
    uint32_t btn_nkeys;
    nloc_t btn_table_space;
    nloc_t btn_free_space;
    nloc_t btn_key_free_list;
    nloc_t btn_val_free_list;
    uint64_t btn_data[];
};
typedef struct btree_node_phys btree_node_phys_t;
```

The locations of the key and value areas aren't stored explicitly. The key area begins after the end of the table of contents and ends before the start of the shared free space. The value area begins after the end of shared free space and ends at the end of the B-tree node (for nonroot nodes) or before the instance of `btree_info_t` that's at the end of a root node.

### Fields

#### btn_o
The object's header.
```c
obj_phys_t btn_o;
```

#### btn_flags
The B-tree node's flags.
```c
uint16_t btn_flags;
```
For the values used in this bit field, see [B-Tree Node Flags](#b-tree-node-flags).

#### btn_level
The number of child levels below this node.
```c
uint16_t btn_level;
```
For example, the value of this field is zero for a leaf node and one for the immediate parent of a leaf node. Likewise, the height of a tree is one plus the value of this field on the tree's root node.

#### btn_nkeys
The number of keys stored in this node.
```c
uint32_t btn_nkeys;
```

#### btn_table_space
The location of the table of contents.
```c
nloc_t btn_table_space;
```
The offset for the table of contents is counted from the beginning of the node's `btn_data` field to the beginning of the table of contents.

If the `BTNODE_FIXED_KV_SIZE` flag is set, the table of contents is an array of instances of `kvoff_t`; otherwise, it's an array of instances of `kvloc_t`.

#### btn_free_space
The location of the shared free space for keys and values.
```c
nloc_t btn_free_space;
```
The location's offset is counted from the beginning of the key area to the beginning of the free space.

#### btn_key_free_list
A linked list that tracks free key space.
```c
nloc_t btn_key_free_list;
```
The offset from the beginning of the key area to the first available space for a key is stored in the `off` field, and the total amount of free key space is stored in the `len` field. Each free space stores an instance of `nloc_t` whose `len` field indicates the size of that free space and whose `off` field contains the location of the next free space.

#### btn_val_free_list
A linked list that tracks free value space.
```c
nloc_t btn_val_free_list;
```
The offset from the end of the value area to the first available space for a value is stored in the `off` field, and the total amount of free value space is stored in the `len` field. Each free space stores an instance of `nloc_t` whose `len` field indicates the size of that free space and whose `off` field contains the location of the next free space.

#### btn_data
The node's storage area.
```c
uint64_t btn_data[];
```
This area contains the table of contents, keys, free space, and values. A root node also has as an instance of `btree_info_t` at the end of its storage area. For more information, see B-trees.

## btree_info_fixed_t

Static information about a B-tree.

```c
struct btree_info_fixed {
    uint32_t bt_flags;
    uint32_t bt_node_size;
    uint32_t bt_key_size;
    uint32_t bt_val_size;
};
typedef struct btree_info_fixed btree_info_fixed_t;
```

This information doesn't change over time as the B-tree is modified. It's stored separately from the rest of the information in `btree_info_t`, which does change, to enable this information to be cached more easily.

### Fields

#### bt_flags
The B-tree's flags.
```c
uint32_t bt_flags;
```
For the values used in this bit field, see [B-Tree Flags](#b-tree-flags).

#### bt_node_size
The on-disk size, in bytes, of a node in this B-tree.
```c
uint32_t bt_node_size;
```
Leaf nodes, nonleaf nodes, and the root node are all the same size.

#### bt_key_size
The size of a key, or zero if the keys have variable size.
```c
uint32_t bt_key_size;
```
If this field has a value of zero, the `btn_flags` field of instances of `btree_node_phys_t` in this tree must not include `BTNODE_FIXED_KV_SIZE`.

#### bt_val_size
The size of a value, or zero if the values have variable size.
```c
uint32_t bt_val_size;
```
If this field has a value of zero, the `btn_flags` field of instances of `btree_node_phys_t` for leaf nodes in this tree must not include `BTNODE_FIXED_KV_SIZE`. Nonleaf nodes in a tree with variable-size values include `BTNODE_FIXED_KV_SIZE`, because the values stored in those nodes are the object identifiers of their child nodes, and object identifiers have a fixed size.

## btree_info_t

Information about a B-tree.

```c
struct btree_info {
    btree_info_fixed_t bt_fixed;
    uint32_t bt_longest_key;
    uint32_t bt_longest_val;
    uint64_t bt_key_count;
    uint64_t bt_node_count;
};
typedef struct btree_info btree_info_t;
```

This information appears only in a root node, stored at the end of the node.

### Fields

#### bt_fixed
Information about the B-tree that doesn't change over time.
```c
btree_info_fixed_t bt_fixed;
```

#### bt_longest_key
The length, in bytes, of the longest key that has ever been stored in the B-tree.
```c
uint32_t bt_longest_key;
```

#### bt_longest_val
The length, in bytes, of the longest value that has ever been stored in the B-tree.
```c
uint32_t bt_longest_val;
```

#### bt_key_count
The number of keys stored in the B-tree.
```c
uint64_t bt_key_count;
```

#### bt_node_count
The number of nodes stored in the B-tree.
```c
uint64_t bt_node_count;
```

## btn_index_node_val_t

The value used by hashed B-trees for nonleaf nodes.

```c
struct btn_index_node_val {
    oid_t binv_child_oid;
    uint8_t binv_child_hash[BTREE_NODE_HASH_SIZE_MAX];
};
typedef struct btn_index_node_val btn_index_node_val_t;

#define BTREE_NODE_HASH_SIZE_MAX 64
```

For nonhashed B-trees, instead of using this structure, the values are instances of `oid_t`. Because this structure's `oid_t` field comes first, code that's expecting only the object identifier of the child node as the B-tree value is still able to read the hashed B-tree by ignoring the hashes.

### Fields

#### binv_child_oid
The object identifier of the child node.
```c
oid_t binv_child_oid;
```

#### binv_child_hash
The hash of the child node.
```c
uint8_t binv_child_hash[BTREE_NODE_HASH_SIZE_MAX];
```
The hash algorithm used by this tree determines the length of the hash. See the `im_hash_type` field of `integrity_meta_phys_t`, and the `hash_size` field of `j_file_data_hash_val_t`.

To compute the hash, use the entire child node object as the input for the hash algorithm specified for this tree. If the output from that hash algorithm is smaller than the `BTREE_NODE_HASH_SIZE_MAX` bytes, treat the remaining bytes as padding — set them to zero when you create a new node, and preserve their value when you modify an existing node.

### Constants

#### BTREE_NODE_HASH_SIZE_MAX
The maximum length of a hash that can be stored in this structure.
```c
#define BTREE_NODE_HASH_SIZE_MAX 64
```
This value is the same as `APFS_HASH_MAX_SIZE`.

## nloc_t

A location within a B-tree node.

```c
struct nloc {
    uint16_t off;
    uint16_t len;
};
typedef struct nloc nloc_t;

#define BTOFF_INVALID 0xffff
```

### Fields

#### off
The offset, in bytes.
```c
uint16_t off;
```
Depending on the data type that contains this location, the offset is either implicitly positive or negative, and is counted starting at different points in the B-tree node.

#### len
The length, in bytes.
```c
uint16_t len;
```

### Constants

#### BTOFF_INVALID
An invalid offset.
```c
#define BTOFF_INVALID 0xffff
```
This value is stored in the `off` field of `nloc_t` to indicate that there's no offset. For example, the last entry in a free list has no entry after it, so it uses this value for its `off` field.

## kvloc_t

The location, within a B-tree node, of a key and value.

```c
struct kvloc {
    nloc_t k;
    nloc_t v;
};
typedef struct kvloc kvloc_t;
```

The B-tree node's table of contents uses this structure when the keys and values are not both fixed in size.

### Fields

#### k
The location of the key.
```c
nloc_t k;
```

#### v
The location of the value.
```c
nloc_t v;
```

## kvoff_t

The location, within a B-tree node, of a fixed-size key and value.

```c
struct kvoff {
    uint16_t k;
    uint16_t v;
};
typedef struct kvoff kvoff_t;
```

The B-tree node's table of contents uses this structure when the keys and values are both fixed in size. The meaning of the offsets stored in this structure's `k` and `v` fields is the same as the meaning of the `off` field in an instance of `nloc_t`. This structure doesn't have a field that's equivalent to the `len` field of `nloc_t` — the key and value lengths are always the same, and omitting them from the table of contents saves space.

### Fields

#### k
The offset of the key.
```c
uint16_t k;
```

#### v
The offset of the value.
```c
uint16_t v;
```

## B-Tree Flags

The flags used to describe configuration options for a B-tree.

```c
#define BTREE_UINT64_KEYS 0x00000001
#define BTREE_SEQUENTIAL_INSERT 0x00000002
#define BTREE_ALLOW_GHOSTS 0x00000004
#define BTREE_EPHEMERAL 0x00000008
#define BTREE_PHYSICAL 0x00000010
#define BTREE_NONPERSISTENT 0x00000020
#define BTREE_KV_NONALIGNED 0x00000040
#define BTREE_HASHED 0x00000080
#define BTREE_NOHEADER 0x00000100
```

### BTREE_UINT64_KEYS
Code that works with the B-tree should enable optimizations to make comparison of keys fast.
```c
#define BTREE_UINT64_KEYS 0x00000001
```
This is a hint used by Apple's implementation.

### BTREE_SEQUENTIAL_INSERT
Code that works with the B-tree should enable optimizations to keep the B-tree compact during sequential insertion of entries.
```c
#define BTREE_SEQUENTIAL_INSERT 0x00000002
```
This is a hint used by Apple's implementation.

Normally, nodes are split in half when they become almost full. With this flag set, a new node is added to provide the needed space, instead of splitting a node that's almost full. This yields a tree with nodes that are almost full instead of nodes that are about half full.

### BTREE_ALLOW_GHOSTS
The table of contents is allowed to contain keys that have no corresponding value.
```c
#define BTREE_ALLOW_GHOSTS 0x00000004
```
In the table of contents, a ghost is indicated by a value whose location offset is `BTOFF_INVALID`.

The meaning of a ghost depends on context — it can indicate a key that has been deleted and should be ignored, or a key whose value is implicit from context. For example, in the space manager's free queue, a ghost indicates a free extent that's one block long.

Using ghosts to store an implicit value allows more entries to be stored in some circumstances because no space in the value area is used by the ghost.

### BTREE_EPHEMERAL
The nodes in the B-tree use ephemeral object identifiers to link to child nodes.
```c
#define BTREE_EPHEMERAL 0x00000008
```
If this flag is set, `BTREE_PHYSICAL` must not be set. If neither flag is set, nodes in the B-tree use virtual object identifiers to link to their child nodes.

### BTREE_PHYSICAL
The nodes in the B-tree use physical object identifiers to link to child nodes.
```c
#define BTREE_PHYSICAL 0x00000010
```
If this flag is set, `BTREE_EPHEMERAL` must not be set. If neither flag is set, nodes in the B-tree use virtual object identifiers to link to their child nodes.

### BTREE_NONPERSISTENT
The B-tree isn't persisted across unmounting.
```c
#define BTREE_NONPERSISTENT 0x00000020
```
This flag is valid only when `BTREE_EPHEMERAL` is also set, and only on in-memory B-trees.

### BTREE_KV_NONALIGNED
The keys and values in the B-tree aren't required to be aligned to eight-byte boundaries.
```c
#define BTREE_KV_NONALIGNED 0x00000040
```
Aligning to eight-byte boundaries avoids unaligned reads on 64-bit platforms, which improves performance, but wastes space on disk for structures whose size isn't a multiple of eight bytes.

### BTREE_HASHED
The nonleaf nodes of this B-tree store a hash of their child nodes.
```c
#define BTREE_HASHED 0x00000080
```
If this flag is set, all nodes of this B-tree have the `BTNODE_HASHED` flag set.

The hash is stored in the `binv_child_hash` field of `btn_index_node_val_t`.

### BTREE_NOHEADER
The nodes of this B-tree are stored without object headers.
```c
#define BTREE_NOHEADER 0x00000100
```
If this flag is set, all nodes of this B-tree have the `BTNODE_NOHEADER` flag set.

## B-Tree Table of Contents Constants

Constants used in managing the size of the table of contents in a B-tree node.

```c
#define BTREE_TOC_ENTRY_INCREMENT 8
#define BTREE_TOC_ENTRY_MAX_UNUSED (2 * BTREE_TOC_ENTRY_INCREMENT)
```

These values are used by Apple's implementation; other implementations can choose different values. If you don't use these values, profile your implementation to determine the performance impact of your chosen values.

### BTREE_TOC_ENTRY_INCREMENT
The number of entries that are added or removed when changing the size of the table of contents.
```c
#define BTREE_TOC_ENTRY_INCREMENT 8
```

### BTREE_TOC_ENTRY_MAX_UNUSED
The maximum allowed number of unused entries in the table of contents.
```c
#define BTREE_TOC_ENTRY_MAX_UNUSED (2 * BTREE_TOC_ENTRY_INCREMENT)
```

## B-Tree Node Flags

The flags used with a B-tree node.

```c
#define BTNODE_ROOT 0x0001
#define BTNODE_LEAF 0x0002
#define BTNODE_FIXED_KV_SIZE 0x0004
#define BTNODE_HASHED 0x0008
#define BTNODE_NOHEADER 0x0010
#define BTNODE_CHECK_KOFF_INVAL 0x8000
```

### BTNODE_ROOT
The B-tree node is a root node.
```c
#define BTNODE_ROOT 0x0001
```
If this flag is set, the node's object type is `OBJECT_TYPE_BTREE`. If this is the tree's only node, both `BTNODE_ROOT` and `BTNODE_LEAF` are set. Otherwise, the `BTNODE_LEAF` flag must not be set.

### BTNODE_LEAF
The B-tree node is a leaf node.
```c
#define BTNODE_LEAF 0x0002
```
If this is the tree's only node, the node object's type is `OBJECT_TYPE_BTREE`, and both `BTNODE_ROOT` and `BTNODE_LEAF` are set. Otherwise, the node's object type is `OBJECT_TYPE_BTREE_NODE`, and the `BTNODE_ROOT` flag must not be set.

### BTNODE_FIXED_KV_SIZE
The B-tree node has keys and values of a fixed size, and the table of contents omits their lengths.
```c
#define BTNODE_FIXED_KV_SIZE 0x0004
```
If the keys and values both have a fixed size, this flag must be set.

Within the same B-tree, it's valid to have a mix of nodes that have this flag set and nodes that don't. For example, consider a B-tree with fixed-sized keys and variable-sized values. Leaf nodes in that tree don't have this flag set because of the variable-sized values. However, nonleaf nodes in in the same tree do have this flag set. The values stored in nonleaf nodes are object identifiers, which are fixed-sized values; therefore, this flag can be applied to nonleaf nodes of any tree with fixed-size keys.

### BTNODE_HASHED
The B-tree node contains child hashes.
```c
#define BTNODE_HASHED 0x0008
```
This flag is valid only on B-trees that have the `BTREE_HASHED` flag. You can this flag on a leaf node, for consistency with the nonleaf nodes in the same tree, but it doesn't mean anything there and is ignored.

If this flag isn't set, the `binv_child_hash` field of `btn_index_node_val_t` is unused.

### BTNODE_NOHEADER
The B-tree node is stored without an object header.
```c
#define BTNODE_NOHEADER 0x0010
```
This flag is valid only on B-trees that have the `BTREE_NOHEADER` flag.

If this flag is set, the `btn_o` field of this instance of `btree_node_phys_t` is always zero.

### BTNODE_CHECK_KOFF_INVAL
The B-tree node is in a transient state.
```c
#define BTNODE_CHECK_KOFF_INVAL 0x8000
```
Objects with this flag never appear on disk. If you find an object of this type in production, file a bug against the Apple File System implementation.

This flag isn't reserved by Apple; non-Apple implementations of Apple File System can set it on B-tree nodes in memory.

## B-Tree Node Constants

Constants used to determine the size of a B-tree node.

```c
#define BTREE_NODE_SIZE_DEFAULT 4096
#define BTREE_NODE_MIN_ENTRY_COUNT 4
```

A node is almost always one logical block in size. Smaller nodes waste space, and larger nodes can experience allocation issues when space is fragmented. For example, a five-block node requires five adjacent blocks to all be free, but on a fragmented disk such a large free space might not exist.

### BTREE_NODE_SIZE_DEFAULT
The default size, in bytes, of a B-tree node.
```c
#define BTREE_NODE_SIZE_DEFAULT 4096
```

### BTREE_NODE_MIN_ENTRY_COUNT
The minimum number of entries that must be able to fit in a nonleaf B-tree node.
```c
#define BTREE_NODE_MIN_ENTRY_COUNT 4
```

To satisfy this requirement, reduce the size of the keys stored in the node. The maximum key size is computed as follows:

```c
uint32_t btree_key_max_size(uint32_t nodesize) {
    uint32_t dataspace, esize, count, kvspace;
    dataspace = nodesize - offsetof(btree_node_phys_t, btn_data)
        - sizeof(btree_info_t);
    esize = sizeof(kvloc_t);
    count = BTREE_TOC_ENTRY_INCREMENT;
    kvspace = dataspace - (count * esize);
    return ((kvspace / BTREE_NODE_MIN_ENTRY_COUNT) - sizeof(oid_t));
}
```

> **Note**: This requirement comes from logic in Apple's implementation that performs proactive splitting of B-tree nodes.

# Encryption

Apple File System supports encryption in the data structures used for containers, volumes, and files. When a volume is encrypted, both its file-system tree and the contents of files in that volume are encrypted.

Depending on the device's capabilities, Apple File System uses either hardware or software encryption, which impacts encryption process and the meaning of several data structures. Hardware encryption is used for internal storage on devices that support it, including macOS (with T2 security chip) and iOS devices. Software encryption is used for external storage, and for internal storage on devices that don't support hardware encryption. When hardware encryption is in use, only the kernel can interact with internal storage.

> **Important**: This document describes only software encryption.

The keys used to access file data are stored on disk in a wrapped state. You access these keys through a chain of key-unwrapping operations. The volume encryption key (VEK) is the default key used to access encrypted content on the volume. The key encryption key (KEK) is used to unwrap the VEK. The KEK is unwrapped in one of several ways:

- **User password** - The user enters their password, which is used to unwrap the KEK.
- **Personal recovery key** - This key is generated when the drive is formatted and is saved by the user on a paper printout. The string on that printout is used to unwrap the KEK.
- **Institutional recovery key** - This key is enabled by the user in Settings and allows the corresponding corporate master key to unwrap the KEK.
- **iCloud recovery key** - This key is used by customers working with Apple Support, and isn't described in this document.

For example, to access a file given the user's password on a volume that uses per-volume encryption, the chain of key unwrapping and data decryption consists of the following high-level operations:

1. Unwrap the KEK using the user's password.
2. Unwrap the VEK using the KEK.
3. Decrypt the file-system B-tree using the VEK.
4. Decrypt the file data using the VEK.

The detailed steps are described in [Accessing Encrypted Objects](#accessing-encrypted-objects) below.

## Keybag

On macOS devices, both the container and the volume have a keybag (an instance of `kb_locker_t`). The container's keybag is stored at the location indicated by the `nx_keylocker` field of `nx_superblock_t`. For each volume, the container's keybag stores the volume's wrapped VEK and the location of the volume's keybag. The volume's keybag contains several copies of the volume's KEK, wrapped using user passwords and recovery keys.

Keybags are encrypted using the UUID of the container or volume, which makes it possible to quickly and securely destroy the contents of an encrypted volume by changing or deleting the UUID. For a volume, destroying the UUID by securely erasing a volume superblock makes the corresponding keybag unreadable, which in turn makes the encrypted content of that volume inaccessible. For a container superblock, you need to destroy all of the copies of that block in the checkpoint descriptor area and the copy at block zero.

## Accessing Encrypted Objects

Before accessing an encrypted object, confirm that the `APFS_FS_ONEKEY` flag is set on the volume. Volumes that use per-file encryption require hardware encryption, and the steps below describe only software encryption.

To obtain the unwrapped VEK for a volume, do the following:

1. Locate the container's keybag using the `nx_keylocker` field of `nx_superblock_t`.
2. Unwrap the container's keybag using the container's UUID, according to the algorithm described in RFC 3394.
3. Find an entry in the container's keybag whose UUID matches the volume's UUID and whose tag is `KB_TAG_VOLUME_KEY`. The key data for that entry is the wrapped VEK for this volume.
4. Find an entry in the container's keybag whose UUID matches the volume's UUID and whose tag is `KB_TAG_VOLUME_UNLOCK_RECORDS`. The key data for that entry is the location of the volume's keybag.
5. Unwrap the volume's keybag using the volume's UUID according to the algorithm described in RFC 3394.
6. Find an entry in the volume's keybag whose UUID matches the user's Open Directory UUID and whose tag is `KB_TAG_VOLUME_UNLOCK_RECORDS`. The key data for that entry is the wrapped KEK for this volume.
7. Unwrap the KEK using the user's password, and then unwrap the VEK using the KEK, both according to the algorithm described in RFC 3394.

The volume's keybag might contain a passphrase hint for the user (`KB_TAG_VOLUME_PASSPHRASE_HINT`), which you can display when prompting for the password. It also might contain an entry for a personal recovery key, using `APFS_FV_PERSONAL_RECOVERY_KEY_UUID` as the UUID. You follow the same process for a personal recovery key as you do for a password: Unwrap the KEK with the user-entered string, and then use the unwrapped KEK to unwrap the VEK, both according to the algorithm described in RFC 3394.

To decrypt a file, do the following:

1. Decrypt the blocks where the volume's root file-system tree is stored, using the VEK as an AES-XTS key. The file-system tree is accessed using the `apfs_root_tree_oid` field of `apfs_superblock_t`.
2. Find the file extent record (`APFS_TYPE_FILE_EXTENT`) for the encrypted file.
3. Find the encryption state record (`APFS_TYPE_CRYPTO_STATE`) whose identifier matches the `crypto_id` field of `j_file_extent_val_t`.
4. Decrypt the blocks where the file's data is stored, using the VEK as an AES-XTS key and the value of `crypto_id` as the tweak.

## j_crypto_key_t

The key half of a per-file encryption state record.

```c
struct j_crypto_key {
    j_key_t hdr;
} __attribute__((packed));
typedef struct j_crypto_key j_crypto_key_t;
```

Several encryption state objects always have the same identifier, as listed in [Encryption Identifiers](#encryption-identifiers).

### Fields

#### hdr
The record's header.
```c
j_key_t hdr;
```
The object identifier in the header is the file-system object's identifier. The type in the header is always `APFS_TYPE_CRYPTO_STATE`.

## j_crypto_val_t

The value half of a per-file encryption state record.

```c
struct j_crypto_val {
    uint32_t refcnt;
    wrapped_crypto_state_t state;
} __attribute__((aligned(4),packed));
typedef struct j_crypto_val j_crypto_val_t;
```

### Fields

#### refcnt
The reference count.
```c
int32_t refcnt;
```
The encryption state record can be deleted when its reference count reaches zero.

#### state
The encryption state information.
```c
wrapped_crypto_state_t state;
```
If this encryption state record is used by the file-system tree rather than by a file, this field is an instance of `wrapped_meta_crypto_state_t` and the key used is always the volume encryption key (VEK).

## wrapped_crypto_state_t

A wrapped key used for per-file encryption.

```c
struct wrapped_crypto_state {
    uint16_t major_version;
    uint16_t minor_version;
    crypto_flags_t cpflags;
    cp_key_class_t persistent_class;
    cp_key_os_version_t key_os_version;
    cp_key_revision_t key_revision;
    uint16_t key_len;
    uint8_t persistent_key[0];
} __attribute__((aligned(2), packed));
typedef struct wrapped_crypto_state wrapped_crypto_state_t;

#define CP_MAX_WRAPPEDKEYSIZE 128
```

This structure is used inside of `j_crypto_val_t`.

### Fields

#### major_version
The major version for this structure's layout.
```c
uint16_t major_version;
```
The current value of this field is five. If backward-incompatible changes are made to this data structure in the future, the major version number will be incremented.

This structure is equivalent to a structure used by iOS for per-file encryption on HFS-Plus; versions four and earlier were used by previous versions of that structure.

#### minor_version
The minor version for this structure's layout.
```c
uint16_t minor_version;
```
The current value of this field is zero. If backward-compatible changes are made to this data structure in the future, the minor version number will be incremented.

#### cpflags
The encryption state's flags.
```c
crypto_flags_t cpflags;
```
There are currently none defined.

#### persistent_class
The protection class associated with the key.
```c
cp_key_class_t persistent_class;
```
For possible values and the bit mask that must be used, see [Protection Classes](#protection-classes).

#### key_os_version
The version of the OS that created this structure.
```c
cp_key_os_version_t key_os_version;
```
This field is used as part of key rolling. For information about how the major version number, minor version number, and build number are packed into 32 bits, see `cp_key_os_version_t`.

#### key_revision
The version of the key.
```c
cp_key_revision_t key_revision;
```
Set this field to one when creating a new instance, and increment it by one when rolling to a new key.

#### key_len
The size, in bytes, of the wrapped key data.
```c
uint16_t key_len;
```
The maximum value of this field is `CP_MAX_WRAPPEDKEYSIZE`.

#### persistent_key
The wrapped key data.
```c
uint8_t persistent_key[0];
```

### Constants

#### CP_MAX_WRAPPEDKEYSIZE
The size, in bytes, of the largest possible key.
```c
#define CP_MAX_WRAPPEDKEYSIZE 128
```

## wrapped_meta_crypto_state_t

Information about how the volume encryption key (VEK) is used to encrypt a file.

```c
struct wrapped_meta_crypto_state {
    uint16_t major_version;
    uint16_t minor_version;
    crypto_flags_t cpflags;
    cp_key_class_t persistent_class;
    cp_key_os_version_t key_os_version;
    cp_key_revision_t key_revision;
    uint16_t unused;
} __attribute__((aligned(2), packed));
typedef struct wrapped_meta_crypto_state wrapped_meta_crypto_state_t;
```

This structure is used inside of `j_crypto_val_t`. The fields in this structure are the same as `wrapped_crypto_state_t`, except this structure doesn't contain a wrapped key.

### Fields

#### major_version
The major version for this structure's layout.
```c
uint16_t major_version;
```
The value of this field is always five. This structure is equivalent to a structure used by iOS for per-file encryption on HFS-Plus; versions four and earlier were used by previous versions of that structure.

#### minor_version
The minor version for this structure's layout.
```c
uint16_t minor_version;
```
The value of this field is always zero.

#### cpflags
The encryption state's flags.
```c
crypto_flags_t cpflags;
```
There are currently none defined.

#### persistent_class
The protection class associated with the key.
```c
cp_key_class_t persistent_class;
```
For possible values, see [Protection Classes](#protection-classes).

#### key_os_version
The version of the OS that created this structure.
```c
cp_key_os_version_t key_os_version;
```
For information about how the major version number, minor version number, and build number are packed into 32 bits, see `cp_key_os_version_t`.

#### key_revision
The version of the key.
```c
cp_key_revision_t key_revision;
```
Set this field to one when creating a new instance.

#### unused
Reserved.
```c
uint16_t unused;
```
Populate this field with zero when you create a new instance of this structure, and preserve its value when you modify an existing instance.

## Encryption Types

Data types used in encryption-related structures.

```c
typedef uint32_t cp_key_class_t;
typedef uint32_t cp_key_os_version_t;
typedef uint16_t cp_key_revision_t;
typedef uint32_t crypto_flags_t;
```

### cp_key_class_t
A protection class.
```c
typedef uint32_t cp_key_class_t;
```
For possible values, see [Protection Classes](#protection-classes).

### cp_key_os_version_t
An OS version and build number.
```c
typedef uint32_t cp_key_os_version_t;
```
This type stores an OS version and build number as follows:

- Two bytes for the major version number as an unsigned integer
- Two bytes for the minor version letter as an ASCII character
- Four bytes for the build number as an unsigned integer

For example, to store the build number 18A391:

1. Store the number 18 (0x12) in the highest two bytes, yielding 0x12000000.
2. Store the character A (0x41) in the next two bytes, yielding 0x12410000.
3. Store the number 391 (0x0187) in the lowest four bytes, yielding 0x12410187.

### cp_key_revision_t
A version number for an encryption key.
```c
typedef uint16_t cp_key_revision_t;
```

### crypto_flags_t
Flags used by an encryption state.
```c
typedef uint32_t crypto_flags_t;
```
These flags are used by the `cpflags` field of `wrapped_crypto_state_t` and `wrapped_meta_crypto_state_t`. There are currently none defined.

## Protection Classes

Constants that indicate the data protection class of a file.

```c
#define PROTECTION_CLASS_DIR_NONE 0
#define PROTECTION_CLASS_A 1
#define PROTECTION_CLASS_B 2
#define PROTECTION_CLASS_C 3
#define PROTECTION_CLASS_D 4
#define PROTECTION_CLASS_F 6
#define PROTECTION_CLASS_M 14
#define CP_EFFECTIVE_CLASSMASK 0x0000001f
```

These values are used by the `persistent_class` field of `wrapped_meta_crypto_state_t`.

For more information about protection classes, see iOS Security Guide and FileProtectionType.

### PROTECTION_CLASS_DIR_NONE
Directory default.
```c
#define PROTECTION_CLASS_DIR_NONE 0
```
This protection class is used only on devices running iOS.

Files with this protection class use their containing directory's default protection class, which is set by the `default_protection_class` field of `j_inode_val_t`.

### PROTECTION_CLASS_A
Complete protection.
```c
#define PROTECTION_CLASS_A 1
```
This value corresponds to `FileProtectionType.complete`.

### PROTECTION_CLASS_B
Protected unless open.
```c
#define PROTECTION_CLASS_B 2
```
This value corresponds to `FileProtectionType.completeUnlessOpen`.

### PROTECTION_CLASS_C
Protected until first user authentication.
```c
#define PROTECTION_CLASS_C 3
```
This value corresponds to `FileProtectionType.completeUntilFirstUserAuthentication`.

### PROTECTION_CLASS_D
No protection.
```c
#define PROTECTION_CLASS_D 4
```
This value corresponds to `FileProtectionType.none`.

### PROTECTION_CLASS_F
No protection with nonpersistent key.
```c
#define PROTECTION_CLASS_F 6
```
The behavior of this protection class is the same as Class D, except the key isn't stored in any persistent way. This protection class is suitable for temporary files that aren't needed after rebooting the device, such as a virtual machine's swap file.

### PROTECTION_CLASS_M
No overview available.
```c
#define PROTECTION_CLASS_M 14
```

### CP_EFFECTIVE_CLASSMASK
The bit mask used to access the protection class.
```c
#define CP_EFFECTIVE_CLASSMASK 0x0000001f
```
All other bits are reserved. Populate those bits with zero when you create a wrapped key, and preserve their value when you modify an existing wrapped key.

## Encryption Identifiers

Encryption state objects whose identifier is always the same.

```c
#define CRYPTO_SW_ID 4
#define CRYPTO_RESERVED_5 5
#define APFS_UNASSIGNED_CRYPTO_ID (~0ULL)
```

### CRYPTO_SW_ID
The identifier of a placeholder encryption state used when software encryption is in use.
```c
#define CRYPTO_SW_ID 4
```
There is no associated encryption key for this encryption state. All the fields of the corresponding `j_crypto_val_t` structure have a value of zero.

### CRYPTO_RESERVED_5
Reserved.
```c
#define CRYPTO_RESERVED_5 5
```
Don't create an encryption state object with this identifier. If you find an object with this identifier in production, file a bug against the Apple File System implementation.

### APFS_UNASSIGNED_CRYPTO_ID
The identifier of a placeholder encryption state used when cloning files.
```c
#define APFS_UNASSIGNED_CRYPTO_ID (~0ULL)
```
As a performance optimization when cloning a file, Apple's implementation sets this placeholder value on the clone and continues to use the original file's encryption state for both that file and its clone. If the clone is modified, a new encryption state object is created for the clone. Creating a new encryption state object is relatively expensive, and usually takes longer than the cloning process.

## kb_locker_t

A keybag.

```c
struct kb_locker {
    uint16_t kl_version;
    uint16_t kl_nkeys;
    uint32_t kl_nbytes;
    uint8_t padding[8];
    keybag_entry_t kl_entries[];
};
typedef struct kb_locker kb_locker_t;

#define APFS_KEYBAG_VERSION 2
```

A keybag stores wrapped encryption keys and information that's needed to unwrap them. The container and each volume have their own keybag.

The container's keybag stores wrapped VEKs and the location of each volume's keybag. A volume's keybag stores wrapped KEKs.

### Fields

#### kl_version
The keybag version.
```c
uint16_t kl_version;
```
The value of this field is `APFS_KEYBAG_VERSION`.

#### kl_nkeys
The number of entries in the keybag.
```c
uint16_t kl_nkeys;
```

#### kl_nbytes
The size, in bytes, of the data stored in the `kl_entries` field.
```c
uint32_t kl_nbytes;
```

#### padding
Reserved.
```c
uint8_t padding[8];
```
Populate this field with zero when you create a new keybag, and preserve its value when you modify an existing keybag. This field is padding.

#### kl_entries
The entries.
```c
keybag_entry_t kl_entries[];
```

### Constants

#### APFS_KEYBAG_VERSION
The first version of the keybag.
```c
#define APFS_KEYBAG_VERSION 2
```
Version one was used during prototyping of Apple File System, and uses an incompatible, undocumented layout. If you find a keybag in production whose version is less than two, file a bug against the Apple File System implementation.

## keybag_entry_t

An entry in a keybag.

```c
struct keybag_entry {
    uuid_t ke_uuid;
    uint16_t ke_tag;
    uint16_t ke_keylen;
    uint8_t padding[4];
    uint8_t ke_keydata[];
};
typedef struct keybag_entry keybag_entry_t;

#define APFS_VOL_KEYBAG_ENTRY_MAX_SIZE 512
#define APFS_FV_PERSONAL_RECOVERY_KEY_UUID "EBC6C064-0000-11AA-AA11-00306543ECAC"
```

### Fields

#### ke_uuid
In a container's keybag, the UUID of a volume; in a volume's keybag, the UUID of a user.
```c
uuid_t ke_uuid;
```

#### ke_tag
A description of the kind of data stored in this keybag entry.
```c
uint16_t ke_tag;
```
For possible values, see [Keybag Tags](#keybag-tags).

#### ke_keylen
The length, in bytes, of the keybag entry's data.
```c
uint16_t ke_keylen;
```
The value of this field must be less than `APFS_VOL_KEYBAG_ENTRY_MAX_SIZE`.

#### padding
Reserved.
```c
uint8_t padding[4];
```
Populate this field with zero when you create a new keybag entry, and preserve its value when you modify an existing entry. This field is padding.

#### ke_keydata
The keybag entry's data.
```c
uint8_t ke_keydata[];
```
The data stored this field depends on the tag and whether this is an entry in a container or volume's keybag, as described in [Keybag Tags](#keybag-tags).

### Constants

#### APFS_VOL_KEYBAG_ENTRY_MAX_SIZE
The largest size, in bytes, of a keybag entry.
```c
#define APFS_VOL_KEYBAG_ENTRY_MAX_SIZE 512
```

#### APFS_FV_PERSONAL_RECOVERY_KEY_UUID
The user UUID used by a keybag record that contains a personal recovery key.
```c
#define APFS_FV_PERSONAL_RECOVERY_KEY_UUID "EBC6C064-0000-11AA-AA11-00306543ECAC"
```
The personal recovery key is generated during the initial volume-encryption process, and it's stored by the user as a paper printout. You use it the same way you use a user's password to unwrap the corresponding KEK.

## media_keybag_t

A keybag, wrapped up as a container-layer object.

```c
struct media_keybag {
    obj_phys_t mk_obj;
    kb_locker_t mk_locker;
};
typedef struct media_keybag media_keybag_t;
```

### Fields

#### mk_obj
The object's header.
```c
obj_phys_t mk_obj;
```

#### mk_locker
The keybag data.
```c
kb_locker_t mk_locker;
```

## Keybag Tags

A description of what kind of information is stored by a keybag entry.

```c
enum {
    KB_TAG_UNKNOWN = 0,
    KB_TAG_RESERVED_1 = 1,
    KB_TAG_VOLUME_KEY = 2,
    KB_TAG_VOLUME_UNLOCK_RECORDS = 3,
    KB_TAG_VOLUME_PASSPHRASE_HINT = 4,
    KB_TAG_WRAPPING_M_KEY = 5,
    KB_TAG_VOLUME_M_KEY = 6,
    KB_TAG_RESERVED_F8 = 0xF8
};
```

### KB_TAG_UNKNOWN
Reserved.
```c
KB_TAG_UNKNOWN = 0
```
This tag never appears on disk. If you find a keybag entry with this tag in production, file a bug against the Apple File System implementation.

This value isn't reserved by Apple; non-Apple implementations of Apple File System can use it in memory. For example, Apple's implementation uses this value as a wildcard that matches any tag.

### KB_TAG_RESERVED_1
Reserved.
```c
KB_TAG_RESERVED_1 = 1
```
Don't create keybag entries with this tag, but preserve any existing entries.

### KB_TAG_VOLUME_KEY
The key data stores a wrapped VEK.
```c
KB_TAG_VOLUME_KEY = 2
```
This tag is valid only in a container's keybag.

### KB_TAG_VOLUME_UNLOCK_RECORDS
In a container's keybag, the key data stores the location of the volume's keybag; in a volume keybag, the key data stores a wrapped KEK.
```c
KB_TAG_VOLUME_UNLOCK_RECORDS = 3
```
This tag is used only on devices running macOS.

The volume's keybag location is stored as an instance of `prange_t`; the data at that location is an instance of `kb_locker_t`.

### KB_TAG_VOLUME_PASSPHRASE_HINT
The key data stores a user's password hint as plain text.
```c
KB_TAG_VOLUME_PASSPHRASE_HINT = 4
```
This tag is valid only in a volume's keybag, and it's used only on devices running macOS.

### KB_TAG_WRAPPING_M_KEY
The key data stores a key that's used to wrap a media key.
```c
KB_TAG_WRAPPING_M_KEY = 5
```
This tag is used only on devices running iOS.

### KB_TAG_VOLUME_M_KEY
The key data stores a key that's used to wrap media keys on this volume.
```c
KB_TAG_VOLUME_M_KEY = 6
```
This tag is used only on devices running iOS.

### KB_TAG_RESERVED_F8
Reserved.
```c
KB_TAG_RESERVED_F8 = 0xF8
```
Don't create keybag entries with this tag, but preserve any existing entries.
# Sealed Volumes

Sealed volumes contain a hash of their file system, which can be compared to their current content to determine whether the volume has been modified after it was sealed, or compared to a known value to determine whether the volume contains the expected content. On a sealed volume, all of the following must be true:

- The volume's role is `APFS_VOL_ROLE_SYSTEM`.
- The `APFS_INCOMPAT_SEALED_VOLUME` flag is set on the volume.
- The `apfs_integrity_meta_oid` field of `apfs_superblock_t` has a nonzero value.
- The `apfs_fext_tree_oid` field of `apfs_superblock_t` has a nonzero value.
- The `BTREE_HASHED` and `BTREE_NOHEADER` flags are set on the B-tree object that stores the volume's file system.

The B-tree that stores the volume's file system also stores a hash of its contents. A hashed B-tree differs from an nonhashed B-tree as follows:

- The `BTREE_HASHED` flag is set on the root node.
- The `BTNODE_HASHED` flag is set on the nonroot nodes.
- The values stored in nonleaf B-trees are instances of `btn_index_node_val_t`, containing the object identifier of the child node and the hash of the child node.

Conceptually, the hashed B-trees used by sealed volumes are similar to Merkle trees. However, unlike Merkle trees, these hashed B-trees store data as well as a hash of that data.

## integrity_meta_phys_t

Integrity metadata for a sealed volume.

```c
struct integrity_meta_phys {
    obj_phys_t im_o;
    uint32_t im_version;
    uint32_t im_flags;
    apfs_hash_type_t im_hash_type;
    uint32_t im_root_hash_offset;
    xid_t im_broken_xid;
    uint64_t im_reserved[9];
} __attribute__((packed));
typedef struct integrity_meta_phys integrity_meta_phys_t;
```

### Fields

#### im_o
The object's header.
```c
obj_phys_t im_o;
```

#### im_version
The version of this data structure.
```c
uint32_t im_version;
```
The value of this field must be one of the constants listed in [Integrity Metadata Version Constants](#integrity-metadata-version-constants).

#### im_flags
The flags used to describe configuration options.
```c
uint32_t im_flags;
```
For the values used in this bit field, see [Integrity Metadata Flags](#integrity-metadata-flags).

This field appears in version 1 and later of this data structure.

#### im_hash_type
The hash algorithm being used.
```c
apfs_hash_type_t im_hash_type;
```
This field appears in version 1 and later of this data structure.

#### im_root_hash_offset
The offset, in bytes, of the root hash relative to the start of this integrity metadata object.
```c
uint32_t im_root_hash_offset;
```
This field appears in version 1 and later of this data structure.

#### im_broken_xid
The identifier of the transaction that unsealed the volume.
```c
xid_t im_broken_xid;
```
When a sealed volume is modified, breaking its seal, that transaction identifier is recorded in this field and the `APFS_SEAL_BROKEN` flag is set. Otherwise, the value of this field is zero.

This field appears in version 1 and later of this data structure.

#### im_reserved
Reserved.
```c
uint64_t im_reserved[9];
```
This field appears in version 2 and later of this data structure.

## Integrity Metadata Version Constants

Version numbers for the integrity metadata structure.

```c
enum {
    INTEGRITY_META_VERSION_INVALID = 0,
    INTEGRITY_META_VERSION_1 = 1,
    INTEGRITY_META_VERSION_2 = 2,
    INTEGRITY_META_VERSION_HIGHEST = INTEGRITY_META_VERSION_2
};
```

These constants are used as the value of the `im_version` field of the `integrity_meta_phys_t` structure.

### INTEGRITY_META_VERSION_INVALID
An invalid version.
```c
INTEGRITY_META_VERSION_INVALID = 0
```

### INTEGRITY_META_VERSION_1
The first version of the structure.
```c
INTEGRITY_META_VERSION_1 = 1
```

### INTEGRITY_META_VERSION_2
The second version of the structure.
```c
INTEGRITY_META_VERSION_2 = 2
```

### INTEGRITY_META_VERSION_HIGHEST
The highest valid version number.
```c
INTEGRITY_META_VERSION_HIGHEST = INTEGRITY_META_VERSION_2
```

## Integrity Metadata Flags

Flags used by integrity metadata.

```c
#define APFS_SEAL_BROKEN (1U << 0)
```

These flags are used by the `im_flags` field of `integrity_meta_phys_t`.

### APFS_SEAL_BROKEN
The volume was modified after being sealed, breaking its seal.
```c
#define APFS_SEAL_BROKEN (1U << 0)
```
If this flag is set, the `im_broken_xid` field of `integrity_meta_phys_t` contains the transaction identifier for the modification that broke the seal.

## apfs_hash_type_t

Constants used to identify hash algorithms.

```c
typedef enum {
    APFS_HASH_INVALID = 0,
    APFS_HASH_SHA256 = 0x1,
    APFS_HASH_SHA512_256 = 0x2,
    APFS_HASH_SHA384 = 0x3,
    APFS_HASH_SHA512 = 0x4,
    APFS_HASH_MIN = APFS_HASH_SHA256,
    APFS_HASH_MAX = APFS_HASH_SHA512,
    APFS_HASH_DEFAULT = APFS_HASH_SHA256,
} apfs_hash_type_t;

#define APFS_HASH_CCSHA256_SIZE 32
#define APFS_HASH_CCSHA512_256_SIZE 32
#define APFS_HASH_CCSHA384_SIZE 48
#define APFS_HASH_CCSHA512_SIZE 64
#define APFS_HASH_MAX_SIZE 64
```

These constants are used as the value of the `im_hash_type` field of the `integrity_meta_phys_t` structure. The corresponding hash size is used as the value of the `hash_size` field of the `j_file_data_hash_val_t` structure.

### APFS_HASH_INVALID
An invalid hash algorithm.
```c
APFS_HASH_INVALID = 0
```

### APFS_HASH_SHA256
The SHA-256 variant of Secure Hash Algorithm 2.
```c
APFS_HASH_SHA256 = 0x1
```

### APFS_HASH_SHA512_256
The SHA-512/256 variant of Secure Hash Algorithm 2.
```c
APFS_HASH_SHA512_256 = 0x2
```

### APFS_HASH_SHA384
The SHA-384 variant of Secure Hash Algorithm 2.
```c
APFS_HASH_SHA384 = 0x3
```

### APFS_HASH_SHA512
The SHA-512 variant of Secure Hash Algorithm 2.
```c
APFS_HASH_SHA512 = 0x4
```

### APFS_HASH_MIN
The smallest valid value for identifying a hash algorithm.
```c
APFS_HASH_MIN = APFS_HASH_SHA256
```

### APFS_HASH_MAX
The largest valid value for identifying a hash algorithm.
```c
APFS_HASH_MAX = APFS_HASH_SHA512
```

### APFS_HASH_DEFAULT
The default hash algorithm.
```c
APFS_HASH_DEFAULT = APFS_HASH_SHA256
```

### Hash Size Constants

#### APFS_HASH_CCSHA256_SIZE
The size of a SHA-256 hash.
```c
#define APFS_HASH_CCSHA256_SIZE 32
```

#### APFS_HASH_CCSHA512_256_SIZE
The size of a SHA-512/256 hash.
```c
#define APFS_HASH_CCSHA512_256_SIZE 32
```

#### APFS_HASH_CCSHA384_SIZE
The size of a SHA-384 hash.
```c
#define APFS_HASH_CCSHA384_SIZE 48
```

#### APFS_HASH_CCSHA512_SIZE
The size of a SHA-512 hash.
```c
#define APFS_HASH_CCSHA512_SIZE 64
```

#### APFS_HASH_MAX_SIZE
The maximum valid hash size.
```c
#define APFS_HASH_MAX_SIZE 64
```
This value is the same as `BTREE_NODE_HASH_SIZE_MAX`.

## fext_tree_key_t

The key half of a record from a file extent tree.

```c
struct fext_tree_key {
    uint64_t private_id;
    uint64_t logical_addr;
} __attribute__((packed));
typedef struct fext_tree_key fext_tree_key_t;
```

### Fields

#### private_id
The object identifier of the file.
```c
uint64_t private_id;
```
This value corresponds the object identifier portion of the `obj_id_and_type` field of `j_key_t`.

#### logical_addr
The offset within the file's data, in bytes, for the data stored in this extent.
```c
uint64_t logical_addr;
```

## fext_tree_val_t

The value half of a record from a file extent tree.

```c
struct fext_tree_val {
    uint64_t len_and_flags;
    uint64_t phys_block_num;
} __attribute__((packed));
typedef struct fext_tree_val fext_tree_val_t;
```

### Fields

#### len_and_flags
A bit field that contains the length of the extent and its flags.
```c
uint64_t len_and_flags;
```
The extent's length is a `uint64_t` value, accessed as `len_and_kind & J_FILE_EXTENT_LEN_MASK`, and measured in bytes. The length must be a multiple of the block size defined by the `nx_block_size` field of `nx_superblock_t`. The extent's flags are accessed as `(len_and_kind & J_FILE_EXTENT_FLAG_MASK) >> J_FILE_EXTENT_FLAG_SHIFT`.

There are currently no flags defined.

#### phys_block_num
The physical block address that the extent starts at.
```c
uint64_t phys_block_num;
```

## j_file_info_key_t

The key half of a file-info record.

```c
struct j_file_info_key {
    j_key_t hdr;
    uint64_t info_and_lba;
} __attribute__((packed));
typedef struct j_file_info_key j_file_info_key_t;

#define J_FILE_INFO_LBA_MASK 0x00ffffffffffffffULL
#define J_FILE_INFO_TYPE_MASK 0xff00000000000000ULL
#define J_FILE_INFO_TYPE_SHIFT 56
```

### Fields

#### hdr
The record's header.
```c
j_key_t hdr;
```
The object identifier in the header is the file-system object's identifier. The type in the header is always `APFS_TYPE_FILE_INFO`.

#### info_and_lba
A bit field that contains the address and other information.
```c
uint64_t info_and_lba;
```
The address is a `paddr_t` value accessed as `info_and_lba & J_FILE_INFO_LBA_MASK`. The type is a `j_obj_file_info_type` value accessed as `(info_and_lba & J_FILE_INFO_TYPE_MASK) >> J_FILE_INFO_TYPE_SHIFT`.

### Constants

#### J_FILE_INFO_LBA_MASK
The bit mask used to access file-info addresses.
```c
#define J_FILE_INFO_LBA_MASK 0x00ffffffffffffffULL
```

#### J_FILE_INFO_TYPE_MASK
The bit mask used to access file-info types.
```c
#define J_FILE_INFO_TYPE_MASK 0xff00000000000000ULL
```

#### J_FILE_INFO_TYPE_SHIFT
The bit shift used to access file-info types.
```c
#define J_FILE_INFO_TYPE_SHIFT 56
```

## j_file_info_val_t

The value half of a file-info record.

```c
struct j_file_info_val {
    union {
        j_file_data_hash_val_t dhash;
    };
} __attribute__((packed));
typedef struct j_file_info_val j_file_info_val_t;
```

Use the type stored in the `j_file_info_key_t` half of this record to determine which of the union's fields to use.

### Fields

#### dhash
A hash of the file data.
```c
j_file_data_hash_val_t dhash;
```
Use this field of the union if the type stored in the `info_and_lba` field of `j_file_info_val_t` is `APFS_FILE_INFO_DATA_HASH`.

## j_obj_file_info_type

The type of a file-info record.

```c
typedef enum {
    APFS_FILE_INFO_DATA_HASH = 1,
} j_obj_file_info_type;
```

These values are used by the `info_and_lba` field of `j_file_info_key_t`, to indicate how to interpret the data in the corresponding `j_file_info_val_t`.

### APFS_FILE_INFO_DATA_HASH
The file-info record contains a hash of file data.
```c
APFS_FILE_INFO_DATA_HASH = 1
```

## j_file_data_hash_val_t

A hash of file data.

```c
struct j_file_data_hash_val {
    uint16_t hashed_len;
    uint8_t hash_size;
    uint8_t hash[0];
} __attribute__((packed));
typedef struct j_file_data_hash_val j_file_data_hash_val_t;
```

### Fields

#### hashed_len
The length, in blocks, of the data segment that was hashed.
```c
uint16_t hashed_len;
```

#### hash_size
The length, in bytes, of the hash data.
```c
uint8_t hash_size;
```
The value of this field must match the constant that corresponds to the hash algorithm specified in the `im_hash_type` field of `integrity_meta_phys_t`. For a list of algorithms and hash sizes, see `apfs_hash_type_t`.

#### hash
The hash data.
```c
uint8_t hash[0];
```

# Space Manager

The space manager allocates and frees blocks where objects and file data can be stored. There's exactly one instance of this structure in a container.

## chunk_info_t

```c
struct chunk_info {
    uint64_t ci_xid;
    uint64_t ci_addr;
    uint32_t ci_block_count;
    uint32_t ci_free_count;
    paddr_t ci_bitmap_addr;
};
typedef struct chunk_info chunk_info_t;
```

### Fields

#### ci_xid
```c
uint64_t ci_xid;
```

#### ci_addr
```c
uint64_t ci_addr;
```

#### ci_block_count
```c
uint32_t ci_block_count;
```

#### ci_free_count
```c
uint32_t ci_free_count;
```

#### ci_bitmap_addr
```c
paddr_t ci_bitmap_addr;
```

## chunk_info_block

A block that contains an array of chunk-info structures.

```c
struct chunk_info_block {
    obj_phys_t cib_o;
    uint32_t cib_index;
    uint32_t cib_chunk_info_count;
    chunk_info_t cib_chunk_info[];
};
typedef struct chunk_info_block chunk_info_block_t;
```

### Fields

#### cib_o
```c
obj_phys_t cib_o;
```

#### cib_index
```c
uint32_t cib_index;
```

#### cib_chunk_info_count
```c
uint32_t cib_chunk_info_count;
```

#### cib_chunk_info
```c
chunk_info_t cib_chunk_info[];
```

## cib_addr_block

A block that contains an array of chunk-info block addresses.

```c
struct cib_addr_block {
    obj_phys_t cab_o;
    uint32_t cab_index;
    uint32_t cab_cib_count;
    paddr_t cab_cib_addr[];
};
typedef struct cib_addr_block cib_addr_block_t;
```

### Fields

#### cab_o
```c
obj_phys_t cab_o;
```

#### cab_index
```c
uint32_t cab_index;
```

#### cab_cib_count
```c
uint32_t cab_cib_count;
```

#### cab_cib_addr
```c
paddr_t cab_cib_addr[];
```

## spaceman_free_queue_entry_t

```c
struct spaceman_free_queue_entry {
    spaceman_free_queue_key_t sfqe_key;
    spaceman_free_queue_val_t sfqe_count;
};
typedef struct spaceman_free_queue_entry spaceman_free_queue_entry_t;

typedef uint64_t spaceman_free_queue_val_t;
```

### Fields

#### sfqe_key
```c
spaceman_free_queue_key_t sfqe_key;
```

#### sfqe_count
```c
spaceman_free_queue_val_t sfqe_count;
```

## spaceman_free_queue_key_t

```c
struct spaceman_free_queue_key {
    xid_t sfqk_xid;
    paddr_t sfqk_paddr;
};
typedef struct spaceman_free_queue_key spaceman_free_queue_key_t;
```

### Fields

#### sfqk_xid
```c
xid_t sfqk_xid;
```

#### sfqk_paddr
```c
paddr_t sfqk_paddr;
```

## spaceman_free_queue_t

```c
struct spaceman_free_queue {
    uint64_t sfq_count;
    oid_t sfq_tree_oid;
    xid_t sfq_oldest_xid;
    uint16_t sfq_tree_node_limit;
    uint16_t sfq_pad16;
    uint32_t sfq_pad32;
    uint64_t sfq_reserved;
};
typedef struct spaceman_free_queue spaceman_free_queue_t;
```

### Fields

#### sfq_count
```c
uint64_t sfq_count;
```

#### sfq_tree_oid
```c
oid_t sfq_tree_oid;
```

#### sfq_oldest_xid
```c
xid_t sfq_oldest_xid;
```

#### sfq_tree_node_limit
```c
uint16_t sfq_tree_node_limit;
```

#### sfq_pad16
```c
uint16_t sfq_pad16;
```

#### sfq_pad32
```c
uint32_t sfq_pad32;
```

#### sfq_reserved
```c
uint64_t sfq_reserved;
```

## spaceman_device_t

```c
struct spaceman_device {
    uint64_t sm_block_count;
    uint64_t sm_chunk_count;
    uint32_t sm_cib_count;
    uint32_t sm_cab_count;
    uint64_t sm_free_count;
    uint32_t sm_addr_offset;
    uint32_t sm_reserved;
    uint64_t sm_reserved2;
};
typedef struct spaceman_device spaceman_device_t;
```

### Fields

#### sm_block_count
```c
uint64_t sm_block_count;
```

#### sm_chunk_count
```c
uint64_t sm_chunk_count;
```

#### sm_cib_count
```c
uint32_t sm_cib_count;
```

#### sm_cab_count
```c
uint32_t sm_cab_count;
```

#### sm_free_count
```c
uint64_t sm_free_count;
```

#### sm_addr_offset
```c
uint32_t sm_addr_offset;
```

#### sm_reserved
```c
uint32_t sm_reserved;
```

#### sm_reserved2
```c
uint64_t sm_reserved2;
```

## spaceman_allocation_zone_boundaries_t

```c
struct spaceman_allocation_zone_boundaries {
    uint64_t saz_zone_start;
    uint64_t saz_zone_end;
};
typedef struct spaceman_allocation_zone_boundaries spaceman_allocation_zone_boundaries_t;
```

### Fields

#### saz_zone_start
```c
uint64_t saz_zone_start;
```

#### saz_zone_end
```c
uint64_t saz_zone_end;
```

## spaceman_allocation_zone_info_phys_t

```c
struct spaceman_allocation_zone_info_phys {
    spaceman_allocation_zone_boundaries_t saz_current_boundaries;
    spaceman_allocation_zone_boundaries_t saz_previous_boundaries[SM_ALLOCZONE_NUM_PREVIOUS_BOUNDARIES];
    uint16_t saz_zone_id;
    uint16_t saz_previous_boundary_index;
    uint32_t saz_reserved;
};
typedef struct spaceman_allocation_zone_info_phys spaceman_allocation_zone_info_phys_t;

#define SM_ALLOCZONE_INVALID_END_BOUNDARY 0
#define SM_ALLOCZONE_NUM_PREVIOUS_BOUNDARIES 7
```

### Fields

#### saz_current_boundaries
```c
spaceman_allocation_zone_boundaries_t saz_current_boundaries;
```

#### saz_previous_boundaries
```c
spaceman_allocation_zone_boundaries_t saz_previous_boundaries[SM_ALLOCZONE_NUM_PREVIOUS_BOUNDARIES];
```

#### saz_zone_id
```c
uint16_t saz_zone_id;
```

#### saz_previous_boundary_index
```c
uint16_t saz_previous_boundary_index;
```

#### saz_reserved
```c
uint32_t saz_reserved;
```

### Constants

#### SM_ALLOCZONE_INVALID_END_BOUNDARY
```c
#define SM_ALLOCZONE_INVALID_END_BOUNDARY 0
```

#### SM_ALLOCZONE_NUM_PREVIOUS_BOUNDARIES
```c
#define SM_ALLOCZONE_NUM_PREVIOUS_BOUNDARIES 7
```

## spaceman_datazone_info_phys_t

```c
struct spaceman_datazone_info_phys {
    spaceman_allocation_zone_info_phys_t sdz_allocation_zones[SD_COUNT][SM_DATAZONE_ALLOCZONE_COUNT];
};
typedef struct spaceman_datazone_info_phys spaceman_datazone_info_phys_t;

#define SM_DATAZONE_ALLOCZONE_COUNT 8
```

### Fields

#### sdz_allocation_zones
```c
spaceman_allocation_zone_info_phys_t sdz_allocation_zones[SD_COUNT][SM_DATAZONE_ALLOCZONE_COUNT];
```

### Constants

#### SM_DATAZONE_ALLOCZONE_COUNT
```c
#define SM_DATAZONE_ALLOCZONE_COUNT 8
```

## spaceman_phys_t

```c
struct spaceman_phys {
    obj_phys_t sm_o;
    uint32_t sm_block_size;
    uint32_t sm_blocks_per_chunk;
    uint32_t sm_chunks_per_cib;
    uint32_t sm_cibs_per_cab;
    spaceman_device_t sm_dev[SD_COUNT];
    uint32_t sm_flags;
    uint32_t sm_ip_bm_tx_multiplier;
    uint64_t sm_ip_block_count;
    uint32_t sm_ip_bm_size_in_blocks;
    uint32_t sm_ip_bm_block_count;
    paddr_t sm_ip_bm_base;
    paddr_t sm_ip_base;
    uint64_t sm_fs_reserve_block_count;
    uint64_t sm_fs_reserve_alloc_count;
    spaceman_free_queue_t sm_fq[SFQ_COUNT];
    uint16_t sm_ip_bm_free_head;
    uint16_t sm_ip_bm_free_tail;
    uint32_t sm_ip_bm_xid_offset;
    uint32_t sm_ip_bitmap_offset;
    uint32_t sm_ip_bm_free_next_offset;
    uint32_t sm_version;
    uint32_t sm_struct_size;
    spaceman_datazone_info_phys_t sm_datazone;
};
typedef struct spaceman_phys spaceman_phys_t;

#define SM_FLAG_VERSIONED 0x00000001
```

### Fields

#### sm_o
```c
obj_phys_t sm_o;
```

#### sm_block_size
```c
uint32_t sm_block_size;
```

#### sm_blocks_per_chunk
```c
uint32_t sm_blocks_per_chunk;
```

#### sm_chunks_per_cib
```c
uint32_t sm_chunks_per_cib;
```

#### sm_cibs_per_cab
```c
uint32_t sm_cibs_per_cab;
```

#### sm_dev
```c
spaceman_device_t sm_dev[SD_COUNT];
```

#### sm_flags
```c
uint32_t sm_flags;
```

#### sm_ip_bm_tx_multiplier
```c
uint32_t sm_ip_bm_tx_multiplier;
```

#### sm_ip_block_count
```c
uint64_t sm_ip_block_count;
```

#### sm_ip_bm_size_in_blocks
```c
uint32_t sm_ip_bm_size_in_blocks;
```

#### sm_ip_bm_block_count
```c
uint32_t sm_ip_bm_block_count;
```

#### sm_ip_bm_base
```c
paddr_t sm_ip_bm_base;
```

#### sm_ip_base
```c
paddr_t sm_ip_base;
```

#### sm_fs_reserve_block_count
```c
uint64_t sm_fs_reserve_block_count;
```

#### sm_fs_reserve_alloc_count
```c
uint64_t sm_fs_reserve_alloc_count;
```

#### sm_fq
```c
spaceman_free_queue_t sm_fq[SFQ_COUNT];
```

#### sm_ip_bm_free_head
```c
uint16_t sm_ip_bm_free_head;
```

#### sm_ip_bm_free_tail
```c
uint16_t sm_ip_bm_free_tail;
```

#### sm_ip_bm_xid_offset
```c
uint32_t sm_ip_bm_xid_offset;
```

#### sm_ip_bitmap_offset
```c
uint32_t sm_ip_bitmap_offset;
```

#### sm_ip_bm_free_next_offset
```c
uint32_t sm_ip_bm_free_next_offset;
```

#### sm_version
```c
uint32_t sm_version;
```

#### sm_struct_size
```c
uint32_t sm_struct_size;
```

#### sm_datazone
```c
spaceman_datazone_info_phys_t sm_datazone;
```

### Constants

#### SM_FLAG_VERSIONED
```c
#define SM_FLAG_VERSIONED 0x00000001
```

## sfq

```c
enum sfq {
    SFQ_IP = 0,
    SFQ_MAIN = 1,
    SFQ_TIER2 = 2,
    SFQ_COUNT = 3
};
```

### Values

#### SFQ_IP
```c
SFQ_IP = 0
```

#### SFQ_MAIN
```c
SFQ_MAIN = 1
```

#### SFQ_TIER2
```c
SFQ_TIER2 = 2
```

#### SFQ_COUNT
```c
SFQ_COUNT = 3
```

## smdev

```c
enum smdev {
    SD_MAIN = 0,
    SD_TIER2 = 1,
    SD_COUNT = 2
};
```

### Values

#### SD_MAIN
```c
SD_MAIN = 0
```

#### SD_TIER2
```c
SD_TIER2 = 1
```

#### SD_COUNT
```c
SD_COUNT = 2
```

## Chunk Info Block Constants

```c
#define CI_COUNT_MASK 0x000fffff
#define CI_COUNT_RESERVED_MASK 0xfff00000
```

### CI_COUNT_MASK
```c
#define CI_COUNT_MASK 0x000fffff
```

### CI_COUNT_RESERVED_MASK
```c
#define CI_COUNT_RESERVED_MASK 0xfff00000
```

## Internal-Pool Bitmap

```c
#define SPACEMAN_IP_BM_TX_MULTIPLIER 16
#define SPACEMAN_IP_BM_INDEX_INVALID 0xffff
#define SPACEMAN_IP_BM_BLOCK_COUNT_MAX 0xfffe
```

### SPACEMAN_IP_BM_TX_MULTIPLIER
```c
#define SPACEMAN_IP_BM_TX_MULTIPLIER 16
```

### SPACEMAN_IP_BM_INDEX_INVALID
```c
#define SPACEMAN_IP_BM_INDEX_INVALID 0xffff
```

### SPACEMAN_IP_BM_BLOCK_COUNT_MAX
```c
#define SPACEMAN_IP_BM_BLOCK_COUNT_MAX 0xfffe
```
# Reaper

The reaper is a mechanism that allows large objects to be deleted over a period spanning multiple transactions. There's exactly one instance of this structure in a container.

## nx_reaper_phys_t

```c
struct nx_reaper_phys {
    obj_phys_t nr_o;
    uint64_t nr_next_reap_id;
    uint64_t nr_completed_id;
    oid_t nr_head;
    oid_t nr_tail;
    uint32_t nr_flags;
    uint32_t nr_rlcount;
    uint32_t nr_type;
    uint32_t nr_size;
    oid_t nr_fs_oid;
    oid_t nr_oid;
    xid_t nr_xid;
    uint32_t nr_nrle_flags;
    uint32_t nr_state_buffer_size;
    uint8_t nr_state_buffer[];
};
typedef struct nx_reaper_phys nx_reaper_phys_t;
```

### Fields

#### nr_o
```c
obj_phys_t nr_o;
```

#### nr_next_reap_id
```c
uint64_t nr_next_reap_id;
```

#### nr_completed_id
```c
uint64_t nr_completed_id;
```

#### nr_head
```c
oid_t nr_head;
```

#### nr_tail
```c
oid_t nr_tail;
```

#### nr_flags
```c
uint32_t nr_flags;
```

#### nr_rlcount
```c
uint32_t nr_rlcount;
```

#### nr_type
```c
uint32_t nr_type;
```

#### nr_size
```c
uint32_t nr_size;
```

#### nr_fs_oid
```c
oid_t nr_fs_oid;
```

#### nr_oid
```c
oid_t nr_oid;
```

#### nr_xid
```c
xid_t nr_xid;
```

#### nr_nrle_flags
```c
uint32_t nr_nrle_flags;
```

#### nr_state_buffer_size
```c
uint32_t nr_state_buffer_size;
```

#### nr_state_buffer
```c
uint8_t nr_state_buffer[];
```

## nx_reap_list_phys_t

```c
struct nx_reap_list_phys {
    obj_phys_t nrl_o;
    oid_t nrl_next;
    uint32_t nrl_flags;
    uint32_t nrl_max;
    uint32_t nrl_count;
    uint32_t nrl_first;
    uint32_t nrl_last;
    uint32_t nrl_free;
    nx_reap_list_entry_t nrl_entries[];
};
typedef struct nx_reap_list_phys nx_reap_list_phys_t;
```

### Fields

#### nrl_o
```c
obj_phys_t nrl_o;
```

#### nrl_next
```c
oid_t nrl_next;
```

#### nrl_flags
```c
uint32_t nrl_flags;
```

#### nrl_max
```c
uint32_t nrl_max;
```

#### nrl_count
```c
uint32_t nrl_count;
```

#### nrl_first
```c
uint32_t nrl_first;
```

#### nrl_last
```c
uint32_t nrl_last;
```

#### nrl_free
```c
uint32_t nrl_free;
```

#### nrl_entries
```c
nx_reap_list_entry_t nrl_entries[];
```

## nx_reap_list_entry_t

```c
struct nx_reap_list_entry {
    uint32_t nrle_next;
    uint32_t nrle_flags;
    uint32_t nrle_type;
    uint32_t nrle_size;
    oid_t nrle_fs_oid;
    oid_t nrle_oid;
    xid_t nrle_xid;
};
typedef struct nx_reap_list_entry nx_reap_list_entry_t;
```

### Fields

#### nrle_next
```c
uint32_t nrle_next;
```

#### nrle_flags
```c
uint32_t nrle_flags;
```

#### nrle_type
```c
uint32_t nrle_type;
```

#### nrle_size
```c
uint32_t nrle_size;
```

#### nrle_fs_oid
```c
oid_t nrle_fs_oid;
```

#### nrle_oid
```c
oid_t nrle_oid;
```

#### nrle_xid
```c
xid_t nrle_xid;
```

## Volume Reaper States

```c
enum {
    APFS_REAP_PHASE_START = 0,
    APFS_REAP_PHASE_SNAPSHOTS = 1,
    APFS_REAP_PHASE_ACTIVE_FS = 2,
    APFS_REAP_PHASE_DESTROY_OMAP = 3,
    APFS_REAP_PHASE_DONE = 4
};
```

### APFS_REAP_PHASE_START
```c
APFS_REAP_PHASE_START = 0
```

### APFS_REAP_PHASE_SNAPSHOTS
```c
APFS_REAP_PHASE_SNAPSHOTS = 1
```

### APFS_REAP_PHASE_ACTIVE_FS
```c
APFS_REAP_PHASE_ACTIVE_FS = 2
```

### APFS_REAP_PHASE_DESTROY_OMAP
```c
APFS_REAP_PHASE_DESTROY_OMAP = 3
```

### APFS_REAP_PHASE_DONE
```c
APFS_REAP_PHASE_DONE = 4
```

## Reaper Flags

The flags used for general information about a reaper.

```c
#define NR_BHM_FLAG 0x00000001
#define NR_CONTINUE 0x00000002
```

These flags are used by the `nr_flags` field of `nx_reaper_phys_t`.

### NR_BHM_FLAG
Reserved.
```c
#define NR_BHM_FLAG 0x00000001
```
This flag must always be set.

### NR_CONTINUE
The current object is being reaped.
```c
#define NR_CONTINUE 0x00000002
```

## Reaper List Entry Flags

```c
#define NRLE_VALID 0x00000001
#define NRLE_REAP_ID_RECORD 0x00000002
#define NRLE_CALL 0x00000004
#define NRLE_COMPLETION 0x00000008
#define NRLE_CLEANUP 0x00000010
```

### NRLE_VALID
```c
#define NRLE_VALID 0x00000001
```

### NRLE_REAP_ID_RECORD
```c
#define NRLE_REAP_ID_RECORD 0x00000002
```

### NRLE_CALL
```c
#define NRLE_CALL 0x00000004
```

### NRLE_COMPLETION
```c
#define NRLE_COMPLETION 0x00000008
```

### NRLE_CLEANUP
```c
#define NRLE_CLEANUP 0x00000010
```

## Reaper List Flags

```c
#define NRL_INDEX_INVALID 0xffffffff
```

### NRL_INDEX_INVALID
```c
#define NRL_INDEX_INVALID 0xffffffff
```

## omap_reap_state_t

State used when reaping an object map.

```c
struct omap_reap_state {
    uint32_t omr_phase;
    omap_key_t omr_ok;
};
typedef struct omap_reap_state omap_reap_state_t;
```

The reaper uses the state that's stored in this structure to resume after an interruption.

### Fields

#### omr_phase
The current reaping phase.
```c
uint32_t omr_phase;
```
For the values used in this field, see Object Map Reaper Phases.

#### omr_ok
The key of the most recently freed entry in the object map.
```c
omap_key_t omr_ok;
```
This field allows the reaper to resume after the last entry it processed.

## omap_cleanup_state_t

State used when reaping to clean up deleted snapshots.

```c
struct omap_cleanup_state {
    uint32_t omc_cleaning;
    uint32_t omc_omsflags;
    xid_t omc_sxidprev;
    xid_t omc_sxidstart;
    xid_t omc_sxidend;
    xid_t omc_sxidnext;
    omap_key_t omc_curkey;
};
typedef struct omap_cleanup_state omap_cleanup_state_t;
```

### Fields

#### omc_cleaning
A flag that indicates whether the structure has valid data in it.
```c
uint32_t omc_cleaning;
```
If the value of this field is zero, the structure has been allocated and zeroed, but doesn't yet contain valid data. Otherwise, the structure is valid.

#### omc_omsflags
The flags for the snapshot being deleted.
```c
uint32_t omc_omsflags;
```
The value for this field is the same as the value of the snapshot's `omap_snapshot_t.oms_flags` field.

#### omc_sxidprev
The transaction identifier of the snapshot prior to the snapshots being deleted.
```c
xid_t omc_sxidprev;
```

#### omc_sxidstart
The transaction identifier of the first snapshot being deleted.
```c
xid_t omc_sxidstart;
```

#### omc_sxidend
The transaction identifier of the last snapshot being deleted.
```c
xid_t omc_sxidend;
```

#### omc_sxidnext
The transaction identifier of the snapshot after the snapshots being deleted.
```c
xid_t omc_sxidnext;
```

#### omc_curkey
The key of the next object mapping to consider for deletion.
```c
omap_key_t omc_curkey;
```

## apfs_reap_state_t

```c
struct apfs_reap_state {
    uint64_t last_pbn;
    xid_t cur_snap_xid;
    uint32_t phase;
} __attribute__((packed));
typedef struct apfs_reap_state apfs_reap_state_t;
```

### Fields

#### last_pbn
```c
uint64_t last_pbn;
```

#### cur_snap_xid
```c
xid_t cur_snap_xid;
```

#### phase
```c
uint32_t phase;
```

# Encryption Rolling

## er_state_phys_t

```c
struct er_state_phys {
    er_state_phys_header_t ersb_header;
    uint64_t ersb_flags;
    uint64_t ersb_snap_xid;
    uint64_t ersb_current_fext_obj_id;
    uint64_t ersb_file_offset;
    uint64_t ersb_progress;
    uint64_t ersb_total_blk_to_encrypt;
    oid_t ersb_blockmap_oid;
    uint64_t ersb_tidemark_obj_id;
    uint64_t ersb_recovery_extents_count;
    oid_t ersb_recovery_list_oid;
    uint64_t ersb_recovery_length;
};
typedef struct er_state_phys er_state_phys_t;
```

### Fields

#### ersb_header
```c
er_state_phys_header_t ersb_header;
```

#### ersb_flags
```c
uint64_t ersb_flags;
```

#### ersb_snap_xid
```c
uint64_t ersb_snap_xid;
```

#### ersb_current_fext_obj_id
```c
uint64_t ersb_current_fext_obj_id;
```

#### ersb_file_offset
```c
uint64_t ersb_file_offset;
```

#### ersb_progress
```c
uint64_t ersb_progress;
```

#### ersb_total_blk_to_encrypt
```c
uint64_t ersb_total_blk_to_encrypt;
```

#### ersb_blockmap_oid
```c
oid_t ersb_blockmap_oid;
```

#### ersb_tidemark_obj_id
```c
uint64_t ersb_tidemark_obj_id;
```

#### ersb_recovery_extents_count
```c
uint64_t ersb_recovery_extents_count;
```

#### ersb_recovery_list_oid
```c
oid_t ersb_recovery_list_oid;
```

#### ersb_recovery_length
```c
uint64_t ersb_recovery_length;
```

## er_state_phys_v1_t

Version 1 of the encryption rolling state structure.

```c
struct er_state_phys_v1 {
    er_state_phys_header_t ersb_header;
    uint64_t ersb_flags;
    uint64_t ersb_snap_xid;
    uint64_t ersb_current_fext_obj_id;
    uint64_t ersb_file_offset;
    uint64_t ersb_fext_pbn;
    uint64_t ersb_paddr;
    uint64_t ersb_progress;
    uint64_t ersb_total_blk_to_encrypt;
    uint64_t ersb_blockmap_oid;
    uint32_t ersb_checksum_count;
    uint32_t ersb_reserved;
    uint64_t ersb_fext_cid;
    uint8_t ersb_checksum[0];
};
typedef struct er_state_phys_v1 er_state_phys_v1_t;
```

### Fields

#### ersb_header
```c
er_state_phys_header_t ersb_header;
```

#### ersb_flags
```c
uint64_t ersb_flags;
```

#### ersb_snap_xid
```c
uint64_t ersb_snap_xid;
```

#### ersb_current_fext_obj_id
```c
uint64_t ersb_current_fext_obj_id;
```

#### ersb_file_offset
```c
uint64_t ersb_file_offset;
```

#### ersb_fext_pbn
```c
uint64_t ersb_fext_pbn;
```

#### ersb_paddr
```c
uint64_t ersb_paddr;
```

#### ersb_progress
```c
uint64_t ersb_progress;
```

#### ersb_total_blk_to_encrypt
```c
uint64_t ersb_total_blk_to_encrypt;
```

#### ersb_blockmap_oid
```c
uint64_t ersb_blockmap_oid;
```

#### ersb_checksum_count
```c
uint32_t ersb_checksum_count;
```

#### ersb_reserved
```c
uint32_t ersb_reserved;
```

#### ersb_fext_cid
```c
uint64_t ersb_fext_cid;
```

#### ersb_checksum
```c
uint8_t ersb_checksum[0];
```

## er_state_phys_header_t

Header for encryption rolling state structures.

```c
struct er_state_phys_header {
    obj_phys_t ersb_o;
    uint32_t ersb_magic;
    uint32_t ersb_version;
};
typedef struct er_state_phys_header er_state_phys_header_t;
```

### Fields

#### ersb_o
```c
obj_phys_t ersb_o;
```

#### ersb_magic
```c
uint32_t ersb_magic;
```

#### ersb_version
```c
uint32_t ersb_version;
```

## er_phase_t

```c
enum er_phase_enum {
    ER_PHASE_OMAP_ROLL = 1,
    ER_PHASE_DATA_ROLL = 2,
    ER_PHASE_SNAP_ROLL = 3,
};
typedef enum er_phase_enum er_phase_t;
```

### Values

#### ER_PHASE_OMAP_ROLL
```c
ER_PHASE_OMAP_ROLL = 1
```

#### ER_PHASE_DATA_ROLL
```c
ER_PHASE_DATA_ROLL = 2
```

#### ER_PHASE_SNAP_ROLL
```c
ER_PHASE_SNAP_ROLL = 3
```

## er_recovery_block_phys_t

```c
struct er_recovery_block_phys {
    obj_phys_t erb_o;
    uint64_t erb_offset;
    oid_t erb_next_oid;
    uint8_t erb_data[0];
};
typedef struct er_recovery_block_phys er_recovery_block_phys_t;
```

### Fields

#### erb_o
```c
obj_phys_t erb_o;
```

#### erb_offset
```c
uint64_t erb_offset;
```

#### erb_next_oid
```c
oid_t erb_next_oid;
```

#### erb_data
```c
uint8_t erb_data[0];
```

## gbitmap_block_phys_t

```c
struct gbitmap_block_phys {
    obj_phys_t bmb_o;
    uint64_t bmb_field[0];
};
typedef struct gbitmap_block_phys gbitmap_block_phys_t;
```

### Fields

#### bmb_o
```c
obj_phys_t bmb_o;
```

#### bmb_field
```c
uint64_t bmb_field[0];
```

## gbitmap_phys_t

```c
struct gbitmap_phys {
    obj_phys_t bm_o;
    oid_t bm_tree_oid;
    uint64_t bm_bit_count;
    uint64_t bm_flags;
};
typedef struct gbitmap_phys gbitmap_phys_t;
```

### Fields

#### bm_o
```c
obj_phys_t bm_o;
```

#### bm_tree_oid
```c
oid_t bm_tree_oid;
```

#### bm_bit_count
```c
uint64_t bm_bit_count;
```

#### bm_flags
```c
uint64_t bm_flags;
```

## Encryption-Rolling Checksum Block Sizes

```c
enum {
    ER_512B_BLOCKSIZE = 0,
    ER_2KiB_BLOCKSIZE = 1,
    ER_4KiB_BLOCKSIZE = 2,
    ER_8KiB_BLOCKSIZE = 3,
    ER_16KiB_BLOCKSIZE = 4,
    ER_32KiB_BLOCKSIZE = 5,
    ER_64KiB_BLOCKSIZE = 6,
};
```

### ER_512B_BLOCKSIZE
```c
ER_512B_BLOCKSIZE = 0
```

### ER_2KiB_BLOCKSIZE
```c
ER_2KiB_BLOCKSIZE = 1
```

### ER_4KiB_BLOCKSIZE
```c
ER_4KiB_BLOCKSIZE = 2
```

### ER_8KiB_BLOCKSIZE
```c
ER_8KiB_BLOCKSIZE = 3
```

### ER_16KiB_BLOCKSIZE
```c
ER_16KiB_BLOCKSIZE = 4
```

### ER_32KiB_BLOCKSIZE
```c
ER_32KiB_BLOCKSIZE = 5
```

### ER_64KiB_BLOCKSIZE
```c
ER_64KiB_BLOCKSIZE = 6
```

## Encryption Rolling Flags

```c
#define ERSB_FLAG_ENCRYPTING 0x00000001
#define ERSB_FLAG_DECRYPTING 0x00000002
#define ERSB_FLAG_KEYROLLING 0x00000004
#define ERSB_FLAG_PAUSED 0x00000008
#define ERSB_FLAG_FAILED 0x00000010
#define ERSB_FLAG_CID_IS_TWEAK 0x00000020
#define ERSB_FLAG_FREE_1 0x00000040
#define ERSB_FLAG_FREE_2 0x00000080
#define ERSB_FLAG_CM_BLOCK_SIZE_MASK 0x00000F00
#define ERSB_FLAG_CM_BLOCK_SIZE_SHIFT 8
#define ERSB_FLAG_ER_PHASE_MASK 0x00003000
#define ERSB_FLAG_ER_PHASE_SHIFT 12
#define ERSB_FLAG_FROM_ONEKEY 0x00004000
```

### ERSB_FLAG_ENCRYPTING
```c
#define ERSB_FLAG_ENCRYPTING 0x00000001
```

### ERSB_FLAG_DECRYPTING
```c
#define ERSB_FLAG_DECRYPTING 0x00000002
```

### ERSB_FLAG_KEYROLLING
```c
#define ERSB_FLAG_KEYROLLING 0x00000004
```

### ERSB_FLAG_PAUSED
```c
#define ERSB_FLAG_PAUSED 0x00000008
```

### ERSB_FLAG_FAILED
```c
#define ERSB_FLAG_FAILED 0x00000010
```

### ERSB_FLAG_CID_IS_TWEAK
```c
#define ERSB_FLAG_CID_IS_TWEAK 0x00000020
```

### ERSB_FLAG_FREE_1
```c
#define ERSB_FLAG_FREE_1 0x00000040
```

### ERSB_FLAG_FREE_2
```c
#define ERSB_FLAG_FREE_2 0x00000080
```

### ERSB_FLAG_CM_BLOCK_SIZE_MASK
```c
#define ERSB_FLAG_CM_BLOCK_SIZE_MASK 0x00000F00
```

### ERSB_FLAG_CM_BLOCK_SIZE_SHIFT
```c
#define ERSB_FLAG_CM_BLOCK_SIZE_SHIFT 8
```

### ERSB_FLAG_ER_PHASE_MASK
```c
#define ERSB_FLAG_ER_PHASE_MASK 0x00003000
```

### ERSB_FLAG_ER_PHASE_SHIFT
```c
#define ERSB_FLAG_ER_PHASE_SHIFT 12
```

### ERSB_FLAG_FROM_ONEKEY
```c
#define ERSB_FLAG_FROM_ONEKEY 0x00004000
```

## Encryption-Rolling Constants

```c
#define ER_CHECKSUM_LENGTH 8
#define ER_MAGIC 'FLAB'
#define ER_VERSION 1
#define ER_MAX_CHECKSUM_COUNT_SHIFT 16
#define ER_CUR_CHECKSUM_COUNT_MASK 0x0000FFFF
```

### ER_CHECKSUM_LENGTH
```c
#define ER_CHECKSUM_LENGTH 8
```

### ER_MAGIC
```c
#define ER_MAGIC 'FLAB'
```

### ER_VERSION
```c
#define ER_VERSION 1
```

### ER_MAX_CHECKSUM_COUNT_SHIFT
```c
#define ER_MAX_CHECKSUM_COUNT_SHIFT 16
```

### ER_CUR_CHECKSUM_COUNT_MASK
```c
#define ER_CUR_CHECKSUM_COUNT_MASK 0x0000FFFF
```def struct er_state_phys er_state_phys_t;
```

# Fusion

Apple File System supports Fusion drives, which combine a solid-state drive (SSD) with a traditional hard disk drive (HDD) to provide both fast performance and large storage capacity. The Fusion system manages data placement between the two storage tiers automatically.

## fusion_wbc_phys_t

A write-back cache state structure used for Fusion devices.

```c
typedef struct {
    obj_phys_t fwp_objHdr;
    uint64_t fwp_version;
    oid_t fwp_listHeadOid;
    oid_t fwp_listTailOid;
    uint64_t fwp_stableHeadOffset;
    uint64_t fwp_stableTailOffset;
    uint32_t fwp_listBlocksCount;
    uint32_t fwp_reserved;
    uint64_t fwp_usedByRC;
    prange_t fwp_rcStash;
} fusion_wbc_phys_t;
```

Fields
fwp_objHdr
cobj_phys_t fwp_objHdr;
The object's physical header.
fwp_version
cuint64_t fwp_version;
The version of the write-back cache structure.
fwp_listHeadOid
coid_t fwp_listHeadOid;
The object identifier of the head of the write-back cache list.
fwp_listTailOid
coid_t fwp_listTailOid;
The object identifier of the tail of the write-back cache list.
fwp_stableHeadOffset
cuint64_t fwp_stableHeadOffset;
The stable head offset in the write-back cache.
fwp_stableTailOffset
cuint64_t fwp_stableTailOffset;
The stable tail offset in the write-back cache.
fwp_listBlocksCount
cuint32_t fwp_listBlocksCount;
The number of blocks in the write-back cache list.
fwp_reserved
cuint32_t fwp_reserved;
Reserved field for future use.
fwp_usedByRC
cuint64_t fwp_usedByRC;
Space used by the read cache.
fwp_rcStash
cprange_t fwp_rcStash;
Physical range used for read cache stash.
fusion_wbc_list_entry_t
An entry in the write-back cache list for Fusion devices.
ctypedef struct {
    paddr_t fwle_wbcLba;
    paddr_t fwle_targetLba;
    uint64_t fwle_length;
} fusion_wbc_list_entry_t;
Fields
fwle_wbcLba
cpaddr_t fwle_wbcLba;
The logical block address in the write-back cache (on the SSD).
fwle_targetLba
cpaddr_t fwle_targetLba;
The target logical block address (on the HDD where data will eventually be written).
fwle_length
cuint64_t fwle_length;
The length of the cached data in blocks.
fusion_wbc_list_phys_t
A write-back cache list structure for Fusion devices.
ctypedef struct {
    obj_phys_t fwlp_objHdr;
    uint64_t fwlp_version;
    uint64_t fwlp_tailOffset;
    uint32_t fwlp_indexBegin;
    uint32_t fwlp_indexEnd;
    uint32_t fwlp_indexMax;
    uint32_t fwlp_reserved;
    fusion_wbc_list_entry_t fwlp_listEntries[];
} fusion_wbc_list_phys_t;
This mapping keeps track of data from the hard drive that's cached on the solid-state drive. For read caching, the same data is stored on both the hard drive and the solid-state drive. For write caching, the data is stored on the solid-state drive, but space for the data has been allocated on the hard drive, and the data will eventually be copied to that space.
Fields
fwlp_objHdr
cobj_phys_t fwlp_objHdr;
The object's physical header.
fwlp_version
cuint64_t fwlp_version;
The version of the write-back cache list structure.
fwlp_tailOffset
cuint64_t fwlp_tailOffset;
The offset of the tail in the write-back cache list.
fwlp_indexBegin
cuint32_t fwlp_indexBegin;
The beginning index of valid entries in the list.
fwlp_indexEnd
cuint32_t fwlp_indexEnd;
The ending index of valid entries in the list.
fwlp_indexMax
cuint32_t fwlp_indexMax;
The maximum number of entries that can be stored in the list.
fwlp_reserved
cuint32_t fwlp_reserved;
Reserved field for future use.
fwlp_listEntries
cfusion_wbc_list_entry_t fwlp_listEntries[];
Array of write-back cache list entries.
Address Markers
Constants used to distinguish between storage tiers in Fusion drives.
c#define FUSION_TIER2_DEVICE_BYTE_ADDR 0x4000000000000000ULL
#define FUSION_TIER2_DEVICE_BLOCK_ADDR(_blksize) \
    (FUSION_TIER2_DEVICE_BYTE_ADDR >> __builtin_ctzl(_blksize))
#define FUSION_BLKNO(_fusion_tier2, _blkno, _blksize) \
    ((_fusion_tier2) \
     ? (FUSION_TIER2_DEVICE_BLOCK_ADDR(_blksize) | (_blkno)) \
     : (_blkno))
FUSION_TIER2_DEVICE_BYTE_ADDR
c#define FUSION_TIER2_DEVICE_BYTE_ADDR 0x4000000000000000ULL
Byte address marker used to identify the secondary storage device (HDD) in a Fusion drive.
FUSION_TIER2_DEVICE_BLOCK_ADDR
c#define FUSION_TIER2_DEVICE_BLOCK_ADDR(_blksize) \
    (FUSION_TIER2_DEVICE_BYTE_ADDR >> __builtin_ctzl(_blksize))
Macro to calculate the block address marker for the secondary storage device based on block size.
FUSION_BLKNO
c#define FUSION_BLKNO(_fusion_tier2, _blkno, _blksize) \
    ((_fusion_tier2) \
     ? (FUSION_TIER2_DEVICE_BLOCK_ADDR(_blksize) | (_blkno)) \
     : (_blkno))
Macro to encode a block number with the appropriate tier marker. If _fusion_tier2 is true, the block number is marked as being on the secondary storage device.
fusion_mt_key_t
A key used in the Fusion middle tree.
ctypedef paddr_t fusion_mt_key_t;
The key is a physical address representing the location on the HDD that is being tracked in the cache management system.
fusion_mt_val_t
A value used in the Fusion middle tree to track cached blocks.
ctypedef struct {
    paddr_t fmv_lba;
    uint32_t fmv_length;
    uint32_t fmv_flags;
} fusion_mt_val_t;
Fields
fmv_lba
cpaddr_t fmv_lba;
The logical block address where the data is cached on the SSD.
fmv_length
cuint32_t fmv_length;
The length of the cached data in blocks.
fmv_flags
cuint32_t fmv_flags;
Flags describing the state of the cached data. See Fusion Middle-Tree Flags for possible values.
Fusion Middle-Tree Flags
Flags used to describe the state of cached data in the Fusion middle tree.
c#define FUSION_MT_DIRTY (1 << 0)
#define FUSION_MT_TENANT (1 << 1)
#define FUSION_MT_ALLFLAGS (FUSION_MT_DIRTY | FUSION_MT_TENANT)
FUSION_MT_DIRTY
c#define FUSION_MT_DIRTY (1 << 0)
The cached data on the SSD has been modified and needs to be written back to the HDD.
FUSION_MT_TENANT
c#define FUSION_MT_TENANT (1 << 1)
The cached data is associated with a specific tenant or user.
FUSION_MT_ALLFLAGS
c#define FUSION_MT_ALLFLAGS (FUSION_MT_DIRTY | FUSION_MT_TENANT)
Bitmask of all valid Fusion middle-tree flags.

## Revision History

### 2020-06-22

Added the Sealed Volumes chapter.

Added numerous symbols for sealed volumes, file integrity, and other features.

### 2020-05-15

Added various symbols for volume groups, system volumes, Time Machine support, and other macOS and iOS features.

### 2019-02-07

Corrected the discussion of object identifiers in `j_snap_metadata_val_t`.

### 2019-01-24

Added information about software encryption on macOS in the Encryption chapter.

### 2018-09-17

New document that describes the data structures used for read-only access to Apple File System on unencrypted, non-Fusion storage.

---

## Copyright and Notices

**Apple Inc.**  
Copyright © 2020 Apple Inc.  
All rights reserved.

No part of this publication may be reproduced, stored in a retrieval system, or transmitted, in any form or by any means, mechanical, electronic, photocopying, recording, or otherwise, without prior written permission of Apple Inc., with the following exceptions: Any person is hereby authorized to store documentation on a single computer or device for personal use only and to print copies of documentation for personal use provided that the documentation contains Apple's copyright notice.

Apple Inc.  
One Apple Park Way  
Cupertino, CA 95014  
USA  
408-996-1010

Apple is a trademark of Apple Inc., registered in the U.S. and other countries.

APPLE MAKES NO WARRANTY OR REPRESENTATION, EITHER EXPRESS OR IMPLIED, WITH RESPECT TO THIS DOCUMENT, ITS QUALITY, ACCURACY, MERCHANTABILITY, OR FITNESS FOR A PARTICULAR PURPOSE. AS A RESULT, THIS DOCUMENT IS PROVIDED "AS IS," AND YOU, THE READER, ARE ASSUMING THE ENTIRE RISK AS TO ITS QUALITY AND ACCURACY.

IN NO EVENT WILL APPLE BE LIABLE FOR DIRECT, INDIRECT, SPECIAL, INCIDENTAL, OR CONSEQUENTIAL DAMAGES RESULTING FROM ANY DEFECT, ERROR OR INACCURACY IN THIS DOCUMENT, even if advised of the possibility of such damages.

Some jurisdictions do not allow the exclusion of implied warranties or liability, so the above exclusion may not apply to you.

---

*2020-06-22 | Copyright © 2020 Apple Inc. All Rights Reserved.*