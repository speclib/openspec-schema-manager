package e2e

import (
	"os"
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
	s.waitFor("Not in an OpenSpec project")
	s.waitFor("installing needs a project")
}

func TestStartingInsideAProjectOpensOnProject(t *testing.T) {
	s := newSession(t, sessionOptions{workDir: openSpecProject(t)})

	s.waitFor("demo-app")
	s.waitFor("Built in milestone 05")
}

func TestTabMovesToTheNextTabAndWraps(t *testing.T) {
	s := newSession(t, sessionOptions{})

	s.waitFor("Registry ·")

	s.send("\t")
	s.waitFor("Built in milestone 06")

	s.send("\t")
	s.waitFor("Built in milestone 07")

	s.send("\t")
	s.waitFor("Built in milestone 05")

	s.send("\t")
	s.waitFor("Registry ·")
}

func TestATabIsReachableByNumber(t *testing.T) {
	s := newSession(t, sessionOptions{})

	s.waitFor("Registry ·")

	s.send("4")
	s.waitFor("Built in milestone 07")

	s.send("1")
	s.waitFor("Built in milestone 05")
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
