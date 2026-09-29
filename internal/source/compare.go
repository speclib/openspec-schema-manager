package source

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/speclib/openspec-schema-manager/internal/registry"
)

type Verdict int

const (
	VerdictIdentical Verdict = iota
	VerdictDiffers
	VerdictNoEntry
	VerdictSeveralEntries
	VerdictUnreachable
	VerdictBuiltIn
)

type Comparison struct {
	Verdict    Verdict
	Ref        string
	Changed    []string
	Added      []string
	Removed    []string
	Candidates []string
	Reason     string
}

// Explain says what the comparison found and, on a difference, what it cannot
// tell. OpenSpec records nothing about where an installed schema came from, so
// a difference is a difference; calling it an update would be a guess.
func (c Comparison) Explain(name string) string {
	switch c.Verdict {
	case VerdictIdentical:
		return name + " matches what the registry offers at " + c.Ref + "."
	case VerdictDiffers:
		return fmt.Sprintf(
			"%s differs from the registry at %s: %d changed, %d only here, %d only there. ossm cannot tell an upstream change from a local edit, because OpenSpec records nothing about where an installed schema came from.",
			name, c.Ref, len(c.Changed), len(c.Added), len(c.Removed),
		)
	case VerdictNoEntry:
		return "nothing in the registry declares the name " + name + ", so there is nothing to compare against."
	case VerdictSeveralEntries:
		return fmt.Sprintf(
			"%d registry entries declare the name %s (%s). A name is not unique across the registry, so ossm cannot tell which one this came from.",
			len(c.Candidates), name, strings.Join(c.Candidates, ", "),
		)
	case VerdictUnreachable:
		return name + " could not be compared: " + c.Reason
	case VerdictBuiltIn:
		return name + " ships with OpenSpec, so it has no registry source to compare against."
	}

	return ""
}

func Compare(ctx context.Context, f Fetcher, installed string, builtIn bool, name string, entries []registry.Entry) Comparison {
	if builtIn {
		return Comparison{Verdict: VerdictBuiltIn}
	}

	var matches []registry.Entry
	for _, e := range entries {
		if e.Name == name {
			matches = append(matches, e)
		}
	}

	switch len(matches) {
	case 0:
		return Comparison{Verdict: VerdictNoEntry}
	case 1:
	default:
		var ids []string
		for _, e := range matches {
			ids = append(ids, e.ID)
		}
		sort.Strings(ids)
		return Comparison{Verdict: VerdictSeveralEntries, Candidates: ids}
	}

	entry := matches[0]
	src := Source{Repo: entry.Source.Repo, Path: entry.Source.Path, Ref: entry.Source.Ref}

	fetched, err := f.Fetch(ctx, src, true)
	if err != nil {
		return Comparison{Verdict: VerdictUnreachable, Ref: src.RefLabel(), Reason: err.Error()}
	}

	comparison, err := diff(installed, fetched)
	if err != nil {
		return Comparison{Verdict: VerdictUnreachable, Ref: src.RefLabel(), Reason: err.Error()}
	}

	comparison.Ref = src.RefLabel()

	return comparison
}

func diff(installed, fetched string) (Comparison, error) {
	here, err := readTree(installed)
	if err != nil {
		return Comparison{}, err
	}

	there, err := readTree(fetched)
	if err != nil {
		return Comparison{}, err
	}

	var comparison Comparison

	for rel, body := range here {
		other, ok := there[rel]
		switch {
		case !ok:
			comparison.Added = append(comparison.Added, rel)
		case body != other:
			comparison.Changed = append(comparison.Changed, rel)
		}
	}

	for rel := range there {
		if _, ok := here[rel]; !ok {
			comparison.Removed = append(comparison.Removed, rel)
		}
	}

	sort.Strings(comparison.Changed)
	sort.Strings(comparison.Added)
	sort.Strings(comparison.Removed)

	if len(comparison.Changed)+len(comparison.Added)+len(comparison.Removed) == 0 {
		comparison.Verdict = VerdictIdentical
	} else {
		comparison.Verdict = VerdictDiffers
	}

	return comparison, nil
}

// readTree reads a schema folder, skipping source.json. That file is ossm's
// own note about where a cached copy came from, and an installed schema never
// has one; counting it would make every comparison differ.
func readTree(dir string) (map[string]string, error) {
	files := make(map[string]string)

	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !d.Type().IsRegular() {
			return nil
		}

		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)

		if rel == "source.json" {
			return nil
		}

		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		files[rel] = string(body)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return files, nil
}
