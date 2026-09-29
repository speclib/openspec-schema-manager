package tui

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var updateReference = flag.Bool("update", false, "rewrite docs/keys.md from the code")

const referencePath = "../../docs/keys.md"

func TestWriteKeyReference(t *testing.T) {
	want := KeyReference()

	if *updateReference {
		if err := os.WriteFile(referencePath, []byte(want), 0o644); err != nil {
			t.Fatalf("writing the reference: %v", err)
		}
		return
	}

	raw, err := os.ReadFile(referencePath)
	if err != nil {
		t.Fatalf("reading %s: %v\nrun: go test ./internal/tui -run TestWriteKeyReference -update", referencePath, err)
	}

	if string(raw) != want {
		t.Errorf("%s is out of date; run: go test ./internal/tui -run TestWriteKeyReference -update", filepath.Clean(referencePath))
	}
}

func TestTheReferenceCoversEveryScreen(t *testing.T) {
	t.Parallel()

	got := KeyReference()

	for _, want := range []string{"## Everywhere", "## Project", "## Registry", "## Local", "## Composer", "## Modes"} {
		if !strings.Contains(got, want) {
			t.Errorf("the reference has no %q section:\n%s", want, got)
		}
	}

	for _, want := range []string{"ctrl+c", "open a schema folder by path", "install into this project", "send to the composer", "compare with the registry", "write the schema"} {
		if !strings.Contains(got, want) {
			t.Errorf("the reference does not mention %q", want)
		}
	}
}

func TestTheReferenceIsATable(t *testing.T) {
	t.Parallel()

	got := KeyReference()

	if !strings.Contains(got, "| Key") {
		t.Errorf("the reference draws no table:\n%s", got)
	}

	for _, line := range strings.Split(got, "\n") {
		if !strings.HasPrefix(line, "|") {
			continue
		}
		if strings.Count(line, "|") != 3 {
			t.Errorf("a table row is malformed: %q", line)
		}
	}
}

func TestAnEmptyKeyTableSaysSo(t *testing.T) {
	t.Parallel()

	if got := keyTable(nil); !strings.Contains(got, "no keys of its own") {
		t.Errorf("keyTable(nil) = %q", got)
	}
}
