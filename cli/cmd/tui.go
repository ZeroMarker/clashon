package cmd

import (
	"github.com/spf13/cobra"

	"github.com/ZeroMarker/clashon/sdk/app"
	"github.com/ZeroMarker/clashon/tui"
)

func tuiCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "tui",
		Short: "Open the terminal UI",
		RunE: func(cmd *cobra.Command, _ []string) error {
			a, err := app.RequireCore()
			if err != nil {
				return err
			}
			return tui.Run(a)
		},
	}
}
