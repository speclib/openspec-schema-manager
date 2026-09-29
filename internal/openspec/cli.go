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

type CLI interface {
	Schemas(ctx context.Context, dir string) ([]ResolvedSchema, error)
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
		return nil, fmt.Errorf("openspec %s: %w: %s", args[0], err, bytes.TrimSpace(stderr.Bytes()))
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
