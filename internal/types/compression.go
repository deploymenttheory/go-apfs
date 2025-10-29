package types

// Compressed Data
// Compressed data is stored in blocks that are individually compressed using one of several compression algorithms.

// CompressedDataHeaderT is the header for a compressed data block.
type CompressedDataHeaderT struct {
	// The signature identifying this as compressed data.
	// The value of this field is always 0x6370636d ("cpcm" in little-endian, displayed as "fpmc").
	Signature uint32

	// The compression method used for this data block.
	// For possible values, see Compression Methods.
	CompressionMethod uint32

	// The uncompressed data size, in bytes.
	UncompressedSize uint64
}

// Compression Methods

// CompressionMethodType represents a compression method used in APFS.
type CompressionMethodType uint32

const (
	// CompressionMethodDeflate uses the DEFLATE compression algorithm.
	CompressionMethodDeflate CompressionMethodType = 1

	// CompressionMethodLzfse uses the LZFSE compression algorithm.
	CompressionMethodLzfse CompressionMethodType = 2

	// CompressionMethodLzvn uses the LZVN compression algorithm.
	CompressionMethodLzvn CompressionMethodType = 3

	// CompressionMethodLz4 uses the LZ4 compression algorithm.
	CompressionMethodLz4 CompressionMethodType = 4

	// CompressionMethodZstd uses the Zstandard compression algorithm.
	CompressionMethodZstd CompressionMethodType = 5
)

// CompressionSignature is the magic value for a compressed data block.
// In big-endian: 0x6370636d ("cpcm"), in little-endian display: "fpmc"
const CompressionSignature uint32 = 0x6370636d

// JDcompKeyT is the key half of a compressed data record.
type JDcompKeyT struct {
	// The record's header.
	// The object identifier in the header is the file-system object's identifier.
	// The type in the header is always APFS_TYPE_DSTREAM_ID.
	Hdr JKeyT

	// The logical offset within the file's data.
	LogicalAddr uint64
}

// JDcompValT is the value half of a compressed data record.
type JDcompValT struct {
	// A bit field that contains the length of the extent and its flags.
	// The extent's length is a uint64_t value, accessed as ref_count_and_flags & J_DCOMP_LEN_MASK.
	// The flags are accessed as (ref_count_and_flags & J_DCOMP_FLAG_MASK) >> J_DCOMP_FLAG_SHIFT.
	RefCountAndFlags uint64

	// The physical block address of the compressed data.
	PhysBlockNum uint64
}

// JDcompLenMask is the bit mask used to access the extent length.
const JDcompLenMask uint64 = 0x00ffffffffffffff

// JDcompFlagMask is the bit mask used to access the flags.
const JDcompFlagMask uint64 = 0xff00000000000000

// JDcompFlagShift is the bit shift used to access the flags.
const JDcompFlagShift uint64 = 56
