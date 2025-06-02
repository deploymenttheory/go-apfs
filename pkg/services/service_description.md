# APFS Services Description

This document outlines the viable services for the go-apfs library, taking into account the reality of T2/hardware encryption constraints and practical use cases.

## Core Read-Only Operations

### ContainerService
Provides container discovery, superblock reading, and basic container metadata analysis. Works on any accessible APFS partition to enumerate volumes and extract container-level information like space management and checkpoint data.

### VolumeService  
Handles volume enumeration, metadata reading, and basic volume information extraction. Can read volume superblocks, feature flags, and structural metadata even when file content is encrypted.

### FilesystemService
Provides file and directory listing, inode reading, and basic filesystem navigation. Limited to accessible (unencrypted) volumes but provides full filesystem tree traversal and metadata access. Supports operations like `ls` with detailed metadata (file types, timestamps, permissions).

### ExtractionService
Handles file and directory extraction to disk for accessible volumes. Supports selective extraction, preserves metadata and extended attributes, and can extract entire volume contents. Enables operations like `cp` for copying files/directories from APFS volumes to local filesystem.

## Structure Analysis

### AnalysisService
Performs deep APFS structure analysis including B-tree analysis, object statistics, and filesystem health assessment. Works on encrypted volumes for structural analysis without requiring content access.

### ValidationService
Provides integrity checking, corruption detection, and APFS structure validation. Verifies checksums, object headers, and structural consistency regardless of encryption status.

### EncryptionService
Analyzes encryption state, keybag structures, and protection classes without performing decryption. Useful for understanding encryption configuration and compliance verification.

### SnapshotService
Handles snapshot enumeration, metadata extraction, and snapshot comparison. Can analyze snapshot structures and provide snapshot timeline information.

## DMG/Image Handling

### DMGService
Provides DMG file parsing, APFS volume extraction from DMG containers, and DMG metadata analysis. Often works with unencrypted DMG files making them prime targets for analysis. Supports iOS firmware (IPSW) DMG extraction and analysis.

### ImageService
Handles disk image validation, format detection, and image metadata extraction. Supports various image formats and provides image integrity verification.

### FirmwareService
Specialized service for iOS/macOS firmware analysis including IPSW processing, firmware DMG extraction, and system volume analysis. Enables forensic analysis of iOS firmware images and system partitions.

## Recovery Operations

### RecoveryService
Performs deleted file recovery and undelete operations on unencrypted volumes. Limited effectiveness on encrypted volumes but valuable for accessible data recovery scenarios.

### ScanService
Provides raw disk scanning for APFS structures, orphaned data recovery, and low-level structure discovery. Useful for damaged filesystem recovery and forensic analysis.

## File Operations

### FileContentService
Provides file content reading capabilities similar to `cat` operations but with enhanced features like partial reads, streaming for large files, and content type detection.

### DirectoryService
Advanced directory operations beyond basic listing including recursive traversal, directory statistics, and hierarchical analysis. Supports filtering, sorting, and metadata-rich directory exploration.

### MetadataService
Comprehensive file metadata extraction including timestamps, permissions, extended attributes, resource forks, and file flags. Provides detailed file information for forensic analysis.

## Utility Services

### ReportService
Generates comprehensive analysis reports in various formats (JSON, XML, HTML). Aggregates data from multiple services to provide detailed filesystem analysis reports.

### ExportService
Handles metadata export to various formats, data serialization, and integration with external tools. Supports structured data export for further analysis or compliance reporting.

### ComparisonService
Enables comparison between volumes, snapshots, or filesystem states. Useful for change detection, forensic timeline analysis, and system state comparison.

## Use Cases Inspired by Real-World Tools

### iOS Forensics
- Extract and analyze iOS firmware DMG files from IPSW packages
- Navigate iOS system volumes and extract system files
- Analyze iOS application data and system configurations

### macOS Analysis
- Examine macOS system volumes and user data
- Extract system files, preferences, and application data
- Analyze system integrity and configuration

### Digital Forensics
- Timeline reconstruction from filesystem metadata
- Deleted file recovery from unencrypted volumes
- Evidence extraction with preserved metadata and chain of custody

### System Administration
- Backup verification and integrity checking
- System configuration analysis and documentation
- Capacity planning and storage utilization analysis
