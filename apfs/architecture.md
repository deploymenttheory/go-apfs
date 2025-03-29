# APFS Project Architecture

## Overview

This architecture follows the layered design of Apple File System as described in the reference documentation, with a clear separation between the container layer and file-system layer.

### File Architecture

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

# APFS Implementation Architecture with Functions

## pkg/types/
- **constants.go**
  - APFS Magic Numbers (NX_MAGIC, APFS_MAGIC, etc.)
  - Object Types and Flags
  - B-tree Constants
  - File System Constants
  - Error Constants

- **container_types.go**
  - `type NXSuperblock struct`
  - `type OMapPhys struct`
  - `type BTNodePhys struct`
  - `type CheckpointMapPhys struct`
  - `type SpacemanPhys struct`
  - `type ReaperPhys struct`

- **fs_types.go**
  - `type APFSSuperblock struct`
  - `type JInodeVal struct`
  - `type JDrecVal struct`
  - `type JFileExtentVal struct`
  - `type JXattrVal struct`
  - `type JSnapMetadataVal struct`

- **interfaces.go**
  - `type BlockDevice interface`
  - `type Object interface`
  - `type FileSystem interface`
  - `type Transaction interface`
  - `type ObjectResolver interface`
  - `type KeyProvider interface`

- **common.go**
  - `type OID uint64`
  - `type XID uint64`
  - `type PAddr int64`
  - `type UUID [16]byte`
  - `type FileInfo struct`
  - `type DirectoryEntry struct`

## pkg/container/
- **object.go**
  - `func ReadObject(device BlockDevice, addr PAddr) (*Object, error)`
  - `func ValidateObjectChecksum(obj *Object) bool`
  - `func WriteObject(device BlockDevice, obj *Object) error`
  - `func ObjectTypeToString(objType uint32) string`

- **superblock.go**
  - `func ReadNXSuperblock(device BlockDevice, addr PAddr) (*NXSuperblock, error)`
  - `func ValidateNXSuperblock(sb *NXSuperblock) error`
  - `func FindLatestSuperblock(device BlockDevice) (*NXSuperblock, PAddr, error)`
  - `func WriteSuperblock(device BlockDevice, sb *NXSuperblock) error`

- **checkpoint.go**
  - `func ReadCheckpointMap(device BlockDevice, addr PAddr) (*CheckpointMapPhys, error)`
  - `func FindLatestCheckpoint(device BlockDevice, sb *NXSuperblock) (*CheckpointMapPhys, error)`
  - `func ReadCheckpointArea(device BlockDevice, sb *NXSuperblock) ([]CheckpointMapPhys, error)`
  - `func WriteCheckpoint(device BlockDevice, sb *NXSuperblock, objects []EphemeralObject) error`

- **omap.go**
  - `func ReadOMap(device BlockDevice, addr PAddr) (*OMapPhys, error)`
  - `func LookupOMapRecord(device BlockDevice, omap *OMapPhys, oid OID, xid XID) (*OMapVal, error)`
  - `func InsertOMapRecord(device BlockDevice, omap *OMapPhys, oid OID, xid XID, paddr PAddr) error`
  - `func DeleteOMapRecord(device BlockDevice, omap *OMapPhys, oid OID, xid XID) error`

- **btree.go**
  - `func SearchBTree(device BlockDevice, rootNodeAddr PAddr, key []byte, compare KeyCompareFunc) ([]byte, error)`
  - `func InsertBTree(device BlockDevice, rootNodeAddr *PAddr, key, value []byte, compare KeyCompareFunc) error`
  - `func DeleteBTree(device BlockDevice, rootNodeAddr PAddr, key []byte, compare KeyCompareFunc) error`
  - `func IterateBTree(device BlockDevice, rootNodeAddr PAddr, callback IterateCallback) error`

- **btnode.go**
  - `func ReadBTreeNode(device BlockDevice, addr PAddr) (*BTNodePhys, error)`
  - `func WriteBTreeNode(device BlockDevice, node *BTNodePhys, addr PAddr) error`
  - `func GetKeyValue(node *BTNodePhys, index int) ([]byte, []byte, error)`
  - `func InsertKeyValue(node *BTNodePhys, key, value []byte) error`
  - `func DeleteKeyValue(node *BTNodePhys, index int) error`
  - `func SplitNode(node *BTNodePhys) (*BTNodePhys, *BTNodePhys, []byte, error)`

- **spaceman.go**
  - `func ReadSpaceManager(device BlockDevice, addr PAddr) (*SpacemanPhys, error)`
  - `func GetFreeSpace(device BlockDevice, spaceman *SpacemanPhys) (uint64, error)`
  - `func UpdateSpaceManagerStats(device BlockDevice, spaceman *SpacemanPhys) error`
  - `func InitializeSpaceManager(device BlockDevice, blockCount uint64) (*SpacemanPhys, error)`

- **allocation.go**
  - `func AllocateBlocks(device BlockDevice, spaceman *SpacemanPhys, count uint64) (PAddr, error)`
  - `func FreeBlocks(device BlockDevice, spaceman *SpacemanPhys, start PAddr, count uint64) error`
  - `func AllocateObject(device BlockDevice, spaceman *SpacemanPhys, size uint32) (PAddr, error)`
  - `func MarkBlocksReserved(device BlockDevice, spaceman *SpacemanPhys, start PAddr, count uint64) error`

- **reaper.go**
  - `func ReadReaper(device BlockDevice, addr PAddr) (*ReaperPhys, error)`
  - `func AddToReaperQueue(device BlockDevice, reaper *ReaperPhys, oid OID, fsoid OID) error`
  - `func ProcessReaperQueue(device BlockDevice, reaper *ReaperPhys, maxItems int) error`
  - `func IsReapInProgress(reaper *ReaperPhys) bool`

- **encryption.go**
  - `func ReadContainerKeybag(device BlockDevice, sb *NXSuperblock) (*KBLocker, error)`
  - `func UnwrapContainerKeybag(keybag *KBLocker, uuid UUID) error`
  - `func GetVolumeKey(containerKeybag *KBLocker, volUUID UUID) ([]byte, error)`
  - `func UpdateContainerKeybag(device BlockDevice, sb *NXSuperblock, keybag *KBLocker) error`

## pkg/fs/
- **volume.go**
  - `func ReadVolumeSuperblock(device BlockDevice, addr PAddr) (*APFSSuperblock, error)`
  - `func ValidateVolumeSuperblock(sb *APFSSuperblock) error`
  - `func GetVolumeInfo(sb *APFSSuperblock) (*VolumeInfo, error)`
  - `func UpdateVolumeSuperblock(device BlockDevice, sb *APFSSuperblock) error`

- **tree.go**
  - `func ReadFSTree(device BlockDevice, omap *OMapPhys, oid OID) (*BTNodePhys, error)`
  - `func SearchFSTree(device BlockDevice, omap *OMapPhys, rootOID OID, key []byte) ([]byte, error)`
  - `func InsertFSRecord(device BlockDevice, omap *OMapPhys, rootOID OID, key, value []byte) error`
  - `func DeleteFSRecord(device BlockDevice, omap *OMapPhys, rootOID OID, key []byte) error`

- **inode.go**
  - `func ReadInode(device BlockDevice, fs *FileSystem, inodeNum OID) (*JInodeVal, error)`
  - `func CreateInode(fs *FileSystem, parentID OID, name string, mode uint16) (OID, error)`
  - `func UpdateInode(fs *FileSystem, inodeNum OID, inode *JInodeVal) error`
  - `func DeleteInode(fs *FileSystem, inodeNum OID) error`
  - `func GetInodeExtendedField(inode *JInodeVal, fieldType uint8) ([]byte, error)`

- **dentry.go**
  - `func ReadDirectoryEntry(fs *FileSystem, dirID OID, name string) (*JDrecVal, error)`
  - `func AddDirectoryEntry(fs *FileSystem, dirID OID, name string, fileID OID) error`
  - `func RemoveDirectoryEntry(fs *FileSystem, dirID OID, name string) error`
  - `func ListDirectoryEntries(fs *FileSystem, dirID OID) ([]DirectoryEntry, error)`

- **extattr.go**
  - `func GetExtendedAttribute(fs *FileSystem, fileID OID, name string) ([]byte, error)`
  - `func SetExtendedAttribute(fs *FileSystem, fileID OID, name string, data []byte) error`
  - `func RemoveExtendedAttribute(fs *FileSystem, fileID OID, name string) error`
  - `func ListExtendedAttributes(fs *FileSystem, fileID OID) ([]string, error)`

- **datastream.go**
  - `func ReadDataStream(fs *FileSystem, oid OID) (*JDstream, error)`
  - `func WriteDataStream(fs *FileSystem, oid OID, stream *JDstream) error`
  - `func AllocateDataStream(fs *FileSystem, size uint64) (OID, error)`
  - `func ResizeDataStream(fs *FileSystem, oid OID, newSize uint64) error`

- **extent.go**
  - `func ReadFileExtent(fs *FileSystem, fileID OID, logicalAddr uint64) (*JFileExtentVal, error)`
  - `func AllocateFileExtent(fs *FileSystem, fileID OID, logicalAddr uint64, size uint64) error`
  - `func FreeFileExtent(fs *FileSystem, fileID OID, logicalAddr uint64) error`
  - `func ReadDataFromExtent(fs *FileSystem, extent *JFileExtentVal, offset int64, size int) ([]byte, error)`
  - `func GetFileExtents(fs *FileSystem, fileID OID) ([]FileExtent, error)`

- **sibling.go**
  - `func CreateHardLink(fs *FileSystem, targetID OID, dirID OID, name string) error`
  - `func ReadSiblingLink(fs *FileSystem, siblingID uint64) (*JSiblingVal, error)`
  - `func UpdateSiblingLink(fs *FileSystem, siblingID uint64, val *JSiblingVal) error`
  - `func GetSiblingMap(fs *FileSystem, siblingID uint64) (*JSiblingMapVal, error)`
  - `func GetPrimaryLink(fs *FileSystem, inodeID OID) (OID, string, error)`

- **crypto.go**
  - `func ReadVolumeKeybag(device BlockDevice, sb *APFSSuperblock) (*KBLocker, error)`
  - `func UnwrapVolumeKeybag(keybag *KBLocker, passwd string) error`
  - `func GetFileKey(device BlockDevice, sb *APFSSuperblock, fileID OID) ([]byte, error)`
  - `func EncryptFileData(data []byte, key []byte, tweak uint64) ([]byte, error)`
  - `func DecryptFileData(data []byte, key []byte, tweak uint64) ([]byte, error)`

## pkg/io/
- **blockdevice.go**
  - `func NewBlockDevice(path string) (BlockDevice, error)`
  - `func (d *DeviceImpl) ReadBlock(addr PAddr) ([]byte, error)`
  - `func (d *DeviceImpl) WriteBlock(addr PAddr, data []byte) error`
  - `func (d *DeviceImpl) GetBlockSize() uint32`
  - `func (d *DeviceImpl) GetBlockCount() uint64`
  - `func (d *DeviceImpl) Close() error`

- **cache.go**
  - `func NewBlockCache(device BlockDevice, cacheSize int) *BlockCache`
  - `func (c *BlockCache) ReadBlock(addr PAddr) ([]byte, error)`
  - `func (c *BlockCache) WriteBlock(addr PAddr, data []byte) error`
  - `func (c *BlockCache) Flush() error`
  - `func (c *BlockCache) Invalidate(addr PAddr) error`

- **transaction.go**
  - `func BeginTransaction(device BlockDevice, sb *NXSuperblock) (*Transaction, error)`
  - `func (t *Transaction) CreateObject(objType, objSubtype uint32, size uint32) (OID, []byte, error)`
  - `func (t *Transaction) UpdateObject(oid OID, data []byte) error`
  - `func (t *Transaction) DeleteObject(oid OID) error`
  - `func (t *Transaction) Commit() error`
  - `func (t *Transaction) Abort() error`

## pkg/crypto/
- **keybag.go**
  - `func ReadKeybag(data []byte) (*KBLocker, error)`
  - `func CreateKeybag() *KBLocker`
  - `func (k *KBLocker) AddEntry(uuid UUID, tag uint16, keyData []byte) error`
  - `func (k *KBLocker) FindEntry(uuid UUID, tag uint16) (*KeybagEntry, error)`
  - `func (k *KBLocker) Serialize() ([]byte, error)`

- **key.go**
  - `func DeriveKeyFromPassword(password string, salt []byte) ([]byte, error)`
  - `func GenerateVolumeEncryptionKey() ([]byte, error)`
  - `func WrapKey(key []byte, wrappingKey []byte) ([]byte, error)`
  - `func UnwrapKey(wrappedKey []byte, wrappingKey []byte) ([]byte, error)`
  - `func GenerateRecoveryKey() (string, []byte, error)`

- **aes.go**
  - `func NewAESXTSContext(key []byte) (*AESXTSContext, error)`
  - `func (ctx *AESXTSContext) Encrypt(plaintext []byte, tweak uint64) ([]byte, error)`
  - `func (ctx *AESXTSContext) Decrypt(ciphertext []byte, tweak uint64) ([]byte, error)`
  - `func GenerateTweak(blockNum uint64, fileID uint64) uint64`
  - `func ValidateEncryptionParameters(key []byte, data []byte) error`

- **wrappers.go**
  - `func UnwrapMetaCryptoState(state *WrappedMetaCryptoState, key []byte) error`
  - `func UnwrapCryptoState(state *WrappedCryptoState, key []byte) ([]byte, error)`
  - `func WrapCryptoState(key []byte, wrappingKey []byte, class uint32) (*WrappedCryptoState, error)`
  - `func GetKeyClass(state *WrappedCryptoState) uint32`
  - `func IsKeyRollingNeeded(state *WrappedCryptoState) bool`

## pkg/snapshot/
- **metadata.go**
  - `func ReadSnapshotMetadata(fs *FileSystem, xid XID) (*JSnapMetadataVal, error)`
  - `func GetSnapshotByName(fs *FileSystem, name string) (*JSnapMetadataVal, XID, error)`
  - `func ListSnapshots(fs *FileSystem) ([]SnapshotInfo, error)`
  - `func GetSnapshotExtendedMetadata(fs *FileSystem, xid XID) (*SnapMetaExt, error)`
  - `func UpdateSnapshotMetadata(fs *FileSystem, xid XID, metadata *JSnapMetadataVal) error`

- **operations.go**
  - `func CreateSnapshot(fs *FileSystem, name string) (XID, error)`
  - `func DeleteSnapshot(fs *FileSystem, xid XID) error`
  - `func MountSnapshot(fs *FileSystem, xid XID) (*FileSystem, error)`
  - `func RevertToSnapshot(fs *FileSystem, xid XID) error`
  - `func IsSnapshotMountable(fs *FileSystem, xid XID) (bool, error)`

## pkg/seal/
- **integrity.go**
  - `func ReadIntegrityMetadata(device BlockDevice, addr PAddr) (*IntegrityMetaPhys, error)`
  - `func ValidateIntegrityMetadata(meta *IntegrityMetaPhys) error`
  - `func CreateIntegrityMetadata(hashType HashType) (*IntegrityMetaPhys, error)`
  - `func UpdateIntegrityMetadata(device BlockDevice, meta *IntegrityMetaPhys, addr PAddr) error`
  - `func MarkSealBroken(device BlockDevice, meta *IntegrityMetaPhys, xid XID) error`

- **hash.go**
  - `func ComputeFileHash(fs *FileSystem, fileID OID, hashType HashType) ([]byte, error)`
  - `func ComputeNodeHash(node *BTNodePhys, hashType HashType) ([]byte, error)`
  - `func VerifyNodeHash(node *BTNodePhys, storedHash []byte, hashType HashType) (bool, error)`
  - `func GetHashSize(hashType HashType) uint8`
  - `func GetHashFunction(hashType HashType) HashFunction`

## pkg/checksum/
- **fletcher64.go**
  - `func Fletcher64(data []byte) uint64`
  - `func Fletcher64WithZeroedChecksum(data []byte, offset int) uint64`
  - `func VerifyFletcher64Checksum(data []byte) bool`
  - `func UpdateFletcher64Checksum(data []byte, offset int) []byte`
