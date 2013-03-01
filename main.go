package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func errMain() (err error) {
	return nil
}

func main() {
	err := errMain()
	if err == nil {
		os.Exit(0)
	}
	fmt.Fprintf(os.Stderr, "%v: Error: %v\n", filepath.Base(os.Args[0]), err)
	os.Exit(1)
}
