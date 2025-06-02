# APFS Services

This package provides high-level services for working with Apple File System (APFS) containers and volumes. The services layer abstracts the complexity of low-level APFS parsing and provides intuitive APIs for common operations.

## Overview

The services are designed around the principle of **read-only operations** that work within the constraints of T2/hardware encryption. They provide practical functionality for:

- **Container discovery and analysis**
- **Filesystem navigation and metadata extraction**
- **Structural analysis and health checking**
- **File extraction and forensic operations**

## Available Services

### ✅ ContainerService
**Status**: Implemented

Provides container-level operations including:
- Container discovery on accessible devices
- Container superblock reading and parsing
- Volume enumeration
- Space management information
- Checkpoint validation

```go
containerSvc, err := services.GetContainerService()
if err != nil {
    log.Fatal(err)
}

// Discover APFS containers
containers, err := containerSvc.DiscoverContainers(ctx)

// Open a specific container
info, err := containerSvc.OpenContainer(ctx, "/dev/disk1")

// List volumes in the container
volumes, err := containerSvc.ListVolumes(ctx, "/dev/disk1")
```

### ✅ FilesystemService  
**Status**: Implemented (with mock data)

Provides filesystem navigation and operations:
- Directory listing with detailed metadata
- File information retrieval
- Filesystem traversal and searching
- Access checking (encryption status)

```go
fsSvc, err := services.GetFilesystemService()
if err != nil {
    log.Fatal(err)
}

// List directory contents
files, err := fsSvc.ListDirectory(ctx, "/dev/disk1", volumeID, "/", false)

// Get detailed file information
fileInfo, err := fsSvc.GetFileInfo(ctx, "/dev/disk1", volumeID, "/Applications/Calculator.app")

// Check if file is accessible (not encrypted)
accessible, err := fsSvc.CheckAccess(ctx, "/dev/disk1", volumeID, "/Users/john/Documents")
```

### 🚧 VolumeService
**Status**: Interface defined, implementation pending

Will provide volume-level operations:
- Volume superblock reading
- Volume statistics and health
- Snapshot enumeration
- Volume integrity checking

### 🚧 AnalysisService
**Status**: Interface defined, implementation pending

Will provide deep structural analysis:
- B-tree health analysis
- Object statistics and integrity checking
- Performance metrics
- Comprehensive reporting

### 🚧 ExtractionService
**Status**: Interface defined, implementation pending

Will provide file extraction capabilities:
- Single file extraction
- Directory tree extraction
- Metadata preservation
- Integrity verification

## Quick Start

### Using the Default Service Factory

```go
package main

import (
    "context"
    "log"

    "github.com/deploymenttheory/go-apfs/pkg/services"
)

func main() {
    ctx := context.Background()
    
    // Initialize services
    if err := services.InitializeServices(); err != nil {
        log.Fatal("Failed to initialize services:", err)
    }
    defer services.ShutdownServices()

    // Get container service
    containerSvc, err := services.GetContainerService()
    if err != nil {
        log.Fatal("Failed to get container service:", err)
    }

    // Discover containers
    containers, err := containerSvc.DiscoverContainers(ctx)
    if err != nil {
        log.Fatal("Failed to discover containers:", err)
    }

    for _, container := range containers {
        log.Printf("Found container: %s (%d volumes)", 
                   container.DevicePath, container.VolumeCount)
    }
}
```

### Using a Custom Service Factory

```go
factory := services.NewServiceFactory()
defer factory.Shutdown()

// Initialize with custom configuration
if err := factory.Initialize(); err != nil {
    log.Fatal(err)
}

// Get services
containerSvc, _ := factory.ContainerService()
filesystemSvc, _ := factory.FilesystemService()

// Use services...
```

## Service Architecture

The services follow a layered architecture:

```
┌─────────────────────────────────────┐
│           Service Layer             │
├─────────────────────────────────────┤
│  Container │ Filesystem │ Analysis  │
│  Service   │ Service    │ Service   │
├─────────────────────────────────────┤
│         Parser Layer               │
├─────────────────────────────────────┤
│        Internal Types              │
└─────────────────────────────────────┘
```

### Design Principles

1. **Read-Only Operations**: All services are designed for read-only access
2. **Context-Aware**: All operations support cancellation via context
3. **Error-First**: Clear error handling and reporting
4. **Resource Management**: Proper cleanup and resource management
5. **Thread-Safe**: Services can be used concurrently

## Error Handling

Services use structured error handling:

```go
if err != nil {
    switch err {
    case services.ErrServiceNotImplemented:
        // Service not yet implemented
    case services.ErrServiceNotAvailable:
        // Service not available in current configuration
    default:
        // Other errors
    }
}
```

## Configuration

### Container Discovery

The container service looks for APFS containers in:
- `/dev/disk*` - macOS disk devices
- `/Volumes/*` - Mounted volumes  
- `*.dmg` - DMG files in current directory
- `*.img` - IMG files in current directory

### Limitations

Current implementation limitations:

1. **Encryption**: Encrypted content cannot be accessed without keys
2. **T2 Security**: Hardware-encrypted volumes have limited access
3. **Mock Data**: Some services return mock data pending full implementation
4. **Platform**: Primarily designed for macOS, limited cross-platform support

## Testing

Run the test suite:

```bash
cd pkg/services
go test -v
```

Run benchmarks:

```bash
go test -bench=. -benchmem
```

## Use Cases

### Digital Forensics
```go
// Analyze container structure
containers, _ := containerSvc.DiscoverContainers(ctx)
for _, container := range containers {
    info, _ := containerSvc.OpenContainer(ctx, container.DevicePath)
    // Analyze metadata, timestamps, structure
}
```

### System Administration  
```go
// Check filesystem health
volumes, _ := containerSvc.ListVolumes(ctx, devicePath)
for _, volume := range volumes {
    // Check space usage, integrity
}
```

### iOS/macOS Analysis
```go
// Extract system files from DMG
files, _ := filesystemSvc.ListDirectory(ctx, "firmware.dmg", volumeID, "/System", true)
// Process system files
```

## Future Enhancements

Planned enhancements include:

1. **Real B-tree Traversal**: Replace mock data with actual filesystem parsing
2. **Volume Operations**: Complete volume service implementation  
3. **Analysis Tools**: Deep structural analysis and health checking
4. **Extraction Capabilities**: File and directory extraction with metadata
5. **DMG Support**: Enhanced DMG/IPSW handling
6. **Reporting**: Comprehensive analysis reporting in multiple formats

## Contributing

When contributing to services:

1. Follow the established interface patterns
2. Ensure thread safety
3. Add comprehensive tests
4. Document new functionality
5. Handle errors appropriately
6. Support context cancellation

## License

This package is part of the go-apfs project and follows the same licensing terms. 