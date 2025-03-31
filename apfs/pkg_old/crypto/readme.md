/*
Keybag Management:

CreateKeybag(volumeUUID UUID)
AddKeyToKeybag(keybag *Keybag, key []byte, tag KeybagEntryTag)
RetrieveKeyFromKeybag(keybag *Keybag, uuid UUID, tag KeybagEntryTag)
RotateKeybag(keybag *Keybag)
SerializeKeybag(keybag *Keybag)
DeserializeKeybag(data []byte)

Key Management:

GenerateVolumeEncryptionKey()
GenerateKeyEncryptionKey()
WrapKey(key, wrapperKey []byte)
UnwrapKey(wrappedKey, wrapperKey []byte)
DeriveKeyFromPassword(password string, salt []byte)
GenerateRecoveryKey()

Encryption Operations:

EncryptData(data, key []byte)
DecryptData(encryptedData, key []byte)
EncryptMetadata(metadata []byte, vek []byte)
DecryptMetadata(encryptedMetadata []byte, vek []byte)

Volume Encryption:

EncryptVolume(volume *Volume, password string)
DecryptVolume(volume *Volume, password string)
UnlockVolumeWithPassword(volume *Volume, password string)
UnlockVolumeWithRecoveryKey(volume *Volume, recoveryKey string)

Encryption State Management:

InitializeEncryptionState(volume *Volume)
UpdateEncryptionState(volume *Volume)
RollEncryptionKey(volume *Volume)
MarkVolumeEncrypted(volume *Volume)

Advanced Encryption Features:

SupportsSoftwareEncryption()
SupportsHardwareEncryption()
GetEncryptionCapabilities()

Recovery and Backup:

BackupEncryptionKeys(volume *Volume)
RestoreEncryptionKeys(volume *Volume, backupData []byte)
AddRecoveryKeyToVolume(volume *Volume, recoveryKey string)
RemoveRecoveryKeyFromVolume(volume *Volume, recoveryKey string)

Utility Functions:

GenerateSalt()
ComputeKeyDerivationHash()
ValidatePasswordStrength(password string)

Encryption Rolling:

StartEncryptionRolling(volume *Volume)
CompleteEncryptionRolling(volume *Volume)
PauseEncryptionRolling(volume *Volume)
ResumeEncryptionRolling(volume *Volume)

Error Handling and Logging:

LogEncryptionEvent(event EncryptionEvent)
HandleEncryptionError(err EncryptionError)

Protection Class Management:

SetProtectionClass(file *File, class ProtectionClass)
GetProtectionClass(file *File)
ListSupportedProtectionClasses()
Apple File System Reference documentation's encryption section (pages 134-149).
*/

confirm you agree with these set of funcs ? these are funcs are supposed to be the lowlevel funcs, there will be a seperate encryption manager in a services package later.