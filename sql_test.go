package ksuid

import (
	"database/sql"
	"database/sql/driver"
	"testing"
)

func TestSqlValuer(t *testing.T) {
	id, _ := Parse(maxStringEncoded)

	if v, err := id.Value(); err != nil {
		t.Error(err)
	} else if s, ok := v.(string); !ok {
		t.Error("not a string value")
	} else if s != maxStringEncoded {
		t.Error("bad string value:", s)
	}

	randomID := New()
	if v, err := randomID.Value(); err != nil {
		t.Error(err)
	} else if s, ok := v.(string); !ok {
		t.Error("not a string value")
	} else if s != randomID.String() {
		t.Errorf("expected %s, got %s", randomID.String(), s)
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
		name     string
		ksuid    KSUID
		value    interface{}
		hasError bool
	}{
		{"nil", Nil, nil, false},
		{"empty string", Nil, "", false},
		{"empty bytes", Nil, []byte{}, false},
		{"zero string", Nil, minStringEncoded, false},
		{"zero bytes", Nil, make([]byte, byteLength), false},
		{"string id1", id1, id1.String(), false},
		{"bytes id2", id2, id2.Bytes(), false},
		{"invalid type int", Nil, 12345, true},
		{"invalid short string", Nil, "too short", true},
		{"invalid short bytes", Nil, []byte{1, 2, 3}, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var id KSUID
			err := id.Scan(test.value)
			if test.hasError {
				if err == nil {
					t.Errorf("expected error scanning %v, got nil", test.value)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error scanning %v: %v", test.value, err)
			}
			if id != test.ksuid {
				t.Errorf("expected %v, got %v", test.ksuid, id)
			}
		})
	}
}

func TestSqlNilRoundTrip(t *testing.T) {
	// Round trip Nil via driver.Value -> Scan
	v, err := Nil.Value()
	if err != nil {
		t.Fatal(err)
	}
	if v != nil {
		t.Fatalf("expected nil driver.Value, got %v", v)
	}

	var scanned KSUID
	if err := scanned.Scan(v); err != nil {
		t.Fatal(err)
	}
	if scanned != Nil {
		t.Fatalf("expected Nil, got %v", scanned)
	}

	// Round trip Nil string representation (e.g. for NOT NULL columns or range queries)
	zeroStr := Nil.String()
	if err := scanned.Scan(zeroStr); err != nil {
		t.Fatal(err)
	}
	if scanned != Nil {
		t.Fatalf("expected Nil after scanning zero string, got %v", scanned)
	}

	// Round trip Nil binary representation
	zeroBytes := Nil.Bytes()
	if err := scanned.Scan(zeroBytes); err != nil {
		t.Fatal(err)
	}
	if scanned != Nil {
		t.Fatalf("expected Nil after scanning zero bytes, got %v", scanned)
	}
}

func TestNullKSUIDValuer(t *testing.T) {
	// Invalid (null) value
	nullID := NullKSUID{}
	if v, err := nullID.Value(); err != nil {
		t.Error(err)
	} else if v != nil {
		t.Errorf("expected nil value for invalid NullKSUID, got %v", v)
	}

	// Valid with random KSUID
	randomID := New()
	validID := NullKSUID{KSUID: randomID, Valid: true}
	if v, err := validID.Value(); err != nil {
		t.Error(err)
	} else if s, ok := v.(string); !ok || s != randomID.String() {
		t.Errorf("expected %s, got %v", randomID.String(), v)
	}

	// Valid with Nil KSUID (zero value representation)
	validNilID := NullKSUID{KSUID: Nil, Valid: true}
	if v, err := validNilID.Value(); err != nil {
		t.Error(err)
	} else if s, ok := v.(string); !ok || s != minStringEncoded {
		t.Errorf("expected %s, got %v", minStringEncoded, v)
	}
}

func TestNullKSUIDScanner(t *testing.T) {
	randomID := New()

	tests := []struct {
		name      string
		src       interface{}
		wantValid bool
		wantKSUID KSUID
		hasError  bool
	}{
		{"nil", nil, false, Nil, false},
		{"string id", randomID.String(), true, randomID, false},
		{"bytes id", randomID.Bytes(), true, randomID, false},
		{"zero string", minStringEncoded, true, Nil, false},
		{"zero bytes", make([]byte, byteLength), true, Nil, false},
		{"empty string", "", true, Nil, false},
		{"empty bytes", []byte{}, true, Nil, false},
		{"invalid type", 123, false, Nil, true},
		{"invalid string length", "too short", false, Nil, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var n NullKSUID
			err := n.Scan(test.src)
			if test.hasError {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				if n.Valid {
					t.Errorf("expected Valid to be false on error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if n.Valid != test.wantValid {
				t.Errorf("expected Valid=%v, got %v", test.wantValid, n.Valid)
			}
			if n.KSUID != test.wantKSUID {
				t.Errorf("expected KSUID=%v, got %v", test.wantKSUID, n.KSUID)
			}
		})
	}
}

func TestNullKSUIDRoundTrip(t *testing.T) {
	// Round trip SQL NULL
	n1 := NullKSUID{KSUID: Nil, Valid: false}
	v1, err := n1.Value()
	if err != nil {
		t.Fatal(err)
	}
	if v1 != nil {
		t.Fatalf("expected nil driver.Value, got %v", v1)
	}
	var scanned1 NullKSUID
	if err := scanned1.Scan(v1); err != nil {
		t.Fatal(err)
	}
	if scanned1.Valid || scanned1.KSUID != Nil {
		t.Fatalf("expected Valid=false and KSUID=Nil, got %+v", scanned1)
	}

	// Round trip zero value KSUID
	n2 := NullKSUID{KSUID: Nil, Valid: true}
	v2, err := n2.Value()
	if err != nil {
		t.Fatal(err)
	}
	if s, ok := v2.(string); !ok || s != minStringEncoded {
		t.Fatalf("expected minStringEncoded, got %v", v2)
	}
	var scanned2 NullKSUID
	if err := scanned2.Scan(v2); err != nil {
		t.Fatal(err)
	}
	if !scanned2.Valid || scanned2.KSUID != Nil {
		t.Fatalf("expected Valid=true and KSUID=Nil, got %+v", scanned2)
	}

	// Round trip random KSUID
	id := New()
	n3 := NullKSUID{KSUID: id, Valid: true}
	v3, err := n3.Value()
	if err != nil {
		t.Fatal(err)
	}
	var scanned3 NullKSUID
	if err := scanned3.Scan(v3); err != nil {
		t.Fatal(err)
	}
	if !scanned3.Valid || scanned3.KSUID != id {
		t.Fatalf("expected Valid=true and KSUID=%v, got %+v", id, scanned3)
	}
}

// Compile-time interface assertion
var (
	_ driver.Valuer = KSUID{}
	_ sql.Scanner   = (*KSUID)(nil)
	_ driver.Valuer = NullKSUID{}
	_ sql.Scanner   = (*NullKSUID)(nil)
)
