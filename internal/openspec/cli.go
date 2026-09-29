package openspec

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"time"
)

const (
	SourcePackage = "package"
	SourceProject = "project"
	SourceUser    = "user"

	defaultTimeout = 15 * time.Second
)

var ErrNotInstalled = errors.New("the openspec CLI is not on PATH")

type ResolvedSchema struct {
	Name    string   `json:"name"`
	Source  string   `json:"source"`
	Path    string   `json:"path"`
	Shadows []string `json:"shadows"`
}

func (s ResolvedSchema) BuiltIn() bool {
	return s.Source == SourcePackage
}

type Change struct {
	Name           string `json:"name"`
	CompletedTasks int    `json:"completedTasks"`
	TotalTasks     int    `json:"totalTasks"`
	Status         string `json:"status"`
}

type ValidationIssue struct {
	Message string `json:"message"`
	Path    string `json:"path"`
	Level   string `json:"level"`
}

type Validation struct {
	Name   string            `json:"name"`
	Path   string            `json:"path"`
	Valid  bool              `json:"valid"`
	Issues []ValidationIssue `json:"issues"`
}

type CLI interface {
	Schemas(ctx context.Context, dir string) ([]ResolvedSchema, error)
	Changes(ctx context.Context, dir string) ([]Change, error)
	ChangeSchema(ctx context.Context, dir, change string) (string, error)
	ValidateSchema(ctx context.Context, dir, name string) (Validation, error)
}

type Exec struct {
	Binary  string
	Timeout time.Duration
}

func NewExec() Exec {
	return Exec{Binary: "openspec", Timeout: defaultTimeout}
}

func (e Exec) binary() string {
	if e.Binary != "" {
		return e.Binary
	}
	return "openspec"
}

func (e Exec) run(ctx context.Context, dir string, args ...string) ([]byte, error) {
	if _, err := exec.LookPath(e.binary()); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrNotInstalled, e.binary())
	}

	if e.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, e.Timeout)
		defer cancel()
	}

	cmd := exec.CommandContext(ctx, e.binary(), args...)
	cmd.Dir = dir

	// Without this, a cancelled command still blocks in Wait until every
	// process holding the output pipe exits, which a child of the CLI can
	// outlive the CLI itself.
	cmd.WaitDelay = time.Second

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return stdout.Bytes(), fmt.Errorf("openspec %s: %w: %s", args[0], err, bytes.TrimSpace(stderr.Bytes()))
	}

	return stdout.Bytes(), nil
}

func (e Exec) Schemas(ctx context.Context, dir string) ([]ResolvedSchema, error) {
	out, err := e.run(ctx, dir, "schema", "which", "--all", "--json")
	if err != nil {
		return nil, err
	}

	var schemas []ResolvedSchema
	if err := json.Unmarshal(out, &schemas); err != nil {
		return nil, fmt.Errorf("reading openspec schema which output: %w", err)
	}

	return schemas, nil
}

func (e Exec) Changes(ctx context.Context, dir string) ([]Change, error) {
	out, err := e.run(ctx, dir, "list", "--json")
	if err != nil {
		return nil, err
	}

	var listed struct {
		Changes []Change `json:"changes"`
	}
	if err := json.Unmarshal(out, &listed); err != nil {
		return nil, fmt.Errorf("reading openspec list output: %w", err)
	}

	return listed.Changes, nil
}

func (e Exec) ChangeSchema(ctx context.Context, dir, change string) (string, error) {
	out, err := e.run(ctx, dir, "status", "--change", change, "--json")
	if err != nil {
		return "", err
	}

	var status struct {
		SchemaName string `json:"schemaName"`
	}
	if err := json.Unmarshal(out, &status); err != nil {
		return "", fmt.Errorf("reading openspec status output: %w", err)
	}

	return status.SchemaName, nil
}

func (e Exec) ValidateSchema(ctx context.Context, dir, name string) (Validation, error) {
	out, err := e.run(ctx, dir, "schema", "validate", name, "--json")

	// A schema OpenSpec rejects makes the command exit non-zero while still
	// writing the report. The report is the answer; the exit status alone is
	// not, so it is only an error when nothing parseable came back.
	if len(out) > 0 {
		var v Validation
		if jsonErr := json.Unmarshal(out, &v); jsonErr == nil {
			return v, nil
		}
	}

	if err != nil {
		return Validation{}, err
	}

	return Validation{}, fmt.Errorf("reading openspec schema validate output for %s", name)
}
