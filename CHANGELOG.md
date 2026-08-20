# CHANGELOG

## Unreleased

## v0.1.6

FIX: **`App.SetWindowSize` and `App.SetWindowPos` now take effect at the next frame boundary rather than at the moment they are called.** Applying window geometry from inside a frame crashes on Wayland: the GLFW call dispatches a surface configure synchronously, which re-enters the render loop through GLFW's window-refresh callback and opens a second ImGui frame inside the current one, tripping ImGui's "Forgot to call Render() or EndFrame()" assertion. X11 issues the same call asynchronously to the window manager, which is why the hazard stayed invisible there. Both setters now record the request and flush it from the backend's `afterRender` hook, outside any frame; a later request before that boundary supersedes an earlier one, and a boundary with nothing pending does not reassert geometry. Applications that called either setter and immediately read the geometry back will now observe the old value until the frame ends.

## v0.1.5

FEATURE: `App.SetWindowPos(x, y int)` and `App.SetWindowSize(w, h int)` expose runtime window geometry control, delegating to the backend behind the same nil guard as `SetWindowTitle`. Previously geometry was fixed once at `Run()` from `Config`; an application can now move and resize the window live (for example, transitioning from a compact launcher surface to a larger main surface). Both must be called on the UI goroutine, since the underlying GLFW operations are main-thread only.

## v0.1.4

FEATURE: `Config.OnAction func(ActionEvent)` reports every action as it fires, so an application can record which actions and shortcuts actually get used and refine its keybindings from real usage rather than guesswork. Keyboard invocations route through a single dispatch chokepoint that notifies the hook before running the handler; each `ActionEvent` carries the invoked `*Action`, its `Source`, and the `Time`. Menu-click capture is not yet wired — `ActionSourceMenu` is reserved for it.

CHANGE: Reference documentation moved under `docs/current/` (`layout-guide.md` and `child-actions.md`, formerly `docs/LAYOUT_GUIDE.md` and `docs/CHILD_ACTIONS.md`), and a repo-local `AGENTS.md` now carries contributor and agent orientation. README links updated to match.

## v0.1.3

Initial release.