package arch_test

import (
	"encoding/json"
	"io"
	"os/exec"
	"strings"
	"testing"
)

const modulePath = "github.com/speclib/openspec-schema-manager"

type listedPackage struct {
	ImportPath   string
	Deps         []string
	TestImports  []string
	XTestImports []string
}

func listPackages(t *testing.T) []listedPackage {
	t.Helper()

	cmd := exec.Command("go", "list", "-json", "./...")
	cmd.Dir = "../.."

	out, err := cmd.Output()
	if err != nil {
		var stderr string
		if ee, ok := err.(*exec.ExitError); ok {
			stderr = string(ee.Stderr)
		}
		t.Fatalf("go list failed: %v\n%s", err, stderr)
	}

	var pkgs []listedPackage
	dec := json.NewDecoder(strings.NewReader(string(out)))
	for {
		var p listedPackage
		switch err := dec.Decode(&p); err {
		case nil:
			pkgs = append(pkgs, p)
		case io.EOF:
			if len(pkgs) == 0 {
				t.Fatal("go list reported no packages")
			}
			return pkgs
		default:
			t.Fatalf("decoding go list output: %v", err)
		}
	}
}

func TestOnlyTheTUIAndTheCommandDependOnTheTUI(t *testing.T) {
	const tui = modulePath + "/internal/tui"

	allowed := map[string]bool{
		tui:                      true,
		modulePath + "/cmd/ossm": true,
	}

	pkgs := listPackages(t)

	var checked int
	for _, p := range pkgs {
		if !strings.HasPrefix(p.ImportPath, modulePath) || allowed[p.ImportPath] {
			continue
		}
		checked++

		for _, dep := range append(append(append([]string{}, p.Deps...), p.TestImports...), p.XTestImports...) {
			if dep == tui {
				t.Errorf("%s depends on %s; the TUI depends on the other packages and never the reverse", p.ImportPath, tui)
			}
		}
	}

	if checked == 0 {
		t.Fatal("no packages were checked; the boundary test is not looking at anything")
	}
}

func TestEveryInternalPackageIsListed(t *testing.T) {
	want := []string{
		"internal/arch",
		"internal/compose",
		"internal/config",
		"internal/graph",
		"internal/openspec",
		"internal/registry",
		"internal/schema",
		"internal/source",
		"internal/tui",
	}

	seen := make(map[string]bool)
	for _, p := range listPackages(t) {
		seen[strings.TrimPrefix(p.ImportPath, modulePath+"/")] = true
	}

	for _, w := range want {
		if !seen[w] {
			t.Errorf("package %s is missing; the layout in openspec/config.yaml names it", w)
		}
	}
}
