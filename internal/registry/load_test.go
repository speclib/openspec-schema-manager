package registry

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func fixture(t *testing.T, name string) []byte {
	t.Helper()

	raw, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("reading the fixture: %v", err)
	}

	return raw
}

func TestParseReadsTheFixtureRegistry(t *testing.T) {
	t.Parallel()

	doc, err := Parse("registry.json", fixture(t, "registry.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(doc.Schemas) != 4 {
		t.Fatalf("got %d entries, want 4", len(doc.Schemas))
	}

	first := doc.Schemas[0]
	if first.ID != "danielhanold/superspec" {
		t.Errorf("the first entry is %q; file order must be preserved", first.ID)
	}
	if first.Name != "SuperSpec" {
		t.Errorf("name = %q; upstream casing must survive", first.Name)
	}
	if len(first.Artifacts) != 9 {
		t.Errorf("got %d artifacts, want 9", len(first.Artifacts))
	}

	pinned := doc.Schemas[1]
	if !pinned.Pinned() || pinned.RefLabel() != "v0.2.0" {
		t.Errorf("the pinned entry reports %q", pinned.RefLabel())
	}

	retired := doc.Schemas[2]
	if !retired.Deprecated() {
		t.Error("the deprecated entry does not report itself as deprecated")
	}
	if retired.SupersededBy != "speclib/tinychange" {
		t.Errorf("superseded_by = %q", retired.SupersededBy)
	}
}

func TestAnEmptyRegistryIsValid(t *testing.T) {
	t.Parallel()

	doc, err := Parse("empty.json", fixture(t, "empty.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(doc.Schemas) != 0 {
		t.Errorf("got %d entries, want none", len(doc.Schemas))
	}
}

func TestAnUnknownFieldIsIgnored(t *testing.T) {
	t.Parallel()

	doc, err := Parse("newer.json", fixture(t, "newer.json"))
	if err != nil {
		t.Fatalf("a registry newer than ossm was rejected: %v", err)
	}
	if len(doc.Schemas) != 1 {
		t.Fatalf("got %d entries, want 1", len(doc.Schemas))
	}
	if doc.Schemas[0].Name != "tinychange" {
		t.Errorf("the known fields did not survive: %+v", doc.Schemas[0])
	}
}

func TestParseRejectsWhatItCannotTrust(t *testing.T) {
	t.Parallel()

	const valid = `{"id":"a/b","name":"b","description":"d","artifacts":["x"],"source":{"repo":"https://e.test/r","path":"p"}}`

	cases := []struct {
		name    string
		body    string
		wantErr string
	}{
		{name: "not json", body: `{oops`, wantErr: "reading"},
		{name: "top level is an array", body: `[]`, wantErr: "reading"},
		{name: "top level is a string", body: `"a registry"`, wantErr: "reading"},
		{name: "no schemas key", body: `{"entries":[]}`, wantErr: "not a registry"},
		{name: "schemas is not an array", body: `{"schemas":"none"}`, wantErr: "reading"},
		{name: "no id", body: `{"schemas":[{"name":"b","description":"d","artifacts":["x"],"source":{"repo":"https://e.test/r","path":"p"}}]}`, wantErr: "position 0"},
		{name: "id with no owner", body: `{"schemas":[{"id":"nameonly","name":"b","description":"d","artifacts":["x"],"source":{"repo":"https://e.test/r","path":"p"}}]}`, wantErr: "<owner>/<name>"},
		{name: "id with capitals", body: `{"schemas":[{"id":"Owner/Name","name":"b","description":"d","artifacts":["x"],"source":{"repo":"https://e.test/r","path":"p"}}]}`, wantErr: "lowercase"},
		{name: "no name", body: `{"schemas":[{"id":"a/b","description":"d","artifacts":["x"],"source":{"repo":"https://e.test/r","path":"p"}}]}`, wantErr: "name is required"},
		{name: "blank name", body: `{"schemas":[{"id":"a/b","name":"  ","description":"d","artifacts":["x"],"source":{"repo":"https://e.test/r","path":"p"}}]}`, wantErr: "name is required"},
		{name: "no description", body: `{"schemas":[{"id":"a/b","name":"b","artifacts":["x"],"source":{"repo":"https://e.test/r","path":"p"}}]}`, wantErr: "description is required"},
		{name: "no artifacts", body: `{"schemas":[{"id":"a/b","name":"b","description":"d","artifacts":[],"source":{"repo":"https://e.test/r","path":"p"}}]}`, wantErr: "artifacts is required"},
		{name: "no repo", body: `{"schemas":[{"id":"a/b","name":"b","description":"d","artifacts":["x"],"source":{"path":"p"}}]}`, wantErr: "source.repo is required"},
		{name: "no path", body: `{"schemas":[{"id":"a/b","name":"b","description":"d","artifacts":["x"],"source":{"repo":"https://e.test/r"}}]}`, wantErr: "source.path is required"},
		{name: "unknown status", body: `{"schemas":[{"id":"a/b","name":"b","description":"d","artifacts":["x"],"source":{"repo":"https://e.test/r","path":"p"},"status":"retired"}]}`, wantErr: "active or deprecated"},
		{name: "superseded but active", body: `{"schemas":[{"id":"a/b","name":"b","description":"d","artifacts":["x"],"source":{"repo":"https://e.test/r","path":"p"},"superseded_by":"a/c"}]}`, wantErr: "not deprecated"},
		{name: "the second entry is bad", body: `{"schemas":[` + valid + `,{"id":"a/c","name":"c"}]}`, wantErr: "entry a/c"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := Parse("registry.json", []byte(tc.body))
			if err == nil {
				t.Fatalf("expected an error for %s", tc.body)
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("error %q does not mention %q", err, tc.wantErr)
			}
			if !strings.Contains(err.Error(), "registry.json") {
				t.Errorf("error %q does not name the source", err)
			}
		})
	}
}

func TestParseAcceptsALongDescription(t *testing.T) {
	t.Parallel()

	long := strings.Repeat("a", 250)
	body := `{"schemas":[{"id":"a/b","name":"b","description":"` + long + `","artifacts":["x"],"source":{"repo":"https://e.test/r","path":"p"}}]}`

	if _, err := Parse("registry.json", []byte(body)); err != nil {
		t.Fatalf("ossm rejected a description the registry should have caught: %v", err)
	}
}
