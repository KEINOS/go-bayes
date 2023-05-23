package data

import (
	"bytes"
	"encoding/gob"

	"github.com/KEINOS/go-bayes/pkg/utils"
	"github.com/pkg/errors"
)

// -----------------------------------------------------------------------------
//  Type: Item
// -----------------------------------------------------------------------------

type Item struct {
	Value any
	UID   uint64
}

// -----------------------------------------------------------------------------
//  Constructor
// -----------------------------------------------------------------------------

// NewItem returns a new Item object with the given value and UID.
func NewItem(input any) (*Item, error) {
	uid, err := utils.AnyToUint64(input)
	if err != nil {
		return nil, err
	}

	item := Item{
		Value: input,
		UID:   uid,
	}

	return &item, nil
}

// DecodeItem returns the decoded Item object from the given gob encoded data.
//
// The encoded data can be obtained from the Item.Encoded() method.
func DecodeItem(encoded []byte) (*Item, error) {
	var item Item

	dec := gob.NewDecoder(bytes.NewReader(encoded))

	err := dec.Decode(&item)
	if err != nil {
		return nil, errors.Wrap(err, "failed to gob decode the value")
	}

	return &item, nil
}

// -----------------------------------------------------------------------------
//  Methods
// -----------------------------------------------------------------------------

// Encoded returns the gob encoded data of itself.
func (i *Item) Encoded() ([]byte, error) {
	var buf bytes.Buffer

	enc := gob.NewEncoder(&buf)

	item := Item{
		Value: i.Value,
		UID:   i.UID,
	}

	err := enc.Encode(item)
	if err != nil || i.Value == nil {
		if err == nil {
			err = errors.New("nil value given")
		}

		return nil, errors.Wrap(err, "failed to gob encode the value")
	}

	return buf.Bytes(), nil
}

// IsValidUID returns true if the UID set is equivelent to the calculated UID
// from the current item value.
func (i Item) IsValidUID() bool {
	uid, err := utils.AnyToUint64(i.Value)
	if err != nil {
		return false
	}

	return uid == i.UID
}
