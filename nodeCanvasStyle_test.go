package dfx

import (
	"testing"

	"github.com/AllenDang/cimgui-go/imgui"
)

func colorTestStyle() NodeCanvasStyle {
	return NodeCanvasStyle{
		NodeBorderColor:         imgui.Vec4{Z: 0.1, W: 1},
		NodeBorderColorHovered:  imgui.Vec4{Z: 0.2, W: 1},
		NodeBorderColorSelected: imgui.Vec4{Z: 0.3, W: 1},
		TitleBandColor:          imgui.Vec4{X: 0.1, W: 1},
		TitleBandColorSelected:  imgui.Vec4{X: 0.2, W: 1},
		PinColor:                imgui.Vec4{X: 0.3, W: 1},
		PinColorHovered:         imgui.Vec4{X: 0.4, W: 1},
		LinkColor:               imgui.Vec4{X: 0.5, W: 1},
		LinkColorHovered:        imgui.Vec4{X: 0.6, W: 1},
		LinkColorSelected:       imgui.Vec4{X: 0.7, W: 1},
	}
}

// an unset accent falls back to the style for the title band, border, and pins; a set one replaces the band and pin
// colors and gives the selected border a highlight of itself, while hover keeps the style's colors.
func TestNodeAccentFallsBackToStyle(t *testing.T) {
	s := colorTestStyle()
	accent := imgui.Vec4{Y: 0.8, W: 1}

	if got := titleBandColor(&s, NodeFlags{}); got != s.TitleBandColor {
		t.Errorf("unset band = %v, want style %v", got, s.TitleBandColor)
	}
	if got := titleBandColor(&s, NodeFlags{Selected: true}); got != s.TitleBandColorSelected {
		t.Errorf("unset selected band = %v, want style %v", got, s.TitleBandColorSelected)
	}
	if got := titleBandColor(&s, NodeFlags{Accent: accent}); got != accent {
		t.Errorf("accented band = %v, want %v", got, accent)
	}
	if got := titleBandColor(&s, NodeFlags{Selected: true, Accent: accent}); got != accent {
		t.Errorf("accented selected band = %v, want the accent; selection shows on the border", got)
	}

	if got := nodeBorderColor(&s, NodeFlags{Selected: true}, false); got != s.NodeBorderColorSelected {
		t.Errorf("unset selected border = %v, want style %v", got, s.NodeBorderColorSelected)
	}
	if got := nodeBorderColor(&s, NodeFlags{Selected: true, Accent: accent}, false); got != highlight(accent) {
		t.Errorf("accented selected border = %v, want the accent's highlight %v", got, highlight(accent))
	}
	if got := nodeBorderColor(&s, NodeFlags{Accent: accent}, true); got != s.NodeBorderColorHovered {
		t.Errorf("hovered accented border = %v, want style hover %v", got, s.NodeBorderColorHovered)
	}
	if got := nodeBorderColor(&s, NodeFlags{Accent: accent}, false); got != s.NodeBorderColor {
		t.Errorf("unselected accented border = %v, want style %v", got, s.NodeBorderColor)
	}
	if h := highlight(accent); h.W != 1 || h.Y <= accent.Y || h.X <= accent.X {
		t.Errorf("highlight %v is not brighter and opaque", h)
	}

	if got := pinColor(&s, NodeFlags{}, false); got != s.PinColor {
		t.Errorf("unset pin = %v, want style %v", got, s.PinColor)
	}
	if got := pinColor(&s, NodeFlags{Accent: accent}, false); got != accent {
		t.Errorf("accented pin = %v, want %v", got, accent)
	}
	if got := pinColor(&s, NodeFlags{Accent: accent}, true); got != s.PinColorHovered {
		t.Errorf("hovered accented pin = %v, want style hover %v", got, s.PinColorHovered)
	}
}

// an unset link color falls back to the style; a set one is the normal color, selected it draws as a highlight of
// itself, and hover keeps the style's color.
func TestLinkColorFallsBackToStyle(t *testing.T) {
	s := colorTestStyle()
	color := imgui.Vec4{Z: 0.9, W: 1}
	cases := []struct {
		name              string
		selected, hovered bool
		declared, want    imgui.Vec4
	}{
		{"unset", false, false, imgui.Vec4{}, s.LinkColor},
		{"set", false, false, color, color},
		{"set hovered", false, true, color, s.LinkColorHovered},
		{"set selected", true, false, color, highlight(color)},
		{"set selected and hovered", true, true, color, highlight(color)},
		{"unset selected", true, false, imgui.Vec4{}, s.LinkColorSelected},
		{"unset hovered", false, true, imgui.Vec4{}, s.LinkColorHovered},
	}
	for _, c := range cases {
		if got := linkColor(&s, c.selected, c.hovered, c.declared); got != c.want {
			t.Errorf("%v: %v, want %v", c.name, got, c.want)
		}
	}
}
