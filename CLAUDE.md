# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

A Bubble Tea terminal UI for managing Docker Compose stacks on a single host.
Each direct subdirectory of `~/stacks/` (override with `--stacks-dir`) that
contains a compose file is one stack; the directory name is the stack name.

## Commands

```sh
make build        # -> bin/container-toolbox
make run          # go run .
make test         # go test ./...
make fmt          # gofmt -w .
make vet          # go vet ./...
make tidy         # go mod tidy

go test ./internal/stacks/ -run TestProjectName   # single test
```

Requires Go 1.25+, Docker Engine, and the `docker compose` (v2) CLI plugin.

## Architecture

The defining design decision is the **hybrid Docker client** (`internal/docker/`):

- `client.go` uses the **Docker Go SDK** over the local socket to *read* live
  container status (`ListContainersByProject`), grouping containers by the
  `com.docker.compose.project` label.
- `compose.go` shells out to the **`docker compose` CLI** for all *mutations*
  (`Up`/`Down`/`Restart`/`PullUp`) and log streaming. Commands run with
  `cmd.Dir` set to the stack directory so Compose auto-detects the file and
  derives the project name the same way the SDK labels report it.

These two sources are joined in `internal/tui/commands.go` (`join`): on-disk
stacks are matched to live status by the sanitized project name, with the
`com.docker.compose.project.working_dir` label as a fallback join key. The
join key must stay consistent — `stacks.ProjectName` reimplements Compose's own
directory-name normalization (lowercase, keep `[a-z0-9_-]`, trim leading
`_-`) so it lines up with the SDK label. Changing one without the other breaks
status display.

### Package layout

- `internal/stacks/` — pure filesystem discovery (no Docker dependency).
  `Discover` is deliberately tolerant: a missing root returns an empty slice,
  not an error, so the UI shows an empty state.
- `internal/docker/` — the hybrid client described above.
- `internal/tui/` — Bubble Tea Elm-architecture UI.

### TUI conventions (`internal/tui/`)

- Standard Elm loop: `model.go` holds `Model` + `Update`/`View`; `messages.go`
  defines all custom `tea.Msg` types and the `StackStatus`/`runState` domain
  types; `commands.go` holds the `tea.Cmd` functions that do the actual I/O off
  the update loop.
- Two view states (`stateList`, `stateLogs`) switched in `handleKey`.
- All Docker work happens inside `tea.Cmd` closures (in `commands.go`), never
  in `Update` — `Update` only routes messages. Add new behavior by defining a
  message type + a command, not by calling Docker from `Update`.
- **Log streaming pattern**: `StreamLogs` pushes lines into a channel; the UI
  turns that channel into a message stream by having `waitForLogCmd` read one
  line and reschedule itself. Cancellation is via the stored
  `logCancel context.CancelFunc` — call `stopLogs`/`quit` to kill the
  underlying process, never just drop the channel.
- A single `busy` flag serializes lifecycle actions: `dispatch` is a no-op
  while busy, so only one compose mutation runs at a time.
- Key bindings live only in `keys.go` (the `keys` var). Movement is handled
  natively by the list/viewport components; `keys` are app-level actions only.

## Testing

Tests cover the pure logic — `stacks.ProjectName`/`Discover` and the `join`
function. There are no tests against a live Docker daemon; keep daemon-touching
code thin and testable logic in pure functions.
