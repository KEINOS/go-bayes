package inmemory

import (
	"bytes"
	"encoding/gob"
	"io/fs"
	"os"

	"github.com/pkg/errors"
)

const (
	Mode644 = fs.FileMode(0o644) // Permissions for new files. -rw-r--r--
)

// DB is the in-memory key-value store.
// It is an implementation of the kvs.KVS interface.
type DB struct {
	Data  map[uint64]any
	Keys  []uint64
	Index int
}

func New() *DB {
	registGob() // regist DB struct to gob.

	return new(DB)
}

func (d *DB) keyUpdate() {
	keys := make(map[uint64]bool)
	list := []uint64{}

	for _, key := range d.Keys {
		if _, ok := keys[key]; ok {
			continue
		}

		// Append only if the key exists in the map.
		if _, ok := d.Data[key]; ok {
			keys[key] = true
			list = append(list, key)
		}
	}

	d.Keys = list
}

// Delete deletes the value for the given key. It will not error if the key is not found.
func (d *DB) Delete(key uint64) error {
	if _, ok := d.Data[key]; !ok {
		return nil
	}

	// Delete from the map.
	delete(d.Data, key)

	// Delete from the keys.
	d.keyUpdate()

	return nil
}

// Keys returns the keys in the KVS.
func (d *DB) GetKeys() []uint64 {
	d.keyUpdate()

	return d.Keys
}

// GetValue returns the value for the given key.
func (d *DB) GetValue(key uint64) (any, error) {
	v, ok := d.Data[key]
	if !ok {
		return nil, errors.Errorf("key %d not found", key)
	}

	return v, nil
}

// HasNext returns true if there is a next element.
func (d *DB) HasNext() bool {
	return d.Index < len(d.Keys)
}

// Load loads the KVS data from a file.
func (d *DB) Load(filePath string) error {
	registGob()

	byteData, err := os.ReadFile(filePath)
	if err != nil {
		return errors.Wrap(err, "failed to read file")
	}

	buf := bytes.NewBuffer(byteData)

	if err := gob.NewDecoder(buf).Decode(&d); err != nil {
		return errors.Wrap(err, "can not decode gob data from: "+filePath)
	}

	d.Reset()

	return nil
}

// Next returns the next key-value pair.
func (d *DB) Next() (uint64, any, error) {
	if !d.HasNext() {
		return 0, nil, errors.New("no more elements")
	}

	key := d.Keys[d.Index]

	value, ok := d.Data[key]
	if !ok {
		return 0, nil, errors.Errorf("key %d not found", key)
	}

	d.Index++

	return key, value, nil
}

// RegistGob registers the given struct to the gob registry.
func (d *DB) RegistGob(value any) {
	gob.Register(value)
}

// Reset resets the iterator to the first element.
func (d *DB) Reset() {
	d.keyUpdate()

	d.Index = 0
}

// gobNewEncoder is a copy of gob.NewEncoder to ease testing.
var gobNewEncoder = gob.NewEncoder

// Save saves the KVS data to a file.
func (d *DB) Save(filePath string) error {
	registGob()

	w, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, Mode644)
	if err != nil {
		return errors.Wrap(err, "failed to open file")
	}

	defer w.Close()

	if err := gobNewEncoder(w).Encode(d); err != nil {
		return errors.Wrap(err, "can not encode to gob")
	}

	return nil
}

// Set sets the value for the given key. It will overwrite the existing value.
func (d *DB) Set(key uint64, value any) error {
	if d.Data == nil {
		d.Data = make(map[uint64]any)
	}

	d.Data[key] = value

	// Append blindly even if the key is already in the list.
	d.Keys = append(d.Keys, key)

	d.keyUpdate() // update the Keys to be unique. This is needed to keep the order of the keys.

	return nil
}
