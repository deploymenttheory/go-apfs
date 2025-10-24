# Test Fixtures

This directory contains APFS container test data used by the go-apfs test harness.

## Test Containers

### test_container.img
- **Size**: 100MB (104,857,600 bytes)
- **Format**: APFS container with GPT partition table
- **Usage**: Legacy test container used by `container_reader_test.go`
- **Creation**: Created using macOS `diskutil` and `hdiutil` tools

### basic_test.img
- **Size**: ~783KB (783,136 bytes)
- **Format**: Real APFS container (basic)
- **Usage**: Used by `TestNewVolumeService` - basic volume service functionality
- **Content**: Single text file "basic.txt"
- **Creation**: `hdiutil create -fs APFS -volname BasicTest`

### volume_test.img
- **Size**: ~873KB (873,248 bytes)
- **Format**: Real APFS container (volume data)
- **Usage**: Used by `TestGetSpaceUsageStats` - volume space management tests
- **Content**: Directory structure with binary data file
- **Creation**: `hdiutil create -fs APFS -volname VolumeTest` + test data

### metadata_test.img
- **Size**: ~848KB (848,672 bytes)
- **Format**: Real APFS container (metadata)
- **Usage**: Used by `TestGetVolumeMetadata` - volume metadata analysis tests
- **Content**: 20 files, directories, and symlinks for metadata testing
- **Creation**: `hdiutil create -fs APFS -volname MetadataTest` + complex structure

## Test Data Usage

### B-tree Object Resolver Testing
All containers are used to test the B-tree object resolver functionality:

```go
func TestBTreeResolverWithMultipleContainers(t *testing.T) {
    testContainers := []struct {
        name string
        path string
        needsExtraction bool
    }{
        {"Original Test Container", "tests/test_container.img", true},
        {"Real APFS Container (minimal)", "tests/apfs_padded_container.img", false},
        {"Real APFS Container (with data)", "tests/updated_apfs_padded.img", false},
    }
    // ... test logic
}
```

### Container Properties
All containers demonstrate:
- ✅ Valid APFS container superblock parsing
- ✅ B-tree node structure reading  
- ✅ Object map OID resolution
- ✅ Empty B-tree detection and fallback logic

## File Maintenance

### Required Files Only
This directory contains only files required by the test harness. Intermediate files (DMG, sparse images, raw extractions) have been removed to keep the repository clean.

### Regeneration
To recreate test containers if needed:

```bash
# Create new minimal APFS container
hdiutil create -size 100m -fs APFS -volname TestVolume test_new.sparseimage
hdiutil attach test_new.sparseimage
# Add test files as needed
hdiutil detach /dev/disk#

# Extract and format for tests
# (see test creation scripts in project history)
```
