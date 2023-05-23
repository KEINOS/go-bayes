package logmem_test

import (
	"bytes"
	"fmt"
	"log"

	"github.com/KEINOS/go-bayes/pkg/dumptype"
	"github.com/KEINOS/go-bayes/pkg/nodelogger/logmem"
)

// ----------------------------------------------------------------------------
//  NodeLog.Restore()
// ----------------------------------------------------------------------------

func ExampleNodeLog_Restore_gob() {
	logMem := logmem.New(5963)

	fmt.Println("[Before] Restored node ID:", logMem.ID())
	fmt.Println("[Before] Restored total accesses:", logMem.TotalAccesses)

	// Gob encoded data.
	storedData := []byte{
		97, 255, 129, 3, 1, 1, 7, 78, 111, 100, 101, 76, 111, 103, 1, 255, 130,
		0, 1, 6, 1, 8, 67, 108, 97, 115, 115, 77, 97, 112, 1, 16, 0, 1, 8, 70,
		114, 111, 109, 65, 84, 111, 66, 1, 255, 134, 0, 1, 5, 70, 114, 111, 109,
		65, 1, 255, 132, 0, 1, 3, 84, 111, 66, 1, 255, 132, 0, 1, 6, 78, 111,
		100, 101, 73, 68, 1, 6, 0, 1, 13, 84, 111, 116, 97, 108, 65, 99, 99, 101,
		115, 115, 101, 115, 1, 4, 0, 0, 0, 42, 255, 133, 4, 1, 1, 25, 109, 97,
		112, 91, 117, 105, 110, 116, 54, 52, 93, 109, 97, 112, 91, 117, 105, 110,
		116, 54, 52, 93, 105, 110, 116, 1, 255, 134, 0, 1, 6, 1, 255, 132, 0, 0,
		14, 255, 131, 4, 1, 2, 255, 132, 0, 1, 6, 1, 4, 0, 0, 76, 255, 130, 1,
		52, 103, 105, 116, 104, 117, 98, 46, 99, 111, 109, 47, 75, 69, 73, 78,
		79, 83, 47, 103, 111, 45, 98, 97, 121, 101, 115, 47, 112, 107, 103, 47,
		110, 111, 100, 101, 108, 111, 103, 103, 101, 114, 47, 108, 111, 103, 109,
		101, 109, 46, 77, 97, 112, 255, 137, 4, 1, 1, 3, 77, 97, 112, 1, 255,
		138, 0, 1, 6, 1, 255, 136, 0, 0, 27, 255, 135, 3, 1, 2, 255, 136, 0, 1,
		2, 1, 3, 82, 97, 119, 1, 16, 0, 1, 2, 73, 68, 1, 6, 0, 0, 0, 255, 197,
		255, 138, 101, 0, 3, 248, 239, 218, 6, 88, 99, 5, 201, 214, 1, 6, 115,
		116, 114, 105, 110, 103, 12, 3, 0, 1, 65, 1, 248, 239, 218, 6, 88, 99,
		5, 201, 214, 0, 248, 13, 242, 114, 66, 74, 205, 186, 164, 1, 6, 115, 116,
		114, 105, 110, 103, 12, 3, 0, 1, 66, 1, 248, 13, 242, 114, 66, 74, 205,
		186, 164, 0, 248, 95, 127, 47, 204, 166, 21, 248, 80, 1, 6, 115, 116,
		114, 105, 110, 103, 12, 3, 0, 1, 67, 1, 248, 95, 127, 47, 204, 166, 21,
		248, 80, 0, 1, 2, 248, 13, 242, 114, 66, 74, 205, 186, 164, 1, 248, 95,
		127, 47, 204, 166, 21, 248, 80, 2, 248, 239, 218, 6, 88, 99, 5, 201, 214,
		1, 248, 13, 242, 114, 66, 74, 205, 186, 164, 2, 1, 2, 248, 239, 218, 6,
		88, 99, 5, 201, 214, 2, 248, 13, 242, 114, 66, 74, 205, 186, 164, 2, 1,
		2, 248, 13, 242, 114, 66, 74, 205, 186, 164, 2, 248, 95, 127, 47, 204,
		166, 21, 248, 80, 2, 1, 254, 48, 57, 1, 4, 0,
	}

	if err := logMem.Restore(dumptype.GOB, bytes.NewBuffer(storedData)); err != nil {
		log.Fatal(err)
	}

	classMap := logMem.ClassMap

	fmt.Println("[After] Restored node ID:", logMem.ID())
	fmt.Println("[After] Restored total accesses:", logMem.TotalAccesses)
	fmt.Println("[After] Restored class map:", classMap)

	// Output:
	// [Before] Restored node ID: 5963
	// [Before] Restored total accesses: 0
	// [After] Restored node ID: 12345
	// [After] Restored total accesses: 2
}

func ExampleNodeLog_Restore_json() {
	logMem := logmem.New(5963)

	fmt.Println("[Before] Restored node ID:", logMem.ID())
	fmt.Println("[Before] Restored total accesses:", logMem.TotalAccesses)

	storedData := `{
		"from_a_to_b": {"1": { "2": 1 }, "2": { "3": 1 } },
		"from_a": { "1": 1, "2": 1 },
		"to_b": { "2": 1, "3": 1 },
		"NodeID": 12345,
		"total_accesses": 2
	}`

	if err := logMem.Restore(dumptype.JSON, bytes.NewBufferString(storedData)); err != nil {
		log.Fatal(err)
	}

	fmt.Println("[After] Restored node ID:", logMem.ID())
	fmt.Println("[After] Restored total accesses:", logMem.TotalAccesses)

	// Output:
	// [Before] Restored node ID: 5963
	// [Before] Restored total accesses: 0
	// [After] Restored node ID: 12345
	// [After] Restored total accesses: 2
}

// ----------------------------------------------------------------------------
//  NodeLog.Store()
// ----------------------------------------------------------------------------

func ExampleNodeLog_Store_gob() {
	trainer := logmem.New(12345)

	classMap, err := logmem.NewMap("A", "B", "C")
	if err != nil {
		log.Fatal(err)
	}

	A, err := classMap.GetClassID("A")
	if err != nil {
		log.Fatal(err)
	}

	B, err := classMap.GetClassID("B")
	if err != nil {
		log.Fatal(err)
	}

	C, err := classMap.GetClassID("C")
	if err != nil {
		log.Fatal(err)
	}

	// Train
	trainer.Update(A, B) // transition from A to B
	trainer.Update(B, C) // transition from B to C

	// Store the trained node data to the buffer b in GOB format.
	var b bytes.Buffer

	if err := trainer.Store(dumptype.GOB, &classMap, &b); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Written bytes:", b.Len())
	fmt.Printf("Dump:\n%v\n", b.Bytes())
	// fmt.Printf("first 10 bytes: %v\n", b.Bytes()[:10])
	// fmt.Printf("last 10 bytes: %v\n", b.Bytes()[len(b.Bytes())-10:])

	// Output:
	// Written bytes: 175
	// first 10 bytes: [84 255 129 3 1 1 7 78 111 100]
	// last 10 bytes: [2 3 2 1 254 48 57 1 4 0]
}

func ExampleNodeLog_Store_json() {
	logger := logmem.New(12345)

	logger.Update(1, 2)
	logger.Update(2, 3)

	var b bytes.Buffer

	classMap, err := logmem.NewMap("")
	if err != nil {
		log.Fatal(err)
	}

	if err := logger.Store(dumptype.JSON, &classMap, &b); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Written data:", b.String())

	// Output:
	// Written data: {
	// 	"from_a_to_b": {
	// 		"1": {
	// 			"2": 1
	// 		},
	// 		"2": {
	// 			"3": 1
	// 		}
	// 	},
	// 	"from_a": {
	// 		"1": 1,
	// 		"2": 1
	// 	},
	// 	"to_b": {
	// 		"2": 1,
	// 		"3": 1
	// 	},
	// 	"NodeID": 12345,
	// 	"total_accesses": 2
	// }
}

// ----------------------------------------------------------------------------
//  NodeLog.ID()
// ----------------------------------------------------------------------------

func ExampleNodeLog_ID() {
	n := logmem.New(12345)

	fmt.Println(n.ID())

	// Output: 12345
}

// ----------------------------------------------------------------------------
//  NodeLog.Predict()
// ----------------------------------------------------------------------------

func ExampleNodeLog_Predict() {
	const (
		x = uint64(1) // Node ID of node x
		y = uint64(2) // Node ID of node y
		z = uint64(3) // Node ID of node z
	)

	nodeY := logmem.New(y) // Create a new node y

	nodeY.Update(x, z) // from x to z (x -> y -> z) This should be called from node z.
	nodeY.Update(z, x) // from z to x (z -> y -> x) This should be called from node x.
	nodeY.Update(x, z) // from x to z (x -> y -> z) This should be called from node z.
	nodeY.Update(z, x) // from z to x (z -> y -> x) This should be called from node x.

	fmt.Println("Prediction of outgoing node to be x, on incoming node as x:", nodeY.Predict(x, x))
	fmt.Println("Prediction of outgoing node to be y, on incoming node as x:", nodeY.Predict(x, y))
	fmt.Println("Prediction of outgoing node to be z, on incoming node as x:", nodeY.Predict(x, z))

	// Output:
	// Prediction of outgoing node to be x, on incoming node as x: 0
	// Prediction of outgoing node to be y, on incoming node as x: 0
	// Prediction of outgoing node to be z, on incoming node as x: 1
}

// ----------------------------------------------------------------------------
//  NodeLog.PriorPfromAtoB()
// ----------------------------------------------------------------------------

func ExampleNodeLog_PriorPfromAtoB() {
	const (
		x = uint64(1) // Node ID of node x
		y = uint64(2) // Node ID of node y
		z = uint64(3) // Node ID of node z
	)

	nodeY := logmem.New(y) // Create a new node y

	nodeY.Update(x, z) // from x to z (x -> y -> z) This should be called from node z.
	nodeY.Update(z, x) // from z to x (z -> y -> x) This should be called from node x.
	nodeY.Update(x, z) // from x to z (x -> y -> z) This should be called from node z.
	nodeY.Update(z, x) // from z to x (z -> y -> x) This should be called from node x.

	fmt.Println("Prior probability of outgoing node x, on incoming node as x (x -> y -> x):",
		nodeY.PriorPfromAtoB(x, x))
	fmt.Println("Prior probability of outgoing node y, on incoming node as x (x -> y -> y):",
		nodeY.PriorPfromAtoB(x, y))
	fmt.Println("Prior probability of outgoing node z, on incoming node as x (x -> y -> z):",
		nodeY.PriorPfromAtoB(x, z))

	// Output:
	// Prior probability of outgoing node x, on incoming node as x (x -> y -> x): 0
	// Prior probability of outgoing node y, on incoming node as x (x -> y -> y): 0
	// Prior probability of outgoing node z, on incoming node as x (x -> y -> z): 0.5
}

// ----------------------------------------------------------------------------
//  NodeLog.PriorPNotFromAtoB()
// ----------------------------------------------------------------------------

func ExampleNodeLog_PriorPNotFromAtoB() {
	nodeY := logmem.New(12345) // logger ID as 12345

	for _, t := range []struct {
		fromA uint64 // incoming node ID
		toB   uint64 // outgoing node ID
	}{
		{1, 2}, // from node #1 to #2
		{1, 3}, // from node #1 to #3
		{2, 3}, // from node #2 to #3
		{3, 1}, // from node #3 to #1
	} {
		nodeY.Update(t.fromA, t.toB)
	}

	fmt.Println("Prior probability of outgoing node not to be node #1, on incoming node as node #1:",
		nodeY.PriorPNotFromAtoB(1, 1))
	fmt.Println("Prior probability of outgoing node not to be node #1, on incoming node as node #2:",
		nodeY.PriorPNotFromAtoB(2, 1))
	fmt.Println("Prior probability of outgoing node not to be node #1, on incoming node as node #3:",
		nodeY.PriorPNotFromAtoB(3, 1))

	// Output:
	// Prior probability of outgoing node not to be node #1, on incoming node as node #1: 0.5
	// Prior probability of outgoing node not to be node #1, on incoming node as node #2: 0.25
	// Prior probability of outgoing node not to be node #1, on incoming node as node #3: 0
}

// ----------------------------------------------------------------------------
//  NodeLog.PriorPtoB()
// ----------------------------------------------------------------------------

func ExampleNodeLog_PriorPtoB() {
	nodeY := logmem.New(12345) // logger ID as 12345

	for _, t := range []struct {
		fromA uint64 // incoming node ID
		toB   uint64 // outgoing node ID
	}{
		{1, 2}, // from node #1 to #2
		{1, 3}, // from node #1 to #3
		{2, 3}, // from node #2 to #3
		{3, 1}, // from node #3 to #1
	} {
		nodeY.Update(t.fromA, t.toB)
	}

	fmt.Println("Prior probability of outgoing node to node #1:", nodeY.PriorPtoB(1))
	fmt.Println("Prior probability of outgoing node to node #2:", nodeY.PriorPtoB(2))
	fmt.Println("Prior probability of outgoing node to node #3:", nodeY.PriorPtoB(3))

	// Output:
	// Prior probability of outgoing node to node #1: 0.25
	// Prior probability of outgoing node to node #2: 0.25
	// Prior probability of outgoing node to node #3: 0.5
}

func ExampleNodeLog_String() {
	n := logmem.New(12345)

	// Stringer implementation for the NodeLog type.
	fmt.Println(n)

	// Output: 12345
}

func ExampleNodeLog_Update() {
	const (
		x = uint64(1)
		y = uint64(2)
		z = uint64(3)
	)

	nodeY := logmem.New(y)

	// Update the access log of nodeY.
	//
	// It will update the access log as "x -> y -> z".
	// Which means that node x is the predecessor of node y, and node z is the
	// successor of node y. In other words, node x is the incoming node and node
	// z is the outgoing node of node y.
	//
	// Note that it must be called by the next node accessed. In this case,
	// node z should call this function.
	nodeY.Update(x, z)

	fmt.Println("Total access:", nodeY.TotalAccesses)
	fmt.Println("Number of access from node x:", nodeY.FromA[x])
	fmt.Println("Number of access from node z:", nodeY.FromA[z])
	fmt.Println("Number of outgoing node x:", nodeY.ToB[x])
	fmt.Println("Number of outgoing node z:", nodeY.ToB[z])

	// Output:
	// Total access: 1
	// Number of access from node x: 1
	// Number of access from node z: 0
	// Number of outgoing node x: 0
	// Number of outgoing node z: 1
}

// ----------------------------------------------------------------------------
//  Map
// ----------------------------------------------------------------------------

func ExampleMap() {
	// Add items in the map at once on instantiation.
	classMap, err := logmem.NewMap(
		"foo",
		[]byte("bar"),
		1,
		true,
	)
	if err != nil {
		log.Fatal(err)
	}

	// Get the class ID of the item in the map
	classID, err := classMap.GetClassID("foo")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("ClassID of foo:", classID)

	// Add additional items in the map.
	err = classMap.AddAny(
		"foo", // duplicate items are ignored
		2,
		false,
	)
	if err != nil {
		log.Fatal(err)
	}

	// Get list of class ID of all items in the map.
	classIDs := classMap.GetKeys()

	for _, id := range classIDs {
		item, err := classMap.GetClass(id)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("ID: %v, Value: %v\n", item.ID, item.Raw)
	}

	// Output:
	// ClassID of foo: 18354777369014459540
	// ID: 225400234097053844, Value: [98 97 114]
	// ID: 466126457980199386, Value: false
	// ID: 2041971204945576328, Value: true
	// ID: 4518803544265741939, Value: 1
	// ID: 8305819738622092086, Value: 2
	// ID: 18354777369014459540, Value: foo
}

func ExampleMap_GetKeys() {
	classMap, err := logmem.NewMap(
		"foo", "bar", "baz",
	)
	if err != nil {
		log.Fatal(err)
	}

	keys := classMap.GetKeys()

	for _, key := range keys {
		item, err := classMap.GetClass(key)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Key:", key, "Value:", item)
	}

	// Output:
	// Key: 9195856179681587737 Value: bar
	// Key: 10040945619453064280 Value: baz
	// Key: 18354777369014459540 Value: foo
}
