#!/bin/bash

# Script to generate valid APFS test containers with real filesystem data
# Uses hdiutil with -fs apfs to create native APFS volumes
# Reference: https://gist.github.com/darwin/d5df8fbb1c2710a29d7d0908f941b329

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
TESTS_DIR="$PROJECT_ROOT/tests"

echo "Generating APFS test containers in $TESTS_DIR..."

# Create temp directory for building containers  
TEMP_DIR=$(mktemp -d)
cleanup() {
    # Unmount any attached images first
    hdiutil detach /tmp/test_mount 2>/dev/null || true
    rm -rf "$TEMP_DIR"
}
trap cleanup EXIT

# Function to create APFS disk image with content
create_apfs_container() {
    local size_mb=$1
    local name=$2
    local output_file=$3
    
    echo "Creating $name (${size_mb}MB APFS)..."
    
    local temp_img="/tmp/${name}_temp.dmg"
    
    # Create APFS image directly
    hdiutil create -size ${size_mb}m -fs APFS -volname "Temp" "$temp_img" > /dev/null 2>&1 || {
        echo "Error: Failed to create APFS image"
        return 1
    }
    
    if [ ! -f "$temp_img" ]; then
        echo "Error: APFS image file not created at $temp_img"
        return 1
    fi
    
    # Attach the image
    local device=$(hdiutil attach "$temp_img" 2>&1 | grep -o '/dev/disk[0-9]*s[0-9]*' | head -1)
    if [ -z "$device" ]; then
        echo "Error: Could not attach image"
        rm -f "$temp_img"
        return 1
    fi
    
    echo "  Attached to $device"
    
    # Get mount point - for APFS, we need to wait a moment and check
    sleep 1
    local mount_point=$(diskutil info "$device" 2>/dev/null | grep "Mount Point:" | awk -F': ' '{print $2}')
    
    if [ -z "$mount_point" ] || [ "$mount_point" = "Not mounted" ]; then
        echo "Error: APFS volume not mounting automatically. Trying via hdiutil..."
        # Try to get the mount point differently
        mount_point=$(hdiutil attach "$temp_img" -plist 2>/dev/null | grep -A1 "mount-point" | grep "string" | sed 's/.*<string>//;s/<\/string>.*//' | head -1)
        
        if [ -z "$mount_point" ]; then
            echo "Error: Could not determine mount point"
            hdiutil detach "$device" 2>/dev/null
            rm -f "$temp_img"
            return 1
        fi
    fi
    
    echo "  Mounted at $mount_point"
    
    # Add content
    case "$name" in
        basic_test)
            echo "This is a test file" > "$mount_point/test1.txt"
            mkdir -p "$mount_point/Documents"
            echo "Document content" > "$mount_point/Documents/document.txt"
            mkdir -p "$mount_point/Library/Caches"
            echo "Cache data" > "$mount_point/Library/Caches/cache.dat"
            ;;
        metadata_test)
            echo "Old file" > "$mount_point/old_file.txt"
            touch -t 202001011200 "$mount_point/old_file.txt"
            echo "Recent file" > "$mount_point/recent_file.txt"
            mkdir -p "$mount_point/Subdirectory"
            echo "Nested content" > "$mount_point/Subdirectory/nested.txt"
            dd if=/dev/zero of="$mount_point/large_file.bin" bs=1M count=2 2>/dev/null
            ;;
        volume_test)
            mkdir -p "$mount_point"/{Documents,Downloads,Applications,System/Library}
            for i in {1..5}; do
                echo "Document $i" > "$mount_point/Documents/doc_$i.txt"
            done
            for i in {1..3}; do
                dd if=/dev/zero of="$mount_point/Downloads/file_$i.bin" bs=512k count=1 2>/dev/null
            done
            echo "System lib" > "$mount_point/System/Library/system.lib"
            mkdir -p "$mount_point/Applications/Test.app/Contents/MacOS"
            echo "Test" > "$mount_point/Applications/Test.app/Contents/MacOS/Test"
            ;;
        test_container_base)
            echo "Test content" > "$mount_point/test.txt"
            mkdir -p "$mount_point/TestFolder"
            echo "Folder content" > "$mount_point/TestFolder/content.txt"
            ;;
    esac
    
    # Sync and detach
    echo "  Syncing filesystem..."
    sync
    sleep 1
    
    echo "  Detaching image..."
    umount "$mount_point" 2>/dev/null || true
    sleep 1
    hdiutil detach "$device" 2>/dev/null || hdiutil detach "$device" -force 2>/dev/null
    sleep 2
    
    # Verify source file exists
    if [ ! -f "$temp_img" ]; then
        echo "Error: Temporary image not found at $temp_img"
        return 1
    fi
    
    # Copy to output
    echo "  Copying to $output_file..."
    if cp "$temp_img" "$output_file"; then
        echo "  ✓ Successfully copied to $output_file"
    else
        echo "Error: Failed to copy $temp_img to $output_file"
        rm -f "$temp_img"
        return 1
    fi
    
    # Cleanup
    rm -f "$temp_img"
    
    # Verify output
    if [ -f "$output_file" ]; then
        local size=$(du -h "$output_file" | awk '{print $1}')
        echo "  ✓ Created $output_file ($size)"
        return 0
    else
        echo "Error: Output file not created: $output_file"
        return 1
    fi
}

# Create containers
create_apfs_container 20 basic_test "$TESTS_DIR/basic_test.img"
create_apfs_container 30 metadata_test "$TESTS_DIR/metadata_test.img"
create_apfs_container 50 volume_test "$TESTS_DIR/volume_test.img"

# Create test_container with GPT wrapper
echo "Creating test_container.img with GPT wrapper..."
create_apfs_container 25 test_container_base "$TEMP_DIR/apfs_base.img" && \
{
    {
        dd if=/dev/zero bs=512 count=40 2>/dev/null
        cat "$TEMP_DIR/apfs_base.img"
    } > "$TESTS_DIR/test_container.img"
    echo "  ✓ Created test_container.img"
}

echo ""
echo "Test containers ready:"
ls -lh "$TESTS_DIR"/*.img 2>/dev/null || echo "Note: Check $TESTS_DIR"
