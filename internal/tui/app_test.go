package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/speclib/openspec-schema-manager/internal/ansi"
)

func render(m Model) string { return ansi.Strip(m.Render()) }

// flat collapses every run of whitespace to one space, so an assertion on a
// phrase does not fail because the renderer wrapped it.
func flat(s string) string { return strings.Join(strings.Fields(s), " ") }

func press(t *testing.T, m Model, keys ...string) Model {
	t.Helper()

	for _, key := range keys {
		updated, _ := m.Update(keyPress(key))
		next, ok := updated.(Model)
		if !ok {
			t.Fatalf("Update returned %T, want Model", updated)
		}
		m = next
	}

	return m
}

func keyPress(key string) tea.KeyPressMsg {
	switch key {
	case "tab":
		return tea.KeyPressMsg{Code: tea.KeyTab}
	case "shift+tab":
		return tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift}
	case "esc":
		return tea.KeyPressMsg{Code: tea.KeyEscape}
	case "ctrl+c":
		return tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl}
	default:
		return tea.KeyPressMsg{Code: rune(key[0]), Text: key}
	}
}

func outsideProject() Model {
	return New(Options{Version: "0.1.0"})
}

func insideProject() Model {
	return New(Options{Version: "0.1.0", InProject: true, ProjectRoot: "/home/tester/demo-app"})
}

func TestFourTabsInAFixedOrder(t *testing.T) {
	t.Parallel()

	want := []string{"Project", "Registry", "Local", "Composer"}

	for _, m := range []Model{outsideProject(), insideProject()} {
		if len(m.screens) != len(want) {
			t.Fatalf("got %d tabs, want %d", len(m.screens), len(want))
		}
		for i, title := range want {
			if got := m.screens[i].Title(); got != title {
				t.Errorf("tab %d is %q, want %q", i, got, title)
			}
		}

		view := render(m)
		last := -1
		for _, title := range want {
			at := strings.Index(view, title)
			if at < 0 {
				t.Fatalf("the view does not draw the %s tab:\n%s", title, view)
			}
			if at < last {
				t.Errorf("the %s tab is drawn out of order", title)
			}
			last = at
		}
	}
}

func TestTheOpeningTabReflectsWhereTheUserIsStanding(t *testing.T) {
	t.Parallel()

	if got := outsideProject().CurrentTitle(); got != "Registry" {
		t.Errorf("outside a project the opening tab is %q, want Registry", got)
	}
	if got := insideProject().CurrentTitle(); got != "Project" {
		t.Errorf("inside a project the opening tab is %q, want Project", got)
	}
}

func TestOutsideAProjectTheStatusLineSaysSo(t *testing.T) {
	t.Parallel()

	if view := render(outsideProject()); !strings.Contains(view, "no OpenSpec project here") {
		t.Errorf("the status line does not say there is no project:\n%s", view)
	}
}

func TestOutsideAProjectTheProjectTabExplainsItself(t *testing.T) {
	t.Parallel()

	view := flat(render(press(t, outsideProject(), "1")))

	if !strings.Contains(view, "no OpenSpec project here") {
		t.Errorf("the Project tab does not say there is no project:\n%s", view)
	}
	if !strings.Contains(view, "installing one needs a project") {
		t.Errorf("the Project tab does not say what that costs:\n%s", view)
	}
	if !strings.Contains(view, "--path") {
		t.Errorf("the Project tab does not say how to point at one:\n%s", view)
	}
}

func TestInsideAProjectTheStatusLineNamesTheRoot(t *testing.T) {
	t.Parallel()

	if view := render(insideProject()); !strings.Contains(view, "/home/tester/demo-app") {
		t.Errorf("the status line does not name the project root:\n%s", view)
	}
}

func TestCyclingWrapsAtBothEnds(t *testing.T) {
	t.Parallel()

	m := insideProject()

	for _, want := range []string{"Registry", "Local", "Composer", "Project"} {
		m = press(t, m, "tab")
		if got := m.CurrentTitle(); got != want {
			t.Fatalf("after tab the selection is %q, want %q", got, want)
		}
	}

	for _, want := range []string{"Composer", "Local", "Registry", "Project"} {
		m = press(t, m, "shift+tab")
		if got := m.CurrentTitle(); got != want {
			t.Fatalf("after shift+tab the selection is %q, want %q", got, want)
		}
	}
}

func TestSelectingATabByNumber(t *testing.T) {
	t.Parallel()

	cases := map[string]string{"1": "Project", "2": "Registry", "3": "Local", "4": "Composer"}

	for key, want := range cases {
		m := press(t, outsideProject(), key)
		if got := m.CurrentTitle(); got != want {
			t.Errorf("%s selected %q, want %q", key, got, want)
		}
	}
}

func TestADigitWithNoTabChangesNothing(t *testing.T) {
	t.Parallel()

	m := press(t, outsideProject(), "3")
	before := m.CurrentTitle()

	m = press(t, m, "5", "9")

	if got := m.CurrentTitle(); got != before {
		t.Errorf("a digit with no tab moved the selection from %q to %q", before, got)
	}
	if m.Quitting() {
		t.Error("a digit with no tab quit the application")
	}
}

func TestHelpOpensAndLists(t *testing.T) {
	t.Parallel()

	m := press(t, outsideProject(), "?")

	if !m.HelpOpen() {
		t.Fatal("? did not open the help overlay")
	}

	view := flat(render(m))
	for _, want := range []string{"Everywhere", "tab / shift+tab", "ctrl+c", "Registry", "filter over id, name and description"} {
		if !strings.Contains(view, want) {
			t.Errorf("the help overlay does not mention %q:\n%s", want, view)
		}
	}
}

func TestHelpFollowsTheSelectedTab(t *testing.T) {
	t.Parallel()

	m := press(t, outsideProject(), "?", "4")

	view := render(m)
	if !strings.Contains(view, "Composer") {
		t.Errorf("the help overlay did not follow the tab change:\n%s", view)
	}
	if !strings.Contains(view, "write the schema") {
		t.Errorf("the help overlay does not list the Composer keys:\n%s", view)
	}
	if !m.HelpOpen() {
		t.Error("changing tabs closed the help overlay")
	}
}

func TestHelpClosesOnEveryDocumentedKey(t *testing.T) {
	t.Parallel()

	for _, key := range []string{"?", "esc", "q"} {
		m := press(t, outsideProject(), "?", key)

		if m.HelpOpen() {
			t.Errorf("%s did not close the help overlay", key)
		}
		if m.Quitting() {
			t.Errorf("%s quit the application instead of closing the overlay", key)
		}
	}
}

func TestQuitting(t *testing.T) {
	t.Parallel()

	if m := press(t, outsideProject(), "q"); !m.Quitting() {
		t.Error("q with nothing open did not quit")
	}
	if m := press(t, outsideProject(), "ctrl+c"); !m.Quitting() {
		t.Error("ctrl+c did not quit")
	}
	if m := press(t, outsideProject(), "?", "ctrl+c"); !m.Quitting() {
		t.Error("ctrl+c with help open did not quit")
	}
}

func TestTheStatusLineCarriesTheNameAndVersion(t *testing.T) {
	t.Parallel()

	if view := render(outsideProject()); !strings.Contains(view, "ossm 0.1.0") {
		t.Errorf("the status line does not carry the name and version:\n%s", view)
	}
}

func TestAScreenCanSetAndReplaceTheStatus(t *testing.T) {
	t.Parallel()

	m := outsideProject().SetStatus("fetching the registry")
	if !strings.Contains(render(m), "fetching the registry") {
		t.Error("a status message was not drawn")
	}

	m = m.SetStatus("cached 3h ago")
	view := render(m)
	if !strings.Contains(view, "cached 3h ago") {
		t.Error("a replacement status message was not drawn")
	}
	if strings.Contains(view, "fetching the registry") {
		t.Error("the replaced status message is still drawn")
	}

	if m = m.SetStatus(""); strings.Contains(render(m), "·  ·") {
		t.Error("clearing the status left an empty separator")
	}
	if m.Status() != "" {
		t.Errorf("Status() = %q, want empty", m.Status())
	}
}

func TestWindowSizeIsTaken(t *testing.T) {
	t.Parallel()

	updated, _ := outsideProject().Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m, ok := updated.(Model)
	if !ok {
		t.Fatalf("Update returned %T, want Model", updated)
	}

	if m.width != 120 || m.height != 40 {
		t.Errorf("size = %dx%d, want 120x40", m.width, m.height)
	}
}

func TestATinyWindowStillDraws(t *testing.T) {
	t.Parallel()

	updated, _ := outsideProject().Update(tea.WindowSizeMsg{Width: 0, Height: 0})
	m := updated.(Model)

	if render(m) == "" {
		t.Error("a zero-sized window drew nothing")
	}
}

func TestEveryUnbuiltTabSaysWhatItWillHold(t *testing.T) {
	t.Parallel()

	m := outsideProject()

	for i, screen := range m.screens {
		m.current = i
		view := render(m)

		if _, unbuilt := screen.(placeholder); unbuilt && !strings.Contains(view, "Built in milestone") {
			t.Errorf("the %s tab does not name the milestone that builds it:\n%s", screen.Title(), view)
		}
		if len(screen.Keys()) == 0 {
			t.Errorf("the %s tab lists no keys, so the help overlay cannot be honest about it", screen.Title())
		}
	}
}

func TestTheRegistryTabIsNoLongerAPlaceholder(t *testing.T) {
	t.Parallel()

	m := outsideProject()

	if _, isPlaceholder := m.screens[m.indexOf("Registry")].(placeholder); isPlaceholder {
		t.Error("the Registry tab is still a placeholder")
	}
	if view := render(press(t, m, "2")); strings.Contains(view, "Built in milestone") {
		t.Errorf("the Registry tab draws a placeholder:\n%s", view)
	}
}

func TestAnUnhandledKeyReachesTheScreen(t *testing.T) {
	t.Parallel()

	m := press(t, outsideProject(), "z")

	if m.Quitting() {
		t.Error("an unhandled key quit the application")
	}
	if m.CurrentTitle() != "Registry" {
		t.Errorf("an unhandled key moved the selection to %q", m.CurrentTitle())
	}
}

func TestAnUnknownMessageReachesTheScreen(t *testing.T) {
	t.Parallel()

	type odd struct{}

	updated, cmd := outsideProject().Update(odd{})
	if _, ok := updated.(Model); !ok {
		t.Fatalf("Update returned %T, want Model", updated)
	}
	if cmd != nil {
		t.Error("a placeholder screen returned a command")
	}
}

func TestViewIsAnAltScreenView(t *testing.T) {
	t.Parallel()

	v := outsideProject().View()

	if !v.AltScreen {
		t.Error("the view does not ask for the alternate screen")
	}
	if v.Content == "" {
		t.Error("the view has no content")
	}
}

func TestInitAsksForNothingWithoutALoader(t *testing.T) {
	t.Parallel()

	if cmd := outsideProject().Init(); cmd != nil {
		t.Error("Init returned a command with no registry loader configured")
	}
}

type capturingScreen struct{ keylessScreen }

func (capturingScreen) Capturing() bool { return true }

func (s capturingScreen) Update(tea.Msg) (Screen, tea.Cmd) { return s, nil }

func TestAScreenCapturingTextKeepsTheKeys(t *testing.T) {
	t.Parallel()

	m := outsideProject()
	m.screens[m.current] = capturingScreen{}
	before := m.CurrentTitle()

	for _, key := range []string{"4", "1", "q", "?", "tab"} {
		m = press(t, m, key)

		if m.Quitting() {
			t.Fatalf("%q quit while a screen was capturing text", key)
		}
		if m.HelpOpen() {
			t.Fatalf("%q opened help while a screen was capturing text", key)
		}
		if got := m.CurrentTitle(); got != before {
			t.Fatalf("%q moved the selection to %q while a screen was capturing text", key, got)
		}
	}
}

func TestCtrlCQuitsWhileAScreenCapturesText(t *testing.T) {
	t.Parallel()

	m := outsideProject()
	m.screens[m.current] = capturingScreen{}

	if m = press(t, m, "ctrl+c"); !m.Quitting() {
		t.Error("ctrl+c did not quit while a screen was capturing text")
	}
}

func TestIndexOfAnUnknownTitleFallsBackToTheFirst(t *testing.T) {
	t.Parallel()

	if got := outsideProject().indexOf("Nonexistent"); got != 0 {
		t.Errorf("indexOf an unknown title = %d, want 0", got)
	}
}

type keylessScreen struct{}

func (keylessScreen) Title() string                      { return "Keyless" }
func (keylessScreen) Capturing() bool                    { return false }
func (keylessScreen) Keys() []KeyHelp                    { return nil }
func (s keylessScreen) Update(tea.Msg) (Screen, tea.Cmd) { return s, nil }
func (keylessScreen) View(int, int) string               { return "nothing here" }

func TestHelpSaysWhenATabAddsNoKeys(t *testing.T) {
	t.Parallel()

	m := outsideProject()
	m.screens[m.current] = keylessScreen{}

	m = press(t, m, "?")

	if view := render(m); !strings.Contains(view, "adds no keys of its own") {
		t.Errorf("the help overlay does not say the tab adds no keys:\n%s", view)
	}
}

func TestWrapBreaksOnWidth(t *testing.T) {
	t.Parallel()

	got := wrap("one two three four five six seven eight nine ten", 20)

	for _, line := range strings.Split(got, "\n") {
		if len(line) > 20 {
			t.Errorf("line %q is longer than 20 columns", line)
		}
	}
	if !strings.Contains(got, "\n") {
		t.Error("wrap produced no line break")
	}
	if strings.Join(strings.Fields(got), " ") != "one two three four five six seven eight nine ten" {
		t.Errorf("wrap changed the words: %q", got)
	}
}

func TestWrapHandlesEdges(t *testing.T) {
	t.Parallel()

	if got := wrap("", 40); got != "" {
		t.Errorf("wrap of an empty string = %q", got)
	}
	if got := wrap("word", 1); got != "word" {
		t.Errorf("a single word must not be broken: %q", got)
	}
}

func TestRuleIsEmptyWithNoWidth(t *testing.T) {
	t.Parallel()

	if got := rule(0); got != "" {
		t.Errorf("rule(0) = %q, want empty", got)
	}
	if got := ansi.Strip(rule(5)); got != "─────" {
		t.Errorf("rule(5) = %q", got)
	}
}

func TestTheStatusLineNeverOverflowsTheWidth(t *testing.T) {
	t.Parallel()

	m := outsideProject()
	m = m.SetStatus(strings.Repeat("a very long message ", 20))

	for _, width := range []int{1, 2, 10, 40, 100, 200} {
		m.width = width

		got := ansi.Strip(m.renderStatus())
		if len([]rune(got)) > width {
			t.Errorf("at width %d the status line is %d columns:\n%s", width, len([]rune(got)), got)
		}
	}
}

func TestTruncate(t *testing.T) {
	t.Parallel()

	cases := []struct {
		in    string
		width int
		want  string
	}{
		{in: "short", width: 10, want: "short"},
		{in: "exactly", width: 7, want: "exactly"},
		{in: "truncated", width: 5, want: "trun…"},
		{in: "anything", width: 1, want: "…"},
		{in: "anything", width: 0, want: ""},
		{in: "anything", width: -3, want: ""},
		{in: "", width: 5, want: ""},
		{in: "wîdé rünes", width: 6, want: "wîdé …"},
	}

	for _, tc := range cases {
		if got := truncate(tc.in, tc.width); got != tc.want {
			t.Errorf("truncate(%q, %d) = %q, want %q", tc.in, tc.width, got, tc.want)
		}
	}
}

func TestAPlaceholderIgnoresEveryMessage(t *testing.T) {
	t.Parallel()

	type odd struct{}

	p := localScreen()

	updated, cmd := p.Update(odd{})
	if cmd != nil {
		t.Error("a placeholder returned a command")
	}
	if updated.Title() != "Local" {
		t.Errorf("title = %q", updated.Title())
	}

	updated, cmd = p.Update(keyPress("z"))
	if cmd != nil {
		t.Error("a placeholder returned a command for a key press")
	}
	if updated.Capturing() {
		t.Error("a placeholder reports that it captures text")
	}
}

func TestAnInstallFinishedReachesEveryScreen(t *testing.T) {
	t.Parallel()

	root := demoProject(t, "schema: spec-driven\n")
	reader := &ProjectReader{CLI: fullProject(), Root: root}

	m := New(Options{Version: "0.1.0", InProject: true, ProjectRoot: root, ProjectRead: reader})

	updated, cmd := m.Update(InstallFinished{Name: "minimalist"})
	m = updated.(Model)

	if cmd == nil {
		t.Fatal("an install did not reach the Project screen")
	}

	updated, _ = m.Update(cmd())
	m = updated.(Model)

	if got := flat(render(press(t, m, "1"))); !strings.Contains(got, "add-auth") {
		t.Errorf("the project was not re-read after the install:\n%s", got)
	}
}

func TestInitReadsTheProjectInsideOne(t *testing.T) {
	t.Parallel()

	root := demoProject(t, "schema: spec-driven\n")
	reader := &ProjectReader{CLI: fullProject(), Root: root}

	m := New(Options{Version: "0.1.0", InProject: true, ProjectRoot: root, ProjectRead: reader})

	if m.Init() == nil {
		t.Error("Init asked for nothing inside a project with a reader")
	}
}

func TestInitAsksForNothingOutsideAProject(t *testing.T) {
	t.Parallel()

	root := demoProject(t, "schema: spec-driven\n")
	reader := &ProjectReader{CLI: fullProject(), Root: root}

	m := New(Options{Version: "0.1.0", ProjectRead: reader})

	if m.Init() != nil {
		t.Error("Init read a project from outside one")
	}
}
