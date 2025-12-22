package main

import (
	"os"
	"path/filepath"

	"github.com/michaelquigley/df/da"
	"github.com/michaelquigley/df/dd"
	"github.com/michaelquigley/df/dl"
)

func main() {
	cfgPath, err := configPath()
	if err != nil {
		dl.Fatalf("error getting config path: %v", err)
	}

	app := da.NewApplication[config](defaultConfig())

	// register factories - order determines workspace order
	app.Factories = append(app.Factories, &counterPanelFactory{})
	app.Factories = append(app.Factories, &infoPanelFactory{})
	app.Factories = append(app.Factories, &shellFactory{})

	if err := app.InitializeWithPaths(da.OptionalPath(cfgPath)); err != nil {
		dl.Fatalf("error initializing: %v", err)
	}

	if err := app.Start(); err != nil {
		dl.Fatalf("error starting: %v", err)
	}

	// wait for GUI to close
	if shl, ok := da.Get[*shell](app.C); ok {
		shl.app.Wait()
	}

	// save config
	if err := os.MkdirAll(filepath.Dir(cfgPath), 0755); err != nil {
		dl.Errorf("error creating config directory: %v", err)
	}
	if err := dd.UnbindYAMLFile(app.Cfg, cfgPath); err != nil {
		dl.Errorf("error saving config: %v", err)
	}

	if err := app.Stop(); err != nil {
		dl.Errorf("error stopping: %v", err)
	}
}
