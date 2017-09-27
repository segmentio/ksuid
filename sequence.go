package ksuid

import (
	"encoding/binary"
	"errors"
	"math"
	"sync/atomic"
)

// Sequence is a KSUID generator which produces a sequence of ordered KSUIDs
// from a seed.
//
// Up to 65536 KSUIDs can be generated by for a single seed.
//
// A typical usage of a Sequence looks like this:
//
//	seq := ksuid.Sequence{
//		Seed: ksuid.New(),
//	}
//	id, err := seq.Next()
//
type Sequence struct {
	// The seed is used as base for the KSUID generator, all generated KSUIDs
	// share the same leading 18 bytes of the seed.
	Seed  KSUID
	count uint32 // uint32 for overlow, only 2 bytes are used
}

// Next produces the next KSUID in the sequence, or returns an error if the
// sequence has been exhausted.
func (seq *Sequence) Next() (KSUID, error) {
	id := seq.Seed // copy
	count := atomic.AddUint32(&seq.count, 1) - 1
	if count > math.MaxUint16 {
		return Nil, errors.New("too many IDs were generated")
	}
	binary.BigEndian.PutUint16(id[len(id)-2:], uint16(count))
	return id, nil
}
