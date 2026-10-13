package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ZeroMarker/clashon/sdk/desktop"
)

// desktopCmd holds OS integration used by the GUI.
func desktopCmd() *cobra.Command {
	c := &cobra.Command{Use: "desktop", Short: "OS integration for the GUI (autostart, clash:// links)"}

	var exe string
	autostart := &cobra.Command{
		Use:       "autostart [on|off]",
		Short:     "Show or set launching the GUI at login",
		Args:      cobra.MatchAll(cobra.MaximumNArgs(1), cobra.OnlyValidArgs),
		ValidArgs: []string{"on", "off"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				if args[0] == "on" && exe == "" {
					return fmt.Errorf("--exec is required")
				}
				if err := desktop.SetAutostart(exe, args[0] == "on"); err != nil {
					return err
				}
			}
			on, err := desktop.Autostart()
			if err != nil {
				return err
			}
			if jsonOut {
				return printJSON(map[string]bool{"enabled": on})
			}
			fmt.Println("autostart:", map[bool]string{true: "on", false: "off"}[on])
			return nil
		},
	}
	autostart.Flags().StringVar(&exe, "exec", "", "path to the GUI executable")

	var schemeExe string
	register := &cobra.Command{
		Use:   "register",
		Short: "Register the GUI as the clash:// link handler",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if schemeExe == "" {
				return fmt.Errorf("--exec is required")
			}
			return desktop.RegisterSchemes(schemeExe)
		},
	}
	register.Flags().StringVar(&schemeExe, "exec", "", "path to the GUI executable")

	c.AddCommand(autostart, register)
	return c
}
