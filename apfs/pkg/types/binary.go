// File: pkg/types/binary.go
package types

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
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
