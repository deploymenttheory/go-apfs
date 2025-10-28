go-apfs/
├── DOCUMENTATION
│   ├── README.md                        # Project overview, usage examples
│   ├── docs/
│   │   ├── Apple-File-System-Reference.md  # ⭐ OFFICIAL APFS SPEC (36K+lines)
│   │   ├── Apple-File-System-Reference.pdf
│   │   ├── architecture.md               # Project architecture (437 lines)
│   │   ├── SERVICES_IMPLEMENTATION_PLAN.md # Service priorities (284 lines)
│   │   └── dependancies.mermaid
│   ├── CHANGELOG.md
│   ├── CODE_OF_CONDUCT.md
│   ├── CONTRIBUTING.md
│   ├── SECURITY.md
│   ├── LICENSE
│   └── apfs-config.yaml                 # Configuration template
│
├── CORE IMPLEMENTATION
│   ├── go.mod / go.sum                  # Dependencies (google/uuid, viper, testify)
│   ├── main.go                          # CLI entry point
│   │
│   └── internal/
│       │
│       ├── 🔹 INTERFACES/ (22 files)
│       │   └── Contract definitions for all APFS subsystems
│       │       ├── analysis.go
│       │       ├── block_device.go      # Block I/O abstraction
│       │       ├── btree.go             # B-tree operations
│       │       ├── container.go         # Container management
│       │       ├── data_recovery.go     # Deleted file recovery
│       │       ├── data_streams.go      # File extent/stream access
│       │       ├── efi_jumpstart.go     # EFI boot driver
│       │       ├── encryption.go        # Encryption operations
│       │       ├── encryption_rolling.go# Rolling encryption
│       │       ├── extended_fields.go   # Metadata fields
│       │       ├── file_system_objects.go
│       │       ├── filesystem.go        # File traversal
│       │       ├── fusion.go            # Hybrid storage
│       │       ├── mounting.go          # Mount operations
│       │       ├── object_map.go        # Virtual object resolution
│       │       ├── objects.go           # Core object operations
│       │       ├── reaper.go            # Garbage collection
│       │       ├── sealed_volumes.go    # Sealed/signed volumes
│       │       ├── siblings.go          # File cloning
│       │       ├── snapshot.go          # Snapshots
│       │       ├── space_manager.go     # Free space tracking
│       │       └── volumes.go           # Volume operations
│       │
│       ├── 🔹 TYPES/ (19 files) - SINGLE SOURCE OF TRUTH
│       │   └── APFS data structures matching official spec
│       │       ├── general_types.go     # Paddr, UUID, Prange
│       │       ├── objects.go           # obj_phys_t, oid_t, xid_t
│       │       ├── container.go         # nx_superblock_t
│       │       ├── btree.go             # btree_node_phys_t structures
│       │       ├── object_maps.go       # omap_phys_t, omap_key_t
│       │       ├── volumes.go           # apfs_superblock_t
│       │       ├── file_system_constants.go
│       │       ├── file_system_objects.go # Inodes, directory entries
│       │       ├── data_streams.go      # File extents, streams
│       │       ├── extended_fields.go   # Metadata fields
│       │       ├── encryption.go        # Crypto state, keybags
│       │       ├── encryption_rolling.go
│       │       ├── snapshot.go          # Snapshot metadata
│       │       ├── efi_jumpstart.go     # Boot structures
│       │       ├── space_manager.go     # Allocation tracking
│       │       ├── reaper.go            # Deletion tracking
│       │       ├── sealed_volumes.go    # Signature structures
│       │       ├── siblings.go          # Cloning metadata
│       │       └── fusion.go            # Fusion metadata
│       │
│       ├── 🔹 PARSERS/ (90+ files) - Binary format readers
│       │   └── Domain-specific parser implementations
│       │       │
│       │       ├── container/ (5 readers)
│       │       │   ├── container_superblock_reader*.go
│       │       │   ├── checkpoint_map_reader*.go
│       │       │   ├── checkpoint_mapping_reader*.go
│       │       │   ├── container_statistics_reader*.go
│       │       │   └── evict_mapping_reader*.go
│       │       │
│       │       ├── btrees/ (6 readers)
│       │       │   ├── btree_info_reader*.go
│       │       │   ├── btree_node_reader*.go
│       │       │   ├── btree_location_reader*.go
│       │       │   ├── btree_kv_location_reader*.go
│       │       │   ├── btree_kv_offset_reader*.go
│       │       │   └── btree_index_node_value_reader*.go
│       │       │
│       │       ├── volumes/ (25 readers!)
│       │       │   ├── volume_superblock_reader*.go
│       │       │   ├── volume_flags_reader*.go
│       │       │   ├── volume_features_reader*.go
│       │       │   ├── volume_encryption_reader*.go
│       │       │   ├── volume_identity_reader*.go
│       │       │   ├── volume_snapshot_metadata_reader*.go
│       │       │   ├── volume_space_management_reader*.go
│       │       │   ├── volume_checkpoint_management_reader*.go
│       │       │   ├── volume_quota_limits_reader*.go
│       │       │   ├── volume_incompatible_features_reader*.go
│       │       │   └── [15+ more specialized readers]
│       │       │
│       │       ├── file_system_objects/ (13 readers)
│       │       │   ├── inode_reader*.go
│       │       │   ├── directory_entry_reader*.go
│       │       │   ├── directory_stats_reader*.go
│       │       │   ├── extended_attribute_reader*.go
│       │       │   ├── jkey_reader*.go
│       │       │   └── [8+ more]
│       │       │
│       │       ├── data_streams/ (5 readers)
│       │       │   ├── file_extent_reader*.go
│       │       │   ├── physical_extent_reader*.go
│       │       │   ├── data_stream_reader*.go
│       │       │   └── [2+ more]
│       │       │
│       │       ├── object_maps/ (11 readers)
│       │       │   ├── omap_reader*.go
│       │       │   └── [10+ more specialized readers]
│       │       │
│       │       ├── encryption/ (3 readers)
│       │       │   ├── crypto_state_reader*.go
│       │       │   ├── keybag_reader*.go
│       │       │   └── media_keybag_reader*.go
│       │       │
│       │       ├── encryption_rolling/ (3 readers)
│       │       ├── extended_fields/ (4 readers)
│       │       ├── efi_jumpstart/ (10 readers)
│       │       ├── snapshot/ (6 readers)
│       │       ├── space_manager/ (10 readers)
│       │       ├── sealed_volumes/ (3 readers)
│       │       ├── siblings/ (2 readers)
│       │       ├── reaper/ (5 readers)
│       │       └── objects/ (6 readers)
│       │       
│       │       (* = includes _test.go files)
│       │
│       ├── 🔹 SERVICES/ (15 files)
│       │   └── High-level business logic
│       │       ├── interfaces.go         # Service contracts (6 services)
│       │       ├── models.go             # DTO structures
│       │       ├── container_reader.go   # Low-level block I/O
│       │       ├── btree_object_resolver.go  # Virtual object resolution
│       │       ├── volume_service.go     # Volume operations
│       │       ├── filesystem_service.go # File traversal
│       │       ├── checkpoint_discovery_service.go # Checkpoint recovery
│       │       └── [test files]
│       │
│       ├── 🔹 DEVICE/ (1 file)
│       │   └── device/dmg.go            # DMG image format support
│       │
│       └── 🔹 HELPERS/ (2 files)
│           └── Utility functions
│               ├── encryption.go        # Crypto utilities
│               └── encryption_test.go
│
├── TESTING
│   ├── tests/
│   │   ├── basic_apfs.dmg              # Test image (minimal APFS)
│   │   ├── empty_apfs.dmg              # Test image (empty volume)
│   │   ├── full_apfs.dmg               # Test image (complex structure)
│   │   ├── deleted_files_apfs.dmg      # Test image (deleted files)
│   │   └── populated_apfs.dmg          # Test image (full data)
│   │
│   └── scripts/
│       └── create_test_dmgs.sh         # Generate test images
│
└── BUILD
    ├── go.mod                          # Module definition
    └── go.sum                          # Dependency checksums