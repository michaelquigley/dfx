package dfx

import (
	"math"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/michaelquigley/dfx/fonts"
)

// HCollapse is a horizontal collapsible component that contains content to its right.
// when collapsed, only the toggle button is visible. when expanded, shows a header
// bar with title and the content below.
type HCollapse struct {
	Container
	Title         string              // displayed in header when expanded (also used for imgui ID)
	Expanded      bool                // current state
	ExpandedWidth float32             // width when fully expanded
	CurrentWidth  float32             // animated width (internal)
	MinWidth      float32             // collapsed width (toggle button only)
	MaxWidth      float32             // maximum width when resizing (0 = no limit)
	TransitionMs  int                 // animation duration
	Resizable     bool                // allow drag-to-resize when expanded
	Content       Component           // the component to show/hide
	OnToggle      func(expanded bool) // optional callback on state change
}

// HCollapseConfig provides configuration options for NewHCollapse.
type HCollapseConfig struct {
	Title         string
	ExpandedWidth float32
	MinWidth      float32 // defaults to HCollapseDefaultMinWidth
	MaxWidth      float32 // 0 = no limit
	TransitionMs  int     // defaults to HCollapseDefaultTransition
	Resizable     bool
	Expanded      bool // initial state
}

// HCollapse constants
const (
	HCollapseHeaderHeight      = 24
	HCollapseDefaultMinWidth   = 36
	HCollapseDefaultTransition = 80
	HCollapseResizeHandleSize  = 20
)

// NewHCollapse creates a new horizontal collapsible component.
func NewHCollapse(content Component, cfg HCollapseConfig) *HCollapse {
	minWidth := cfg.MinWidth
	if minWidth <= 0 {
		minWidth = HCollapseDefaultMinWidth
	}
	transitionMs := cfg.TransitionMs
	if transitionMs <= 0 {
		transitionMs = HCollapseDefaultTransition
	}
	expandedWidth := cfg.ExpandedWidth
	if expandedWidth < minWidth {
		expandedWidth = minWidth
	}

	currentWidth := minWidth
	if cfg.Expanded {
		currentWidth = expandedWidth
	}

	return &HCollapse{
		Container: Container{
			Visible: true,
		},
		Title:         cfg.Title,
		Expanded:      cfg.Expanded,
		ExpandedWidth: expandedWidth,
		CurrentWidth:  currentWidth,
		MinWidth:      minWidth,
		MaxWidth:      cfg.MaxWidth,
		TransitionMs:  transitionMs,
		Resizable:     cfg.Resizable,
		Content:       content,
	}
}

// id returns the unique imgui identifier for this instance.
func (h *HCollapse) imguiID() string {
	return "##hcollapse_" + h.Title
}

// Toggle toggles the expanded state.
func (h *HCollapse) Toggle() {
	h.Expanded = !h.Expanded
	if h.OnToggle != nil {
		h.OnToggle(h.Expanded)
	}
}

// Draw implements Component.
func (h *HCollapse) Draw(state *State) {
	if !h.Visible {
		return
	}

	// animate toward target width
	h.animate()

	// when collapsed or collapsing, just draw the toggle button without any child windows
	// this avoids scrollbar issues when the panel is narrow
	if !h.isFullyExpanded() {
		h.drawCollapsedToggle(state)
		return
	}

	// draw the full expanded panel with child window
	imgui.SetNextWindowBgAlpha(DashBackgroundAlpha)
	imgui.PushStyleVarFloat(imgui.StyleVarWindowRounding, DashWindowRounding)

	windowFlags := imgui.WindowFlagsNoCollapse | imgui.WindowFlagsNoTitleBar | imgui.WindowFlagsNoResize | imgui.WindowFlagsNoScrollbar | imgui.WindowFlagsNoScrollWithMouse

	childSize := imgui.Vec2{X: h.CurrentWidth, Y: state.Size.Y}
	imgui.BeginChildStrV(h.imguiID(), childSize, imgui.ChildFlagsNone, windowFlags)

	// draw header bar
	h.drawHeader()

	// draw content
	h.drawContent(state)

	// draw resize handle if applicable
	if h.Resizable {
		h.drawResizeHandle(state)
	}

	imgui.EndChild()
	imgui.PopStyleVar()
}

// drawCollapsedToggle draws just the toggle button when collapsed (no child windows).
func (h *HCollapse) drawCollapsedToggle(state *State) {
	// draw a simple background for the collapsed state
	imgui.SetNextWindowBgAlpha(DashBackgroundAlpha)
	imgui.PushStyleVarFloat(imgui.StyleVarWindowRounding, DashWindowRounding)

	// use a minimal child just for the background, with no scrollbars
	windowFlags := imgui.WindowFlagsNoCollapse | imgui.WindowFlagsNoTitleBar | imgui.WindowFlagsNoResize | imgui.WindowFlagsNoScrollbar | imgui.WindowFlagsNoScrollWithMouse

	childSize := imgui.Vec2{X: h.CurrentWidth, Y: state.Size.Y}
	imgui.BeginChildStrV(h.imguiID(), childSize, imgui.ChildFlagsNone, windowFlags)

	windowPadding := imgui.CurrentStyle().WindowPadding()
	imgui.SetCursorPos(windowPadding)

	// toggle button only
	imgui.PushStyleColorVec4(imgui.ColButton, imgui.Vec4{})
	imgui.PushStyleColorVec4(imgui.ColButtonHovered, imgui.CurrentStyle().Colors()[imgui.ColHeaderHovered])
	imgui.PushStyleColorVec4(imgui.ColButtonActive, imgui.CurrentStyle().Colors()[imgui.ColHeaderActive])

	if imgui.Button(fonts.ICON_CHEVRON_RIGHT + h.imguiID() + "_toggle") {
		h.Toggle()
	}
	if imgui.IsItemHovered() && h.Title != "" {
		imgui.SetTooltip(h.Title)
	}

	imgui.PopStyleColorV(3)

	imgui.EndChild()
	imgui.PopStyleVar()
}

// drawHeader draws the header bar with toggle button and title.
func (h *HCollapse) drawHeader() {
	windowPadding := imgui.CurrentStyle().WindowPadding()
	imgui.SetCursorPos(windowPadding)

	// toggle button
	icon := fonts.ICON_CHEVRON_RIGHT
	if h.Expanded {
		icon = fonts.ICON_CHEVRON_LEFT
	}

	imgui.PushStyleColorVec4(imgui.ColButton, imgui.Vec4{})
	imgui.PushStyleColorVec4(imgui.ColButtonHovered, imgui.CurrentStyle().Colors()[imgui.ColHeaderHovered])
	imgui.PushStyleColorVec4(imgui.ColButtonActive, imgui.CurrentStyle().Colors()[imgui.ColHeaderActive])

	if imgui.Button(icon + h.imguiID() + "_toggle") {
		h.Toggle()
	}

	imgui.PopStyleColorV(3)

	// title (only if there's room)
	if h.CurrentWidth > h.MinWidth+50 && h.Title != "" {
		imgui.SameLine()
		imgui.TextUnformatted(h.Title)
	}
}

// drawContent draws the content area.
// only called when fully expanded to avoid scrollbar overlap with toggle button.
func (h *HCollapse) drawContent(state *State) {
	// position cursor below header
	imgui.SetCursorPosY(HCollapseHeaderHeight)

	// get available region inside the outer child window
	avail := imgui.ContentRegionAvail()

	// reserve space for resize handle if present
	contentWidth := avail.X
	if h.Resizable {
		contentWidth -= DashDragHandleOffset
	}
	contentHeight := avail.Y

	if contentWidth <= 0 || contentHeight <= 0 {
		return
	}

	imgui.PushStyleVarVec2(imgui.StyleVarWindowPadding, imgui.Vec2{})
	contentFlags := imgui.WindowFlagsNoScrollbar | imgui.WindowFlagsNoScrollWithMouse
	imgui.BeginChildStrV(h.imguiID()+"_content", imgui.Vec2{X: contentWidth, Y: contentHeight}, 0, contentFlags)
	imgui.PopStyleVar() // window padding

	if h.Content != nil {
		childState := &State{
			Size:     imgui.Vec2{X: contentWidth, Y: contentHeight},
			Position: imgui.Vec2{},
			IO:       state.IO,
			App:      state.App,
			Parent:   h,
		}
		h.Content.Draw(childState)
	}

	imgui.EndChild()
}

// drawResizeHandle draws the resize handle on the right edge.
func (h *HCollapse) drawResizeHandle(state *State) {
	handlePos := imgui.Vec2{
		X: h.CurrentWidth - DashDragHandleOffset,
		Y: DefaultItemSpacing + 1,
	}
	imgui.SetCursorPos(handlePos)

	imgui.PushStyleColorVec4(imgui.ColText, imgui.CurrentStyle().Colors()[imgui.ColHeaderActive])
	imgui.TextUnformatted(fonts.ICON_DRAG_INDICATOR)
	imgui.PopStyleColor()

	imgui.SetCursorPos(handlePos)
	imgui.InvisibleButton(h.imguiID()+"_resize", imgui.Vec2{X: HCollapseResizeHandleSize, Y: HCollapseResizeHandleSize})

	if imgui.IsItemHovered() {
		imgui.SetMouseCursor(imgui.MouseCursorResizeEW)
	}

	if imgui.IsItemActive() {
		delta := imgui.CurrentIO().MouseDelta().X
		h.CurrentWidth += delta
		h.ExpandedWidth += delta

		// clamp to bounds
		if h.CurrentWidth < h.MinWidth {
			h.CurrentWidth = h.MinWidth
			h.ExpandedWidth = h.MinWidth
		}
		if h.MaxWidth > 0 && h.CurrentWidth > h.MaxWidth {
			h.CurrentWidth = h.MaxWidth
			h.ExpandedWidth = h.MaxWidth
		}
		if h.CurrentWidth > state.Size.X-50 {
			h.CurrentWidth = state.Size.X - 50
			h.ExpandedWidth = state.Size.X - 50
		}
	}
}

// animate updates CurrentWidth toward the target width.
func (h *HCollapse) animate() {
	target := h.MinWidth
	if h.Expanded {
		target = h.ExpandedWidth
	}

	if h.CurrentWidth < target {
		h.CurrentWidth += h.pxPerFrame()
		if h.CurrentWidth > target {
			h.CurrentWidth = target
		}
	} else if h.CurrentWidth > target {
		h.CurrentWidth -= h.pxPerFrame()
		if h.CurrentWidth < target {
			h.CurrentWidth = target
		}
	}
}

// pxPerFrame calculates pixels to animate per frame for smooth transitions.
func (h *HCollapse) pxPerFrame() float32 {
	msFrame := FramerateToMs / imgui.CurrentIO().Framerate()
	frames := float32(h.TransitionMs) / msFrame
	return float32(math.Ceil(float64(h.ExpandedWidth) / float64(frames)))
}

// isFullyExpanded returns true if the animation has completed to expanded state.
func (h *HCollapse) isFullyExpanded() bool {
	return h.Expanded && h.CurrentWidth >= h.ExpandedWidth
}

// Actions implements Component by delegating to the content component.
func (h *HCollapse) Actions() *ActionRegistry {
	if h.Content != nil {
		return h.Content.Actions()
	}
	return h.Container.Actions()
}
