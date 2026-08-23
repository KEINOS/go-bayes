package main

import (
	"strings"
	"testing"
)

func TestReadWine_rejectsInvalidData(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		"empty class":         ",14.23,1.71,2.43,15.6,127,2.8,3.06,.28,2.29,5.64,1.04,3.92,1065\n",
		"empty dataset":       "",
		"invalid measurement": "1,not-a-number,1.71,2.43,15.6,127,2.8,3.06,.28,2.29,5.64,1.04,3.92,1065\n",
		"missing field":       "1,14.23,1.71,2.43\n",
		"unknown class":       "4,14.23,1.71,2.43,15.6,127,2.8,3.06,.28,2.29,5.64,1.04,3.92,1065\n",
	}

	for name, input := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, err := readWine(strings.NewReader(input))
			if err == nil {
				t.Fatal("readWine() error = nil, want an error")
			}
		})
	}
}
