package ksuid

import "encoding/binary"

// CompressedSet is an immutable data type which stores a set of KSUIDs.
type CompressedSet []byte

// Iter returns an iterator that produces all KSUIDs in the set.
func (set CompressedSet) Iter() CompressedSetIter {
	return CompressedSetIter{
		content: []byte(set),
	}
}

// Compress creates and returns a compressed set of KSUIDs from the list given
// as arguments.
func Compress(ids ...KSUID) CompressedSet {
	c := byteLength + (len(ids) / 5)
	b := make([]byte, 0, c)
	return AppendCompressed(b, ids...)
}

// AppendCompressed uses the given byte slice as pre-allocated storage space to
// build a KSUID set.
//
// Note that the set uses a compression technique to store the KSUIDs, so the
// resuling length is not 20 x len(ids). The rule of thumb here is for the given
// byte slice to reserve the amount of memory that the application would be OK
// to waste.
func AppendCompressed(set []byte, ids ...KSUID) CompressedSet {
	if len(ids) != 0 {
		Sort(ids) // TODO: this is the only dynamic memory allocation :'(
		one := makeUint128(0, 1)

		// The first KSUID is always written to the set, this is the starting
		// point for all deltas.
		set = append(set, ids[0][:]...)

		timestamp := ids[0].Timestamp()
		lastKSUID := ids[0]
		lastValue := makeUint128FromPayload(ids[0][timestampLengthInBytes:])

		for i := 1; i != len(ids); i++ {
			id := ids[i]

			if id == lastKSUID {
				continue
			}

			t := id.Timestamp()
			p := id[timestampLengthInBytes:]
			v := makeUint128FromPayload(p)

			if t != timestamp {
				d := t - timestamp
				n := varintLength32(d)

				set = append(set, timeDelta|byte(n))
				set = appendVarint32(set, d, n)
				set = append(set, p...)

				timestamp = t
			} else {
				d := sub128(v, lastValue)

				if d != one {
					n := varintLength128(d)

					set = append(set, payloadDelta|byte(n))
					set = appendVarint128(set, d, n)
				} else {
					l := rangeLength(ids[i:], t, v)
					m := uint64(l + 1)
					n := varintLength64(m)

					set = append(set, payloadRange|byte(n))
					set = appendVarint64(set, m, n)

					i += l
				}
			}

			lastKSUID = id
			lastValue = v
		}
	}
	return CompressedSet(set)
}

func rangeLength(ids []KSUID, timestamp uint32, lastValue uint128) int {
	one := makeUint128(0, 1)

	for i := range ids {
		if ids[i].Timestamp() != timestamp {
			return i
		}

		p := ids[i][timestampLengthInBytes:]
		v := makeUint128FromPayload(p)

		if v != one {
			return i
		}

		lastValue = v
	}

	return len(ids)
}

func appendVarint128(b []byte, v uint128, n int) []byte {
	c := v.bytes()
	return append(b, c[len(c)-n:]...)
}

func appendVarint64(b []byte, v uint64, n int) []byte {
	c := [8]byte{}
	binary.BigEndian.PutUint64(c[:], v)
	return append(b, c[len(c)-n:]...)
}

func appendVarint32(b []byte, v uint32, n int) []byte {
	c := [4]byte{}
	binary.BigEndian.PutUint32(c[:], v)
	return append(b, c[len(c)-n:]...)
}

func varint128(b []byte) uint128 {
	a := [16]byte{}
	copy(a[16-len(b):], b)
	return makeUint128FromPayload(a[:])
}

func varint64(b []byte) uint64 {
	a := [8]byte{}
	copy(a[8-len(b):], b)
	return binary.BigEndian.Uint64(a[:])
}

func varint32(b []byte) uint32 {
	a := [4]byte{}
	copy(a[4-len(b):], b)
	return binary.BigEndian.Uint32(a[:])
}

func varintLength128(v uint128) int {
	if v[1] != 0 {
		return 8 + varintLength64(v[1])
	}
	return varintLength64(v[0])
}

func varintLength64(v uint64) int {
	switch {
	case (v & 0xFFFFFFFFFFFFFF00) == 0:
		return 1
	case (v & 0xFFFFFFFFFFFF0000) == 0:
		return 2
	case (v & 0xFFFFFFFFFF000000) == 0:
		return 3
	case (v & 0xFFFFFFFF00000000) == 0:
		return 4
	case (v & 0xFFFF000000000000) == 0:
		return 6
	default:
		return 8
	}
}

func varintLength32(v uint32) int {
	switch {
	case (v & 0xFFFFFF00) == 0:
		return 1
	case (v & 0xFFFF0000) == 0:
		return 2
	case (v & 0xFF000000) == 0:
		return 3
	default:
		return 4
	}
}

const (
	timeDelta    = (1 << 6)
	payloadDelta = (1 << 7)
	payloadRange = (1 << 6) | (1 << 7)
)

// CompressedSetIter is an iterator type returned by Set.Iter to produce the
// list of KSUIDs stored in a set.
//
// Here's is how the iterator type is commonly used:
//
//	for it := set.Iter(); it.Next(); {
//		id := it.KSUID
//		// ...
//	}
//
type CompressedSetIter struct {
	// KSUID is modified by calls to the Next method to hold the KSUID loaded
	// by the iterator.
	KSUID KSUID

	content []byte
	offset  int

	sequence  Sequence
	seqlength uint64

	timestamp uint32
	lastValue uint128
}

// Next moves the iterator forward, returning true if there a KSUID was found,
// or false if the iterator as reached the end of the set it was created from.
func (it *CompressedSetIter) Next() bool {
	if it.offset == len(it.content) {
		return false
	}

	if it.offset == 0 {
		copy(it.KSUID[:], it.content[:byteLength])
		it.offset += byteLength
		it.timestamp = it.KSUID.Timestamp()
		it.lastValue = makeUint128FromPayload(it.KSUID[timestampLengthInBytes:])
		return true
	}

	if it.seqlength != 0 {
		var err error
		it.KSUID, err = it.sequence.Next()
		if err != nil {
			panic(err)
		}
		it.seqlength--
		it.lastValue = makeUint128FromPayload(it.KSUID[timestampLengthInBytes:])
		return true
	}

	b := it.content[it.offset]
	it.offset++

	const mask = timeDelta | payloadDelta | payloadRange
	tag := int(b) & mask
	cnt := int(b) & ^mask

	switch tag {
	case timeDelta:
		off0 := it.offset
		off1 := off0 + cnt
		off2 := off1 + payloadLengthInBytes

		it.timestamp += varint32(it.content[off0:off1])
		it.KSUID = it.makeFromPayload(it.content[off1:off2])
		it.offset = off2
		it.lastValue = makeUint128FromPayload(it.content[off1:off2])

	case payloadDelta:
		off0 := it.offset
		off1 := off0 + cnt

		delta := varint128(it.content[off0:off1])
		value := add128(it.lastValue, delta)

		it.KSUID = it.makeFromUint128(value)
		it.offset = off1
		it.lastValue = value

	case payloadRange:
		off0 := it.offset
		off1 := off0 + cnt

		it.sequence.Seed = it.KSUID
		it.sequence.count = uint32(binary.BigEndian.Uint16(it.KSUID[byteLength-2:])) + 1
		it.seqlength = varint64(it.content[off0:off1])
		it.offset = off1

		// 2^16 IDs can be generated, the first one will not fail.
		it.KSUID, _ = it.sequence.Next()
		it.seqlength--
		it.lastValue = makeUint128FromPayload(it.KSUID[timestampLengthInBytes:])

	default:
		panic("KSUID set iterator is reading malformed data")
	}

	return true
}

func (it *CompressedSetIter) makeFromPayload(payload []byte) (id KSUID) {
	binary.BigEndian.PutUint32(id[:timestampLengthInBytes], it.timestamp)
	copy(id[timestampLengthInBytes:], payload)
	return
}

func (it *CompressedSetIter) makeFromUint128(value uint128) KSUID {
	payload := value.bytes()
	return it.makeFromPayload(payload[:])
}
