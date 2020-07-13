package ksuid

import (
	"database/sql/driver"
	"fmt"
)

// Scan implements the sql.Scanner interface. It supports converting from
// string, []byte, or nil into a KSUID value. Attempting to convert from
// another type will return an error.
func (i *KSUID) Scan(src interface{}) error {
	switch src := src.(type) {
	case nil:
		return nil
	case string:
		// if an empty KSUID comes from a table, we return a null KSUID
		if src == "" {
			return nil
		}
		k, err := Parse(src)
		if err != nil {
			return fmt.Errorf("Scan: %v", err)
		}
		*i = k
	case []byte:
		// if an empty KSUID comes from a table, we return a null KSUID
		if len(src) == 0 {
			return nil
		}
		// assumes a simple slice of bytes if [byteLength] bytes
		if len(src) == byteLength {
			copy((*i)[:], src)
			return nil
		}

		if len(src) == stringEncodedLength {
			return i.Scan(string(src))
		}

		return i.Scan(string(src))

	default:
		return fmt.Errorf("Scan: unable to scan type %T into KSUID", src)
	}
	return nil
}

// Value implements sql.Valuer so that KSUIDs can be written to databases
// transparently. Currently, KSUIDs map to strings. Please consult database-specific
// driver documentation for matching types.
func (i *KSUID) Value() (driver.Value, error) {
	if i == nil || i.IsNil() {
		return nil, nil
	}
	return i.String(), nil
}
