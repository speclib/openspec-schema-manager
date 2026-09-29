package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/speclib/openspec-schema-manager/internal/compose"
	"github.com/speclib/openspec-schema-manager/internal/config"
	"github.com/speclib/openspec-schema-manager/internal/graph"
	"github.com/speclib/openspec-schema-manager/internal/openspec"
	"github.com/speclib/openspec-schema-manager/internal/registry"
	"github.com/speclib/openspec-schema-manager/internal/source"
	"github.com/speclib/openspec-schema-manager/internal/tui"
)

var version = "dev"

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr, isTerminal(os.Stdout)))
}

func isTerminal(f *os.File) bool {
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}

func run(args []string, stdout, stderr io.Writer, terminal bool) int {
	opts, err := parseFlags(args, stdout)
	switch {
	case errors.Is(err, errUsage):
		return 0
	case err != nil:
		fmt.Fprintln(stderr, "ossm:", err)
		return 2
	}

	if opts.showVersion {
		fmt.Fprintln(stdout, strings.TrimSpace(version))
		return 0
	}

	if !terminal {
		fmt.Fprintln(stderr, "ossm: this is a terminal application and standard output is not a terminal.")
		fmt.Fprintln(stderr, "Run it in a terminal, or use --version or --help, which need none.")
		return 1
	}

	if err := start(opts); err != nil {
		fmt.Fprintln(stderr, "ossm:", err)
		return 1
	}

	return 0
}

func appOptions(opts options) (tui.Options, error) {
	cfg, paths, err := config.Load()
	if err != nil {
		return tui.Options{}, err
	}

	workDir, err := os.Getwd()
	if err != nil {
		return tui.Options{}, err
	}

	appOpts := tui.Options{
		Version:  strings.TrimSpace(version),
		WorkDir:  workDir,
		Config:   cfg,
		Paths:    paths,
		Recents:  config.Recents{Path: paths.RecentsFile, Cap: cfg.RecentsCap},
		Drafts:   compose.Drafts{Dir: paths.DraftsDir},
		OpenPath: opts.path,
	}

	if root, err := openspec.FindProjectRoot(workDir); err == nil {
		appOpts.InProject = true
		appOpts.ProjectRoot = root
	}

	appOpts.Resolver = &tui.Resolver{
		Fetcher: source.NewGit(paths.SchemaCache),
		ASCII:   !graph.UnicodeAvailable(),
	}

	cli := openspec.NewExec()

	appOpts.Installer = &tui.Installer{
		Fetcher:   source.NewGit(paths.SchemaCache),
		CLI:       cli,
		Root:      appOpts.ProjectRoot,
		InProject: appOpts.InProject,
	}

	if appOpts.InProject {
		appOpts.ProjectRead = &tui.ProjectReader{CLI: cli, Root: appOpts.ProjectRoot}
	}

	appOpts.Comparer = &tui.Comparer{
		Fetcher: source.NewGit(paths.SchemaCache),
	}

	appOpts.Registry = &tui.RegistryLoader{
		Store: registry.Store{
			Path: paths.RegistryFile,
			URL:  cfg.RegistryURL,
		},
		Fetcher: registry.NewHTTPFetcher(),
		CLI:     cli,
		WorkDir: workDir,
		TTL:     cfg.RegistryTTL,
	}

	return appOpts, nil
}

func start(opts options) error {
	appOpts, err := appOptions(opts)
	if err != nil {
		return err
	}

	_, err = tea.NewProgram(tui.New(appOpts)).Run()

	return err
}
