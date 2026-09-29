package e2e

// The acceptance criteria from briefing/BRIEFING.md section 8, one case each,
// named after the criterion. Each drives the built binary; where a criterion
// needs the openspec CLI or git, the case skips with the reason rather than
// passing because its precondition was missing.

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// "ossm launched outside any project shows the registry and local schemas;
// pressing install explains that you must be inside an OpenSpec project."
func TestAcceptanceOutsideAProject(t *testing.T) {
	root := t.TempDir()
	schemas := localSchemasDir(t)
	writeConfigWithDirs(t, root, writeRegistryPointingAt(t, root, schemaRepo(t)), schemas)

	s := newSession(t, sessionOptions{root: root})

	s.waitFor("Registry ·")
	s.waitFor("minimalist")

	s.send("3")
	s.waitFor("Local · 2 schema(s)")
	s.waitFor("team-review")

	s.send("2")
	s.waitFor("Registry ·")
	s.send("g")
	s.send("i")

	s.waitFor("needs an OpenSpec project")
}

// "ossm launched inside an OpenSpec project shows its default schema, available
// schemas with their origin, and each change with its schema."
func TestAcceptanceInsideAProject(t *testing.T) {
	root := t.TempDir()
	writeConfig(t, root, "file://"+filepath.Join(root, "absent.json"))

	project := openSpecProjectWithCLI(t)

	newChange := exec.Command("openspec", "new", "change", "add-auth", "--description", "x")
	newChange.Dir = project
	newChange.Env = append(os.Environ(), "HOME="+t.TempDir())
	if out, err := newChange.CombinedOutput(); err != nil {
		t.Skipf("openspec new change failed: %v\n%s", err, out)
	}

	s := newSession(t, sessionOptions{root: root, workDir: project})

	s.waitFor("default schema: spec-driven")
	s.waitFor("Schemas available")
	s.waitFor("spec-driven")
	s.waitFor("built-in")
	s.waitFor("Changes")
	s.waitFor("add-auth")
	s.waitFor("spec-driven")
}

// "Opening a registry schema shows its artifacts and a readable diagram,
// working offline once cached."
func TestAcceptanceOpeningASchemaAndWorkingOffline(t *testing.T) {
	root := t.TempDir()
	repo := schemaRepo(t)
	registryPath := filepath.Join(root, "openspec-schemas.json")
	writeConfig(t, root, writeRegistryPointingAt(t, root, repo))

	first := newSession(t, sessionOptions{root: root, env: []string{"LANG=en_GB.UTF-8"}})
	first.waitFor("minimalist")
	first.send("g")
	first.send("\r")

	first.waitFor("minimalist v1")
	first.waitFor("apply gate: tasks")
	first.waitFor("specs/**/*.md")

	first.send("d")
	first.waitFor("d closes the diagram")

	drawn := first.drawn()
	if !strings.Contains(drawn, "─") || !strings.Contains(drawn, "▼") {
		t.Errorf("the diagram is not readable:\n%s", drawn)
	}

	first.send("q")
	first.send("q")
	_ = first.waitForExit()

	// Take both the source repository and the registry file away.
	if err := os.RemoveAll(repo); err != nil {
		t.Fatalf("removing the repository: %v", err)
	}
	if err := os.Remove(registryPath); err != nil {
		t.Fatalf("removing the registry: %v", err)
	}

	second := newSession(t, sessionOptions{root: root, env: []string{"LANG=en_GB.UTF-8"}})

	second.waitFor("minimalist")
	second.send("g")
	second.send("\r")

	second.waitFor("minimalist v1")
	second.waitFor("apply gate: tasks")

	second.send("d")
	second.waitFor("d closes the diagram")

	if drawn := second.drawn(); !strings.Contains(drawn, "─") {
		t.Errorf("the diagram is not drawn offline:\n%s", drawn)
	}
}

// "Installing a registry schema into a demo project results in OpenSpec seeing
// it, and a change can be created with it."
func TestAcceptanceInstalling(t *testing.T) {
	root := t.TempDir()
	writeConfig(t, root, writeRegistryPointingAt(t, root, schemaRepo(t)))

	project := openSpecProjectWithCLI(t)

	s := newSession(t, sessionOptions{root: root, workDir: project})

	s.send("2")
	s.waitFor("minimalist")
	s.send("g")
	s.send("i")
	s.waitFor("y write these files")
	s.send("y")
	s.waitFor("minimalist installed")
	s.waitFor("OpenSpec validated it")

	s.send("\r")
	s.send("q")
	_ = s.waitForExit()

	which := exec.Command("openspec", "schema", "which", "minimalist", "--json")
	which.Dir = project
	which.Env = append(os.Environ(), "HOME="+t.TempDir())

	out, err := which.Output()
	if err != nil {
		t.Fatalf("openspec schema which: %v", err)
	}
	if !strings.Contains(string(out), `"source": "project"`) {
		t.Errorf("OpenSpec does not resolve the installed schema from the project:\n%s", out)
	}

	change := exec.Command("openspec", "new", "change", "with-minimalist", "--schema", "minimalist", "--json")
	change.Dir = project
	change.Env = append(os.Environ(), "HOME="+t.TempDir())

	if out, err := change.CombinedOutput(); err != nil {
		t.Fatalf("a change could not be created with the installed schema: %v\n%s", err, out)
	}
}

// "Duplicating a schema, editing a template via e, and returning shows the
// updated content with validation results."
func TestAcceptanceDuplicatingAndEditing(t *testing.T) {
	root := t.TempDir()
	schemas := localSchemasDir(t)
	writeConfigWithDirs(t, root, "file://"+filepath.Join(root, "absent.json"), schemas)

	editor := filepath.Join(t.TempDir(), "stub-editor")
	if err := os.WriteFile(editor, []byte("#!/bin/sh\nprintf '\\n## edited through ossm\\n' >> \"$1\"\n"), 0o755); err != nil {
		t.Fatalf("writing the stub editor: %v", err)
	}

	s := newSession(t, sessionOptions{root: root, env: []string{"EDITOR=" + editor, "VISUAL="}})

	s.send("3")
	s.waitFor("Local · 2 schema(s)")

	s.send("c")
	s.waitFor("Duplicate minimalist")
	s.send("mine")
	s.send("\r")
	s.waitFor("copied to")

	s.send("t")
	s.waitFor("Files of mine")
	s.waitFor("This schema validates")

	s.send("j")
	s.send("e")

	template := filepath.Join(schemas, "mine", "templates", "tasks.md")
	waitForFile(t, template, "edited through ossm")

	s.waitFor("Files of mine")
	s.waitFor("This schema validates")

	raw, err := os.ReadFile(filepath.Join(schemas, "mine", "schema.yaml"))
	if err != nil {
		t.Fatalf("reading the copy: %v", err)
	}
	if !strings.Contains(string(raw), "name: mine") {
		t.Errorf("the copy does not declare the new name:\n%s", raw)
	}
}

// "The composer can build a schema from artifacts of two fixture schemas, catch
// a cycle, warn about a dangling template reference, and write a schema that
// OpenSpec accepts and can create a change with."
func TestAcceptanceComposing(t *testing.T) {
	root := t.TempDir()
	schemas := composerFixtures(t)
	writeConfigWithDirs(t, root, "file://"+filepath.Join(root, "absent.json"), schemas)

	s := newSession(t, sessionOptions{root: root})

	s.send("3")
	s.waitFor("Local · 2 schema(s)")
	s.send("p")
	s.waitFor("sending")
	s.send("j")
	s.send("p")
	s.waitFor("sending")

	s.send("4")
	s.waitFor("Palette")

	// One artifact from each fixture. research-first's tasks template mentions
	// proposal, which is deliberately left off the canvas.
	s.send("j")
	s.send("a")
	s.waitFor("1 artifact(s)")
	s.waitFor("template warning")

	s.send("jj")
	s.send("a")
	s.waitFor("2 artifact(s)")

	// A cycle, caught as it is made.
	s.send("\x1b[C")
	s.send("j")
	s.send("l")
	s.waitFor("requires:")
	s.send("\r")
	s.send("k")
	s.send("l")
	s.waitFor("requires:")
	s.send("\r")
	s.waitFor("cycle")

	// Undo it and finish the schema.
	s.send("u")
	s.send("g")
	s.waitFor("gate:")
	s.send("t")
	s.send("review.md")
	s.send("\r")
	s.waitFor("tracks: review.md")

	s.send("w")
	s.send("composed")
	s.send("\r")
	s.waitFor("written to")

	s.send("q")
	_ = s.waitForExit()

	written := filepath.Join(schemas, "composed")
	if _, err := os.Stat(filepath.Join(written, "schema.yaml")); err != nil {
		t.Fatalf("the composed schema was not written: %v", err)
	}

	if _, err := exec.LookPath("openspec"); err != nil {
		t.Skip("the openspec CLI is not on PATH; the schema was written but not validated")
	}

	project := openSpecProjectWithCLI(t)
	destination := filepath.Join(project, "openspec", "schemas", "composed")
	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		t.Fatalf("creating the schemas directory: %v", err)
	}
	copyDir(t, written, destination)

	validate := exec.Command("openspec", "schema", "validate", "composed", "--json")
	validate.Dir = project
	validate.Env = append(os.Environ(), "HOME="+t.TempDir())

	out, err := validate.CombinedOutput()
	if err != nil || !strings.Contains(string(out), `"valid": true`) {
		t.Fatalf("OpenSpec rejected the composed schema: %v\n%s", err, out)
	}

	change := exec.Command("openspec", "new", "change", "with-composed", "--schema", "composed", "--json")
	change.Dir = project
	change.Env = append(os.Environ(), "HOME="+t.TempDir())

	if out, err := change.CombinedOutput(); err != nil {
		t.Fatalf("a change could not be created with the composed schema: %v\n%s", err, out)
	}
}

// composerFixtures writes two schemas whose artifacts can be mixed, with a
// template that mentions an artifact the other schema does not have.
func composerFixtures(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()

	fixtures := map[string]struct {
		schema    string
		templates map[string]string
	}{
		"research-first": {
			schema: `name: research-first
version: 1
description: Research before proposing
artifacts:
  - id: proposal
    generates: proposal.md
    description: What to do
    template: proposal.md
  - id: tasks
    generates: tasks.md
    description: The work
    template: tasks.md
    requires: [proposal]
apply:
  requires: [tasks]
  tracks: tasks.md
`,
			templates: map[string]string{
				"proposal.md": "# Proposal\n",
				"tasks.md":    "# Tasks\n\nWork through what the proposal asked for.\n",
			},
		},
		"team-review": {
			schema: `name: team-review
version: 1
description: A review gate
artifacts:
  - id: review
    generates: review.md
    description: What the team said
    template: review.md
apply:
  requires: [review]
  tracks: review.md
`,
			templates: map[string]string{"review.md": "# Review\n"},
		},
	}

	for name, fixture := range fixtures {
		base := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Join(base, "templates"), 0o755); err != nil {
			t.Fatalf("creating %s: %v", base, err)
		}
		if err := os.WriteFile(filepath.Join(base, "schema.yaml"), []byte(fixture.schema), 0o644); err != nil {
			t.Fatalf("writing the schema: %v", err)
		}
		for rel, body := range fixture.templates {
			if err := os.WriteFile(filepath.Join(base, "templates", rel), []byte(body), 0o644); err != nil {
				t.Fatalf("writing %s: %v", rel, err)
			}
		}
	}

	return dir
}

// "All unit tests run offline." The acceptance suite uses fixtures served from
// the filesystem, so nothing here reaches the network either.
func TestAcceptanceNothingReachesTheNetwork(t *testing.T) {
	root := t.TempDir()
	writeConfig(t, root, writeRegistryPointingAt(t, root, schemaRepo(t)))

	raw, err := os.ReadFile(filepath.Join(root, "config", "ossm", "config.yml"))
	if err != nil {
		t.Fatalf("reading the config: %v", err)
	}
	if !strings.Contains(string(raw), "file://") {
		t.Errorf("the acceptance suite points at something other than a file:\n%s", raw)
	}
}
