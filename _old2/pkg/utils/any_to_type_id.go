package utils

// -----------------------------------------------------------------------------
//  Type: TypeID
// -----------------------------------------------------------------------------

// TypeID is the type ID of the input value.
type TypeID int

// List of supported type IDs.
const (
	// UnknownType represents an unknown type.
	UnknownType     TypeID = iota
	INTType                // INTType represents an int type.
	INT8Type               // INT8Type represents an int8 type.
	INT16Type              // INT16Type represents an int16 type.
	INT32Type              // INT32Type represents an int32 type.
	INT64Type              // INT64Type represents an int64 type.
	UINTType               // UINTType represents an uint type.
	UINT8Type              // UTIN8Type represents an uint8/a byte type.
	UINT16Type             // UINT16Type represents an uint16 type.
	UINT32Type             // UINT32Type represents an uint32 type.
	UINT64Type             // UINT64Type represents an uint64 type.
	FLOAT32Type            // FLOAT32Type represents a float32 type.
	FLOAT64Type            // FLOAT64Type represents a float64 type.
	StringType             // StringType represents a string type.
	BytesType              // BytesType represents a byte slice type.
	BoolType               // BoolType represents a bool type.
	SliceStringType        // SliceStringType represents a string slice type.
	SliceIntType           // SliceIntType represents an int slice type.
)

// -----------------------------------------------------------------------------
//  AnyToTypeID()
// -----------------------------------------------------------------------------

// AnyToTypeID returns the ID of the type representing the input value as uint8.
//
// This value is used as a salt to generate a unique hash from the same value but
// different type.
//nolint: gocyclo,cyclop // ignore cyclomatic complexity for readability
func AnyToTypeID(input any) byte {
	switch input.(type) {
	case int:
		return uint8(INTType)
	case int8:
		return uint8(INT8Type)
	case int16:
		return uint8(INT16Type)
	case int32:
		return uint8(INT32Type)
	case int64:
		return uint8(INT64Type)
	case uint:
		return uint8(UINTType)
	case uint8:
		return uint8(UINT8Type)
	case uint16:
		return uint8(UINT16Type)
	case uint32:
		return uint8(UINT32Type)
	case uint64:
		return uint8(UINT64Type)
	case float32:
		return uint8(FLOAT32Type)
	case float64:
		return uint8(FLOAT64Type)
	case string:
		return uint8(StringType)
	case []byte:
		return uint8(BytesType)
	case bool:
		return uint8(BoolType)
	case []string:
		return uint8(SliceStringType)
	case []int:
		return uint8(SliceIntType)
	}

	return uint8(UnknownType) // zero = unknown/unsupported type
}
