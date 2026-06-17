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
- **Action subsystem** (`action.go`) — keybindings resolved child → parent-local → app-global, first match wins. Modifier matching is **exact** (Ctrl+S does not match Ctrl+Shift+S). Conflicts are detected per-registry at registration. See `docs/current/child-actions.md` for the full model.
- **Composition** — `Container`, `Workspace` (`workspace.go`), `DashManager`/`Dash` (`dashManager.go`, `dash.go`), `HCollapse` (`hCollapse.go`), `MultiGrid` (`multiGrid.go`). Widgets: `Fader`, `VuMeter`, `VuWaterfall`, `FileTree`, `LogViewer`, `Toolbar`, `Controls`.

## Conventions

- **File naming**: camelCase for multi-word files (`dashManager.go`, `fileTree.go`, `hCollapse.go`, `logViewer.go`). Doc files in `docs/` are kebab-case.
- **UI primitives**: `github.com/AllenDang/cimgui-go/imgui`.
- **Logging**: `github.com/michaelquigley/df/dl`. **YAML/JSON marshaling**: `df/dd`. **Dependency injection**: `df/da`. **Errors**: `github.com/pkg/errors`.
- **Changelog**: there is no `CHANGELOG.md` yet. If one is added, use the in-house format — newest-first releases, prose entries each led by `FEATURE` / `CHANGE` / `FIX`, and an `## Unreleased` slot new entries go into. (Not Keep a Changelog.)
- **Docs**: built behavior is documented under `docs/current/`.

## Reference docs

- [`docs/current/layout-guide.md`](docs/current/layout-guide.md) — Dear ImGui layout and sizing (cursor model, child windows, tables, practical patterns) as used in dfx.
- [`docs/current/child-actions.md`](docs/current/child-actions.md) — the action subsystem: registries, precedence/cascade, key-string format, menu actions.
