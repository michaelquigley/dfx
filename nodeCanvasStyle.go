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
	}
}

func withAlpha(c imgui.Vec4, a float32) imgui.Vec4 {
	c.W = a
	return c
}
