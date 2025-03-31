// File: pkg/crypto/volume_encryption.go
package crypto

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/deploymenttheory/go-apfs/apfs/pkg/container"
	"github.com/deploymenttheory/go-apfs/apfs/pkg/types"
)

// EncryptVolume encrypts a volume using the specified password
func EncryptVolume(volume *types.APFSSuperblock, password string, device types.BlockDevice) error {
	if volume == nil {
		return errors.New("volume is nil")
	}

	// Check if volume is already encrypted
	if !volume.IsUnencrypted() {
		return errors.New("volume is already encrypted")
	}

	// Generate a random Volume Encryption Key (VEK)
	vek, err := GenerateVolumeEncryptionKey()
	if err != nil {
		return fmt.Errorf("failed to generate volume encryption key: %w", err)
	}

	// Derive a Key Encryption Key (KEK) from the password
	salt := volume.UUID[:]
	kek, err := DeriveKeyFromPassword(password, salt)
	if err != nil {
		return fmt.Errorf("failed to derive key encryption key: %w", err)
	}

	// Wrap the VEK with the KEK
	wrappedVEK, err := WrapKey(vek, kek)
	if err != nil {
		return fmt.Errorf("failed to wrap volume encryption key: %w", err)
	}

	// Create a keybag for the volume
	keybag := CreateKeybag(volume.UUID)

	// Add the wrapped VEK to the keybag
	err = AddKeyToKeybag(keybag, wrappedVEK, types.KBTagVolumeKey, volume.UUID)
	if err != nil {
		return fmt.Errorf("failed to add volume key to keybag: %w", err)
	}

	// Add the wrapped KEK as an unlock record
	unlockRecord, err := WrapKey(kek, kek) // Self-wrapped for unlock records
	if err != nil {
		return fmt.Errorf("failed to create unlock record: %w", err)
	}

	err = AddKeyToKeybag(keybag, unlockRecord, types.KBTagVolumeUnlockRecords, volume.UUID)
	if err != nil {
		return fmt.Errorf("failed to add unlock record to keybag: %w", err)
	}

	// Initialize encryption state
	encryptionState, err := InitializeEncryptionState(volume)
	if err != nil {
		return fmt.Errorf("failed to initialize encryption state: %w", err)
	}

	// Save the keybag to the volume
	// This would require the SpaceManager to allocate space

	// Mark the volume as being encrypted
	volume.FSFlags &= ^types.APFSFSUnencrypted
	volume.IncompatFeatures |= types.APFSIncompatEncRolled

	// Start the encryption process
	// In a real implementation, this would start an asynchronous process

	return nil
}

// UnlockVolumeWithPassword unlocks an encrypted volume using a password
func UnlockVolumeWithPassword(volume *types.APFSSuperblock, password string, device types.BlockDevice) ([]byte, error) {
	if volume == nil {
		return nil, errors.New("volume is nil")
	}

	// Check if volume is encrypted
	if volume.IsUnencrypted() {
		return nil, errors.New("volume is not encrypted")
	}

	// Read the volume's keybag
	// In a full implementation, this would use ReadKeybagFromDevice

	// For demonstration purposes, assume we have the keybag
	keybag := &Keybag{} // This would be loaded from disk

	// Derive the KEK from the password
	salt := volume.UUID[:]
	kek, err := DeriveKeyFromPassword(password, salt)
	if err != nil {
		return nil, fmt.Errorf("failed to derive key encryption key: %w", err)
	}

	// Get the wrapped VEK from the keybag
	wrappedVEK, err := RetrieveKeyFromKeybag(keybag, volume.UUID, types.KBTagVolumeKey)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve volume key: %w", err)
	}

	// Unwrap the VEK using the KEK
	vek, err := UnwrapKey(wrappedVEK, kek)
	if err != nil {
		return nil, fmt.Errorf("incorrect password or corrupted key: %w", err)
	}

	// Return the unwrapped VEK
	return vek, nil
}

// UnlockVolumeWithRecoveryKey unlocks an encrypted volume using a recovery key
func UnlockVolumeWithRecoveryKey(volume *types.APFSSuperblock, recoveryKey string, device types.BlockDevice) ([]byte, error) {
	if volume == nil {
		return nil, errors.New("volume is nil")
	}

	// Check if volume is encrypted
	if volume.IsUnencrypted() {
		return nil, errors.New("volume is not encrypted")
	}

	// Parse the recovery key
	recoveryKeyBytes, err := ParseRecoveryKey(recoveryKey)
	if err != nil {
		return nil, fmt.Errorf("invalid recovery key format: %w", err)
	}

	// Read the volume's keybag
	// In a full implementation, this would use ReadKeybagFromDevice

	// For demonstration purposes, assume we have the keybag
	keybag := &Keybag{} // This would be loaded from disk

	// Get the wrapped VEK from the keybag using the recovery key UUID
	recoveryKeyUUID, _ := types.UUIDFromString(types.APFSFVPersonalRecoveryKeyUUID)
	wrappedVEK, err := RetrieveKeyFromKeybag(keybag, recoveryKeyUUID, types.KBTagVolumeKey)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve volume key for recovery: %w", err)
	}

	// Unwrap the VEK using the recovery key
	vek, err := UnwrapKey(wrappedVEK, recoveryKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("incorrect recovery key or corrupted key: %w", err)
	}

	// Return the unwrapped VEK
	return vek, nil
}

// InitializeEncryptionState creates an initial encryption state object for a volume
func InitializeEncryptionState(volume *types.APFSSuperblock) (*types.ERStatePhys, error) {
	if volume == nil {
		return nil, errors.New("volume is nil")
	}

	// Create a new encryption rolling state
	state := &types.ERStatePhys{
		Header: types.ObjectHeader{
			// Initialize the header
			Type:    types.ObjEphemeral | types.ObjectTypeERState,
			Subtype: 0,
		},
		Magic:   types.ERMagic,
		Version: types.ERVersion,
		Flags:   types.ERSBFlagEncrypting, // Set flag to indicate encryption is in progress
	}

	// Generate a unique transaction ID for this encryption operation
	// In a real implementation, this would use the container's next transaction ID

	// Initialize progress tracking
	state.Progress = 0
	state.TotalBlkToEncrypt = volume.AllocCount // Total blocks to encrypt

	return state, nil
}

// UpdateEncryptionState updates the encryption state during the encryption process
func UpdateEncryptionState(state *types.ERStatePhys, progress uint64) error {
	if state == nil {
		return errors.New("encryption state is nil")
	}

	// Update progress
	state.Progress = progress

	// Check if encryption is complete
	if progress >= state.TotalBlkToEncrypt {
		// Clear the encrypting flag and mark as completed
		state.Flags &= ^types.ERSBFlagEncrypting
	}

	return nil
}

// AddRecoveryKeyToVolume adds a recovery key to a volume
func AddRecoveryKeyToVolume(volume *types.APFSSuperblock, password string, device types.BlockDevice) (string, error) {
	if volume == nil {
		return "", errors.New("volume is nil")
	}

	// Check if volume is encrypted
	if volume.IsUnencrypted() {
		return "", errors.New("volume is not encrypted")
	}

	// First unlock the volume with the password to get the VEK
	vek, err := UnlockVolumeWithPassword(volume, password, device)
	if err != nil {
		return "", fmt.Errorf("failed to unlock volume with password: %w", err)
	}

	// Generate a new recovery key
	recoveryKeyBytes, formattedRecoveryKey, err := GenerateRecoveryKey()
	if err != nil {
		return "", fmt.Errorf("failed to generate recovery key: %w", err)
	}

	// Read the volume's keybag
	// In a full implementation, this would use ReadKeybagFromDevice

	// For demonstration purposes, assume we have the keybag
	keybag := &Keybag{} // This would be loaded from disk

	// Wrap the VEK with the recovery key
	wrappedVEK, err := WrapKey(vek, recoveryKeyBytes)
	if err != nil {
		return "", fmt.Errorf("failed to wrap VEK with recovery key: %w", err)
	}

	// Add the recovery key entry to the keybag
	recoveryKeyUUID, _ := types.UUIDFromString(types.APFSFVPersonalRecoveryKeyUUID)
	err = AddKeyToKeybag(keybag, wrappedVEK, types.KBTagVolumeKey, recoveryKeyUUID)
	if err != nil {
		return "", fmt.Errorf("failed to add recovery key to keybag: %w", err)
	}

	// Write the updated keybag back to the device
	// In a full implementation, this would use WriteKeybagToDevice

	return formattedRecoveryKey, nil
}

// RemoveRecoveryKeyFromVolume removes a recovery key from a volume
func RemoveRecoveryKeyFromVolume(volume *types.APFSSuperblock, password string, device types.BlockDevice) error {
	if volume == nil {
		return errors.New("volume is nil")
	}

	// Check if volume is encrypted
	if volume.IsUnencrypted() {
		return errors.New("volume is not encrypted")
	}

	// First unlock the volume with the password to confirm it's valid
	_, err := UnlockVolumeWithPassword(volume, password, device)
	if err != nil {
		return fmt.Errorf("failed to unlock volume with password: %w", err)
	}

	// Read the volume's keybag
	// In a full implementation, this would use ReadKeybagFromDevice

	// For demonstration purposes, assume we have the keybag
	keybag := &Keybag{} // This would be loaded from disk

	// Remove the recovery key entry from the keybag
	recoveryKeyUUID, _ := types.UUIDFromString(types.APFSFVPersonalRecoveryKeyUUID)
	err = RemoveKeyFromKeybag(keybag, recoveryKeyUUID, types.KBTagVolumeKey)
	if err != nil {
		return fmt.Errorf("failed to remove recovery key from keybag: %w", err)
	}

	// Write the updated keybag back to the device
	// In a full implementation, this would use WriteKeybagToDevice

	return nil
}

// SupportsSoftwareEncryption checks if the device supports software encryption
func SupportsSoftwareEncryption() bool {
	// Software encryption is always supported as our implementation is in software
	return true
}

// SupportsHardwareEncryption checks if the device supports hardware encryption
func SupportsHardwareEncryption() bool {
	// This would need to check the actual hardware capabilities
	// For a portable Go implementation, we'll return false as we're implementing software encryption
	return false
}

// GetEncryptionCapabilities returns information about supported encryption features
func GetEncryptionCapabilities() map[string]bool {
	return map[string]bool{
		"software_encryption": SupportsSoftwareEncryption(),
		"hardware_encryption": SupportsHardwareEncryption(),
		"key_rolling":         true,
		"recovery_key":        true,
		"password_auth":       true,
	}
}

// StartEncryptionRolling initiates the encryption key rolling process
func StartEncryptionRolling(volume *types.APFSSuperblock, oldPassword, newPassword string, device types.BlockDevice, spaceman types.SpaceManager) error {
	if volume == nil {
		return errors.New("volume is nil")
	}

	// Check if volume is encrypted
	if volume.IsUnencrypted() {
		return errors.New("volume is not encrypted")
	}

	// Check if key rolling is already in progress
	if volume.ERStateOID != 0 {
		erState, err := ReadEncryptionRollingState(volume.ERStateOID, device)
		if err == nil && (erState.Flags&types.ERSBFlagKeyrolling) != 0 {
			return errors.New("encryption key rolling is already in progress")
		}
	}

	// Verify the old password
	oldVEK, err := UnlockVolumeWithPassword(volume, oldPassword, device)
	if err != nil {
		return fmt.Errorf("failed to unlock volume with old password: %w", err)
	}

	// Generate a new VEK
	newVEK, err := GenerateVolumeEncryptionKey()
	if err != nil {
		return fmt.Errorf("failed to generate new volume encryption key: %w", err)
	}

	// Derive a new KEK from the new password
	salt := volume.UUID[:]
	newKEK, err := DeriveKeyFromPassword(newPassword, salt)
	if err != nil {
		return fmt.Errorf("failed to derive new key encryption key: %w", err)
	}

	// Wrap the new VEK with the new KEK
	wrappedNewVEK, err := WrapKey(newVEK, newKEK)
	if err != nil {
		return fmt.Errorf("failed to wrap new volume encryption key: %w", err)
	}

	// Get current transaction ID
	nxsb, err := container.ReadNXSuperblock(device, 0)
	if err != nil {
		return fmt.Errorf("failed to read container superblock: %w", err)
	}
	currentXID := nxsb.NextXID

	// Initialize encryption rolling state
	state := &types.ERStatePhys{
		Header: types.ObjectHeader{
			OID:     0, // Will be set when allocated
			XID:     currentXID,
			Type:    types.ObjEphemeral | types.ObjectTypeERState,
			Subtype: 0,
		},
		Magic:             types.ERMagic,
		Version:           types.ERVersion,
		Flags:             types.ERSBFlagKeyrolling, // Set flag to indicate key rolling is in progress
		SnapXID:           currentXID,
		Progress:          0,
		TotalBlkToEncrypt: volume.AllocCount, // Total blocks to re-encrypt
	}

	// Extension to store the old and new VEK
	// Assuming we have an extension field in ERStatePhys
	// If not, we'd need to add it or use a separate data structure
	keyData := struct {
		OldVEK [VolumeEncryptionKeySize]byte
		NewVEK [VolumeEncryptionKeySize]byte
	}{}
	copy(keyData.OldVEK[:], oldVEK)
	copy(keyData.NewVEK[:], newVEK)

	// Encrypt this sensitive data with the new KEK
	encryptedKeyData, err := EncryptData(keyData, newKEK, EncryptionModeAESCBC)
	if err != nil {
		return fmt.Errorf("failed to encrypt key data: %w", err)
	}
	state.KeyData = encryptedKeyData // Assuming we add this field to ERStatePhys

	// Allocate space for the encryption state
	stateData, err := serializeERStatePhys(state)
	if err != nil {
		return fmt.Errorf("failed to serialize encryption state: %w", err)
	}

	// Allocate space in the container
	stateAddr, err := spaceman.AllocateBlock()
	if err != nil {
		return fmt.Errorf("failed to allocate block for encryption state: %w", err)
	}

	// Write the state to disk
	err = device.WriteBlock(stateAddr, stateData)
	if err != nil {
		// Free the block if write fails
		spaceman.FreeBlock(stateAddr)
		return fmt.Errorf("failed to write encryption state: %w", err)
	}

	// Allocate an OID for the state
	stateOID := nxsb.NextOID
	nxsb.NextOID++

	// Update the OMap to map the OID to the physical address
	omap, err := container.ReadOMapPhys(device, types.PAddr(nxsb.OMapOID))
	if err != nil {
		return fmt.Errorf("failed to read object map: %w", err)
	}

	omapWrapper := &container.OMap{
		Phys:     omap,
		Device:   device,
		Spaceman: spaceman,
	}

	err = omapWrapper.Set(stateOID, stateAddr)
	if err != nil {
		return fmt.Errorf("failed to update object map: %w", err)
	}

	// Update volume flags to indicate key rolling is in progress
	volume.IncompatFeatures |= types.APFSIncompatEncRolled

	// Update the volume superblock to reference the encryption rolling state
	volume.ERStateOID = stateOID

	// Write the updated volume superblock
	// This depends on how you write volume superblocks in your system
	// For example:
	err = container.WriteVolumeSuperblock(device, volume)
	if err != nil {
		return fmt.Errorf("failed to update volume superblock: %w", err)
	}

	// Write the updated container superblock to persist NextOID
	err = container.WriteNXSuperblock(device, 0, nxsb)
	if err != nil {
		return fmt.Errorf("failed to update container superblock: %w", err)
	}

	return nil
}

// ReadEncryptionRollingState reads the encryption rolling state from disk
func ReadEncryptionRollingState(stateOID types.OID, device types.BlockDevice) (*types.ERStatePhys, error) {
	if stateOID == 0 {
		return nil, errors.New("invalid state object ID: zero")
	}

	if device == nil {
		return nil, errors.New("device is nil")
	}

	// Read the container superblock (block 0)
	nxsb, err := container.ReadNXSuperblock(device, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to read container superblock: %w", err)
	}

	// Read the OMap structure
	omap, err := container.ReadOMapPhys(device, types.PAddr(nxsb.OMapOID))
	if err != nil {
		return nil, fmt.Errorf("failed to read object map: %w", err)
	}

	// Create OMap wrapper
	omapWrapper := &container.OMap{
		Phys:   omap,
		Device: device,
		// Spaceman not needed for read-only operation
	}

	// Resolve the physical address of the encryption rolling state
	paddr, err := omapWrapper.Resolve(stateOID)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve encryption state object (OID %d): %w", stateOID, err)
	}

	// Read the object data
	blockData, err := device.ReadBlock(paddr)
	if err != nil {
		return nil, fmt.Errorf("failed to read encryption state block at address %d: %w", paddr, err)
	}

	// Parse the object data into an ERStatePhys structure
	state := &types.ERStatePhys{}

	// First read the object header
	r := bytes.NewReader(blockData)
	err = binary.Read(r, binary.LittleEndian, &state.Header)
	if err != nil {
		return nil, fmt.Errorf("failed to read encryption state header: %w", err)
	}

	// Verify the object type
	if state.Header.Type != (types.ObjEphemeral | types.ObjectTypeERState) {
		return nil, fmt.Errorf("invalid encryption state object type: 0x%x", state.Header.Type)
	}

	// Read the rest of the structure
	err = binary.Read(r, binary.LittleEndian, &state.Magic)
	if err != nil {
		return nil, fmt.Errorf("failed to read encryption state magic: %w", err)
	}

	err = binary.Read(r, binary.LittleEndian, &state.Version)
	if err != nil {
		return nil, fmt.Errorf("failed to read encryption state version: %w", err)
	}

	err = binary.Read(r, binary.LittleEndian, &state.Flags)
	if err != nil {
		return nil, fmt.Errorf("failed to read encryption state flags: %w", err)
	}

	err = binary.Read(r, binary.LittleEndian, &state.SnapXID)
	if err != nil {
		return nil, fmt.Errorf("failed to read encryption state snapshot XID: %w", err)
	}

	err = binary.Read(r, binary.LittleEndian, &state.Progress)
	if err != nil {
		return nil, fmt.Errorf("failed to read encryption state progress: %w", err)
	}

	err = binary.Read(r, binary.LittleEndian, &state.TotalBlkToEncrypt)
	if err != nil {
		return nil, fmt.Errorf("failed to read encryption state total blocks: %w", err)
	}

	// Check magic value
	if state.Magic != types.ERMagic {
		return nil, fmt.Errorf("invalid encryption state magic: 0x%x", state.Magic)
	}

	// Check version
	if state.Version != types.ERVersion {
		return nil, fmt.Errorf("unsupported encryption state version: %d", state.Version)
	}

	return state, nil
}
