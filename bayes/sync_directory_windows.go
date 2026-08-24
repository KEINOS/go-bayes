//go:build windows

package bayes

// syncDirectory is unavailable on Windows. The database file is synced before
// replacement, but Windows does not support flushing a directory through
// os.File.Sync.
func syncDirectory(_ string) error {
	return nil
}
