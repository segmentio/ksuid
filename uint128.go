package ksuid

import (
	"encoding/binary"
	"fmt"
)

// uint128 represents an unsigned 128 bits little endian integer.
type uint128 [2]uint64

func uint128Payload(ksuid KSUID) uint128 {
	return makeUint128FromPayload(ksuid[timestampLengthInBytes:])
}

func makeUint128(high uint64, low uint64) uint128 {
	return uint128{low, high}
}

func makeUint128FromPayload(payload []byte) uint128 {
	return uint128{
		binary.BigEndian.Uint64(payload[8:]), // low
		binary.BigEndian.Uint64(payload[:8]), // high
	}
}

func (v uint128) ksuid(timestamp uint32) (out KSUID) {
	binary.BigEndian.PutUint32(out[:4], timestamp) // time
	binary.BigEndian.PutUint64(out[4:], v[1])      // high
	binary.BigEndian.PutUint64(out[12:], v[0])     // low
	return
}

func (v uint128) bytes() (out [16]byte) {
	binary.BigEndian.PutUint64(out[:8], v[1])
	binary.BigEndian.PutUint64(out[8:], v[0])
	return
}

func (v uint128) String() string {
	return fmt.Sprintf("0x%016X%016X", v[0], v[1])
}

const wordBitSize = 64

func cmp128(x, y uint128) int {
	if x[1] < y[1] {
		return -1
	}
	if x[1] > y[1] {
		return 1
	}
	if x[0] < y[0] {
		return -1
	}
	if x[0] > y[0] {
		return 1
	}
	return 0
}

func add128(x, y uint128) (z uint128) {
	x0 := x[0]
	y0 := y[0]
	z0 := x0 + y0
	z[0] = z0

	c := (x0&y0 | (x0|y0)&^z0) >> (wordBitSize - 1)

	z[1] = x[1] + y[1] + c
	return
}

func sub128(x, y uint128) (z uint128) {
	x0 := x[0]
	y0 := y[0]
	z0 := x0 - y0
	z[0] = z0

	c := (y0&^x0 | (y0|^x0)&z0) >> (wordBitSize - 1)

	z[1] = x[1] - y[1] - c
	return
}

func incr128(x uint128) uint128 {
	return add128(x, uint128{1, 0})
}
