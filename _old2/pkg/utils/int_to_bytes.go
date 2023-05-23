package utils

import "unsafe"

// IntToBytes converts an integer (int64) to a byte slice in big endian format.
func IntToBytes(num int64) []byte {
	size := int(unsafe.Sizeof(num))
	arr := make([]byte, size)

	for i := 0; i < size; i++ {
		byt := *(*uint8)(unsafe.Pointer(uintptr(unsafe.Pointer(&num)) + uintptr(i)))
		arr[size-i-1] = byt
	}

	return arr
}
