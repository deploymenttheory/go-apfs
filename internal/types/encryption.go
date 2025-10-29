package types

// Encryption (pages 135-149)
// Apple File System supports encryption in the data structures used for containers, volumes, and files.

// JCryptoKeyT is the key half of a per-file encryption state record.
// Reference: page 137
type JCryptoKeyT struct {
	// The record's header. (page 137)
	// The object identifier in the header is the file-system object's identifier.
	// The type in the header is always APFS_TYPE_CRYPTO_STATE.
	Hdr JKeyT
}

// JCryptoValT is the value half of a per-file encryption state record.
// Reference: page 137
type JCryptoValT struct {
	// The reference count. (page 137)
	// The encryption state record can be deleted when its reference count reaches zero.
	Refcnt uint32

	// The encryption state information. (page 138)
	// If this encryption state record is used by the file-system tree rather than by a file,
	// this field is an instance of wrapped_meta_crypto_state_t and the key used is always
	// the volume encryption key (VEK).
	State WrappedCryptoStateT
}

// WrappedCryptoStateT is a wrapped key used for per-file encryption.
// Reference: page 138
type WrappedCryptoStateT struct {
	// The major version for this structure's layout. (page 138)
	// The current value of this field is five. If backward-incompatible changes are made to this data structure
	// in the future, the major version number will be incremented.
	// This structure is equivalent to a structure used by iOS for per-file encryption on HFS-Plus;
	// versions four and earlier were used by previous versions of that structure.
	MajorVersion uint16

	// The major version for this structure's layout. (page 138)
	// The current value of this field is zero. If backward-compatible changes are made to this data structure
	// in the future, the minor version number will be incremented.
	MinorVersion uint16

	// The encryption state's flags. (page 139)
	// There are currently none defined.
	Cpflags CryptoFlagsT

	// The protection class associated with the key. (page 139)
	// For possible values and the bit mask that must be used, see Protection Classes.
	PersistentClass CpKeyClassT

	// The version of the OS that created this structure. (page 139)
	// This field is used as part of key rolling.
	KeyOsVersion CpKeyOsVersionT

	// The version of the key. (page 139)
	// Set this field to one when creating a new instance, and increment it by one when rolling to a new key.
	KeyRevision CpKeyRevisionT

	// The size, in bytes, of the wrapped key data. (page 139)
	// The maximum value of this field is CP_MAX_WRAPPEDKEYSIZE.
	KeyLen uint16

	// The wrapped key data, limited to CP_MAX_WRAPPEDKEYSIZE bytes.
	// Reference: page 139
	PersistentKey [CpMaxWrappedkeysize]byte
}

// CpMaxWrappedkeysize is the size, in bytes, of the largest possible key.
// Reference: page 139
const CpMaxWrappedkeysize uint16 = 128

// WrappedMetaCryptoStateT contains information about how the volume encryption key (VEK)
// is used to encrypt a file.
// Reference: page 140
type WrappedMetaCryptoStateT struct {
	// The major version for this structure's layout. (page 140)
	// The value of this field is always five. This structure is equivalent to a structure used by iOS
	// for per-file encryption on HFS-Plus; versions four and earlier were used
	// by previous versions of that structure.
	MajorVersion uint16

	// The major version for this structure's layout. (page 140)
	// The value of this field is always zero.
	MinorVersion uint16

	// The encryption state's flags. (page 140)
	// There are currently none defined.
	Cpflags CryptoFlagsT

	// The protection class associated with the key. (page 140)
	// For possible values, see Protection Classes.
	PersistentClass CpKeyClassT

	// The version of the OS that created this structure. (page 141)
	// For information about how the major version number, minor version number,
	// and build number are packed into 32 bits, see cp_key_os_version_t.
	KeyOsVersion CpKeyOsVersionT

	// The version of the key. (page 141)
	// Set this field to one when creating a new instance.
	KeyRevision CpKeyRevisionT

	// Reserved. (page 141)
	// Populate this field with zero when you create a new instance of this structure,
	// and preserve its value when you modify an existing instance.
	Unused uint16
}

// Encryption Types (page 141)

// CpKeyClassT is a protection class.
// Reference: page 141
type CpKeyClassT uint32

// CpKeyOsVersionT is an OS version and build number.
// Reference: page 141
type CpKeyOsVersionT uint32

// CpKeyRevisionT is a version number for an encryption key.
// Reference: page 142
type CpKeyRevisionT uint16

// CryptoFlagsT contains flags used by an encryption state.
// Reference: page 142
type CryptoFlagsT uint32

// Protection Classes (pages 142-143)

const (
	// ProtectionClassDirNone is the directory default.
	// Reference: page 142
	// This protection class is used only on devices running iOS.
	// Files with this protection class use their containing directory's default protection class,
	// which is set by the default_protection_class field of j_inode_val_t.
	ProtectionClassDirNone CpKeyClassT = 0

	// ProtectionClassA indicates complete protection.
	// Reference: page 142
	ProtectionClassA CpKeyClassT = 1

	// ProtectionClassB indicates protected unless open.
	// Reference: page 143
	ProtectionClassB CpKeyClassT = 2

	// ProtectionClassC indicates protected until first user authentication.
	// Reference: page 143
	ProtectionClassC CpKeyClassT = 3

	// ProtectionClassD indicates no protection.
	// Reference: page 143
	ProtectionClassD CpKeyClassT = 4

	// ProtectionClassF indicates no protection with nonpersistent key.
	// Reference: page 143
	// The behavior of this protection class is the same as Class D, except the key isn't stored in any
	// persistent way. This protection class is suitable for temporary files that aren't needed after
	// rebooting the device, such as a virtual machine's swap file.
	ProtectionClassF CpKeyClassT = 6

	// ProtectionClassM has no overview available.
	// Reference: page 143
	ProtectionClassM CpKeyClassT = 14
)

// CpEffectiveClassmask is the bit mask used to access the protection class.
// Reference: page 143
// All other bits are reserved. Populate those bits with zero when you create a wrapped key,
// and preserve their value when you modify an existing wrapped key.
const CpEffectiveClassmask CpKeyClassT = 0x0000001f

// Encryption Identifiers (page 144)

// CryptoSwId is the identifier of a placeholder encryption state used when software encryption is in use.
// Reference: page 144
// There is no associated encryption key for this encryption state.
// All the fields of the corresponding j_crypto_val_t structure have a value of zero.
const CryptoSwId uint64 = 4

// CryptoReserved5 is reserved.
// Reference: page 144
// Don't create an encryption state object with this identifier.
// If you find an object with this identifier in production, file a bug against the Apple File System implementation.
const CryptoReserved5 uint64 = 5

// ApfsUnassignedCryptoId is the identifier of a placeholder encryption state used when cloning files.
// Reference: page 144
// As a performance optimization when cloning a file, Apple's implementation sets this placeholder value
// on the clone and continues to use the original file's encryption state for both that file and its clone.
// If the clone is modified, a new encryption state object is created for the clone.
// Creating a new encryption state object is relatively expensive, and usually takes longer than the cloning process.
const ApfsUnassignedCryptoId uint64 = ^uint64(0) // ~0ULL

// KbLockerT is a keybag.
// Reference: page 144
type KbLockerT struct {
	// The keybag version. (page 145)
	// The value of this field is APFS_KEYBAG_VERSION.
	KlVersion uint16

	// The number of entries in the keybag. (page 145)
	KlNkeys uint16

	// The size, in bytes, of the data stored in the kl_entries field. (page 145)
	KlNbytes uint32

	// Reserved. (page 145)
	// Populate this field with zero when you create a new keybag,
	// and preserve its value when you modify an existing keybag.
	// This field is padding.
	Padding [8]byte

	// The entries. (page 145)
	KlEntries []KeybagEntryT
}

// ApfsKeybagVersion is the first version of the keybag.
// Reference: page 145
// Version one was used during prototyping of Apple File System, and uses an incompatible, undocumented layout.
// If you find a keybag in production whose version is less than two, file a bug against the Apple File System
// implementation.
const ApfsKeybagVersion uint16 = 2

// KeybagEntryT is an entry in a keybag.
// Reference: page 146
type KeybagEntryT struct {
	// In a container's keybag, the UUID of a volume; in a volume's keybag, the UUID of a user. (page 146)
	KeUuid UUID

	// A description of the kind of data stored in this keybag entry. (page 146)
	// For possible values, see Keybag Tags.
	KeTag uint16

	// The length, in bytes, of the keybag entry's data. (page 146)
	// The value of this field must be less than APFS_VOL_KEYBAG_ENTRY_MAX_SIZE.
	KeKeylen uint16

	// Reserved. (page 146)
	// Populate this field with zero when you create a new keybag entry,
	// and preserve its value when you modify an existing entry.
	// This field is padding.
	Padding [4]byte

	// The keybag entry's data. (page 146)
	// The data stored this field depends on the tag and whether this is an entry in a container
	// or volume's keybag, as described in Keybag Tags.
	KeKeydata []byte
}

// ApfsVolKeybagEntryMaxSize is the largest size, in bytes, of a keybag entry.
// Reference: page 147
const ApfsVolKeybagEntryMaxSize uint16 = 512

// ApfsFvPersonalRecoveryKeyUuid is the user UUID used by a keybag record that contains a personal recovery key.
// Reference: page 147
// The personal recovery key is generated during the initial volume-encryption process,
// and it's stored by the user as a paper printout.
// You use it the same way you use a user's password to unwrap the corresponding KEK.
var ApfsFvPersonalRecoveryKeyUuid = UUID{
	0xEB, 0xC6, 0xC0, 0x64,
	0x00, 0x00, 0x11, 0xAA,
	0xAA, 0x11, 0x00, 0x30,
	0x65, 0x43, 0xEC, 0xAC,
}

// ApfsFvInstitutionalRecoveryKeyUuid is the UUID used for institutional recovery keys.
// Reference: APFS Advent Challenge Day 15
// The institutional recovery key allows organizations to recover encrypted volumes
// using a centrally-managed recovery key.
var ApfsFvInstitutionalRecoveryKeyUuid = UUID{
	0xC0, 0x64, 0xEB, 0xC6,
	0x00, 0x00, 0x11, 0xAA,
	0xAA, 0x11, 0x00, 0x30,
	0x65, 0x43, 0xEC, 0xAC,
}

// ApfsFvInstitutionalUserUuid is the UUID used for institutional user recovery.
// Reference: APFS Advent Challenge Day 15
// This UUID is used in volume keybags for institutional recovery scenarios.
var ApfsFvInstitutionalUserUuid = UUID{
	0x2F, 0xA3, 0x14, 0x00,
	0xBA, 0xFF, 0x4D, 0xE7,
	0xAE, 0x2A, 0xC3, 0xAA,
	0x6E, 0x1F, 0xD3, 0x40,
}

// MediaKeybagT is a keybag, wrapped up as a container-layer object.
// Reference: page 147
type MediaKeybagT struct {
	// The object's header. (page 147)
	MkObj ObjPhysT

	// The keybag data. (page 147)
	MkLocker KbLockerT
}

// Keybag Tags (pages 147-149)

// KbTag represents a description of what kind of information is stored by a keybag entry.
// Reference: page 147
type KbTag uint16

const (
	// KbTagUnknown is reserved.
	// Reference: page 148
	// This tag never appears on disk. If you find a keybag entry with this tag in production,
	// file a bug against the Apple File System implementation.
	// This value isn't reserved by Apple; non-Apple implementations of Apple File System can use it in memory.
	// For example, Apple's implementation uses this value as a wildcard that matches any tag.
	KbTagUnknown KbTag = 0

	// KbTagReserved1 is reserved.
	// Reference: page 148
	// Don't create keybag entries with this tag, but preserve any existing entries.
	KbTagReserved1 KbTag = 1

	// KbTagVolumeKey indicates the key data stores a wrapped VEK.
	// Reference: page 148
	// This tag is valid only in a container's keybag.
	KbTagVolumeKey KbTag = 2

	// KbTagVolumeUnlockRecords indicates in a container's keybag, the key data stores the location
	// of the volume's keybag; in a volume keybag, the key data stores a wrapped KEK.
	// Reference: page 148
	// This tag is used only on devices running macOS.
	// The volume's keybag location is stored as an instance of prange_t;
	// the data at that location is an instance of kb_locker_t.
	KbTagVolumeUnlockRecords KbTag = 3

	// KbTagVolumePassphraseHint indicates the key data stores a user's password hint as plain text.
	// Reference: page 148
	// This tag is valid only in a volume's keybag, and it's used only on devices running macOS.
	KbTagVolumePassphraseHint KbTag = 4

	// KbTagWrappingMKey indicates the key data stores a key that's used to wrap a media key.
	// Reference: page 149
	// This tag is used only on devices running iOS.
	KbTagWrappingMKey KbTag = 5

	// KbTagVolumeMKey indicates the key data stores a key that's used to wrap media keys on this volume.
	// Reference: page 149
	// This tag is used only on devices running iOS.
	KbTagVolumeMKey KbTag = 6

	// KbTagReservedF8 is reserved.
	// Reference: page 149
	// Don't create keybag entries with this tag, but preserve any existing entries.
	KbTagReservedF8 KbTag = 0xF8
)
