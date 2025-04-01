package extendedfields

import "github.com/deploymenttheory/go-apfs/internal/interfaces"

type ExtendedFieldTypeResolver struct {
	types map[uint8]struct {
		name, desc string
	}
}

func NewExtendedFieldTypeResolver() interfaces.ExtendedFieldTypeResolver {
	return &ExtendedFieldTypeResolver{
		types: map[uint8]struct {
			name, desc string
		}{
			1:  {"Document ID", "The APFS document ID for the file."},
			2:  {"Finder Info", "Extended Finder information blob."},
			3:  {"Sparse Bytes", "Count of sparse bytes in the file's data stream."},
			4:  {"Previous File Size", "Used for crash recovery; file size prior to last update."},
			5:  {"Snapshot XID", "Transaction ID of the snapshot that created this inode."},
			6:  {"Delta Tree OID", "Tree OID for snapshot delta information."},
			7:  {"Device ID", "Device identifier for special files."},
			8:  {"Original Sync Root ID", "Original hierarchy inode ID for sync root."},
			9:  {"Sibling ID", "Hardlink sibling ID."},
			10: {"Filesystem UUID", "UUID of mounted filesystem target."},
		},
	}
}

func (r *ExtendedFieldTypeResolver) ResolveName(fieldType uint8) string {
	if info, ok := r.types[fieldType]; ok {
		return info.name
	}
	return "Unknown"
}

func (r *ExtendedFieldTypeResolver) ResolveDescription(fieldType uint8) string {
	if info, ok := r.types[fieldType]; ok {
		return info.desc
	}
	return "No description available"
}

func (r *ExtendedFieldTypeResolver) ListSupportedFieldTypes() []uint8 {
	out := make([]uint8, 0, len(r.types))
	for t := range r.types {
		out = append(out, t)
	}
	return out
}
