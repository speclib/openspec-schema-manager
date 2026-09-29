package e2e

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/charmbracelet/x/vt"
	"github.com/creack/pty"
)

const (
	cols        = 100
	rows        = 30
	waitTimeout = 10 * time.Second
	pollEvery   = 20 * time.Millisecond
)

var (
	binaryPath string
	buildErr   error
)

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "ossm-e2e-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "e2e: creating a build directory: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(dir)

	binaryPath = filepath.Join(dir, "ossm")

	cmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/ossm")
	cmd.Dir = filepath.Join("..", "..")

	if out, err := cmd.CombinedOutput(); err != nil {
		buildErr = fmt.Errorf("building ossm: %w\n%s", err, out)
	}

	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}

type session struct {
	t      *testing.T
	cmd    *exec.Cmd
	tty    *os.File
	mu     sync.Mutex
	screen *vt.Emulator
	done   chan struct{}
	wait   error
	once   sync.Once
}

type sessionOptions struct {
	workDir string
	root    string
	args    []string
	env     []string
}

func newSession(t *testing.T, opts sessionOptions) *session {
	t.Helper()

	if buildErr != nil {
		t.Fatal(buildErr)
	}

	root := opts.root
	if root == "" {
		root = t.TempDir()
	}
	if opts.workDir == "" {
		opts.workDir = t.TempDir()
	}

	cmd := exec.Command(binaryPath, opts.args...)
	cmd.Dir = opts.workDir
	cmd.Env = append(isolatedEnv(t, root), opts.env...)

	tty, err := pty.StartWithSize(cmd, &pty.Winsize{Cols: cols, Rows: rows})
	if err != nil {
		if errors.Is(err, os.ErrPermission) || errors.Is(err, os.ErrNotExist) {
			t.Skipf("no pseudo terminal available in this environment: %v", err)
		}
		t.Skipf("could not allocate a pseudo terminal: %v", err)
	}

	s := &session{
		t:      t,
		cmd:    cmd,
		tty:    tty,
		screen: vt.NewEmulator(cols, rows),
		done:   make(chan struct{}),
	}

	go s.read()
	go s.answer()

	t.Cleanup(func() { s.stop() })

	return s
}

func isolatedEnv(t *testing.T, root string) []string {
	t.Helper()

	env := []string{
		"XDG_CONFIG_HOME=" + filepath.Join(root, "config"),
		"XDG_CACHE_HOME=" + filepath.Join(root, "cache"),
		"XDG_STATE_HOME=" + filepath.Join(root, "state"),
		"HOME=" + filepath.Join(root, "home"),
		"TERM=xterm-256color",
	}

	for _, keep := range []string{"PATH", "LANG", "LC_ALL", "TMPDIR"} {
		if v, ok := os.LookupEnv(keep); ok {
			env = append(env, keep+"="+v)
		}
	}

	return env
}

// answer forwards what the emulator replies to the pty.
//
// The emulator answers a query (a device attributes request, a colour probe)
// by writing into an internal pipe, and that write blocks until something
// reads it. Bubble Tea sends those queries on startup, so without this the
// read loop stops mid-write while still holding the lock, and the session
// deadlocks before drawing anything.
func (s *session) answer() {
	_, _ = io.Copy(s.tty, s.screen)
}

func (s *session) read() {
	defer close(s.done)

	buf := make([]byte, 4096)
	for {
		n, err := s.tty.Read(buf)
		if n > 0 {
			s.mu.Lock()
			_, _ = s.screen.Write(buf[:n])
			s.mu.Unlock()
		}
		if err != nil {
			return
		}
	}
}

func (s *session) drawn() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.screen.String()
}

// flat collapses every run of whitespace to one space. The screen pads every
// line to its full width and wraps long text, so a phrase a user reads as one
// is several fragments in the buffer.
func flat(s string) string { return strings.Join(strings.Fields(s), " ") }

func (s *session) send(keys string) {
	s.t.Helper()

	if _, err := s.tty.WriteString(keys); err != nil {
		s.t.Fatalf("sending %q: %v", keys, err)
	}
}

func (s *session) waitFor(text string) {
	s.t.Helper()

	want := flat(text)

	deadline := time.Now().Add(waitTimeout)
	for time.Now().Before(deadline) {
		if strings.Contains(flat(s.drawn()), want) {
			return
		}
		time.Sleep(pollEvery)
	}

	s.t.Fatalf("waited %s for %q to appear; the screen held:\n%s", waitTimeout, text, s.drawn())
}

func (s *session) waitForAbsence(text string) {
	s.t.Helper()

	gone := flat(text)

	deadline := time.Now().Add(waitTimeout)
	for time.Now().Before(deadline) {
		if !strings.Contains(flat(s.drawn()), gone) {
			return
		}
		time.Sleep(pollEvery)
	}

	s.t.Fatalf("waited %s for %q to go away; the screen held:\n%s", waitTimeout, text, s.drawn())
}

func (s *session) waitForExit() error {
	s.once.Do(func() {
		select {
		case <-s.done:
		case <-time.After(waitTimeout):
			s.t.Fatalf("ossm did not exit within %s; the screen held:\n%s", waitTimeout, s.drawn())
		}
		s.wait = s.cmd.Wait()
	})

	return s.wait
}

func (s *session) stop() {
	if s.cmd.Process != nil && s.cmd.ProcessState == nil {
		_ = s.cmd.Process.Kill()
	}
	_ = s.tty.Close()
	s.once.Do(func() { s.wait = s.cmd.Wait() })
}

func exitCode(err error) int {
	if err == nil {
		return 0
	}

	var ee *exec.ExitError
	if errors.As(err, &ee) {
		return ee.ExitCode()
	}

	return -1
}

func openSpecProject(t *testing.T) string {
	t.Helper()

	dir := filepath.Join(t.TempDir(), "demo-app")
	if err := os.MkdirAll(filepath.Join(dir, "openspec", "changes"), 0o755); err != nil {
		t.Fatalf("creating a demo project: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "openspec", "config.yaml"), []byte("schema: spec-driven\n"), 0o644); err != nil {
		t.Fatalf("writing the project config: %v", err)
	}

	return dir
}

func testCommand(t *testing.T, root string, args ...string) *exec.Cmd {
	t.Helper()

	cmd := exec.Command(binaryPath, args...)
	cmd.Dir = root
	cmd.Env = isolatedEnv(t, root)

	return cmd
}
