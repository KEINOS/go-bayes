package main

import (
	"strings"
	"testing"
)

func TestReadMushrooms_rejectsInvalidData(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		"empty dataset": "",
		"empty feature": "p,,s,n,t,p,f,c,n,k,e,e,s,s,w,w,p,w,o,p,k,s,u\n",
		"long feature":  "p,xx,s,n,t,p,f,c,n,k,e,e,s,s,w,w,p,w,o,p,k,s,u\n",
		"missing field": "p,x,s,n\n",
		"unknown class": "q,x,s,n,t,p,f,c,n,k,e,e,s,s,w,w,p,w,o,p,k,s,u\n",
	}

	for name, input := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, err := readMushrooms(strings.NewReader(input))
			if err == nil {
				t.Fatal("readMushrooms() error = nil, want an error")
			}
		})
	}
}
