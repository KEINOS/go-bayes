package utils_test

import (
	"bytes"
	"fmt"
	"log"

	"github.com/KEINOS/go-bayes/pkg/utils"
)

// ----------------------------------------------------------------------------
//  AnyToBytes()
// ----------------------------------------------------------------------------

func ExampleAnyToBytes() {
	for _, tt := range []any{
		// below two are equivalent
		int(1),
		int16(1),
		// below two are equivalent
		int(-1),
		int32(-1),
		// below two are equivalent
		[]byte("hello, world"),
		[]string{"h", "e", "l", "l", "o", ",", " ", "w", "o", "r", "l", "d"},
		// below two are equivalent
		true,
		int(-1),
		// below two are equivalent
		false,
		int(0),
		// below two are equivalent
		float32(1.0),
		float64(1.0),
		// below two are equivalent
		[]int{1, 2, 3},
		[]byte{
			0, 0, 0, 0, 0, 0, 0, 1,
			0, 0, 0, 0, 0, 0, 0, 2,
			0, 0, 0, 0, 0, 0, 0, 3,
		},
	} {
		b, err := utils.AnyToBytes(tt)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(b)
	}
	// Output:
	// [0 0 0 0 0 0 0 1]
	// [0 0 0 0 0 0 0 1]
	// [255 255 255 255 255 255 255 255]
	// [255 255 255 255 255 255 255 255]
	// [104 101 108 108 111 44 32 119 111 114 108 100]
	// [104 101 108 108 111 44 32 119 111 114 108 100]
	// [255 255 255 255 255 255 255 255]
	// [255 255 255 255 255 255 255 255]
	// [0 0 0 0 0 0 0 0]
	// [0 0 0 0 0 0 0 0]
	// [63 240 0 0 0 0 0 0]
	// [63 240 0 0 0 0 0 0]
	// [0 0 0 0 0 0 0 1 0 0 0 0 0 0 0 2 0 0 0 0 0 0 0 3]
	// [0 0 0 0 0 0 0 1 0 0 0 0 0 0 0 2 0 0 0 0 0 0 0 3]
}

// ----------------------------------------------------------------------------
//  AnyToHash()
// ----------------------------------------------------------------------------

func ExampleAnyToHash() {
	hashed1, err := utils.AnyToHash("some data")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%x\n", hashed1)

	// Same value and type will result in the same hash
	hashed2, err := utils.AnyToHash("some", " ", "data")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%x\n", hashed2)

	// Same data but different type will result in different hash
	hashed3, err := utils.AnyToHash([]byte("some data"))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%x\n", hashed3)

	// Output:
	// 046f088042c332812cad6c0dea15c7cf903c3900ab165e4272f8eb485ac8fd02
	// 046f088042c332812cad6c0dea15c7cf903c3900ab165e4272f8eb485ac8fd02
	// 0b03dee6ad7385d489e40e065992a3385ea938e7f15da3919f37a8a1e7c89e23
}

// ----------------------------------------------------------------------------
//  AnyToStateID()
// ----------------------------------------------------------------------------

func ExampleAnyToStateID() {
	for _, tt := range []struct {
		value  any
		expect uint64
	}{
		{1, 4518803544265741878},
		{int32(1), 16914795258483526959},
		{int64(1), 6971089086649022728},
	} {
		expect := tt.expect

		actual, err := utils.AnyToStateID(tt.value) // example usage
		if err != nil {
			log.Fatal(err)
		}

		if expect != actual {
			log.Fatalf("type %T --> expected: %d, got: %d", tt.value, expect, actual)
		}
	}

	fmt.Println("ok")
	// Output: ok
}

// ----------------------------------------------------------------------------
//  AnyToTypeID()
// ----------------------------------------------------------------------------

func ExampleAnyToTypeID() {
	for _, tt := range []struct {
		value  any
		expect uint8 // expect ID of the type of the value
	}{
		{int(1), 1},
		{int8(1), 2},
		{int16(1), 3},
		{int32(1), 4},
		{int64(1), 5},
		{uint(1), 6},
		{uint8(1), 7},
		{uint16(1), 8},
		{uint32(1), 9},
		{uint64(1), 10},
		{float32(1.0), 11},
		{float64(1.0), 12},
		{"hello, world", 13},
		{[]byte("hello, world"), 14},
		{true, 15},
		{false, 15},
		{[]string{"hello", "world"}, 16},
		{[]int{1, 2, 3}, 17},
		{make(chan int), 0}, // unsupported type should return zero
		{struct{}{}, 0},     // unsupported type should return zero
	} {
		expect := tt.expect
		actual := utils.AnyToTypeID(tt.value) // example usage

		if expect != actual {
			log.Fatalf("type %T --> expected: %d, got: %d", tt.value, expect, actual)
		}
	}

	fmt.Println("ok")
	// Output: ok
}

// ----------------------------------------------------------------------------
//  BytesToHash()
// ----------------------------------------------------------------------------

func ExampleBytesToHash() {
	items := [][]byte{
		[]byte("some"),
		[]byte(" "),
		[]byte("data"),
	}

	fmt.Printf("%x\n", utils.BytesToHash(items...))
	fmt.Printf("%x\n", utils.BytesToHash([]byte("some data")))

	// Output:
	// b224a1da2bf5e72b337dc6dde457a05265a06dec8875be379e2ad2be5edb3bf2
	// b224a1da2bf5e72b337dc6dde457a05265a06dec8875be379e2ad2be5edb3bf2
}

// ----------------------------------------------------------------------------
//  BytesToStateID()
// ----------------------------------------------------------------------------

func ExampleBytesToStateID() {
	input := []byte{2, 3, 4, 5, 6, 7, 8, 9}
	result := utils.BytesToStateID(input)

	fmt.Println(result)
	fmt.Printf("%x\n", result)
	// Output:
	// 144964032628459520
	// 203040506070800
}

// ----------------------------------------------------------------------------
//  IntToBytes()
// ----------------------------------------------------------------------------

func ExampleIntToBytes() {
	for _, tt := range []struct {
		expect []byte
		value  int64
	}{
		{[]byte{0, 0, 0, 0, 0, 0, 0, 1}, 1},
		{[]byte{255, 255, 255, 255, 255, 255, 255, 255}, -1},
		{[]byte{0, 0, 0, 0, 0, 0, 0, 255}, 255},
		{[]byte{0, 0, 0, 0, 0, 0, 1, 0}, 256},
	} {
		expect := tt.expect
		actual := utils.IntToBytes(tt.value) // example usage

		if res := bytes.Compare(expect, actual); res != 0 {
			log.Fatalf("value %d --> expected: %v, got: %v", tt.value, expect, actual)
		}
	}

	fmt.Println("ok")
	// Output: ok
}

// ----------------------------------------------------------------------------
//  IntsToStateID()
// ----------------------------------------------------------------------------

func ExampleIntsToStateID() {
	input := []uint64{2, 3, 4, 5, 6, 7, 8, 9}
	result := utils.IntsToStateID(input)

	fmt.Println(result)
	fmt.Printf("%x\n", result)

	// Output:
	// 15423143640648593391
	// d60a036d04af2fef
}
