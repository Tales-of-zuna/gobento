package generate

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// retryDelays are the waits between attempts of a rename that fails because
// another process, such as an antivirus scanner or indexer on Windows,
// briefly holds a new file open.
func retryDelays() []time.Duration {
	return []time.Duration{50 * time.Millisecond, 100 * time.Millisecond, 200 * time.Millisecond, 400 * time.Millisecond, 800 * time.Millisecond}
}

// place moves the project from tmp into target. A missing target is created
// by renaming tmp. An existing empty target receives each top-level entry of
// tmp, after which tmp is removed. It returns the paths it moved into an
// existing target, so they can be removed if a later step fails.
func place(tmp, target string) ([]string, error) {
	if _, err := os.Stat(target); errors.Is(err, fs.ErrNotExist) {
		if err := rename(tmp, target); err != nil {
			return nil, fmt.Errorf("move project into place: %w", err)
		}
		return nil, nil
	}

	entries, err := os.ReadDir(tmp)
	if err != nil {
		return nil, err
	}
	var moved []string
	for _, e := range entries {
		dst := filepath.Join(target, e.Name())
		if err := rename(filepath.Join(tmp, e.Name()), dst); err != nil {
			return moved, fmt.Errorf("move %s into place: %w", e.Name(), err)
		}
		moved = append(moved, dst)
	}
	if err := os.Remove(tmp); err != nil {
		return moved, fmt.Errorf("remove temporary directory: %w", err)
	}
	return moved, nil
}

// rename renames oldpath to newpath, retrying on Windows when access is
// denied.
func rename(oldpath, newpath string) error {
	return retry(func() error { return os.Rename(oldpath, newpath) }, isTransient, retryDelays())
}

// isTransient reports whether err may go away on its own: an access-denied
// error on Windows.
func isTransient(err error) bool {
	return runtime.GOOS == "windows" && errors.Is(err, fs.ErrPermission)
}

// retry calls op until it succeeds, returns an error retryable rejects, or
// the delays are used up.
func retry(op func() error, retryable func(error) bool, delays []time.Duration) error {
	err := op()
	for _, d := range delays {
		if err == nil || !retryable(err) {
			return err
		}
		time.Sleep(d)
		err = op()
	}
	return err
}

// cleanup removes the temporary directory and every entry already moved
// into a pre-existing target. The target directory itself is never removed.
func cleanup(tmp string, moved []string) error {
	var errs []error
	if err := os.RemoveAll(tmp); err != nil {
		errs = append(errs, fmt.Errorf("remove %s: %w", tmp, err))
	}
	for _, p := range moved {
		if err := os.RemoveAll(p); err != nil {
			errs = append(errs, fmt.Errorf("remove %s: %w", p, err))
		}
	}
	return errors.Join(errs...)
}
