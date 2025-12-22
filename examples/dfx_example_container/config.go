package main

import (
	"os"
	"path/filepath"
)

type config struct {
	WindowX      int
	WindowY      int
	WindowWidth  int
	WindowHeight int
	Counter      int
}

func defaultConfig() config {
	return config{
		WindowX:      100,
		WindowY:      100,
		WindowWidth:  800,
		WindowHeight: 600,
		Counter:      0,
	}
}

func configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".dfx_example_container", "config.yml"), nil
}
