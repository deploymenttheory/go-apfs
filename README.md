# afps

**afps** is a cross-platform, read-only command-line tool for exploring, extracting, recovering, and validating Apple File System (APFS) volumes — directly from raw disks, partitions, or `.dmg` images, without mounting or relying on macOS.

---

## Purpose

APFS is a sophisticated filesystem with support for snapshots, encryption, crash-safe structures, and sparse containers. But outside macOS, there's little tooling to inspect or recover data from it. `afps` fills that gap by parsing APFS **on-disk structures directly**, without requiring kernel extensions, drivers, or mounted volumes.

This makes it ideal for:

- Data recovery engineers
- Forensic analysts
- Backup verification tools
- Security auditing workflows
- Developers analyzing `.dmg` payloads

---

## Key Features

### General

- 📦 Works with physical disks, partitions, and `.dmg` images
- 🔍 Explore containers, volumes, snapshots, and files
- 🗃️ Extract single files, directories, or full volumes
- 🧯 Recover deleted files from unmounted APFS volumes
- 🔐 Inspect encryption metadata and protection classes
- ✅ Fully **read-only** and **cross-platform**
- 🚫 Does **not mount** anything

---

### Volume and Container Inspection

- Discover APFS containers and volumes
- Show volume metadata, flags, block sizes, roles
- Read checkpoints and recover historical states

### File Extraction

- Extract:
  - A single file
  - A directory (with optional recursion)
  - An entire volume
- Preserve metadata and extended attributes
- Extract files from snapshots or specific checkpoints

### Snapshot Management

- List available snapshots in a volume
- Extract contents of a snapshot
- Compare snapshots (planned)

### Deleted File Recovery

- Recover deleted files via extent and inode scanning
- Filter by filename, path, time, or type
- Dump recovered files to `lost+found` structure

### Filesystem Integrity Verification

- Verify object checksums (Fletcher-64)
- Detect corruption in containers, volumes, and trees
- Validate free space bitmap and space manager

### Encryption Metadata Inspection

- Decode encryption state, keybags, and protection classes
- Read encryption metadata from volumes or snapshots
- Operates even if file contents are encrypted (read-only)

### `.dmg` Support

- Locate and extract APFS volumes from within `.dmg` files
- Support for raw, sparse, and compressed images (planned)
- All other operations (extract, inspect, recover) work the same on embedded volumes

---

## Example Usage

```bash
# Show info about all APFS volumes on a device
afps list --device /dev/disk2

# Extract a folder recursively
afps extract --src /Users/alice/Documents --out ./backup --recursive

# Recover deleted .jpg files from a volume
afps recover --filter '*.jpg' --out ./lostfound

# List snapshots and extract one
afps list-snapshots --volume /dev/disk3s1
afps extract --src / --snapshot Snap1 --out ./Snap1-root

# Work with a .dmg image
afps extract --from-dmg ./mac_backup.dmg --src /Library --out ./lib_dump
```

Why No Mounting?
Unlike tools like mount, hdiutil, or fuse-apfs, afps does not mount the filesystem. Instead, it reads the disk structures directly:

Parses superblocks, B-trees, and extent trees in user space

Avoids kernel drivers, FUSE, or macOS-only APIs

Prevents any chance of modification or write-back

Enables full support for Linux and Windows

## Status

afps is under active development. Many core features are implemented or in-progress. Contributions welcome!