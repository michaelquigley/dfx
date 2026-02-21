package main

import (
	"fmt"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/michaelquigley/df/da"
	"github.com/michaelquigley/dfx"
)

type counterPanelFactory struct{}

func (f *counterPanelFactory) Build(a *da.Application[config]) error {
	panel := &counterPanel{cfg: &a.Cfg}
	da.AddTagged(a.C, workspacesTag, panel)
	return nil
}

type counterPanel struct {
	cfg *config
}

func (p *counterPanel) Draw(state *dfx.State) {
	imgui.Text("counter workspace")
	imgui.Separator()
	imgui.Spacing()

	imgui.Text(fmt.Sprintf("current value: %d", p.cfg.Counter))
	imgui.Spacing()

	if imgui.Button("increment") {
		p.cfg.Counter++
	}
	imgui.SameLine()
	if imgui.Button("decrement") {
		p.cfg.Counter--
	}
	imgui.SameLine()
	if imgui.Button("reset") {
		p.cfg.Counter = 0
	}
}

func (p *counterPanel) Actions() *dfx.ActionRegistry {
	return nil
}

var _ dfx.Component = (*counterPanel)(nil)
