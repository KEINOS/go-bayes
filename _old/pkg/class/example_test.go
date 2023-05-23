package class_test

import (
	"fmt"
	"log"

	"github.com/KEINOS/go-bayes/pkg/class"
)

// ============================================================================
//  Functions
// ============================================================================

// ----------------------------------------------------------------------------
//  AnyToHash
// ----------------------------------------------------------------------------

func ExampleAnyToHash() {
	for _, tt := range []any{
		5963,
		true,
		false,
		-1,
		"foo",
		[]byte("bar"),
	} {
		h, err := class.AnyToHash(tt)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Len: %v, Hash: %x\n", len(h), h)
	}
	// Output:
	// Len: 64, Hash: 5fd421ff34e35be3fbbb54427dbb7736909cca3923df216f228fb76a3e165fb0567d8ce0b578d7c5db391da7c6741437b735765b4b41a34b68c1b49c5e271031
	// Len: 64, Hash: 1c5689fdf85b81b8901a83feeda69603a725bb5e413b615b66cd70e8d26842167492f22a249a0f116ce4a7b56e68fab6d7a289c7eaa24f4701d2b824658cf79b
	// Len: 64, Hash: 0678038b3c5221aedf79557131c206b27bbafc05ec822e7242089ef07f6b4c007170868d1ef30a8e23fd968eb34b794af504c558f21e4b3f600e1744b2ab3dfd
	// Len: 64, Hash: 242acd2edc6eddbfda82ab24a2b2d2fcc62b9130382991be1f9d401776122babcd3edbda316213e6006d87ca74716ad8d103af2b873287ef80ebfc8288e4e94e
	// Len: 64, Hash: feb944c4f37c5c3849960d374496864bdc52c586cf5b5e617ec16a280b3a755b96591d4b89ca76bf3f09c0c514532569c9ff033f21e21584f6c1ae412716e5b9
	// Len: 64, Hash: 0320c85195a510dc335cefd870a3f8f080f5ee71a2e73cf20782a913249be4c82353a67d1bd679a1cc434b3235918570904b9b1371f6d7e74039933cf44ec3ce
}

// ----------------------------------------------------------------------------
//  AnyToClassID
// ----------------------------------------------------------------------------

func ExampleAnyToClassID() {
	// List of supported types
	for _, v := range []any{
		"foo",         // string. The ID will be the hash value of the string.
		[]byte("bar"), // []byte. The ID will be the hash value of the byte array.
		true,          // bool. The ID will be 18446744073709551615.
		false,         // bool. The ID will be 0.
		100,           // int
		-100,          // int
		int(1),        // int
		int(-1),       // int
		uint(1),       // uint
		int16(1),      // int16
		int16(-1),     // int16
		uint16(1),     // uint16
		int32(1),      // int32
		uint32(1),     // uint32
		int64(1),      // int64
		uint64(1),     // uint64
		float32(1.5),  // float32
		float64(1.6),  // float64
	} {
		classID, err := class.AnyToClassID(v)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Value: %v (%T), ID: %v (%T)\n", v, v, classID, classID)
	}

	// Output:
	// Value: foo (string), ID: 18354777369014459540 (uint64)
	// Value: [98 97 114] ([]uint8), ID: 225400234097053844 (uint64)
	// Value: true (bool), ID: 2041971204945576328 (uint64)
	// Value: false (bool), ID: 466126457980199386 (uint64)
	// Value: 100 (int), ID: 15939380025483151630 (uint64)
	// Value: -100 (int), ID: 16649709537238761602 (uint64)
	// Value: 1 (int), ID: 4518803544265741939 (uint64)
	// Value: -1 (int), ID: 2606120935537696001 (uint64)
	// Value: 1 (uint), ID: 6971089086649022800 (uint64)
	// Value: 1 (int16), ID: 4025381293389327996 (uint64)
	// Value: -1 (int16), ID: 11551872110596953430 (uint64)
	// Value: 1 (uint16), ID: 14069011374507549038 (uint64)
	// Value: 1 (int32), ID: 11525237366295240881 (uint64)
	// Value: 1 (uint32), ID: 15668921813877113886 (uint64)
	// Value: 1 (int64), ID: 16914795258483526914 (uint64)
	// Value: 1 (uint64), ID: 4346251990183516104 (uint64)
	// Value: 1.5 (float32), ID: 8221774876508395890 (uint64)
	// Value: 1.6 (float64), ID: 13421046981767916794 (uint64)
}

func ExampleGetSalt() {
	// List of supported types
	for _, v := range []any{
		int(1),        // int
		int16(1),      // int16
		int32(1),      // int32
		int64(1),      // int64
		uint(1),       // uint
		uint16(1),     // uint16
		uint32(1),     // uint32
		uint64(1),     // uint64
		float32(1.5),  // float32
		float64(1.6),  // float64
		"foo",         // string
		[]byte("bar"), // []byte
		true,          // bool
	} {
		saltID, err := class.GetSalt(v)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Input type: %T -> SaltID: %v (Type: %T)\n", v, saltID, saltID)
	}
	// Output:
	// Input type: int -> SaltID: 1 (Type: uint8)
	// Input type: int16 -> SaltID: 2 (Type: uint8)
	// Input type: int32 -> SaltID: 3 (Type: uint8)
	// Input type: int64 -> SaltID: 4 (Type: uint8)
	// Input type: uint -> SaltID: 5 (Type: uint8)
	// Input type: uint16 -> SaltID: 6 (Type: uint8)
	// Input type: uint32 -> SaltID: 7 (Type: uint8)
	// Input type: uint64 -> SaltID: 8 (Type: uint8)
	// Input type: float32 -> SaltID: 9 (Type: uint8)
	// Input type: float64 -> SaltID: 10 (Type: uint8)
	// Input type: string -> SaltID: 11 (Type: uint8)
	// Input type: []uint8 -> SaltID: 12 (Type: uint8)
	// Input type: bool -> SaltID: 13 (Type: uint8)
}

func ExampleNew() {
	// List of supported types
	for _, v := range []interface{}{
		"foo",         // string. The ID will be the hash value of the string.
		[]byte("bar"), // []byte. The ID will be the hash value of the byte array.
		true,          // bool. The ID will be 18446744073709551615.
		false,         // bool. The ID will be 0.
		100,           // int
		-100,          // int
		int(1),        // int
		uint(1),       // uint
		int16(1),      // int16
		uint16(1),     // uint16
		int32(1),      // int32
		uint32(1),     // uint32
		int64(1),      // int64
		uint64(1),     // uint64
		float32(1),    // float32
		float64(1),    // float64
	} {
		c, err := class.New(v)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf(
			"Input Value: %v, Type: %T, Class ID: %v (HexID: %v))\n",
			c.Raw,
			c.Raw,
			c.ID,
			c.HexID(),
		)
	}

	// Unsupported type will return an error.
	_, err := class.New(struct{}{})
	if err != nil {
		fmt.Println("Unsupported type will return an error")
	}

	// Output:
	// Input Value: foo, Type: string, Class ID: 18354777369014459540 (HexID: feb944c4f37c5c94))
	// Input Value: [98 97 114], Type: []uint8, Class ID: 225400234097053844 (HexID: 0320c85195a51094))
	// Input Value: true, Type: bool, Class ID: 2041971204945576328 (HexID: 1c5689fdf85b8188))
	// Input Value: false, Type: bool, Class ID: 466126457980199386 (HexID: 0678038b3c5221da))
	// Input Value: 100, Type: int, Class ID: 15939380025483151630 (HexID: dd340daec6e1490e))
	// Input Value: -100, Type: int, Class ID: 16649709537238761602 (HexID: e70fa69c348cc082))
	// Input Value: 1, Type: int, Class ID: 4518803544265741939 (HexID: 3eb603e1c1bc5673))
	// Input Value: 1, Type: uint, Class ID: 6971089086649022800 (HexID: 60be489703513950))
	// Input Value: 1, Type: int16, Class ID: 4025381293389327996 (HexID: 37dd06f57df92e7c))
	// Input Value: 1, Type: uint16, Class ID: 14069011374507549038 (HexID: c33f2b4de673e16e))
	// Input Value: 1, Type: int32, Class ID: 11525237366295240881 (HexID: 9ff1de52421620b1))
	// Input Value: 1, Type: uint32, Class ID: 15668921813877113886 (HexID: d973315f4983c41e))
	// Input Value: 1, Type: int64, Class ID: 16914795258483526914 (HexID: eabd6cab4e5f4d02))
	// Input Value: 1, Type: uint64, Class ID: 4346251990183516104 (HexID: 3c50fd285b3917c8))
	// Input Value: 1, Type: float32, Class ID: 14094205019098768387 (HexID: c398acca62e8e403))
	// Input Value: 1, Type: float64, Class ID: 7938632105198602139 (HexID: 6e2bafd6c8dfc79b))
	// Unsupported type will return an error
}

// ============================================================================
//  Class and Methods
// ============================================================================

// ----------------------------------------------------------------------------
//  Class
// ----------------------------------------------------------------------------

func ExampleClass_HexID() {
	for _, id := range []uint64{
		0,
		255,
		0xffffffffffffffff,
		0xfffffffffffffff,  // should be equal as below
		0x0fffffffffffffff, // should be equal as above
	} {
		c, err := class.New(id)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(c.HexID())
	}
	// Output:
	// 167a18361c56a8ac
	// c025b50f8be99ca5
	// 66e44d5d75abc0cb
	// 61766c7f90e02d55
	// 61766c7f90e02d55
}

func ExampleClass_String() {
	for _, tt := range []any{
		5963,
		true,
		false,
		-1,
		"foo",
		[]byte("bar"),
	} {
		c, err := class.New(tt)
		if err != nil {
			log.Fatal(err)
		}

		// Class object supports stringer interface.
		fmt.Println(c, c.String())
	}
	// Output:
	// 5963 5963
	// true true
	// false false
	// -1 -1
	// foo foo
	// [98 97 114] [98 97 114]
}
