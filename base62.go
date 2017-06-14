package ksuid

const (
	// lexographic ordering (based on Unicode table) is 0-9A-Za-z
	base62Characters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	zeroString       = "000000000000000000000000000"
)

// Converts a base 62 byte into the number value that it represents.
func base62Value(digit byte) byte {
	switch {
	case digit >= '0' && digit <= '9':
		return digit - '0'
	case digit >= 'A' && digit <= 'Z':
		return 10 + (digit - 'A')
	default:
		return 36 + (digit - 'a')
	}
}

// This function encodes the base 62 representation of the src KSUID in binary
// form into dst.
//
// In order to support a couple of optimizations the function assumes that src
// is 20 bytes long and dst is 27 bytes long.
//
// Any unused bytes in dst will be set to the padding '0' byte.
func fastEncodeBase62(dst []byte, src []byte) {
	const srcBase = 4294967296
	const dstBase = 62

	// Split src into 5 4-byte words, this is where most of the efficiency comes
	// from because this is a O(N^2) algorithm, and we make N = N / 4 by working
	// on 32 bits at a time.
	parts := [5]uint32{
		uint32(src[0])<<24 | uint32(src[1])<<16 | uint32(src[2])<<8 | uint32(src[3]),
		uint32(src[4])<<24 | uint32(src[5])<<16 | uint32(src[6])<<8 | uint32(src[7]),
		uint32(src[8])<<24 | uint32(src[9])<<16 | uint32(src[10])<<8 | uint32(src[11]),
		uint32(src[12])<<24 | uint32(src[13])<<16 | uint32(src[14])<<8 | uint32(src[15]),
		uint32(src[16])<<24 | uint32(src[17])<<16 | uint32(src[18])<<8 | uint32(src[19]),
	}

	n := len(dst)
	bp := parts[:]
	bq := [5]uint32{}

	for len(bp) != 0 {
		quotient := bq[:0]
		remainder := uint64(0)

		for _, c := range bp {
			value := uint64(c) + uint64(remainder)*srcBase
			digit := value / dstBase
			remainder = value % dstBase

			if len(quotient) != 0 || digit != 0 {
				quotient = append(quotient, uint32(digit))
			}
		}

		// Writes at the end of the destination buffer because we computed the
		// lowest bits first.
		n--
		dst[n] = base62Characters[remainder]
		bp = quotient
	}

	// Add padding at the head of the destination buffer for all bytes that were
	// not set.
	copy(dst[:n], zeroString)
}

// This function appends the base 62 representation of the KSUID in src to dst,
// and returns the extended byte slice.
// The result is left-padded with '0' bytes to always append 27 bytes to the
// destination buffer.
func fastAppendEncodeBase62(dst []byte, src []byte) []byte {
	dst = reserve(dst, stringEncodedLength)
	n := len(dst)
	fastEncodeBase62(dst[n:n+stringEncodedLength], src)
	return dst[:n+stringEncodedLength]
}

// This function decodes the base 62 representation of the src KSUID to the
// binary form into dst.
//
// In order to support a couple of optimizations the function assumes that src
// is 27 bytes long and dst is 20 bytes long.
//
// Any unused bytes in dst will be set to zero.
func fastDecodeBase62(dst []byte, src []byte) {
	const srcBase = 62
	const dstBase = 4294967296

	parts := [27]byte{
		base62Value(src[0]),
		base62Value(src[1]),
		base62Value(src[2]),
		base62Value(src[3]),
		base62Value(src[4]),
		base62Value(src[5]),
		base62Value(src[6]),
		base62Value(src[7]),
		base62Value(src[8]),
		base62Value(src[9]),

		base62Value(src[10]),
		base62Value(src[11]),
		base62Value(src[12]),
		base62Value(src[13]),
		base62Value(src[14]),
		base62Value(src[15]),
		base62Value(src[16]),
		base62Value(src[17]),
		base62Value(src[18]),
		base62Value(src[19]),

		base62Value(src[20]),
		base62Value(src[21]),
		base62Value(src[22]),
		base62Value(src[23]),
		base62Value(src[24]),
		base62Value(src[25]),
		base62Value(src[26]),
	}

	n := len(dst)
	bp := parts[:]
	bq := [stringEncodedLength]byte{}

	for len(bp) > 0 {
		quotient := bq[:0]
		remainder := uint64(0)

		for _, c := range bp {
			value := uint64(c) + uint64(remainder)*srcBase
			digit := value / dstBase
			remainder = value % dstBase

			if len(quotient) != 0 || digit != 0 {
				quotient = append(quotient, byte(digit))
			}
		}

		dst[n-1] = byte(remainder >> 24)
		dst[n-2] = byte(remainder >> 16)
		dst[n-3] = byte(remainder >> 8)
		dst[n-4] = byte(remainder)
		n -= 4
		bp = quotient
	}
}

// This function appends the base 62 decoded version of src into dst.
func fastAppendDecodeBase62(dst []byte, src []byte) []byte {
	dst = reserve(dst, byteLength)
	n := len(dst)
	fastDecodeBase62(dst[n:n+byteLength], src)
	return dst[:n+byteLength]
}

// Ensures that at least nbytes are available in the remaining capacity of the
// destination slice, if not, a new copy is made and returned by the function.
func reserve(dst []byte, nbytes int) []byte {
	c := cap(dst)
	n := len(dst)

	if avail := c - n; avail < nbytes {
		c *= 2
		if (c - n) < nbytes {
			c = n + nbytes
		}
		b := make([]byte, n, c)
		copy(b, dst)
		dst = b
	}

	return dst
}
