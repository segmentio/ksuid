package ksuid

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
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

	x, _ := FromBytes(make([]byte, byteLength))
	if !x.IsNil() {
		t.Fatal("Zero-byte array should be Nil!")
	}
}

func TestEncoding(t *testing.T) {
	x, _ := FromBytes(make([]byte, byteLength))
	if !x.IsNil() {
		t.Fatal("Zero-byte array should be Nil!")
	}

	encoded := x.String()
	expected := strings.Repeat("0", stringEncodedLength)

	if encoded != expected {
		t.Fatal("expected", expected, "encoded", encoded)
	}
}

func TestPadding(t *testing.T) {
	b := make([]byte, byteLength)
	for i := 0; i < byteLength; i++ {
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

	parsed, err := Parse(strings.Repeat("0", stringEncodedLength))
	if err != nil {
		t.Fatal("Unexpected error", err)
	}

	if Compare(parsed, Nil) != 0 {
		t.Fatal("Parsing all-zeroes string should equal Nil value",
			"expected:", Nil,
			"actual:", parsed)
	}

	maxBytes := make([]byte, byteLength)
	for i := 0; i < byteLength; i++ {
		maxBytes[i] = 255
	}
	maxBytesKSUID, err := FromBytes(maxBytes)
	if err != nil {
		t.Fatal("Unexpected error", err)
	}

	maxParseKSUID, err := Parse(maxStringEncoded)
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

func TestMarshalText(t *testing.T) {
	var id1 = New()
	var id2 KSUID

	if err := id2.UnmarshalText([]byte(id1.String())); err != nil {
		t.Fatal(err)
	}

	if id1 != id2 {
		t.Fatal(id1, "!=", id2)
	}

	if b, err := id2.MarshalText(); err != nil {
		t.Fatal(err)
	} else if s := string(b); s != id1.String() {
		t.Fatal(s)
	}
}

func TestMarshalBinary(t *testing.T) {
	var id1 = New()
	var id2 KSUID

	if err := id2.UnmarshalBinary(id1.Bytes()); err != nil {
		t.Fatal(err)
	}

	if id1 != id2 {
		t.Fatal(id1, "!=", id2)
	}

	if b, err := id2.MarshalBinary(); err != nil {
		t.Fatal(err)
	} else if bytes.Compare(b, id1.Bytes()) != 0 {
		t.Fatal("bad binary form:", id2)
	}
}

func TestMashalJSON(t *testing.T) {
	var id1 = New()
	var id2 KSUID

	if b, err := json.Marshal(id1); err != nil {
		t.Fatal(err)
	} else if err := json.Unmarshal(b, &id2); err != nil {
		t.Fatal(err)
	} else if id1 != id2 {
		t.Error(id1, "!=", id2)
	}
}

func TestFlag(t *testing.T) {
	var id1 = New()
	var id2 KSUID

	fset := flag.NewFlagSet("test", flag.ContinueOnError)
	fset.Var(&id2, "id", "the KSUID")

	if err := fset.Parse([]string{"-id", id1.String()}); err != nil {
		t.Fatal(err)
	}

	if id1 != id2 {
		t.Error(id1, "!=", id2)
	}
}

func TestSqlValuer(t *testing.T) {
	id := parse(maxStringEncoded)

	if v, err := id.Value(); err != nil {
		t.Error(err)
	} else if s, ok := v.(string); !ok {
		t.Error("not a string value")
	} else if s != maxStringEncoded {
		t.Error("bad string value::", s)
	}
}

func TestSqlScanner(t *testing.T) {
	tests := []struct {
		ksuid KSUID
		value interface{}
	}{
		{Nil, nil},
		{parse(maxStringEncoded), maxStringEncoded},
		{parse(maxStringEncoded), []byte(maxStringEncoded)},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%T", test.value), func(t *testing.T) {
			var id KSUID

			if err := id.Scan(test.value); err != nil {
				t.Error(err)
			}

			if id != test.ksuid {
				t.Error("bad KSUID:")
				t.Logf("expected %v", test.ksuid)
				t.Logf("found    %v", id)
			}
		})
	}
}

func parse(s string) KSUID {
	id, _ := Parse(s)
	return id
}
