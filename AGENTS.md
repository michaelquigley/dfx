# AGENTS.md — dfx

dfx is a Go immediate-mode GUI framework built on Dear ImGui (via [cimgui-go](https://github.com/AllenDang/cimgui-go)). It wraps ImGui's per-frame draw model in a small, Go-idiomatic component/action API for building desktop tools. Everything lives in a single `package dfx` at the repo root.

This file is the orientation entry point for agents working in this repo. Read it before making changes; the deeper reference docs are linked at the bottom.

## Build / test / run

Via the `Makefile`:

- `make build` — `go install ./...`
- `make test` — `go test ./... -count=1` then `go vet ./...`
- `make clean` — `go clean` and remove installed binaries

Examples live in `examples/dfx_example_*`, each a standalone `main.go`:

- Run one with `go run ./examples/dfx_example_simple` (and so on).
- These open a real window — they need a display. Tests do not: the GLFW backend is created through the `createBackend` indirection (`app.go`), which tests override to run headless. Add display-free unit tests in the `*_test.go` pattern already used (`app_run_test.go`, `app_actions_test.go`, `fileTree_test.go`, etc.).

## Architecture

The whole model is small enough to hold in your head:

- **`Component`** (`component.go`) — the core interface: `Draw(*State)` to render each frame, `Actions() *ActionRegistry` for keyboard shortcuts. `Container` and `Func` are the embeddable/base implementations. Composite components may also implement `ChildActionProvider` (`ChildActions() []Component`) and `LocalActionProvider` (`LocalActions() *ActionRegistry`) to participate in hierarchical action resolution.
- **`State`** (`component.go`) — passed to every `Draw`: `Size`, `Position`, `IO` (imgui input/output), `App`, and `Parent`. Use `state.Size` for the component's allocated drawing area.
- **App loop** (`app.go`) — `dfx.New(root, Config{...})` then `app.Run()`. `Config` carries window setup plus lifecycle callbacks: `OnSetup`, `OnTick`, `OnClose`, `OnShutdown`, `OnSizeChange`, plus `MenuBar`, `Theme`, and font/theme toggles.
- **Action subsystem** (`action.go`) — keybindings resolved child → parent-local → app-global, first match wins. Modifier matching is **exact** (Ctrl+S does not match Ctrl+Shift+S). Conflicts are detected per-registry at registration. Invocations can be observed via `Config.OnAction` for usage telemetry. See `docs/current/child-actions.md` for the full model.
- **Composition** — `Container`, `Workspace` (`workspace.go`), `DashManager`/`Dash` (`dashManager.go`, `dash.go`), `HCollapse` (`hCollapse.go`), `MultiGrid` (`multiGrid.go`). Widgets: `Fader`, `VuMeter`, `VuWaterfall`, `FileTree`, `LogViewer`, `Toolbar`, `Controls`.

## Conventions

- **File naming**: camelCase for multi-word files (`dashManager.go`, `fileTree.go`, `hCollapse.go`, `logViewer.go`). Doc files in `docs/` are kebab-case.
- **UI primitives**: `github.com/AllenDang/cimgui-go/imgui`.
- **Logging**: `github.com/michaelquigley/df/dl`. **YAML/JSON marshaling**: `df/dd`. **Dependency injection**: `df/da`. **Errors**: `github.com/pkg/errors`.
- **Changelog**: `CHANGELOG.md` uses the in-house format — a `## Unreleased` slot pinned at the top that new entries are written into, then newest-first `## vX.Y.Z` releases below. Entries are prose paragraphs (not bullets), each led by `FEATURE` / `CHANGE` / `FIX` and ordered most-important-first; breaking changes lead with a **bolded** clause (no separate tag). No dates, no `### Added`-style subsections — not Keep a Changelog.
- **Docs**: built behavior is documented under `docs/current/`.

## Reference docs

- [`docs/current/layout-guide.md`](docs/current/layout-guide.md) — Dear ImGui layout and sizing (cursor model, child windows, tables, practical patterns) as used in dfx.
- [`docs/current/child-actions.md`](docs/current/child-actions.md) — the action subsystem: registries, precedence/cascade, key-string format, menu actions.

## Project memory

Durable knowledge about this project lives in `docs/journal/`, dated files `docs/journal/YYYY-MM-DD.md`. This is project memory; it does not go in harness-local storage (`.claude/` or equivalent), where it's invisible to every other harness and collaborator and dies with the host. Concretely: do not write to your harness's memory directory or memory tool for this project — even when the harness presents it as the default place for durable knowledge. That tool is the silo this convention exists to replace; the journal is the only durable home.

On arrival, read the most recent entries to pick up where the last session left off, before you start changing things. Treat them as prior-session context, not verified truth — if an entry conflicts with the code or a `docs/current/` doc, the code wins.

Write the smallest entry that carries the session's durable insight, and nothing more. The test for every line: *would a competent agent get this wrong, or waste time rediscovering it, working from the tree alone?* If it's recoverable by reading the code, the diff, `docs/current/`, or git history, leave it out.

That filter keeps four kinds of thing and discards the rest:

- **Decisions whose rationale isn't visible in the result** — why a value was chosen, what a line guards against, why something that looks like dead code or a no-op is load-bearing.
- **Deliberate non-actions** — a change you considered and chose not to make, so the next agent doesn't "fix" it. An unchanged file leaves no trace in a diff.
- **Couplings that span files** — two places that must move together, an ordering that matters, an assumption one file makes about another.
- **Live state** — what's unverified, unfinished, or waiting on something external.

Skip change inventories, restatements of the diff, and play-by-play of how you worked. There's no write-time approval gate; Michael reviews on commit. Append to the day's file if it exists, and write the few lines you'd want the next agent to read — honest and self-contained.

## Commits

The operator commits; agents don't. Never run `git commit` or `git push` in this repo. Finish the edit, leave the change in the working tree (staged is fine), report what changed, and hand off — the uncommitted diff is the review queue and the commit is the operator's act of acceptance. Approval of a change is not direction to commit; only an explicit instruction to commit is, and only for that commit.
