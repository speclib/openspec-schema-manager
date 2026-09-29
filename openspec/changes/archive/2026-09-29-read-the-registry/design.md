# Design

## The model comes from the registry, not from here

`internal/registry`'s entry struct mirrors
`schema/openspec-schemas.schema.json` in the registry repository field for
field. Where the two disagree the registry wins, and the difference is recorded
in `NOTES.md`.

ossm validates what it needs to read an entry safely: the six required fields,
`id` shaped as `<owner>/<name>`, `artifacts` non-empty. It does not reimplement
the registry's linting. The 100-character cap on `description` is the registry's
gate to enforce before an entry is merged; a consumer that also refuses it just
means a bad entry breaks every user instead of one pull request.

Unknown fields are ignored. Go's `encoding/json` does this by default, which is
the behaviour wanted here and the opposite of what `internal/config` wants for
`config.yml`. The difference is deliberate: a config file is written by the user
in front of you, so a typo is worth stopping for. A registry file is written by
someone else and may be newer than the ossm reading it.

## Absent is not zero

`source.ref` absent means the default branch, and the default branch is not
always `main`. An entry model that stores `ref string` loses that distinction the
moment something writes `"main"` into the empty case. So `Ref` stays a string
and `Pinned()` reports whether it was set, and nothing in ossm ever substitutes a
branch name for an absent ref. Resolving it is the source fetcher's job in
milestone 04, and it asks the repository.

The same reasoning gives `Status()`, `Language()` and `License()` accessors that
apply the documented default rather than handing out a zero value.

## The cache file

```json
{"fetched_at": "2026-09-29T17:58:34Z", "url": "...", "registry": { ... }}
```

The fetch time is recorded rather than taken from the file's modification time,
because a checkout, a restore or a backup tool moves that time without the
content having been fetched. The URL is recorded so that changing
`registry_url` invalidates the cache instead of silently serving the old one.

Writing is to a temporary file in the same directory followed by a rename, so an
interrupted write cannot leave a half-written cache. That is also why validation
happens before the write and not after: a bad response has to be rejected while
the good cache is still the only thing on disk.

## Fetching

`internal/registry` takes an interface with one method rather than an
`*http.Client`, so the tests need no server:

```go
type Fetcher interface {
    Fetch(ctx context.Context, url string) ([]byte, error)
}
```

The HTTP implementation sets a timeout, refuses a body over a few megabytes, and
turns a non-success status into an error naming the status. A `file://` URL is
supported by the same interface, which is how the end to end tests point ossm at
a fixture registry without a server.

No unit test makes a network request. There is no networked test at all: the
registry repository already has a workflow checking that its sources resolve,
and duplicating that here would fail on a flaky connection and teach people to
ignore a red badge.

## Load order in the TUI

The cache is read synchronously on start because it is a local file read and
blocking on it is shorter than one frame. The fetch is a `tea.Cmd` returning a
message, which is what makes the list update behind the cached one without the
screen ever being empty while it works.

A refresh that fails leaves the list alone and writes to the status line. The
list is never replaced by an error, because the cached data is still the best
answer available.

## Built-in schemas

`openspec schema which --all --json` is the source. The adapter grows one
method here and the rest of it arrives in milestone 05. Its output shape is read
from the installed CLI rather than guessed, and the parser tolerates fields it
does not know for the same reason the registry parser does.

A missing `openspec` binary is not an error that hides the registry. The list
shows the registry and says the built-in schemas could not be listed. Refusing to
show anything because one of two sources is unavailable is the behaviour that
makes a tool useless on a machine that is only half set up.

## Ordering

Rows sort by `name`, case-insensitively, then by origin, then by `id`. Sorting
by name is what a user scanning the list expects; the tiebreakers exist because
`name` is explicitly not unique across the registry and two rows with the same
name must not swap places between runs.

## Capturing keys

The frame currently takes `tab` and the digits before the screen sees them,
which is what keeps the tab bar working. A filter breaks that: typing `4` into a
filter must not jump to the Composer.

So `Screen` grows `Capturing() bool`. While a screen is capturing, the frame
takes only `ctrl+c` and passes everything else down. This keeps the rule in one
place, and a screen that forgets to report it has a bug that a test in the frame
catches rather than one a user finds.

## Search

`id`, `name` and `description` are matched case-insensitively as substrings.
Fuzzy matching is not used here. With a registry of a few dozen entries a
substring match is predictable, and predictability is worth more than cleverness
on a list a user is scanning to compare things. If the registry grows past a few
hundred entries this is worth revisiting, and the search is one function.
