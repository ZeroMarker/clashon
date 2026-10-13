package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/ZeroMarker/clashon/sdk/config"
	"github.com/ZeroMarker/clashon/sdk/deeplink"
	"github.com/ZeroMarker/clashon/sdk/privilege"
)

func tunCmd() *cobra.Command {
	return &cobra.Command{
		Use:       "tun [on|off]",
		Short:     "Show or toggle TUN mode (route all traffic, needs privileges)",
		Args:      cobra.MatchAll(cobra.MaximumNArgs(1), cobra.OnlyValidArgs),
		ValidArgs: []string{"on", "off"},
		RunE: func(cmd *cobra.Command, args []string) error {
			e, err := loadEnv()
			if err != nil {
				return err
			}
			if len(args) == 1 {
				on := args[0] == "on"
				if on {
					if err := privilege.Check(); err != nil {
						return err
					}
				}
				e.Settings.TUN = on
				if err := config.SaveSettings(e.Layout.Settings(), e.Settings); err != nil {
					return err
				}
				if err := e.reloadIfRunning(cmd); err != nil {
					return err
				}
			}
			ready := privilege.Check() == nil
			if jsonOut {
				return printJSON(map[string]bool{"enabled": e.Settings.TUN, "ready": ready})
			}
			state := "off"
			if e.Settings.TUN {
				state = "on"
			}
			fmt.Println("tun:", state)
			if !ready {
				fmt.Println("note:", privilege.ErrMissing)
			}
			return nil
		},
	}
}

func serviceCmd() *cobra.Command {
	c := &cobra.Command{Use: "service", Short: "Set up privileges for TUN mode"}
	exe := func() (string, error) {
		p, err := os.Executable()
		if err != nil {
			return "", err
		}
		return filepath.EvalSymlinks(p)
	}
	c.AddCommand(&cobra.Command{
		Use:   "install",
		Short: "Grant this binary TUN privileges (run with sudo)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			p, err := exe()
			if err != nil {
				return err
			}
			if err := privilege.Grant(p); err != nil {
				return err
			}
			fmt.Println("granted TUN privileges to", p)
			return nil
		},
	}, &cobra.Command{
		Use:   "uninstall",
		Short: "Remove TUN privileges (run with sudo)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			p, err := exe()
			if err != nil {
				return err
			}
			return privilege.Revoke(p)
		},
	})
	return c
}

func importCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "import <clash://install-config?url=...>",
		Short: "Add a subscription from a clash:// link",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			in, err := deeplink.Parse(args[0])
			if err != nil {
				return err
			}
			e, err := loadEnv()
			if err != nil {
				return err
			}
			// Keep names unique so repeated imports do not fail.
			name := in.Name
			for i := 2; ; i++ {
				if _, err := e.Store.Get(name); err != nil {
					break
				}
				name = fmt.Sprintf("%s-%d", in.Name, i)
			}
			p, err := e.Store.Add(cmd.Context(), name, in.URL)
			if err != nil {
				return err
			}
			if jsonOut {
				return printJSON(map[string]string{"name": p.Name})
			}
			fmt.Printf("added %s\n", p.Name)
			return nil
		},
	}
}
