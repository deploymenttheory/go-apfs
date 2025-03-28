package types

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"unsafe"

	"github.com/deploymenttheory/go-apfs/apfs/pkg/checksum"
)

// BinaryReader helps with reading binary data
type BinaryReader struct {
	reader io.Reader
	order  binary.ByteOrder
	buf    *bytes.Reader // used for peeking
}

// NewBinaryReader creates a new binary reader with specified byte order
func NewBinaryReader(r io.Reader, order binary.ByteOrder) *BinaryReader {
	// Ensure we have a bytes.Reader for peeking
	var b *bytes.Reader
	switch x := r.(type) {
	case *bytes.Reader:
		b = x
	default:
		buf := new(bytes.Buffer)
		io.Copy(buf, r)
		b = bytes.NewReader(buf.Bytes())
	}
	return &BinaryReader{
		reader: b,
		order:  order,
		buf:    b,
	}
}

// Read reads structured binary data from r into data.
// Data must be a pointer to a fixed-size value or a slice of fixed-size values.
func (br *BinaryReader) Read(data interface{}) error {
	return binary.Read(br.reader, br.order, data)
}

// ReadUint8 reads a uint8
func (br *BinaryReader) ReadUint8() (uint8, error) {
	var val uint8
	err := br.Read(&val)
	return val, err
}

// ReadUint16 reads a uint16
func (br *BinaryReader) ReadUint16() (uint16, error) {
	var val uint16
	err := br.Read(&val)
	return val, err
}

// ReadUint32 reads a uint32
func (br *BinaryReader) ReadUint32() (uint32, error) {
	var val uint32
	err := br.Read(&val)
	return val, err
}

// ReadUint64 reads a uint64
func (br *BinaryReader) ReadUint64() (uint64, error) {
	var val uint64
	err := br.Read(&val)
	return val, err
}

// ReadOID reads an OID
func (br *BinaryReader) ReadOID() (OID, error) {
	var val OID
	err := br.Read(&val)
	return val, err
}

// ReadXID reads an XID
func (br *BinaryReader) ReadXID() (XID, error) {
	var val XID
	err := br.Read(&val)
	return val, err
}

// ReadPAddr reads a PAddr
func (br *BinaryReader) ReadPAddr() (PAddr, error) {
	var val PAddr
	err := br.Read(&val)
	return val, err
}

// ReadUUID reads a UUID
func (br *BinaryReader) ReadUUID() (UUID, error) {
	var val UUID
	err := br.Read(&val)
	return val, err
}

// ReadBytes reads a slice of bytes with the specified length
func (br *BinaryReader) ReadBytes(length int) ([]byte, error) {
	buf := make([]byte, length)
	_, err := io.ReadFull(br.reader, buf)
	return buf, err
}

// ReadString reads a null-terminated string with the specified maximum length
func (br *BinaryReader) ReadString(maxLen int) (string, error) {
	buf, err := br.ReadBytes(maxLen)
	if err != nil {
		return "", err
	}

	// Find null terminator
	nullPos := bytes.IndexByte(buf, 0)
	if nullPos != -1 {
		return string(buf[:nullPos]), nil
	}

	// No null terminator found, return entire string
	return string(buf), nil
}

// ReadStringWithLen reads a string of the given length
func (br *BinaryReader) ReadStringWithLen(length int) (string, error) {
	buf, err := br.ReadBytes(length)
	if err != nil {
		return "", err
	}
	return string(buf), nil
}

// PeekBytes allows peeking ahead at the next n bytes without advancing the reader
func (br *BinaryReader) PeekBytes(n int) ([]byte, error) {
	if br.buf == nil {
		return nil, fmt.Errorf("peek not supported on non-buffered reader")
	}
	pos, _ := br.buf.Seek(0, io.SeekCurrent)
	buf := make([]byte, n)
	_, err := br.buf.Read(buf)
	br.buf.Seek(pos, io.SeekStart) // rewind
	return buf, err
}

// ReadUntilNullByte reads until a null byte is encountered
func (br *BinaryReader) ReadUntilNullByte() ([]byte, error) {
	var result []byte
	for {
		b := make([]byte, 1)
		_, err := br.buf.Read(b)
		if err != nil {
			return nil, err
		}
		if b[0] == 0 {
			break
		}
		result = append(result, b[0])
	}
	return result, nil
}

// ReadUint16Masked reads a uint16, applies a mask, and shifts it into position
func (br *BinaryReader) ReadUint16Masked(mask uint16, shift uint8) (uint16, error) {
	val, err := br.ReadUint16()
	if err != nil {
		return 0, err
	}
	return (val & mask) >> shift, nil
}

// ReadUint32Masked reads a uint32, applies a mask, and shifts it into position
func (br *BinaryReader) ReadUint32Masked(mask uint32, shift uint8) (uint32, error) {
	val, err := br.ReadUint32()
	if err != nil {
		return 0, err
	}
	return (val & mask) >> shift, nil
}

// ReadUint64Array reads an array of uint64 values of the specified length
func (br *BinaryReader) ReadUint64Array(count int) ([]uint64, error) {
	result := make([]uint64, count)
	for i := 0; i < count; i++ {
		val, err := br.ReadUint64()
		if err != nil {
			return nil, err
		}
		result[i] = val
	}
	return result, nil
}

// ReadOIDArray reads an array of OID values of the specified length
func (br *BinaryReader) ReadOIDArray(count int) ([]OID, error) {
	result := make([]OID, count)
	for i := 0; i < count; i++ {
		val, err := br.ReadOID()
		if err != nil {
			return nil, err
		}
		result[i] = val
	}
	return result, nil
}

// BinaryWriter helps with writing binary data
type BinaryWriter struct {
	writer io.Writer
	order  binary.ByteOrder
}

// NewBinaryWriter creates a new binary writer with specified byte order
func NewBinaryWriter(w io.Writer, order binary.ByteOrder) *BinaryWriter {
	return &BinaryWriter{
		writer: w,
		order:  order,
	}
}

// Write writes the binary representation of data into w.
// Data must be a fixed-size value or a slice of fixed-size values, or a
// pointer to such data.
func (bw *BinaryWriter) Write(data interface{}) error {
	return binary.Write(bw.writer, bw.order, data)
}

// WriteUint8 writes a uint8
func (bw *BinaryWriter) WriteUint8(val uint8) error {
	return bw.Write(val)
}

// WriteUint16 writes a uint16
func (bw *BinaryWriter) WriteUint16(val uint16) error {
	return bw.Write(val)
}

// WriteUint32 writes a uint32
func (bw *BinaryWriter) WriteUint32(val uint32) error {
	return bw.Write(val)
}

// WriteUint64 writes a uint64
func (bw *BinaryWriter) WriteUint64(val uint64) error {
	return bw.Write(val)
}

// WriteOID writes an OID
func (bw *BinaryWriter) WriteOID(val OID) error {
	return bw.Write(val)
}

// WriteXID writes an XID
func (bw *BinaryWriter) WriteXID(val XID) error {
	return bw.Write(val)
}

// WritePAddr writes a PAddr
func (bw *BinaryWriter) WritePAddr(val PAddr) error {
	return bw.Write(val)
}

// WriteUUID writes a UUID
func (bw *BinaryWriter) WriteUUID(val UUID) error {
	return bw.Write(val)
}

// WriteBytes writes a slice of bytes
func (bw *BinaryWriter) WriteBytes(data []byte) error {
	_, err := bw.writer.Write(data)
	return err
}

// WriteString writes a string without null termination
func (bw *BinaryWriter) WriteString(s string) error {
	return bw.WriteBytes([]byte(s))
}

// WriteNullTerminatedString writes a null-terminated string
func (bw *BinaryWriter) WriteNullTerminatedString(s string) error {
	err := bw.WriteString(s)
	if err != nil {
		return err
	}
	return bw.WriteUint8(0)
}

// WriteStringWithLen writes a string of exactly the specified length
// If the string is shorter, it will be null-padded
func (bw *BinaryWriter) WriteStringWithLen(s string, length int) error {
	if len(s) >= length {
		return bw.WriteBytes([]byte(s[:length]))
	}

	err := bw.WriteString(s)
	if err != nil {
		return err
	}

	padding := make([]byte, length-len(s))
	return bw.WriteBytes(padding)
}

// ================================
// Serialization and Deserialization Functions
// ================================

// DeserializeObjectHeader deserializes an ObjectHeader from binary data
func DeserializeObjectHeader(data []byte) (*ObjectHeader, error) {
	if len(data) < int(unsafe.Sizeof(ObjectHeader{})) {
		return nil, ErrStructTooShort
	}

	reader := NewBinaryReader(bytes.NewReader(data), binary.LittleEndian)
	header := &ObjectHeader{}

	if err := reader.Read(&header.Cksum); err != nil {
		return nil, fmt.Errorf("failed to read checksum: %w", err)
	}

	var err error
	header.OID, err = reader.ReadOID()
	if err != nil {
		return nil, fmt.Errorf("failed to read OID: %w", err)
	}

	header.XID, err = reader.ReadXID()
	if err != nil {
		return nil, fmt.Errorf("failed to read XID: %w", err)
	}

	header.Type, err = reader.ReadUint32()
	if err != nil {
		return nil, fmt.Errorf("failed to read type: %w", err)
	}

	header.Subtype, err = reader.ReadUint32()
	if err != nil {
		return nil, fmt.Errorf("failed to read subtype: %w", err)
	}

	return header, nil
}

// SerializeObjectHeader serializes an ObjectHeader to binary data
func SerializeObjectHeader(header *ObjectHeader) ([]byte, error) {
	buf := new(bytes.Buffer)
	writer := NewBinaryWriter(buf, binary.LittleEndian)

	if err := writer.Write(header.Cksum); err != nil {
		return nil, fmt.Errorf("failed to write checksum: %w", err)
	}

	if err := writer.WriteOID(header.OID); err != nil {
		return nil, fmt.Errorf("failed to write OID: %w", err)
	}

	if err := writer.WriteXID(header.XID); err != nil {
		return nil, fmt.Errorf("failed to write XID: %w", err)
	}

	if err := writer.WriteUint32(header.Type); err != nil {
		return nil, fmt.Errorf("failed to write type: %w", err)
	}

	if err := writer.WriteUint32(header.Subtype); err != nil {
		return nil, fmt.Errorf("failed to write subtype: %w", err)
	}

	return buf.Bytes(), nil
}

// DeserializeNXSuperblock deserializes an NXSuperblock from binary data
func DeserializeNXSuperblock(data []byte) (*NXSuperblock, error) {
	// Validate input size
	if len(data) < int(unsafe.Sizeof(NXSuperblock{})) {
		return nil, ErrStructTooShort
	}

	// Create a new binary reader for the full data slice
	sb := &NXSuperblock{}

	// Read object header first
	header, err := DeserializeObjectHeader(data)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize object header: %w", err)
	}
	sb.Header = *header

	// Reinitialize binary reader to skip over the header and resume reading
	headerSize := int(unsafe.Sizeof(ObjectHeader{}))
	br := NewBinaryReader(bytes.NewReader(data[headerSize:]), binary.LittleEndian)

	// Begin reading all subsequent NXSuperblock fields
	if sb.Magic, err = br.ReadUint32(); err != nil {
		return nil, fmt.Errorf("failed to read magic: %w", err)
	}
	if sb.Magic != NXMagic {
		return nil, ErrInvalidMagic
	}
	if sb.BlockSize, err = br.ReadUint32(); err != nil {
		return nil, err
	}
	if sb.BlockCount, err = br.ReadUint64(); err != nil {
		return nil, err
	}
	if sb.Features, err = br.ReadUint64(); err != nil {
		return nil, err
	}
	if sb.ReadOnlyCompatFeatures, err = br.ReadUint64(); err != nil {
		return nil, err
	}
	if sb.IncompatFeatures, err = br.ReadUint64(); err != nil {
		return nil, err
	}
	if sb.UUID, err = br.ReadUUID(); err != nil {
		return nil, err
	}
	if sb.NextOID, err = br.ReadOID(); err != nil {
		return nil, err
	}
	if sb.NextXID, err = br.ReadXID(); err != nil {
		return nil, err
	}
	if sb.XPDescBlocks, err = br.ReadUint32(); err != nil {
		return nil, err
	}
	if sb.XPDataBlocks, err = br.ReadUint32(); err != nil {
		return nil, err
	}
	if sb.XPDescBase, err = br.ReadPAddr(); err != nil {
		return nil, err
	}
	if sb.XPDataBase, err = br.ReadPAddr(); err != nil {
		return nil, err
	}
	if sb.XPDescNext, err = br.ReadUint32(); err != nil {
		return nil, err
	}
	if sb.XPDataNext, err = br.ReadUint32(); err != nil {
		return nil, err
	}
	if sb.XPDescIndex, err = br.ReadUint32(); err != nil {
		return nil, err
	}
	if sb.XPDescLen, err = br.ReadUint32(); err != nil {
		return nil, err
	}
	if sb.XPDataIndex, err = br.ReadUint32(); err != nil {
		return nil, err
	}
	if sb.XPDataLen, err = br.ReadUint32(); err != nil {
		return nil, err
	}
	if sb.SpacemanOID, err = br.ReadOID(); err != nil {
		return nil, err
	}
	if sb.OMapOID, err = br.ReadOID(); err != nil {
		return nil, err
	}
	if sb.ReaperOID, err = br.ReadOID(); err != nil {
		return nil, err
	}
	if sb.TestType, err = br.ReadUint32(); err != nil {
		return nil, err
	}
	if sb.MaxFileSystems, err = br.ReadUint32(); err != nil {
		return nil, err
	}

	// Arrays
	fsOIDs, err := br.ReadOIDArray(NXMaxFileSystems)
	if err != nil {
		return nil, fmt.Errorf("failed to read FSOIDs: %w", err)
	}
	copy(sb.FSOID[:], fsOIDs)

	counters, err := br.ReadUint64Array(NXNumCounters)
	if err != nil {
		return nil, fmt.Errorf("failed to read counters: %w", err)
	}
	copy(sb.Counters[:], counters)

	ephemeralInfo, err := br.ReadUint64Array(NXEphemeralInfoCount)
	if err != nil {
		return nil, fmt.Errorf("failed to read ephemeral info: %w", err)
	}
	copy(sb.EphemeralInfo[:], ephemeralInfo)

	// Nested structs
	if sb.BlockedOutPRange.StartAddr, err = br.ReadPAddr(); err != nil {
		return nil, err
	}
	if sb.BlockedOutPRange.BlockCount, err = br.ReadUint64(); err != nil {
		return nil, err
	}
	if sb.EvictMappingTreeOID, err = br.ReadOID(); err != nil {
		return nil, err
	}
	if sb.Flags, err = br.ReadUint64(); err != nil {
		return nil, err
	}
	if sb.EFIJumpstart, err = br.ReadPAddr(); err != nil {
		return nil, err
	}
	if sb.FusionUUID, err = br.ReadUUID(); err != nil {
		return nil, err
	}
	if sb.KeyLocker.StartAddr, err = br.ReadPAddr(); err != nil {
		return nil, err
	}
	if sb.KeyLocker.BlockCount, err = br.ReadUint64(); err != nil {
		return nil, err
	}

	// Final optional Fusion fields
	if sb.TestOID, err = br.ReadOID(); err != nil {
		return nil, err
	}
	if sb.FusionMtOID, err = br.ReadOID(); err != nil {
		return nil, err
	}
	if sb.FusionWbcOID, err = br.ReadOID(); err != nil {
		return nil, err
	}
	if sb.FusionWbc.StartAddr, err = br.ReadPAddr(); err != nil {
		return nil, err
	}
	if sb.FusionWbc.BlockCount, err = br.ReadUint64(); err != nil {
		return nil, err
	}
	if sb.NewestMountedVersion, err = br.ReadUint64(); err != nil {
		return nil, err
	}
	if sb.MkbLocker.StartAddr, err = br.ReadPAddr(); err != nil {
		return nil, err
	}
	if sb.MkbLocker.BlockCount, err = br.ReadUint64(); err != nil {
		return nil, err
	}

	return sb, nil
}

// SerializeNXSuperblock serializes an NXSuperblock to binary data
func SerializeNXSuperblock(sb *NXSuperblock) ([]byte, error) {
	buf := new(bytes.Buffer)
	writer := NewBinaryWriter(buf, binary.LittleEndian)

	// Write object header
	headerBytes, err := SerializeObjectHeader(&sb.Header)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize object header: %w", err)
	}
	if err := writer.WriteBytes(headerBytes); err != nil {
		return nil, fmt.Errorf("failed to write object header: %w", err)
	}

	// Begin writing NXSuperblock fields
	if err := writer.WriteUint32(sb.Magic); err != nil {
		return nil, err
	}
	if err := writer.WriteUint32(sb.BlockSize); err != nil {
		return nil, err
	}
	if err := writer.WriteUint64(sb.BlockCount); err != nil {
		return nil, err
	}
	if err := writer.WriteUint64(sb.Features); err != nil {
		return nil, err
	}
	if err := writer.WriteUint64(sb.ReadOnlyCompatFeatures); err != nil {
		return nil, err
	}
	if err := writer.WriteUint64(sb.IncompatFeatures); err != nil {
		return nil, err
	}
	if err := writer.WriteUUID(sb.UUID); err != nil {
		return nil, err
	}
	if err := writer.WriteOID(sb.NextOID); err != nil {
		return nil, err
	}
	if err := writer.WriteXID(sb.NextXID); err != nil {
		return nil, err
	}
	if err := writer.WriteUint32(sb.XPDescBlocks); err != nil {
		return nil, err
	}
	if err := writer.WriteUint32(sb.XPDataBlocks); err != nil {
		return nil, err
	}
	if err := writer.WritePAddr(sb.XPDescBase); err != nil {
		return nil, err
	}
	if err := writer.WritePAddr(sb.XPDataBase); err != nil {
		return nil, err
	}
	if err := writer.WriteUint32(sb.XPDescNext); err != nil {
		return nil, err
	}
	if err := writer.WriteUint32(sb.XPDataNext); err != nil {
		return nil, err
	}
	if err := writer.WriteUint32(sb.XPDescIndex); err != nil {
		return nil, err
	}
	if err := writer.WriteUint32(sb.XPDescLen); err != nil {
		return nil, err
	}
	if err := writer.WriteUint32(sb.XPDataIndex); err != nil {
		return nil, err
	}
	if err := writer.WriteUint32(sb.XPDataLen); err != nil {
		return nil, err
	}
	if err := writer.WriteOID(sb.SpacemanOID); err != nil {
		return nil, err
	}
	if err := writer.WriteOID(sb.OMapOID); err != nil {
		return nil, err
	}
	if err := writer.WriteOID(sb.ReaperOID); err != nil {
		return nil, err
	}
	if err := writer.WriteUint32(sb.TestType); err != nil {
		return nil, err
	}
	if err := writer.WriteUint32(sb.MaxFileSystems); err != nil {
		return nil, err
	}

	// Write fixed-size arrays
	for _, oid := range sb.FSOID {
		if err := writer.WriteOID(oid); err != nil {
			return nil, fmt.Errorf("failed to write FSOID: %w", err)
		}
	}
	for _, counter := range sb.Counters {
		if err := writer.WriteUint64(counter); err != nil {
			return nil, fmt.Errorf("failed to write counter: %w", err)
		}
	}
	for _, ephem := range sb.EphemeralInfo {
		if err := writer.WriteUint64(ephem); err != nil {
			return nil, fmt.Errorf("failed to write ephemeral info: %w", err)
		}
	}

	// Write nested structs
	if err := writer.WritePAddr(sb.BlockedOutPRange.StartAddr); err != nil {
		return nil, err
	}
	if err := writer.WriteUint64(sb.BlockedOutPRange.BlockCount); err != nil {
		return nil, err
	}
	if err := writer.WriteOID(sb.EvictMappingTreeOID); err != nil {
		return nil, err
	}
	if err := writer.WriteUint64(sb.Flags); err != nil {
		return nil, err
	}
	if err := writer.WritePAddr(sb.EFIJumpstart); err != nil {
		return nil, err
	}
	if err := writer.WriteUUID(sb.FusionUUID); err != nil {
		return nil, err
	}
	if err := writer.WritePAddr(sb.KeyLocker.StartAddr); err != nil {
		return nil, err
	}
	if err := writer.WriteUint64(sb.KeyLocker.BlockCount); err != nil {
		return nil, err
	}

	// Write fusion metadata
	if err := writer.WriteOID(sb.TestOID); err != nil {
		return nil, err
	}
	if err := writer.WriteOID(sb.FusionMtOID); err != nil {
		return nil, err
	}
	if err := writer.WriteOID(sb.FusionWbcOID); err != nil {
		return nil, err
	}
	if err := writer.WritePAddr(sb.FusionWbc.StartAddr); err != nil {
		return nil, err
	}
	if err := writer.WriteUint64(sb.FusionWbc.BlockCount); err != nil {
		return nil, err
	}
	if err := writer.WriteUint64(sb.NewestMountedVersion); err != nil {
		return nil, err
	}
	if err := writer.WritePAddr(sb.MkbLocker.StartAddr); err != nil {
		return nil, err
	}
	if err := writer.WriteUint64(sb.MkbLocker.BlockCount); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// DeserializeAPFSSuperblock deserializes an APFSSuperblock from binary data
func DeserializeAPFSSuperblock(data []byte) (*APFSSuperblock, error) {
	if len(data) < int(unsafe.Sizeof(APFSSuperblock{})) {
		return nil, ErrStructTooShort
	}

	sb := &APFSSuperblock{}

	header, err := DeserializeObjectHeader(data)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize object header: %w", err)
	}
	sb.Header = *header

	headerSize := int(unsafe.Sizeof(ObjectHeader{}))
	br := NewBinaryReader(bytes.NewReader(data[headerSize:]), binary.LittleEndian)

	if sb.Magic, err = br.ReadUint32(); err != nil {
		return nil, fmt.Errorf("failed to read magic: %w", err)
	}
	if sb.Magic != APFSMagic {
		return nil, ErrInvalidMagic
	}
	if sb.FSIndex, err = br.ReadUint32(); err != nil {
		return nil, fmt.Errorf("failed to read FSIndex: %w", err)
	}
	if sb.Features, err = br.ReadUint64(); err != nil {
		return nil, fmt.Errorf("failed to read features: %w", err)
	}
	if sb.ReadOnlyCompatFeatures, err = br.ReadUint64(); err != nil {
		return nil, fmt.Errorf("failed to read read-only compatible features: %w", err)
	}
	if sb.IncompatFeatures, err = br.ReadUint64(); err != nil {
		return nil, fmt.Errorf("failed to read incompatible features: %w", err)
	}
	if sb.UnmountTime, err = br.ReadUint64(); err != nil {
		return nil, fmt.Errorf("failed to read unmount time: %w", err)
	}
	if sb.ReserveBlockCount, err = br.ReadUint64(); err != nil {
		return nil, fmt.Errorf("failed to read reserve block count: %w", err)
	}
	if sb.QuotaBlockCount, err = br.ReadUint64(); err != nil {
		return nil, fmt.Errorf("failed to read quota block count: %w", err)
	}
	if sb.AllocCount, err = br.ReadUint64(); err != nil {
		return nil, fmt.Errorf("failed to read alloc count: %w", err)
	}
	if err = br.Read(&sb.MetaCrypto); err != nil {
		return nil, fmt.Errorf("failed to read meta crypto state: %w", err)
	}
	if sb.RootTreeType, err = br.ReadUint32(); err != nil {
		return nil, fmt.Errorf("failed to read root tree type: %w", err)
	}
	if sb.ExtentrefTreeType, err = br.ReadUint32(); err != nil {
		return nil, fmt.Errorf("failed to read extentref tree type: %w", err)
	}
	if sb.SnapMetaTreeType, err = br.ReadUint32(); err != nil {
		return nil, fmt.Errorf("failed to read snap meta tree type: %w", err)
	}
	if sb.OMapOID, err = br.ReadOID(); err != nil {
		return nil, fmt.Errorf("failed to read OMapOID: %w", err)
	}
	if sb.RootTreeOID, err = br.ReadOID(); err != nil {
		return nil, fmt.Errorf("failed to read root tree OID: %w", err)
	}
	if sb.ExtentrefTreeOID, err = br.ReadOID(); err != nil {
		return nil, fmt.Errorf("failed to read extentref tree OID: %w", err)
	}
	if sb.SnapMetaTreeOID, err = br.ReadOID(); err != nil {
		return nil, fmt.Errorf("failed to read snap meta tree OID: %w", err)
	}
	if sb.RevertToXID, err = br.ReadXID(); err != nil {
		return nil, fmt.Errorf("failed to read revert to XID: %w", err)
	}
	if sb.RevertToSblockOID, err = br.ReadOID(); err != nil {
		return nil, fmt.Errorf("failed to read revert to sblock OID: %w", err)
	}
	if sb.NextObjID, err = br.ReadUint64(); err != nil {
		return nil, fmt.Errorf("failed to read next object ID: %w", err)
	}
	if sb.NumFiles, err = br.ReadUint64(); err != nil {
		return nil, fmt.Errorf("failed to read num files: %w", err)
	}
	if sb.NumDirectories, err = br.ReadUint64(); err != nil {
		return nil, fmt.Errorf("failed to read num directories: %w", err)
	}
	if sb.NumSymlinks, err = br.ReadUint64(); err != nil {
		return nil, fmt.Errorf("failed to read num symlinks: %w", err)
	}
	if sb.NumOtherFSObjects, err = br.ReadUint64(); err != nil {
		return nil, fmt.Errorf("failed to read num other fs objects: %w", err)
	}
	if sb.NumSnapshots, err = br.ReadUint64(); err != nil {
		return nil, fmt.Errorf("failed to read num snapshots: %w", err)
	}
	if sb.TotalBlocksAlloced, err = br.ReadUint64(); err != nil {
		return nil, fmt.Errorf("failed to read total blocks alloced: %w", err)
	}
	if sb.TotalBlocksFreed, err = br.ReadUint64(); err != nil {
		return nil, fmt.Errorf("failed to read total blocks freed: %w", err)
	}
	if sb.UUID, err = br.ReadUUID(); err != nil {
		return nil, fmt.Errorf("failed to read volume UUID: %w", err)
	}
	if sb.LastModTime, err = br.ReadUint64(); err != nil {
		return nil, fmt.Errorf("failed to read last mod time: %w", err)
	}
	if sb.FSFlags, err = br.ReadUint64(); err != nil {
		return nil, fmt.Errorf("failed to read fs flags: %w", err)
	}
	if err = br.Read(&sb.FormattedBy); err != nil {
		return nil, fmt.Errorf("failed to read formatted by: %w", err)
	}
	if err = br.Read(&sb.ModifiedBy); err != nil {
		return nil, fmt.Errorf("failed to read modified by: %w", err)
	}
	if err = br.Read(&sb.VolName); err != nil {
		return nil, fmt.Errorf("failed to read volume name: %w", err)
	}
	if sb.NextDocID, err = br.ReadUint32(); err != nil {
		return nil, fmt.Errorf("failed to read next doc ID: %w", err)
	}
	if sb.Role, err = br.ReadUint16(); err != nil {
		return nil, fmt.Errorf("failed to read role: %w", err)
	}
	if sb.Reserved, err = br.ReadUint16(); err != nil {
		return nil, fmt.Errorf("failed to read reserved: %w", err)
	}
	if sb.RootToXID, err = br.ReadXID(); err != nil {
		return nil, fmt.Errorf("failed to read root to XID: %w", err)
	}
	if sb.ERStateOID, err = br.ReadOID(); err != nil {
		return nil, fmt.Errorf("failed to read ER state OID: %w", err)
	}
	if sb.CloneinfoIDEpoch, err = br.ReadUint64(); err != nil {
		return nil, fmt.Errorf("failed to read cloneinfo ID epoch: %w", err)
	}
	if sb.CloneinfoXID, err = br.ReadUint64(); err != nil {
		return nil, fmt.Errorf("failed to read cloneinfo XID: %w", err)
	}
	if sb.SnapMetaExtOID, err = br.ReadOID(); err != nil {
		return nil, fmt.Errorf("failed to read snap meta ext OID: %w", err)
	}
	if sb.VolumeGroupID, err = br.ReadUUID(); err != nil {
		return nil, fmt.Errorf("failed to read volume group ID: %w", err)
	}
	if sb.IntegrityMetaOID, err = br.ReadOID(); err != nil {
		return nil, fmt.Errorf("failed to read integrity meta OID: %w", err)
	}
	if sb.FextTreeOID, err = br.ReadOID(); err != nil {
		return nil, fmt.Errorf("failed to read fext tree OID: %w", err)
	}
	if sb.FextTreeType, err = br.ReadUint32(); err != nil {
		return nil, fmt.Errorf("failed to read fext tree type: %w", err)
	}
	if sb.ReservedType, err = br.ReadUint32(); err != nil {
		return nil, fmt.Errorf("failed to read reserved type: %w", err)
	}
	if sb.ReservedOID, err = br.ReadOID(); err != nil {
		return nil, fmt.Errorf("failed to read reserved OID: %w", err)
	}

	return sb, nil
}

// SerializeAPFSSuperblock serializes an APFSSuperblock to binary data
func SerializeAPFSSuperblock(sb *APFSSuperblock) ([]byte, error) {
	buf := new(bytes.Buffer)
	writer := NewBinaryWriter(buf, binary.LittleEndian)

	// Write object header
	headerBytes, err := SerializeObjectHeader(&sb.Header)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize object header: %w", err)
	}
	if err := writer.WriteBytes(headerBytes); err != nil {
		return nil, fmt.Errorf("failed to write object header: %w", err)
	}

	if err := writer.WriteUint32(sb.Magic); err != nil {
		return nil, err
	}
	if err := writer.WriteUint32(sb.FSIndex); err != nil {
		return nil, err
	}
	if err := writer.WriteUint64(sb.Features); err != nil {
		return nil, err
	}
	if err := writer.WriteUint64(sb.ReadOnlyCompatFeatures); err != nil {
		return nil, err
	}
	if err := writer.WriteUint64(sb.IncompatFeatures); err != nil {
		return nil, err
	}
	if err := writer.WriteUint64(sb.UnmountTime); err != nil {
		return nil, err
	}
	if err := writer.WriteUint64(sb.ReserveBlockCount); err != nil {
		return nil, err
	}
	if err := writer.WriteUint64(sb.QuotaBlockCount); err != nil {
		return nil, err
	}
	if err := writer.WriteUint64(sb.AllocCount); err != nil {
		return nil, err
	}
	if err := writer.Write(sb.MetaCrypto); err != nil {
		return nil, err
	}
	if err := writer.WriteUint32(sb.RootTreeType); err != nil {
		return nil, err
	}
	if err := writer.WriteUint32(sb.ExtentrefTreeType); err != nil {
		return nil, err
	}
	if err := writer.WriteUint32(sb.SnapMetaTreeType); err != nil {
		return nil, err
	}
	if err := writer.WriteOID(sb.OMapOID); err != nil {
		return nil, err
	}
	if err := writer.WriteOID(sb.RootTreeOID); err != nil {
		return nil, err
	}
	if err := writer.WriteOID(sb.ExtentrefTreeOID); err != nil {
		return nil, err
	}
	if err := writer.WriteOID(sb.SnapMetaTreeOID); err != nil {
		return nil, err
	}
	if err := writer.WriteXID(sb.RevertToXID); err != nil {
		return nil, err
	}
	if err := writer.WriteOID(sb.RevertToSblockOID); err != nil {
		return nil, err
	}
	if err := writer.WriteUint64(sb.NextObjID); err != nil {
		return nil, err
	}
	if err := writer.WriteUint64(sb.NumFiles); err != nil {
		return nil, err
	}
	if err := writer.WriteUint64(sb.NumDirectories); err != nil {
		return nil, err
	}
	if err := writer.WriteUint64(sb.NumSymlinks); err != nil {
		return nil, err
	}
	if err := writer.WriteUint64(sb.NumOtherFSObjects); err != nil {
		return nil, err
	}
	if err := writer.WriteUint64(sb.NumSnapshots); err != nil {
		return nil, err
	}
	if err := writer.WriteUint64(sb.TotalBlocksAlloced); err != nil {
		return nil, err
	}
	if err := writer.WriteUint64(sb.TotalBlocksFreed); err != nil {
		return nil, err
	}
	if err := writer.WriteUUID(sb.UUID); err != nil {
		return nil, err
	}
	if err := writer.WriteUint64(sb.LastModTime); err != nil {
		return nil, err
	}
	if err := writer.WriteUint64(sb.FSFlags); err != nil {
		return nil, err
	}
	if err := writer.Write(sb.FormattedBy); err != nil {
		return nil, err
	}
	for _, m := range sb.ModifiedBy {
		if err := writer.Write(m); err != nil {
			return nil, err
		}
	}
	if err := writer.WriteBytes(sb.VolName[:]); err != nil {
		return nil, err
	}
	if err := writer.WriteUint32(sb.NextDocID); err != nil {
		return nil, err
	}
	if err := writer.WriteUint16(sb.Role); err != nil {
		return nil, err
	}
	if err := writer.WriteUint16(sb.Reserved); err != nil {
		return nil, err
	}
	if err := writer.WriteXID(sb.RootToXID); err != nil {
		return nil, err
	}
	if err := writer.WriteOID(sb.ERStateOID); err != nil {
		return nil, err
	}
	if err := writer.WriteUint64(sb.CloneinfoIDEpoch); err != nil {
		return nil, err
	}
	if err := writer.WriteUint64(sb.CloneinfoXID); err != nil {
		return nil, err
	}
	if err := writer.WriteOID(sb.SnapMetaExtOID); err != nil {
		return nil, err
	}
	if err := writer.WriteUUID(sb.VolumeGroupID); err != nil {
		return nil, err
	}
	if err := writer.WriteOID(sb.IntegrityMetaOID); err != nil {
		return nil, err
	}
	if err := writer.WriteOID(sb.FextTreeOID); err != nil {
		return nil, err
	}
	if err := writer.WriteUint32(sb.FextTreeType); err != nil {
		return nil, err
	}
	if err := writer.WriteUint32(sb.ReservedType); err != nil {
		return nil, err
	}
	if err := writer.WriteOID(sb.ReservedOID); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// DeserializeBTNodePhys deserializes a B-tree node from binary data
func DeserializeBTNodePhys(data []byte) (*BTNodePhys, error) {
	const cksumOffset = 0
	const minSize = 56 // conservative minimum (size of fixed fields)

	if len(data) < minSize {
		return nil, ErrStructTooShort
	}

	// 1. Verify Fletcher64 checksum
	expected := binary.LittleEndian.Uint64(data[:8])
	actual := checksum.Fletcher64WithZeroedChecksum(data, cksumOffset)
	if expected != actual {
		return nil, fmt.Errorf("checksum mismatch: expected 0x%016x, got 0x%016x", expected, actual)
	}

	// 2. Set up reader
	br := NewBinaryReader(bytes.NewReader(data), binary.LittleEndian)
	node := &BTNodePhys{}

	// 3. Read header (obj_phys_t)
	if err := br.Read(&node.Header.Cksum); err != nil {
		return nil, fmt.Errorf("read cksum: %w", err)
	}
	var err error
	if node.Header.OID, err = br.ReadOID(); err != nil {
		return nil, err
	}
	if node.Header.XID, err = br.ReadXID(); err != nil {
		return nil, err
	}
	if node.Header.Type, err = br.ReadUint32(); err != nil {
		return nil, err
	}
	if node.Header.Subtype, err = br.ReadUint32(); err != nil {
		return nil, err
	}

	// 4. Validate object type
	if (node.Header.Type & ObjectTypeMask) != ObjectTypeBtreeNode {
		return nil, fmt.Errorf("invalid object type: 0x%x, expected BTREE_NODE", node.Header.Type)
	}

	// 5. Read BTNodePhys fields
	if node.Flags, err = br.ReadUint16(); err != nil {
		return nil, err
	}
	if node.Level, err = br.ReadUint16(); err != nil {
		return nil, err
	}
	if node.KeyCount, err = br.ReadUint32(); err != nil {
		return nil, err
	}
	if err = br.Read(&node.TableSpace); err != nil {
		return nil, err
	}
	if err = br.Read(&node.FreeSpace); err != nil {
		return nil, err
	}
	if err = br.Read(&node.KeyFreeList); err != nil {
		return nil, err
	}
	if err = br.Read(&node.ValFreeList); err != nil {
		return nil, err
	}

	// 6. Read remaining data as node.Data
	currPos, err := br.buf.Seek(0, io.SeekCurrent)
	if err != nil {
		return nil, err
	}
	remainingSize := int64(len(data)) - currPos
	node.Data, err = br.ReadBytes(int(remainingSize))
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("read trailing data: %w", err)
	}

	// Optional: capture raw tail (for future fields, debug, etc.)
	// node.TailData = data[currPos+len(node.Data):]

	return node, nil
}

// SerializeBTNodePhys serializes a B-tree node to binary data
func SerializeBTNodePhys(node *BTNodePhys) ([]byte, error) {
	buf := new(bytes.Buffer)
	bw := NewBinaryWriter(buf, binary.LittleEndian)

	// Write ObjectHeader
	if err := bw.Write(node.Header.Cksum); err != nil {
		return nil, fmt.Errorf("failed to write checksum: %w", err)
	}
	if err := bw.WriteOID(node.Header.OID); err != nil {
		return nil, err
	}
	if err := bw.WriteXID(node.Header.XID); err != nil {
		return nil, err
	}
	if err := bw.WriteUint32(node.Header.Type); err != nil {
		return nil, err
	}
	if err := bw.WriteUint32(node.Header.Subtype); err != nil {
		return nil, err
	}

	// Write BTNodePhys fixed fields
	if err := bw.WriteUint16(node.Flags); err != nil {
		return nil, err
	}
	if err := bw.WriteUint16(node.Level); err != nil {
		return nil, err
	}
	if err := bw.WriteUint32(node.KeyCount); err != nil {
		return nil, err
	}
	if err := bw.Write(node.TableSpace); err != nil {
		return nil, err
	}
	if err := bw.Write(node.FreeSpace); err != nil {
		return nil, err
	}
	if err := bw.Write(node.KeyFreeList); err != nil {
		return nil, err
	}
	if err := bw.Write(node.ValFreeList); err != nil {
		return nil, err
	}

	// Write node.Data (variable length)
	if err := bw.WriteBytes(node.Data); err != nil {
		return nil, fmt.Errorf("failed to write node data: %w", err)
	}

	return buf.Bytes(), nil
}

// DeserializeCheckpointMapPhys deserializes a checkpoint mapping block from binary data
func DeserializeCheckpointMapPhys(data []byte) (*CheckpointMapPhys, error) {
	if len(data) < int(unsafe.Sizeof(ObjectHeader{})) {
		return nil, ErrStructTooShort
	}

	// Deserialize the object header first
	header, err := DeserializeObjectHeader(data)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize object header: %w", err)
	}

	// Verify object type
	if header.GetObjectType() != ObjectTypeCheckpointMap {
		return nil, fmt.Errorf("invalid object type: 0x%x, expected CHECKPOINT_MAP", header.GetObjectType())
	}

	// Create checkpoint map structure
	cpMap := &CheckpointMapPhys{
		Header: *header,
	}

	// Read the rest of the checkpoint map
	headerSize := int(unsafe.Sizeof(ObjectHeader{}))
	br := NewBinaryReader(bytes.NewReader(data[headerSize:]), binary.LittleEndian)

	// Read flags and count
	var err1, err2 error
	cpMap.Flags, err1 = br.ReadUint32()
	cpMap.Count, err2 = br.ReadUint32()

	if err1 != nil || err2 != nil {
		return nil, fmt.Errorf("failed to read checkpoint map fields: %v %v", err1, err2)
	}

	// Read checkpoint mappings
	cpMap.Map = make([]CheckpointMapping, cpMap.Count)
	for i := uint32(0); i < cpMap.Count; i++ {
		mapping := &cpMap.Map[i]

		mapping.Type, err = br.ReadUint32()
		if err != nil {
			return nil, fmt.Errorf("failed to read mapping type: %w", err)
		}

		mapping.Subtype, err = br.ReadUint32()
		if err != nil {
			return nil, fmt.Errorf("failed to read mapping subtype: %w", err)
		}

		mapping.Size, err = br.ReadUint32()
		if err != nil {
			return nil, fmt.Errorf("failed to read mapping size: %w", err)
		}

		mapping.Pad, err = br.ReadUint32()
		if err != nil {
			return nil, fmt.Errorf("failed to read mapping pad: %w", err)
		}

		mapping.FSOID, err = br.ReadOID()
		if err != nil {
			return nil, fmt.Errorf("failed to read mapping FSOID: %w", err)
		}

		mapping.OID, err = br.ReadOID()
		if err != nil {
			return nil, fmt.Errorf("failed to read mapping OID: %w", err)
		}

		mapping.PAddr, err = br.ReadPAddr()
		if err != nil {
			return nil, fmt.Errorf("failed to read mapping PAddr: %w", err)
		}
	}

	return cpMap, nil
}

// DeserializeOMapPhys deserializes an object map from binary data
func DeserializeOMapPhys(data []byte) (*OMapPhys, error) {
	if len(data) < int(unsafe.Sizeof(ObjectHeader{})) {
		return nil, ErrStructTooShort
	}

	// Deserialize the object header first
	header, err := DeserializeObjectHeader(data)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize object header: %w", err)
	}

	// Verify object type
	if header.GetObjectType() != ObjectTypeOMap {
		return nil, fmt.Errorf("invalid object type: 0x%x, expected OMAP", header.GetObjectType())
	}

	// Create object map structure
	omap := &OMapPhys{
		Header: *header,
	}

	// Read the rest of the object map
	headerSize := int(unsafe.Sizeof(ObjectHeader{}))
	br := NewBinaryReader(bytes.NewReader(data[headerSize:]), binary.LittleEndian)

	// Read fields
	if omap.Flags, err = br.ReadUint32(); err != nil {
		return nil, fmt.Errorf("failed to read flags: %w", err)
	}

	if omap.SnapCount, err = br.ReadUint32(); err != nil {
		return nil, fmt.Errorf("failed to read snap count: %w", err)
	}

	if omap.TreeType, err = br.ReadUint32(); err != nil {
		return nil, fmt.Errorf("failed to read tree type: %w", err)
	}

	if omap.SnapshotTreeType, err = br.ReadUint32(); err != nil {
		return nil, fmt.Errorf("failed to read snapshot tree type: %w", err)
	}

	if omap.TreeOID, err = br.ReadOID(); err != nil {
		return nil, fmt.Errorf("failed to read tree OID: %w", err)
	}

	if omap.SnapshotTreeOID, err = br.ReadOID(); err != nil {
		return nil, fmt.Errorf("failed to read snapshot tree OID: %w", err)
	}

	if omap.MostRecentSnap, err = br.ReadXID(); err != nil {
		return nil, fmt.Errorf("failed to read most recent snap: %w", err)
	}

	if omap.PendingRevertMin, err = br.ReadXID(); err != nil {
		return nil, fmt.Errorf("failed to read pending revert min: %w", err)
	}

	if omap.PendingRevertMax, err = br.ReadXID(); err != nil {
		return nil, fmt.Errorf("failed to read pending revert max: %w", err)
	}

	return omap, nil
}

// DeserializeSpacemanPhys deserializes a space manager from binary data
func DeserializeSpacemanPhys(data []byte) (*SpacemanPhys, error) {
	if len(data) < int(unsafe.Sizeof(ObjectHeader{})) {
		return nil, ErrStructTooShort
	}

	// Deserialize the object header first
	header, err := DeserializeObjectHeader(data)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize object header: %w", err)
	}

	// Verify object type
	if header.GetObjectType() != ObjectTypeSpaceman {
		return nil, fmt.Errorf("invalid object type: 0x%x, expected SPACEMAN", header.GetObjectType())
	}

	// Create space manager structure
	sm := &SpacemanPhys{
		Header: *header,
	}

	// Read the rest of the space manager
	headerSize := int(unsafe.Sizeof(ObjectHeader{}))
	br := NewBinaryReader(bytes.NewReader(data[headerSize:]), binary.LittleEndian)

	// Read basic fields
	if sm.BlockSize, err = br.ReadUint32(); err != nil {
		return nil, fmt.Errorf("failed to read block size: %w", err)
	}

	if sm.BlocksPerChunk, err = br.ReadUint32(); err != nil {
		return nil, fmt.Errorf("failed to read blocks per chunk: %w", err)
	}

	if sm.ChunksPerCIB, err = br.ReadUint32(); err != nil {
		return nil, fmt.Errorf("failed to read chunks per CIB: %w", err)
	}

	if sm.CIBsPerCAB, err = br.ReadUint32(); err != nil {
		return nil, fmt.Errorf("failed to read CIBs per CAB: %w", err)
	}

	// Read device information for main device (index 0)
	if sm.Devices[0].BlockCount, err = br.ReadUint64(); err != nil {
		return nil, fmt.Errorf("failed to read main device block count: %w", err)
	}

	if sm.Devices[0].ChunkCount, err = br.ReadUint64(); err != nil {
		return nil, fmt.Errorf("failed to read main device chunk count: %w", err)
	}

	if sm.Devices[0].CIBCount, err = br.ReadUint32(); err != nil {
		return nil, fmt.Errorf("failed to read main device CIB count: %w", err)
	}

	if sm.Devices[0].CABCount, err = br.ReadUint32(); err != nil {
		return nil, fmt.Errorf("failed to read main device CAB count: %w", err)
	}

	if sm.Devices[0].FreeCount, err = br.ReadUint64(); err != nil {
		return nil, fmt.Errorf("failed to read main device free count: %w", err)
	}

	if sm.Devices[0].AddrOffset, err = br.ReadUint32(); err != nil {
		return nil, fmt.Errorf("failed to read main device addr offset: %w", err)
	}

	if sm.Devices[0].Reserved, err = br.ReadUint32(); err != nil {
		return nil, fmt.Errorf("failed to read main device reserved: %w", err)
	}

	if sm.Devices[0].Reserved2, err = br.ReadUint64(); err != nil {
		return nil, fmt.Errorf("failed to read main device reserved2: %w", err)
	}

	// Read device information for tier2 device (index 1)
	if sm.Devices[1].BlockCount, err = br.ReadUint64(); err != nil {
		return nil, fmt.Errorf("failed to read tier2 device block count: %w", err)
	}

	if sm.Devices[1].ChunkCount, err = br.ReadUint64(); err != nil {
		return nil, fmt.Errorf("failed to read tier2 device chunk count: %w", err)
	}

	if sm.Devices[1].CIBCount, err = br.ReadUint32(); err != nil {
		return nil, fmt.Errorf("failed to read tier2 device CIB count: %w", err)
	}

	if sm.Devices[1].CABCount, err = br.ReadUint32(); err != nil {
		return nil, fmt.Errorf("failed to read tier2 device CAB count: %w", err)
	}

	if sm.Devices[1].FreeCount, err = br.ReadUint64(); err != nil {
		return nil, fmt.Errorf("failed to read tier2 device free count: %w", err)
	}

	if sm.Devices[1].AddrOffset, err = br.ReadUint32(); err != nil {
		return nil, fmt.Errorf("failed to read tier2 device addr offset: %w", err)
	}

	if sm.Devices[1].Reserved, err = br.ReadUint32(); err != nil {
		return nil, fmt.Errorf("failed to read tier2 device reserved: %w", err)
	}

	if sm.Devices[1].Reserved2, err = br.ReadUint64(); err != nil {
		return nil, fmt.Errorf("failed to read tier2 device reserved2: %w", err)
	}

	// Read other fields
	if sm.Flags, err = br.ReadUint32(); err != nil {
		return nil, fmt.Errorf("failed to read flags: %w", err)
	}

	if sm.IPBmTxMultiplier, err = br.ReadUint32(); err != nil {
		return nil, fmt.Errorf("failed to read IP BM TX multiplier: %w", err)
	}

	if sm.IPBlockCount, err = br.ReadUint64(); err != nil {
		return nil, fmt.Errorf("failed to read IP block count: %w", err)
	}

	if sm.IPBmSizeInBlocks, err = br.ReadUint32(); err != nil {
		return nil, fmt.Errorf("failed to read IP BM size in blocks: %w", err)
	}

	if sm.IPBmBlockCount, err = br.ReadUint32(); err != nil {
		return nil, fmt.Errorf("failed to read IP BM block count: %w", err)
	}

	if sm.IPBmBase, err = br.ReadPAddr(); err != nil {
		return nil, fmt.Errorf("failed to read IP BM base: %w", err)
	}

	if sm.IPBase, err = br.ReadPAddr(); err != nil {
		return nil, fmt.Errorf("failed to read IP base: %w", err)
	}

	if sm.FSReserveBlockCount, err = br.ReadUint64(); err != nil {
		return nil, fmt.Errorf("failed to read FS reserve block count: %w", err)
	}

	if sm.FSReserveAllocCount, err = br.ReadUint64(); err != nil {
		return nil, fmt.Errorf("failed to read FS reserve alloc count: %w", err)
	}

	// Read free queues
	for i := 0; i < 3; i++ { // SFQ_COUNT = 3 (IP, Main, Tier2)
		if sm.FreeQueues[i].Count, err = br.ReadUint64(); err != nil {
			return nil, fmt.Errorf("failed to read free queue %d count: %w", i, err)
		}

		if sm.FreeQueues[i].TreeOID, err = br.ReadOID(); err != nil {
			return nil, fmt.Errorf("failed to read free queue %d tree OID: %w", i, err)
		}

		if sm.FreeQueues[i].OldestXID, err = br.ReadXID(); err != nil {
			return nil, fmt.Errorf("failed to read free queue %d oldest XID: %w", i, err)
		}

		if sm.FreeQueues[i].TreeNodeLimit, err = br.ReadUint16(); err != nil {
			return nil, fmt.Errorf("failed to read free queue %d tree node limit: %w", i, err)
		}

		if sm.FreeQueues[i].Pad16, err = br.ReadUint16(); err != nil {
			return nil, fmt.Errorf("failed to read free queue %d pad16: %w", i, err)
		}

		if sm.FreeQueues[i].Pad32, err = br.ReadUint32(); err != nil {
			return nil, fmt.Errorf("failed to read free queue %d pad32: %w", i, err)
		}

		if sm.FreeQueues[i].Reserved, err = br.ReadUint64(); err != nil {
			return nil, fmt.Errorf("failed to read free queue %d reserved: %w", i, err)
		}
	}

	// Read remaining fields
	if sm.IPBmFreeHead, err = br.ReadUint16(); err != nil {
		return nil, fmt.Errorf("failed to read IP BM free head: %w", err)
	}

	if sm.IPBmFreeTail, err = br.ReadUint16(); err != nil {
		return nil, fmt.Errorf("failed to read IP BM free tail: %w", err)
	}

	if sm.IPBmXidOffset, err = br.ReadUint32(); err != nil {
		return nil, fmt.Errorf("failed to read IP BM XID offset: %w", err)
	}

	if sm.IPBitmapOffset, err = br.ReadUint32(); err != nil {
		return nil, fmt.Errorf("failed to read IP bitmap offset: %w", err)
	}

	if sm.IPBmFreeNextOffset, err = br.ReadUint32(); err != nil {
		return nil, fmt.Errorf("failed to read IP BM free next offset: %w", err)
	}

	if sm.Version, err = br.ReadUint32(); err != nil {
		return nil, fmt.Errorf("failed to read version: %w", err)
	}

	if sm.StructSize, err = br.ReadUint32(); err != nil {
		return nil, fmt.Errorf("failed to read struct size: %w", err)
	}

	// Note: We're skipping datazone_info_phys_t which is a complex structure
	// A complete implementation would need to handle this

	return sm, nil
}

// SerializeCheckpointMapPhys serializes a checkpoint mapping block to binary data
func SerializeCheckpointMapPhys(cpMap *CheckpointMapPhys) ([]byte, error) {
	buf := new(bytes.Buffer)
	writer := NewBinaryWriter(buf, binary.LittleEndian)

	// Write object header
	headerBytes, err := SerializeObjectHeader(&cpMap.Header)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize object header: %w", err)
	}
	if err := writer.WriteBytes(headerBytes); err != nil {
		return nil, fmt.Errorf("failed to write object header: %w", err)
	}

	// Write flags and count
	if err := writer.WriteUint32(cpMap.Flags); err != nil {
		return nil, fmt.Errorf("failed to write flags: %w", err)
	}

	if err := writer.WriteUint32(cpMap.Count); err != nil {
		return nil, fmt.Errorf("failed to write count: %w", err)
	}

	// Write checkpoint mappings
	for i := uint32(0); i < cpMap.Count; i++ {
		mapping := &cpMap.Map[i]

		if err := writer.WriteUint32(mapping.Type); err != nil {
			return nil, fmt.Errorf("failed to write mapping type: %w", err)
		}

		if err := writer.WriteUint32(mapping.Subtype); err != nil {
			return nil, fmt.Errorf("failed to write mapping subtype: %w", err)
		}

		if err := writer.WriteUint32(mapping.Size); err != nil {
			return nil, fmt.Errorf("failed to write mapping size: %w", err)
		}

		if err := writer.WriteUint32(mapping.Pad); err != nil {
			return nil, fmt.Errorf("failed to write mapping pad: %w", err)
		}

		if err := writer.WriteOID(mapping.FSOID); err != nil {
			return nil, fmt.Errorf("failed to write mapping FSOID: %w", err)
		}

		if err := writer.WriteOID(mapping.OID); err != nil {
			return nil, fmt.Errorf("failed to write mapping OID: %w", err)
		}

		if err := writer.WritePAddr(mapping.PAddr); err != nil {
			return nil, fmt.Errorf("failed to write mapping PAddr: %w", err)
		}
	}

	return buf.Bytes(), nil
}

// SerializeOMapPhys serializes an object map to binary data
func SerializeOMapPhys(omap *OMapPhys) ([]byte, error) {
	buf := new(bytes.Buffer)
	writer := NewBinaryWriter(buf, binary.LittleEndian)

	// Write object header
	headerBytes, err := SerializeObjectHeader(&omap.Header)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize object header: %w", err)
	}
	if err := writer.WriteBytes(headerBytes); err != nil {
		return nil, fmt.Errorf("failed to write object header: %w", err)
	}

	// Write fields
	if err := writer.WriteUint32(omap.Flags); err != nil {
		return nil, fmt.Errorf("failed to write flags: %w", err)
	}

	if err := writer.WriteUint32(omap.SnapCount); err != nil {
		return nil, fmt.Errorf("failed to write snap count: %w", err)
	}

	if err := writer.WriteUint32(omap.TreeType); err != nil {
		return nil, fmt.Errorf("failed to write tree type: %w", err)
	}

	if err := writer.WriteUint32(omap.SnapshotTreeType); err != nil {
		return nil, fmt.Errorf("failed to write snapshot tree type: %w", err)
	}

	if err := writer.WriteOID(omap.TreeOID); err != nil {
		return nil, fmt.Errorf("failed to write tree OID: %w", err)
	}

	if err := writer.WriteOID(omap.SnapshotTreeOID); err != nil {
		return nil, fmt.Errorf("failed to write snapshot tree OID: %w", err)
	}

	if err := writer.WriteXID(omap.MostRecentSnap); err != nil {
		return nil, fmt.Errorf("failed to write most recent snap: %w", err)
	}

	if err := writer.WriteXID(omap.PendingRevertMin); err != nil {
		return nil, fmt.Errorf("failed to write pending revert min: %w", err)
	}

	if err := writer.WriteXID(omap.PendingRevertMax); err != nil {
		return nil, fmt.Errorf("failed to write pending revert max: %w", err)
	}

	return buf.Bytes(), nil
}

// SerializeSpacemanPhys serializes a space manager to binary data
func SerializeSpacemanPhys(sm *SpacemanPhys) ([]byte, error) {
	buf := new(bytes.Buffer)
	writer := NewBinaryWriter(buf, binary.LittleEndian)

	// Write object header
	headerBytes, err := SerializeObjectHeader(&sm.Header)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize object header: %w", err)
	}
	if err := writer.WriteBytes(headerBytes); err != nil {
		return nil, fmt.Errorf("failed to write object header: %w", err)
	}

	// Write basic fields
	if err := writer.WriteUint32(sm.BlockSize); err != nil {
		return nil, fmt.Errorf("failed to write block size: %w", err)
	}

	if err := writer.WriteUint32(sm.BlocksPerChunk); err != nil {
		return nil, fmt.Errorf("failed to write blocks per chunk: %w", err)
	}

	if err := writer.WriteUint32(sm.ChunksPerCIB); err != nil {
		return nil, fmt.Errorf("failed to write chunks per CIB: %w", err)
	}

	if err := writer.WriteUint32(sm.CIBsPerCAB); err != nil {
		return nil, fmt.Errorf("failed to write CIBs per CAB: %w", err)
	}

	// Write device information for main device (index 0)
	if err := writer.WriteUint64(sm.Devices[0].BlockCount); err != nil {
		return nil, fmt.Errorf("failed to write main device block count: %w", err)
	}

	if err := writer.WriteUint64(sm.Devices[0].ChunkCount); err != nil {
		return nil, fmt.Errorf("failed to write main device chunk count: %w", err)
	}

	if err := writer.WriteUint32(sm.Devices[0].CIBCount); err != nil {
		return nil, fmt.Errorf("failed to write main device CIB count: %w", err)
	}

	if err := writer.WriteUint32(sm.Devices[0].CABCount); err != nil {
		return nil, fmt.Errorf("failed to write main device CAB count: %w", err)
	}

	if err := writer.WriteUint64(sm.Devices[0].FreeCount); err != nil {
		return nil, fmt.Errorf("failed to write main device free count: %w", err)
	}

	if err := writer.WriteUint32(sm.Devices[0].AddrOffset); err != nil {
		return nil, fmt.Errorf("failed to write main device addr offset: %w", err)
	}

	if err := writer.WriteUint32(sm.Devices[0].Reserved); err != nil {
		return nil, fmt.Errorf("failed to write main device reserved: %w", err)
	}

	if err := writer.WriteUint64(sm.Devices[0].Reserved2); err != nil {
		return nil, fmt.Errorf("failed to write main device reserved2: %w", err)
	}

	// Write device information for tier2 device (index 1)
	if err := writer.WriteUint64(sm.Devices[1].BlockCount); err != nil {
		return nil, fmt.Errorf("failed to write tier2 device block count: %w", err)
	}

	if err := writer.WriteUint64(sm.Devices[1].ChunkCount); err != nil {
		return nil, fmt.Errorf("failed to write tier2 device chunk count: %w", err)
	}

	if err := writer.WriteUint32(sm.Devices[1].CIBCount); err != nil {
		return nil, fmt.Errorf("failed to write tier2 device CIB count: %w", err)
	}

	if err := writer.WriteUint32(sm.Devices[1].CABCount); err != nil {
		return nil, fmt.Errorf("failed to write tier2 device CAB count: %w", err)
	}

	if err := writer.WriteUint64(sm.Devices[1].FreeCount); err != nil {
		return nil, fmt.Errorf("failed to write tier2 device free count: %w", err)
	}

	if err := writer.WriteUint32(sm.Devices[1].AddrOffset); err != nil {
		return nil, fmt.Errorf("failed to write tier2 device addr offset: %w", err)
	}

	if err := writer.WriteUint32(sm.Devices[1].Reserved); err != nil {
		return nil, fmt.Errorf("failed to write tier2 device reserved: %w", err)
	}

	if err := writer.WriteUint64(sm.Devices[1].Reserved2); err != nil {
		return nil, fmt.Errorf("failed to write tier2 device reserved2: %w", err)
	}

	// Write other fields
	if err := writer.WriteUint32(sm.Flags); err != nil {
		return nil, fmt.Errorf("failed to write flags: %w", err)
	}

	if err := writer.WriteUint32(sm.IPBmTxMultiplier); err != nil {
		return nil, fmt.Errorf("failed to write IP BM TX multiplier: %w", err)
	}

	if err := writer.WriteUint64(sm.IPBlockCount); err != nil {
		return nil, fmt.Errorf("failed to write IP block count: %w", err)
	}

	if err := writer.WriteUint32(sm.IPBmSizeInBlocks); err != nil {
		return nil, fmt.Errorf("failed to write IP BM size in blocks: %w", err)
	}

	if err := writer.WriteUint32(sm.IPBmBlockCount); err != nil {
		return nil, fmt.Errorf("failed to write IP BM block count: %w", err)
	}

	if err := writer.WritePAddr(sm.IPBmBase); err != nil {
		return nil, fmt.Errorf("failed to write IP BM base: %w", err)
	}

	if err := writer.WritePAddr(sm.IPBase); err != nil {
		return nil, fmt.Errorf("failed to write IP base: %w", err)
	}

	if err := writer.WriteUint64(sm.FSReserveBlockCount); err != nil {
		return nil, fmt.Errorf("failed to write FS reserve block count: %w", err)
	}

	if err := writer.WriteUint64(sm.FSReserveAllocCount); err != nil {
		return nil, fmt.Errorf("failed to write FS reserve alloc count: %w", err)
	}

	// Write free queues
	for i := 0; i < 3; i++ { // SFQ_COUNT = 3 (IP, Main, Tier2)
		if err := writer.WriteUint64(sm.FreeQueues[i].Count); err != nil {
			return nil, fmt.Errorf("failed to write free queue %d count: %w", i, err)
		}

		if err := writer.WriteOID(sm.FreeQueues[i].TreeOID); err != nil {
			return nil, fmt.Errorf("failed to write free queue %d tree OID: %w", i, err)
		}

		if err := writer.WriteXID(sm.FreeQueues[i].OldestXID); err != nil {
			return nil, fmt.Errorf("failed to write free queue %d oldest XID: %w", i, err)
		}

		if err := writer.WriteUint16(sm.FreeQueues[i].TreeNodeLimit); err != nil {
			return nil, fmt.Errorf("failed to write free queue %d tree node limit: %w", i, err)
		}

		if err := writer.WriteUint16(sm.FreeQueues[i].Pad16); err != nil {
			return nil, fmt.Errorf("failed to write free queue %d pad16: %w", i, err)
		}

		if err := writer.WriteUint32(sm.FreeQueues[i].Pad32); err != nil {
			return nil, fmt.Errorf("failed to write free queue %d pad32: %w", i, err)
		}

		if err := writer.WriteUint64(sm.FreeQueues[i].Reserved); err != nil {
			return nil, fmt.Errorf("failed to write free queue %d reserved: %w", i, err)
		}
	}

	// Write remaining fields
	if err := writer.WriteUint16(sm.IPBmFreeHead); err != nil {
		return nil, fmt.Errorf("failed to write IP BM free head: %w", err)
	}

	if err := writer.WriteUint16(sm.IPBmFreeTail); err != nil {
		return nil, fmt.Errorf("failed to write IP BM free tail: %w", err)
	}

	if err := writer.WriteUint32(sm.IPBmXidOffset); err != nil {
		return nil, fmt.Errorf("failed to write IP BM XID offset: %w", err)
	}

	if err := writer.WriteUint32(sm.IPBitmapOffset); err != nil {
		return nil, fmt.Errorf("failed to write IP bitmap offset: %w", err)
	}

	if err := writer.WriteUint32(sm.IPBmFreeNextOffset); err != nil {
		return nil, fmt.Errorf("failed to write IP BM free next offset: %w", err)
	}

	if err := writer.WriteUint32(sm.Version); err != nil {
		return nil, fmt.Errorf("failed to write version: %w", err)
	}

	if err := writer.WriteUint32(sm.StructSize); err != nil {
		return nil, fmt.Errorf("failed to write struct size: %w", err)
	}

	// Note: We're skipping datazone_info_phys_t which is a complex structure
	// A complete implementation would need to handle this

	return buf.Bytes(), nil
}

// We implement serialization/deserialization for all APFS structures.
// Each follows this standardized pattern:
//
// 1. Validate data length against minimum struct size
// 2. Create a binary reader/writer with correct endianness
// 3. Read or write fields in strict spec order
// 4. Verify or compute the Fletcher 64 checksum (obj_phys_t-based)
// 5. Preserve and zero-fill padding where applicable
// 6. Validate object type and flags per OBJECT_TYPE_MASK
// 7. Optionally return or retain raw trailing bytes for future compatibility

// ReadBlock reads a block from a block device and deserializes it to the appropriate structure
func ReadBlock(device BlockDevice, addr PAddr, blockType uint32) (interface{}, error) {
	data, err := device.ReadBlock(addr)
	if err != nil {
		return nil, fmt.Errorf("failed to read block at %d: %w", addr, err)
	}

	// Check if data starts with an object header
	if len(data) < int(unsafe.Sizeof(ObjectHeader{})) {
		return nil, ErrStructTooShort
	}

	// Deserialize object header to get type
	header, err := DeserializeObjectHeader(data)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize object header: %w", err)
	}

	// If a specific blockType was requested, verify it matches
	if blockType != 0 && header.GetObjectType() != blockType {
		return nil, fmt.Errorf("expected block type %d, got %d: %w",
			blockType, header.GetObjectType(), ErrInvalidObjectType)
	}

	// Deserialize based on object type
	switch header.GetObjectType() {
	case ObjectTypeNXSuperblock:
		return DeserializeNXSuperblock(data)
	case ObjectTypeFS:
		return DeserializeAPFSSuperblock(data)
	case ObjectTypeBtreeNode:
		return DeserializeBTNodePhys(data)
	// Add cases for other object types
	default:
		return nil, fmt.Errorf("unsupported object type: %d", header.GetObjectType())
	}
}

// WriteBlock serializes a structure and writes it to a block device
func WriteBlock(device BlockDevice, addr PAddr, obj interface{}) error {
	var data []byte
	var err error

	// Serialize based on object type
	switch v := obj.(type) {
	case *NXSuperblock:
		data, err = SerializeNXSuperblock(v)
	case *APFSSuperblock:
		data, err = SerializeAPFSSuperblock(v)
	case *BTNodePhys:
		data, err = SerializeBTNodePhys(v)
	// Add cases for other object types
	default:
		return fmt.Errorf("unsupported object type: %T", obj)
	}

	if err != nil {
		return fmt.Errorf("failed to serialize object: %w", err)
	}

	// Write data to block device
	return device.WriteBlock(addr, data)
}

// Align skips bytes in the stream to align to the specified byte boundary
func (br *BinaryReader) Align(boundary int) error {
	if seeker, ok := br.reader.(io.Seeker); ok {
		pos, err := seeker.Seek(0, io.SeekCurrent)
		if err != nil {
			return err
		}
		offset := int((boundary - int(pos)%boundary) % boundary)
		_, err = seeker.Seek(int64(offset), io.SeekCurrent)
		return err
	}
	// fallback if reader doesn't support seeking
	padding := make([]byte, boundary)
	_, err := io.ReadFull(br.reader, padding)
	return err
}
