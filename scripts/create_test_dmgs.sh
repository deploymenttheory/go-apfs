#!/bin/bash

# Script to create comprehensive test APFS images for testing
# Creates APFS volumes with populated object maps using UDRW (read/write) format for raw data access

set -e

TESTS_DIR="/Users/dafyddwatkins/GitHub/deploymenttheory/go-apfs/tests"
TEMP_DIR="/tmp/apfs_test_build_$$"

trap "cleanup_and_exit" EXIT INT TERM

cleanup_and_exit() {
	echo "Cleaning up temporary files..."
	
	# Unmount any remaining volumes
	for vol in /Volumes/TestAPFS* /Volumes/EmptyAPFS /Volumes/BasicAPFS /Volumes/PopulatedAPFS /Volumes/FullAPFS /Volumes/BTree100MB; do
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
	
	# Create a sparse disk image
	local disk_image="$TEMP_DIR/${name}.img"
	
	echo "1. Creating sparse disk image ($size)..."
	hdiutil create -size "$size" -type SPARSE -fs APFS -volname "$name" -o "$disk_image"
	
	# The output is a .sparseimage file
	local sparse_image="${disk_image}.sparseimage"
	if [ ! -f "$sparse_image" ]; then
		echo "ERROR: Failed to create disk image"
		return 1
	fi
	
	echo "2. Mounting volume..."
	hdiutil attach "$sparse_image" >/dev/null 2>&1
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
	hdiutil detach "$mount_point" >/dev/null 2>&1
	sleep 2
	
	# Convert sparse image to UDRW (read-write) format
	# This preserves raw APFS structure with GPT and populated object maps
	echo "7. Converting to UDRW format..."
	local temp_udrw="${TEMP_DIR}/temp_udrw_$$"
	
	if hdiutil convert "$sparse_image" -format UDRW -o "$temp_udrw" >/dev/null 2>&1; then
		if [ -f "${temp_udrw}.dmg" ]; then
			cp "${temp_udrw}.dmg" "$output_file"
			rm -f "${temp_udrw}.dmg"
		else
			echo "ERROR: UDRW conversion failed to create output file"
			return 1
		fi
	else
		echo "ERROR: Failed to convert to UDRW format"
		return 1
	fi
	
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
	
	ln -s "$mount_point/Documents/file_1.txt" "$mount_point/symlink_test.txt" 2>/dev/null || true
}

populate_full() {
	local mount_point=$1
	mkdir -p "$mount_point/bigfiles"

	# Create files to fill most of the space
	dd if=/dev/random of="$mount_point/bigfiles/large_1.bin" bs=1024 count=800 2>/dev/null
	dd if=/dev/random of="$mount_point/bigfiles/large_2.bin" bs=1024 count=800 2>/dev/null

	echo "File in full volume" > "$mount_point/test.txt"
}

populate_btree_100mb() {
	local mount_point=$1

	echo "   Creating extensive directory structure to populate B-tree..."

	# Create deep directory structures to populate B-tree nodes
	for dir_num in {1..20}; do
		mkdir -p "$mount_point/dir_$dir_num/subdir_a/subdir_b/subdir_c"
		mkdir -p "$mount_point/dir_$dir_num/subdir_x/subdir_y/subdir_z"
	done

	# Create many small files across directories to populate file extent records
	for dir_num in {1..20}; do
		for file_num in {1..50}; do
			echo "Content for dir $dir_num file $file_num - $(date)" > "$mount_point/dir_$dir_num/file_${file_num}.txt"
		done
	done

	# Create files of varying sizes to populate extent B-trees
	mkdir -p "$mount_point/data"
	for i in {1..30}; do
		size=$((i * 100))
		dd if=/dev/random of="$mount_point/data/file_${i}.bin" bs=1024 count=$size 2>/dev/null
	done

	# Create many nested directories
	mkdir -p "$mount_point/nested"
	for i in {1..10}; do
		mkdir -p "$mount_point/nested/level_$i"
		for j in {1..10}; do
			echo "Nested content level $i item $j" > "$mount_point/nested/level_$i/item_$j.txt"
		done
	done

	# Create symlinks to add to the B-tree complexity
	mkdir -p "$mount_point/links"
	for i in {1..5}; do
		ln -s "$mount_point/dir_1/file_1.txt" "$mount_point/links/link_$i.txt" 2>/dev/null || true
	done

	echo "   Created $(find "$mount_point" -type f | wc -l) files"
	echo "   Created $(find "$mount_point" -type d | wc -l) directories"
}

# Create test volumes

create_apfs_volume "5m" "EmptyAPFS" "$TESTS_DIR/empty_apfs.dmg" "populate_empty"

create_apfs_volume "8m" "BasicAPFS" "$TESTS_DIR/basic_apfs.dmg" "populate_basic"

create_apfs_volume "10m" "PopulatedAPFS" "$TESTS_DIR/populated_apfs.dmg" "populate_populated"

create_apfs_volume "6m" "FullAPFS" "$TESTS_DIR/full_apfs.dmg" "populate_full"

create_apfs_volume "100m" "BTree100MB" "$TESTS_DIR/btree_100mb.dmg" "populate_btree_100mb"

# Summary
echo ""
echo "=========================================="
echo "✓ Test image creation complete!"
echo "=========================================="
echo ""
echo "Created test images:"
ls -lh "$TESTS_DIR"/*.dmg 2>/dev/null || echo "ERROR: No DMG files found!"

exit 0 
