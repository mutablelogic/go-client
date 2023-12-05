package writer

import (
	"strconv"

	// Packages
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Marshaller interface {
	Marshal() ([]byte, error)
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Convert any value to a byte array. If quote is true, then the value is
// quoted if it is a string.
func Marshal(v any, quote bool) ([]byte, error) {
	// Returns nil if v is nil
	if v == nil {
		return nil, nil
	}
	// Use marshaller if implemented
	if m, ok := v.(Marshaller); ok {
		return m.Marshal()
	}
	// Switch the type
	switch v := v.(type) {
	case string:
		if quote {
			return []byte(strconv.Quote(v)), nil
		} else {
			return []byte(v), nil
		}
	case bool:
		if v {
			return []byte("true"), nil
		} else {
			return []byte("false"), nil
		}
	case int:
		return []byte(strconv.FormatInt(int64(v), 10)), nil
	case int8:
		return []byte(strconv.FormatInt(int64(v), 10)), nil
	case int16:
		return []byte(strconv.FormatInt(int64(v), 10)), nil
	case int32:
		return []byte(strconv.FormatInt(int64(v), 10)), nil
	case int64:
		return []byte(strconv.FormatInt(int64(v), 10)), nil
	case uint:
		return []byte(strconv.FormatUint(uint64(v), 10)), nil
	case uint8:
		return []byte(strconv.FormatUint(uint64(v), 10)), nil
	case uint16:
		return []byte(strconv.FormatUint(uint64(v), 10)), nil
	case uint32:
		return []byte(strconv.FormatUint(uint64(v), 10)), nil
	case uint64:
		return []byte(strconv.FormatUint(uint64(v), 10)), nil
	case float32:
		return []byte(strconv.FormatFloat(float64(v), 'f', -1, 64)), nil
	case float64:
		return []byte(strconv.FormatFloat(float64(v), 'f', -1, 64)), nil
	default:
		return nil, ErrBadParameter.Withf("Unable to marshal: %T", v)
	}
}
