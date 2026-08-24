//go:build !windows

package bayes

import (
	"errors"
	"fmt"
	"os"
)

func syncDirectory(path string) error {
	directory, err := os.Open(path) // #nosec G304 -- path is the validated destination parent.
	if err != nil {
		return fmt.Errorf("failed to open destination directory: %w", err)
	}

	syncErr := directory.Sync()
	closeErr := directory.Close()

	return errors.Join(syncErr, closeErr)
}
