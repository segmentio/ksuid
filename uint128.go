package ksuid

import (
	"encoding/binary"
	"fmt"
)

// uint128 represents an unsigned 128 bits little endian integer.
type uint128 [2]uint64

func makeUint128(high uint64, low uint64) uint128 {
	return uint128{low, high}
}

func makeUint128FromPayload(payload []byte) uint128 {
	return makeUint128(
		binary.BigEndian.Uint64(payload[:8]),
		binary.BigEndian.Uint64(payload[8:]),
	)
}

func (v uint128) bytes() (b [16]byte) {
	binary.BigEndian.PutUint64(b[:8], v[1])
	binary.BigEndian.PutUint64(b[8:], v[0])
	return
}

func (v uint128) String() string {
	return fmt.Sprintf("0x%016X%016X", v[0], v[1])
}

const wordBitSize = 64

func cmp128(x, y uint128) int {
	for i := len(x); i != 0; {
		i--
		switch {
		case x[i] < y[i]:
			return -1
		case x[i] > y[i]:
			return +1
		}
	}
	return 0
}

func add128(x, y uint128) (z uint128) {
	var c uint64
	for i, xi := range x {
		yi := y[i]
		zi := xi + yi + c
		z[i] = zi
		c = (xi&yi | (xi|yi)&^zi) >> (wordBitSize - 1)
	}
	return
}

func sub128(x, y uint128) (z uint128) {
	var c uint64
	for i, xi := range x {
		yi := y[i]
		zi := xi - yi - c
		z[i] = zi
		c = (yi&^xi | (yi|^xi)&zi) >> (wordBitSize - 1)
	}
	return
}
