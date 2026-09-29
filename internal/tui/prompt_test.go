package tui

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/speclib/openspec-schema-manager/internal/ansi"
)

func promptAt(t *testing.T, home string) *pathPrompt {
	t.Helper()

	p := newPathPrompt()
	p.home = func() (string, error) { return home, nil }
	p.Open()

	return p
}

func TestThePromptOpensAndCloses(t *testing.T) {
	t.Parallel()

	p := promptAt(t, t.TempDir())

	if !p.open {
		t.Error("Open did not open the prompt")
	}

	p.Type("abc")
	p.Close()

	if p.open {
		t.Error("Close did not close the prompt")
	}

	p.Open()
	if p.typed != "" {
		t.Errorf("reopening left %q typed", p.typed)
	}
}

func TestSubmittingAPathHoldingASchema(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	dir := localSchema(t, root, "quick", "quick")

	p := promptAt(t, root)
	p.Type(dir)

	got, ok := p.Submit()
	if !ok {
		t.Fatalf("a valid path was refused: %s", p.problem)
	}
	if got != dir {
		t.Errorf("Submit = %q, want %q", got, dir)
	}
}

func TestSubmittingRefusals(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	empty := filepath.Join(root, "empty")
	if err := os.Mkdir(empty, 0o755); err != nil {
		t.Fatalf("creating the directory: %v", err)
	}

	file := filepath.Join(root, "a-file")
	if err := os.WriteFile(file, []byte("x\n"), 0o644); err != nil {
		t.Fatalf("writing the file: %v", err)
	}

	cases := []struct {
		name    string
		typed   string
		wantErr string
	}{
		{name: "nothing typed", typed: "", wantErr: "type a directory"},
		{name: "whitespace", typed: "   ", wantErr: "type a directory"},
		{name: "does not exist", typed: filepath.Join(root, "absent"), wantErr: "does not exist"},
		{name: "not a directory", typed: file, wantErr: "is not a directory"},
		{name: "no schema", typed: empty, wantErr: "holds no schema.yaml"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p := promptAt(t, root)
			p.Type(tc.typed)

			if _, ok := p.Submit(); ok {
				t.Fatalf("%q was accepted", tc.typed)
			}
			if !strings.Contains(p.problem, tc.wantErr) {
				t.Errorf("problem = %q, want it to mention %q", p.problem, tc.wantErr)
			}
			if !p.open {
				t.Error("a refusal closed the prompt")
			}
		})
	}
}

func TestALeadingTildeExpands(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	localSchema(t, home, "mine", "mine")

	p := promptAt(t, home)
	p.Type("~/mine")

	got, ok := p.Submit()
	if !ok {
		t.Fatalf("a tilde path was refused: %s", p.problem)
	}
	if got != filepath.Join(home, "mine") {
		t.Errorf("Submit = %q, want %q", got, filepath.Join(home, "mine"))
	}

	if got := p.expand("~"); got != home {
		t.Errorf("a bare tilde expanded to %q", got)
	}
	if got := p.expand("/srv/~backup"); got != "/srv/~backup" {
		t.Errorf("a tilde that is not leading expanded to %q", got)
	}
}

func TestATildeWithNoHomeIsLeftAlone(t *testing.T) {
	t.Parallel()

	p := newPathPrompt()
	p.home = func() (string, error) { return "", errors.New("no home") }

	if got := p.expand("~/mine"); got != "~/mine" {
		t.Errorf("expand = %q, want the path unchanged", got)
	}
}

func TestCompletingOneMatch(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "only-one"), 0o755); err != nil {
		t.Fatalf("creating the directory: %v", err)
	}

	p := promptAt(t, root)
	p.Type(filepath.Join(root, "onl"))
	p.Complete()

	if p.typed != filepath.Join(root, "only-one") {
		t.Errorf("typed = %q, want the completed path", p.typed)
	}
	if len(p.matches) != 0 {
		t.Errorf("a single match was listed: %v", p.matches)
	}
}

func TestCompletingSeveralMatches(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	for _, name := range []string{"schemas-one", "schemas-two"} {
		if err := os.Mkdir(filepath.Join(root, name), 0o755); err != nil {
			t.Fatalf("creating %s: %v", name, err)
		}
	}

	p := promptAt(t, root)
	p.Type(filepath.Join(root, "sch"))
	p.Complete()

	if p.typed != filepath.Join(root, "schemas-") {
		t.Errorf("typed = %q, want the longest common prefix", p.typed)
	}
	if len(p.matches) != 2 {
		t.Errorf("matches = %v, want both", p.matches)
	}

	if got := flat(ansi.Strip(p.View(120))); !strings.Contains(got, "schemas-one") {
		t.Errorf("the matches are not shown:\n%s", got)
	}
}

func TestCompletingNothing(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	p := promptAt(t, root)
	p.Type(filepath.Join(root, "nothing-like-this"))

	before := p.typed
	p.Complete()

	if p.typed != before {
		t.Errorf("typed changed from %q to %q with nothing to complete", before, p.typed)
	}
	if p.problem != "" {
		t.Errorf("completing nothing reported %q", p.problem)
	}
}

func TestCompletingIgnoresFiles(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "candidate.md"), []byte("x\n"), 0o644); err != nil {
		t.Fatalf("writing the file: %v", err)
	}

	p := promptAt(t, root)
	p.Type(filepath.Join(root, "cand"))

	before := p.typed
	p.Complete()

	if p.typed != before {
		t.Errorf("a file was completed to: %q", p.typed)
	}
}

func TestTypingAndBackspacing(t *testing.T) {
	t.Parallel()

	p := promptAt(t, t.TempDir())

	p.Type("a")
	p.Type("b")
	p.Type("é")

	if p.typed != "abé" {
		t.Errorf("typed = %q", p.typed)
	}

	p.Backspace()
	if p.typed != "ab" {
		t.Errorf("after backspace, typed = %q; a multi-byte rune must go as one", p.typed)
	}

	p.Backspace()
	p.Backspace()
	p.Backspace()

	if p.typed != "" {
		t.Errorf("typed = %q, want empty", p.typed)
	}
}

func TestTheViewShowsWhatIsTyped(t *testing.T) {
	t.Parallel()

	p := promptAt(t, t.TempDir())
	p.Type("/some/path")

	got := flat(ansi.Strip(p.View(120)))
	for _, want := range []string{"Open a schema folder", "/some/path", "enter open", "tab complete", "esc cancel"} {
		if !strings.Contains(got, want) {
			t.Errorf("the prompt does not carry %q:\n%s", want, got)
		}
	}
}

func TestTheViewShowsAProblem(t *testing.T) {
	t.Parallel()

	p := promptAt(t, t.TempDir())
	p.Type(filepath.Join(t.TempDir(), "absent"))
	_, _ = p.Submit()

	if got := flat(ansi.Strip(p.View(120))); !strings.Contains(got, "does not exist") {
		t.Errorf("the problem is not shown:\n%s", got)
	}
}

func TestTheViewTruncatesALongMatchList(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	for i := range 12 {
		if err := os.Mkdir(filepath.Join(root, "dir-"+string(rune('a'+i))), 0o755); err != nil {
			t.Fatalf("creating a directory: %v", err)
		}
	}

	p := promptAt(t, root)
	p.Type(filepath.Join(root, "dir-"))
	p.Complete()

	if got := flat(ansi.Strip(p.View(120))); !strings.Contains(got, "and more") {
		t.Errorf("a long match list is not truncated:\n%s", got)
	}
}

func TestCommonPrefix(t *testing.T) {
	t.Parallel()

	cases := []struct {
		a, b, want string
	}{
		{a: "abc", b: "abd", want: "ab"},
		{a: "abc", b: "abc", want: "abc"},
		{a: "abc", b: "ab", want: "ab"},
		{a: "", b: "abc", want: ""},
		{a: "xyz", b: "abc", want: ""},
	}

	for _, tc := range cases {
		if got := commonPrefix(tc.a, tc.b); got != tc.want {
			t.Errorf("commonPrefix(%q, %q) = %q, want %q", tc.a, tc.b, got, tc.want)
		}
	}
}

func TestEditorCommand(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		env      map[string]string
		wantName string
		wantArgs []string
	}{
		{name: "VISUAL wins", env: map[string]string{"VISUAL": "nvim", "EDITOR": "vi"}, wantName: "nvim"},
		{name: "EDITOR when VISUAL is unset", env: map[string]string{"EDITOR": "nano"}, wantName: "nano"},
		{name: "vi when neither is set", env: nil, wantName: "vi"},
		{name: "blank VISUAL falls through", env: map[string]string{"VISUAL": "   ", "EDITOR": "nano"}, wantName: "nano"},
		{name: "arguments are kept", env: map[string]string{"EDITOR": "code --wait"}, wantName: "code", wantArgs: []string{"--wait"}},
		{name: "several arguments", env: map[string]string{"EDITOR": "emacsclient -c -a ''"}, wantName: "emacsclient", wantArgs: []string{"-c", "-a", "''"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			name, args := EditorCommand(func(key string) string { return tc.env[key] })

			if name != tc.wantName {
				t.Errorf("name = %q, want %q", name, tc.wantName)
			}
			if len(args) != len(tc.wantArgs) {
				t.Fatalf("args = %v, want %v", args, tc.wantArgs)
			}
			for i, want := range tc.wantArgs {
				if args[i] != want {
					t.Errorf("args[%d] = %q, want %q", i, args[i], want)
				}
			}
		})
	}
}

func TestEditorFromEnvReadsTheProcess(t *testing.T) {
	t.Setenv("VISUAL", "")
	t.Setenv("EDITOR", "my-editor")

	if name, _ := editorFromEnv(); name != "my-editor" {
		t.Errorf("name = %q, want my-editor", name)
	}
}
