package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/ZeroMarker/clashon/sdk/units"
)

var (
	accent   = lipgloss.Color("12")
	muted    = lipgloss.Color("8")
	sHeader  = lipgloss.NewStyle().Bold(true).Foreground(accent)
	sMuted   = lipgloss.NewStyle().Foreground(muted)
	sTab     = lipgloss.NewStyle().Padding(0, 1).Foreground(muted)
	sTabOn   = lipgloss.NewStyle().Padding(0, 1).Bold(true).Foreground(lipgloss.Color("0")).Background(accent)
	sCursor  = lipgloss.NewStyle().Reverse(true)
	sCurrent = lipgloss.NewStyle().Bold(true).Foreground(accent)
	sErr     = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	sGood    = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	sOK      = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	sSlow    = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
)

var tabNames = [numTabs]string{"Proxies", "Connections", "Rules", "Logs"}

var help = [numTabs]string{
	"↑↓ move  ←→/enter pane  enter select  t test  m mode  1-4 tab  q quit",
	"↑↓ move  x close  X close all  m mode  1-4 tab  q quit",
	"↑↓ move  / filter  esc clear  m mode  1-4 tab  q quit",
	"↑↓ scroll  G follow  c clear  m mode  1-4 tab  q quit",
}

// bodyHeight is the space between the header/tab bar and the footer.
func (m *Model) bodyHeight() int { return max(m.h-3, 1) }

func (m *Model) View() string {
	if m.w == 0 {
		return ""
	}
	var body string
	switch m.tab {
	case tabProxies:
		body = m.viewProxies()
	case tabConns:
		body = m.viewConns()
	case tabRules:
		body = m.viewRules()
	case tabLogs:
		body = m.viewLogs()
	}
	// Pad/clip the body to an exact height so the footer stays at the bottom.
	lines := strings.Split(body, "\n")
	h := m.bodyHeight()
	if len(lines) > h {
		lines = lines[:h]
	}
	for len(lines) < h {
		lines = append(lines, "")
	}
	return strings.Join([]string{m.viewHeader(), m.viewTabs(), strings.Join(lines, "\n"), m.viewFooter()}, "\n")
}

func (m *Model) viewHeader() string {
	left := sHeader.Render("ClashOn") + sMuted.Render(fmt.Sprintf("  profile %s  mode %s", orDash(m.profile), orDash(m.mode)))
	right := fmt.Sprintf("↑ %s/s  ↓ %s/s", units.Bytes(m.traffic.Up), units.Bytes(m.traffic.Down))
	gap := m.w - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		return fit(left, m.w)
	}
	return left + strings.Repeat(" ", gap) + right
}

func (m *Model) viewTabs() string {
	parts := make([]string, numTabs)
	for i, n := range tabNames {
		label := fmt.Sprintf("%d %s", i+1, n)
		if i == m.tab {
			parts[i] = sTabOn.Render(label)
		} else {
			parts[i] = sTab.Render(label)
		}
	}
	return fit(strings.Join(parts, ""), m.w)
}

func (m *Model) viewFooter() string {
	switch {
	case m.err != nil:
		return fit(sErr.Render("error: "+m.err.Error()), m.w)
	case m.filtering:
		return fit(m.filter.View(), m.w)
	case m.status != "":
		return fit(m.status, m.w)
	}
	return fit(sMuted.Render(help[m.tab]), m.w)
}

func (m *Model) viewProxies() string {
	h := m.bodyHeight()
	lw := min(max(m.w/3, 16), 32)
	rw := m.w - lw - 1

	left := make([]string, 0, h)
	start, end := window(len(m.groups), m.gi, h)
	for i := start; i < end; i++ {
		row := fit(" "+m.groups[i], lw)
		if i == m.gi {
			if m.inMembers {
				row = sCurrent.Render(row)
			} else {
				row = sCursor.Render(row)
			}
		}
		left = append(left, row)
	}

	right := make([]string, 0, h)
	if g := m.group(); g != nil {
		right = append(right, sMuted.Render(fit(fmt.Sprintf(" %s · %s · %d", g.Type, orDash(g.Now), len(g.All)), rw)))
		mem := g.All
		start, end := window(len(mem), m.pi, h-1)
		for i := start; i < end; i++ {
			name := mem[i]
			mark, typ, delay := "  ", "", sMuted.Render("    -")
			if name == g.Now {
				mark = "● "
			}
			if p := m.proxies[name]; p != nil {
				typ = p.Type
				delay = fmtDelay(p.LastDelay())
			}
			dw := lipgloss.Width(delay)
			row := col(" "+mark+name, rw-dw-12) + col(sMuted.Render(typ), 12) + delay
			if m.inMembers && i == m.pi {
				row = sCursor.Render(ansi.Strip(row))
			} else if name == g.Now {
				row = sCurrent.Render(ansi.Strip(row))
			}
			right = append(right, row)
		}
	}

	rows := make([]string, h)
	for i := range rows {
		l, r := strings.Repeat(" ", lw), ""
		if i < len(left) {
			l = left[i]
		}
		if i < len(right) {
			r = right[i]
		}
		rows[i] = l + sMuted.Render("│") + r
	}
	return strings.Join(rows, "\n")
}

func fmtDelay(d int) string {
	switch {
	case d <= 0:
		return sMuted.Render("    -")
	case d < 300:
		return sGood.Render(fmt.Sprintf("%4dms", d))
	case d < 800:
		return sOK.Render(fmt.Sprintf("%4dms", d))
	}
	return sSlow.Render(fmt.Sprintf("%4dms", d))
}

func (m *Model) viewConns() string {
	if m.conns == nil {
		return sMuted.Render(" loading…")
	}
	list := m.conns.Connections
	h := m.bodyHeight() - 1
	// target | rule | chain | ↓ | ↑ | age
	fixed := 9 + 9 + 6 + 4
	flex := max(m.w-fixed, 30)
	tw, rw := flex*45/100, flex*20/100
	cw := flex - tw - rw
	head := col(" TARGET", tw) + col("RULE", rw) + col("CHAIN", cw) + rpad("DOWN", 9) + rpad("UP", 9) + rpad("AGE", 6)
	out := []string{sMuted.Render(fmt.Sprintf("%s  (%d conns, ↓ %s ↑ %s total)", head, len(list), units.Bytes(m.conns.DownloadTotal), units.Bytes(m.conns.UploadTotal)))}
	out[0] = fit(out[0], m.w)
	start, end := window(len(list), m.ci, h)
	for i := start; i < end; i++ {
		c := list[i]
		rule := c.Rule
		if c.RulePayload != "" {
			rule += "," + c.RulePayload
		}
		row := col(" "+c.Target(), tw) + col(rule, rw) + col(chain(c.Chains), cw) +
			rpad(units.Bytes(c.Download), 9) + rpad(units.Bytes(c.Upload), 9) + rpad(age(c.Start), 6)
		if i == m.ci {
			row = sCursor.Render(row)
		}
		out = append(out, row)
	}
	return strings.Join(out, "\n")
}

// chain renders mihomo's chains (innermost first) as "Group → Node".
func chain(c []string) string {
	r := make([]string, len(c))
	for i, s := range c {
		r[len(c)-1-i] = s
	}
	return strings.Join(r, " → ")
}

func age(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%ds", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	return fmt.Sprintf("%dh", int(d.Hours()))
}

func (m *Model) viewRules() string {
	rules := m.visibleRules()
	h := m.bodyHeight()
	tw := 16
	pw := max((m.w-tw)*2/3, 10)
	out := make([]string, 0, h)
	start, end := window(len(rules), m.ri, h)
	for i := start; i < end; i++ {
		r := rules[i]
		row := col(" "+r.Type, tw) + col(r.Payload, pw) + fit(r.Proxy, m.w-tw-pw)
		if i == m.ri {
			row = sCursor.Render(row)
		}
		out = append(out, row)
	}
	if len(rules) == 0 {
		out = append(out, sMuted.Render(" no rules"))
	}
	return strings.Join(out, "\n")
}

func (m *Model) viewLogs() string {
	h := m.bodyHeight()
	end := len(m.logs) - m.logShift
	start := max(end-h, 0)
	out := make([]string, 0, h)
	for _, l := range m.logs[start:end] {
		lvl := l.Type
		switch lvl {
		case "error":
			lvl = sErr.Render(fit(lvl, 8))
		case "warning":
			lvl = sOK.Render(fit(lvl, 8))
		default:
			lvl = sMuted.Render(fit(lvl, 8))
		}
		out = append(out, lvl+fit(l.Payload, m.w-8))
	}
	if len(m.logs) == 0 {
		out = append(out, sMuted.Render(" waiting for logs…"))
	}
	return strings.Join(out, "\n")
}

// window returns the [start, end) slice of n rows of height h keeping sel visible.
func window(n, sel, h int) (int, int) {
	if h <= 0 || n == 0 {
		return 0, 0
	}
	start := 0
	if sel >= h {
		start = sel - h + 1
	}
	return start, min(start+h, n)
}

// col is fit with a one-cell gutter on the right.
func col(s string, w int) string { return fit(s, w-1) + " " }

// fit truncates or pads s to exactly w terminal cells (CJK-aware).
func fit(s string, w int) string {
	if w <= 0 {
		return ""
	}
	s = ansi.Truncate(s, w, "…")
	if pad := w - ansi.StringWidth(s); pad > 0 {
		s += strings.Repeat(" ", pad)
	}
	return s
}

func rpad(s string, w int) string {
	if pad := w - ansi.StringWidth(s); pad > 0 {
		return strings.Repeat(" ", pad) + s
	}
	return s
}

func orDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}
