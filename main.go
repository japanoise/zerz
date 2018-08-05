package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: %s [files]\n", os.Args[0])
		os.Exit(1)
	}

	global, errors := InitEditor(os.Args[1:])
	if len(global.Buffers) == 0 {
		fmt.Fprintln(os.Stderr, "All files had errors.")
		for _, err := range errors {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(1)
	}

	zerz(global, errors)
}
