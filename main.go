package main

import (
	"os"

	"github.com/wameedh/ccflow/cmd/ccflow"
)

func main() {
	if err := ccflow.Execute(); err != nil {
		os.Exit(1)
	}
}
