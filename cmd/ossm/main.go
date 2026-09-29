package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

var version = "dev"

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
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

	fmt.Fprintln(stderr, "ossm: the interface is not built yet")
	return 1
}
