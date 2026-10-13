package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
)

func subCmd() *cobra.Command {
	c := &cobra.Command{Use: "sub", Short: "Manage subscriptions"}

	var name string
	add := &cobra.Command{
		Use:   "add <url>",
		Short: "Download a subscription",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			e, err := loadEnv()
			if err != nil {
				return err
			}
			p, err := e.store.Add(cmd.Context(), name, args[0])
			if err != nil {
				return err
			}
			fmt.Printf("added %s\n", p.Name)
			return nil
		},
	}
	add.Flags().StringVarP(&name, "name", "n", "default", "profile name")

	ls := &cobra.Command{
		Use:     "ls",
		Aliases: []string{"list"},
		Short:   "List subscriptions",
		RunE: func(cmd *cobra.Command, _ []string) error {
			e, err := loadEnv()
			if err != nil {
				return err
			}
			ps, cur, err := e.store.List()
			if err != nil {
				return err
			}
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "\tNAME\tUPDATED\tUSED/TOTAL\tEXPIRE")
			for _, p := range ps {
				mark, usage, expire := "", "-", "-"
				if p.Name == cur {
					mark = "*"
				}
				if u := p.UserInfo; u != nil {
					usage = fmt.Sprintf("%s/%s", bytes(u.Upload+u.Download), bytes(u.Total))
					if !u.Expire.IsZero() {
						expire = u.Expire.Format("2006-01-02")
					}
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", mark, p.Name, p.Updated.Format(time.DateTime), usage, expire)
			}
			return w.Flush()
		},
	}

	update := &cobra.Command{
		Use:   "update [name...]",
		Short: "Re-download subscriptions (all if none given) and reload the core",
		RunE: func(cmd *cobra.Command, args []string) error {
			e, err := loadEnv()
			if err != nil {
				return err
			}
			if len(args) == 0 {
				ps, _, err := e.store.List()
				if err != nil {
					return err
				}
				for _, p := range ps {
					args = append(args, p.Name)
				}
			}
			for _, n := range args {
				if _, err := e.store.Update(cmd.Context(), n); err != nil {
					return fmt.Errorf("%s: %w", n, err)
				}
				fmt.Printf("updated %s\n", n)
			}
			return e.reloadIfRunning(cmd)
		},
	}

	use := &cobra.Command{
		Use:   "use <name>",
		Short: "Switch the current subscription and reload the core",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			e, err := loadEnv()
			if err != nil {
				return err
			}
			if err := e.store.Use(args[0]); err != nil {
				return err
			}
			fmt.Printf("using %s\n", args[0])
			return e.reloadIfRunning(cmd)
		},
	}

	rm := &cobra.Command{
		Use:   "rm <name>",
		Short: "Remove a subscription",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			e, err := loadEnv()
			if err != nil {
				return err
			}
			return e.store.Remove(args[0])
		},
	}

	c.AddCommand(add, ls, update, use, rm)
	return c
}

func (e *env) reloadIfRunning(cmd *cobra.Command) error {
	if _, err := e.daemon.PID(); err != nil {
		return nil
	}
	path, err := e.writeRuntimeConfig()
	if err != nil {
		return err
	}
	if err := e.client.Reload(cmd.Context(), path); err != nil {
		return err
	}
	fmt.Println("core reloaded")
	return nil
}

func bytes(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%dB", n)
	}
	div, exp := int64(unit), 0
	for m := n / unit; m >= unit; m /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%ciB", float64(n)/float64(div), "KMGTPE"[exp])
}
