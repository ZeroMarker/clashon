package main

import (
	"os"

	"github.com/ZeroMarker/clashon/cli/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
