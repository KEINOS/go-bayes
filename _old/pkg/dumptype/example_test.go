package dumptype_test

import (
	"fmt"

	"github.com/KEINOS/go-bayes/pkg/dumptype"
)

func ExampleDumpType_String() {
	for _, t := range []dumptype.DumpType{
		dumptype.Unknown,
		dumptype.JSON,
		dumptype.CSV,
		dumptype.SQL,
		dumptype.GOB,
	} {
		fmt.Println(int(t), t, t.String())
	}
	// Output:
	// 0 unknown unknown
	// 1 json json
	// 2 csv csv
	// 3 sql sql
	// 4 gob gob
}
