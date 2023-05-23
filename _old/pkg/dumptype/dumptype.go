package dumptype

// DumpType is the type of format to store the trained data.
type DumpType int

const (
	// Unknown represents an unknown dump type.
	Unknown DumpType = iota
	// JSON represents dump type as JSON.
	JSON
	// CSV represents dump type as CSV.
	CSV
	// SQL represents dump type as SQL.
	SQL
	// GOB represents dump type as gob encoding, an exchangeable binary values of Go.
	GOB
)

// String is an implementation of Stringer interface.
func (d DumpType) String() string {
	switch d {
	case JSON:
		return "json"
	case CSV:
		return "csv"
	case SQL:
		return "sql"
	case GOB:
		return "gob"
	case Unknown:
		fallthrough
	default:
		return "unknown"
	}
}
