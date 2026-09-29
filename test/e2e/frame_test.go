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

	s.waitFor("Built in milestone 02")

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

	s.waitFor("Built in milestone 02")

	s.send("\t")
	s.waitFor("Built in milestone 06")

	s.send("\t")
	s.waitFor("Built in milestone 07")

	s.send("\t")
	s.waitFor("Built in milestone 05")

	s.send("\t")
	s.waitFor("Built in milestone 02")
}

func TestATabIsReachableByNumber(t *testing.T) {
	s := newSession(t, sessionOptions{})

	s.waitFor("Built in milestone 02")

	s.send("4")
	s.waitFor("Built in milestone 07")

	s.send("1")
	s.waitFor("Built in milestone 05")
}

func TestHelpOpensAndClosesWithoutQuitting(t *testing.T) {
	s := newSession(t, sessionOptions{})

	s.waitFor("Built in milestone 02")

	s.send("?")
	s.waitFor("Everywhere")
	s.waitFor("quit from anywhere")

	s.send("q")
	s.waitFor("Built in milestone 02")

	s.send("?")
	s.waitFor("Everywhere")
	s.send("\x1b")
	s.waitFor("Built in milestone 02")
}

func TestQuittingExitsZero(t *testing.T) {
	s := newSession(t, sessionOptions{})

	s.waitFor("Built in milestone 02")
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
	s.waitFor("Built in milestone 02")
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
