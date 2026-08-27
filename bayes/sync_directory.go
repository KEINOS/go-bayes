//go:build !windows

package bayes

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

func syncDirectory(path string) error {
	directory, err := os.Open(filepath.Clean(path))
	if err != nil {
		return fmt.Errorf("failed to open destination directory: %w", err)
	}

	syncErr := directory.Sync()
	closeErr := directory.Close()

	return errors.Join(syncErr, closeErr)
}
