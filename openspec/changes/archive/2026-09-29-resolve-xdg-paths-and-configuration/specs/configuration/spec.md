# configuration

## ADDED Requirements

### Requirement: Configuration is read from one XDG location

ossm SHALL read `config.yml` from `ossm/` inside `XDG_CONFIG_HOME`, and from
`~/.config/ossm/` when `XDG_CONFIG_HOME` is unset or empty. It SHALL NOT search
any other directory.

#### Scenario: XDG_CONFIG_HOME is set

- **WHEN** `XDG_CONFIG_HOME` is set to a directory
- **THEN** the configuration is read from `<that directory>/ossm/config.yml`

#### Scenario: XDG_CONFIG_HOME is unset

- **WHEN** `XDG_CONFIG_HOME` is unset or empty
- **THEN** the configuration is read from `~/.config/ossm/config.yml`

### Requirement: A missing configuration file is not an error

ossm SHALL run on its defaults when no configuration file exists. Every setting
SHALL have a default that works without any configuration.

#### Scenario: No file exists

- **WHEN** no `config.yml` exists at the resolved location
- **THEN** loading succeeds and every setting holds its default

#### Scenario: A file sets some keys

- **WHEN** `config.yml` sets some keys and omits others
- **THEN** the keys it sets take effect and the omitted ones hold their defaults

### Requirement: A file that cannot be trusted is rejected by name

ossm SHALL fail with an error naming the file when `config.yml` is not valid
YAML, when its top level is not a mapping, when a value has the wrong type, or
when it carries a key ossm does not define. A typo SHALL be reported rather than
taking silent effect.

#### Scenario: The file is not valid YAML

- **WHEN** `config.yml` cannot be parsed
- **THEN** loading fails with an error naming the file

#### Scenario: A value has the wrong type

- **WHEN** a setting expecting a duration holds a number, or a setting expecting
  a list holds a string
- **THEN** loading fails with an error naming the file and the field

#### Scenario: An unknown key is present

- **WHEN** `config.yml` carries a key ossm does not define
- **THEN** loading fails with an error naming the file and the unknown key,
  rather than ignoring it

### Requirement: The settings and their defaults

The configuration SHALL carry the registry URL, the registry cache time to
live, the local schema directories and the recents cap, and nothing else. The
registry URL SHALL default to the published address of the OpenSpec schema
registry, the time to live to 24 hours, the schema directories to an empty list
and the recents cap to 20.

#### Scenario: Defaults are read with no file present

- **WHEN** no configuration file exists
- **THEN** the registry URL is the published registry address, the time to live
  is 24 hours, the schema directories are empty and the recents cap is 20

#### Scenario: A time to live of zero

- **WHEN** the time to live is set to zero
- **THEN** it is accepted and means the cache is always treated as stale, which
  is how a user asks for a fetch on every launch

#### Scenario: A negative recents cap

- **WHEN** the recents cap is set below zero
- **THEN** loading fails with an error naming the field, because a negative cap
  has no meaning

### Requirement: Every path ossm writes to is derived in one place

The configuration package SHALL be the only place that composes a path for the
registry cache file, the fetched schema cache directory, the recents file and
the drafts directory. Those SHALL sit under `XDG_CACHE_HOME` and
`XDG_STATE_HOME`, falling back to `~/.cache` and `~/.local/state`.

#### Scenario: The cache root is relocated

- **WHEN** `XDG_CACHE_HOME` is set to a different directory
- **THEN** the registry cache file and the schema cache directory both move with
  it, and the recents file and drafts directory do not

#### Scenario: The state root is relocated

- **WHEN** `XDG_STATE_HOME` is set to a different directory
- **THEN** the recents file and the drafts directory both move with it, and the
  cache paths do not

#### Scenario: Neither variable is set

- **WHEN** `XDG_CACHE_HOME` and `XDG_STATE_HOME` are both unset
- **THEN** the cache paths sit under `~/.cache/ossm` and the state paths under
  `~/.local/state/ossm`

### Requirement: A tilde in a configured directory expands to the home directory

A `schemas_dirs` entry beginning with `~/` SHALL expand to the user's home
directory. A tilde elsewhere in a path SHALL be left alone.

#### Scenario: A leading tilde

- **WHEN** a `schemas_dirs` entry is `~/work/schemas`
- **THEN** it resolves to `work/schemas` inside the user's home directory

#### Scenario: A tilde that is not leading

- **WHEN** a `schemas_dirs` entry is `/srv/~backup/schemas`
- **THEN** it is used unchanged, because the tilde is part of a real name

### Requirement: Directories are created when written to, not when read

Loading the configuration SHALL create no directory. A directory SHALL be
created by the call that writes into it.

#### Scenario: Configuration is loaded and nothing is written

- **WHEN** the configuration is loaded and no cache, state or draft is written
- **THEN** no directory is created under the cache root or the state root

#### Scenario: Something is written

- **WHEN** a caller asks for a path and declares it is about to write there
- **THEN** the parent directory is created if it does not exist
