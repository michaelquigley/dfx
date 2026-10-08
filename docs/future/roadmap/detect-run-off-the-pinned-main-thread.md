---
title: detect Run off the pinned main thread
state: researching
created: 2026-10-07
tags: [defect]
milestone: v0.1.x
source: patchbay stage 3; baab 2026-08-20
---

make `App.Run` fail loudly, or at least log at warn, when it is called on a goroutine that is not locked to the main OS thread. two projects have now lost their wayland window decorations to the same cause: glfw and libdecor's gtk plugin initialize off the main thread and `gtk_init_check` fails inside the plugin, intermittently, depending on where the scheduler left the goroutine. baab hit it when `Run` ran on a spawned goroutine; patchbay hit it because `cmd/patchbay` had no `runtime.LockOSThread()` in `init()` at all. the check is cheap (`runtime.LockOSThread` is idempotent, but the condition to detect is "not already locked", which needs a sentinel: lock in an `init()` of dfx itself and compare thread identity, or document the requirement in `New`/`Run` and have `Run` call `LockOSThread` and warn that doing so late is not sufficient). document the requirement in AGENTS.md either way.

## why

the failure is silent, intermittent, and looks environmental (it points at libdecor and the desktop), so each project spends a diagnosis on it. the library is the one place that knows the requirement.
