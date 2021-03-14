package main

import (
	"crypto/rand"
	"fmt"
)

func printNewSeed() error {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return err
	}
	fmt.Printf("%X\n", b)
	return nil
}
