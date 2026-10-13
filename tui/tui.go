// Package tui is the Bubble Tea terminal UI for a running core.
package tui

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/ZeroMarker/clashon/sdk/api"
	"github.com/ZeroMarker/clashon/sdk/app"
)

const (
	tabProxies = iota
	tabConns
	tabRules
	tabLogs
	numTabs
)

const (
	maxLogs  = 500
	testURL  = "https://www.gstatic.com/generate_204"
	testWait = 5 * time.Second
)

var modes = []string{"rule", "global", "direct"}

// Model is the root Bubble Tea model.
type Model struct {
	client  *api.Client
	profile string
	ctx     context.Context
	cancel  context.CancelFunc
	events  chan tea.Msg // traffic and log streams

	w, h   int
	tab    int
	status string
	err    error

	mode    string
	traffic api.Traffic

	proxies   map[string]*api.Proxy
	groups    []string
	gi, pi    int
	inMembers bool
	testing   bool

	conns *api.Connections
	ci    int

	rules     []api.Rule
	ri        int
	filter    textinput.Model
	filtering bool

	logs     []api.LogEntry
	logShift int // lines scrolled up from the newest entry
}

// New returns a model bound to a's controller.
func New(a *app.App) *Model {
	profile := ""
	if p, err := a.Store.Current(); err == nil {
		profile = p.Name
	}
	return newModel(a.Client, profile)
}

func newModel(c *api.Client, profile string) *Model {
	ctx, cancel := context.WithCancel(context.Background())
	m := &Model{
		client:  c,
		profile: profile,
		ctx:     ctx,
		cancel:  cancel,
		events:  make(chan tea.Msg, 256),
		filter:  textinput.New(),
	}
	m.filter.Prompt = "/"
	return m
}

// Run starts the TUI and blocks until the user quits.
func Run(a *app.App) error {
	m := New(a)
	defer m.cancel()
	_, err := tea.NewProgram(m, tea.WithAltScreen()).Run()
	return err
}

type (
	proxiesMsg map[string]*api.Proxy
	configsMsg *api.Configs
	rulesMsg   []api.Rule
	connsMsg   *api.Connections
	trafficMsg api.Traffic
	logMsg     api.LogEntry
	statusMsg  string
	errMsg     struct{ error }
	streamErr  struct{ error } // a stream ended; its waiter must be re-armed
	tickMsg    time.Time
	testedMsg  struct {
		group string
		n     int
	}
)

func (m *Model) Init() tea.Cmd {
	m.startStreams()
	return tea.Batch(m.loadProxies(), m.loadConfigs(), m.loadRules(), m.waitEvent(), tick())
}

func (m *Model) startStreams() {
	push := func(msg tea.Msg) bool {
		select {
		case m.events <- msg:
		case <-m.ctx.Done():
			return false
		default: // drop when the UI falls behind
		}
		return true
	}
	go func() {
		err := m.client.Traffic(m.ctx, func(t api.Traffic) bool { return push(trafficMsg(t)) })
		if err != nil {
			push(streamErr{err})
		}
	}()
	go func() {
		err := m.client.Logs(m.ctx, "info", func(l api.LogEntry) bool { return push(logMsg(l)) })
		if err != nil {
			push(streamErr{err})
		}
	}()
}

func (m *Model) waitEvent() tea.Cmd {
	return func() tea.Msg {
		select {
		case msg := <-m.events:
			return msg
		case <-m.ctx.Done():
			return nil
		}
	}
}

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg { return tickMsg(t) })
}

func (m *Model) call(fn func(ctx context.Context) (tea.Msg, error)) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 15*time.Second)
		defer cancel()
		msg, err := fn(ctx)
		if err != nil {
			return errMsg{err}
		}
		return msg
	}
}

func (m *Model) loadProxies() tea.Cmd {
	return m.call(func(ctx context.Context) (tea.Msg, error) {
		p, err := m.client.Proxies(ctx)
		return proxiesMsg(p), err
	})
}

func (m *Model) loadConfigs() tea.Cmd {
	return m.call(func(ctx context.Context) (tea.Msg, error) {
		c, err := m.client.Configs(ctx)
		return configsMsg(c), err
	})
}

func (m *Model) loadRules() tea.Cmd {
	return m.call(func(ctx context.Context) (tea.Msg, error) {
		r, err := m.client.Rules(ctx)
		return rulesMsg(r), err
	})
}

func (m *Model) loadConns() tea.Cmd {
	return m.call(func(ctx context.Context) (tea.Msg, error) {
		c, err := m.client.Connections(ctx)
		return connsMsg(c), err
	})
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.w, m.h = msg.Width, msg.Height
	case tea.KeyMsg:
		return m, m.key(msg)
	case tickMsg:
		cmds := []tea.Cmd{tick()}
		if m.tab == tabConns {
			cmds = append(cmds, m.loadConns())
		}
		return m, tea.Batch(cmds...)
	case proxiesMsg:
		m.setProxies(msg)
	case configsMsg:
		m.mode = msg.Mode
	case rulesMsg:
		m.rules = msg
	case connsMsg:
		sort.SliceStable(msg.Connections, func(i, j int) bool {
			return msg.Connections[i].Start.After(msg.Connections[j].Start)
		})
		m.conns = msg
		m.ci = clamp(m.ci, len(msg.Connections))
	case trafficMsg:
		m.traffic = api.Traffic(msg)
		return m, m.waitEvent()
	case logMsg:
		m.logs = append(m.logs, api.LogEntry(msg))
		if len(m.logs) > maxLogs {
			m.logs = m.logs[len(m.logs)-maxLogs:]
		}
		if m.logShift > 0 {
			m.logShift++ // keep the viewed lines in place
		}
		return m, m.waitEvent()
	case testedMsg:
		m.testing = false
		m.status = fmt.Sprintf("tested %s: %d reachable", msg.group, msg.n)
		return m, m.loadProxies()
	case statusMsg:
		m.status, m.err = string(msg), nil
	case errMsg:
		m.testing = false
		m.err = msg.error
	case streamErr:
		m.err = msg.error
		return m, m.waitEvent()
	}
	return m, nil
}

// setProxies orders groups as the config does (GLOBAL lists them in order).
func (m *Model) setProxies(p map[string]*api.Proxy) {
	m.proxies = p
	m.groups = m.groups[:0]
	if g, ok := p["GLOBAL"]; ok {
		for _, n := range g.All {
			if q := p[n]; q != nil && q.IsGroup() {
				m.groups = append(m.groups, n)
			}
		}
		m.groups = append(m.groups, "GLOBAL")
	}
	m.gi = clamp(m.gi, len(m.groups))
	m.pi = clamp(m.pi, len(m.members()))
}

func (m *Model) group() *api.Proxy {
	if m.gi < len(m.groups) {
		return m.proxies[m.groups[m.gi]]
	}
	return nil
}

func (m *Model) members() []string {
	if g := m.group(); g != nil {
		return g.All
	}
	return nil
}

func (m *Model) visibleRules() []api.Rule {
	f := strings.ToLower(m.filter.Value())
	if f == "" {
		return m.rules
	}
	var out []api.Rule
	for _, r := range m.rules {
		if strings.Contains(strings.ToLower(r.Type+" "+r.Payload+" "+r.Proxy), f) {
			out = append(out, r)
		}
	}
	return out
}

func (m *Model) key(k tea.KeyMsg) tea.Cmd {
	if m.filtering {
		switch k.String() {
		case "enter":
			m.filtering = false
			m.filter.Blur()
		case "esc":
			m.filtering = false
			m.filter.Blur()
			m.filter.SetValue("")
		default:
			var cmd tea.Cmd
			m.filter, cmd = m.filter.Update(k)
			m.ri = 0
			return cmd
		}
		return nil
	}

	m.status, m.err = "", nil
	switch k.String() {
	case "q", "ctrl+c":
		m.cancel()
		return tea.Quit
	case "tab":
		return m.setTab((m.tab + 1) % numTabs)
	case "shift+tab":
		return m.setTab((m.tab + numTabs - 1) % numTabs)
	case "1", "2", "3", "4":
		return m.setTab(int(k.String()[0] - '1'))
	case "m":
		return m.cycleMode()
	case "r":
		return tea.Batch(m.loadProxies(), m.loadConfigs(), m.loadRules(), m.loadConns())
	}

	switch m.tab {
	case tabProxies:
		return m.keyProxies(k)
	case tabConns:
		return m.keyConns(k)
	case tabRules:
		switch k.String() {
		case "/":
			m.filtering = true
			return m.filter.Focus()
		case "esc":
			m.filter.SetValue("")
		default:
			m.ri = move(k, m.ri, len(m.visibleRules()), m.bodyHeight())
		}
	case tabLogs:
		switch k.String() {
		case "up", "k":
			m.logShift = min(m.logShift+1, max(len(m.logs)-1, 0))
		case "down", "j":
			m.logShift = max(m.logShift-1, 0)
		case "G", "end":
			m.logShift = 0
		case "c":
			m.logs, m.logShift = nil, 0
		}
	}
	return nil
}

func (m *Model) setTab(t int) tea.Cmd {
	m.tab = t
	if t == tabConns {
		return m.loadConns()
	}
	return nil
}

func (m *Model) cycleMode() tea.Cmd {
	next := modes[0]
	for i, md := range modes {
		if md == m.mode {
			next = modes[(i+1)%len(modes)]
		}
	}
	return m.call(func(ctx context.Context) (tea.Msg, error) {
		if err := m.client.PatchConfigs(ctx, map[string]any{"mode": next}); err != nil {
			return nil, err
		}
		return configsMsg(&api.Configs{Mode: next}), nil
	})
}

func (m *Model) keyProxies(k tea.KeyMsg) tea.Cmd {
	switch k.String() {
	case "t":
		if m.testing || m.group() == nil {
			return nil
		}
		m.testing = true
		name := m.groups[m.gi]
		m.status = "testing " + name + "…"
		return m.call(func(ctx context.Context) (tea.Msg, error) {
			res, err := m.client.Delay(ctx, name, testURL, testWait)
			n := 0
			for _, d := range res {
				if d > 0 {
					n++
				}
			}
			return testedMsg{name, n}, err
		})
	case "right", "l":
		m.inMembers = true
		return nil
	case "left", "h", "esc":
		m.inMembers = false
		return nil
	case "enter":
		if !m.inMembers {
			m.inMembers = true
			if g := m.group(); g != nil {
				m.pi = max(indexOf(g.All, g.Now), 0)
			}
			return nil
		}
		g, mem := m.group(), m.members()
		if g == nil || m.pi >= len(mem) {
			return nil
		}
		if g.Type != "Selector" {
			return func() tea.Msg { return statusMsg(g.Type + " groups choose automatically") }
		}
		group, proxy := m.groups[m.gi], mem[m.pi]
		return tea.Sequence(m.call(func(ctx context.Context) (tea.Msg, error) {
			return statusMsg(group + " → " + proxy), m.client.Select(ctx, group, proxy)
		}), m.loadProxies())
	}
	if m.inMembers {
		m.pi = move(k, m.pi, len(m.members()), m.bodyHeight())
	} else {
		old := m.gi
		m.gi = move(k, m.gi, len(m.groups), m.bodyHeight())
		if m.gi != old {
			m.pi = 0
		}
	}
	return nil
}

func (m *Model) keyConns(k tea.KeyMsg) tea.Cmd {
	var list []api.Connection
	if m.conns != nil {
		list = m.conns.Connections
	}
	switch k.String() {
	case "x":
		if m.ci < len(list) {
			id := list[m.ci].ID
			return tea.Sequence(m.call(func(ctx context.Context) (tea.Msg, error) {
				return statusMsg("connection closed"), m.client.CloseConnection(ctx, id)
			}), m.loadConns())
		}
	case "X":
		return tea.Sequence(m.call(func(ctx context.Context) (tea.Msg, error) {
			return statusMsg("all connections closed"), m.client.CloseConnection(ctx, "")
		}), m.loadConns())
	default:
		m.ci = move(k, m.ci, len(list), m.bodyHeight()-1)
	}
	return nil
}

// move handles list navigation keys and returns the new index.
func move(k tea.KeyMsg, i, n, page int) int {
	switch k.String() {
	case "up", "k":
		i--
	case "down", "j":
		i++
	case "pgup", "ctrl+u":
		i -= max(page, 1)
	case "pgdown", "ctrl+d":
		i += max(page, 1)
	case "home", "g":
		i = 0
	case "end", "G":
		i = n - 1
	}
	return clamp(i, n)
}

func clamp(i, n int) int {
	if i >= n {
		i = n - 1
	}
	return max(i, 0)
}

func indexOf(s []string, v string) int {
	for i, x := range s {
		if x == v {
			return i
		}
	}
	return -1
}
