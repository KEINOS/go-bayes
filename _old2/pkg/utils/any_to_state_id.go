package utils

func AnyToStateID(input any) (uint64, error) {
	hashed, err := AnyToHash(input)
	if err != nil {
		return 0, err
	}

	stateID := BytesToStateID(hashed)

	return stateID, nil
}
