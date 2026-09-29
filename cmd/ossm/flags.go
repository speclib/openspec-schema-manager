package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
)

var errUsage = errors.New("usage requested")

type options struct {
	showVersion bool
	path        string
}

func parseFlags(args []string, out io.Writer) (options, error) {
	var opts options

	fs := flag.NewFlagSet("ossm", flag.ContinueOnError)
	fs.SetOutput(out)
	fs.BoolVar(&opts.showVersion, "version", false, "print the version and exit")
	fs.StringVar(&opts.path, "path", "", "open the schema folder at this path on launch")
	fs.Usage = func() {
		fmt.Fprintf(out, "ossm browses, installs, authors and composes OpenSpec workflow schemas.\n\nUsage:\n  ossm [flags]\n\nFlags:\n")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return opts, errUsage
		}
		return opts, err
	}

	if rest := fs.Args(); len(rest) > 0 {
		return opts, fmt.Errorf("unexpected argument %q", rest[0])
	}

	return opts, nil
}
