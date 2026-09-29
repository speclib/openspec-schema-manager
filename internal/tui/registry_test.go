package tui

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/speclib/openspec-schema-manager/internal/ansi"
	"github.com/speclib/openspec-schema-manager/internal/openspec"
	"github.com/speclib/openspec-schema-manager/internal/registry"
)

const fixtureRegistry = `{
  "schemas": [
    {"id":"danielhanold/superspec","name":"SuperSpec","description":"Spec-driven workflow wired into Superpowers skills and worktrees.","artifacts":["brainstorm","proposal","design","specs","tasks","plan","apply","verify","finalize"],"source":{"repo":"https://github.com/danielhanold/superspec","path":"openspec/schemas/superspec"}},
    {"id":"lukk17/e2e-runbooks","name":"e2e-runbooks","description":"Capability-level e2e runbooks with behaviour-only assertions.","artifacts":["proposal","test-spec","tasks-template","run"],"source":{"repo":"https://github.com/Lukk17/openspec-schemas","path":"openspec/schemas/e2e-runbooks","ref":"v0.2.0"}},
    {"id":"speclib/retired","name":"retired","description":"An older workflow kept so its id still resolves.","artifacts":["proposal","tasks"],"source":{"repo":"https://e.test/r","path":"p"},"status":"deprecated","superseded_by":"speclib/tinychange"},
    {"id":"speclib/tinychange","name":"tinychange","description":"Lean specs to tasks workflow for very small changes.","artifacts":["specs","tasks"],"source":{"repo":"https://github.com/speclib/openspec-tinychange-schema","path":"openspec/schemas/tinychange"}}
  ]
}`

var loaderNow = time.Date(2026, 9, 29, 18, 0, 0, 0, time.UTC)

const loaderURL = "https://registry.speclib.test/openspec-schemas.json"

func newLoader(t *testing.T, body string, fetchErr error) *RegistryLoader {
	t.Helper()

	store := registry.Store{
		Path: filepath.Join(t.TempDir(), "registry.json"),
		URL:  loaderURL,
		Now:  func() time.Time { return loaderNow },
	}

	return &RegistryLoader{
		Store: store,
		Fetcher: registry.FetcherFunc(func(context.Context, string) ([]byte, error) {
			if fetchErr != nil {
				return nil, fetchErr
			}
			return []byte(body), nil
		}),
		CLI: &openspec.Fake{SchemasResult: []openspec.ResolvedSchema{
			{Name: "spec-driven", Source: openspec.SourcePackage, Path: "/nix/store/x/spec-driven"},
		}},
		TTL: 24 * time.Hour,
		Now: func() time.Time { return loaderNow },
	}
}

func seed(t *testing.T, l *RegistryLoader, body string, age time.Duration) {
	t.Helper()

	doc, err := registry.Parse("fixture", []byte(body))
	if err != nil {
		t.Fatalf("parsing the fixture: %v", err)
	}

	store := l.Store
	store.Now = func() time.Time { return loaderNow.Add(-age) }

	if _, err := store.Write(doc); err != nil {
		t.Fatalf("seeding the cache: %v", err)
	}
}

func screenWith(t *testing.T, l *RegistryLoader) *registryScreenModel {
	t.Helper()

	s := newRegistryScreen(l, nil, nil)

	updated, _ := s.Update(l.Load(context.Background()))

	model, ok := updated.(*registryScreenModel)
	if !ok {
		t.Fatalf("Update returned %T", updated)
	}

	return model
}

func view(s Screen) string { return flat(ansi.Strip(s.View(120, 20))) }

func key(s Screen, keys ...string) (Screen, tea.Cmd) {
	var cmd tea.Cmd
	for _, k := range keys {
		s, cmd = s.Update(keyPress(k))
	}
	return s, cmd
}

func TestTheCacheIsShownWithoutTheNetwork(t *testing.T) {
	t.Parallel()

	l := newLoader(t, fixtureRegistry, errors.New("no route to host"))
	seed(t, l, fixtureRegistry, 3*time.Hour)

	s := screenWith(t, l)

	got := view(s)
	for _, want := range []string{"5 schemas", "cached 3h ago", "SuperSpec", "tinychange", "spec-driven"} {
		if !strings.Contains(got, want) {
			t.Errorf("the view does not carry %q:\n%s", want, got)
		}
	}
}

func TestNoCacheAndNoNetworkSaysSo(t *testing.T) {
	t.Parallel()

	l := newLoader(t, fixtureRegistry, errors.New("no route to host"))

	got := view(screenWith(t, l))

	if !strings.Contains(got, "never been fetched") {
		t.Errorf("the view does not say the registry has never been fetched:\n%s", got)
	}
	if !strings.Contains(got, "Press r") {
		t.Errorf("the view does not offer a refresh:\n%s", got)
	}
	if !strings.Contains(got, "spec-driven") {
		t.Errorf("the built-in schema is missing even though OpenSpec answered:\n%s", got)
	}
}

func TestAStaleCacheIsRefreshedBehindIt(t *testing.T) {
	t.Parallel()

	l := newLoader(t, fixtureRegistry, nil)
	seed(t, l, `{"schemas":[]}`, 48*time.Hour)

	if !l.Stale() {
		t.Fatal("a 48-hour-old cache is not stale against a 24h time to live")
	}

	s := screenWith(t, l)
	if got := view(s); !strings.Contains(got, "1 schemas") {
		t.Errorf("the stale cache was not shown first:\n%s", got)
	}

	msg := l.RefreshCmd()()
	refreshed, ok := msg.(RegistryRefreshed)
	if !ok {
		t.Fatalf("RefreshCmd returned %T", msg)
	}
	if refreshed.Err != nil {
		t.Fatalf("the refresh failed: %v", refreshed.Err)
	}

	_, _ = s.Update(refreshed)
	if got := view(s); !strings.Contains(got, "5 schemas") {
		t.Errorf("the refreshed list did not replace the cached one:\n%s", got)
	}
}

func TestAFreshCacheIsNotStale(t *testing.T) {
	t.Parallel()

	l := newLoader(t, fixtureRegistry, nil)
	seed(t, l, fixtureRegistry, time.Hour)

	if l.Stale() {
		t.Error("an hour-old cache is stale against a 24h time to live")
	}
}

func TestAFailedRefreshLeavesTheListAlone(t *testing.T) {
	t.Parallel()

	l := newLoader(t, fixtureRegistry, errors.New("no route to host"))
	seed(t, l, fixtureRegistry, 3*time.Hour)

	s := screenWith(t, l)

	_, cmd := key(s, "r")
	if cmd == nil {
		t.Fatal("r produced no refresh command")
	}
	if got := view(s); !strings.Contains(got, "refreshing") {
		t.Errorf("the view does not report that a refresh is running:\n%s", got)
	}

	_, _ = s.Update(cmd())

	got := view(s)
	if !strings.Contains(got, "refresh failed") {
		t.Errorf("the view does not report the failure:\n%s", got)
	}
	if !strings.Contains(got, "SuperSpec") {
		t.Errorf("the failed refresh emptied the list:\n%s", got)
	}
}

func TestARefreshIsAlwaysAllowed(t *testing.T) {
	t.Parallel()

	l := newLoader(t, fixtureRegistry, nil)
	seed(t, l, `{"schemas":[]}`, time.Minute)

	s := screenWith(t, l)

	_, cmd := key(s, "r")
	if cmd == nil {
		t.Fatal("r produced no refresh command on a fresh cache")
	}

	_, _ = s.Update(cmd())
	if got := view(s); !strings.Contains(got, "5 schemas") {
		t.Errorf("the forced refresh did not take:\n%s", got)
	}
}

func TestTheBuiltInSchemaCannotBeInstalled(t *testing.T) {
	t.Parallel()

	l := newLoader(t, fixtureRegistry, nil)
	seed(t, l, fixtureRegistry, time.Hour)

	for _, row := range screenWith(t, l).rows {
		if row.Origin == registry.OriginBuiltIn && row.Installable {
			t.Errorf("%s is built in and offered as installable", row.Name)
		}
	}
}

func TestAMissingCLIDoesNotHideTheRegistry(t *testing.T) {
	t.Parallel()

	l := newLoader(t, fixtureRegistry, nil)
	l.CLI = &openspec.Fake{SchemasErr: openspec.ErrNotInstalled}
	seed(t, l, fixtureRegistry, time.Hour)

	got := view(screenWith(t, l))

	if !strings.Contains(got, "SuperSpec") {
		t.Errorf("the registry is hidden because the CLI is missing:\n%s", got)
	}
	if !strings.Contains(got, "openspec CLI is not on PATH") {
		t.Errorf("the missing CLI is passed over in silence:\n%s", got)
	}
}

func TestAFailingCLIIsReported(t *testing.T) {
	t.Parallel()

	l := newLoader(t, fixtureRegistry, nil)
	l.CLI = &openspec.Fake{SchemasErr: errors.New("exit status 3")}
	seed(t, l, fixtureRegistry, time.Hour)

	if got := view(screenWith(t, l)); !strings.Contains(got, "exit status 3") {
		t.Errorf("a failing CLI is not reported:\n%s", got)
	}
}

func TestNoCLIAtAllIsReported(t *testing.T) {
	t.Parallel()

	l := newLoader(t, fixtureRegistry, nil)
	l.CLI = nil
	seed(t, l, fixtureRegistry, time.Hour)

	if got := view(screenWith(t, l)); !strings.Contains(got, "no OpenSpec adapter") {
		t.Errorf("a missing adapter is not reported:\n%s", got)
	}
}

func TestFiltering(t *testing.T) {
	t.Parallel()

	l := newLoader(t, fixtureRegistry, nil)
	seed(t, l, fixtureRegistry, time.Hour)

	s := Screen(screenWith(t, l))

	s, _ = key(s, "/")
	if !s.Capturing() {
		t.Fatal("/ did not put the screen into capturing mode")
	}

	s, _ = key(s, "t", "i", "n", "y")

	got := view(s)
	if !strings.Contains(got, "filter: tiny") {
		t.Errorf("the filter text is not drawn:\n%s", got)
	}
	if !strings.Contains(got, "tinychange") {
		t.Errorf("the matching entry is missing:\n%s", got)
	}
	if strings.Contains(got, "SuperSpec") {
		t.Errorf("a non-matching entry survived the filter:\n%s", got)
	}

	s, _ = key(s, "esc")
	if s.Capturing() {
		t.Error("esc left the screen capturing")
	}
	if got := view(s); !strings.Contains(got, "SuperSpec") {
		t.Errorf("esc did not restore the full list:\n%s", got)
	}
}

func TestFilteringMatchesNothing(t *testing.T) {
	t.Parallel()

	l := newLoader(t, fixtureRegistry, nil)
	seed(t, l, fixtureRegistry, time.Hour)

	s, _ := key(Screen(screenWith(t, l)), "/", "z", "z", "z")

	got := view(s)
	if !strings.Contains(got, `Nothing matches "zzz"`) {
		t.Errorf("the view does not say what matched nothing:\n%s", got)
	}
}

func TestBackspaceWidensTheFilter(t *testing.T) {
	t.Parallel()

	l := newLoader(t, fixtureRegistry, nil)
	seed(t, l, fixtureRegistry, time.Hour)

	s, _ := key(Screen(screenWith(t, l)), "/", "t", "i", "n", "y", "backspace", "backspace", "backspace", "backspace")

	if got := view(s); !strings.Contains(got, "SuperSpec") {
		t.Errorf("backspacing the filter away did not restore the list:\n%s", got)
	}

	s, _ = key(s, "backspace")
	if got := view(s); !strings.Contains(got, "SuperSpec") {
		t.Errorf("backspacing an empty filter broke the list:\n%s", got)
	}
}

func TestEnterClosesTheFilterButKeepsIt(t *testing.T) {
	t.Parallel()

	l := newLoader(t, fixtureRegistry, nil)
	seed(t, l, fixtureRegistry, time.Hour)

	s, _ := key(Screen(screenWith(t, l)), "/", "t", "i", "n", "y", "enter")

	if s.Capturing() {
		t.Error("enter left the screen capturing")
	}

	got := view(s)
	if !strings.Contains(got, "tinychange") || strings.Contains(got, "SuperSpec") {
		t.Errorf("enter did not keep the filter applied:\n%s", got)
	}

	s, _ = key(s, "esc")
	if got := view(s); !strings.Contains(got, "SuperSpec") {
		t.Errorf("esc after enter did not clear the filter:\n%s", got)
	}
}

func TestTheSelectionSurvivesFilteringWhereItCan(t *testing.T) {
	t.Parallel()

	l := newLoader(t, fixtureRegistry, nil)
	seed(t, l, fixtureRegistry, time.Hour)

	s := screenWith(t, l)

	for s.selectedRow() == nil || s.selectedRow().Name != "tinychange" {
		if s.selected >= len(s.shown)-1 {
			t.Fatal("tinychange is not in the list")
		}
		s.move(1)
	}

	s.filter = "tiny"
	s.reselect()

	if row := s.selectedRow(); row == nil || row.Name != "tinychange" {
		t.Errorf("the selection did not survive a filter it still matches: %+v", row)
	}

	s.filter = "superspec"
	s.reselect()

	if row := s.selectedRow(); row == nil || row.Name != "SuperSpec" {
		t.Errorf("the selection did not move to the first match: %+v", row)
	}
}

func TestMovementStaysInBounds(t *testing.T) {
	t.Parallel()

	l := newLoader(t, fixtureRegistry, nil)
	seed(t, l, fixtureRegistry, time.Hour)

	s := screenWith(t, l)

	for range 20 {
		s.move(-1)
	}
	if s.selected != 0 {
		t.Errorf("moving up past the top left the selection at %d", s.selected)
	}

	for range 20 {
		s.move(1)
	}
	if s.selected != len(s.shown)-1 {
		t.Errorf("moving down past the bottom left the selection at %d of %d", s.selected, len(s.shown))
	}

	empty := &registryScreenModel{}
	empty.move(1)
	if empty.selected != 0 {
		t.Errorf("moving in an empty list left the selection at %d", empty.selected)
	}
}

func TestMovementKeys(t *testing.T) {
	t.Parallel()

	l := newLoader(t, fixtureRegistry, nil)
	seed(t, l, fixtureRegistry, time.Hour)

	s := screenWith(t, l)

	if _, _ = key(s, "down", "down"); s.selected != 2 {
		t.Errorf("two downs left the selection at %d, want 2", s.selected)
	}
	if _, _ = key(s, "up"); s.selected != 1 {
		t.Errorf("up left the selection at %d, want 1", s.selected)
	}
	if _, _ = key(s, "G"); s.selected != len(s.shown)-1 {
		t.Errorf("G left the selection at %d, want the last row", s.selected)
	}
	if _, _ = key(s, "g"); s.selected != 0 {
		t.Errorf("g left the selection at %d, want 0", s.selected)
	}
	if _, _ = key(s, "j", "j", "k"); s.selected != 1 {
		t.Errorf("jjk left the selection at %d, want 1", s.selected)
	}
}

func TestADeprecatedRowSaysSo(t *testing.T) {
	t.Parallel()

	l := newLoader(t, fixtureRegistry, nil)
	seed(t, l, fixtureRegistry, time.Hour)

	got := view(screenWith(t, l))

	if !strings.Contains(got, "deprecated") {
		t.Errorf("the deprecated row does not say so:\n%s", got)
	}
	if !strings.Contains(got, "speclib/tinychange") {
		t.Errorf("the deprecated row does not name its replacement:\n%s", got)
	}
}

func TestARowCarriesItsRefOrTheDefaultBranch(t *testing.T) {
	t.Parallel()

	l := newLoader(t, fixtureRegistry, nil)
	seed(t, l, fixtureRegistry, time.Hour)

	got := view(screenWith(t, l))

	if !strings.Contains(got, "@ v0.2.0") {
		t.Errorf("the pinned row does not carry its ref:\n%s", got)
	}
	if !strings.Contains(got, "@ default branch") {
		t.Errorf("an unpinned row does not say it tracks the default branch:\n%s", got)
	}
	if strings.Contains(got, "@ main") {
		t.Errorf("an absent ref was rendered as main:\n%s", got)
	}
}

func TestAnEmptyRegistrySaysSo(t *testing.T) {
	t.Parallel()

	l := newLoader(t, fixtureRegistry, nil)
	l.CLI = &openspec.Fake{}
	seed(t, l, `{"schemas":[]}`, time.Hour)

	if got := view(screenWith(t, l)); !strings.Contains(got, "holds no entries") {
		t.Errorf("an empty registry does not say so:\n%s", got)
	}
}

func TestTheListScrollsWhenItIsLongerThanThePane(t *testing.T) {
	t.Parallel()

	var entries []map[string]any
	for i := range 40 {
		entries = append(entries, map[string]any{
			"id":          "owner/schema" + string(rune('a'+i%26)) + string(rune('a'+i/26)),
			"name":        "schema" + string(rune('a'+i%26)) + string(rune('a'+i/26)),
			"description": "one of many",
			"artifacts":   []string{"specs", "tasks"},
			"source":      map[string]string{"repo": "https://e.test/r", "path": "p"},
		})
	}
	body, err := json.Marshal(map[string]any{"schemas": entries})
	if err != nil {
		t.Fatalf("building the fixture: %v", err)
	}

	l := newLoader(t, string(body), nil)
	l.CLI = &openspec.Fake{}
	seed(t, l, string(body), time.Hour)

	s := screenWith(t, l)

	first := flat(ansi.Strip(s.View(120, 10)))
	if strings.Count(first, "schema") < 5 {
		t.Errorf("the pane drew almost nothing:\n%s", first)
	}

	for range 39 {
		s.move(1)
	}

	last := flat(ansi.Strip(s.View(120, 10)))
	if first == last {
		t.Error("the list did not scroll when the selection moved past the pane")
	}
	if s.top == 0 {
		t.Error("the viewport did not move with the selection")
	}
}

func TestHumanAge(t *testing.T) {
	t.Parallel()

	cases := []struct {
		in   time.Duration
		want string
	}{
		{in: 10 * time.Second, want: "moments"},
		{in: 5 * time.Minute, want: "5m"},
		{in: 3 * time.Hour, want: "3h"},
		{in: 40 * time.Hour, want: "40h"},
		{in: 72 * time.Hour, want: "3d"},
	}

	for _, tc := range cases {
		if got := humanAge(tc.in); got != tc.want {
			t.Errorf("humanAge(%v) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestPad(t *testing.T) {
	t.Parallel()

	if got := pad("ab", 5); got != "ab   " {
		t.Errorf("pad = %q", got)
	}
	if got := pad("abcdef", 3); got != "abcdef" {
		t.Errorf("pad must not shorten: %q", got)
	}
}

func TestTheScreenReportsItsKeys(t *testing.T) {
	t.Parallel()

	s := newRegistryScreen(nil, nil, nil)

	var found []string
	for _, k := range s.Keys() {
		found = append(found, k.Key)
	}

	for _, want := range []string{"/", "esc", "r"} {
		if !contains(found, want) {
			t.Errorf("the screen does not report %q among %v", want, found)
		}
	}
}

func contains(haystack []string, needle string) bool {
	for _, h := range haystack {
		if h == needle {
			return true
		}
	}
	return false
}

func TestRefreshDoesNothingWithoutALoader(t *testing.T) {
	t.Parallel()

	s, cmd := key(newRegistryScreen(nil, nil, nil), "r")
	if cmd != nil {
		t.Error("r produced a command with no loader")
	}
	if s.Capturing() {
		t.Error("r left the screen capturing")
	}
}

func TestAnUnknownMessageIsIgnored(t *testing.T) {
	t.Parallel()

	type odd struct{}

	s, cmd := newRegistryScreen(nil, nil, nil).Update(odd{})
	if cmd != nil {
		t.Error("an unknown message produced a command")
	}
	if s.Title() != "Registry" {
		t.Errorf("title = %q", s.Title())
	}
}

func TestTheFrameLoadsTheRegistryOnInit(t *testing.T) {
	t.Parallel()

	l := newLoader(t, fixtureRegistry, nil)
	seed(t, l, fixtureRegistry, 48*time.Hour)

	m := New(Options{Version: "0.1.0", Registry: l})

	if cmd := m.Init(); cmd == nil {
		t.Fatal("Init returned no command with a registry loader configured")
	}
}

func TestLoaderUsesTheRealClockByDefault(t *testing.T) {
	t.Parallel()

	l := &RegistryLoader{}
	before := time.Now().Add(-time.Second)

	if l.now().Before(before) {
		t.Error("the loader did not fall back to the real clock")
	}
}

func TestLoadTreatsACorruptCacheAsNoCache(t *testing.T) {
	t.Parallel()

	l := newLoader(t, fixtureRegistry, nil)
	if err := os.MkdirAll(filepath.Dir(l.Store.Path), 0o755); err != nil {
		t.Fatalf("creating the cache directory: %v", err)
	}
	if err := os.WriteFile(l.Store.Path, []byte("{oops"), 0o644); err != nil {
		t.Fatalf("writing the cache: %v", err)
	}

	loaded := l.Load(context.Background())
	if loaded.HasCache {
		t.Error("a corrupt cache was reported as a cache")
	}
	if !l.Stale() {
		t.Error("a corrupt cache was not reported as stale")
	}
}

func TestTheFrameCarriesTheCachedRegistryBeforeInit(t *testing.T) {
	t.Parallel()

	l := newLoader(t, fixtureRegistry, nil)
	seed(t, l, fixtureRegistry, time.Hour)

	m := New(Options{Version: "0.1.0", Registry: l})

	if got := flat(ansi.Strip(m.Render())); !strings.Contains(got, "SuperSpec") {
		t.Errorf("the cached registry is not in the model before Init:\n%s", got)
	}
	if cmd := m.Init(); cmd != nil {
		t.Error("Init asked for a refresh on a fresh cache")
	}
}

func TestARefreshIsNotOverwrittenByTheCacheRead(t *testing.T) {
	t.Parallel()

	l := newLoader(t, fixtureRegistry, nil)

	m := New(Options{Version: "0.1.0", Registry: l})

	if got := flat(ansi.Strip(m.Render())); !strings.Contains(got, "never been fetched") {
		t.Fatalf("the empty state is not drawn before the refresh:\n%s", got)
	}

	cmd := m.Init()
	if cmd == nil {
		t.Fatal("Init asked for no refresh with no cache at all")
	}

	updated, _ := m.Update(cmd())
	m = updated.(Model)

	got := flat(ansi.Strip(m.Render()))
	if !strings.Contains(got, "SuperSpec") {
		t.Errorf("the refreshed registry is not drawn:\n%s", got)
	}
	if strings.Contains(got, "never fetched") {
		t.Errorf("the refresh was overwritten by the cache read:\n%s", got)
	}
}

func TestEnterOpensAndEscReturnsWithTheSameRowSelected(t *testing.T) {
	t.Parallel()

	l := newLoader(t, fixtureRegistry, nil)
	seed(t, l, fixtureRegistry, time.Hour)

	fetcher := &fakeFetcher{dir: schemaDir(t, "chain.yaml")}
	s := newRegistryScreen(l, &Resolver{Fetcher: fetcher, ASCII: true}, nil)
	_, _ = s.Update(l.Load(context.Background()))

	model := s.(*registryScreenModel)
	_, _ = key(s, "down")
	before := model.selectedRow().Name

	_, cmd := key(s, "enter")
	if cmd == nil {
		t.Fatal("enter produced no command")
	}
	if model.detail == nil {
		t.Fatal("enter did not open the detail")
	}

	_, _ = s.Update(cmd())

	if got := view(s); !strings.Contains(got, "apply gate") {
		t.Errorf("the detail is not drawn:\n%s", got)
	}

	_, _ = key(s, "esc")

	if model.detail != nil {
		t.Fatal("esc did not close the detail")
	}
	if got := model.selectedRow().Name; got != before {
		t.Errorf("the selection moved from %q to %q", before, got)
	}
}

func TestQClosesTheDetailRatherThanQuitting(t *testing.T) {
	t.Parallel()

	l := newLoader(t, fixtureRegistry, nil)
	seed(t, l, fixtureRegistry, time.Hour)

	fetcher := &fakeFetcher{dir: schemaDir(t, "chain.yaml")}
	s := newRegistryScreen(l, &Resolver{Fetcher: fetcher, ASCII: true}, nil)
	_, _ = s.Update(l.Load(context.Background()))

	_, cmd := key(s, "enter")
	_, _ = s.Update(cmd())

	_, _ = key(s, "q")

	if s.(*registryScreenModel).detail != nil {
		t.Error("q did not close the detail")
	}
	if got := view(s); !strings.Contains(got, "schemas ·") {
		t.Errorf("the list did not return:\n%s", got)
	}
}

func TestEnterOnAnEmptyListDoesNothing(t *testing.T) {
	t.Parallel()

	s := newRegistryScreen(nil, &Resolver{ASCII: true}, nil)

	if _, cmd := key(s, "enter"); cmd != nil {
		t.Error("enter on an empty list produced a command")
	}
	if s.(*registryScreenModel).detail != nil {
		t.Error("enter on an empty list opened a detail")
	}
}

func TestEnterWithNoResolverDoesNothing(t *testing.T) {
	t.Parallel()

	l := newLoader(t, fixtureRegistry, nil)
	seed(t, l, fixtureRegistry, time.Hour)

	s := newRegistryScreen(l, nil, nil)
	_, _ = s.Update(l.Load(context.Background()))

	if _, cmd := key(s, "enter"); cmd != nil {
		t.Error("enter produced a command with no resolver")
	}
}

func TestTheDetailTakesTheKeysAndTheHelp(t *testing.T) {
	t.Parallel()

	l := newLoader(t, fixtureRegistry, nil)
	seed(t, l, fixtureRegistry, time.Hour)

	fetcher := &fakeFetcher{dir: schemaDir(t, "chain.yaml")}
	s := newRegistryScreen(l, &Resolver{Fetcher: fetcher, ASCII: true}, nil)
	_, _ = s.Update(l.Load(context.Background()))

	_, cmd := key(s, "enter")
	_, _ = s.Update(cmd())

	var keys []string
	for _, k := range s.Keys() {
		keys = append(keys, k.Key)
	}
	if !contains(keys, "d") {
		t.Errorf("the help does not follow into the detail: %v", keys)
	}

	_, _ = key(s, "d")
	if got := view(s); !strings.Contains(got, "d closes the diagram") {
		t.Errorf("the detail did not take the d key:\n%s", got)
	}
}

func TestTheListKeysMentionEnter(t *testing.T) {
	t.Parallel()

	var keys []string
	for _, k := range newRegistryScreen(nil, nil, nil).Keys() {
		keys = append(keys, k.Key)
	}

	if !contains(keys, "enter") {
		t.Errorf("the list does not report enter: %v", keys)
	}
}
