package ksuid

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"sort"
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

func TestIssue25(t *testing.T) {
	// https://github.com/segmentio/ksuid/issues/25
	for _, s := range []string{
		"aaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"aWgEPTl1tmebfsQzFP4bxwgy80!",
	} {
		_, err := Parse(s)
		if err != errStrValue {
			t.Error("invalid KSUID representations cannot be successfully parsed, got err =", err)
		}
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
	id, _ := Parse(maxStringEncoded)

	if v, err := id.Value(); err != nil {
		t.Error(err)
	} else if s, ok := v.(string); !ok {
		t.Error("not a string value")
	} else if s != maxStringEncoded {
		t.Error("bad string value::", s)
	}
}

func TestSqlValuerNilValue(t *testing.T) {
	if v, err := Nil.Value(); err != nil {
		t.Error(err)
	} else if v != nil {
		t.Errorf("bad nil value: %v", v)
	}
}

func TestSqlScanner(t *testing.T) {
	id1 := New()
	id2 := New()

	tests := []struct {
		ksuid KSUID
		value interface{}
	}{
		{Nil, nil},
		{id1, id1.String()},
		{id2, id2.Bytes()},
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

func TestAppend(t *testing.T) {
	for _, repr := range []string{"0pN1Own7255s7jwpwy495bAZeEa", "aWgEPTl1tmebfsQzFP4bxwgy80V"} {
		k, _ := Parse(repr)
		a := make([]byte, 0, stringEncodedLength)

		a = append(a, "?: "...)
		a = k.Append(a)

		if s := string(a); s != "?: "+repr {
			t.Error(s)
		}
	}
}

func TestSort(t *testing.T) {
	ids1 := [11]KSUID{}
	ids2 := [11]KSUID{}

	for i := range ids1 {
		ids1[i] = New()
	}

	ids2 = ids1
	sort.Slice(ids2[:], func(i, j int) bool {
		return Compare(ids2[i], ids2[j]) < 0
	})

	Sort(ids1[:])

	if !IsSorted(ids1[:]) {
		t.Error("not sorted")
	}

	if ids1 != ids2 {
		t.Error("bad order:")
		t.Log(ids1)
		t.Log(ids2)
	}
}

func TestPrevNext(t *testing.T) {
	tests := []struct {
		id   KSUID
		prev KSUID
		next KSUID
	}{
		{
			id:   Nil,
			prev: Max,
			next: KSUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
		},
		{
			id:   Max,
			prev: KSUID{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xfe},
			next: Nil,
		},
	}

	for _, test := range tests {
		t.Run(test.id.String(), func(t *testing.T) {
			testPrevNext(t, test.id, test.prev, test.next)
		})
	}
}

func TestGetTimestamp(t *testing.T) {
	nowTime := time.Now()
	x, _ := NewRandomWithTime(nowTime)
	xTime := int64(x.Timestamp())
	unix := nowTime.Unix()
	if xTime != unix-epochStamp {
		t.Fatal(xTime, "!=", unix)
	}
}

func testPrevNext(t *testing.T, id, prev, next KSUID) {
	id1 := id.Prev()
	id2 := id.Next()

	if id1 != prev {
		t.Error("previous id of the nil KSUID is wrong:", id1, "!=", prev)
	}

	if id2 != next {
		t.Error("next id of the nil KSUID is wrong:", id2, "!=", next)
	}
}

func BenchmarkAppend(b *testing.B) {
	a := make([]byte, 0, stringEncodedLength)
	k := New()

	for i := 0; i != b.N; i++ {
		k.Append(a)
	}
}

func BenchmarkString(b *testing.B) {
	k := New()

	for i := 0; i != b.N; i++ {
		_ = k.String()
	}
}

func BenchmarkParse(b *testing.B) {
	for i := 0; i != b.N; i++ {
		Parse(maxStringEncoded)
	}
}

func BenchmarkCompare(b *testing.B) {
	k1 := New()
	k2 := New()

	for i := 0; i != b.N; i++ {
		Compare(k1, k2)
	}
}

func BenchmarkSort(b *testing.B) {
	ids1 := [101]KSUID{}
	ids2 := [101]KSUID{}

	for i := range ids1 {
		ids1[i] = New()
	}

	for i := 0; i != b.N; i++ {
		ids2 = ids1
		Sort(ids2[:])
	}
}

func BenchmarkNew(b *testing.B) {
	b.Run("with crypto rand", func(b *testing.B) {
		SetRand(nil)
		for i := 0; i != b.N; i++ {
			New()
		}
	})
	b.Run("with math rand", func(b *testing.B) {
		SetRand(FastRander)
		for i := 0; i != b.N; i++ {
			New()
		}
	})
}

func TestFromPartsTimestampBounds(t *testing.T) {
	payload := make([]byte, payloadLengthInBytes)

	// Valid exact epoch timestamp: May 13, 2014
	epochTime := time.Unix(epochStamp, 0)
	id, err := FromParts(epochTime, payload)
	if err != nil {
		t.Fatalf("expected epoch time to be valid, got: %v", err)
	}
	if id.Timestamp() != 0 {
		t.Errorf("expected timestamp 0 at epoch, got: %d", id.Timestamp())
	}

	// Valid exact max timestamp
	maxTime := time.Unix(epochStamp+math.MaxUint32, 0)
	id, err = FromParts(maxTime, payload)
	if err != nil {
		t.Fatalf("expected max time to be valid, got: %v", err)
	}
	if id.Timestamp() != math.MaxUint32 {
		t.Errorf("expected timestamp %d at maxTime, got: %d", uint32(math.MaxUint32), id.Timestamp())
	}

	// Invalid: 1 second before epoch
	preEpoch := time.Unix(epochStamp-1, 0)
	if _, err := FromParts(preEpoch, payload); err != errTime {
		t.Errorf("expected errTime for pre-epoch time, got: %v", err)
	}
	if id := FromPartsOrNil(preEpoch, payload); id != Nil {
		t.Errorf("expected Nil for FromPartsOrNil with pre-epoch time, got: %v", id)
	}
	if _, err := NewRandomWithTime(preEpoch); err != errTime {
		t.Errorf("expected errTime for NewRandomWithTime with pre-epoch time, got: %v", err)
	}

	// Invalid: year 2000 (issue #58 reproduction)
	y2k := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	if _, err := FromParts(y2k, payload); err != errTime {
		t.Errorf("expected errTime for year 2000, got: %v", err)
	}
	if id := FromPartsOrNil(y2k, payload); id != Nil {
		t.Errorf("expected Nil for FromPartsOrNil with year 2000, got: %v", id)
	}

	// Invalid: zero value time.Time{}
	var zeroTime time.Time
	if _, err := FromParts(zeroTime, payload); err != errTime {
		t.Errorf("expected errTime for zero time, got: %v", err)
	}

	// Invalid: 1 second after max timestamp
	postMax := time.Unix(epochStamp+math.MaxUint32+1, 0)
	if _, err := FromParts(postMax, payload); err != errTime {
		t.Errorf("expected errTime for post-max time, got: %v", err)
	}
	if id := FromPartsOrNil(postMax, payload); id != Nil {
		t.Errorf("expected Nil for FromPartsOrNil with post-max time, got: %v", id)
	}
	if _, err := NewRandomWithTime(postMax); err != errTime {
		t.Errorf("expected errTime for NewRandomWithTime with post-max time, got: %v", err)
	}

	// Invalid payload size still takes precedence or returns errPayloadSize
	if _, err := FromParts(epochTime, make([]byte, 10)); err != errPayloadSize {
		t.Errorf("expected errPayloadSize, got: %v", err)
	}
}
