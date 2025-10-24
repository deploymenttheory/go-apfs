package services

// FileSystemService provides high-level filesystem operations
type FileSystemService interface {
	GetInodeByPath(path string) (*FileNode, error)
	ListDirectoryContents(inodeID uint64) ([]*FileNode, error)
	GetFileExtents(inodeID uint64) ([]ExtentMapping, error)
	FindFilesByName(pattern string, maxResults int) ([]*FileNode, error)
	GetFileMetadata(inodeID uint64) (*FileNode, error)
	GetParentDirectory(inodeID uint64) (*FileNode, error)
	IsPathAccessible(path string) (bool, error)
}

// ObjectLocatorService provides object discovery and analysis
type ObjectLocatorService interface {
	FindObjectByID(oid uint64) (interface{}, error)
	ResolveObjectPath(oid uint64) ([]uint64, error)
	IsObjectValid(oid uint64) (bool, error)
	GetObjectDependencies(oid uint64) ([]uint64, error)
	AnalyzeObjectReferences() (*ReferenceGraph, error)
	FindOrphanedObjects() ([]uint64, error)
	GetObjectType(oid uint64) (string, error)
	GetObjectSize(oid uint64) (uint64, error)
}

// SnapshotService provides snapshot management and analysis
type SnapshotService interface {
	ListAllSnapshots() ([]*SnapshotInfo, error)
	GetSnapshotMetadata(xid uint64) (*SnapshotInfo, error)
	CompareSnapshots(xid1, xid2 uint64) (*DiffReport, error)
	GetChangedFiles(xid1, xid2 uint64) ([]FileChange, error)
	GetSnapshotSize(xid uint64) (uint64, error)
	FindSnapshotByName(name string) (*SnapshotInfo, error)
	GetSnapshotFileCount(xid uint64) (uint64, error)
}

// VolumeService provides volume-level operations
type VolumeService interface {
	GetVolumeMetadata() (*VolumeReport, error)
	GetSpaceUsageStats() (*SpaceStats, error)
	AnalyzeVolumeFragmentation() (map[string]interface{}, error)
	DetectCorruption() ([]VolumeCorruptionAnomaly, error)
	GenerateVolumeReport() (*VolumeReport, error)
	GetFileCount() (uint64, error)
	GetDirectoryCount() (uint64, error)
	GetSymlinkCount() (uint64, error)
}

// DataRecoveryService provides data recovery and rescue capabilities
type DataRecoveryService interface {
	FindUnlinkedInodes() ([]uint64, error)
	FindOrphanedExtents() ([]ExtentMapping, error)
	ScanForDeletedFiles(pattern string) (*RecoveryReport, error)
	EstimateRecoveryPotential() (*RecoveryReport, error)
	BuildRecoveryMap() (map[uint64]*RecoverableFile, error)
	GetRecoverableFileCount() (int, error)
	GetRecoverableDataSize() (uint64, error)
	FindFilesByInode(inode uint64) ([]*RecoverableFile, error)
}

// EncryptionService provides encryption verification and analysis
type EncryptionService interface {
	GetEncryptionStatus() (*EncryptionState, error)
	VerifyEncryptionConsistency() (bool, []string, error)
	AnalyzeKeyRolling() (map[string]interface{}, error)
	CheckProtectionClasses() (map[string]interface{}, error)
	ValidateEncryptionMetadata() (bool, []string, error)
	IsFileEncrypted(inode uint64) (bool, error)
	GetEncryptionKeys() (map[string]interface{}, error)
	VerifyFileEncryption(inode uint64) (bool, error)
}

// CacheService provides optional caching for performance
type CacheService interface {
	CacheContainerSuperblock() error
	CacheVolumeSuperblock() error
	CacheBTreeNodes(count int) error
	CacheObjectMap() error
	PrefetchFrequentObjects() error
	ClearCache() error
	GetCacheStats() (map[string]interface{}, error)
	IsCached(oid uint64) bool
}
