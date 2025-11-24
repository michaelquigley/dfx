package main

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/michaelquigley/dfx"
	"github.com/michaelquigley/dfx/fonts"
	"github.com/michaelquigley/df/dl"
	"github.com/sqweek/dialog"
	"golang.design/x/clipboard"
)

func main() {
	// initialize clipboard for copy functionality
	if err := clipboard.Init(); err != nil {
		fmt.Printf("error initializing clipboard: %v\n", err)
		return
	}

	// create log buffer
	buffer := dfx.NewLogBuffer(512)

	// create slog handler for df/dl integration
	handler := dfx.NewSlogHandler(buffer, &dfx.SlogHandlerOptions{
		TrimPrefix: "git.hq.quigley.com/products/baab/",
		MinLevel:   slog.LevelDebug,
		StartTime:  time.Now(),
	})

	// initialize df/dl logger with our handler
	opts := dl.DefaultOptions().SetTrimPrefix("git.hq.quigley.com/products/baab/")
	opts.CustomHandler = handler
	dl.Init(opts)

	// create log viewer
	viewer := dfx.NewLogViewer(buffer)
	viewer.AutoScroll = true
	viewer.LevelFilter = slog.LevelDebug
	viewer.ShowTime = true
	viewer.ShowFunc = true
	viewer.ShowFields = true

	// create toolbar with controls
	toolbar := dfx.NewFunc(func(state *dfx.State) {
		// copy button
		if dfx.Button(fonts.ICON_COPY_ALL) {
			text := buffer.AllText()
			clipboard.Write(clipboard.FmtText, []byte(text))
			dl.Log().Info("copied log to clipboard")
		}

		dfx.SameLine()
		if dfx.Button(fonts.ICON_SAVE) {
			if filename, err := dialog.File().Filter("log file", "log").Title("Save Log").Save(); err == nil {
				// ensure .log extension
				if filepath.Ext(filename) != ".log" {
					filename = filename + ".log"
				}
				if err := os.WriteFile(filename, []byte(buffer.AllText()), 0644); err != nil {
					dl.Log().Errorf("error saving log file '%v': %v", filename, err)
				} else {
					dl.Log().Infof("saved log to '%v'", filename)
				}
			}
		}

		dfx.SameLine()
		if dfx.Button(fonts.ICON_CLEAR_ALL) {
			buffer.Clear()
			dl.Log().Info("cleared log buffer")
		}

		dfx.SameLine()
		dfx.Text(fmt.Sprintf("Messages: %d", buffer.Count()))

		dfx.SameLine()
		if dfx.Button("Generate Logs") {
			generateTestLogs()
		}
	})

	// create container with toolbar and viewer
	root := &dfx.Container{
		Visible: true,
		OnDraw: func(state *dfx.State) {
			toolbar.Draw(state)
			dfx.Separator()
		},
		Children: []dfx.Component{viewer},
	}

	// generate some initial log messages
	dl.Log().Info("application starting")
	dl.Log().Debug("debug message example")
	dl.Log().Warn("warning message example")
	dl.Log().Error("error message example")
	dl.Log().With("key1", "value1").With("key2", 42).Info("message with fields")

	// create and run application
	app := dfx.New(root, dfx.Config{
		Title:  "Log Viewer Example",
		Width:  1000,
		Height: 600,
		OnSetup: func(app *dfx.App) {
			// register keyboard shortcuts
			app.Actions().Register("quit", "Ctrl+Q", func() {
				dl.Log().Info("quitting application")
				app.Stop()
			})
		},
	})

	dl.Log().Info("starting application")

	go func() {
		for {
			time.Sleep(500 * time.Millisecond)
			generateTestLogs()
		}
	}()

	app.Run()
}

func generateTestLogs() {
	dl.Log().Debug("this is a debug message")
	dl.Log().Info("this is an info message")
	dl.Log().Warn("this is a warning message")
	dl.Log().Error("this is an error message")
	dl.Log().With("timestamp", time.Now()).With("count", 123).Info("info message with fields")
	dl.Log().With("details", "some detailed information").Warn("warning with context")
}
