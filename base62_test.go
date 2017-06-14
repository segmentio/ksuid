package ksuid

import (
	"bytes"
	"sort"
	"strings"
	"testing"
)

func TestBase10ToBase62AndBack(t *testing.T) {
	number := []byte{1, 2, 3, 4}
	encoded := base2base(number, 10, 62)
	decoded := base2base(encoded, 62, 10)

	if bytes.Compare(number, decoded) != 0 {
		t.Fatal(number, " != ", decoded)
	}
}

func TestBase256ToBase62AndBack(t *testing.T) {
	number := []byte{255, 254, 253, 251}
	encoded := base2base(number, 256, 62)
	decoded := base2base(encoded, 62, 256)

	if bytes.Compare(number, decoded) != 0 {
		t.Fatal(number, " != ", decoded)
	}
}

func TestEncodeAndDecodeBase62(t *testing.T) {
	helloWorld := []byte("hello world")
	encoded := encodeBase62(helloWorld)
	decoded := decodeBase62(encoded)

	if len(encoded) < len(helloWorld) {
		t.Fatal("length of encoded base62 string", encoded, "should be >= than raw bytes!")

	}

	if bytes.Compare(helloWorld, decoded) != 0 {
		t.Fatal(decoded, " != ", helloWorld)
	}
}

func TestLexographicOrdering(t *testing.T) {
	unsortedStrings := make([]string, 256)
	for i := 0; i < 256; i++ {
		s := string(encodeBase62([]byte{0, byte(i)}))
		unsortedStrings[i] = strings.Repeat("0", 2-len(s)) + s
	}

	if !sort.StringsAreSorted(unsortedStrings) {
		sortedStrings := make([]string, len(unsortedStrings))
		for i, s := range unsortedStrings {
			sortedStrings[i] = s
		}
		sort.Strings(sortedStrings)

		t.Fatal("base62 encoder does not produce lexographically sorted output.",
			"expected:", sortedStrings,
			"actual:", unsortedStrings)
	}
}

func TestBase62Value(t *testing.T) {
	s := base62Characters

	for i := range s {
		v := int(base62Value(s[i]))

		if v != i {
			t.Error("bad value:")
			t.Log("<<<", i)
			t.Log(">>>", v)
		}
	}
}

func TestFastAppendEncodeBase62(t *testing.T) {
	id, _ := Parse("aWgEPTl1tmebfsQzFP4bxwgy80V")

	b0 := id[:]
	b1 := appendEncodeBase62(nil, b0)
	b2 := fastAppendEncodeBase62(nil, b0)

	s1 := string(b1)
	s2 := string(b2)

	if s1 != s2 {
		t.Error("bad base62 representation:")
		t.Log("<<<", s1, len(s1))
		t.Log(">>>", s2, len(s2))
	}
}

func TestFastAppendDecodeBase62(t *testing.T) {
	b1 := appendDecodeBase62(nil, []byte("aWgEPTl1tmebfsQzFP4bxwgy80V"))
	b2 := fastAppendDecodeBase62(nil, []byte("aWgEPTl1tmebfsQzFP4bxwgy80V"))

	if !bytes.Equal(b1, b2) {
		t.Error("bad binary representation:")
		t.Log("<<<", b1)
		t.Log(">>>", b2)
	}
}

func BenchmarkAppendEncodeBase62(b *testing.B) {
	a := [stringEncodedLength]byte{}
	id := New()

	for i := 0; i != b.N; i++ {
		appendEncodeBase62(a[:0], id[:])
	}
}

func BenchmarkAppendFastEncodeBase62(b *testing.B) {
	a := [stringEncodedLength]byte{}
	id := New()

	for i := 0; i != b.N; i++ {
		fastAppendEncodeBase62(a[:0], id[:])
	}
}

func BenchmarkAppendDecodeBase62(b *testing.B) {
	a := [byteLength]byte{}
	id := []byte(New().String())

	for i := 0; i != b.N; i++ {
		b := [stringEncodedLength]byte{}
		copy(b[:], id)
		appendDecodeBase62(a[:0], b[:])
	}
}

func BenchmarkAppendFastDecodeBase62(b *testing.B) {
	a := [byteLength]byte{}
	id := []byte(New().String())

	for i := 0; i != b.N; i++ {
		fastAppendDecodeBase62(a[:0], id)
	}
}

// The functions bellow were the initial implementation of the base conversion
// algorithms, they were replaced by optimized versions later on. We keep them
// in the test files as a reference to ensure compatibility between the generic
// and optimized implementations.

func appendBase2Base(dst []byte, src []byte, inBase int, outBase int) []byte {
	off := len(dst)
	bs := src[:]
	bq := [stringEncodedLength]byte{}

	for len(bs) > 0 {
		length := len(bs)
		quotient := bq[:0]
		remainder := 0

		for i := 0; i != length; i++ {
			acc := int(bs[i]) + remainder*inBase
			d := acc/outBase | 0
			remainder = acc % outBase

			if len(quotient) > 0 || d > 0 {
				quotient = append(quotient, byte(d))
			}
		}

		// Appends in reverse order, the byte slice gets reversed before it's
		// returned by the function.
		dst = append(dst, byte(remainder))
		bs = quotient
	}

	reverse(dst[off:])
	return dst
}

func base2base(src []byte, inBase int, outBase int) []byte {
	return appendBase2Base(nil, src, inBase, outBase)
}

func appendEncodeBase62(dst []byte, src []byte) []byte {
	off := len(dst)
	dst = appendBase2Base(dst, src, 256, 62)
	for i, c := range dst[off:] {
		dst[off+i] = base62Characters[c]
	}
	return dst
}

func encodeBase62(in []byte) []byte {
	return appendEncodeBase62(nil, in)
}

func appendDecodeBase62(dst []byte, src []byte) []byte {
	// Kind of intrusive, we modify the input buffer... it's OK here, it saves
	// a memory allocation in Parse.
	for i, b := range src {
		// O(1)... technically. Has better real-world perf than a map
		src[i] = byte(strings.IndexByte(base62Characters, b))
	}
	return appendBase2Base(dst, src, 62, 256)
}

func decodeBase62(src []byte) []byte {
	return appendDecodeBase62(
		make([]byte, 0, len(src)*2),
		append(make([]byte, 0, len(src)), src...),
	)
}

func reverse(b []byte) {
	i := 0
	j := len(b) - 1

	for i < j {
		b[i], b[j] = b[j], b[i]
		i++
		j--
	}
}
