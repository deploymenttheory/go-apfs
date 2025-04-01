package objectmaps

import (
	"errors"

	"github.com/deploymenttheory/go-apfs/internal/interfaces"
	"github.com/deploymenttheory/go-apfs/internal/types"
)

type ObjectMapInspectorImpl struct {
	entries []interfaces.ObjectMapEntryReader
}

func NewObjectMapInspector(entries []interfaces.ObjectMapEntryReader) *ObjectMapInspectorImpl {
	return &ObjectMapInspectorImpl{entries: entries}
}

func (i *ObjectMapInspectorImpl) ListObjects() ([]interfaces.ObjectMapEntryReader, error) {
	return i.entries, nil
}

func (i *ObjectMapInspectorImpl) FindObjectByID(objectID types.OidT, txID ...types.XidT) (interfaces.ObjectMapEntryReader, error) {
	for _, entry := range i.entries {
		if entry.ObjectID() != objectID {
			continue
		}
		if len(txID) > 0 && entry.TransactionID() != txID[0] {
			continue
		}
		return entry, nil
	}
	return nil, errors.New("object not found")
}

func (i *ObjectMapInspectorImpl) CountObjects() (int, error) {
	return len(i.entries), nil
}

func (i *ObjectMapInspectorImpl) FindDeletedObjects() ([]interfaces.ObjectMapEntryReader, error) {
	var deleted []interfaces.ObjectMapEntryReader
	for _, entry := range i.entries {
		if entry.IsDeleted() {
			deleted = append(deleted, entry)
		}
	}
	return deleted, nil
}

func (i *ObjectMapInspectorImpl) FindEncryptedObjects() ([]interfaces.ObjectMapEntryReader, error) {
	var encrypted []interfaces.ObjectMapEntryReader
	for _, entry := range i.entries {
		if entry.IsEncrypted() {
			encrypted = append(encrypted, entry)
		}
	}
	return encrypted, nil
}
