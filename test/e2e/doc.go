// Package e2e drives the built ossm binary the way a user would.
//
// Every case runs the compiled binary attached to a pseudo terminal, sends key
// presses to it and asserts on what it drew. Nothing here calls into ossm's
// packages: an end to end test that imports the code it is testing stops being
// one.
//
// Each case gets its own XDG config, cache and state roots inside a temporary
// directory, so a run cannot read or damage the machine's real ossm state.
package e2e
