// Package config resolves ossm's XDG locations and reads its configuration.
//
// It owns config.yml under XDG_CONFIG_HOME, the cache root under
// XDG_CACHE_HOME and the state root under XDG_STATE_HOME, along with the
// defaults for the registry URL, the cache TTL and the local schema
// directories.
//
// It does not read or write anything inside those roots beyond its own
// configuration and the recents list. Fetched registries belong to
// internal/registry and fetched schema folders to internal/source.
package config
