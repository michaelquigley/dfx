package dfx

import "github.com/AllenDang/cimgui-go/imgui"

// cimgui-go v1.6.0's vertical slider reserves 2px at each end, then half the
// grab height. keep paint and scale on the same travel as SliderBehavior.
type faderGeometry struct {
	min, max                              imgui.Vec2
	top, bottom                           float32
	centerX                               float32
	handleWidth, handleHeight, trackWidth float32
}

func newFaderGeometry(min, max imgui.Vec2, params FaderParams) faderGeometry {
	return faderGeometry{
		min: min, max: max,
		top:         min.Y + 2 + params.HandleHeight/2,
		bottom:      max.Y - 2 - params.HandleHeight/2,
		centerX:     (min.X + max.X) / 2,
		handleWidth: params.HandleWidth, handleHeight: params.HandleHeight,
		trackWidth: params.TrackWidth,
	}
}

func (g faderGeometry) y(position float32) float32 {
	return g.bottom - position*(g.bottom-g.top)
}

// faderPalette is the fader's resolved colors for one draw. style alpha is applied
// when they are converted for drawing, once.
type faderPalette struct {
	track, fill, border, body imgui.Vec4
}

// resolveFaderPalette picks the fader's colors from the current style through the
// selected role bindings. explicit AccentColor and TrackColor win.
func resolveFaderPalette(params FaderParams, hovered, active bool) faderPalette {
	roles := currentStyleRoles()
	p := faderPalette{
		track:  styleColor(imgui.ColFrameBg),
		fill:   styleColor(roles.FaderFill),
		border: styleColor(imgui.ColBorder),
		body:   styleColor(imgui.ColFrameBg),
	}
	if params.TrackColor != nil {
		p.track = *params.TrackColor
	}
	if params.AccentColor != nil {
		p.fill = *params.AccentColor
	}
	if hovered {
		p.border = p.fill
		p.body = styleColor(imgui.ColFrameBgHovered)
	}
	if active {
		p.border = styleColor(roles.FaderActive)
		if params.AccentColor != nil {
			p.border = p.fill
		}
		p.body = styleColor(imgui.ColFrameBgActive)
	}
	return p
}

func drawFader(params FaderParams, g faderGeometry, position float32, hovered, active bool) {
	dl := imgui.WindowDrawList()
	y := g.y(position)
	palette := resolveFaderPalette(params, hovered, active)
	left, right := g.centerX-g.trackWidth/2, g.centerX+g.trackWidth/2
	dl.AddRectFilled(imgui.Vec2{X: left, Y: g.top}, imgui.Vec2{X: right, Y: g.bottom}, imgui.ColorU32Vec4(palette.track))
	if y < g.bottom {
		dl.AddRectFilled(imgui.Vec2{X: left, Y: y}, imgui.Vec2{X: right, Y: g.bottom}, imgui.ColorU32Vec4(palette.fill))
	}

	handleMin := imgui.Vec2{X: g.centerX - g.handleWidth/2, Y: y - g.handleHeight/2}
	handleMax := imgui.Vec2{X: g.centerX + g.handleWidth/2, Y: y + g.handleHeight/2}
	rounding := min(g.handleWidth, g.handleHeight) / 2
	// two soft shadow layers, inset enough to stay within the item's bounds.
	shadow := imgui.ColorU32Vec4(styleColor(imgui.ColBorderShadow))
	for _, offset := range []float32{2, 1} {
		dl.AddRectFilledV(imgui.Vec2{X: handleMin.X, Y: handleMin.Y + offset}, imgui.Vec2{X: handleMax.X, Y: handleMax.Y + offset}, shadow, rounding, imgui.DrawFlagsNone)
	}
	dl.AddRectFilledV(handleMin, handleMax, imgui.ColorU32Vec4(palette.border), rounding, imgui.DrawFlagsNone)
	bevel := min(3, min(g.handleWidth, g.handleHeight)/4)
	dl.AddRectFilledV(imgui.Vec2{X: handleMin.X + bevel, Y: handleMin.Y + bevel}, imgui.Vec2{X: handleMax.X - bevel, Y: handleMax.Y - bevel}, imgui.ColorU32Vec4(palette.body), max(0, rounding-bevel), imgui.DrawFlagsNone)
	markerHalfWidth := g.handleWidth * 0.22
	markerHalfHeight := min(1.5, g.handleHeight/8)
	dl.AddRectFilled(imgui.Vec2{X: g.centerX - markerHalfWidth, Y: y - markerHalfHeight}, imgui.Vec2{X: g.centerX + markerHalfWidth, Y: y + markerHalfHeight}, imgui.ColorU32Vec4(styleColor(imgui.ColText)))
}
