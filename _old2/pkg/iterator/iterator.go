package iterator

// Iterator is an interface that defines the methods that a database engine
// must implement.
type Iterator[Item any] interface {
	// HasNext must return true if there is still data to read.
	HasNext() bool
	// Next must return the next item read. It returns nil if there is no more data.
	Next() *Item
	// Reset must reset the iterator to the beginning of the table.
	Reset()
}
