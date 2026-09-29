package tui

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/speclib/openspec-schema-manager/internal/schema"
)

type pathPrompt struct {
	open    bool
	typed   string
	problem string
	matches []string
	home    func() (string, error)
}

func newPathPrompt() *pathPrompt {
	return &pathPrompt{home: os.UserHomeDir}
}

func (p *pathPrompt) Open() {
	p.open = true
	p.typed = ""
	p.problem = ""
	p.matches = nil
}

func (p *pathPrompt) Close() {
	p.open = false
	p.problem = ""
	p.matches = nil
}

func (p *pathPrompt) expand(path string) string {
	if path != "~" && !strings.HasPrefix(path, "~/") {
		return path
	}

	home, err := p.home()
	if err != nil {
		return path
	}

	if path == "~" {
		return home
	}

	return filepath.Join(home, filepath.FromSlash(path[2:]))
}

// Submit resolves what was typed. It returns the directory when it holds a
// schema, and otherwise leaves the prompt open with a reason.
func (p *pathPrompt) Submit() (string, bool) {
	typed := strings.TrimSpace(p.typed)
	if typed == "" {
		p.problem = "type a directory holding a schema.yaml"
		return "", false
	}

	dir := p.expand(typed)

	info, err := os.Stat(dir)
	switch {
	case err != nil:
		p.problem = dir + " does not exist"
		return "", false
	case !info.IsDir():
		p.problem = dir + " is not a directory"
		return "", false
	}

	if _, err := os.Stat(filepath.Join(dir, schema.SchemaFile)); err != nil {
		p.problem = dir + " holds no " + schema.SchemaFile
		return "", false
	}

	return dir, true
}

func (p *pathPrompt) Complete() {
	p.problem = ""

	expanded := p.expand(p.typed)

	entries, err := filepath.Glob(expanded + "*")
	if err != nil {
		return
	}

	var dirs []string
	for _, entry := range entries {
		if info, err := os.Stat(entry); err == nil && info.IsDir() {
			dirs = append(dirs, entry)
		}
	}

	if len(dirs) == 0 {
		p.matches = nil
		return
	}

	sort.Strings(dirs)
	p.matches = dirs

	common := dirs[0]
	for _, dir := range dirs[1:] {
		common = commonPrefix(common, dir)
	}

	if len(common) > len(expanded) {
		p.typed = common
	}

	if len(dirs) == 1 {
		p.typed = dirs[0]
		p.matches = nil
	}
}

func commonPrefix(a, b string) string {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}

	for i := range n {
		if a[i] != b[i] {
			return a[:i]
		}
	}

	return a[:n]
}

func (p *pathPrompt) Type(text string) {
	p.typed += text
	p.problem = ""
	p.matches = nil
}

func (p *pathPrompt) Backspace() {
	if p.typed == "" {
		return
	}

	runes := []rune(p.typed)
	p.typed = string(runes[:len(runes)-1])
	p.problem = ""
	p.matches = nil
}

func (p *pathPrompt) View(width int) string {
	var b strings.Builder

	b.WriteString(headingStyle.Render("Open a schema folder"))
	b.WriteString("\n\n")
	b.WriteString("path: " + p.typed + "▏")

	if p.problem != "" {
		b.WriteString("\n\n")
		b.WriteString(p.problem)
	}

	if len(p.matches) > 0 {
		b.WriteString("\n")
		for i, match := range p.matches {
			if i == 8 {
				b.WriteString("\n  " + dimStyle.Render("and more"))
				break
			}
			b.WriteString("\n  " + dimStyle.Render(truncate(match, width-2)))
		}
	}

	b.WriteString("\n\n")
	b.WriteString(dimStyle.Render("enter open · tab complete · esc cancel"))

	return b.String()
}
