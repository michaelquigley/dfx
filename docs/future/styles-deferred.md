# Styles — Deferred Work

Complete styles are built and accepted; see the [style guide](../current/styles.md). This note preserves the remaining questions from that work. None blocks the release.

## Layout follow-ups

**Side dash widths.** Left and right dashes are not clamped to the manager's available width. The exhausted-allocation and `TopMargin` fixes bounded height; a horizontal fix still needs its own sizing and precedence checks under both styles.

**HCollapse title fit.** Titles appear once the panel exceeds `MinWidth+50`, rather than when the measured title fits. The Waterfall (2ch) example exposed clipping. Revisit the rule under both styles; changing it also changes Classic.

## Naming question

Michael reads “theme” as fonts and colors together, so the gallery's color-only menu became “Colors.” Renaming `FaceplateTheme` to `FaceplateColors`, and renaming the public color-only `Theme` API, remain undecided. Consider compatibility before proposing either change; neither rename is approved implementation work.

## Features waiting for a client

**Live complete-style switching.** Startup selection covers current needs. Revisit when a client needs font-atlas rebuilding and layout refresh during a running application.

**Style editors, serialization, registries, inheritance, and automatic density modes.** Copyable configuration and local overrides cover current use. Add machinery only for a demonstrated application need.

**Rotary widgets and plugin-specific controls.** Sharing Faceplate's visual language did not require a larger widget inventory. Revisit for a tool that needs those controls.

**Shared token generation and automatic Faceplate synchronization.** dfx owns one Go/ImGui realization with pinned asset provenance. Revisit extraction when another renderer needs shared machine-readable tokens; metrics still require deliberate audition.
