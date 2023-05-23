package data_test

import (
	"fmt"
	"log"

	"github.com/KEINOS/go-bayes/pkg/data"
)

// ============================================================================
//  Functions
// ============================================================================

// ----------------------------------------------------------------------------
//  NewItem()
// ----------------------------------------------------------------------------

func ExampleNewItem() {
	var str *string
	s := "1"
	str = &s // str is a pointer to s

	for i, v := range []any{
		int(1),
		int8(1),
		int16(1),
		int32(1),
		int64(1),
		uint(1),
		uint8(1), // uint8(x) == byte(x)
		byte(1),  // byte(x) == uint8(x)
		uint16(1),
		uint32(1),
		uint64(1),
		float32(1),
		float64(1),
		[]int8{1},
		[]uint8{1}, // []uint8{x} == []byte{x}
		[]byte{1},  // []byte{x} == []uint8{x}
		"1",
		*str, // pointer is not allowed so you need to plug in its actual value
		[]string{"1"},
	} {
		item, err := data.NewItem(v)
		if err != nil {
			log.Fatal("item #", i, "error msg:", err)
		}

		if !item.IsValidUID() {
			log.Fatal("item #", i, "UID is not valid")
		}

		fmt.Printf("%-10T --> %T, %v\n", v, item.Value, item.UID)
	}

	// Output:
	// int        --> int, 5499574118996709625
	// int8       --> int8, 9267480926123474231
	// int16      --> int16, 8302621575810894237
	// int32      --> int32, 5099122415246212248
	// int64      --> int64, 17661666406993562769
	// uint       --> uint, 3336975496170789273
	// uint8      --> uint8, 17793169530411995228
	// uint8      --> uint8, 17793169530411995228
	// uint16     --> uint16, 3619354590694453937
	// uint32     --> uint32, 13669797085912375188
	// uint64     --> uint64, 4019400696499439099
	// float32    --> float32, 14522299580392283628
	// float64    --> float64, 2679853070506119653
	// []int8     --> []int8, 10376156057276158199
	// []uint8    --> []uint8, 2782361327378389732
	// []uint8    --> []uint8, 2782361327378389732
	// string     --> string, 12263243927642032000
	// string     --> string, 12263243927642032000
	// []string   --> []string, 16832052253739009849
}

func ExampleNewItem_unsupported_types() {
	var v interface{}

	// List of unsupported types
	for _, input := range []any{
		nil,
		v,             // Invalid reflect.Value is error
		new(struct{}), // Pointers are not allowed
		new(string),
	} {
		// Unsupported type returns error.
		if _, err := data.NewItem(input); err != nil {
			fmt.Printf("Type %T returns an error\n", input)
		}
	}

	// Output:
	// Type <nil> returns an error
	// Type <nil> returns an error
	// Type *struct {} returns an error
	// Type *string returns an error
}

// ----------------------------------------------------------------------------
//  DecodeItem()
// ----------------------------------------------------------------------------

func ExampleDecodeItem() {
	encData := []byte{
		36, 255, 129, 3, 1, 1, 4, 73, 116, 101, 109, 1, 255, 130, 0, 1, 2, 1,
		5, 86, 97, 108, 117, 101, 1, 16, 0, 1, 3, 85, 73, 68, 1, 6, 0, 0, 0, 26,
		255, 130, 1, 6, 115, 116, 114, 105, 110, 103, 12, 3, 0, 1, 49, 1, 248,
		170, 47, 203, 96, 245, 17, 199, 128, 0,
	}

	item, err := data.DecodeItem(encData)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Object type: %T\n", item)
	fmt.Printf("Value: %v (Type: %T)\n", item.Value, item.Value)
	fmt.Printf("UID: %v\n", item.UID)
	fmt.Println("Is valid UID:", item.IsValidUID())

	// Output:
	// Object type: *data.Item
	// Value: 1 (Type: string)
	// UID: 12263243927642032000
	// Is valid UID: true
}

// ============================================================================
//  Methods
// ============================================================================

// ----------------------------------------------------------------------------
//  Item
// ----------------------------------------------------------------------------

func ExampleItem() {
	ExitOnError := func(err error) {
		if err != nil {
			log.Fatal(err)
		}
	}

	item, err := data.NewItem("1")
	ExitOnError(err)

	// Item.Encoded() returns the item object in gob encoding format.
	encData, err := item.Encoded()
	ExitOnError(err)

	decItem, err := data.DecodeItem(encData)
	ExitOnError(err)

	fmt.Printf("Object type: %T\n", decItem)
	fmt.Printf("Value: %v (Type: %T)\n", decItem.Value, decItem.Value)
	fmt.Printf("UID: %v\n", decItem.UID)

	// Output:
	// Object type: *data.Item
	// Value: 1 (Type: string)
	// UID: 12263243927642032000
}

// ----------------------------------------------------------------------------
//  Item.Encoded()
// ----------------------------------------------------------------------------

func ExampleItem_Encoded_must_be_created_via_constructor() {
	item := data.Item{}

	if _, err := item.Encoded(); err != nil {
		fmt.Println(err)
	}

	// Output: failed to gob encode the value: nil value given
}

// ----------------------------------------------------------------------------
//  Item.IsValidUID()
// ----------------------------------------------------------------------------

func ExampleItem_IsValidUID() {
	item, err := data.NewItem("1")
	if err != nil {
		log.Fatal(err)
	}

	if item.IsValidUID() {
		fmt.Println("UID is valid")
	}
	// Output: UID is valid
}
