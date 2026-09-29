package tui

import (
	"os"
	"strings"
)

// EditorCommand resolves the editor from the environment, following the
// convention $VISUAL, then $EDITOR, then vi.
//
// The value is split on spaces so EDITOR="code --wait" works. No shell is
// involved: passing it through sh -c would let a crafted value do more than
// edit, and buys only quoting no real editor setting needs.
func EditorCommand(getenv func(string) string) (string, []string) {
	for _, key := range []string{"VISUAL", "EDITOR"} {
		if value := strings.TrimSpace(getenv(key)); value != "" {
			fields := strings.Fields(value)
			return fields[0], fields[1:]
		}
	}

	return "vi", nil
}

func editorFromEnv() (string, []string) {
	return EditorCommand(os.Getenv)
}
