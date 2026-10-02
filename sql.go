package ksuid

import (
	"database/sql/driver"
	"fmt"
)

// Value converts the KSUID into a SQL driver value which can be used to
// directly use the KSUID as parameter to a SQL query.
//
// By default, if id is Nil, Value returns nil, nil (mapping to SQL NULL).
// In SQL schemas where columns are defined with NOT NULL constraints
// (such as CHAR(27) NOT NULL or BINARY(20) NOT NULL), passing a zero-value
// KSUID directly will result in a NOT NULL constraint violation.
// Similarly, using a zero-value KSUID in query parameters such as:
//
//	SELECT * FROM users WHERE id > ?
//
// will evaluate to "WHERE id > NULL", returning no rows.
//
// To store or query a zero-value KSUID without producing a SQL NULL, callers can:
//   - Pass id.String() (e.g. "000000000000000000000000000") for text columns.
//   - Pass id.Bytes() for binary columns.
//   - Use NullKSUID{KSUID: id, Valid: true}.
func (i KSUID) Value() (driver.Value, error) {
	if i.IsNil() {
		return nil, nil
	}
	return i.String(), nil
}

// Scan implements the sql.Scanner interface. It supports converting from
// string, []byte, or nil into a KSUID value. Attempting to convert from
// another type will return an error.
func (i *KSUID) Scan(src interface{}) error {
	switch v := src.(type) {
	case nil:
		return i.scan(nil)
	case []byte:
		return i.scan(v)
	case string:
		return i.scan([]byte(v))
	default:
		return fmt.Errorf("Scan: unable to scan type %T into KSUID", v)
	}
}

func (i *KSUID) scan(b []byte) error {
	switch len(b) {
	case 0:
		*i = Nil
		return nil
	case byteLength:
		return i.UnmarshalBinary(b)
	case stringEncodedLength:
		return i.UnmarshalText(b)
	default:
		return errSize
	}
}

// NullKSUID represents a KSUID that may be null.
// NullKSUID implements the sql.Scanner and driver.Valuer interfaces
// so it can be used as a scan destination, similar to sql.NullString.
type NullKSUID struct {
	KSUID KSUID
	Valid bool // Valid is true if KSUID is not NULL
}

// NullID is an alias for NullKSUID.
type NullID = NullKSUID

// Scan implements the sql.Scanner interface.
func (n *NullKSUID) Scan(src interface{}) error {
	if src == nil {
		n.KSUID, n.Valid = Nil, false
		return nil
	}
	n.Valid = false
	if err := n.KSUID.Scan(src); err != nil {
		return err
	}
	n.Valid = true
	return nil
}

// Value implements the driver.Valuer interface.
func (n NullKSUID) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return n.KSUID.String(), nil
}
