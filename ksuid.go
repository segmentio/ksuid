package ksuid

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"strings"
	"time"
)

const (
	// KSUID's epoch starts more recently so that the 32-bit number space gives a
	// significantly higher useful lifetime of around 136 years from March 2017
	EpochStamp int64 = 1400000000

	// Timestamp is a uint32
	TimestampLengthInBytes = 4

	// Payload is 16-bytes
	PayloadLengthInBytes = 16

	// KSUIDs are 20 bytes when binary encoded
	ByteLength = TimestampLengthInBytes + PayloadLengthInBytes

	// The length of a KSUID when string (base62) encoded
	StringEncodedLength = 27

	// A string-encoded maximum value for a KSUID
	MaxStringEncoded = "aWgEPTl1tmebfsQzFP4bxwgy80V"
)

// KSUIDs are 20 bytes:
//  00-03 byte: uint32 BE UTC timestamp with custom epoch
//  04-19 byte: random "payload"
type KSUID [ByteLength]byte

var (
	rander = rand.Reader

	errSize    = fmt.Errorf("Valid KSUIDs are %v bytes", ByteLength)
	errStrSize = fmt.Errorf("Valid encoded KSUIDs are %v characters", StringEncodedLength)

	// Represents a completely empty (invalid) KSUID
	Nil KSUID
)

// The timestamp portion of the ID as a Time object
func (i KSUID) Time() time.Time {
	return correctedUTCTimestampToTime(i.Timestamp())
}

// The timestamp portion of the ID as a bare integer which is uncorrected
// for KSUID's special epoch.
func (i KSUID) Timestamp() uint32 {
	return binary.BigEndian.Uint32(i[:TimestampLengthInBytes])
}

// The 16-byte random payload without the timestamp
func (i KSUID) Payload() []byte {
	return i[TimestampLengthInBytes:]
}

// String-encoded representation that can be passed through Parse()
func (i KSUID) String() string {
	encoded := encodeBase62(i[:])

	padAmount := StringEncodedLength - len(encoded)
	if padAmount > 0 {
		return strings.Repeat("0", padAmount) + encoded
	}

	return encoded
}

// The underlying buffer for this KSUID
func (i KSUID) buffer() []byte {
	return i[:]
}

// Raw byte representation of KSUID
func (i KSUID) Bytes() []byte {
	out := make([]byte, ByteLength)
	for i, b := range i {
		out[i] = b
	}
	return out
}

// Returns true if this is a "nil" KSUID
func (i KSUID) Nil() bool {
	return Compare(Nil, i) == 0
}

// Decodes a string-encoded representation of a KSUID object
func Parse(s string) (KSUID, error) {
	if len(s) != StringEncodedLength {
		return Nil, errStrSize
	}

	decoded := decodeBase62(s)
	padAmount := ByteLength - len(decoded)
	if padAmount > 0 {
		decoded = append(make([]byte, padAmount), decoded...)
	}

	return FromBytes(decoded)
}

func timeToCorrectedUTCTimestamp(t time.Time) uint32 {
	return uint32(t.Unix() - EpochStamp)
}

func correctedUTCTimestampToTime(ts uint32) time.Time {
	return time.Unix(int64(ts)+EpochStamp, 0)
}

// Generates a new KSUID
func New() KSUID {
	var ksuid KSUID

	// Fill in the payload bytes with random stuff
	_, err := io.ReadFull(rander, ksuid[4:])
	if err != nil {
		return Nil
	}

	// Grab the current timestamp and fill in the timestamp portion
	now := time.Now()
	ts := timeToCorrectedUTCTimestamp(now)
	binary.BigEndian.PutUint32(ksuid[:4], ts)

	return ksuid
}

// Constructs a KSUID from a 20-byte binary representation
func FromBytes(b []byte) (KSUID, error) {
	var ksuid KSUID

	if len(b) != ByteLength {
		return Nil, errSize
	}

	for i := 0; i < ByteLength; i++ {
		ksuid[i] = b[i]
	}

	return ksuid, nil
}

// Sets the global source of random bytes for KSUID generation.
func SetRand(r io.Reader) {
	if r == nil {
		rander = rand.Reader
		return
	}
	rander = r
}

// Implements comparison for KSUID type
func Compare(a, b KSUID) int {
	return bytes.Compare(a.buffer(), b.buffer())
}
