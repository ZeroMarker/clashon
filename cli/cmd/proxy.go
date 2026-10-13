package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/ZeroMarker/clashon/sdk/api"
)

func proxyCmd() *cobra.Command {
	c := &cobra.Command{Use: "proxy", Short: "List, select and test proxies"}

	ls := &cobra.Command{
		Use:     "ls [group]",
		Aliases: []string{"list"},
		Short:   "List groups, or the members of a group",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			e, err := loadEnv()
			if err != nil {
				return err
			}
			all, err := e.client.Proxies(cmd.Context())
			if err != nil {
				return err
			}
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			defer w.Flush()
			if len(args) == 0 {
				fmt.Fprintln(w, "GROUP\tTYPE\tNOW\tMEMBERS")
				var names []string
				for n, p := range all {
					if p.IsGroup() && n != "GLOBAL" {
						names = append(names, n)
					}
				}
				sort.Strings(names)
				for _, n := range names {
					p := all[n]
					fmt.Fprintf(w, "%s\t%s\t%s\t%d\n", n, p.Type, p.Now, len(p.All))
				}
				return nil
			}
			g, ok := all[args[0]]
			if !ok || !g.IsGroup() {
				return fmt.Errorf("no such group %q", args[0])
			}
			fmt.Fprintln(w, "\tPROXY\tTYPE\tDELAY")
			for _, n := range g.All {
				mark, typ, delay := "", "", "-"
				if n == g.Now {
					mark = "*"
				}
				if p := all[n]; p != nil {
					typ = p.Type
					if d := p.LastDelay(); d > 0 {
						delay = fmt.Sprintf("%dms", d)
					}
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", mark, n, typ, delay)
			}
			return nil
		},
	}

	use := &cobra.Command{
		Use:   "use <group> <proxy>",
		Short: "Select a proxy in a group",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			e, err := loadEnv()
			if err != nil {
				return err
			}
			if err := e.client.Select(cmd.Context(), args[0], args[1]); err != nil {
				return err
			}
			fmt.Printf("%s -> %s\n", args[0], args[1])
			return nil
		},
	}

	var testURL string
	var timeout time.Duration
	test := &cobra.Command{
		Use:   "test <group|proxy>",
		Short: "Measure latency",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			e, err := loadEnv()
			if err != nil {
				return err
			}
			res, err := e.client.Delay(cmd.Context(), args[0], testURL, timeout)
			if err != nil {
				return err
			}
			names := make([]string, 0, len(res))
			for n := range res {
				names = append(names, n)
			}
			sort.Slice(names, func(i, j int) bool { return less(res[names[i]], res[names[j]], names[i], names[j]) })
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			for _, n := range names {
				d := "timeout"
				if res[n] > 0 {
					d = fmt.Sprintf("%dms", res[n])
				}
				fmt.Fprintf(w, "%s\t%s\n", n, d)
			}
			return w.Flush()
		},
	}
	test.Flags().StringVar(&testURL, "url", "https://www.gstatic.com/generate_204", "test URL")
	test.Flags().DurationVar(&timeout, "timeout", 5*time.Second, "per-proxy timeout")

	c.AddCommand(ls, use, test)
	return c
}

// less sorts reachable proxies by delay first, timeouts last, then by name.
func less(a, b int, an, bn string) bool {
	if (a > 0) != (b > 0) {
		return a > 0
	}
	if a != b {
		return a < b
	}
	return strings.Compare(an, bn) < 0
}

func modeCmd() *cobra.Command {
	return &cobra.Command{
		Use:       "mode [rule|global|direct]",
		Short:     "Show or set the routing mode",
		Args:      cobra.MaximumNArgs(1),
		ValidArgs: []string{"rule", "global", "direct"},
		RunE: func(cmd *cobra.Command, args []string) error {
			e, err := loadEnv()
			if err != nil {
				return err
			}
			if len(args) == 0 {
				cfg, err := e.client.Configs(cmd.Context())
				if err != nil {
					return err
				}
				fmt.Println(cfg.Mode)
				return nil
			}
			if err := cobra.OnlyValidArgs(cmd, args); err != nil {
				return err
			}
			return e.client.PatchConfigs(cmd.Context(), map[string]any{"mode": args[0]})
		},
	}
}

func ruleCmd() *cobra.Command {
	c := &cobra.Command{Use: "rule", Short: "Inspect rules"}
	c.AddCommand(&cobra.Command{
		Use:     "ls [filter]",
		Aliases: []string{"list"},
		Short:   "List rules, optionally filtered by substring",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			e, err := loadEnv()
			if err != nil {
				return err
			}
			rules, err := e.client.Rules(cmd.Context())
			if err != nil {
				return err
			}
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			for _, r := range rules {
				line := fmt.Sprintf("%s\t%s\t%s", r.Type, r.Payload, r.Proxy)
				if len(args) == 1 && !strings.Contains(strings.ToLower(line), strings.ToLower(args[0])) {
					continue
				}
				fmt.Fprintln(w, line)
			}
			return w.Flush()
		},
	})
	return c
}

func logCmd() *cobra.Command {
	var level string
	c := &cobra.Command{
		Use:   "log",
		Short: "Stream core logs",
		RunE: func(cmd *cobra.Command, _ []string) error {
			e, err := loadEnv()
			if err != nil {
				return err
			}
			ctx, stop := signalContext()
			defer stop()
			return e.client.Logs(ctx, level, func(l api.LogEntry) bool {
				fmt.Printf("%s %-7s %s\n", time.Now().Format(time.TimeOnly), l.Type, l.Payload)
				return true
			})
		},
	}
	c.Flags().StringVarP(&level, "level", "l", "info", "debug|info|warning|error")
	return c
}
