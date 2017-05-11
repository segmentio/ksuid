package ksuid

import (
	"strings"
	"testing"
	"time"
)

func TestConstructionTimestamp(t *testing.T) {
	x := New()
	nowTime := time.Now().Round(1 * time.Minute)
	xTime := x.Time().Round(1 * time.Minute)

	if xTime != nowTime {
		t.Fatal(xTime, "!=", nowTime)
	}
}

func TestNil(t *testing.T) {
	if !Nil.IsNil() {
		t.Fatal("Nil should be Nil!")
	}

	x, _ := FromBytes(make([]byte, ByteLength))
	if !x.IsNil() {
		t.Fatal("Zero-byte array should be Nil!")
	}
}

func TestEncoding(t *testing.T) {
	x, _ := FromBytes(make([]byte, ByteLength))
	if !x.IsNil() {
		t.Fatal("Zero-byte array should be Nil!")
	}

	encoded := x.String()
	expected := strings.Repeat("0", StringEncodedLength)

	if encoded != expected {
		t.Fatal("expected", expected, "encoded", encoded)
	}
}

func TestPadding(t *testing.T) {
	b := make([]byte, ByteLength)
	for i := 0; i < ByteLength; i++ {
		b[i] = 255
	}

	x, _ := FromBytes(b)
	xEncoded := x.String()
	nilEncoded := Nil.String()

	if len(xEncoded) != len(nilEncoded) {
		t.Fatal("Encoding should produce equal-length strings for zero and max case")
	}
}

func TestParse(t *testing.T) {
	_, err := Parse("123")
	if err != errStrSize {
		t.Fatal("Expected Parsing a 3-char string to return an error")
	}

	parsed, err := Parse(strings.Repeat("0", StringEncodedLength))
	if err != nil {
		t.Fatal("Unexpected error", err)
	}

	if Compare(parsed, Nil) != 0 {
		t.Fatal("Parsing all-zeroes string should equal Nil value",
			"expected:", Nil,
			"actual:", parsed)
	}

	maxBytes := make([]byte, ByteLength)
	for i := 0; i < ByteLength; i++ {
		maxBytes[i] = 255
	}
	maxBytesKSUID, err := FromBytes(maxBytes)
	if err != nil {
		t.Fatal("Unexpected error", err)
	}

	maxParseKSUID, err := Parse(MaxStringEncoded)
	if err != nil {
		t.Fatal("Unexpected error", err)
	}

	if Compare(maxBytesKSUID, maxParseKSUID) != 0 {
		t.Fatal("String decoder broke for max string")
	}
}

func TestEncodeAndDecode(t *testing.T) {
	x := New()
	builtFromEncodedString, err := Parse(x.String())
	if err != nil {
		t.Fatal("Unexpected error", err)
	}

	if Compare(x, builtFromEncodedString) != 0 {
		t.Fatal("Parse(X).String() != X")
	}
}
