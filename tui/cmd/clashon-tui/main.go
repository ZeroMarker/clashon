package main

import (
	"fmt"
	"os"

	"github.com/ZeroMarker/clashon/sdk/app"
	"github.com/ZeroMarker/clashon/tui"
)

func main() {
	a, err := app.RequireCore()
	if err == nil {
		err = tui.Run(a)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
