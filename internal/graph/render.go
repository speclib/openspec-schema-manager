package graph

import (
	"errors"
	"os"
	"strings"

	"github.com/pgavlin/mermaid-ascii/pkg/diagram"
	"github.com/pgavlin/mermaid-ascii/pkg/render"
)

var ErrNothingToDraw = errors.New("the schema declares no artifacts, so there is nothing to draw")

func (g Graph) Draw(ascii bool) (string, error) {
	if g.Len() == 0 {
		return "", ErrNothingToDraw
	}

	src, err := g.Mermaid()
	if err != nil {
		return "", err
	}

	cfg := diagram.DefaultConfig()
	cfg.UseAscii = ascii
	cfg.GraphDirection = "TD"
	cfg.StyleType = "cli"

	out, err := render.Render(src, cfg)
	if err != nil {
		return "", err
	}

	return strings.TrimRight(out, "\n"), nil
}

// UnicodeAvailable reports whether the environment's locale says UTF-8. It is
// the strongest signal a terminal application can read without querying the
// terminal, which a Bubble Tea model must not do mid-render.
func UnicodeAvailable() bool {
	for _, key := range []string{"LC_ALL", "LC_CTYPE", "LANG"} {
		if v := os.Getenv(key); v != "" {
			return strings.Contains(strings.ToUpper(v), "UTF-8") || strings.Contains(strings.ToUpper(v), "UTF8")
		}
	}

	return false
}
