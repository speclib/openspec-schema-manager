# Design

## Fetching with git

```go
type Fetcher interface {
    Fetch(ctx context.Context, src Source) (string, error)   // returns a directory
}
```

`Source` is the entry's repository, path and ref. The return is a directory on
disk holding `schema.yaml` and `templates/`, whether it was just fetched or
already cached.

The git implementation does, into a temporary directory beside the cache:

```
git init
git remote add origin <repo>
git fetch --depth 1 origin <ref>        # or: git fetch --depth 1 origin HEAD
git checkout --detach FETCH_HEAD
```

Not `git clone --branch`, because that fails on a commit SHA and an entry may
pin one. `git fetch origin HEAD` is what resolves the default branch: the
repository answers, which is the whole point, because one of the seeded
repositories uses `master` and computing `main` would fail on it.

Then the named path is copied out of the checkout into the cache location, and
the checkout is thrown away. Copying rather than keeping the clone means the
cache holds schema folders and nothing else, and the cache key does not have to
encode which clone a folder came from.

## No fallback for a source that is not a git repository

The briefing allows for fetching raw files instead. Nothing is built for it.
Every entry in the registry today is a git repository, and a raw-file fetch
cannot list a directory without a host-specific API, so the fallback would be a
GitHub adapter wearing a general name. `Fetcher` is one interface; adding one
later changes nothing above it.

## The cache key

```
$XDG_CACHE_HOME/ossm/schemas/<sha256(repo + "\x00" + ref + "\x00" + path)>/
```

A hash rather than a readable path. A repository URL, a ref and a path can all
hold characters a directory name cannot, and sanitising three of them into one
name is how two different sources end up sharing a directory. A `source.json`
inside the directory records the three in full, so the cache is still
inspectable by hand.

The ref recorded is what the entry asked for, not what it resolved to. Resolving
a branch to a commit and keying on that would refetch on every upstream push,
which is the opposite of caching. Pinning is the consumer's job at install time,
which is milestone 05.

## A failed fetch never damages the cache

Fetch into a temporary directory in the cache root, verify a `schema.yaml` is
there and parses, then rename into place. A rename within one filesystem is
atomic, and the temporary directory is in the cache root so it always is one.

## What is not fetched

A built-in schema and a local schema already have a directory. `Fetch` is not
called for them. The detail screen takes a resolved directory, and where it came
from is decided before the screen is built.

## Rendering

`github.com/pgavlin/mermaid-ascii/pkg/render.Render(src, *diagram.Config)`
is exported and works in process, under MIT. That settles open question 6: no
binary is shipped and nothing is shelled out to.

The `Config` carries `UseAscii` and `GraphDirection`. `UseAscii` is set from
whether the environment's locale says UTF-8, which is the same test a terminal
application can make without querying the terminal.

The renderer already draws `{{"x"}}` as a hexagon, so the gate shape chosen in
milestone 03 arrives on screen with nothing further to do.

Rendering happens once, into a string, when the diagram is opened or the schema
changes. The viewport then slices that string. Re-rendering per keystroke would
be visible: the layout is not cheap on a branchy graph.

## The viewport

The diagram is split into lines once. Drawing takes `lines[top:top+height]` and
slices each line by rune from `left`. Slicing by rune rather than by byte is what
keeps box-drawing characters intact; slicing by byte cuts them into mojibake.

Scrolling clamps at all four edges. `top` and `left` never go below zero and
never past the point where the last line or column is at the pane's edge, so a
diagram smaller than the pane simply never moves.

## The detail screen owns its schema

The screen is built from a resolved schema, its graph, its findings and its
origin. Fetching is a `tea.Cmd` returning a message, the same shape the registry
refresh uses, so an uncached schema does not block the frame while git runs.

The screen is pushed over the Registry tab rather than becoming a fifth tab. It
belongs to whatever listed it, and milestone 06 will open the same screen from
the Local tab. The frame gains no knowledge of it: the Registry screen holds it
and delegates, which is what keeps the frame's routing rules unchanged.

## Tests make no network request

`git` can clone from a local path. The tests build a real repository in a
temporary directory, with a schema in it, and fetch from that. That exercises the
real git commands, the default branch resolution and the ref handling, offline.
A repository whose default branch is `master` is one line of setup, so the case
that broke a naive implementation is covered rather than assumed.
