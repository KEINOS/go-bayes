package inmemory

var isGobRegistered = false

func registGob() {
	if isGobRegistered {
		return
	}

	// gob.Register(New())

	isGobRegistered = true
}
