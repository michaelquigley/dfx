package dfx

import "github.com/AllenDang/cimgui-go/imgui"

// NodeCanvasStyle holds every color and metric the NodeCanvas renders with.
// it is plain data and a complete value, not a sparse overlay: a partially
// filled struct is used as-is, zero fields included. callers wanting partial
// customization start from DefaultNodeCanvasStyle() and mutate fields —
// necessarily after an imgui context exists (e.g. in OnSetup, or later via
// SetStyle).
//
// metric unit spaces are pinned per field group: render metrics are
// canvas-space and are multiplied by the current zoom before reaching the
// drawlist, so chrome scales with the nodes it decorates; hit and snap
// tolerances are screen-pixel constants, so targets stay equally grabbable
// at every detent.
type NodeCanvasStyle struct {
	// colors
	GridColor               imgui.Vec4
	NodeBodyColor           imgui.Vec4
	NodeBorderColor         imgui.Vec4
	NodeBorderColorHovered  imgui.Vec4
	NodeBorderColorSelected imgui.Vec4
	TitleBandColor          imgui.Vec4
	TitleBandColorSelected  imgui.Vec4
	PinColor                imgui.Vec4
	PinColorHovered         imgui.Vec4
	LinkColor               imgui.Vec4
	LinkColorHovered        imgui.Vec4
	LinkColorSelected       imgui.Vec4
	BoxSelectFillColor      imgui.Vec4
	BoxSelectBorderColor    imgui.Vec4
	ZoomOverlayBgColor      imgui.Vec4
	ZoomOverlayTextColor    imgui.Vec4

	// render metrics, canvas-space
	NodeRounding            float32
	NodePadding             float32
	BorderThickness         float32
	BorderThicknessSelected float32
	PinRadius               float32
	LinkThickness           float32
	LinkTangent             float32

	// hit and snap tolerances, screen px
	PinHitRadius    float32
	LinkHitDistance float32
	LinkSnapRadius  float32

	// zoom overlay metrics, screen px: the overlay is viewport chrome and
	// does not scale with the graph.
	ZoomOverlayPadding  float32
	ZoomOverlayMargin   float32
	ZoomOverlayRounding float32
}

// hitParams extracts the screen-pixel tolerances the hit-testing and gesture
// machinery consume.
func (s *NodeCanvasStyle) hitParams() hitParams {
	return hitParams{
		pinHitRadius:    s.PinHitRadius,
		linkHitDistance: s.LinkHitDistance,
		linkSnapRadius:  s.LinkSnapRadius,
	}
}

// DefaultNodeCanvasStyle derives a complete style from the active dfx theme,
// reading imgui style colors the way the theme system writes them. it
// requires a live imgui context: components are constructed before app.Run
// creates the context and applies the theme, so a zero-valued
// NodeCanvasConfig.Style is resolved lazily at the first Begin rather than
// at construction. apps that rebuild style on theme change assign a fresh
// derivation via SetStyle at any time.
func DefaultNodeCanvasStyle() NodeCanvasStyle {
	colors := imgui.CurrentStyle().Colors()
	border := colors[imgui.ColBorder]
	text := colors[imgui.ColText]
	body := colors[imgui.ColPopupBg]
	title := colors[imgui.ColTitleBgActive]
	emphasis := colors[imgui.ColHeaderActive]
	hover := colors[imgui.ColHeaderHovered]

	return NodeCanvasStyle{
		GridColor:               withAlpha(border, border.W*0.5),
		NodeBodyColor:           body,
		NodeBorderColor:         border,
		NodeBorderColorHovered:  hover,
		NodeBorderColorSelected: emphasis,
		TitleBandColor:          title,
		TitleBandColorSelected:  emphasis,
		PinColor:                withAlpha(text, 0.8),
		PinColorHovered:         withAlpha(emphasis, 1),
		LinkColor:               withAlpha(text, 0.6),
		LinkColorHovered:        withAlpha(hover, 1),
		LinkColorSelected:       withAlpha(emphasis, 1),
		BoxSelectFillColor:      withAlpha(emphasis, 0.2),
		BoxSelectBorderColor:    withAlpha(emphasis, 0.8),
		ZoomOverlayBgColor:      withAlpha(body, 0.75),
		ZoomOverlayTextColor:    withAlpha(text, 0.85),

		NodeRounding:            4,
		NodePadding:             8,
		BorderThickness:         1,
		BorderThicknessSelected: 2,
		PinRadius:               5,
		LinkThickness:           2,
		LinkTangent:             50,

		PinHitRadius:    10,
		LinkHitDistance: 6,
		LinkSnapRadius:  24,

		ZoomOverlayPadding:  4,
		ZoomOverlayMargin:   8,
		ZoomOverlayRounding: 3,
	}
}

func withAlpha(c imgui.Vec4, a float32) imgui.Vec4 {
	c.W = a
	return c
}

// highlight is the selected form of a declared color: lifted 30% toward white, fully opaque. selection shows as a
// thicker line in the element's own color, brighter than its unselected state.
func highlight(c imgui.Vec4) imgui.Vec4 {
	const lift = 0.3
	return imgui.Vec4{X: c.X + (1-c.X)*lift, Y: c.Y + (1-c.Y)*lift, Z: c.Z + (1-c.Z)*lift, W: 1}
}

// titleBandColor resolves a node's title band: the declared accent when set, whether or not the node is selected;
// otherwise the style's selected or normal band color.
func titleBandColor(s *NodeCanvasStyle, flags NodeFlags) imgui.Vec4 {
	if flags.Accent != (imgui.Vec4{}) {
		return flags.Accent
	}
	if flags.Selected {
		return s.TitleBandColorSelected
	}
	return s.TitleBandColor
}

// nodeBorderColor resolves a node's border: selected, a highlight of the accent when set or the style's selected
// color; otherwise the style's hover or normal color.
func nodeBorderColor(s *NodeCanvasStyle, flags NodeFlags, hovered bool) imgui.Vec4 {
	switch {
	case flags.Selected && flags.Accent != (imgui.Vec4{}):
		return highlight(flags.Accent)
	case flags.Selected:
		return s.NodeBorderColorSelected
	case hovered:
		return s.NodeBorderColorHovered
	}
	return s.NodeBorderColor
}

// pinColor resolves a pin marker: the style's hover color while hovered, the node's accent when set, the style's
// pin color otherwise.
func pinColor(s *NodeCanvasStyle, flags NodeFlags, hovered bool) imgui.Vec4 {
	if hovered {
		return s.PinColorHovered
	}
	if flags.Accent != (imgui.Vec4{}) {
		return flags.Accent
	}
	return s.PinColor
}

// linkColor resolves a link: selected, a highlight of the declared color when set or the style's selected color;
// then the style's hover color; then the declared color; then the style's link color.
func linkColor(s *NodeCanvasStyle, selected, hovered bool, declared imgui.Vec4) imgui.Vec4 {
	set := declared != (imgui.Vec4{})
	switch {
	case selected && set:
		return highlight(declared)
	case selected:
		return s.LinkColorSelected
	case hovered:
		return s.LinkColorHovered
	case set:
		return declared
	}
	return s.LinkColor
}
