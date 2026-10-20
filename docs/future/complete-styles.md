# Complete Styles for dfx

dfx's default face should belong to the same family as faceplate and michael9: Chakra Petch for everyday interface text, charcoal surfaces, a single configurable identity accent, and enough space for controls and groups to read clearly. Applications should obtain that face by selecting one complete style. The existing compact appearance remains available as Classic.

This is a change to the framework's default presentation and the infrastructure that carries it. It covers ordinary ImGui controls, dfx's own chrome, and the allocation of space around content. The implementation plan is [complete-styles-work-order.md](complete-styles-work-order.md).

## Decisions and starting positions

Michael accepted Chakra Petch as the everyday face, more margin and spacing, preservation of the existing appearance as a pluggable style, and a three-stage sequence: establish Classic, audition Faceplate, then change the default after hands-on acceptance.

The architecture and compatibility details below are the proposed translation of those decisions. Typography sizes, padding, and interaction colors are starting positions for the running comparison, rather than claims of visual acceptance.

## The visual language

Faceplate provides the identity reference: cool charcoal background, a distinguishable raised panel, soft primary text, quieter secondary text, and one accent. The accent identifies active values, selection, focus, and enabled state. Hover should be legible without turning every surface into a solid accent rectangle. Filled accent controls need foreground text that remains readable against that accent.

Chakra Petch Regular becomes the everyday UI face. Real Bold is available for section headings and emphasis. JetBrains Mono remains available for code, logs, and numeric material where alignment matters. Icons retain the existing artwork and public constants. Font sizes are judged on dfx's real controls and reading surfaces; the plugin role sizes do not transfer automatically between renderers.

The spacing vocabulary distinguishes the application's outer margin, padding inside panels, gaps between related items, and separation between sections. Framed controls also receive comfortable internal padding. A log or dense table may deliberately use tighter local spacing while its containing panel follows the normal rhythm. dfx cannot infer semantic sections from arbitrary application draw calls; applications use the exposed section spacing deliberately.

The new face applies to desktop tools with trees, tables, graphs, and logs. The plugin preference for knob-first layouts does not constrain what a dfx application can build. Meter zones, errors, warnings, and application-declared graph colors keep their informational meaning independently of the identity accent.

## Complete styles and color themes

A style is a complete, copyable configuration for typography, ImGui presentation metrics and colors, and the additional defaults consumed by dfx components. Built-in constructors return independent configuration fields that callers can customize before application startup. Font bytes are the explicit exception: they are borrowed, shared, immutable assets, never cloned when constructing a preset, selecting a style, or querying its configuration. Callers customize a face by assigning another immutable font asset before startup; they must not edit bytes through a style copy. Classic and Faceplate exercise the same public configuration path; applications can supply a customized value without registering a plugin or subclassing a renderer.

The mutable built-in Faceplate theme belongs to the configuration snapshot: startup selection and `CurrentStyle()` copy it so editing the source or a returned snapshot cannot change the active configuration. Application-supplied custom `Theme` implementations retain their existing ownership rules; this does not require a general theme-cloning interface.

The existing `Theme` interface remains the color-only extension point. `Config.Style` selects the complete style. `Config.Theme`, when supplied, overrides that style's default color theme. Existing `SetTheme` calls continue to change colors in the current ImGui context without rebuilding fonts or changing spacing. `ModernTheme`, `ModernDark`, hue schemes, `DefaultStyle`, `SetupFonts`, and the existing font slot constants retain their existing explicit meanings. Selecting Modern Dark colors alone does not select Classic's typography and layout; migration documentation shows the complete Classic selection.

Full style selection is a startup operation in this version. The comparison application selects a style before opening its window; it may offer live color-theme changes within that style. No complete-style switching API is promised inside an active frame.

The startup order is complete-style selection, font setup, metric application, effective color-theme application, then `OnSetup`. Application setup and scoped drawing overrides therefore have the final word. `DisableFonts` continues to skip font loading. `DisableTheming` continues to skip color-theme application while dfx metrics are applied; its current behavior is preserved and documented rather than silently broadening the flag.

## How defaults reach components

Ordinary controls use the effective ImGui style directly. dfx-specific color roles give custom chrome a meaningful source for accent, panel, and secondary text. Their default values resolve through the effective ImGui color table so legacy themes and local ImGui color overrides remain useful. Classic retains the existing per-widget color relationships, including the fader's current hover and active colors. Faceplate can use its identity accent independently of its neutral button-hover treatment.

Faders follow effective colors on each draw and retain their explicit accent and track overrides. NodeCanvas retains its existing complete-value `NodeCanvasStyle` contract: its implicit style is derived on first `Begin`, and an application that changes color theme after that explicitly re-derives and assigns a style with `SetStyle`. A supplied canvas style, node accent, or link color remains authoritative. This work does not add automatic canvas invalidation or change gesture tolerances as a side effect of a theme change.

Meters retain their explicit color fields and signal palettes. Log severity colors and the exported log-color overrides remain available; neither is recolored wholesale into the identity accent. The surrounding typography, panels, and layout provide the family resemblance.

Components are often constructed before an ImGui context exists. Constructors must not read the current context to acquire style defaults. Internal presentation constants can become late-resolved style defaults. Existing public dimensions, gaps, margins, colors, and complete component styles remain authoritative, even where constructors initialized those fields to historical defaults. The implementation must not guess whether an application assigned a value by comparing it with an old default. The comparison example explicitly demonstrates adopting style spacing through those older public fields.

## Space and allocation

The root window continues to fill the native client area below the main menu. The outer margin is an inset of application content, not a second window rectangle. A child component receives the usable area at its drawing origin, after its parent's padding and headers have been accounted for. Nested containers must neither subtract the same padding twice nor bypass it with cursor coordinates anchored to the outer window.

The larger defaults require a targeted allocation pass through the root, Workspace, Dash, HCollapse, and both MultiGrid layouts. Header reservation follows measured controls and active typography. Resize strips and canvas hit tolerances retain their interaction contracts. Tiny available areas produce bounded, nonnegative allocations rather than interpreting accidental negative dimensions as ImGui's fill-relative size syntax.

Classic preserves the existing font choices, metric values, color themes, and widget treatments. Correcting an allocation that previously exceeded its content area can change how an existing layout fits; this is an explicit compatibility correction, documented alongside the default change. Classic is the route to the old styling, rather than a guarantee that every historical sizing defect is retained.

## Source ownership

faceplate remains the reference for the family identity and font provenance. dfx owns the Go/ImGui realization: a documented snapshot of the palette and bundled font assets, with source revision, checksums, and the font license. There is no JUCE dependency, sibling checkout requirement, or build-time fetch. Later changes in faceplate are adopted deliberately and auditioned in dfx.

## Scenarios and acceptance

**A new tool.** With neither style nor theme specified, the completed work opens in Faceplate. Chakra Petch, panels, control padding, margins, and dfx chrome agree without application setup calls.

**An existing tool.** Selecting Classic restores the old typography and styling defaults. Existing hue themes still work, custom `Theme` implementations still compile, and application-defined widget colors and dimensions survive. Allocation corrections are called out rather than attributed to a font change.

**An application identity.** Two Faceplate instances with different accents retain the same neutral chrome and type. Accent text stays readable; warning/error and meter colors retain their meanings. An application can customize spacing from a copy without modifying another application's preset.

**A working surface.** The comparison contains ordinary and dense sections, nested containers, menu and popup states, a real FileTree and LogViewer, meters, faders, and NodeCanvas. At a normal and a narrow window size, controls remain operable, headers fit, and fills stay inside their allocation. Icons align beside the new text. The canvas is checked at all existing zoom detents, including its overview chrome.

Headless checks establish precedence, default preservation, glyph availability, and allocation behavior. Michael judges typography, spacing, state readability, and practical working density in the running comparison before Faceplate becomes the default. Successful tests or review do not stand in for that acceptance.

## Seam census

- **Style / application layout — separate.** The style supplies default presentation and a spacing vocabulary; explicit application geometry and semantic grouping remain application decisions. Revisit automatic density behavior only when a real client needs it.
- **Complete style / color theme — separate.** Fonts and layout are selected at startup; the existing palette interface remains usable at runtime. Revisit full live switching when a client requires it.
- **Default / explicit widget configuration — separate.** Defaults are late-resolved where possible, and explicit values win. Retain NodeCanvas's complete-style ownership; no automatic overwrite of application styles.
- **Family identity / renderer implementation — separate.** Adopt faceplate's assets and intent with recorded provenance; own the ImGui translation in dfx. A second renderer-independent consumer of shared machine-readable tokens would justify reconsidering extraction.
- **Context / process — retain the existing single-active-UI execution model.** Do not expand this work into concurrent App/context support. New active-style state must be scoped to setup and teardown and must not leak into a later App or a headless test.

## Deferred (and Why)

**Complete live style and font switching.** Startup selection provides the comparison and application configuration needed here without atlas rebuilding and layout-cache invalidation during drawing.

**Editors, serialization, registries, inheritance, and automatic density modes.** Copyable configuration and local overrides cover the current need; additional machinery waits for an application that needs it.

**New rotary widgets and plugin-specific control vocabulary.** Family styling does not require expanding the widget inventory. Existing controls establish the first usable surface.

**Cross-repository token generation and automatic faceplate synchronization.** There is one Go realization today, and renderer-specific metrics need deliberate audition. A pinned documented snapshot is sufficient.

**General layout redesign and large-graph optimization.** Fix allocations touched by the spacing change, preserving existing interaction behavior. New layout engines, virtualization, and NodeCanvas performance targets remain separate work.
