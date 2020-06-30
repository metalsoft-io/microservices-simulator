package main

import (
	"testing"
)

func TestGenIndex(t *testing.T) {
	arr := generateRandomUniqueIntegers(10)

	if len(arr) != 10 {
		t.Fatal("len should be 10")
	}

	for i := 0; i < len(arr); i++ {
		for j := i; j < len(arr); j++ {
			if i != j && arr[i] == arr[j] {
				t.Fatal("not unique")
			}
		}
	}
}

func TestGenChain(t *testing.T) {
	idx := []string{"a", "b", "c"}
	chain := generateChain(idx, 2)
	if len(chain) != 2 {
		t.Fatal("len should be 2")
	}
}
