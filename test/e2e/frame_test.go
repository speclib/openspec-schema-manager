package e2e

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestStartingOutsideAProjectOpensOnRegistry(t *testing.T) {
	s := newSession(t, sessionOptions{})

	s.waitFor("no OpenSpec project here")

	for _, tab := range []string{"Project", "Registry", "Local", "Composer"} {
		s.waitFor(tab)
	}

	s.waitFor("Registry ·")

	s.send("1")
	s.waitFor("There is no OpenSpec project here")
	s.waitFor("installing one needs a project")
}

func TestStartingInsideAProjectOpensOnProject(t *testing.T) {
	s := newSession(t, sessionOptions{workDir: openSpecProject(t)})

	s.waitFor("demo-app")
	s.waitFor("default schema")
}

func TestTabMovesToTheNextTabAndWraps(t *testing.T) {
	s := newSession(t, sessionOptions{})

	s.waitFor("Registry ·")

	s.send("\t")
	s.waitFor("Built in milestone 06")

	s.send("\t")
	s.waitFor("Built in milestone 07")

	s.send("\t")
	s.waitFor("There is no OpenSpec project here")

	s.send("\t")
	s.waitFor("Registry ·")
}

func TestATabIsReachableByNumber(t *testing.T) {
	s := newSession(t, sessionOptions{})

	s.waitFor("Registry ·")

	s.send("4")
	s.waitFor("Built in milestone 07")

	s.send("1")
	s.waitFor("There is no OpenSpec project here")
}

func TestHelpOpensAndClosesWithoutQuitting(t *testing.T) {
	s := newSession(t, sessionOptions{})

	s.waitFor("Registry ·")

	s.send("?")
	s.waitFor("Everywhere")
	s.waitFor("quit from anywhere")

	s.send("q")
	s.waitFor("Registry ·")

	s.send("?")
	s.waitFor("Everywhere")
	s.send("\x1b")
	s.waitFor("Registry ·")
}

func TestQuittingExitsZero(t *testing.T) {
	s := newSession(t, sessionOptions{})

	s.waitFor("Registry ·")
	s.send("q")

	if code := exitCode(s.waitForExit()); code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
}

func TestARedirectedRunSaysItNeedsATerminal(t *testing.T) {
	if buildErr != nil {
		t.Fatal(buildErr)
	}

	root := t.TempDir()

	cmd := testCommand(t, root)
	out, err := cmd.CombinedOutput()

	if exitCode(err) != 1 {
		t.Errorf("exit code = %d, want 1", exitCode(err))
	}
	if !strings.Contains(string(out), "not a terminal") {
		t.Errorf("output does not say it needs a terminal:\n%s", out)
	}
	if strings.Contains(string(out), "\x1b[") {
		t.Errorf("a redirected run emitted an escape sequence:\n%q", out)
	}
}

func TestARunWritesNothingOutsideItsRoots(t *testing.T) {
	root := t.TempDir()
	home := filepath.Join(root, "untouched-home")
	if err := os.MkdirAll(home, 0o755); err != nil {
		t.Fatalf("creating %s: %v", home, err)
	}

	s := newSession(t, sessionOptions{env: []string{"HOME=" + home}})
	s.waitFor("Registry ·")
	s.send("q")
	_ = s.waitForExit()

	entries, err := os.ReadDir(home)
	if err != nil {
		t.Fatalf("reading %s: %v", home, err)
	}
	if len(entries) != 0 {
		var names []string
		for _, e := range entries {
			names = append(names, e.Name())
		}
		t.Errorf("a run wrote into the home directory: %s", strings.Join(names, ", "))
	}
}

func TestVersionAndHelpNeedNoTerminal(t *testing.T) {
	if buildErr != nil {
		t.Fatal(buildErr)
	}

	for _, arg := range []string{"--version", "--help"} {
		cmd := testCommand(t, t.TempDir(), arg)

		out, err := cmd.CombinedOutput()
		if exitCode(err) != 0 {
			t.Errorf("%s exit code = %d, want 0\n%s", arg, exitCode(err), out)
		}
		if len(out) == 0 {
			t.Errorf("%s wrote nothing", arg)
		}
	}
}

func writeFixtureRegistry(t *testing.T, dir string) string {
	t.Helper()

	path := filepath.Join(dir, "openspec-schemas.json")
	body := `{
  "schemas": [
    {"id":"speclib/tinychange","name":"tinychange","description":"Lean specs to tasks workflow for very small changes.","artifacts":["specs","tasks"],"source":{"repo":"https://github.com/speclib/openspec-tinychange-schema","path":"openspec/schemas/tinychange"}},
    {"id":"danielhanold/superspec","name":"SuperSpec","description":"Spec-driven workflow wired into Superpowers skills and worktrees.","artifacts":["brainstorm","proposal","design","specs","tasks","plan","apply","verify","finalize"],"source":{"repo":"https://github.com/danielhanold/superspec","path":"openspec/schemas/superspec","ref":"v1.0.0"}}
  ]
}`

	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("writing the fixture registry: %v", err)
	}

	return "file://" + path
}

func writeConfig(t *testing.T, root, registryURL string) {
	t.Helper()

	dir := filepath.Join(root, "config", "ossm")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("creating the config directory: %v", err)
	}

	body := "registry_url: " + registryURL + "\nregistry_ttl: 0s\n"
	if err := os.WriteFile(filepath.Join(dir, "config.yml"), []byte(body), 0o644); err != nil {
		t.Fatalf("writing config.yml: %v", err)
	}
}

func TestTheRegistryTabListsAFixtureRegistry(t *testing.T) {
	root := t.TempDir()
	url := writeFixtureRegistry(t, root)
	writeConfig(t, root, url)

	s := newSession(t, sessionOptions{root: root})

	s.waitFor("tinychange")
	s.waitFor("SuperSpec")
	s.waitFor("@ v1.0.0")
	s.waitFor("@ default branch")
}

func TestFilteringNarrowsAndRestoresTheList(t *testing.T) {
	root := t.TempDir()
	url := writeFixtureRegistry(t, root)
	writeConfig(t, root, url)

	s := newSession(t, sessionOptions{root: root})
	s.waitFor("SuperSpec")

	s.send("/tiny")
	s.waitFor("filter: tiny")
	s.waitForAbsence("SuperSpec")

	s.send("\x1b")
	s.waitFor("SuperSpec")
}

func TestADigitTypedIntoTheFilterDoesNotSwitchTabs(t *testing.T) {
	root := t.TempDir()
	url := writeFixtureRegistry(t, root)
	writeConfig(t, root, url)

	s := newSession(t, sessionOptions{root: root})
	s.waitFor("tinychange")

	s.send("/4")
	s.waitFor("filter: 4")

	if drawn := flat(s.drawn()); strings.Contains(drawn, "Built in milestone 07") {
		t.Errorf("typing 4 into the filter jumped to the Composer:\n%s", s.drawn())
	}
}

func TestAnUnreachableRegistryWithNoCacheSaysSo(t *testing.T) {
	root := t.TempDir()
	writeConfig(t, root, "file://"+filepath.Join(root, "absent.json"))

	s := newSession(t, sessionOptions{root: root})

	s.waitFor("never been fetched")
	s.waitFor("Press r to try again")
}

func TestACachedRegistrySurvivesAnUnreachableURL(t *testing.T) {
	root := t.TempDir()
	url := writeFixtureRegistry(t, root)
	writeConfig(t, root, url)

	first := newSession(t, sessionOptions{root: root})
	first.waitFor("tinychange")
	first.send("q")
	_ = first.waitForExit()

	if err := os.Remove(strings.TrimPrefix(url, "file://")); err != nil {
		t.Fatalf("removing the fixture registry: %v", err)
	}

	second := newSession(t, sessionOptions{root: root})
	second.waitFor("tinychange")
	second.waitFor("cached")
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_CONFIG_GLOBAL=/dev/null",
		"GIT_CONFIG_SYSTEM=/dev/null",
		"GIT_AUTHOR_NAME=Test", "GIT_AUTHOR_EMAIL=test@example.test",
		"GIT_COMMITTER_NAME=Test", "GIT_COMMITTER_EMAIL=test@example.test",
	)

	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
}

func schemaRepo(t *testing.T) string {
	t.Helper()

	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not on PATH")
	}

	dir := t.TempDir()
	runGit(t, dir, "init", "--quiet", "--initial-branch", "master")

	base := filepath.Join(dir, "openspec", "schemas", "minimalist")
	if err := os.MkdirAll(filepath.Join(base, "templates", "specs"), 0o755); err != nil {
		t.Fatalf("creating the schema folder: %v", err)
	}

	files := map[string]string{
		filepath.Join(base, "schema.yaml"): `name: minimalist
version: 1
description: Lightweight schema for well-scoped, low-risk changes
artifacts:
  - id: specs
    generates: specs/**/*.md
    description: Specifications as user stories
    template: specs/spec.md
  - id: tasks
    generates: tasks.md
    description: Implementation checklist derived from the specs
    template: tasks.md
    requires: [specs]
apply:
  requires: [tasks]
  tracks: tasks.md
`,
		filepath.Join(base, "templates", "tasks.md"):         "# tasks\n",
		filepath.Join(base, "templates", "specs", "spec.md"): "# spec\n",
	}

	for path, body := range files {
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatalf("writing %s: %v", path, err)
		}
	}

	runGit(t, dir, "add", "-A")
	runGit(t, dir, "commit", "--quiet", "-m", "the schema")

	return dir
}

func writeRegistryPointingAt(t *testing.T, dir, repo string) string {
	t.Helper()

	path := filepath.Join(dir, "openspec-schemas.json")
	body := `{"schemas":[{"id":"speclib/minimalist","name":"minimalist","description":"Lightweight schema for well-scoped, low-risk changes.","artifacts":["specs","tasks"],"source":{"repo":"` + repo + `","path":"openspec/schemas/minimalist"}}]}`

	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("writing the registry: %v", err)
	}

	return "file://" + path
}

func TestOpeningASchemaShowsItsArtifacts(t *testing.T) {
	root := t.TempDir()
	repo := schemaRepo(t)
	writeConfig(t, root, writeRegistryPointingAt(t, root, repo))

	s := newSession(t, sessionOptions{root: root})
	s.waitFor("minimalist")

	// The built-in schema is listed before the registry arrives, so it holds
	// the selection. g moves back to the top of the settled list.
	s.send("g")
	s.send("\r")

	s.waitFor("minimalist v1")
	s.waitFor("apply gate: tasks")
	s.waitFor("tracks: tasks.md")
	s.waitFor("2 artifacts")
	s.waitFor("longest chain 2")
	s.waitFor("specs/**/*.md")
}

func TestTheDiagramDrawsBoxes(t *testing.T) {
	root := t.TempDir()
	repo := schemaRepo(t)
	writeConfig(t, root, writeRegistryPointingAt(t, root, repo))

	s := newSession(t, sessionOptions{root: root, env: []string{"LANG=en_GB.UTF-8"}})
	s.waitFor("minimalist")

	s.send("g")
	s.send("\r")
	s.waitFor("apply gate: tasks")

	s.send("d")
	s.waitFor("d closes the diagram")

	drawn := s.drawn()
	if !strings.Contains(drawn, "─") || !strings.Contains(drawn, "│") {
		t.Errorf("the diagram drew no box:\n%s", drawn)
	}
	if !strings.Contains(drawn, "▼") {
		t.Errorf("the diagram drew no arrow:\n%s", drawn)
	}
}

func TestEscReturnsToTheListWithTheSameRowSelected(t *testing.T) {
	root := t.TempDir()
	repo := schemaRepo(t)
	writeConfig(t, root, writeRegistryPointingAt(t, root, repo))

	s := newSession(t, sessionOptions{root: root})
	s.waitFor("minimalist")

	s.send("g")
	s.send("\r")
	s.waitFor("apply gate: tasks")

	s.send("\x1b")
	s.waitFor("Registry ·")
	s.waitForAbsence("apply gate")
}

func TestABuiltInSchemaOpensWithoutFetching(t *testing.T) {
	root := t.TempDir()
	writeConfig(t, root, "file://"+filepath.Join(root, "absent.json"))

	if _, err := exec.LookPath("openspec"); err != nil {
		t.Skip("the openspec CLI is not on PATH")
	}

	s := newSession(t, sessionOptions{root: root})
	s.waitFor("spec-driven")

	s.send("\r")
	s.waitFor("apply gate")
	s.waitFor("proposal")
}

func openSpecProjectWithCLI(t *testing.T) string {
	t.Helper()

	if _, err := exec.LookPath("openspec"); err != nil {
		t.Skip("the openspec CLI is not on PATH")
	}

	dir := filepath.Join(t.TempDir(), "demo-app")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("creating the project: %v", err)
	}

	cmd := exec.Command("openspec", "init", "--tools", "none", "--no-animation")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "HOME="+t.TempDir())

	if out, err := cmd.CombinedOutput(); err != nil {
		t.Skipf("openspec init failed: %v\n%s", err, out)
	}

	return dir
}

func TestTheProjectTabShowsTheProject(t *testing.T) {
	root := t.TempDir()
	writeConfig(t, root, "file://"+filepath.Join(root, "absent.json"))

	project := openSpecProjectWithCLI(t)

	s := newSession(t, sessionOptions{root: root, workDir: project})

	s.waitFor("demo-app")
	s.waitFor("default schema: spec-driven")
	s.waitFor("Schemas available")
	s.waitFor("spec-driven")
	s.waitFor("built-in")
	s.waitFor("no changes yet")
}

func TestInstallingASchemaIsSeenByOpenSpec(t *testing.T) {
	root := t.TempDir()
	repo := schemaRepo(t)
	writeConfig(t, root, writeRegistryPointingAt(t, root, repo))

	project := openSpecProjectWithCLI(t)

	s := newSession(t, sessionOptions{root: root, workDir: project})

	s.send("2")
	s.waitFor("minimalist")
	s.send("g")
	s.send("i")

	s.waitFor("Install minimalist")
	s.waitFor("openspec/schemas/minimalist")
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

	change := exec.Command("openspec", "new", "change", "try-it", "--schema", "minimalist", "--json")
	change.Dir = project
	change.Env = append(os.Environ(), "HOME="+t.TempDir())

	if out, err := change.CombinedOutput(); err != nil {
		t.Fatalf("a change could not be created with the installed schema: %v\n%s", err, out)
	}
}

func TestDecliningAnInstallLeavesTheProjectUntouched(t *testing.T) {
	root := t.TempDir()
	repo := schemaRepo(t)
	writeConfig(t, root, writeRegistryPointingAt(t, root, repo))

	project := openSpecProjectWithCLI(t)

	s := newSession(t, sessionOptions{root: root, workDir: project})

	s.send("2")
	s.waitFor("minimalist")
	s.send("g")
	s.send("i")
	s.waitFor("y write these files")

	s.send("n")
	s.waitFor("Registry ·")

	if _, err := os.Stat(filepath.Join(project, "openspec", "schemas", "minimalist")); err == nil {
		t.Error("declining wrote the schema anyway")
	}
}

func TestInstallingOutsideAProjectIsRefused(t *testing.T) {
	root := t.TempDir()
	repo := schemaRepo(t)
	writeConfig(t, root, writeRegistryPointingAt(t, root, repo))

	s := newSession(t, sessionOptions{root: root})

	s.waitFor("minimalist")
	s.send("g")
	s.send("i")

	s.waitFor("needs an OpenSpec project")
}
