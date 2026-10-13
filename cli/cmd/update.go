package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ZeroMarker/clashon/sdk/update"
)

func updateCmd() *cobra.Command {
	c := &cobra.Command{Use: "update", Short: "Check for a newer ClashOn release"}
	c.AddCommand(&cobra.Command{
		Use:   "check",
		Short: "Compare this build with the latest GitHub release",
		RunE: func(cmd *cobra.Command, _ []string) error {
			r, err := update.Latest(cmd.Context(), nil, "")
			if err != nil {
				return err
			}
			out := map[string]any{"current": Version, "newer": false}
			if r != nil {
				out["latest"], out["url"] = r.Version, r.URL
				out["newer"] = update.Newer(r.Version, Version)
			}
			if jsonOut {
				return printJSON(out)
			}
			switch {
			case r == nil:
				fmt.Println("no releases published yet")
			case out["newer"] == true:
				fmt.Printf("ClashOn %s is available (you have %s): %s\n", r.Version, Version, r.URL)
			default:
				fmt.Printf("up to date (%s)\n", Version)
			}
			return nil
		},
	})
	return c
}
