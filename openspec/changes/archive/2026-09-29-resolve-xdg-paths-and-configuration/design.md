# Design

## Shape

```go
type Config struct {
    RegistryURL  string
    RegistryTTL  time.Duration
    SchemasDirs  []string
    RecentsCap   int
}

type Paths struct {
    ConfigFile   string
    RegistryFile string
    SchemaCache  string
    RecentsFile  string
    DraftsDir    string
}
```

`Load` returns both. A caller asks `Paths` where a file belongs and never joins
a path itself. The two are separate types because `Paths` is derived from the
environment and `Config` from a file, and confusing the two is how a setting
ends up controlling a location it should not.

## Strict decoding

`yaml.v3`'s decoder has `KnownFields(true)`, which turns an unknown key into an
error. That is the whole implementation of the typo requirement. The cost is
that a key added in a later version makes the file fail on an older ossm; that
is the right trade for a tool where a silently ignored setting looks exactly
like a setting that did not work.

## Durations

`registry_ttl` is a Go duration string (`24h`, `30m`). A bare number is
rejected, because `ttl: 30` reads as minutes to one person and seconds to
another. yaml.v3 does not decode into `time.Duration`, so the field is decoded
as a string and parsed, which also puts the field name in the error.

## Tilde expansion

Only a leading `~/` expands, and only in `schemas_dirs`. A tilde is a shell
convention, and this is the one setting a user writes by hand where a home
relative path is natural. The registry URL is a URL and the cache and state
roots come from the environment, so neither needs it.

Expansion happens at load, so every consumer sees an absolute path and none has
to remember to expand.

## Where the environment is read

`os.UserHomeDir` and `os.Getenv` are read through function fields on an
unexported struct so tests can supply an environment without touching the
process. `t.Setenv` would work, but it forbids parallel tests and this package
is pure enough to deserve them.

## Creating directories

`Load` creates nothing. The package exposes `EnsureDir(path string) error` and
callers use it immediately before writing. A read-only run of ossm leaves no
trace, which matters because browsing the registry is the common case and the
drafts directory belongs to a feature most users will never open.

## The default registry URL

`https://registry.speclib.org/api/v1/openspec-schemas.json`, the address
committed to in `openspec-schema-registry-9hi8`. It is a setting so a fork, a
mirror or a `file://` URL can be used without a rebuild, which is also how the
end to end tests point ossm at a fixture registry.

## Example configuration

`docs/config.example.yml` carries every key with its default and a line saying
what it does. It is not read by the program. A test asserts that every key in it
decodes, so the example cannot drift away from the struct.
