package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

func install(path, name string, flags *Flags) error {
	var info os.FileInfo

	// stat the api binary
	if v, err := os.Stat(filepath.Join(path, name)); err != nil {
		return err
	} else {
		info = v
	}

	// check for an existing symlink
	var result error
	for _, cmd := range flags.cmds {
		// Check for existing symlink
		if stat, err := os.Stat(filepath.Join(path, cmd.Name)); err == nil {
			if !statEquals(info, stat) {
				result = errors.Join(result, fmt.Errorf("failed to install %q: file exists but doesn't match", cmd.Name))
			} else {
				continue
			}
		} else if !errors.Is(err, os.ErrNotExist) {
			result = errors.Join(result, fmt.Errorf("failed to install %q: %w", cmd.Name, err))
		}
		// Make symlink
		if err := os.Symlink(filepath.Join(path, name), filepath.Join(path, cmd.Name)); err != nil {
			result = errors.Join(result, fmt.Errorf("failed to install %q: %w", cmd.Name, err))
		}
	}
	// Return any errors
	return result
}

func statEquals(a, b os.FileInfo) bool {
	return a.Size() == b.Size() && a.ModTime() == b.ModTime()
}
