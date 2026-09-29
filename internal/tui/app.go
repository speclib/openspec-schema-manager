package tui

import (
	"context"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/speclib/openspec-schema-manager/internal/config"
)

type Options struct {
	Version     string
	WorkDir     string
	InProject   bool
	ProjectRoot string
	Config      config.Config
	Paths       config.Paths
	Registry    *RegistryLoader
}

type Model struct {
	opts     Options
	screens  []Screen
	current  int
	helpOpen bool
	status   string
	width    int
	height   int
	quit     bool
}

func New(opts Options) Model {
	m := Model{
		opts: opts,
		screens: []Screen{
			projectScreen(opts.InProject),
			newRegistryScreen(opts.Registry),
			localScreen(),
			composerScreen(),
		},
		width:  80,
		height: 24,
	}

	if !opts.InProject {
		m.current = m.indexOf("Registry")
		m.status = "no OpenSpec project here"
	} else {
		m.current = m.indexOf("Project")
		m.status = opts.ProjectRoot
	}

	if opts.Registry != nil {
		at := m.indexOf("Registry")
		m.screens[at], _ = m.screens[at].Update(opts.Registry.Load(context.Background()))
	}

	return m
}

func (m Model) indexOf(title string) int {
	for i, s := range m.screens {
		if s.Title() == title {
			return i
		}
	}
	return 0
}

// Init asks for a refresh and nothing else. The cached registry is already in
// the model, applied by New: it is a local file read, shorter than a frame, and
// batching it alongside the fetch let a fast fetch be overwritten by the stale
// read that started before it.
func (m Model) Init() tea.Cmd {
	if m.opts.Registry == nil || !m.opts.Registry.Stale() {
		return nil
	}

	return m.opts.Registry.RefreshCmd()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyPressMsg:
		return m.handleKey(msg)
	}

	updated, cmd := m.screens[m.current].Update(msg)
	m.screens[m.current] = updated
	return m, cmd
}

func (m Model) handleKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	if key == "ctrl+c" {
		m.quit = true
		return m, tea.Quit
	}

	if m.helpOpen {
		switch key {
		case "?", "esc", "q":
			m.helpOpen = false
			return m, nil
		}

		moved, _ := m.selectTab(key)
		return moved, nil
	}

	if !m.screens[m.current].Capturing() {
		switch key {
		case "q":
			m.quit = true
			return m, tea.Quit
		case "?":
			m.helpOpen = true
			return m, nil
		}

		if moved, ok := m.selectTab(key); ok {
			return moved, nil
		}
	}

	updated, cmd := m.screens[m.current].Update(msg)
	m.screens[m.current] = updated
	return m, cmd
}

func (m Model) selectTab(key string) (Model, bool) {
	switch key {
	case "tab":
		m.current = (m.current + 1) % len(m.screens)
		return m, true
	case "shift+tab":
		m.current = (m.current - 1 + len(m.screens)) % len(m.screens)
		return m, true
	}

	if len(key) == 1 && key[0] >= '1' && key[0] <= '9' {
		if index := int(key[0] - '1'); index < len(m.screens) {
			m.current = index
		}
		return m, true
	}

	return m, false
}

func (m Model) View() tea.View {
	v := tea.NewView(m.Render())
	v.AltScreen = true
	return v
}

func (m Model) Render() string {
	var b strings.Builder

	b.WriteString(m.renderTabs())
	b.WriteString("\n")
	b.WriteString(rule(m.width))
	b.WriteString("\n")

	paneHeight := m.height - 5
	if paneHeight < 1 {
		paneHeight = 1
	}

	if m.helpOpen {
		b.WriteString(renderHelp(m.screens[m.current], m.width))
	} else {
		b.WriteString(m.screens[m.current].View(m.width, paneHeight))
	}

	b.WriteString("\n")
	b.WriteString(rule(m.width))
	b.WriteString("\n")
	b.WriteString(m.renderStatus())

	return b.String()
}

func (m Model) renderTabs() string {
	var b strings.Builder

	b.WriteString(" ")
	for i, s := range m.screens {
		if i > 0 {
			b.WriteString(dimStyle.Render(" │ "))
		}
		if i == m.current {
			b.WriteString(activeTabStyle.Render(s.Title()))
		} else {
			b.WriteString(inactiveTabStyle.Render(s.Title()))
		}
	}

	return b.String()
}

func (m Model) renderStatus() string {
	const help = "  ·  ? help"

	left := " ossm " + m.opts.Version
	line := left + help

	if m.status != "" {
		line = left + "  ·  " + m.status + help
	}

	return dimStyle.Render(truncate(line, m.width))
}

func (m Model) SetStatus(s string) Model {
	m.status = s
	return m
}

func (m Model) Status() string { return m.status }

func (m Model) CurrentTitle() string { return m.screens[m.current].Title() }

func (m Model) HelpOpen() bool { return m.helpOpen }

func (m Model) Quitting() bool { return m.quit }
