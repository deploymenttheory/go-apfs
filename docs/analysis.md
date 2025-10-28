APFS Advent Challenge 2022

2022 APFS Advent Challenge
Sunday, November 27, 2022
As an exercise in self-discipline, I’ve decided to get an early start on my 2023 New Year’s resolution of writing more and sharing what research I can with the community. As a sort of Digital Forensics Advent Calendar, I’m going to attempt to publish a daily series of informative blog posts detailing internals of Apple’s APFS file system.

Why APFS?
I’ve chosen the topic of APFS for three main reasons:

APFS has been Apple’s file system of choice for all of its devices, including Macs, since 2017, yet most of the resources that are available are somewhat lacking in completeness and correctness. Even Apple’s own Apple File System Reference document contains several errors and glaring omissions.

APFS is a fairly unique file system with many interesting design decisions. It’s also complex enough to justify a 24-part blog series.

Over the last five years, I’ve written three separate APFS parsing implementations, so it’s a topic that I am well versed in and qualified to discuss.

The Rules
All good challenges needs rules in order to keep us honest. Luckily for me, I’m challenging myself to this task, so I get to set the requirements.

I will publish a short blog post every day between December 1st and 24th every weekday in December. While I am aware that the Christian season of Advent technically starts on November 27th and ends on Christmas, I am not doing this as a part of any religious observance and most modern Advent calendars start on December 1st anyway. Besides, I only just had this idea and haven’t had time to plan out my topics.

As long as it’s before the clock strikes midnight on a given day somewhere on the planet Earth, it counts. Since I will have to limit myself to writing after work hours and teaching my night classes, I am not limiting myself to my own timezone. This challenge is supposed to be rewarding, not stressful.

For each failed deadline, I will donate $100 to a Ukrainian relief fund. As a result of the unprovoked war of aggression and the continued occupation of Ukraine, the people of Ukraine are suffering loss at a massive scale. The intentional targeting and destruction of Ukrainian civilian infrastructure and power generation capabilities as winter sets in has only increased the suffering of the civilian population. For each day that I do not meet my self imposed deadline of publishing, I will donate additional money towards humanitarian relief for the people of Ukraine. If I do meet my publishing goals (or even if I don’t), and it is within your means, I would ask that you too consider making your own donation this holiday season.

The Posts
[Day 1 - Anatomy of An APFS Object](#2022-apfs-advent-challenge-day-1---anatomy-of-an-object)
[Day 2 - Kinds of APFS Objects](#2022-apfs-advent-challenge-day-2---kinds-of-objects)
[Day 3 - APFS Containers](#2022-apfs-advent-challenge-day-3---containers)
[Day 4 - NX Superblock Objects](#2022-apfs-advent-challenge-day-4---nx-superblock-objects)
[Day 5 - Checkpoint Maps and Ephemeral Objects](#2022-apfs-advent-challenge-day-5---checkpoint-maps-and-ephemeral-objects)
[Day 6 - B-Trees (Part 1)](#2022-apfs-advent-challenge-day-6---b-trees-part-1)
[Day 7 - B-Trees (Part 2)](#2022-apfs-advent-challenge-day-7---b-trees-part-2)
[Day 8 - Object Maps](#2022-apfs-advent-challenge-day-8---object-maps)
Day 9 - Volume Superblock Objects
Day 10 - $100 Donation
Day 11 - File System Trees
Day 12 - Inode and Directory Records
Day 13 - Data Streams
Day 14 - Sealed Volumes
Day 15 - Keybags
Day 16 - Wrapped Keys
Day 17 - Blazingly Fast Checksums with SIMD
Day 18 - Decryption
Day 19 - $100 Donation
Day 20 - Snapshot Metadata
Day 21 - Fusion Containers
Day 22 - Retrospective

2022 APFS Advent Challenge Day 1 - Anatomy of an Object
Thursday, December 1, 2022
APFS is a copy-on-write file system, consisting of a set of immutable objects that are the fundamental building blocks of the file system’s design. APFS objects are made up of one or more fixed-size blocks. Block sizes are configurable at the time of formatting a new container. Valid block sizes are any power-of-two sized value between 4 KiB and 64 KiB of data, and must always be an integer multiple of the block size of the underlying storage device. At the time of this writing, the default (and thus most common) block size is 4 KiB.

Object Headers
While some objects are headerless, most begin with an obj_phys_t structure as their header. Like all APFS on-disk objects, this structure is stored with little-endian values.

#define MAX_CKSUM_SIZE 8

typedef uint64_t oid_t;
typedef uint64_t xid_t;

typedef struct obj_phys {
    uint8_t o_cksum[MAX_CKSUM_SIZE]; // 0x00
    oid_t o_oid;                     // 0x08
    xid_t o_xid;                     // 0x10
    uint32_t o_type;                 // 0x18
    uint32_t o_subtype;              // 0x1C
} obj_phys_t;                        // 0x20
The object headers are immediately followed by type-specific data, and any remaining space between the object’s data and the end of its last block is always zeroed and is reserved for future use.

Checksum
The integrity of an on-disk APFS object’s data can be verified by calculating a Fletcher-64 checksum of all the object’s data after the first 8 bytes. This checksum can be compared with the value of the o_cksum field in the object’s header. If these values do not match, then the object is either only partially flushed to disk or is otherwise corrupted. Note that like most uses of checksums, this is not a security feature, but is only used to detect unintentionally corrupted data.

uint64_t fletcher64(const void* data, size_t size) {
    uint64_t sum1 = 0;
    uint64_t sum2 = 0;

    // Calculate the number of 32-bit words
    size_t words_left = size / sizeof(uint32_t);

    // Interpret the data as a set of 32-bit words
    const uint32_t* words = static_cast<const uint32_t*>(data);

    while (words_left > 0) {
        // Truncate sums after a maximum of 1024 words
        const n = std::min(words_left, 1024);

        // Compute the checksums
        for (size_t i = 0; i < n; i++) {
            sum1 += words[i];
            sum2 += sum1;
        }

        // Calculate the modulo of the sums
        sum1 %= UINT32_MAX;
        sum2 %= UINT32_MAX;

        words_left -= n;
        words += n;
    }

    // Calculate the value needed to be able to get a checksum of zero
    const uint64_t ck_low = UINT32_MAX - ((sum1 + sum2) % UINT32_MAX);
    const uint64_t ck_high = UINT32_MAX - ((sum1 + ck_low) % UINT32_MAX);

    // Combine the sums
    return ck_low | (ck_high << 32);
}
Object and Transaction IDs
Each object has a unique 8-byte object identifier (oid), which is stored in the header’s o_oid field, along with an 8-byte transaction identifier (xid). Most APFS objects are immutable. When a change is made and flushed to disk, an entirely new object is created elsewhere on disk and is assigned the same oid as the original object, but with a higher xid.

Once the updated object has been fully flushed to disk, and all other objects that reference the original object have been updated to reference the newer object, the transaction is considered complete and the original object’s blocks are free to be reused by APFS. While these blocks are not immediately wiped for reuse, the lifetime of unreferenced objects is relatively short.

Types and Subtypes
The remaining two fields in the header encode the object’s type and (optional) subtype identifiers. Each distinct APFS object type is assigned a unique type identifier. With few exceptions, this identifier is stored in the 16 least-significant bits of the o_type field in the header, with the 16 most-significant bits being used for type flags.

The following is a list of all currently-known object types and their identifiers. We will discuss the details of many of them throughout the course of this blog series.

Object Type	Type Identifier	Description	Structure
NX_SUPERBLOCK	0x01	Container Superblock	nx_superblock_t
BTREE	0x02	B-Tree Root Node	btree_node_phys_t
BTREE_NODE	0x03	B-Tree Node	btree_node_phys_t
MTREE	0x04	M-Tree	undocumented type
SPACEMAN	0x05	Space Manager	spaceman_phys_t
SPACEMAN_CAB	0x06	Space Manager Chunk-Info Address Block	cib_addr_block
SPACEMAN_CIB	0x07	Space Manager Chunk-Info Block	chunk_info_block
SPACEMAN_BITMAP	0x08	Space Manager Free-Space Bitmap	raw block of bits
OMAP	0x0b	Object Map	omap_phys_t
CHECKPOINT_MAP	0x0c	Checkpoint Map	checkpoint_map_phys_t
FS	0x0d	Volume	apfs_superblock_t
NX_REAPER	0x11	Reaper	nx_reaper_phys_t
NX_REAP_LIST	0x12	Reaper List	nx_reap_list_phys_t
EFI_JUMPSTART	0x14	EFI Boot Information	nx_efi_jumpstart_t
NX_FUSION_WBC	0x16	Fusion Write-Back Cache State	fusion_wbc_phys_t
NX_FUSION_WBC_LIST	0x17	Fusion Write-Back Cache List	fusion_wbc_list_phys_t
ER_STATE	0x18	Rolling Encryption State	er_state_phys_t
GBITMAP	0x19	General Purpose Bitmap	gbitmap_phys_t
GBITMAP_BLOCK	0x1b	General Purpose Bitmap Block	gbitmap_block_phys_t
ER_RECOVERY_BLOCK	0x1c	Rolling Encryption Recovery State	er_recovery_block_phys_t
SNAP_META_EXT	0x1d	Additional Snapshot Metadata	snap_meta_ext_obj_phys_t
INTEGRITY_META	0x1e	Integrity Metadata	integrity_meta_phys_t
There are three additional known object types that use all 32-bits of the o_type header field and do not contain type flags.

Object Type	Type Identifier	Description	Structure
CONTAINER_KEYBAG	0x7379656b	Container Keybag	media_keybag_t
VOLUME_KEYBAG	0x73636572	Volume Keybag	media_keybag_t
MEDIA_KEYBAG	0x79656b6d	Media Keybag	media_keybag_t
B-Tree objects also contain subtypes, which help identify the specific purpose of the tree. These subtype identifiers are stored in the header’s o_subtype field. The following is a list of known b-tree subtypes and the structures that they map.

Object Subtype	Subtype Identifier	Description	Key Structure	Value Structure
SPACEMAN_FREE_QUEUE	0x09	Space Manager Free-Space Queue	spaceman_free_queue_key_t	spaceman_free_queue_t
EXTENT_LIST_TREE	0x0a	Logical to Physical Mapping of Extents	paddr_t	prange_t
OMAP	0x0b	Object Map	omap_key_t	omap_val_t
FSTREE	0x0e	File-System Record Tree	j_key_t	variable
BLOCKREFTREE	0x0f	Extent Reference Tree	j_phys_ext_key_t	j_phys_ext_val_t
SNAPMETATREE	0x10	Snapshot Metadata Tree	j_key_t	variable
OMAP_SNAPSHOT	0x13	Omap Snapshot Info	xid_t	omap_snapshot_t
FUSION_MIDDLE_TREE	0x15	Tracks Cached SSD Fusion Blocks	fusion_mt_key_t	fusion_mt_val_t
GBITMAP_TREE	0x1a	General Purpose Bitmap Tree	uint64_t	uint64_t
FEXT_TREE	0x1f	File Extents	fext_tree_key_t	fext_tree_val_t
Type Flags
As previously mentioned, when object types are indicated in object headers and other APFS structures, they are usually combined with up to 16 bits of flags that give extra information. The currently defined flags are as follows:

// Object Kind Flags
#define OBJ_VIRTUAL 0x00000000
#define OBJ_EPHEMERAL 0x80000000
#define OBJ_PHYSICAL 0x40000000

// Other Flags
#define OBJ_NOHEADER 0x20000000
#define OBJ_ENCRYPTED 0x10000000
#define OBJ_NONPERSISTENT 0x08000000
The two most-significant bits are used to denote the kind of APFS object: virtual, ephemeral, or physical. All APFS objects fit into one of these three categories. The difference between these will be the subject of tomorrow’s post.

If the OBJ_NOHEADER flag is set, then the object type in question does not start with an obj_phys_t header. These types of objects are rare, and so far I’ve only seen it used for space manager bitmap objects. Note that these objects are different than the headerless objects that are used in sealed volumes, which we will discuss in a future posts.

The OBJ_ENCRYPTED flag denotes an object that is always encrypted on disk, and the OBJ_NONPERSISTENT flag denotes an object that is never written to disk at all (this flag will only be set for ephemeral objects in memory that do not require persistence).

https://jtsylve.blog/post/2022/12/02/Kinds-of-APFS-Objects

2022 APFS Advent Challenge Day 2 - Kinds of Objects
Friday, December 2, 2022
As we discussed in our last post, objects are the fundamental building blocks of APFS. While there are many different object types, each individual object can be one of three kinds: physical, virtual, or ephemeral. While each of these objects can be found on disk, there are differences in their lifetimes as well as the techniques needed to locate them.

Physical Objects
Physical Objects are objects that are always stored at a fixed location on disk. You can think of these kinds of objects as being “owned” and managed directly by the APFS Container itself. (We will discuss containers in an upcoming post in this series.)

Locating a physical object on disk via its object identifier (oid) is trivial, because its oid will always be the same as its starting block number. APFS block numbers are zero-indexed; therefore, we can locate a physical object on disk by multiplying its oid by the block size of the container. For example, if we have an APFS container that is configured with a 0x1000 byte (4 KiB) block size, the physical object with oid 5 will be located at byte offset 0x5000 of the physical storage media.

NOTE: These calculations can be slightly more complicated when dealing with Fusion Containers, but we’ll discuss those details in a future post.

Because physical objects use direct addressing, there can only ever be one version of these objects on disk. If a physical object is copied, an entirely new physical object will be created with a different oid. Once an APFS object is no longer being referenced by any other object, its underlying storage blocks are subject to reuse. If they are ever reused by another physical object, the newer object will have the same oid, but can be differentiated from the original object given that it will have a higher transaction identifier (xid).

Virtual Objects
Virtual Objects represent the majority of all objects in APFS. They are not stored at fixed locations on disk, and do not have direct relationships between their oid and storage location. All virtual objects are “owned” by an Object Map (OMAP). OMAPs are tree-like structures that are used to manage the lifetimes of virtual objects and are used to lookup the their block-storage location on physical. You can look forward to a detailed discussion of object maps in a future post in this series. Suffice it to say that OMAPs are key/value stores that allow you to efficiently locate a virtual object’s blocks by using it’s oid and xid as keys.

This indirect mapping of virtual objects allows for quite a bit of flexibility. Because they are not limited to a fixed block-address, virtual objects can be relocated on disk at any time by updating their storage location in their object map.

APFS is a copy-on-write (CoW) file system, which means that the contents of APFS virtual objects are immutable on disk. Changes to virtual objects are handled by the creation of an entirely new virtual object with the same oid and an updated xid. As newer transactions are flushed to disk, older transactions are invalidated, their objects will no longer be referenced by any active objects, and their block storage may be reused.

Another feature of OMAPs is that they can extend the lifetime of objects from earlier transactions by maintaining references to more than one version of the same object. Since these earlier transactions are still “reachable”, they are considered to be active and will be preserved. This can be very useful for “rolling back” the state of APFS to an earlier point-in-time. The APFS Snapshot feature is built upon this inherent property of OMAPs (more on this in a future post).

Ephemeral Objects
The copy-on-write nature of APFS objects is great for fault tolerance. For example, if power is lost while a transaction is only partially through the process of being written to disk, data corruption can be avoided, because the system can just revert to the latest valid transaction. This completely removes the need for a file system transaction log.

Of course, CoW comes at a cost. There are some objects that need to be updated much too frequently to be flushed to disk as part of a transaction on every change. These frequently-updated objects also tend to not require strict data integrity guarantees or strong fault tolerances. Two prime examples are the objects that track performance counters and garbage collection state.

Some objects require no persistence at all and are only resident in-memory. Since we won’t find those on disk, we don’t need to be concerned with them at the moment. Other’s spend the majority of their lifetimes in-memory when APFS is mounted and are only flushed to disk periodically (or when APFS is cleanly unmounted) for persistence. These types of objects are known as ephemeral.

Like virtual objects, ephemeral objects are not located at fixed locations on disk and there is no direct mapping between oid and storage blocks. Ephemeral objects are “owned” by checkpoint maps, which are responsible for managing their on-disk lifetime and providing translation capabilities between oid and on-disk storage locations (more on these in the future).

Given that we have limited guarantees about when and how often these objects are flushed to disk, ephemeral objects are limited to only those that are not critical to preserving the integrity of users’ data, and that can be reconstructed entirely (or from previous versions) on-demand. Still, Ephemeral objects play an important role in APFS. We will discuss several of the kinds of objects throughout this series.

https://jtsylve.blog/post/2022/12/05/APFS-Containers

2022 APFS Advent Challenge Day 3 - Containers
Monday, December 5, 2022
APFS is a pooled storage, transactional, copy-on-write file system. Its design relies on a core management layer known as the Container. APFS containers consist of a collection of several specialized components: The Space Manager, the Checkpoint Areas, and the Reaper. In today’s post, we will give an overview of APFS containers and these components.

History
Prior to the introduction of APFS, Apple’s primary file system of choice was HFS+. HFS+ is a journaling file system that was introduced by Apple in 1998 as an improvement over its legacy HFS file system.

Like most file systems of its era, each HFS+ volume can only manage the space of a single physical disk partition. While it is possible to have more than one HFS+ volume on a disk, the limitation of “one volume per partition” requires that the storage space for each volume be fixed and pre-allocated. This means that HFS+ volumes that are low on storage space cannot make use of any available free space elsewhere on disk.

In 2012, Apple introduced its hybrid Fusion Drives, which consist of a larger hard disk drive (HDD) combined with a smaller, but faster solid state drive (SSD) in a single package. The HDD is intended to be used as the primary storage device, providing the baseline storage capacity, and the SSD provides faster access to the most recently accessed data by acting as a cache.

This caching logic is not built into the fusion drive hardware. The two drives are presented to the operating system as separate storage devices. HFS+ does not have the ability to span a volume across multiple partitions, and it was not designed to support the desired caching mechanisms.

Rather than massively overhauling HFS+ to support these new capabilities, Apple decided instead to add an additional storage layer, called Core Storage. Core Storage acts as a logical volume manager that has the ability to pool the storage of multiple devices on the same drive into a single, logical volume. It also implements a tiered storage model that allows blocks to be duplicated and cached on Fusion drives. Incidentally, Core Storage also provides the mechanism for the volume-level encryption facilities of File Vault on HFS+ systems. Because HFS+ only sees a single logical volume, these complexities are completely transparent to the file system’s implementation.

Apple introduced APFS in 2017. The design of APFS takes many lessons from both HFS+ and Core Storage, and eliminates the need for both of them.

Space Manager
APFS containers provide pooled and tiered storage capabilities, without the need for a Core Storage layer. It presents one logical view of storage, whose blocks can be shared among multiple volumes without the need for pre-partitioning and pre-allocation of space. As volumes’ storage requirements change over time, blocks are allocated or returned to the container. This allows for quite a bit of flexibility, as you can now have multiple volumes that serve different roles without having to figure out their space requirements ahead of time. For example, you can now have more than one system volume with different versions on macOS installed that can share the same user data volume.

It supports storage devices as small as 1 MiB in size (APFS on a 1.44 MiB HD floppy, anyone?) and has no apparent upper storage limit. It supports the sharing of blocks among as many as 100 volumes (with some limitations). In addition to that hard-coded upper maximum of 100 volumes, APFS requires that there can be no more than one volume per 512 MiB of storage space. This helps limit storage contention and reduces the amount of space needed to maintain file system metadata on-disk.

The Space Manager instance keep track of which blocks across storage tiers are in-use. It also is responsible for the allocation and freeing of blocks for volumes on-demand.

Checkpoint Areas
As mentioned in last Friday’s post, APFS provides fault tolerance by batching together copies of updated objects and committing them to disk in transactions known as checkpoints. This transactional, copy-on-write strategy ensures that there is always at least one valid and complete set of APFS objects on disk. The latest checkpoint may be used as the authoritative source of information and since checkpoints aren’t immediately invalidated, the entire state of APFS can be reverted to an earlier point in time.

APFS containers maintain two distinct checkpoint areas. The Checkpoint Data Area, which is reserved for storage of ephemeral objects, and the Checkpoint Descriptor Area.

The Checkpoint Descriptor Area provides a logically (but not necessarily physically) contiguous area on disk that is reserved to act as a circular buffer to store two types of objects that used to parse information about checkpoints: Checkpoint Map Objects and NX Superblock Objects.

After a checkpoint is flushed to disk, both types of objects are written to the descriptor area. The Checkpoint Map Objects provide a list of all ephemeral objects, their types, and their storage location within the checkpoint data area. A NX Superblock object is written to the descriptor area buffer after the map objects. This superblock is the root object of APFS and serves as the initial source of information about the state of the container in each checkpoint. All other valid objects in APFS are either directly or indirectly reachable from the NX superblock object.

Reaper
Once a checkpoint transaction is successfully flushed to disk, APFS may choose to invalidate the oldest checkpoint. At this point, all newly unreferenced objects are subject to a process of garbage collection, where their blocks can be wiped and returned to the space manager for reuse. The Reaper is responsible for managing this garbage collection process, keeping track of the state of objects so that they may be freed across transactions.

Conclusion
Containers provide the core management layer of APFS using several specialized subsystems. This post gives a general overview of each of these components. Future posts in this series will discuss each of these components in more detail, including information on how to interpret their on-disk structures.

2022 APFS Advent Challenge Day 4 - NX Superblock Objects
Tuesday, December 6, 2022
The NX Superblock Object is a crucial component of APFS. It stores key information about the Container, such as the block size, total number of blocks, supported features, and the object IDs of various trees and other structures used to track and maintain other objects. The on-disk nx_superblock_t structure is used as the root source of information to locate all other objects in the checkpoint. In this post, we will go into detail about this structure as well as discuss methodology that can be used to locate them on-disk.

On-Disk Structures
typedef uint8_t uuid_t[0x10];
typedef int64_t paddr_t;

typedef struct prange {
    paddr_t pr_start_paddr;  // 0x00
    uint64_t pr_block_count; // 0x08
} prange_t;                  // 0x10

#define NX_MAGIC 0x4253584E  // NXSB
#define NX_MAX_FILE_SYSTEMS 100
#define NX_EPH_INFO_COUNT 4
#define NX_NUM_COUNTERS 32

typedef struct nx_superblock {
    obj_phys_t nx_o;                                // 0x00
    uint32_t nx_magic;                              // 0x20
    uint32_t nx_block_size;                         // 0x24
    uint64_t nx_block_count;                        // 0x28
    uint64_t nx_features;                           // 0x30
    uint64_t nx_readonly_compatible_features;       // 0x38
    uint64_t nx_incompatible_features;              // 0x40
    uuid_t nx_uuid;                                 // 0x48
    oid_t nx_next_oid;                              // 0x58
    xid_t nx_next_xid;                              // 0x60
    uint32_t nx_xp_desc_blocks;                     // 0x68
    uint32_t nx_xp_data_blocks;                     // 0x6C
    paddr_t nx_xp_desc_base;                        // 0x70
    paddr_t nx_xp_data_base;                        // 0x78
    uint32_t nx_xp_desc_next;                       // 0x80
    uint32_t nx_xp_data_next;                       // 0x84
    uint32_t nx_xp_desc_index;                      // 0x88
    uint32_t nx_xp_desc_len;                        // 0x8C
    uint32_t nx_xp_data_index;                      // 0x90
    uint32_t nx_xp_data_len;                        // 0x94
    oid_t nx_spaceman_oid;                          // 0x98
    oid_t nx_omap_oid;                              // 0xA0
    oid_t nx_reaper_oid;                            // 0xA8
    uint32_t nx_test_type;                          // 0xB0
    uint32_t nx_max_file_systems;                   // 0xB4
    oid_t nx_fs_oid[NX_MAX_FILE_SYSTEMS];           // 0xB8
    uint64_t nx_counters[NX_NUM_COUNTERS];          // 0x3D8
    prange_t nx_blocked_out_prange;                 // 0x4D8
    oid_t nx_evict_mapping_tree_oid;                // 0x5D8
    uint64_t nx_flags;                              // 0x5E0
    paddr_t nx_efi_jumpstart;                       // 0x5E8
    uuid_t nx_fusion_uuid;                          // 0x5F8
    prange_t nx_keylocker;                          // 0x608
    uint64_t nx_ephemeral_info[NX_EPH_INFO_COUNT];  // 0x618
    oid_t nx_test_oid;                              // 0x638
    oid_t nx_fusion_mt_oid;                         // 0x640
    oid_t nx_fusion_wbc_oid;                        // 0x648
    prange_t nx_fusion_wbc;                         // 0x650
    uint64_t nx_newest_mounted_version;             // 0x660
    prange_t nx_mkb_locker;                         // 0x668
} nx_superblock_t;                                  // 0x678
prange_t
prange_t structures keep track of contiguous ranges of blocks. It is used in various other data structures.

pr_start_addr: The physical address of the first block in the range.
pr_block_count: The number of blocks in the range
nx_superblock_t
nx_superblock_t structures store key information about the Container and act as the root source of information to locate all other objects in the checkpoint. We’ll go in detail of most of these as needed, but below is a brief description of each.

nx_o: The object’s header
nx_magic: A number that can be used to verify that you’re reading an instance of nx_superblock_t.This should always be the value defined by NX_MAGIC
nx_block_size: The logical block size used in the container
nx_block_count: The total number of blocks in the container
nx_features: A bit-field of optional features supported by the container
nx_readonly_compatible_features: A bit-field of optional read-only features supported by the container
nx_incompatible_features: A bit-field of backwards-incompatible features that are in use
nx_next_oid: The next object identifier that will be used by a new virtual or ephemeral object
nx_next_oid: The next transaction identifier that will be used
nx_xp_desc_blocks: Encodes the number of blocks in the Checkpoint Descriptor Area
nx_xp_data_blocks: Encodes the number of blocks in the Checkpoint Data Area
nx_xp_desc_base: Encodes information that can be used to locate the ranges of blocks used by the checkpoint descriptor area
nx_xp_data_base: Encodes information that can be used to locate the ranges of blocks used by the checkpoint data area
nx_xp_desc_next: The next index that will be used in the checkpoint descriptor area
nx_xp_data_next: The next index that will be used in the checkpoint data area
nx_xp_desc_index: The index of the first valid item in the checkpoint descriptor area
nx_xp_desc_len: The number of blocks in the checkpoint descriptor area used by the checkpoint for which this superblock belongs
nx_xp_data_index: The index of the first valid item in the checkpoint data area
nx_xp_data_len: The number of blocks in the checkpoint data area used by the checkpoint for which this superblock belongs
nx_spaceman_oid: The ephemeral object identifier of the container’s Space Manager
nx_omap_oid: The physical object identifier of the container’s Object Map
nx_reaper_oid: The ephemeral object identifier of the container’s Reaper
nx_test_type: Reserved
nx_max_file_systems: The maximum number of file system volumes that can be used with this container
nx_fs_oid: An array of virtual object identifiers for File System Superblock Objects
nx_counters: An array of performance counters
nx_blocked_out_prange: A physical range of blocks where space will not be allocated (used when shrinking a partition)
nx_evict_mapping_tree_oid: The physical object identifier of a tree used to keep track of objects that must be moved out of blocked-out storage. (used when shrinking a partition)
nx_flags: Miscellaneous container flags
nx_efi_jumpstart: The physical object identifier of a tree used to keep track of objects that must be moved out of blocked-out storage.
nx_fusion_uuid: The universally unique identifier of the container’s Fusion set, or zero for non-Fusion containers
nx_keylocker: The location of the container’s keybag.
nx_ephemeral_info: An array of fields used in the management of ephemeral data
nx_test_oid: Reserved
nx_fusion_mt_oid: The physical object identifier of the Fusion Middle Tree or zero on non-fusion drives
nx_fusion_wbc_oid: The ephemeral object identifier of the Fusion write-back cache state or zero on non-fusion drives
nx_fusion_wbc: The blocks used for the Fusion write-back cache area, or zero for non-Fusion drives
nx_newest_mounted_version: Reserved, but generally used to encode the version number of the APFS KEXT with the highest version number that was used to mount this container read/write.
nx_mkb_locker: The blocks used to store the wrapped media key
Locating the NX Superblock
Block 0 of the disk partition always contains a copy of the container’s nx_superblock_t object, but it is not guaranteed to be the most up-to-date version, depending on whether the container was last unmounted cleanly. Rather than relying on a possibly invalid superblock object, we can use the information in this block-zero copy to locate the latest, valid checkpoint.

Step 1: Validating Block Zero
First, it is necessary to determine whether in fact we are dealing with an APFS container in the first place, and (if so) to identify the container’s fixed block size.

Start by reading at least 4 KiB of data from the start of the partition. On an APFS formatted partition, this should always be a valid nx_superblock_t structure.

Validate the object type in the nx_o.o_type field and that the nx_magic field is set to the NX_MAGIC value.

Read the container’s block size from the nx_block_size field. If it is larger than 4 KiB, re-read the block-zero superblock into memory with the correct block size.

Calculate the object’s checksum and validate it against the value in its object header.

If all goes well, we’re in business.

Step 2: Locate the Checkpoint Descriptor Area
NX Superblock objects are stored in the container’s Checkpoint Descriptor Area. In order to locate these superblocks, we must first identify and scan the descriptor area blocks, looking for valid NX Superblock objects. In some cases, the blocks of the descriptor area are all physically contiguous on disk, which means we only have a single range of blocks to scan. In other cases, we may need to locate multiple non-contiguous ranges of blocks and scan them in order.

Read the nx_xp_desc_base field of the block-zero superblock. The most-significant bit (MSB) of this value is a flag, and the remaining 63 least-significant bits (LSBs) contain a physical block address.

If the MSB is unset, the descriptor area consists of only a single range of contiguous blocks. The rest of nx_xp_desc_base contains the block number of the starting block, and the nx_xp_desc_blocks field contains the number of blocks in the area.

Things are a bit more complicated if the MSB is set. The descriptor area is stored non-contiguously and we’ll need to scan multiple ranges. Rather than the starting block, the LSBs of nx_xp_desc_base encode the physical address of a B-Tree Root Node object.

We will discuss B-Trees, and how to parse them later this week, but for now it is only necessary to understand that B-Trees in APFS are essentially just ordered key/value stores.

This particular B-Tree maps uint64_t logical starting offsets inside the checkpoint descriptor area to prange_t physical block ranges. Enumerating through the entries in this B-Tree allows us to identify the order and location of each ranges of blocks in the checkpoint descriptor area.

Step 3: Search the Checkpoint Descriptor Area
As discussed in our last post, the container’s Checkpoint Descriptor Area stores two types of objects: NX Superblocks and Checkpoint Maps. These objects can be differentiated by the o_type member of their object headers.

Search each range of blocks in the descriptor area, looking for NX Superblock objects. There should be more than one, which each superblock representing a specific checkpoint. Validate each superblock as before, and keep track of the valid superblock with the highest transaction identifier (xid). Since NX Superblock objects are the last objects flushed to disk during a checkpoint transaction, this should mean you’ve located the information needed to parse the most up-to-date state of the container. If you run into problems parsing information from a container later down the road, you can always try starting from the NX Superblock with the next-highest xid.

Conclusion
Understanding how to interpret and locate NX Superblock Objects is the first step in parsing APFS. These objects are essential to the process of locating all other objects in a checkpoint. In our next post, we will discuss Checkpoint Maps, and how they can be used to locate ephemeral objects on disk.

2022 APFS Advent Challenge Day 5 - Checkpoint Maps and Ephemeral Objects
Wednesday, December 7, 2022
In our last post, we discussed NX Superblock Objects and how they can be used to locate the Checkpoint Descriptor Area in which they are stored. Today, we will discuss the other type of objects that are stored in the descriptor area, Checkpoint Maps, and how they can be used to find persistent, ephemeral objects on disk.

On-Disk Structures
Each Checkpoint Mapping structure gives information about a single ephemeral object that is stored in the Checkpoint Data Area.

typedef struct checkpoint_mapping {
    uint32_t cpm_type;     // 0x00
    uint32_t cpm_subtype;  // 0x04
    uint32_t cpm_size;     // 0x08
    uint32_t cpm_pad;      // 0x0C
    oid_t cpm_fs_oid;      // 0x10
    oid_t cpm_oid;         // 0x18
    oid_t cpm_paddr;       // 0x30
} checkpoint_mapping_t;    // 0x38
cpm_type: The type of the mapped object
cpm_subtype: The (optional) subtype of the mapped object
cmp_size: The size (in-bytes) of the mapped object
cmp_pad: reserved padding
cpm_fs_oid: The virtual object identifier of the file system that owns the ephemeral object
cpm_oid: The object identifier of the mapped object
cpm_paddr: The physical address of the start of the object
Checkpoint Map Objects contain a simple array of checkpoint_mapping_t structures. Each entry in the map corresponds to an ephemeral object stored in the Checkpoint Data Area. If there are more mappings than can fit into a single Checkpoint Map Object, additional map objects are added to the Checkpoint Descriptor Area. A Checkpoint’s final Checkpoint Map Object is marked with the CHECKPOINT_MAP_LAST flag.

#define CHECKPOINT_MAP_LAST 0x00000001

typedef struct checkpoint_map_phys {
    obj_phys_t cpm_o;               // 0x00
    uint32_t cpm_flags;             // 0x20
    uint32_t cpm_count;             // 0x24
    checkpoint_mapping_t cpm_map[]; // 0x28
} checkpoint_map_phys_t;
cpm_o: The object header
cpm_flags: A set of bit-flags. Currently, only CHECKPOINT_MAP_LAST is defined
cmp_count: The number of mappings stored in this Checkpoint Map
cmp_map: An array of cmp_count Checkpoint Mappings
Locating Ephemeral Objects
Once you’ve identified the location of the Checkpoint Data Area, enumeration of on-disk ephemeral objects is fairly straight forward. NOTE: You cannot rely on the zero-block NX Superblock copy. You must locate the NX Superblock that belongs to the Checkpoint you’re examining.

Because there are relatively few persistent, ephemeral objects, linear time enumeration of all of a checkpoint’s mappings is practical. This means that there aren’t any complex data structures that get in between us and the objects that we’re looking for.

The nx_xp_desc_index member of the Checkpoint’s nx_superblock_t stores a zero-based block index into the Checkpoint Descriptor Area. This is the location of the first Checkpoint Map Object. Locate this object and validate it using the checksum stored in its object header.

Read the cmp_count member of the Checkpoint Map. This contains the number of Checkpoint Mappings stored in the current map.

Enumerate the mappings stored in the cpm_map array. These mappings each contain information about an on-disk ephemeral object, including the physical block address in which it is stored.

Once all mappings have been enumerated, read the cmp_flags member. If the bit defined in CHECKPOINT_MAP_LAST is set, you’ve reached the end of your journey; otherwise, there are more ephemeral objects to enumerate.

The next Checkpoint Map Object should follow the current map object, but it is important to remember that the Checkpoint Descriptor Area acts as a circular buffer. You can determine the number of blocks in the Checkpoint Descriptor Area, by reading the nx_xp_desc_blocks member of the NX Superblock and ignoring the most-significant bit. If the current map is stored in the last block of the descriptor area, then the next map will be stored in the first.

// calculating the next index in the circular buffer
next_index = current_index + 1;
if (next_index == (nx_cp_desc_blocks & 0x7FFFFFFF)) {
    next_index = 0;
}

// alternatively...
next_index = (current_index + 1) % (nx_cp_desc_blocks & 0x7FFFFFFF);
Conclusion
Compared to other kinds of objects in APFS, each checkpoint only maintains a relatively small amount of on-disk ephemeral objects. Due to their nature, these objects are likely all read into memory at once when the Checkpoint is mounted. Thanks to these facts, ephemeral objects are stored on disk in a way that is relatively simple for us to find and enumerate.

If only it were always that simple… Next up in this series we will discuss B-Trees – APFS’s method of choice for referencing potentially large sets of data on disk.

2022 APFS Advent Challenge Day 6 - B-Trees (Part 1)
Thursday, December 8, 2022
In yesterday’s post, we discussed Checkpoint Maps, the simple linear-time data structures that APFS uses to manage persistent, ephemeral objects. Today, we will give a general overview of B-Trees and detail the layout and on-disk structures of B-Tree Nodes.

Background
A B-Tree is a self-balancing tree data structure that maintains sorted data and allows for fast search, insertion, and deletion operations. B-Trees are a generalization of binary trees that enable each node to have more than two children, increasing the number of key/value pairs and thus the fan-out. B-Trees are self-balancing – automatically adjusting their structure to maintain a specific balance factor that guarantees operations in logarithmic time.

B-Trees are used in APFS to store and reference many different objects and data types. APFS B-Trees differ from traditional implementations in several ways. Leaf nodes store values, and non-leaf nodes store their children’s object identifiers (oids). Because oids are usually smaller than values, non-leaf nodes in APFS can reference more children than other implementations, reducing tree depth and increasing fan-out. B-Tree nodes in APFS support fixed or variable-length keys and values. This flexibility enables the storage of heterogeneous types of data. APFS provides for the specialization of B-Tree sub-objects that define their configuration and storage order. APFS B-Trees are well-suited for storing large amounts of data on disk or other sequential access storage devices.

On-Disk Structure and Layout
There are two types of APFS B-Tree node objects, and their layer differs slightly: B-Tree Root Node Objects and B-Tree Node Objects. Traversal of APFS B-Trees must start from a Root Node Object because nodes only store references to their children, not their siblings.

Structure of a Root B-Tree Node

Structure of a Root B-Tree Node

Structure of a Non-Root B-Tree Node

Structure of a Non-Root B-Tree Node

The layout of root nodes differs from non-root nodes in that they end with a btree_info_t structure. This structure gives information about the entire B-Tree. To avoid data duplication and to make more efficient use of space, non-root nodes do not store a copy of this information.

B-Tree Info
typedef struct btree_info_fixed {
    uint32_t bt_flags;     // 0x00
    uint32_t bt_node_size; // 0x04
    uint32_t bt_key_size;  // 0x08
    uint32_t bt_val_size;  // 0x0C
} btree_info_fixed_t;      // 0x10
bt_flags: A set of B-Tree node flags
bt_node_size: The on-disk size, in bytes, of a node in this B-Tree
bt_key_size: The size of a key, or zero if the keys have variable size
bt_val_size: The size of a key, or zero if the values have variable size
typedef struct btree_info {
    btree_info_fixed_t bt_fixed; // 0x00
    uint32_t bt_longest_key;     // 0x10
    uint32_t bt_longest_val;     // 0x14
    uint64_t bt_key_count;       // 0x18
    uint64_t bt_node_count;      // 0x20
} btree_info_t;                  // 0x28
bt_fixed: Information about the B-tree that doesnʼt change over time
bt_longest_key: The size (in bytes) of the largest key stored in the tree
bt_longest_value: The size (in bytes) of the largest value stored in the tree
bt_key_count: The number of keys stored in the B-Tree
bt_node_count: The number of nodes that make up the B-Tree
B-Tree Flags
Name	Value	Description
BTREE_UINT64_KEYS	0x00000001	Keys are all of type uint64_t
BTREE_SEQUENTIAL_INSERT	0x00000002	Keys are inserted sequentially
BTREE_ALLOW_GHOSTS	0x00000004	Allow keys with no corresponding value
BTREE_EPHEMERAL	0x00000008	Child node oids are ephemeral
BTREE_PHYSICAL	0x00000010	Child node oids are physical
BTREE_NONPERSISTENT	0x00000020	never found on disk
BTREE_KV_NONALIGNED	0x00000040	Keys and values aren’t 8-byte aligned
BTREE_HASHED	0x00000080	Non-leaf nodes store child hashes
BTREE_NOHEADER	0x00000100	Nodes are all stored without headers
Header
Both types of node objects begin with a btree_node_phys_t structure.

// A location within a B-tree node.
typedef struct nloc {
    uint16_t off; // 0x00
    uint16_t len; // 0x02
} nloc_t;         // 0x04
off: The context-specific offset (in bytes)
len: The length (in bytes)
typedef struct btree_node_phys {
    obj_phys_t btn_o;         // 0x00
    uint16_t btn_flags;       // 0x20
    uint16_t btn_level;       // 0x22
    uint32_t btn_nkeys;       // 0x24
    nloc_t btn_table_space;   // 0x28
    nloc_t btn_free_space;    // 0x2C
    nloc_t btn_key_free_list; // 0x30
    nloc_t btn_val_free_list; // 0x34
} btree_node_phys_t;          // 0x38
btn_o: The node’s object header
btn_flags: A set of node-specific bit-flags
btn_level: The number of child levels below this node
btn_nkeys: The number of keys stored in this node
btn_table_space: The location of the table of contents relative to the end of the btree_node_phys_t structure
btn_free_space: The location of the table of contents relative to the beginning of the key area
brn_key_free_list: A linked list that tracks free space in the key area
brn_val_free_list: A linked list that tracks free space in the value area
Node Flags
Name	Value	Description
BTNODE_ROOT	0x0001	Root node
BTNODE_LEAF	0x0002	Leaf node
BTNODE_FIXED_KV_SIZE	0x0004	Fixed-sized keys and values
BTNODE_HASHED	0x0008	Node contains child hashes
BTNODE_NOHEADER	0x0010	Node is stored without an object header
BTNODE_CHECK_KOFF_INVAL	0x8000	never found on disk
Table of Contents
Each node has a table of contents (ToC) which stores an entry to each key/value pair contained in the node. The table space can be located by reading the btn_table_space member of the header. The start of this area is btn_table_space.off bytes after the btree_node_phys_t and is btn_table_space.len bytes in size.

ToC entries identify the location of key and value data in their respective storage areas. The structure of these entries differs depending on whether the BTNODE_FIXED_KV_SIZE flag is set in the header. When this flag is set, kvoff_t structures in the ToC store offsets of keys and values relative to their respective storage areas.

// The location, within a B-tree node, of a fixed-size key and value.
typedef struct kvoff {
    uint16_t k; // 0x00
    uint16_t v; // 0x02
} kvoff_t;      // 0x04
k: Offset of the key (relative to the start of the key area)
v: Offset of the value (relative to the end of the value area)
It is necessary to store both offsets and sizes for nodes with variable keys.

// The location, within a B-tree node, of a key and value.
typedef struct kvloc {
    nloc_t k; // 0x00
    nloc_t v; // 0x04
} kvloc_t;    // 0x08
k: Location of the key (relative to the start of the key area)
v: Location of the value (relative to the end of the value area)
Key Area
The area where key data is stored is called the key area and begins immediately after the table space. The size of this area grows downward towards the end of the node. Key offsets in the ToC are relative to the start of this area.

Value Area
The area where value data is stored is called the value area. For non-root nodes, the value area starts at the end of the node. For root nodes, the start is before the btree_info_t structure. This area grows upward towards the beginning of the node, and value offsets in the ToC are interpreted as negative offsets relative to this point.

Conclusion
B-Trees are tree-like data structures that provide fast access and efficient storage of key/value data. B-Trees are widely used in APFS and serve several specialized purposes.

In the next post, we will continue our discussion of B-Trees and detail methods of enumerating and looking up their referenced objects.

2022 APFS Advent Challenge Day 7 - B-Trees (Part 2)
Friday, December 9, 2022
Mastering the skill of B-Tree traversal is essential in parsing information from APFS. Our last post gave an overview of APFS B-Trees, their layout, and on-disk node structures. Today, we will discuss applying that knowledge to perform enumeration and fast lookups of referenced objects.

Overview
Traversal of APFS B-Trees always starts at the root node, which can be identified by having the BTNODE_ROOT bit-flag set in the bt_flags field of its btree_node_phys_t header. Each B-Tree can only have a single root node. Root nodes only differ from the other nodes in that their storage space is slightly more limited to make room for the btree_info_fixed_t structure, which stores information about the entire tree.

You can visualize an APFS B-Tree as a root node on the highest level, branching downward for each generation of children. The nodes at the lowest level (level-zero) are called leaf nodes and have the BTNODE_LEAF bit-flag in their header. Single-level B-Trees consist of a solitary node that is both a root and leaf node. Two-level B-Trees have a single root node whose immediate children are leaf nodes. B-Trees with more than two levels will have intermediary nodes that are neither root nor leaf nodes.

Unique, sortable keys reference each value in a B-Tree. When a B-Tree is created, the key/value pairs (k/v) are sorted in a context-specific manner that is specific to how the tree is used. The pairs are then written to as few level-zero leaf nodes as possible. The leaf nodes are evenly allocated among parent nodes, which reference their children by storing a copy of the first key in the child, which is mapped to the child node’s object identifier (oid). Higher order levels are created, as necessary, until reaching a single root node.

Enumerating B-Trees
APFS B-Tree nodes have three storage areas: the table space, the key area, and the value area. The table space contains the table of contents – the table of contents stores the location of each key-value pair within the node. If the BTNODE_FIXED_KV_SIZE flag is set in the node’s header flags, the table of contents only stores offsets for keys and values. Otherwise, it also stores their lengths.

Identify the location of the table space within the node by reading the btn_table_space member of the node’s header. This table space begins at btn_table_space.off bytes after the node’s header and is btn_table_space.len bytes in length. The remaining storage space in the node is used for the key and value areas. We will refer to this space collectively as the k/v area.

Remember: The k/v area is 0x28 bytes smaller in root nodes due to the btree info stored at the end.

The table of contents is an array of either type kvoff_t (for fixed-size k/v pairs) or kvloc_t elements. Read the size of this array from the btn_nkeys field of the node header. Each array element corresponds to a k/v pair. Key offsets are relative to the beginning of the k/v area and value offsets from the end. Enumerating through this array allows you to locate the data associated with each key and value in the node.

NOTE: B-Trees with the BTREE_ALLOW_GHOSTS flag can sparsely populate values. The value 0xFFFF stored in the offset field of an entry in the table of contents indicates that there is no associated value for a given key stored in the node.

Leaf nodes give direct access to values, so no further effort is required in these cases. For non-leaf nodes, the stored value can be interpreted as an object identifier. The btree info may contain either the BTREE_EPHEMERAL or BTREE_PHYSICAL flags or neither. This indicates whether non-root nodes are physical, ephemeral, or virtual objects. Physical nodes can be directly addressed by their object identifiers. Ephemeral nodes need to be looked up using Checkpoint Maps. Virtual nodes require querying Object Maps (discussed in the next post). In all cases, locate each child node using the appropriate method and continue the enumeration process.

Faster Lookups of Specific Values
We could use enumeration to look up a value by its key as with Checkpoint Maps, but that would require linear time. We can use the copies of the key information stored in non-leaf nodes for much faster logarithmic-time lookups. When analyzing a non-leaf node, identify the key closest to, but not ordered after, the desired key. We can then continue the search by analyzing a specific child node without enumerating the rest of the node’s children. Because APFS B-Tree Nodes are optimized for minimal depth, we can identify a particular k/v pair with minimal enumeration.

Conclusion
Understanding the structure and traversal of APFS B-Trees is crucial for effectively parsing information from this file system. We discuss processes that allow the enumeration of all values in linear time and faster logarithmic-time lookups of specific values.

B-Trees are used in many ways in APFS. In the next post, we will discuss Object Map B-Trees and how they can be used to access virtual objects.

2022 APFS Advent Challenge Day 8 - Object Maps
Monday, December 12, 2022
Earlier in this series, we discussed APFS Containers and how they address physical objects via a fixed block size. This was followed up with a discussion on enumerating Checkpoint Maps to locate ephemeral objects. The last remaining kind of objects that we need to know how to find are virtual objects. Today, we will discuss an essential specialization of B-Trees, the Object Map (OMAP), and their critical role in managing these virtual objects in APFS.

Object Maps perform two essential roles in APFS. They facilitate the address translation needed to locate virtual objects on disk and provide snapshot capabilities that can instantly roll back the set of virtual objects to an earlier point in time. The Container and each Volume maintain their own independent OMAPs. Each OMAP has its own virtual address space, so when dereferencing a virtual object, it is essential to understand how that object is used to know which object map to query.

On-Disk Structures
OMAP Objects have a relatively straightforward on-disk structure. Along with some minor metadata, the primary use of the OMAP Object is to store the physical address of its tree. Optionally, they also store an address to a Snapshot Tree. Both trees are structured as B-Tree Objects.

Incidentally, there are indications that APFS was intended to support more than one type of tree structure. Until recently, Apple’s apfs.kext contained references to an undocumented “M-Tree” type, which appeared to be designed to be used in place of B-Trees in some instances. M-Trees are never mentioned in Apple’s APFS File System Reference, I have never seen them on disk, and all references of M-Trees were removed from the macOS 13 apfs.kext.

typedef struct omap_phys {
    obj_phys_t om_o;                // 0x00
    uint32_t om_flags;              // 0x20
    uint32_t om_snap_count;         // 0x24
    uint32_t om_tree_type;          // 0x28
    uint32_t om_snapshot_tree_type; // 0x2C
    oid_t om_tree_oid;              // 0x30
    oid_t om_snapshot_tree_oid;     // 0x38
    xid_t om_most_recent_snap;      // 0x40
    xid_t om_pending_revert_min;    // 0x48
    xid_t om_pending_revert_max;    // 0x50
} omap_phys_t;                      // 0x58
om_o: The object header
om_flags: OMAP flags (see below)
om_snap_count: The number of snapshots
om_tree_type: The type of OMAP tree. This is currently always a physical B-Tree Root Node
om_snapshot_tree_type: The type of Snapshot tree. This is currently always a physical B-Tree Root Node
om_tree_oid: The physical object identifier of the omap’s B-Tree
om_snapshot_tree_oid: The physical object identifier of the snapshot’s B-Tree or zero if there is none.
om_most_recent_snap: The transaction identifier of the latest snapshot
om_pending_revert_min: The earliest transaction identifier for an in-progress revert
om_pending_revert_max: The latest transaction identifier for an in-progress revert
NOTE: Apple’s APFS File System Reference incorrectly lists the om_tree_oid and om_snapshot_tree_oid members as _virtual object identifiers when they are, in fact, physical.

Object Map Flags
Name	Value	Description
OMAP_MANUALLY_MANAGED	0x00000001	Does not support snapshots
OMAP_ENCRYPTING	0x00000002	Encryption in progress
OMAP_DECRYPTING	0x00000004	Decryption in progress
OMAP_KEYROLLING	0x00000008	Encryption key change in progress
OMAP_CRYPTO_GENERATION	0x00000010	Encryption change marker (more on this later)
Object Map B-Tree
An OMAP tree is a specialized B-Tree type that maps fixed-size omap_key_t and omap_value_t key/value pairs. There can be more than one referenced object with the same virtual object identifier (oid) stored in the tree. Transaction identifiers (xid) are stored to identify these versions.

Keys are sorted in ascending order, first by oid and then xid. Values contain the size and physical address of the mapped object. We will discuss APFS encryption in the future, but for now, it is sufficient to note that omap_value_t marks encrypted objects via the OMAP_VAL_ENCRYPTED bit-flag.

typedef struct omap_key {
    oid_t ok_oid; // 0x00
    xid_t ok_xid; // 0x08
} omap_key_t;     // 0x10
ok_oid: The mapped object’s virtual object identifier
ok_xid: The mapped object’s transaction identifier
typedef struct omap_val {
    uint32_t ov_flags; // 0x00
    uint32_t ov_size;  // 0x04
    paddr_t ov_paddr;  // 0x08
} omap_val_t;          // 0x10
ov_flags: OMAP value flags (see below)
ov_size: Size of the mapped object (in bytes)
ov_paddr: The physical address of the start of the mapped object
Object Map Value Flags
Name	Value	Description
OMAP_VAL_DELETED	0x00000001	Object mapping has been removed from the map and this is a placeholder
OMAP_VAL_SAVED	0x00000002	This object mapping shouldnʼt be replaced when the object is updated. (currently unused)
OMAP_VAL_ENCRYPTED	0x00000004	The mapped object is encrypted
OMAP_VAL_NOHEADER	0x00000008	The mapped object has a zero’d object header
OMAP_VAL_CRYPTO_GENERATION	0x00000010	Encryption change marker
Object Map Snapshot Tree
Checkpoint OMAPs also maintain an entry for each of their snapshots in a Snapshot Tree. These trees map xid_t transaction identifiers to omap_snapshot_t values. Other than the deletion state of a snapshot, there is very little information to be gained from enumerating this tree. Volumes maintain additional trees that store more detailed metadata about snapshots. We will discuss those trees in the future.

typedef struct omap_snapshot {
    uint32_t oms_flags; // 0x00
    uint32_t oms_pad;   // 0x04
    oid_t oms_oid;      // 0x08
} omap_snapshot_t;      // 0x10
oms_flags - OMAP Snapshot Flags (see below)
oms_pad - padding
oms_oid - reserved and unused
Name	Value	Description
OMAP_SNAPSHOT_DELETED	0x00000001	The snapshot has been deleted
OMAP_SNAPSHOT_REVERTED	0x00000002	The snapshot has been deleted as part of a revert
Parsing Object Maps
Once you understand the structure of the OMAP key/value pairs, the process of parsing an OMAP is the same as parsing other B-Trees. When looking up a virtual object from the active filesystem state, choose the key with the highest xid available. If parsing the state from a snapshot, ignore all keys whose xid is greater than the xid of the snapshot in question.

Conclusion
Object Maps are an essential component of APFS that serve two key roles. They provide the address translation needed to locate virtual objects on disk and enable snapshot capabilities that allow instant rollback to earlier points in time.

Now that we’ve covered the basics, it’s time for us to start diving deeper into APFS and learn how to parse information from the file systems themselves. Check back tomorrow when we begin discussing APFS Volumes and their file systems.

2022 APFS Advent Challenge Day 9 - Volume Superblock Objects
Tuesday, December 13, 2022
In this blog post, we will explore the Volume Superblock in APFS, a critical data structure containing important information about an individual APFS volume. We will discuss locating the Volume Superblock on disk and describe some fields in the on-disk format. By the end of this post, you should better understand the volume Superblock’s role in the APFS file system and how to parse its on-disk structure.

Locating Volume Superblocks
Every volume in APFS has a Volume Superblock Object that serves as the root source of on-disk information about the volume and its file system. These are virtual objects whose references are managed by the Container’s Object Map. Once you understand how to parse an Object Map, identifying the on-disk location of these superblocks is simple.

Read the nx_max_file_systems field of the NX Superblock Object to determine the maximum number of volumes that the container can support.

Enumerate the nx_fs_oid array looking for the non-zero virtual object identifiers of each Volume Superblock Object. The identifiers may not always be stored contiguously in the array, but all array entries after nx_max_file_systems will always be zero.

Query the Container’s Object Map to locate the physical address of each Volume Superblock.

On-Disk Structures
Volume Superblock Objects are stored on disk as apfs_superblock_t structures. We will discuss many of these structure fields in detail throughout this series. Below is a short description of each.

#define APFS_MAGIC 0x41504342 // APSB
#define APFS_MAX_HIST 8
#define APFS_VOLNAME_LEN 256

// Volume Superblock
typedef struct apfs_superblock {
    obj_phys_t apfs_o;                                  // 0x00
    uint32_t apfs_magic;                                // 0x20
    uint32_t apfs_fs_index;                             // 0x24
    uint64_t apfs_features;                             // 0x28
    uint64_t apfs_readonly_compatible_features;         // 0x30
    uint64_t apfs_incompatible_features;                // 0x38
    uint64_t apfs_unmount_time;                         // 0x40
    uint64_t apfs_fs_reserve_block_count;               // 0x48
    uint64_t apfs_fs_quota_block_count;                 // 0x50
    uint64_t apfs_fs_alloc_count;                       // 0x58
    wrapped_meta_crypto_state_t apfs_meta_crypto;       // 0x60
    uint32_t apfs_root_tree_type;                       // 0x74
    uint32_t apfs_extentref_tree_type;                  // 0x78
    uint32_t apfs_snap_meta_tree_type;                  // 0x7C
    oid_t apfs_omap_oid;                                // 0x80
    oid_t apfs_root_tree_oid;                           // 0x88
    oid_t apfs_extentref_tree_oid;                      // 0x90
    oid_t apfs_snap_meta_tree_oid;                      // 0x98
    xid_t apfs_revert_to_xid;                           // 0xA0
    oid_t apfs_revert_to_sblock_oid;                    // 0xA8
    uint64_t apfs_next_obj_id;                          // 0xB0
    uint64_t apfs_num_files;                            // 0xB8
    uint64_t apfs_num_directories;                      // 0xC0
    uint64_t apfs_num_symlinks;                         // 0xC8
    uint64_t apfs_num_other_fsobjects;                  // 0xD0
    uint64_t apfs_num_snapshots;                        // 0xD8
    uint64_t apfs_total_blocks_alloced;                 // 0xE0
    uint64_t apfs_total_blocks_freed;                   // 0xE8
    uuid_t apfs_vol_uuid;                               // 0xF0
    uint64_t apfs_last_mod_time;                        // 0x100
    uint64_t apfs_fs_flags;                             // 0x108
    apfs_modified_by_t apfs_formatted_by;               // 0x110
    apfs_modified_by_t apfs_modified_by[APFS_MAX_HIST]; // 0x140
    uint8_t apfs_volname[APFS_VOLNAME_LEN];             // 0x2C0
    uint32_t apfs_next_doc_id;                          // 0x3C0
    uint16_t apfs_role;                                 // 0x3C4
    uint16_t reserved;                                  // 0x3C6
    xid_t apfs_root_to_xid;                             // 0x3C8
    oid_t apfs_er_state_oid;                            // 0x3D0
    uint64_t apfs_cloneinfo_id_epoch;                   // 0x3D8
    uint64_t apfs_cloneinfo_xid;                        // 0x3E0
    oid_t apfs_snap_meta_ext_oid;                       // 0x3E8
    uuid_t apfs_volume_group_id;                        // 0x3F0
    oid_t apfs_integrity_meta_oid;                      // 0x400
    oid_t apfs_fext_tree_oid;                           // 0x408
    uint32_t apfs_fext_tree_type;                       // 0x410
    uint32_t reserved_type;                             // 0x414
    oid_t reserved_oid;                                 // 0x418
} apfs_superblock_t;                                    // 0x420
apfs_o: The object header
apfs_magic: Magic value (always APFS_MAGIC)
apfs_fs_index: The index of the object identifier for this volume’s file system in the container’s array of file systems
apfs_features: A bit-field of the optional features being used by this volume
apfs_readonly_compatible_features: A bit-field of the read-only compatible features being used by this volume. (none currently defined)
apfs_incompatible_features: A bit-field of the backward-incompatible features being used by this volume
apfs_unmount_time: The time that this volume was last unmounted
apfs_fs_reserve_block_count: The number of blocks that have been reserved for this volume to allocate
apfs_fs_quota_block_count: The maximum number of blocks that this volume can allocate (or zero if no limit)
apfs_fs_alloc_count: The number of blocks currently allocated for this volume’s file system
apfs_meta_crypto: Information about the key used to encrypt metadata for this volume
apfs_root_tree_type: The type of the root file-system tree
apfs_extentref_tree_type: The type of the extent-reference tree
apfs_snap_meta_tree_type: The type of the snapshot metadata tree
apfs_omap_oid: The physical object identifier of the volume’s object map
apfs_root_tree_oid: The virtual object identifier of the root file-system tree
apfs_extentref_tree_oid: The physical object identifier of the extent-reference tree
apfs_snap_meta_tree_oid: The physical object identifier of the snapshot metadata tree
apfs_revert_to_xid: The transaction identifier of a snapshot that the volume will revert to (or zero if not reverting)
apfs_revert_to_sblock_oid: The physical object identifier of a volume superblock that the volume will revert to (or zero i not reverting)
apfs_next_obj_id: The next identifier that will be assigned to a file-system object in this volume
apfs_num_files: The number of regular files in this volume
apfs_num_directories: The number of directories in this volume
apfs_num_symlinks: The number of symbolic links in this volume
apfs_num_other_fsobjects: The number of other files in this volume
apfs_num_snapshots: The number of snapshots in this volume
apfs_total_blocks_alloced: The total number of blocks that have been allocated by this volume
apfs_total_blocks_freed: The total number of blocks that have been freed by this volume
apfs_vol_uuid: The universally unique identifier for this volume
apfs_last_mod_time: The universally unique identifier for this volume
apfs_fs_flags: The volume’s flags
apfs_formatted_by: Information about the software that created this volume
apfs_modified_by: Information about the software that has modified this volume
apfs_volname: The name of the volume, represented as a null-terminated UTF-8 string
apfs_next_doc_id: The next document identifier that will be assigned
apfs_role: The role of this volume within the container
reserved: reserved
apfs_root_to_xid: The transaction identifier of the snapshot to root from, or zero to root normally
apfs_er_state_oid: The object id of the encryption rolling state (or zero if no encryption change is in progress)
apfs_cloneinfo_id_epoch: The largest object identifier used by this volume at the time INODE_WAS_EVER_CLONED started storing valid information
apfs_cloneinfo_xid: A transaction identifier used with apfs_cloneinfo_id_epoch
apfs_snap_meta_ext_oid: The virtual object identifier of the extended snapshot metadata object
apfs_volume_group_id: The volume group the volume belongs to (or null UUID if not part of a volume group)
apfs_integrity_meta_oid: The virtual object identifier of the integrity metadata object
apfs_fext_tree_oid: The virtual object identifier of the file extent tree
apfs_fext_tree_type: The type of the file extent tree
reserved_type: reserved
reserved_oid: reserved
Optional Feature Flags
Optional Feature Flags indicate that the volume uses newer features that older APFS implementations may not support but should be backward compatible with older standards. Implementations that do not support these features should still be able to mount the volume read/write without problems. These bit-flags are stored in the apfs_features field of the Volume Superblock.

Name	Value	Description
APFS_FEATURE_DEFRAG_PRERELEASE	0x00000001	reserved
APFS_FEATURE_HARDLINK_MAP_RECORDS	0x00000002	The volume has hardlink map records
APFS_FEATURE_DEFRAG	0x00000004	The volume supports defragmentation
APFS_FEATURE_STRICTATIME	0x00000008	This volume updates file access times every time the file is read
APFS_FEATURE_VOLGRP_SYSTEM_INO_SPACE	0x00000010	This volume supports mounting a system and data volume as a single user-visible volume
Incompatible Volume Feature Flags
Incompatible Volume Feature Flags indicate that the volume uses newer features that may not be supported by older APFS implementations and are not fully backward compatible with older standards. Implementations not supporting these features will likely have problems mounting the volume and are encouraged not to do so. These bit-flags are stored in the apfs_incompatible_features field of the Volume Superblock.

Name	Value	Description
APFS_INCOMPAT_CASE_INSENSITIVE	0x00000001	Filenames on this volume are case insensitive
APFS_INCOMPAT_DATALESS_SNAPS	0x00000002	At least one snapshot with no data exists for this volume
APFS_INCOMPAT_ENC_ROLLED	0x00000004	This volumeʼs encryption has changed keys at least once
APFS_INCOMPAT_NORMALIZATION_INSENSITIVE	0x00000008	Filenames on this volume are normalization insensitive
APFS_INCOMPAT_INCOMPLETE_RESTORE	0x00000010	This volume is being restored, or a restore operation to this volume was uncleanly aborted
APFS_INCOMPAT_SEALED_VOLUME	0x00000020	This volume can’t be modified
APFS_INCOMPAT_RESERVED_40	0x00000040	reserved
Volume Flags
Volume Flags are used to indicate additional information about the volume’s status. This bit-field is stored in the apfs_flags field of the Volume Superblock.

Name	Value	Description
APFS_FS_UNENCRYPTED	0x00000001	The volume is not encrypted
APFS_FS_RESERVED_2	0x00000002	reserved
APFS_FS_RESERVED_4	0x00000004	reserved
APFS_FS_ONEKEY	0x00000008	Files on the volume are all encrypted using the volume encryption key (VEK)
APFS_FS_SPILLEDOVER	0x00000010	The volume has run out of allocated space on the solid-state drive
APFS_FS_RUN_SPILLOVER_CLEANER	0x00000020	The volume has spilled over and the spillover cleaner must be run
APFS_FS_ALWAYS_CHECK_EXTENTREF	0x00000040	The volume’s extent reference tree is always consulted when deciding whether to overwrite an extent
APFS_FS_RESERVED_80	0x00000080	reserved
APFS_FS_RESERVED_100	0x00000100	reserved
Volume Roles
In most instances, an APFS Volume is marked as having a defined role. The presence of these roles is context specific, depending on whether the device being analyzed is running macOS or iOS. A Volume can only have a single defined role whose value is stored in the apfs_role member of the Volume Superblock.

Name	Value	Description
APFS_VOL_ROLE_NONE	0x0000	The volume has no defined role
APFS_VOL_ROLE_SYSTEM	0x0001	The volume contains a root directory for the system
APFS_VOL_ROLE_USER	0x0002	The volume contains users’ home directories
APFS_VOL_ROLE_RECOVERY	0x0004	The volume contains a recovery system
APFS_VOL_ROLE_VM	0x0008	The volume is used as swap space for virtual memory
APFS_VOL_ROLE_PREBOOT	0x0010	The volume contains files needed to boot from an encrypted volume
APFS_VOL_ROLE_INSTALLER	0x0020	The volume is used by the OS installer
APFS_VOL_ROLE_DATA	0x0040	The volume contains mutable data
APFS_VOL_ROLE_BASEBAND	0x0080	The volume is used by the radio firmware
APFS_VOL_ROLE_UPDATE	0x00C0	The volume is used by the software update mechanism
APFS_VOL_ROLE_XART	0x0100	The volume is used to manage OS access to secure user data
APFS_VOL_ROLE_HARDWARE	0x0140	The volume is used for firmware data
APFS_VOL_ROLE_BACKUP	0x0180	The volume is used by Time Machine to store backups
APFS_VOL_ROLE_RESERVED_7	0x01C0	reserved
APFS_VOL_ROLE_RESERVED_8	0x0200	reserved
APFS_VOL_ROLE_ENTERPRISE	0x0240	This volume is used to store enterprise-managed data
APFS_VOL_ROLE_RESERVED_10	0x0280	reserved
APFS_VOL_ROLE_PRELOGIN	0x02C0	This volume is used to store system data used before login
apfs_modified_by_t
Volume Superblocks store record-keeping information about the tool used to create them in the apfs_formatted_by field. In this case, the id field of the apfs_modified_by_t structure will give the name and version number of the userland program used to create the volume.

These apfs_modified_by_t structures are also used to keep a history of the apfs.kext versions used for the last eight times the volume was mounted read/write. These are stored in the apfs_modified_by array field.

#define APFS_MODIFIED_NAMELEN 32

// Information about a program that modified the volume.
typedef struct apfs_modified_by {
    uint8_t id[APFS_MODIFIED_NAMELEN]; // 0x00
    uint64_t timestamp;                // 0x20
    xid_t last_xid;                    // 0x28
} apfs_modified_by_t;                  // 0x30
id: A string that identifies the program and its version
timestamp: The time that the program last modified this volume
last_xid: The time that the program last modified this volume
Conclusion
Locating and analyzing Volume Superblocks are essential early steps in being able to parse the contents of their file systems. Our next post will discuss the volume’s File System Tree, a specialized B-Tree that stores the bulk of the file systems’ metadata objects.

2022 APFS Advent Challenge Day 11 - File System Trees
Thursday, December 15, 2022
Each APFS volume has a logical file system stored on disk as a collection of File System Objects. Unlike other APFS Objects, File System Objects consist of one or more File System Records, which are stored in the volume’s File System Tree (FS-Tree). Each record stores specific information about a file or directory. Analyzing each record and associating them with other records with the same identifier gives a complete picture of the file system entry. This post will discuss how these records are organized in the volume’s FS-Tree.

Overview
The File System Tree is a specialized B-Tree that differs in several ways from the other trees that we’ve discussed so far:

FS-Trees are virtual B-Trees. Each node in the tree is a virtual object owned by the Volume’s Object Map. This means that querying the FS-Tree requires using the Object Map to locate each node.

FS-Tree nodes can be optionally encrypted. (We will discuss encryption in a future post.) This allows for select volumes to encrypt not only their files’ contents but their metadata as well.

FS-Trees store a heterogeneous set of records – multiple types of keys and values are stored in the same tree.

One advantage of being virtual trees is that FS-Trees can take full advantage of the Object Map’s snapshotting capabilities to restore their state to previous points in time. Apple also uses the snapshots to compare an FS-Tree with an earlier version of itself to create deltas for Time Machine backups.

Keys
Because FS-Trees have multiple key types, they require a way to identify record types. All keys begin with a common structure for this purpose. Specific types may add additional fields to their keys.

#define OBJ_ID_MASK 0x0fffffff'ffffffff
#define OBJ_TYPE_MASK 0xf0000000'00000000
#define OBJ_TYPE_SHIFT 60

typedef struct j_key {
    uint64_t obj_id_and_type;
} j_key_t;
obj_id_and_type: A bit field that encodes the record’s object identifier (in the 60 least-significant bits) and type (in the found most-significant bits).
Keys are ordered first by an object identifier and then by type. A File System Object’s records will be stored together sequentially. Search the FS-Tree for the first record with a given identifier and then enumerate subsequent records until reaching one with a different ID.

File System Record Types
Below is a table of the documented File System Record Types. We will discuss the on-disk format of each record type soon.

Name	Value	Description
APFS_TYPE_SNAP_METADATA	1	Metadata about a snapshot
APFS_TYPE_EXTENT	2	A physical extent record
APFS_TYPE_INODE	3	An inode
APFS_TYPE_XATTR	4	An extended attribute
APFS_TYPE_SIBLING_LINK	5	A mapping from an inode to hard links
APFS_TYPE_DSTREAM_ID	6	A data stream
APFS_TYPE_CRYPTO_STATE	7	A per-file encryption state
APFS_TYPE_FILE_EXTENT	8	A physical extent record for a file
APFS_TYPE_DIR_REC	9	A directory entry
APFS_TYPE_DIR_STATS	10	Information about a directory
APFS_TYPE_SNAP_NAME	11	The name of a snapshot
APFS_TYPE_SIBLING_MAP	12	A mapping from a hard link to its target inode
APFS_TYPE_FILE_INFO	13	Additional information about file data
Conclusion
The File System Tree (FS-Tree) in an APFS volume is a specialized B-Tree that stores information about the files and directories on the volume. A unique object identifier and type identify each record in the tree, and the FS-Tree is ordered by these keys. FS-Tree nodes can be encrypted, and the tree takes advantage of the Object Map’s snapshotting capabilities. By analyzing the records in the FS-Tree, one can gain a complete understanding of the volume’s file system. In our next post, we will discuss the details of some of these records.

2022 APFS Advent Challenge Day 12 - Inode and Directory Records
Friday, December 16, 2022
Each APFS file system entry has both an inode and directory record. The inode record stores metadata such as the entry’s timestamps, ownership, type, and permissions (among others). Directory records store information about where the entry is stored within the file system’s hierarchy. A single inode may be referenced by more than one directory record, meaning the same file or folder may be present at multiple paths in the file system, as is the case with hard links.

Inode Records
The first record stored for each file system entry in a File System Tree should be an inode record.The key for an inode record only consists of the standard j_key_t structure with the “type” identified as APFS_TYPE_INODE.

// FS-Tree key for inode records
typedef struct j_inode_key {
    j_key_t hdr; // 0x00
} j_inode_key_t; // 0x08
hdr: The record’s header
The value for an inode record is variable sized to account for any extended fields that may be stored after the record.

// Type Aliases
typedef uint16_t mode_t;
typedef uint32_t uid_t;
typedef uint32_t gid_t;
typedef uint32_t cp_key_class_t;

// FS-Tree value for inode records
typedef struct j_inode_val {
    uint64_t parent_id;                      // 0x00 
    uint64_t private_id;                     // 0x08
    uint64_t create_time;                    // 0x10
    uint64_t mod_time;                       // 0x18
    uint64_t change_time;                    // 0x20
    uint64_t access_time;                    // 0x28
    uint64_t internal_flags;                 // 0x30
    union {
        int32_t nchildren;                   // 0x38
        int32_t nlink;                       // 0x38
    };
    cp_key_class_t default_protection_class; // 0x3C
    uint32_t write_generation_counter;       // 0x40
    uint32_t bsd_flags;                      // 0x44
    uid_t owner;                             // 0x48
    gid_t group;                             // 0x4C
    mode_t mode;                             // 0x50
    uint16_t pad1;                           // 0x52
    uint64_t uncompressed_size;              // 0x54
    uint8_t xfields[];                       // 0x5C
} j_inode_val_t;
parent_id: The identifier of the file system record for the parent directory
private_id: The unique identifier used by this file’s data stream
create_time: The time that this record was created
mod_time: The time that this record was last modified
change_time: The time that this record’s attributes were last modified
access_time: The time that this record was last accessed
internal_flags: The inode’s flags
nchildren: The number of directory entries
nlink: The number of hard links whose target is this inode
default_protection_class: The default protection class for this inode
write_generation_counter: A monotonically increasing counter that’s incremented each time this inode or its data is modified
bsd_flags: The inode’s BSD flags
owner: The user identifier of the inode’s owner
group: The group identifier of the inode’s group
mode: The file’s mode
pad1: reserved
uncompressed_size: The size of the file without compression
xfields: The inode’s extended fields
Inode Flags
Name	Value	Description
INODE_IS_APFS_PRIVATE	0x00000001	The inode is used internally by an implementation of Apple File System
INODE_MAINTAIN_DIR_STATS	0x00000002	The inode tracks the size of all of its children
INODE_DIR_STATS_ORIGIN	0x00000004	The inode has the INODE_MAINTAIN_DIR_STATS flag set explicitly, not due to inheritance
INODE_PROT_CLASS_EXPLICIT	0x00000008	The inode’s data protection class was set explicitly when the inode was created
INODE_WAS_CLONED	0x00000010	The inode was created by cloning another inode
INODE_FLAG_UNUSED	0x00000020	reserved
INODE_HAS_SECURITY_EA	0x00000040	The inode has an access control list
INODE_BEING_TRUNCATED	0x00000080	The inode was truncated
INODE_HAS_FINDER_INFO	0x00000100	The inode has a Finder info extended field
INODE_IS_SPARSE	0x00000200	The inode has a sparse byte count extended field
INODE_WAS_EVER_CLONED	0x00000400	The inode has been cloned at least once
INODE_ACTIVE_FILE_TRIMMED	0x00000800	The inode is an overprovisioning file that has been trimmed
INODE_PINNED_TO_MAIN	0x00001000	The inode’s file content is always on the main storage device
INODE_PINNED_TO_TIER2	0x00002000	The inode’s file content is always on the secondary storage device
INODE_HAS_RSRC_FORK	0x00004000	The inode has a resource fork
INODE_NO_RSRC_FORK	0x00008000	The inode doesn’t have a resource fork
INODE_ALLOCATION_SPILLEDOVER	0x00010000	The inode’s file content has some space allocated outside of the preferred storage tier for that file
INODE_FAST_PROMOTE	0x00020000	This inode is scheduled for promotion from slow storage to fast storage
INODE_HAS_UNCOMPRESSED_SIZE	0x00040000	This inode stores its uncompressed size in the inode
INODE_IS_PURGEABLE	0x00080000	This inode will be deleted at the next purge
INODE_WANTS_TO_BE_PURGEABLE	0x00100000	This inode should become purgeable when its link count drops to one
INODE_IS_SYNC_ROOT	0x00200000	This inode is the root of a sync hierarchy for fileproviderd
INODE_SNAPSHOT_COW_EXEMPTION	0x00400000	This inode is exempt from copy-on-write behavior if the data is part of a snapshot
Directory Records
Every folder in the file system will store a directory record for each of its children. A directory record’s key begins with the standard key header with the “type” encoded as APFS_TYPE_DIR_REC followed by and encoded hash and name of the directory entry.

#define J_DREC_LEN_MASK 0x000003ff
#define J_DREC_HASH_MASK 0xfffff400
#define J_DREC_HASH_SHIFT 10

// A directory record key
typedef struct j_drec_hashed_key {
    j_key_t hdr;                // 0x00
    uint32_t name_len_and_hash; // 0x08
    uint8_t name[0];            // 0x0C
} j_drec_hashed_key_t;
hdr: The record’s header
name_len_and_hash: Encodes the length (in-bytes) of the UTF-8 encoded name in the 10 least significant bits and a hash of the name in the most significant bits.
name: A null-terminated UTF-8 encoded name of the entry
The value for an directory record is variable sized to account for any extended fields that may be stored after the record.

typedef struct j_drec_val {
    uint64_t file_id;    // 0x00
    uint64_t date_added; // 0x08
    uint16_t flags;      // 0x10
    uint8_t xfields[];   // 0x12
} j_drec_val_t;
file_id: The identifier of the inode that this directory entry represents
date_added: The time that this directory entry was added to the directory
flags: The directory entry’s flags
xfields: The directory entry’s extended fields
Directory Entry Flags
Name	Value	Description
DREC_TYPE_MASK	0x000f	Directory Entry Type Mask (see below)
RESERVED_10	0x0010	reserved
Directory Entry File Types
The eight least significant bits of the directory record flags encode the type of the directory entry as defined below.

Name	Value	Description
DT_UNKNOWN	0	An unknown directory entry
DT_FIFO	1	A named pipe
DT_CHR	2	A character-special file
DT_DIR	4	A directory
DT_BLK	6	A block device
DT_REG	8	A regular file
DT_LNK	10	A symbolic link
DT_SOCK	12	A socket
DT_WHT	14	A whiteout
Extended Fields
Both inode and directory records contain an optional set of extended fields that are used to store additional information. To check if a directory entry or inode has extended fields, the structure size can be compared to the recorded size in the table of contents entry for the file-system record. If the recorded size is different from the structure’s size, then extended fields are present.

Both the j_drec_val_t and j_inode_val_t structures have a field called xfields that stores the extended field data. The xfields field consists of three parts: an xf_blob_t header that indicates the number of extended fields and their size, an array of x_field_t instances that provide the type and size of each extended field, and an array of the extended field data itself, which is aligned to eight-byte boundaries.

The xf_blob_t header is stored directly after the value structure.

// A collection of extended attributes.
typedef struct xf_blob {
    uint16_t xf_num_exts;  // 0x00
    uint16_t xf_used_data; // 0x02
    uint8_t xf_data[];     // 0x04
} xf_blob_t;
xf_num_exts: The number of extended fields
xf_used_data: The amount of space (in bytes) used to store the xfields
xf_data: The extended fields
An x_field_t array is stored at the end of the xf_blob_t header. There is one entry for each extended field.

// An extended field's metadata.
typedef struct x_field {
    uint8_t x_type;  // 0x00
    uint8_t x_flags; // 0x01
    uint16_t x_size; // 0x02
} x_field_t;         // 0x04
x_type: The extended field’s data type
x_flags: The extended field’s flags
x_size: The size, in bytes, of the data stored in the extended field
The data for the extended fields is stored after the x_field_t array. Importantly, this data is stored in the same order as the x_field_t array and is aligned on eight-byte boundaries. The padding bytes are not included in the x_size field. The type of this data depends on the type of the extended field.

Extended Field Types (Inode Records)
Name	Value	Value Type	Description
INO_EXT_TYPE_SNAP_XID	1	xid_t	The transaction identifier for a snapshot
INO_EXT_TYPE_DELTA_TREE_OID	2	oid_t	The virtual object identifier of the file-system tree that corresponds to a snapshot’s extent delta list
INO_EXT_TYPE_DOCUMENT_ID	3	uint32_t	The file’s document identifier
INO_EXT_TYPE_NAME	4	UTF-8 string	The name of the file
INO_EXT_TYPE_PREV_FSIZE	5	uint64_t	The file’s previous size
INO_EXT_TYPE_RESERVED_6	6	 	reserved
INO_EXT_TYPE_FINDER_INFO	7	32 bytes	Opaque data used by Finder
INO_EXT_TYPE_DSTREAM	8	j_dstream_t	A data stream
INO_EXT_TYPE_RESERVED_9	9	 	reserved
INO_EXT_TYPE_DIR_STATS_KEY	10	j_dir_stats_val_t	Statistics about a directory
INO_EXT_TYPE_FS_UUID	11	uuid_t	The UUID of a file system that’s automatically mounted in this directory
INO_EXT_TYPE_RESERVED_12	12	 	reserved
INO_EXT_TYPE_SPARSE_BYTES	13	uint64_t	The number of sparse bytes in the data stream
INO_EXT_TYPE_RDEV	14	uint32_t	The device identifier for a block- or character-special device
INO_EXT_TYPE_PURGEABLE_FLAGS	15	 	reserved
INO_EXT_TYPE_ORIG_SYNC_ROOT_ID	16	uint64_t	The inode number of the sync-root hierarchy that this file originally belonged to
Extended Field Types (Directory Records)
Name	Value	Value Type	Description
DREC_EXT_TYPE_SIBLING_ID	1	uint64_t	The sibling identifier for a directory record
Extended Field Flags
Name	Value	Description
XF_DATA_DEPENDENT	0x0001	The data in this extended field depends on the file’s data
XF_DO_NOT_COPY	0x0002	When copying this file, omit this extended field from the copy
XF_RESERVED_4	0x0004	reserved
XF_CHILDREN_INHERIT	0x0008	When creating a new entry in this directory, copy this extended field to the new directory entry
XF_USER_FIELD	0x0010	This extended field was added by a user-space program
XF_SYSTEM_FIELD	0x0020	This extended field was added by the kernel
XF_RESERVED_40	0x0040	reserved
XF_RESERVED_80	0x0080	reserved

2022 APFS Advent Challenge Day 13 - Data Streams
Monday, December 19, 2022
Data in APFS that is too large to store within records are stored elsewhere on disk and referenced by data streams (dstreams). Similar to non-resident attributes in NTFS, APFS data streams manage a set of extents that reference the number and order of blocks on the disk which contain external data. In this post, we will discuss how data streams are used in APFS to manage one or more forks of data in inodes as well as their record structures in the File System Tree.

Inode Default Data Streams
Each file has a default data stream that stores what we typically refer to as the file’s data. This stream’s object identifier may or may not be different from the inode’s. It is stored in the private_id field of the inode’s j_inode_val_t structure. Metadata about the default data stream is stored as a j_dstream_t structure in an inode extended field with the type of INO_EXT_TYPE_DSTREAM.

typedef struct j_dstream {
    uint64_t size;                // 0x00
    uint64_t alloced_size;        // 0x08
    uint64_t default_crypto_id;   // 0x10
    uint64_t total_bytes_written; // 0x18
    uint64_t total_bytes_read;    // 0x20
} j_stream_t;                     // 0x28
size: The size of the logical data (in bytes)
alloced_size: The total space allocated for the data stream (in bytes), including any unused space
default_crypto_id: The default encryption key or tweak used in this data stream
total_bytes_written: The total number of bytes that have been written to this data stream
total_bytes_read: The total number of bytes that have been read from this data stream
The logical size and allocated size of a dstream may differ. The allocated size is always a factor of the container’s block size. If the file contents do not fill up the last block, then the allocated size may be larger than the logical size. APFS also allows dstreams to be sparsely allocated. Some extent ranges that logically contain all zero-bytes may not be stored on disk. In these instances, the allocated size may be smaller than the logical size of the stream.

The default_crypto_id comes in to play when we’re dealing with encrypted volumes. We will discuss more about APFS encryption in a future post.

The total_bytes_written and total_bytes_read fields are performance counters we can use to determine how often a data stream has been read-from or written-to. They are only periodically updated, and more research is needed to determine what triggers these values to be flushed to disk. Both values are allowed to overflow and reset from zero, so their utility for forensic analysis is relatively limited.

Extended Attributes
Along with the default data stream, files in APFS can also contain other forks. Like in HFS+, these additional data streams are called extended attributes but are similar in concept to alternate data streams in NTFS.

Extended attributes are stored in the File System Tree as records with a type identifier of APFS_TYPE_XATTR and the same object identifier as the inode record. The key half of an extended attribute record is a j_xattr_key_t structure.

typedef struct j_xattr_key {
    j_key_t hdr;       // 0x00
    uint16_t name_len; // 0x08
    uint8_t name[0];   // 0x0A
} j_xattr_key_t;
hdr: The record’s header
name_len: The length of the extended attribute’s name (in bytes), including the final null character.
name: The null-terminated, UTF-8 encoded name of the extended attribute
The value half of the extended attribute record is a j_xattr_val_t structure.

typedef struct j_xattr_val {
    uint16_t flags;     // 0x00
    uint16_t xdata_len; // 0x02
    uint8_t xdata[0];   // 0x04
} j_xattr_val_t;
flags: The extended attribute record’s flags
xdata_len: The length of the extended attribute’s inline data
xdata: The extended attribute data or the identifier of a data stream that contains the data
Extended Attribute Value Flags
Name	Value	Description
XATTR_DATA_STREAM	0x00000001	The attribute data is stored in a data stream
XATTR_DATA_EMBEDDED	0x00000002	The attribute data is stored directly in the record
XATTR_FILE_SYSTEM_OWNED	0x00000004	The extended attribute record is owned by the file system
XATTR_RESERVED_8	0x00000008	reserved
Like NTFS attributes, APFS extended attributes that are small enough can store their data directly in the attribute record itself. In these instances, the XATTR_DATA_EMBEDDED flag will be set and the stream’s data is stored in the xdata field.

Instead, when the XATTR_DATA_STREAM flag is set, xdata stores a j_xattr_dstream_t structure.

typedef struct j_xattr_dstream {
    uint64_t xattr_obj_id; // 0x00
    j_dstream_t dstream;   // 0x08
};                         // 0x30
xattr_obj_id: The object identifier of the extended attribute’s data stream
dstream: The metadata of the extended attribute’s data stream (see above)
Data Stream Extents
Except for Sealed Volumes__ (which we will discuss in the future), the _extents of a dstream are stored in the volume’s File System Tree as a set of records with the type APFS_TYPE_FILE_EXTENT. For streams with non-contiguous data, there will be more than one extent record.

The file extent record keys are of the type j_file_extent_key_t and encode the object identifier of the dstream in their record header, along with the logical offset of the extent in the stream.

typedef struct j_file_extent_key {
    j_key_t hdr;           // 0x00
    uint64_t logical_addr; // 0x08
} j_file_extent_key_t;     // 0x10
hdr: The record’s header
logical_addr: The offset within the file’s data (in bytes) for the data stored in this extent
The value half of a file extent record takes the form of a j_file_extent_val_t structure and is used to denote the physical location of the extent data on disk.

// length and flags masks
#define J_FILE_EXTENT_LEN_MASK 0x00ffffffffffffffULL
#define J_FILE_EXTENT_FLAG_MASK 0xff00000000000000ULL
#define J_FILE_EXTENT_FLAG_SHIFT 56

typedef struct j_file_extent_val {
    uint64_t len_and_flags;  // 0x00
    uint64_t phys_block_num; // 0x08
    uint64_t crypto_id;      // 0x10
} j_file_extent_val_t;       // 0x18
len_and_flags: A bit-field encoding the length (in bytes) of the extent in the 56 least significant bits and its flags in the most significant bits
phys_block_num: The physical block number of the first block in the extent
crypto_id: The encryption key or tweak used in this extent (or zero if not encrypted)
The eight most significant bits of the len_and_flags field are reserved for flags, but no flags are currently defined.

If the value of phys_block_num is zero, then the extent is sparse and should be interpreted as containing all zero bytes.

The crypto_id field is specific to encrypted volumes and will be discussed in a future post.

Conclusion
Understanding data streams and their on-disk structures are essential to analyzing APFS. This post discussed the default data stream, extended attributes, and file extents. Later this week, we will discuss how parsing this information differs in both Sealed and Encrypted volumes.

2022 APFS Advent Challenge Day 14 - Sealed Volumes
Tuesday, December 20, 2022
With the release of macOS 11, Apple added a security feature to APFS called sealed volumes. Sealed volumes can be used to cryptographically verify the contents of the read-only system volume as an additional layer of protection against rootkits and other malware that may attempt to replace critical components of the operating system. Sealed volumes have subtle differences from some of the properties of file systems that we’ve discussed so far.

Identifying a Sealed Volume
Sealed volumes can be identified by checking for the APFS_INCOMPAT_SEALED_VOLUME flag in the apfs_incompatible_features field of their Volume Superblock. In addition, the apfs_integrity_meta_oid and apfs_fext_tree_oid fields must have non-zero values.

An Integrity Metadata Object stores information about the sealed volume. This is a virtual object that is owned by the volume’s Object Map and whose object identifier can be found in the apfs_integrity_meta_oid field of the Volume Superblock. On disk, it is stored as an integrity_meta_phys_t structure.

typedef struct integrity_meta_phys {
    obj_phys_t im_o;               // 0x00
    uint32_t im_version;           // 0x20
    uint32_t im_flags;             // 0x24
    apfs_hash_type_t im_hash_type; // 0x28
    uint32_t im_root_hash_offset;  // 0x2C
    xid_t im_broken_xid;           // 0x30
    uint64_t im_reserved[9];       // 0x38
} integrity_meta_phys_t;           // 0x80
im_o: The object’s header
im_version: The version of the data structure
im_flags: The configuration flags
im_hash_type: The hash algorithm that is used
im_root_hash_offset: The offset (in bytes) of the root hash relative to the start of the object
im_broken_xid: The identifier of the transaction that unsealed the volume
im_reserved: reserved (only in version 2 or above)
Integrity Metadata Flags
Name	Value	Description
APFS_SEAL_BROKEN	0x00000001	The volume was modified after being sealed, breaking its seal
Hash Types
Name	Value	Description
APFS_HASH_INVALID	0	An invalid hash algorithm
APFS_HASH_SHA256	0x1	The SHA-256 variant of Secure Hash Algorithm 2
APFS_HASH_SHA512_256	0x2	The SHA-512/256 variant of Secure Hash Algorithm 2
APFS_HASH_SHA384	0x3	The SHA-384 variant of Secure Hash Algorithm 2
APFS_HASH_SHA512	0x4	The SHA-512 variant of Secure Hash Algorithm 2
File System Tree
Sealed Volumes can ensure integrity by hashing the contents of their File System Trees. This hashing necessitates some slight differences to the B-Tree. These modified B-Trees can be identified by the BTREE_HASHED and BTREE_NOHEADER flags being set in their B-Tree Info.

In standard B-Trees, non-leaf nodes store the object identifier of their children in the value-half of their entries. “Hashed” B-Trees instead use btn_index_node_val_t structures for this purpose, which store the cryptographic hash of the child node’s contents along with its identifier. Hashed nodes are also stored as headerless objects, with their 32-byte header being zeroed out.

#define BTREE_NODE_HASH_SIZE_MAX 64

typedef struct btn_index_node_val {
    oid_t binv_child_oid;                              // 0x00
    uint8_t binv_child_hash[BTREE_NODE_HASH_SIZE_MAX]; // 0x08
} btn_index_node_val_t;                                // 0x48
binv_child_oid: The object identifier of the child node
binv_child_hash: The hash of the child node
Data Stream Extents
As we discussed yesterday, Data Streams store their extents as file system records in the File System Tree. Sealed Volumes store extents in a separate File Extent Tree, whose virtual object identifier is stored in the apfs_fext_tree_oid of the Volume Superblock.

The key-half of the File Extent Tree entries are fext_tree_key_t structures and are sorted first by private_id and then by logical_addr.

typedef struct fext_tree_key {
    uint64_t private_id;   // 0x00
    uint64_t logical_addr; // 0x08
} fext_tree_key_t;         // 0x10
private_id: The object identifier of the file
logical_addr: The offset (in bytes) within the file’s data for the data stored in this extent
The value-half takes the form of a fext_tree_val_t structure. Its fields are interpreted in the same way as the j_file_extent_val fields. There is no crypto_id because sealed system volumes are never encrypted.

typedef struct fext_tree_val {
    uint64_t len_and_flags;  // 0x00
    uint64_t phys_block_num; // 0x08
} fext_tree_val_t;           // 0x10
len_and_flags: A bit field that contains the length of the extent and its flags
phys_block_num: The starting physical block address of the extent
Conclusion
Sealed Volumes in APFS provide an extra layer of security by allowing macOS to verify its system volume cryptographically. This post described some of the subtle differences in analyzing sealed volumes.

2022 APFS Advent Challenge Day 15 - Keybags
Wednesday, December 21, 2022
APFS is designed with encryption in mind and removes the need for the Core Storage layer used to provide encryption in HFS+. When you enable encryption on a volume, the entire File System Tree and the contents of files within that volume are encrypted. The type of encryption depends on the capabilities of the hardware that it is running on. For example, hardware encryption is used for internal storage on devices that support it, such as macOS computers with T2, M1, or M2 security chips and all iOS devices. Software encryption is used for external and internal storage devices without hardware encryption support. It’s worth noting that when hardware encryption is used, the data cannot be decrypted on any other device. For our purposes, we will focus on the software encryption mechanisms used in macOS. The hardware encryption functions similarly, but the security chip must broker all decryption operations.

Encryption Keys
In macOS, APFS uses a single Volume Encryption Key (VEK) to access encrypted content on a volume. This VEK is stored on disk wrapped in several layers of encryption that allow any authorized user on the system to access the volume’s contents. In addition, several recovery keys can be used to access the VEK.

The VEK is stored encrypted on disk by a Key Encryption Key (KEK). Multiple copies of the KEK are stored on disk, each encrypted (wrapped) with a different key to allow indirect access to the VEK by various users on a system. The keys that are used to encrypt the KEK can be derived from the following:

Each user’s password
The drive’s Personal Recovery Key
An organization’s Institutional Recovery Key
Each user’s iCloud Recovery Key
These wrapped keys are stored securely on disk in encrypted objects known as Keybags.

Keybags
Once decrypted, a Keybag is stored as a media_keybag_t structure on disk.

// A keybag object
typedef struct media_keybag {
    obj_phys_t mk_obj;     // 0x00
    kb_locker_t mk_locker; // 0x20
mk_obj: The object’s header
mk_locker: The keybag data
The main component of a Keybag is a kb_locker_t structure.

#define APFS_KEYBAG_VERSION 2

// A keybag
typedef struct kb_locker {
    uint16_t kl_version;         // 0x00
    uint16_t kl_nkeys;           // 0x02
    uint32_t kl_nbytes;          // 0x04
    uint8_t padding[8];          // 0x08
    keybag_entry_t kl_entries[]; // 0x10
} kb_locker_t;
kl_version: The keybag’s version (currently always 2)
kl_nkeys: Number of entries stored in the keybag
kl_nbytes: The size (in bytes) of the data stored in the kl_entries field
padding: reserved
kl_entries: The start of the entries
Immediately following the kb_locker_t structure is a keybag_entry_t structure for the first entry in the Keybag. After this structure is the data for the entry, followed by the structure for the next entry.

// An entry in a keybag
typedef struct keybag_entry {
    uuid_t ke_uuid;       // 0x00
    uint16_t ke_tag;      // 0x10
    uint16_t ke_keylen;   // 0x12
    uint8_t padding[4];   // 0x14
    uint8_t ke_keydata[]; // 0x18
} keybag_entry_t;
ke_uiid: A context-specific UUID that identifies the entry
ke_tag: A description of the kind of data stored in this keybag entry
ke_keylen: The length (in bytes) of the keybag entry’s data
padding: reserved
ke_keydata: ke_keylen bytes of entry data
Keybag Tags
Name	Value	Description
KB_TAG_UNKNOWN	0	reserved (never found on disk)
KB_TAG_RESERVED_1	1	reserved
KB_TAG_VOLUME_KEY	2	The key data stored a wrapped VEK
KB_TAG_VOLUME_UNLOCK_RECORDS	3	In a container’s keybag, the key data stores the location of the volumeʼs keybag; in a volume keybag, the key data stores a wrapped KEK.
KB_TAG_VOLUME_PASSPHRASE_HINT	4	The key data stores a user’s password hint as plain text
KB_TAG_WRAPPING_M_KEY	5	The key data stored a key that’s used to wrap a media key
KB_TAG_VOLUME_M_KEY	6	The key data stored a key that’s used to wrap volume media keys
KB_TAG_RESERVED_F8	0xF	reserved
Container Keybags
The nx_keylocker field of the container’s NX Superblock is used to locate the encrypted blocks on disk that store the Container Keybag. The XTS-AES-128 encryption key is a 256-bit key, derived from the container’s UUID. Read the 128-bit UUID from the nx_uuid field of the NX Superblock and concatinate it with itself.

container_keybag_key = container_uuid + container_uuid
Once decrypted the container keybag stores the location of each encrypted volume’s keybag as well as the wrapped VEK for each.

Volume Unlock Records
Volume Unlock Records are stored in the Container Keybag with a ke_tag value of KB_TAG_VOLUME_UNLOCK_RECORDS. The ke_uuid field stores the same UUID as the apfs_vol_uuid field of the Volume Superblock. The ke_keydata is a prange_t structure that gives the location of the encrypted blocks for the volume’s keybag.

Wrapped VEK
The wrapped VEK of a volume is stored in the Container Keybag with a ke_tag volume of KB_TAG_VOLUME_KEY with the ke_uuid also being the same as the volume’s UUID. This KEK must be unwrapped using the Key Encryption Key.

Volume Keybags
Each encrypted volume has its own keybag that stores the wrapped KEKs needed to access the VEK. For software encrypted APFS, these keybags are encrypted in the same fashion as the container keybags, using two copies of the volume’s UUID as a 256-bit XTS-AES-128 encryption key. Volume keybags can also store human-readable hints to remind user’s of their passphrases.

Volume Unlock Records
In the context of Volume Keybags, Volume Unlock Records store DER encoded information about wrapped KEKs. The ke_tag value is always KB_TAG_VOLUME_UNLOCK_RECORDS and the ke_uuid is either a cryptograpic user’s UUID or a hardcoded value to denote a recovery key. We will discuss more about the wrapped keys that are found in the ke_keydata field in tomorrow’s post.

Recovery Key UUIDs

Name	UUID
INSTITUTIONAL_RECOVERY_UUID	{C064EBC6-0000-11AA-AA11-00306543ECAC}
INSTITUTIONAL_USER_UUID	{2FA31400-BAFF-4DE7-AE2A-C3AA6E1FD340}
PERSIONAL_RECOVERY_UUID	{EBC6C064-0000-11AA-AA11-00306543ECAC}
ICLOUD_RECOVERY_UUID	{64C0C6EB-0000-11AA-AA11-00306543ECAC}
ICLOUD_USER_UUID	{EC1C2AD9-B618-4ED6-BD8D-50F361C27507}
Passphrase Hints
Passphrase Hint Records are stored with the ke_tag value of KB_TAG_VOLUME_PASSPHRASE_HINT and a cryptographic user’s UUID. The ke_keydata field contains a null-terminated UTF-8 string with the user’s provided passphrase hint.

Conclusion
This post discusses a general overview of APFS Keybags and their on-disk structures. In our next post, we will discuss methods of unwrapping and using the decryption keys.

2022 APFS Advent Challenge Day 16 - Wrapped Keys
Thursday, December 22, 2022
In our last post, we discussed both [Volume and Container Keybags](/post/2022/12/21/APFS-Keybags and how they protect wrapped Volume Encryption and Key Encryption Keys. Depending on whether the encrypted volume was migrated from an HFS+ encrypted Core Storage volume, there are subtle differences in how these keys are used. In this post, we will discuss the structure of these wrapped keys and how they can be used to access the raw Volume Encryption Keys that encrypt data on the file system.

Key Encryption Key Blobs
Each Key Encryption Key (KEK) is encoded in a binary DER blob with the following structure:

KEKBLOB ::= SEQUENCE {
    unknown [0] INTEGER
    hmac    [1] OCTET STRING
    salt    [2] OCTET STRING
    keyblob [3] SEQUENCE {
        unknown     [0] INTEGER
        uuid        [1] OCTET STRING 
        flags       [2] INTEGER
        wrapped_key [3] OCTET STRING
        iterations  [4] INTEGER
        salt        [5] OCTET STRING
    }
}
The keys begin with a header that contains an HMAC-SHA256 hash of the key blob data. The HMAC key is generated from the SHA-256 hash of a magic value concatenated with the given salt.

hmac_key := SHA256("\x01\x16\x20\x17\x15\x05" + salt)
The key blob encodes the wrapped KEK and additional information needed for unwrapping, including a set of bit-flags.

KEK Flags
Name	Value	Description
KEK_FLAG_CORESTORAGE	0x00010000’0000000000	Key is a legacy CoreStorage KEK
KEK_FLAG_HARDWARE	0x00020000’0000000000	Key is hardware encrypted
If the KEK_FLAG_CORESTORAGE flag is set, then the wrapped KEK was migrated from a Core Storage encrypted HFS+ volume and used a 128-bit key to encrypt the KEK; otherwise, a 256-bit key is used.

Generate a key using the PBKDF2-HMAC-SHA256 algorithm, the user’s password, the provided salt, and the number of iterations.

// Calculate size of wrapping key (in bytes)
key_size := (flags & KEK_FLAG_CORESTORAGE) ? 16 : 32

// Generate unwrapping key from user's password
key := pbkdf2_hmac_sha256(password, salt, iterations, key_size)

// Unwrap the encrypted KEK
kek := rfc3394_unwrap(key, wrapped_key);
If the encrypted volume was migrated from Core Storage and the user changed their password afterward, it’s possible to have a non-Core-Storage wrapped KEK containing only a 128-bit key. In these instances, the last 128 bits of the unwrapped KEK will be zeros and should be ignored.

// Shorten the KEK if needed
if is_zeroed(kek[16:]) {
    kek = kek[:16];
}
Volume Encryption Key Blobs
Volume Encryption Key (VEK) blobs have a very similar structure to the KEK blobs that we just discussed. Depending on if they were migrated from Core Storage, they can also be 128-bit or 256-bit keys.

VEKBLOB ::= SEQUENCE {
    unknown [0] INTEGER
    hmac    [1] OCTET STRING
    salt    [2] OCTET STRING
    keyblob [3] SEQUENCE {
        unknown     [0] INTEGER
        uuid        [1] OCTET STRING
        flags       [2] INTEGER
        wrapped_key [3] OCTET STRING
    }
}
VEK Flags
Name	Value	Description
VEK_FLAG_CORESTORAGE	0x00010000’0000000000	Key is a legacy CoreStorage VEK
VEK_FLAG_HARDWARE	0x00020000’0000000000	Key is hardware encrypted
Use the KEK to unwrap the VEK using the RFC3394 key wrapping algorithm. If the wrapped VEK is a 128-bit Core Storage VEK, then only the first 128-bits of the KEK are used.

// Calculate size of wrapping key (in bytes)
vek_size = (flags & VEK_FLAG_CORESTORAGE) ? 16 : 32;

if (vek_size == 16) {
    kek = kek[:16];
}

// Unwrap the VEK
vek = rfc3394_unwrap(vek, wrapped_key)
128-bit Core Storage VEKs must be extended to 256-bit encryption keys. This is accomplished by using the first 128 bits of the SHA256 hash of the VEK and its UUID as the second half of the key.

// 128-bit veks need to be combined with the first 128-bits of a hash
if vek_size == 16 {
    vek = append(vek, SHA256(vek + uuid)[16:])
}
Conclusion
In this post, we discussed utilizing the wrapped keys stored in APFS key bags to gain access to the Volume Encryption Key that protects a user’s data in APFS. Tomorrow, we will conclude our discussion about APFS encryption by describing how to identify and decrypt protected information using these keys.

2022 APFS Advent Challenge Day 17 - Blazingly Fast Checksums with SIMD
Friday, December 23, 2022
Today’s post will take on a bit of a different style than the previous posts in this series. Among other things, I spent my day putting off writing the final APFS encryption blog post by pursuing another one of my New Year goals. Along the way, I wrote a Fletcher64 hashing function that can validate APFS objects at over 31 GiB/s on my 2017 iMac Pro. Rather than fighting my procrastination, I decided it would be better to share my findings. Given that my chosen learning path was directly relevant to APFS, I’m counting this as a valid APFS Advent Challenge post (and you can’t stop me!). I hope you enjoy this brief detour into the dark arts of cross-platform SIMD programming.

SIMD Background
I’ve recently become interested in learning more about SIMD programming and how to utilize it to make my code faster. SIMD stands for “Single Instruction, Multiple Data.” It is a technique used in computer architectures to perform the same operation on multiple data elements in parallel, using a single instruction.

Here’s an example to help illustrate the concept:

Imagine that you have a list of numbers and want to add 1 to each of them. Without SIMD, you would have to write a loop that goes through each number in the list and performs an increment operation. This may be a very time-consuming process if the list is long.

On the other hand, if your computer has SIMD support, it can simultaneously perform the same operation on multiple numbers using a single instruction. This process, known as vectorization, can significantly speed things up, especially for long lists of numbers. We’re not limited to simple increment operations; SIMD supports many arithmetic and logical operations on most architectures.

Speeding up Fletcher64
During my journey, I came across some prior work by James D Guilford and Vinodh Gopal describing using SIMD for the Fast Computation of Fletcher Checksums in ZFS. While ZFS uses a different variant of a Fletcher checksum than APFS, this seemed like a great first project to get my hands dirty with vectorization.

Portability Concerns
The authors of the Intel whitepaper use hand-coded AVX assembly instructions to perform their vectorization. OpenZFS seems to have taken the same approach. They have independent implementations full of inline assembly for Intel’s SSE, AVX2, AVX-512, and ARM’s NEON architectures. Apple takes a similar approach. The Intel version of apfs.kext contains an SSE, AVX, and AVX2 vectorized implementations and a fallback serialized version if, for some reason, none of these instruction sets are supported. The arm64 version of the KEXT uses NEON vectorization instructions.

While these approaches work and produce high-performing and optimized code, having to hand-code an implementation for each instruction set seems to defeat the purpose of writing code in a portable language like C++. Besides, I’m a programmer, and programmers, by our very nature, are lazy. The compiler knows what it’s doing and usually can generate better-performing code that I could hand optimize, so let’s find a way to let it do its job.

C++ TS N4808 and std::experimental::simd
It turns out that the C++ standardization committee, along with many brilliant minds, has been working on this problem for years. Document N4808, the Working Draft, C++ Extensions for Parallelism Version 2, is a proposal to add (among other things) support for portable data parallel types to the C++ standard library.

This technical document proposes a generalized model of the most common SIMD operations that standard library implementations can use to allow programmers to write vectorized code that can be compiled to architecture-specific instructions without requiring architecture-specific inline assembly. That sounds like exactly what we want! While this has not officially been added to the language, GCC’s libstdc++ and Clang’s libc++, have at least partial implementations in their std::experimental namespace. GCC support seems the most complete, so I decided to experiment with gcc-12.

The Implementation
std::experimental::simd allows you to define native C++ vector types whose storage capacity depends on the underlying target architecture. For example, NEON supports 128-bit SIMD registers, holding two 64-bit or four 32-bit integers. AVX2 supports twice the storage with 256-bit registers, and the aptly named AVX-512 supports 512-bit registers. We can write code once, and the size of the vectors will be architecture specific.

namespace stdx = std::experimental;

// SIMD vector of 64-bit unsigned integers
using vu64 = stdx::native_simd<uint64_t>;

// SIMD vector of 32-bit unsigned integers
using vu32 = stdx::native_simd<uint32_t>;
These SIMD vectors can be used almost exactly like native integer types, and once I got over the lack of documentation, I found that they were pretty easy to use. By taking lessons from the existing vectorized implementations and making some improvements of my own, this is what I was able to come up with:

// N, N-2, N-4, ..., 2
static const vu64 even_m{[](const auto i) { return vu32::size() - (2 * i); }};

// N-1, N-3, N-5, ..., 1
static const vu64 odd_m = even_m - 1;

static constexpr auto max32 = std::numeric_limits<uint32_t>::max();

static uint64_t fletcher64_simd(std::span<const uint32_t, 1024> words) {
  vu64 sum1{};
  sum1[0] = -(static_cast<uint64_t>(words[0]) + words[1]);

  vu64 sum2{};
  sum2[0] = words[1];

  for (size_t n = 0; n < words.size(); n += vu32::size()) {
    sum2 += vu32::size() * sum1;

    const vu64 all{reinterpret_cast<const uint64_t*>(std::addressof(words[n])),
                   stdx::vector_aligned};

    const vu64 evens = all & max32;
    const vu64 odds = all >> 32;

    sum1 += evens + odds;
    sum2 += evens * even_m + odds * odd_m;
  }

  // Fold the 64-bit overflow back into the 32-bit value
  const auto fold = [&](uint64_t x) {
    x = (x & max32) + (x >> 32);
    return (x == max32) ? 0 : x;
  };

  const uint64_t low = fold(stdx::reduce(sum1));
  const uint64_t high = fold(stdx::reduce(sum2));

  const uint64_t ck_low = max32 - ((low + high) % max32);
  const uint64_t ck_high = max32 - ((low + ck_low) % max32);

  return ck_low | ck_high << 32;
}
Results
Below are the speed comparisons between the above SIMD function and the following serialized implementation (non-threaded, single core performance). The times reported are the average time per checksum calculation of a 4KiB APFS object.

static uint64_t fletcher64_serial(std::span<const uint32_t, 1024> words) {
  uint64_t sum1 = -(static_cast<uint64_t>(words[0]) + words[1]);
  uint64_t sum2 = words[1];

  for (const uint32_t word : words) {
    sum1 += word;
    sum2 += sum1;
  }

  sum1 %= max32;
  sum2 %= max32;

  const uint64_t ck_low = max32 - ((sum1 + sum2) % max32);
  const uint64_t ck_high = max32 - ((sum1 + ck_low) % max32);

  static constexpr size_t high_shift = 32;
  return ck_low | ck_high << high_shift;
}
My 2017 iMac Pro supports enabling 128-bit SSE, 256-bit AVX2, and 512-bit AVX-512, so it’s a great candidate to show the speedups that can be achieved via vectorization.

Target Architecture	Time per Checksum	Throughput	Speedup
Serialized	730ns	5.21734 GiB/s	-
SSE	509ns	7.49126 GiB/s	1.4x
AVX2	292ns	13.0277 GiB/s	2.5x
AVX-512	122ns	31.1448 GiB/s	6x
The relative performance of my 2021 M1 Max MacBook Pro is somewhat less impressive due to the ARM NEON architecture being limited to only 128-bit vector registers. This computer is still very fast, and I love it.

Target Architecture	Time per Checksum	Throughput	Speedup
Serialized	458ns	8.31391 GiB/s	-
NEON	368ns	10.3417 GiB/s	1.2x
Conclusion
For the proper application, SIMD vectorization can provide fantastic performance benefits. In my testing, I demonstrated a 6x speedup and hashed APFS objects at over 31 Gigabytes per second on an iMac Pro from 2017! The proposed SIMD additions to the C++ standard library are easy to use and generate high-performing, portable code. I absolutely will be using this whenever I can.

Update (December 24, 2022)
I further improved this code’s performance to achieve even better performance!

2022 APFS Advent Challenge Day 18 - Decryption
Monday, December 26, 2022
Now that we know how to parse the File System Tree, Analyze Keybags, and Unwrap Decryption Keys, it’s time to put it all together and learn how to decrypt file system metadata and file data on encrypted volumes in APFS.

Tweaks
All encryption in APFS is based on the XTS-AES-128 cipher, which uses a 256-bit key and a 64-bit “tweak” value. This tweak value is position dependent. It allows the same plaintext to be encrypted and stored in different locations on disk and have drastically different ciphertext while using the same AES key. Every 512 bytes of encrypted data uses a tweak based on the container offset of the block’s initial storage.

Knowledge of the AES key alone is not always enough for successful decryption. If the encrypted block is ever relocated on disk, the data is not guaranteed to be re-encrypted with a new tweak. In these cases, the tweak can not be inferred based on the block’s on-disk location, so we must learn the original tweak value used for encryption.

Identifing Encrypted Blocks
There are primarily two sets of data protected with the APFS Volume Encryption Key: File System Tree Nodes and File Extents. As we’ve discussed, File System Tree Nodes store the File System Records that contain the file system’s metadata, and File Extents contain the bulk of the data stored in a file’s Data Streams.

Encrypted FS-Tree Nodes
A volume’s Object Map is never encrypted, but its referenced virtual objects may be, as is the case with FS-Tree Node on encrypted volumes.

Let’s revisit the value half of an Object Map entry.

typedef struct omap_val {
  uint32_t ov_flags; // 0x00
  uint32_t ov_size;  // 0x04
  paddr_t ov_paddr;  // 0x08
} omap_val_t;        // 0x10
If the ov_flags bit-field member has the OMAP_VAL_ENCRYPTED flag set, then the virtual object located at ov_paddr is encrypted. These objects are never related without being re-encrypted, so the tweak of the first 512 bytes of data can be determined by the physical location of the data using the following logic, with the following tweak values incremented for each subsequent 512 bytes of data:

uint64_t tweak0 = (ov_paddr * block_size) / 512;
Encrypted Extents
Extent data can be relocated on disk and is not guaranteed to be re-encrypted. Due to this, the initial tweak value is stored in the crypto_id field of the j_file_extent_val_t file system record:

typedef struct j_file_extent_val {
  uint64_t len_and_flags;  // 0x00
  uint64_t phys_block_num; // 0x08
  uint64_t crypto_id;      // 0x10
} j_file_extent_val_t;     // 0x18
Conclusion
We’ve now discussed all of the information needed to access data on software-encrypted APFS volumes. This decryption requires the knowledge of the password of any user on the system or one of the various recovery keys. While APFS hardware encryption works in largely the same manner, the encryption also depends on keys that are stored within the specific security chip on a given system. There are currently no known methods of extracting these chip-specific keys; therefore, the data on hardware-encrypted devices must be decrypted at acquisition time on the device itself. The only software that I am aware of that is capable of this is Cellebrite’s Digital Collector.

Full disclosure: I currently work for Cellebrite and helped develop these capabilities. I do not directly profit from the sales of Digital Collector but felt it appropriate to disclose my association when linking to a commercial product. I am not trying to sell you anything. Unfortunately, I am also not at liberty to discuss the methodology used to facilitate this decryption.

2022 APFS Advent Challenge Day 20 - Snapshot Metadata
Wednesday, December 28, 2022
Our previous discussion discussed how Object Maps facilitate the implementation of point-in-time Snapshots of APFS file systems by preserving File System Tree Nodes from earlier transactions. In that discussion, I outlined the on-disk structure of the Object Map Snapshot Tree and how it can be used to enumerate the transaction identifiers of each Volume Snapshot. Today, we will briefly discuss two other sources of information that store additional metadata about each Snapshot.

Snapshot Metadata Tree
The Snapshot Metadata Tree is a B-Tree whose physical address can be located by reading the apfs_snap_meta_tree_oid field of the Volume Superblock. It stores two types of objects, structured as File System Records.

Snapshot Metadata Records
Snapshot Metadata Records store the bulk of metadata about Volume Snapshots. The key-half is a j_snap_metadata_key structure with an encoded type of APFS_TYPE_SNAP_METADATA.

typedef struct j_snap_metadata_key {
  j_key_t hdr;           // 0x00
} j_snap_metadata_key_t; // 0x08
hdr: The record’s header. The object identifier in the header is the snapshot’s transaction identifier.
The value-half of the record is a j_snap_metadata_val_t structure and is immediately followed by the UTF-8 encoded name of the snapshot.

typedef struct j_snap_metadata_val {
  oid_t extentref_tree_oid;       // 0x00
  oid_t sblock_oid;               // 0x08
  uint64_t create_time;           // 0x10
  uint64_t change_time;           // 0x18
  uint64_t inum;                  // 0x20
  uint32_t extentref_tree_type;   // 0x28
  uint32_t flags;                 // 0x2C
  uint16_t name_len;              // 0x30
  uint8_t name[0];                // 0x32
} j_snap_metadata_val_t;
extentref_tree_oid: The physical object identifier of the B-Tree that stores extent references for the snapshot.
sblock_oid: The physical object identifier of a backup of the snapshot’s Volume Superblock
create_time: The time when the snapshot was created
change_time: The time that this snapshot was last modified
inum: reserved
extentref_tree_type: The type of the Extent Reference Tree
flags: A bit field that contains additional information about a snapshot metadata record
name_len: The length of the name that follows this structure (in bytes)
Snapshot Metadata Record Flags
Name	Value	Description
SNAP_META_PENDING_DATALESS	0x00000001	This snapshot is dataless, meaning that it does not preserve the file extents
SNAP_META_MERGE_IN_PROGRESS	0x00000002	The snapshot is in the process of being merged with another
Snapshot Name Records
Snapshot Name Records are used to map snapshot names to their transaction identifiers. The key-half of the record is a j_snap_name_key_t structure with an encoded type of APFS_TYPE_SNAP_NAME. It is followed by the UTF-8 encoded name of the snapshot.

typedef struct j_snap_name_key {
  j_key_t hdr;        // 0x00
  uint16_t name_len;  // 0x08
  uint8_t name[0];    // 0x0A
} j_snap_name_key_t;
hdr: The record’s header. The object identifier can be ignored.
name_len: The length of the name (in bytes)
name: The start of the UTF-8 encoded name
The value-half is a j_snap_name_val_t structure.

typedef struct j_snap_name_val {
  xid_t snap_xid;    // 0x00
} j_snap_name_val_t; // 0x08
snap_xid: The transaction identifier of the snapshot
Snapshot Extended Metadata Object
Each snapshot has a virtual Snapshot Extended Metadata Object in the volume’s Object Map. The virtual object identifier of this object is stored in the apfs_snap_meta_ext_oid field of the Volume Superblock. There are multiple versions of this object whose transaction identifiers correspond to each snapshot.

typedef struct snap_meta_ext_obj_phys {
  obj_phys_t smeop_o;        // 0x00
  snap_meta_ext_t smeop_sme; // 0x20
} snap_meta_ext_obj_phys_t;  // 0x48
smeop_o: The object’s header
smeop_sme: The snapshot’s extended metadata
typedef struct snap_meta_ext {
  uint32_t sme_version; // 0x00
  uint32_t sme_flags;   // 0x04
  xid_t sme_snap_xid;   // 0x08
  uuid_t sme_uuid;      // 0x10
  uint64_t sme_token;   // 0x20
} snap_meta_ext_t;      // 0x28
sme_version: The version of this structure (currently 1)
sme_flags: A bitfield of flags (none are currently defined)
sme_snap_xid: The transaction identifier of the snapshot
sme_uuid: The unique identifier of the snapshot
sme_token: An opaque token (reserved)

2022 APFS Advent Challenge Day 21 - Fusion Containers
Thursday, December 29, 2022
As we discussed in an earlier post, Apple’s Fusion Drives combine the storage capacity of a hard disk drive (HDD) with the faster access speed of a solid state drive (SSD). The HDD is the primary storage device, and the SSD acts as a cache for recently accessed data. However, the Fusion Drive does not have built-in caching logic, and the operating system treats the two drives as separate storage devices. Apple created Core Storage to support the desired caching capabilities and the ability to pool the storage of each device into a single logical volume. APFS removes the need for Core Storage by having first-class support for this tiered storage model. This post will go into more detail about APFS Fusion Containers.

Physical Stores
Both the SSD and HDD of a Fusion Drive appear to macOS as separate physical disk devices. Both disks are GPT partitioned with a standard EFI partition and a second, larger partition, which takes up the bulk of the space on disk. For example, running the command diskutil list may show the HDD as /dev/disk0 with its primary partition as /dev/disk0s2 and the SSD as /dev/disk1 and /dev/disk1s2. These two partitions make up the physical stores of the Fusion Container.

Each physical store is formatted separately in much the same way as any other APFS container. Both will share the same nx_uuid in their [NX _Superblocks and have a separate, nearly-identical UUID in the nx_fusion_uuid field, with the most significant bit being cleared on the tier1 SSD partition and set on the tier2 HDD partition. The combination of these UUIDs can be used to identify the physical storage tiers of the container.

Synthesized Container
Both tiers are mapped together as a single “synthesized” container and are presented to macOS as a single logical block device (for example, /dev/disk2). The tier1 blocks are mapped at logical byte offset zero, and the tier2 blocks at 4 EiB. The offsets within the exabyte-scale gap between the two sets of blocks cannot be read.

APFS objects and blocks can be stored on either (or both) tiers, and their physical addresses will require some simple translation as follows:

#define FUSION_TIER2_DEVICE_BYTE_ADDR 0x4000000000000000ULL
const paddr_t first_tier2_block = FUSION_TIER2_DEVICE_BYTE_ADDR / nxsb->block_size;

if (paddr < first_tier2_block) {
  tier1->read_block(paddr); 
} else {
  tier2->read_block(paddr – first_tier2_block);
}
The logically exabyte-scale gap separating the two tiers presents a unique problem during digital forensic imaging Fusion Containers. To preserve the logical offsets of the evidence without having to use a data center worth of storage, you must use an evidence storage format that supports sparse imaging. As long as this is considered along with the additional physical address translation described above, analyzing fusion containers does not generally differ from other APFS containers.

COMPREHENSIVE LIST OF APPLE DOCUMENTATION OMISSIONS

  Based on my systematic analysis of the blog vs Apple's documentation, here are the critical
  omissions:

  1. ALGORITHMIC IMPLEMENTATION DETAILS

  ❌ Missing from Apple Docs:
  - Fletcher-64 Algorithm: Apple mentions "Fletcher 64 checksum" but provides NO implementation
  details
    - No mention it uses 32-bit words (not 64-bit)
    - No modulo operations with UINT32_MAX
    - No proper checksum calculation formula
    - No mention of 1024-word chunking for modulo

  ❌ Missing from Apple Docs:
  - B-tree Traversal Algorithms: No detailed traversal procedures
  - Object Map Search Logic: No search algorithms for virtual object resolution
  - Checkpoint Discovery Process: No step-by-step mounting procedures

  2. CRITICAL FIELD DESCRIPTIONS

  ❌ Apple Error - Object Map Fields:
  Blog states: "Apple's APFS File System Reference incorrectly lists the om_tree_oid and
  om_snapshot_tree_oid members as virtual object identifiers when they are, in fact, physical."

  ❌ Apple Omission - Missing Field Descriptions:
  - omap_phys_t structure has NO field descriptions in Apple docs
  - Many other structures lack complete field explanations

  3. UNDOCUMENTED OBJECT TYPES

  ❌ Missing from Apple Docs:
  - M-Tree Type (0x04): Blog reveals this was in Apple's kernel but never documented
  - Comprehensive Object Type Table: Blog provides complete list with descriptions

  4. ENCRYPTION IMPLEMENTATION DETAILS

  ❌ Missing from Apple Docs:
  - Tweak Calculation Formula: uint64_t tweak0 = (ov_paddr * block_size) / 512;
  - Key Wrapping Details: DER blob structures and unwrapping procedures
  - Hardware vs Software Encryption Differences
  - Recovery Key UUIDs: Specific hardcoded values

  5. PRACTICAL IMPLEMENTATION GUIDANCE

  ❌ Missing from Apple Docs:
  - Container Mounting Procedures: Step-by-step process for finding latest checkpoint
  - Object Resolution Workflows: How to actually resolve virtual objects
  - Error Handling: What to do when objects are malformed
  - DMG Handling: No mention of disk image specifics

  6. PERFORMANCE OPTIMIZATIONS

  ❌ Missing from Apple Docs:
  - SIMD Fletcher-64 Implementation: 31 GiB/s performance techniques
  - B-tree Optimization Strategies: Minimal depth traversal
  - Caching Approaches: Block-level caching strategies

  7. FUSION DRIVE SPECIFICS

  ❌ Missing from Apple Docs:
  - Tier Address Translation: FUSION_TIER2_DEVICE_BYTE_ADDR = 0x4000000000000000ULL
  - Physical Store Mapping: How SSD/HDD tiers are combined
  - UUID Bit Patterns: MSB differences between tiers

  8. FORENSIC ANALYSIS DETAILS

  ❌ Missing from Apple Docs:
  - Sparse Imaging Requirements: For fusion containers
  - Evidence Preservation: Logical offset considerations
  - Recovery Procedures: Finding older checkpoints when current is malformed

  9. OBJECT FLAGS AND BEHAVIORAL DETAILS

  ❌ Missing from Apple Docs:
  - Complete Flag Interpretations: Object type flag combinations
  - Behavioral Specifications: How flags affect object handling
  - Validation Requirements: When and how to validate checksums

  10. VERSION-SPECIFIC INFORMATION

  ❌ Missing from Apple Docs:
  - Historical Changes: M-Tree removal in macOS 13
  - Kernel Implementation Details: What Apple's actual implementation does
  - Version Compatibility: Which features require which APFS versions