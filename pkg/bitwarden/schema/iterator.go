package schema

import "fmt"

/////////////////////////////////////////////////////////////////////////////////
// TYPES

type Iterable interface {
	*Folder | *Cipher
}

// Iterate over values
type Iterator[T Iterable] interface {
	// Return next value or nil if there are no more values
	Next() T

	// Decrypt the value
	Decrypt(T) T
}

// Concrete iterator
type iterator[T Iterable] struct {
	n      int
	values []T
}

/////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewIterator[T Iterable](profile *Profile, values []T) Iterator[T] {
	iterator := new(iterator[T])
	iterator.values = values
	return iterator
}

/////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (i *iterator[T]) Next() T {
	var result T
	if i.n < len(i.values) {
		result = i.values[i.n]
		i.n++
	}
	return result
}

func (i *iterator[T]) Decrypt(v T) T {
	fmt.Printf("TODO: Decrypt %T\n", v)
	return v
}
