package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ZeroMarker/clashon/sdk/sysproxy"
)

func sysproxyCmd() *cobra.Command {
	return &cobra.Command{
		Use:       "sysproxy [on|off]",
		Short:     "Show or set the OS proxy to point at clashon",
		Args:      cobra.MatchAll(cobra.MaximumNArgs(1), cobra.OnlyValidArgs),
		ValidArgs: []string{"on", "off"},
		RunE: func(cmd *cobra.Command, args []string) error {
			e, err := loadEnv()
			if err != nil {
				return err
			}
			const host = "127.0.0.1"
			port := e.Settings.MixedPort
			if len(args) == 1 {
				if args[0] == "on" {
					err = sysproxy.Enable(host, port)
				} else {
					err = sysproxy.Disable()
				}
				if err != nil {
					return err
				}
			}
			on, err := sysproxy.Enabled(host, port)
			if errors.Is(err, sysproxy.ErrUnsupported) {
				on, err = false, nil
			}
			if err != nil {
				return err
			}
			if jsonOut {
				return printJSON(map[string]bool{"enabled": on})
			}
			state := "off"
			if on {
				state = fmt.Sprintf("on (%s:%d)", host, port)
			}
			fmt.Println("system proxy:", state)
			return nil
		},
	}
}
