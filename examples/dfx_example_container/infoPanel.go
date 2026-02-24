package main

import (
	"fmt"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/michaelquigley/df/da"
	"github.com/michaelquigley/dfx"
)

type infoPanelFactory struct{}

func (f *infoPanelFactory) Build(a *da.Application[config]) error {
	panel := &infoPanel{cfg: &a.Cfg}
	da.AddTagged(a.C, workspacesTag, panel)
	return nil
}

type infoPanel struct {
	cfg *config
}

func (p *infoPanel) Draw(state *dfx.State) {
	imgui.Text("info workspace")
	imgui.Separator()
	imgui.Spacing()

	imgui.Text("this example demonstrates:")
	imgui.BulletText("da.Application container-based lifecycle")
	imgui.BulletText("factory pattern for component creation")
	imgui.BulletText("tagged components for modular registration")
	imgui.BulletText("dd.UnbindYAMLFile for config persistence")
	imgui.BulletText("dfx.Workspace for view switching")

	imgui.Spacing()
	imgui.Separator()
	imgui.Spacing()

	imgui.Text("current configuration:")
	imgui.Text(fmt.Sprintf("  window: %dx%d at (%d,%d)",
		p.cfg.WindowWidth, p.cfg.WindowHeight,
		p.cfg.WindowX, p.cfg.WindowY))
	imgui.Text(fmt.Sprintf("  counter: %d", p.cfg.Counter))
}

func (p *infoPanel) Actions() *dfx.ActionRegistry {
	return nil
}

var _ dfx.Component = (*infoPanel)(nil)
