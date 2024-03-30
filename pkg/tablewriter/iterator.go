package tablewriter

import (
	"fmt"
	"reflect"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type iterator struct {
	slice reflect.Value
	index int
}

///////////////////////////////////////////////////////////////////////////////
// CONSTRUCTOR

// NewTableMeta returns a new table metadata object, from a single struct
// value or an array of one or more struct values which are of the same type
func NewIterator(v any) (*iterator, error) {
	self := new(iterator)

	// Get the type
	rt, isSlice, err := typeOf(v)
	if err != nil {
		return nil, err
	}
	// Set the slice parameter
	if isSlice {
		self.slice = reflect.ValueOf(v)
	} else {
		self.slice = reflect.MakeSlice(reflect.SliceOf(rt), 1, 1)
		rv := reflect.ValueOf(v)
		if rv.Kind() == reflect.Ptr {
			self.slice.Index(0).Set(rv.Elem())
		} else {
			self.slice.Index(0).Set(rv)
		}
	}
	// Set the index parameter
	self.index = 0

	// Return success
	return self, nil
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (i *iterator) String() string {
	str := "<iterator"
	str += fmt.Sprint(" len=", i.slice.Len())
	str += fmt.Sprint(" i=", i.index)
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return the number of elements
func (i *iterator) Reset() {
	i.index = 0
}

// Return the number of elements
func (i *iterator) Len() int {
	return i.slice.Len()
}

// Return the next element, or nil
func (i *iterator) Next() any {
	if i.index >= i.slice.Len() {
		return nil
	}
	v := i.slice.Index(i.index).Interface()
	i.index++
	return v
}
