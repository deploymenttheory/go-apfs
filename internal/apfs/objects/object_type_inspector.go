package objects

import (
	"github.com/deploymenttheory/go-apfs/internal/types"
)

// ObjectInspector implements ObjectTypeInspector for a given ObjPhysT
type ObjectInspector struct {
	Obj *types.ObjPhysT
}

func NewObjectInspector(obj *types.ObjPhysT) *ObjectInspector {
	return &ObjectInspector{Obj: obj}
}

func (o *ObjectInspector) Type() uint32 {
	return o.Obj.OType & types.ObjectTypeMask
}

func (o *ObjectInspector) Subtype() uint32 {
	return o.Obj.OSubtype
}

func (o *ObjectInspector) TypeName() string {
	switch o.Type() {
	case types.ObjectTypeNxSuperblock:
		return "NX Superblock"
	case types.ObjectTypeBtree:
		return "B-tree Root"
	case types.ObjectTypeBtreeNode:
		return "B-tree Node"
	case types.ObjectTypeFs:
		return "APFS Volume"
	case types.ObjectTypeOmap:
		return "Object Map"
	case types.ObjectTypeInvalid:
		return "Invalid"
	default:
		return "Unknown"
	}
}

func (o *ObjectInspector) IsVirtual() bool {
	return (o.Obj.OType & types.ObjStorageTypeMask) == types.ObjVirtual
}

func (o *ObjectInspector) IsEphemeral() bool {
	return (o.Obj.OType & types.ObjStorageTypeMask) == types.ObjEphemeral
}

func (o *ObjectInspector) IsPhysical() bool {
	return (o.Obj.OType & types.ObjStorageTypeMask) == types.ObjPhysical
}

func (o *ObjectInspector) IsEncrypted() bool {
	return (o.Obj.OType & types.ObjEncrypted) != 0
}

func (o *ObjectInspector) IsNonpersistent() bool {
	return (o.Obj.OType & types.ObjNonpersistent) != 0
}

func (o *ObjectInspector) HasHeader() bool {
	return (o.Obj.OType & types.ObjNoheader) == 0
}
