package ksuid

import (
	"strings"
	"testing"
)

func TestScan(t *testing.T) {
	stringTest := "1al9byIH8Ze6OLkD5tZqmByJkSX"
	badTypeTest := 6
	invalidTest := "1al9byIH8Ze6OLkD5tZqmByJk"

	byteTest := make([]byte, byteLength)
	byteTestKSUID := Must(Parse(stringTest))
	copy(byteTest, byteTestKSUID[:])
	textTest := []byte(stringTest)

	// valid tests

	var ksuid KSUID
	err := (&ksuid).Scan(stringTest)
	if err != nil {
		t.Fatal(err)
	}

	err = (&ksuid).Scan([]byte(stringTest))
	if err != nil {
		t.Fatal(err)
	}
	err = (&ksuid).Scan(byteTest)
	if err != nil {
		t.Fatal(err)
	}

	err = (&ksuid).Scan(textTest)
	if err != nil {
		t.Fatal(err)
	}

	// bad type tests

	err = (&ksuid).Scan(badTypeTest)
	if err == nil {
		t.Error("int correctly parsed and shouldn't have")
	}
	if !strings.Contains(err.Error(), "unable to scan type") {
		t.Error("attempting to parse an int returned an incorrect error message")
	}

	// invalid/incomplete ksuids
	err = (&ksuid).Scan(invalidTest)
	if err == nil {
		t.Error("invalid uuid was parsed without error")
	}
	if !strings.Contains(err.Error(), "Valid encoded KSUIDs") {
		t.Error("attempting to parse an invalid KSUID returned an incorrect error message")
	}

	err = (&ksuid).Scan(byteTest[:len(byteTest)-2])
	if err == nil {
		t.Error("invalid byte ksuid was parsed without error")
	}
	if !strings.Contains(err.Error(), "Valid encoded KSUIDs") {
		t.Error("attempting to parse an invalid byte KSUID returned an incorrect error message")
	}

	// empty tests

	ksuid = KSUID{}
	var emptySlice []byte
	err = (&ksuid).Scan(emptySlice)
	if err != nil {
		t.Fatal(err)
	}

	for _, v := range ksuid {
		if v != 0 {
			t.Error("KSUID was not nil after scanning empty byte slice")
		}
	}

	ksuid = KSUID{}
	var emptyString string
	err = (&ksuid).Scan(emptyString)
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range ksuid {
		if v != 0 {
			t.Error("KSUID was not nil after scanning empty string")
		}
	}

	ksuid = KSUID{}
	err = (&ksuid).Scan(nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range ksuid {
		if v != 0 {
			t.Error("KSUID was not nil after scanning nil")
		}
	}
}

func TestValue(t *testing.T) {
	stringTest := "1al9byIH8Ze6OLkD5tZqmByJkSX"
	ksuid := Must(Parse(stringTest))
	val, _ := ksuid.Value()
	if val != stringTest {
		t.Error("Value() did ot return expected string")
	}
}
