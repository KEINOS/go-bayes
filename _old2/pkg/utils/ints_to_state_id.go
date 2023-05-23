package utils

// IntsToStateID returns the state ID from the given slice of state IDs.
//
// It is used to label a transition as a state ID. It converts the state IDs to
// a single unique state ID.
func IntsToStateID(stateIDs []uint64) uint64 {
	salt := AnyToTypeID(stateIDs)
	listBytes := []byte{}

	for _, stateID := range stateIDs {
		stateIDBytes := IntToBytes(int64(stateID))

		listBytes = append(listBytes, stateIDBytes...)
	}
	// Append salt
	listBytes = append(listBytes, salt)

	hashed := BytesToHash(listBytes)

	return BytesToStateID(hashed)
}
