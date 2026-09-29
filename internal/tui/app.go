package tui

import (
	"context"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/speclib/openspec-schema-manager/internal/compose"
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
	Resolver    *Resolver
	Installer   *Installer
	ProjectRead *ProjectReader
	Recents     config.Recents
	Drafts      compose.Drafts
	Comparer    *Comparer
	OpenPath    string
}

type Model struct {
	opts     Options
	screens  []Screen
	current  int
	helpOpen bool
	prompt   *pathPrompt
	status   string
	width    int
	height   int
	quit     bool
}

func New(opts Options) Model {
	m := Model{
		opts: opts,
		screens: []Screen{
			newProjectScreen(opts.InProject, opts.ProjectRoot, opts.ProjectRead, opts.Resolver, opts.Comparer),
			newRegistryScreen(opts.Registry, opts.Resolver, opts.Installer),
			newLocalScreen(opts.Config.SchemasDirs, opts.Recents, opts.Resolver),
			newComposerScreen(opts.Config.SchemasDirs, opts.Drafts, opts.Resolver == nil || opts.Resolver.ASCII),
		},
		prompt: newPathPrompt(),
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

	// The comparison in the Project view needs the registry's entries, and the
	// Registry screen is the one that has them. Wiring it here keeps the
	// screens from knowing about each other.
	if opts.Comparer != nil {
		if registryScreen, ok := m.screens[m.indexOf("Registry")].(*registryScreenModel); ok {
			opts.Comparer.Entries = registryScreen.Entries
		}
	}

	if opts.OpenPath != "" {
		m.openPath(opts.OpenPath)
	}

	return m
}

func (m *Model) local() *localModel {
	screen, _ := m.screens[m.indexOf("Local")].(*localModel)
	return screen
}

func (m *Model) composer() *composerModel {
	screen, _ := m.screens[m.indexOf("Composer")].(*composerModel)
	return screen
}

// SendToComposer adds a schema to the composer's palette without leaving the
// tab it was sent from.
func (m Model) SendToComposer(msg SourceAdded) (Model, tea.Cmd) {
	composer := m.composer()
	if composer == nil {
		return m, nil
	}

	_, cmd := composer.Update(msg)

	return m, cmd
}

func (m *Model) openPath(dir string) {
	local := m.local()
	if local == nil {
		return
	}

	local.Remember(dir)
	m.current = m.indexOf("Local")
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
	var cmds []tea.Cmd

	if m.opts.Registry != nil && m.opts.Registry.Stale() {
		cmds = append(cmds, m.opts.Registry.RefreshCmd())
	}
	if m.opts.ProjectRead != nil && m.opts.InProject {
		cmds = append(cmds, m.opts.ProjectRead.ReadCmd())
	}

	if len(cmds) == 0 {
		return nil
	}

	return tea.Batch(cmds...)
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

	if added, ok := msg.(SourceAdded); ok {
		return m.SendToComposer(added)
	}

	if _, installed := msg.(InstallFinished); installed {
		var cmds []tea.Cmd
		for i, screen := range m.screens {
			updated, cmd := screen.Update(msg)
			m.screens[i] = updated
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
		return m, tea.Batch(cmds...)
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

	if m.prompt.open {
		return m.handlePromptKey(msg, key)
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
		case ":", "ctrl+o":
			m.prompt.Open()
			return m, nil
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

func (m Model) handlePromptKey(msg tea.KeyPressMsg, key string) (tea.Model, tea.Cmd) {
	switch key {
	case "esc":
		m.prompt.Close()
	case "enter":
		if dir, ok := m.prompt.Submit(); ok {
			m.prompt.Close()
			m.openPath(dir)
		}
	case "tab":
		m.prompt.Complete()
	case "backspace":
		m.prompt.Backspace()
	default:
		if msg.Text != "" {
			m.prompt.Type(msg.Text)
		}
	}

	return m, nil
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

	if m.prompt.open {
		b.WriteString(m.prompt.View(m.width))
	} else if m.helpOpen {
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
