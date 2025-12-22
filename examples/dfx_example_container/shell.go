package main

import (
	"fmt"

	"github.com/michaelquigley/df/da"
	"github.com/michaelquigley/dfx"
)

const workspacesTag = "workspaces"

type shellFactory struct{}

func (f *shellFactory) Build(a *da.Application[config]) error {
	shl := &shell{
		cfg:       &a.Cfg,
		workspace: dfx.NewWorkspace(),
	}

	shl.root = dfx.NewFunc(func(state *dfx.State) {
		shl.workspace.Draw(state) // draws combo selector + current workspace
	})

	shl.app = dfx.New(shl.root, dfx.Config{
		Title:  "dfx container example",
		Width:  shl.cfg.WindowWidth,
		Height: shl.cfg.WindowHeight,
		X:      shl.cfg.WindowX,
		Y:      shl.cfg.WindowY,
		OnSizeChange: func(w, h int) {
			shl.cfg.WindowWidth = w
			shl.cfg.WindowHeight = h
		},
		OnClose: func(app *dfx.App) {
			shl.cfg.WindowX, shl.cfg.WindowY = app.GetWindowPos()
		},
	})

	da.Set(a.C, shl)
	return nil
}

type shell struct {
	cfg       *config
	app       *dfx.App
	root      dfx.Component
	workspace *dfx.Workspace
}

// Link wires up all tagged workspace components
func (s *shell) Link(c *da.Container) error {
	for i, ws := range da.TaggedAsType[dfx.Component](c, workspacesTag) {
		s.workspace.Add(fmt.Sprintf("ws-%d", i), fmt.Sprintf("workspace %d", i+1), ws)
	}
	return nil
}

func (s *shell) Start() error {
	go s.app.Run()
	return nil
}
