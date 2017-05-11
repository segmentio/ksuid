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
		s := encodeBase62([]byte{0, byte(i)})
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
