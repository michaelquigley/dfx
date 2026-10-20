# Complete Styles — Work Order

Implement [complete-styles.md](complete-styles.md) in three stages. This is planning for a change in default presentation, not authorization to bypass the stage 2 hands-on acceptance before changing the default. Each implementation stage goes through Terminus to no blocking findings, then Michael's review. The operator commits.

## Grounding

| Surface | Current implementation | Work required |
|---|---|---|
| Startup | `app.go`: `setupFontsAndTheme` loads fonts, calls `DefaultStyle`, then applies `Config.Theme` or Modern Dark before `OnSetup` | Resolve complete style and preserve the color-theme override and disable-flag order |
| Palette | `theme.go`: `Theme.Apply/Name`, hue schemes, Modern Dark, `SetTheme` | Retain the interface; add Faceplate colors and role bindings for custom chrome |
| Metrics | `style.go`: 4 px padding/spacing plus fixed borders, radii, scrollbar | Add copyable complete style data and apply its metrics |
| Type | `fonts.go`, `fonts/embed.go`: three public slots, Gidole 20/16, JetBrains Mono 16, merged icons | Parameterize setup; add house face and real bold without moving existing slots |
| Root | `app.go`: native-sized root and `State.Size = windowSize` despite window padding | Keep native root geometry; hand off content allocation |
| Composition | `workspace.go`, `dash.go`, `dashManager.go`, `hCollapse.go`, `multiGrid.go` | Account for changed padding, headers, origins, and internal spacing |
| Custom chrome | `faderGeometry.go`, `controls.go`, `toolbar.go`, `nodeCanvasStyle.go` | Apply meaningful roles while preserving Classic and explicit overrides |
| Information colors | `vuMeter.go`, `vuWaterfall.go`, `logViewer.go` | Retain current explicit fields/global overrides and signal meanings |
| Demonstration | `examples/dfx_example_themes/main.go` | Replace the limited palette-only demonstration with a representative comparison |

Read the built contracts in `docs/current/layout-guide.md`, `faders.md`, `hcollapse.md`, `icon-fonts.md`, and `node-canvas.md`. Constructors run before startup; NodeCanvas's implicit style derives at its first `Begin`, and its current theme-refresh mechanism is explicit `SetStyle`.

## API and implementation shape

### Complete configuration

Add `Style` as a complete configuration value in `style.go` or a small additional camelCase file. It contains a display name, default `Theme`, typography specification, ImGui metrics, and additional dfx chrome/spacing defaults. Use named nested data types where they reduce ambiguity. Zero numeric values are real values in a supplied complete style; customization starts from a built-in constructor, not a sparse merge.

Expose `ClassicStyle() Style` and `FaceplateStyle() Style`, and add `Style *Style` to `Config`. Expose the accent in the Faceplate color-theme configuration so callers can change it before startup. Constructors are context-free, return independent configuration, and do not mutate shared preset pointers. `FontFace.Data` is borrowed, immutable asset data: share those bytes across presets and configuration copies rather than cloning them. Customization replaces a face's slice before startup; callers must finish preparing any caller-owned bytes before supplying them as `Data`, then keep them immutable while referenced by a style or atlas. Clone other mutable data owned by a preset; custom `Theme` objects supplied by the caller retain their ordinary interface ownership.

Expose a read-only-by-copy `CurrentStyle() Style` view of the active style for application spacing and dfx helpers. Configuration fields copy by value, with font bytes remaining shared and immutable. Copy a built-in `FaceplateTheme` at startup after effective-theme resolution, whether it came from `Config.Style.Theme` or the `Config.Theme` override, and copy it again when returning `CurrentStyle()`. Editing the source theme or any returned snapshot must not change the active theme configuration or another snapshot. Leave application-supplied custom `Theme` implementations under their existing ownership rules; no general cloning interface is required. Keep the active configuration separate from mutable ImGui values: live colors and metric overrides come from the current ImGui context when drawing. The active view describes selected defaults. Outside a dfx setup scope, helpers fall back to Classic metadata so existing raw-context tests and direct helper use keep working.

Use the existing single-active-UI model, as with `Fonts`: install active metadata during setup and restore prior metadata after `App.Run` exits on every path, including panic unwinding. Do not store pointers to destroyed ImGui objects. Headless setup fixtures must restore metadata and font globals they replace. No context registry, global application service, or new concurrency promise is required.

Keep `DefaultStyle()` as the existing Classic metric helper and `SetupFonts()` as the existing Classic font helper. Add an internal parameterized font setup path for App startup. Preserve `MainFont = 0`, `MonospaceFont = 1`, `SmallFont = 2` and Classic's three-entry setup. Faceplate appends its bold face internally; expose a balanced `PushHeadingFont`/`PopFont` usage that selects that face for Faceplate and MainFont for Classic. If dfx fonts are disabled, the heading helper pushes the current ImGui font instead of indexing unloaded dfx slots. Update the size lookup with the selected typography. Explicit direct calls to old helpers continue to do what their callers requested.

### Resolution and lifecycle

| Input | Result |
|---|---|
| `Config.Style == nil` | Classic in stages 1–2; Faceplate in stage 3 |
| Non-nil `Config.Style` | Copy and use that complete value |
| Non-nil `Config.Theme` | Override selected style's default color theme |
| `DisableFonts` | Do not clear/load the atlas or replace font globals |
| `DisableTheming` | Do not apply a color theme; still apply selected metrics |
| `OnSetup` | Runs after selection, fonts, metrics, and effective colors |
| `SetTheme` after startup | Apply colors only; no font/metric reset or component recreation |

A supplied style whose default theme is nil uses Modern Dark as the existing fallback. dfx does not provide an exported full-style apply/switch API in this version. Applications select another complete style on a subsequent launch.

Both bundled themes should establish all ImGui color slots against a known baseline before applying their palette so their result is independent of a previously applied theme. Preserve the existing Modern Dark result on a fresh context and avoid changing arbitrary custom `Theme.Apply` semantics. Characterize previously unspecified slots against the pinned cimgui-go defaults before completing Modern Dark's table; do not inadvertently recolor them.

Semantic roles for accent, panel, and secondary text resolve to current ImGui color slots. For Faceplate use `ColCheckMark`, `ColChildBg`, and `ColTextDisabled` respectively; make those bindings explicit in the selected configuration rather than testing preset names in draw code. Classic's fader fill/hover-border keeps `ColButtonHovered` and its active border keeps `ColButtonActive`. Faceplate binds both to its accent slot. Other fader colors continue to use their existing ImGui slots. Application `AccentColor` and `TrackColor` still win and style alpha is applied exactly once. Use direct color access in hot paths instead of repeatedly copying the whole color table.

Give Faceplate toggle-on controls a dark foreground on accent fill; Classic retains its current text treatment. Add a small optional-parameter Toggle variant for explicit foreground override while retaining the existing Toggle signature and behavior under Classic. Resolve explicit foreground before the preset default, preserving disabled alpha and balanced style pushes. Use a documented contrast choice for the default accent; the gallery must also audition a second hue. This is not a general-purpose contrast engine.

### Reference assets and initial values

Use `../faceplate/src/faceplate/faceplate.h`, `fonts.h`, and `assets/fonts/README.md` as reference. The inspected faceplate revision is `c65b59e8319b8f030b1bbc3ef01a688779dcb677`; verify the revision and asset checksums when copying. Bundle Regular/Bold plus OFL and provenance in dfx's `fonts/`, and embed them through `fonts/embed.go`. The implementation must build with no sibling faceplate checkout. Do not add JUCE or a shared code submodule to import palette constants.

| Faceplate starting value | Value |
|---|---|
| Background / panel / track-border | `#1a1c20` / `#24272e` / `#3a3f48` |
| Primary / secondary text | `#d6d6d6` / `#878b93` |
| Default accent | `#ffd800` |
| Outer content margin / panel padding | 12 / 12 px |
| Item gap / section gap | 8 / 16 px |
| Frame padding | 8 × 5 px |
| Main / small / mono / bold size | 20 / 16 / 16 / 20 px |

Begin with Classic radii, scrollbar width, and interaction tolerances unless audition demonstrates a presentation problem. Set cell padding explicitly for Faceplate tables; leave application row-height choices intact. House sizes are adjustable at stage 2 acceptance. Preserve Classic's original configured metrics and original three font families/sizes exactly.

Merge legacy Material Icons and the supplemental non-overlapping Material Symbols into each UI face offered for labels, including bold if offered there. Preserve icon cmap precedence and tune offsets per face without changing the existing Classic offsets. Keep `PushFont` sizes consistent with loaded fonts and NodeCanvas's detent scaling.

## Stage 1 — Establish Classic and the complete-style path

1. Add the configuration and startup resolution described above with Classic as the implicit default. Move existing metric/font setup behind the parameterized path while retaining the legacy public helpers.
2. Add role bindings and route the fader's default color decisions through them. The Classic bindings must reproduce every existing idle, hovered, active, disabled, and explicit-override case. Do not change its hit rectangle, taper, scale, or handle geometry.
3. Capture meaningful baseline assertions from fresh contexts before refactoring: existing applied metrics, font slots and sizes, representative theme slots, and widget bindings. Test that customizing one returned preset leaves a second preset unchanged.
4. Exercise setup with a custom spy theme and all disable-flag combinations. Prove the effective theme wins after style defaults, `OnSetup` observes final setup, and a subsequent App setup cannot inherit the previous style metadata. Test cleanup through an abnormal exit using the existing test seam where practical.

Acceptance: the new API selects Classic explicitly; implicit startup still uses Classic; existing tests and example packages pass; no visual redesign or default switch is included. Update built documentation for the new opt-in path. Run `make test`, compile all example packages with an output directory under `/tmp` or with `go build ./...`, and run Terminus before Michael's review.

## Stage 2 — Implement Faceplate and audition the whole surface

### Palette, type, and custom chrome

Implement the Faceplate constructor and theme; load bundled Chakra Petch and real Bold through the shared setup path. Set neutral hover colors independently of the accent, readable selection/menu states, and table colors. Keep a distinct visible keyboard-focus treatment. Confirm toggle-on foreground, fader fill/borders, toolbar background, and collapsed-panel indicators agree with those roles.

Keep NodeCanvas's explicit style lifecycle. Adapt `DefaultNodeCanvasStyle` only where new roles are needed, preserving the Classic derivation and all complete-value override semantics. The gallery's live theme selector must re-derive only its own implicit canvas style through `SetStyle`; its separate explicitly customized canvas example must remain unchanged. Do not add per-node color-table copies or trigger measurement callbacks simply to refresh palette.

Retain VU signal colors and log global color overrides. Include their actual widgets in the comparison so an unreadable state is visible during acceptance. Where no private presentation constant needs changing, inheriting fonts and ImGui colors is sufficient; avoid rewriting widget model code.

### Allocation pass

Perform these changes in stage 2, after stage 1 preservation is verified. Shared allocation corrections apply to both presets and are documented separately from their visual defaults.

- **Root (`app.go`).** Keep `rootWindowRect`'s native-window/menu geometry and square outer corners. Push the selected outer padding for the root only, restore panel padding for application drawing, and compute root `State.Size` from the actual usable content region immediately after `Begin`. Do not subtract outer margin again inside the child. Preserve `Position`'s documented coordinate meaning.
- **Workspace (`workspace.go`).** Replace the approximate 30 px selector reservation with the actual cursor advance, including intentional spacing. Bound the remaining height; hide-selector and empty cases must remain correct.
- **Dash / DashManager.** Anchor bounds relative to the component's content origin instead of the enclosing window's outer edge. Resolve internal padding/header/radius defaults from the selected style, account for the header once, and pass actual content allocation to the embedded component. Retain `DashManager.Margin`, `TopMargin`, user dash size/min/max, precedence, and drag behavior as explicit fields. `TopMargin` reserves space above the entire manager layout: offset the layout origin by that inset, subtract it once from the available height, and derive every panel and the inner surface within the remaining area under either precedence mode. Exhausted space follows the existing skip-drawing and animation rules. The gallery assigns the selected style's gap to the manager in `OnSetup`; do not silently reinterpret its historical public margin value.
- **HCollapse.** Measure the header against actual text/control height and padding; keep its anchoring, animation, resize clamp order, and explicit dimensions intact. Do not leave a toggle outside its collapsed strip when Faceplate padding increases: collapsed chrome uses the available strip and measured icon/control height, not full panel padding on both sides. Pass content-region size after the header. Preserve existing public constants for source compatibility even when internal default calculations become style-aware.
- **FlexLayout.** Replace private presentation margins/gaps with selected defaults, retain splitter hit strip sizes, and recompute allocated remainder using the same values used to draw. In child windows, deliberately apply panel padding (including borderless child flags when needed), then pass usable content size.
- **GridLayout.** Anchor to the incoming content cursor, preserve explicit cell-size/span behavior, apply inter-cell gaps within each allocated cell, and give padded children their usable size. Preserve Classic's old gap treatment. Clip cells against the allocation; do not let a negative clamped size become ImGui's relative-fill request.
- **Toolbar.** Replace its private 3 px inset with a selected presentation default and keep centering based on measured text/frame height. Ensure the callback and its final item register the full bar extent without an extra unexplained gap.

For all these paths, a visible panel has one owner for padding at each boundary. Keep raw allocated rectangles separate from content rectangles. Skip content drawing for exhausted allocations; do not pass a computed zero size to `BeginChild` expecting an empty rectangle, because ImGui interprets zero as fill-remaining. Existing public dimensions and per-widget overrides win; any further ambiguous public default stays unchanged and is documented in the style guide rather than inferred from numeric equality. Do not add arbitrary section gaps after every widget: expose `SectionGap` as a value applications use intentionally.

### Comparison application

Extend `examples/dfx_example_themes` with startup flags `-style classic|faceplate` and an optional Faceplate accent value, keeping launch instructions in the example and the current style guide. Validate flag input before window creation. Run separate launches for complete-style comparison; the live selector remains color-only and initially displays the theme actually applied.

Use stable sample data and include menus/popups, buttons/toggles/checkboxes, inputs and sliders, tabs, ordinary and dense tables, FileTree, LogViewer with all severity levels, VUMeter/VuWaterfall, Fader, Toolbar, Workspace, DashManager, HCollapse on both anchors, both grid layouts, and NodeCanvas with mini-map. Organize views so the example is usable rather than one oversized page. Include a deliberate local spacing/color override and explicit canvas style, plus icons beside Regular and Bold text. Do not require a user's real filesystem, log stream, or audio devices.

### Verification and acceptance

Headless tests cover actual root/child allocations with both presets, menu/no-menu, nested nonzero origins, selector on/off, tiny windows, collapsed/expanded panels, and explicit public dimension overrides. Use real ImGui frames for sizing paths rather than only testing arithmetic helpers. Extend existing fader, HCollapse, canvas, and font tests where their contract is touched. Glyph checks must use no-fallback lookup for representative existing and supplemental icons at both UI sizes and existing canvas detents. Assert color themes can change while frame padding and loaded fonts stay unchanged. Verify an explicit canvas style and explicit fader colors survive those changes.

Explicitly check toggle foreground/background pairs at rest, hovered, and pressed, including live switches from Faceplate colors to the legacy hue themes and an explicit foreground override. Exercise these combinations in headless state/color checks and the hands-on audition: `ColButtonHovered` and `ColButtonActive` can replace the accent fill while the toggle's foreground remains in effect.

Run `make test`, build examples, and Terminus. Present the executable comparison instructions to Michael. His hands-on pass checks Classic familiarity, everyday Chakra Petch readability, separation of groups, comfortable controls, usable dense views, icon alignment, hover/active/disabled/focus/selection states, popup and menu readability, panel resizing/collapse, and NodeCanvas at every detent. Check a normal tool window and a narrow/small one, plus default yellow and a second accent. Record concrete acceptance or remaining adjustments; do not infer it from test success. Keep Classic as the default until acceptance arrives.

## Stage 3 — Change the default and finish migration

After Michael accepts stage 2's typography, spacing, and interactions, change nil `Config.Style` resolution to Faceplate. Retain explicit Classic and all color themes. Update the style comparison's startup default and any examples that explicitly request Modern Dark while describing it as the complete default.

Run the examples that exercise composition, mixer, logs, and canvas; fix layout assumptions exposed by the new defaults without removing intentional local styling. Source-build the full examples set. Update `Config` comments, README entry points, and `docs/current/styles.md` with complete-style selection, customized accent/spacing, Classic migration, color-only `Theme`/`SetTheme`, setup order, disabled flags, legacy helper meanings, and explicit NodeCanvas refresh. Update the layout guide for usable `State.Size` and content origins, the fader guide for role bindings, and icon-font docs for the selected face and bold slot.

Add the in-house changelog entry with a leading bold clause announcing the changed default font, spacing, and colors. Mention shared allocation corrections and the explicit Classic escape hatch. Test both nil/default selection and explicit Classic after the flip, then run the ordinary build/test gates and Terminus. Ask for Michael's final hands-on review of the actual default path.

When all stages are accepted, synthesize built behavior into `docs/current/`, carry genuinely live deferrals into a smaller future note, and retire this pair. Check roadmap references before removing intent documents. Do not commit or push.

## Review and session discipline

Mercurius reviews this pair against the code before implementation. Review findings are presented one at a time for Michael's decision; apply only the agreed disposition and pause before advancing. Keep style architecture decisions in the spec and implementation details in this work order. Changes to cosmetic starting values belong in the comparison/implementation and current documentation once accepted, rather than spawning extra compatibility machinery.
