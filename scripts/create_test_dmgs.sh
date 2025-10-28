#!/bin/bash

# Script to create comprehensive test DMG suite for APFS B-tree testing
# This creates DMGs with different filesystem states to test various B-tree scenarios

set -e

TESTS_DIR="/Users/dafyddwatkins/GitHub/deploymenttheory/go-apfs/tests"
DESKTOP_DIR="$HOME/Desktop"

echo "Creating comprehensive APFS test DMG suite..."

# Ensure tests directory exists
mkdir -p "$TESTS_DIR"

# 1. Empty APFS volume (minimal B-tree structures)
echo "1. Creating empty APFS volume..."
hdiutil create -size 5m -fs APFS -volname "EmptyAPFS" -type UDIF "$TESTS_DIR/empty_apfs.dmg"

# 2. APFS volume with a few files (basic B-tree population)
echo "2. Creating APFS volume with basic files..."
hdiutil create -size 8m -fs APFS -volname "BasicAPFS" -type UDIF "$TESTS_DIR/basic_apfs.dmg"
hdiutil attach "$TESTS_DIR/basic_apfs.dmg"

# Add basic files
echo "Basic test file content" > "/Volumes/BasicAPFS/readme.txt"
echo "Another test file" > "/Volumes/BasicAPFS/test.dat"
mkdir "/Volumes/BasicAPFS/subfolder"
echo "File in subdirectory" > "/Volumes/BasicAPFS/subfolder/nested.txt"

# Copy a small file from desktop if available
if [ -f "$DESKTOP_DIR/Apple-File-System-Reference.pdf" ]; then
    cp "$DESKTOP_DIR/Apple-File-System-Reference.pdf" "/Volumes/BasicAPFS/"
fi

hdiutil detach "/Volumes/BasicAPFS"

# 3. APFS volume with many files (populated B-tree structures)
echo "3. Creating APFS volume with many files..."
hdiutil create -size 10m -fs APFS -volname "PopulatedAPFS" -type UDIF "$TESTS_DIR/populated_apfs.dmg"
hdiutil attach "$TESTS_DIR/populated_apfs.dmg"

# Create directory structure
mkdir -p "/Volumes/PopulatedAPFS/Documents"
mkdir -p "/Volumes/PopulatedAPFS/Images"
mkdir -p "/Volumes/PopulatedAPFS/Data"

# Create many files to force B-tree population (reduced for smaller DMG)
for i in {1..20}; do
    echo "Test file content $i" > "/Volumes/PopulatedAPFS/Documents/file_$i.txt"
done

for i in {1..10}; do
    echo "Data file $i with some content to make it larger" > "/Volumes/PopulatedAPFS/Data/data_$i.dat"
    # Make some files larger
    for j in {1..5}; do
        echo "Additional line $j in file $i" >> "/Volumes/PopulatedAPFS/Data/data_$i.dat"
    done
done

# Copy files from desktop if available
if [ -f "$DESKTOP_DIR/DevOps Maturity Model v0.8.xlsx" ]; then
    cp "$DESKTOP_DIR/DevOps Maturity Model v0.8.xlsx" "/Volumes/PopulatedAPFS/Documents/"
fi

if [ -f "$DESKTOP_DIR/Asia Itinerary 20204.xlsx" ]; then
    cp "$DESKTOP_DIR/Asia Itinerary 20204.xlsx" "/Volumes/PopulatedAPFS/Documents/"
fi

# Create some symbolic links and special files
ln -s "/Volumes/PopulatedAPFS/Documents/file_1.txt" "/Volumes/PopulatedAPFS/symlink_test.txt"

# Force synchronization and filesystem consistency to ensure object mappings are committed
echo "Synchronizing filesystem to commit all transactions and object mappings..."
sync
# Give APFS time to commit all pending transactions and create object mappings
sleep 3
# Force filesystem check and repair to ensure consistency
diskutil verifyVolume "/Volumes/PopulatedAPFS" || true
sync
# Additional wait to ensure all background I/O is complete
sleep 2

hdiutil detach "/Volumes/PopulatedAPFS"

# 4. APFS volume that's nearly full (stress test B-tree with space constraints)
echo "4. Creating nearly full APFS volume..."
hdiutil create -size 6m -fs APFS -volname "FullAPFS" -type UDIF "$TESTS_DIR/full_apfs.dmg"
hdiutil attach "$TESTS_DIR/full_apfs.dmg"

# Fill most of the space
mkdir "/Volumes/FullAPFS/bigfiles"
for i in {1..3}; do
    # Create ~1MB files to fill most of the 6MB space
    dd if=/dev/random of="/Volumes/FullAPFS/bigfiles/large_$i.bin" bs=1024 count=1024 2>/dev/null
done

# Add some regular files too
echo "File in full volume" > "/Volumes/FullAPFS/test.txt"
mkdir "/Volumes/FullAPFS/docs"
echo "Documentation file" > "/Volumes/FullAPFS/docs/readme.md"

hdiutil detach "/Volumes/FullAPFS"

# 5. APFS volume with deleted files (B-tree with deleted entries)
echo "5. Creating APFS volume with deleted files..."
cp "$TESTS_DIR/populated_apfs.dmg" "$TESTS_DIR/deleted_files_apfs.dmg"
hdiutil attach "$TESTS_DIR/deleted_files_apfs.dmg"

# Delete half the files to create deleted B-tree entries
for i in {1..10}; do
    rm -f "/Volumes/PopulatedAPFS/Documents/file_$i.txt" 2>/dev/null || true
done

rm -rf "/Volumes/PopulatedAPFS/Data" 2>/dev/null || true
echo "File after deletions" > "/Volumes/PopulatedAPFS/after_delete.txt"

hdiutil detach "/Volumes/PopulatedAPFS"

echo "✓ Created comprehensive test DMG suite (1-10MB sizes):"
echo "  - empty_apfs.dmg (5MB): Empty APFS volume (minimal B-trees)"
echo "  - basic_apfs.dmg (8MB): Basic files (simple B-tree structures)"
echo "  - populated_apfs.dmg (10MB): Many files (fully populated B-trees)"
echo "  - full_apfs.dmg (6MB): Nearly full volume (space-constrained B-trees)"
echo "  - deleted_files_apfs.dmg (10MB): Volume with deleted files (B-tree with deletions)"

echo ""
echo "DMG sizes:"
ls -lh "$TESTS_DIR"/*.dmg

echo ""
echo "Test DMG suite creation complete!"