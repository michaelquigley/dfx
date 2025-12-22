package main

import (
	"fmt"

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
	dfx.Text("counter workspace")
	dfx.Separator()
	dfx.Spacing()

	dfx.Text(fmt.Sprintf("current value: %d", p.cfg.Counter))
	dfx.Spacing()

	if dfx.Button("increment") {
		p.cfg.Counter++
	}
	dfx.SameLine()
	if dfx.Button("decrement") {
		p.cfg.Counter--
	}
	dfx.SameLine()
	if dfx.Button("reset") {
		p.cfg.Counter = 0
	}
}

func (p *counterPanel) Actions() *dfx.ActionRegistry {
	return nil
}

var _ dfx.Component = (*counterPanel)(nil)
