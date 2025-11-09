package main

import (
	"fmt"
	"os"

	"github.com/anish/jotr/cmd"
)

const Version = "1.0.0"

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

