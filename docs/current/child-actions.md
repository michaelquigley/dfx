# child actions and the dfx action subsystem

This document is the implementation-facing reference for keyboard actions in dfx.
It explains how actions are defined, resolved, and executed across components and the app.

## design goals

- predictable shortcut precedence
- clear extension points for custom composite components
- minimal hidden behavior
- safe registration with early conflict detection

## subsystem overview

The action subsystem has four main parts:

1. `Action`:
   - a parsed keybinding and handler
   - fields include `Id`, `Keys`, parsed `key`, parsed modifier bitmask, and optional menu label data
2. `ActionRegistry`:
   - a collection of actions
   - used by both app-level and component-level scopes
3. app event processing:
   - each frame, dfx gathers registries in precedence order
   - checks pressed key + current modifiers
   - executes first matching handler and stops
4. traversal interfaces:
- `ChildActionProvider` and `LocalActionProvider`
   - define how composite components participate in hierarchy resolution

## where actions live

- app-global actions live in `app.Actions()`
- component-local actions live in each component's local registry
- composite components can expose active children through `ChildActions()`

## canonical cascade and precedence

Traversal and execution always follow this order:

1. child component actions (deepest and most specific first)
2. parent-local actions
3. app-global actions

Additional ordering details:

- child traversal uses reverse child order to align with z-order expectations
- within one registry, actions are checked in registration order
- processing stops at the first matching action (`return` after handler call)

## key matching semantics

An action matches only when both conditions are true:

1. `imgui.IsKeyPressedBool(action.key)` is true for the current frame
2. `action.mods == currentMods` (exact modifier equality)

Important implication:

- modifier matching is exact, not subset-based
- for example, a `Ctrl+S` action will not match if `Ctrl+Shift+S` is pressed

## key string format and supported keys

`ActionRegistry.Register(id, keys, handler)` accepts strings like:

- `Ctrl+S`
- `Alt+Shift+F1`
- `Up`
- `[` or `-` (single-character special keys)

Supported key families:

- alpha: `A`-`Z`
- digits: `0`-`9`
- function keys: `F1`-`F12`
- navigation and control keys: `Enter`, `Esc`, `Tab`, `Backspace`, `Delete`, arrows, `Home`, `End`, `PageUp`, `PageDown`, `Space`
- single-character special keys: `-`, `=`, `[`, `]`, `;`, `'`, `,`, `.`, `/`

Supported modifiers:

- `Ctrl`
- `Shift`
- `Alt`
- `Super` (aliases: `Cmd`, `Win`)

## registration and conflict detection

Each `ActionRegistry` enforces unique `(key, modifiers)` combinations.

- `Register(...)` parses the key string and returns an error on invalid bindings or conflicts
- `MustRegister(...)` is the panic-on-error convenience path
- `RegisterAction(...)` and `MustRegisterAction(...)` allow prebuilt actions (for example menu actions)

Conflict scope is registry-local:

- conflicts are checked only within the same registry
- same keybindings may exist in different registries and are resolved by cascade order

## traversal interfaces and component participation

### `ChildActionProvider`

```go
type ChildActionProvider interface {
    ChildActions() []Component
}
```

Use this for components that contain or route to other components.

### `LocalActionProvider`

```go
type LocalActionProvider interface {
    LocalActions() *ActionRegistry
}
```

Use this to explicitly expose the component's local action registry.

If `LocalActionProvider` is not implemented, traversal falls back to `Actions()`.

## framework conventions

- `Container`:
  - `ChildActions()` returns `Children`
  - `LocalActions()` returns its local registry
- `Func`:
  - `LocalActions()` returns its local registry
- composite components such as `Workspace`, `DashManager`, `Dash`, and `HCollapse`:
  - provide active child routing via `ChildActions()`
  - keep local actions in their own registry

## menu actions integration

`NewMenuAction(label, keys, handler)` creates one action that works in both paths:

- keyboard shortcut path via registry registration
- menu click path via `action.DrawMenuItem()`

Behavior notes:

- invalid `keys` passed to `NewMenuAction` panic during construction
- shortcut labels shown in menus are generated from parsed modifier/key values

## recommended patterns for custom components

When implementing a custom composite:

1. embed `Container` unless there is a strong reason not to
2. implement `ChildActions()` to expose only active or relevant children
3. put local shortcuts in `Container.Actions()` (or your local registry)
4. do not replace local action access with delegated child `Actions()` return values

## common pitfalls

- expecting non-exact modifier matches
- registering the same key combo twice in one registry
- forgetting to expose active child components in `ChildActions()`
- assuming conflict checks happen across the entire app (they do not)

## minimal mental model

Use this as the shortest reliable model:

1. actions are grouped by registry
2. registries are traversed child-first, then parent, then app-global
3. first matching action wins
4. key conflicts are prevented per registry at registration time
