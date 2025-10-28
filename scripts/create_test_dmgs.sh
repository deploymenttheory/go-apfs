#!/bin/bash

# Script to create comprehensive test APFS images for testing
# Creates raw APFS volumes that can be read directly without decompression

set -e

TESTS_DIR="/Users/dafyddwatkins/GitHub/deploymenttheory/go-apfs/tests"
TEMP_DIR="/tmp/apfs_test_build_$$"

trap "cleanup_and_exit" EXIT INT TERM

cleanup_and_exit() {
	echo "Cleaning up temporary files..."
	
	# Unmount any remaining volumes
	for vol in /Volumes/TestAPFS* /Volumes/EmptyAPFS /Volumes/BasicAPFS /Volumes/PopulatedAPFS /Volumes/FullAPFS; do
		if [ -d "$vol" ]; then
			hdiutil detach "$vol" 2>/dev/null || true
			sleep 1
		fi
	done
	
	# Remove temp directory
	rm -rf "$TEMP_DIR"
}

echo "=== Creating APFS Test Images ==="
echo "Test directory: $TESTS_DIR"
echo "Temp directory: $TEMP_DIR"

# Ensure clean test directory
echo "Cleaning existing test files..."
mkdir -p "$TESTS_DIR"
rm -f "$TESTS_DIR"/*.dmg 2>/dev/null || true

mkdir -p "$TEMP_DIR"

# Function to create and populate an APFS volume
create_apfs_volume() {
	local size=$1
	local name=$2
	local output_file=$3
	local populate_func=$4
	
	echo ""
	echo "=========================================="
	echo "Creating: $name ($size)"
	echo "=========================================="
	
	# Create a raw disk image file
	local disk_image="$TEMP_DIR/${name}.img"
	
	echo "1. Creating raw disk image ($size)..."
	hdiutil create -size "$size" -type SPARSE -fs APFS -volname "$name" -o "$disk_image"
	
	# The output is actually a .sparseimage file
	local sparse_image="${disk_image}.sparseimage"
	if [ ! -f "$sparse_image" ]; then
		echo "ERROR: Failed to create disk image"
		return 1
	fi
	
	echo "2. Mounting volume..."
	hdiutil attach "$sparse_image"
	sleep 3
	
	# Find the mounted volume
	local mount_point="/Volumes/$name"
	if [ ! -d "$mount_point" ]; then
		echo "ERROR: Volume $name not found at $mount_point after mount"
		hdiutil detach "$sparse_image" 2>/dev/null || true
		return 1
	fi
	
	echo "3. Volume mounted at: $mount_point"
	
	# Populate the volume if function provided
	if [ -n "$populate_func" ] && type "$populate_func" >/dev/null 2>&1; then
		echo "4. Populating volume..."
		"$populate_func" "$mount_point"
	fi
	
	# Sync and unmount
	echo "5. Syncing filesystem..."
	sync
	sleep 2
	
	echo "6. Unmounting volume..."
	hdiutil detach "$mount_point"
	sleep 2
	
	# Extract raw APFS to file (skip partition table, get just the container)
	echo "7. Extracting raw APFS container..."
	
	# The APFS partition with GPT typically starts at block 40 (20480 bytes / 512)
	# We need to find where APFS actually starts by looking for the magic
	# For UDRO format with GPT, APFS is usually at offset 1048576 (256 * 4096)
	# But we'll extract the whole image and let offset detection handle it
	
	# For now, just copy the sparse image directly as the output
	# Our offset detection will find the APFS container
	cp "$sparse_image" "$output_file"
	
	# Clean up sparse image
	rm -f "$sparse_image"
	
	echo "✓ Created: $output_file"
}

# Populate functions for different scenarios

populate_empty() {
	# Empty volume - no additional files needed
	echo "   (empty volume)"
}

populate_basic() {
	local mount_point=$1
	echo "Basic test file content" > "$mount_point/readme.txt"
	echo "Another test file" > "$mount_point/test.dat"
	mkdir -p "$mount_point/subfolder"
	echo "File in subdirectory" > "$mount_point/subfolder/nested.txt"
}

populate_populated() {
	local mount_point=$1
	mkdir -p "$mount_point/Documents"
	mkdir -p "$mount_point/Data"
	
	for i in {1..10}; do
		echo "Test file content $i" > "$mount_point/Documents/file_$i.txt"
	done
	
	for i in {1..5}; do
		echo "Data file $i with some content to make it realistic" > "$mount_point/Data/data_$i.dat"
	done
	
	ln -s "$mount_point/Documents/file_1.txt" "$mount_point/symlink_test.txt"
}

populate_full() {
	local mount_point=$1
	mkdir -p "$mount_point/bigfiles"
	
	# Create files to fill most of the space
	dd if=/dev/random of="$mount_point/bigfiles/large_1.bin" bs=1024 count=800 2>/dev/null
	dd if=/dev/random of="$mount_point/bigfiles/large_2.bin" bs=1024 count=800 2>/dev/null
	
	echo "File in full volume" > "$mount_point/test.txt"
}

# Create test volumes

create_apfs_volume "5m" "EmptyAPFS" "$TESTS_DIR/empty_apfs.dmg" "populate_empty"

create_apfs_volume "8m" "BasicAPFS" "$TESTS_DIR/basic_apfs.dmg" "populate_basic"

create_apfs_volume "10m" "PopulatedAPFS" "$TESTS_DIR/populated_apfs.dmg" "populate_populated"

create_apfs_volume "6m" "FullAPFS" "$TESTS_DIR/full_apfs.dmg" "populate_full"

# Summary
echo ""
echo "=========================================="
echo "✓ Test image creation complete!"
echo "=========================================="
echo ""
echo "Created test images:"
ls -lh "$TESTS_DIR"/*.dmg 2>/dev/null || echo "ERROR: No DMG files found!"

exit 0