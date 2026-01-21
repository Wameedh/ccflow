package main

import (
	"os"

	"github.com/Wameedh/ccflow/cmd/ccflow"
)

func main() {
	if err := ccflow.Execute(); err != nil {
		os.Exit(1)
	}
}
