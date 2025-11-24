package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/michaelquigley/dfx"
)

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("error getting working directory: %v\n", err)
		return
	}

	// build the filesystem tree
	root, err := dfx.BuildTree(cwd, nil)
	if err != nil {
		fmt.Printf("error building tree: %v\n", err)
		return
	}

	// create file tree component
	fileTree := dfx.NewFileTree(root)

	// configure callbacks
	fileTree.OnSelect = func(node *dfx.FileNode) {
		fmt.Printf("selected: %v (dir: %v)\n", node.Path(), node.Dir)
	}

	fileTree.OnDoubleClick = func(node *dfx.FileNode) {
		fmt.Printf("double-clicked: %v\n", node.Path())
		if !node.Dir {
			fmt.Println("  (would open file here)")
		}
	}

	// optional: filter to show only .go files and directories
	fileTree.Filter = func(node *dfx.FileNode) bool {
		if node.Dir {
			return true
		}
		ext := filepath.Ext(node.Name)
		return ext == ".go" || ext == ".md"
	}

	// create application
	app := dfx.New(fileTree, dfx.Config{
		Title:  "File Tree Example",
		Width:  800,
		Height: 600,
	})

	// run application
	app.Run()
}
