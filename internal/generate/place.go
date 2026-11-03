package generate

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

// dirMode is the permission of every directory gobento creates in a
// generated project, before the umask.
const dirMode fs.FileMode = 0o755

// ErrCleanup reports that gobento could not remove everything it created.
var ErrCleanup = errors.New("cleanup failed")

// retryDelays are the waits between attempts of a file operation that fails
// because another process, such as an antivirus scanner or indexer on
// Windows, briefly holds a new file open.
func retryDelays() []time.Duration {
	return []time.Duration{50 * time.Millisecond, 100 * time.Millisecond, 200 * time.Millisecond, 400 * time.Millisecond, 800 * time.Millisecond}
}

// targetExists reports whether target exists, without following a final
// symbolic link. Errors other than "does not exist" are returned.
func targetExists(target string) (bool, error) {
	_, err := os.Lstat(target)
	switch {
	case err == nil:
		return true, nil
	case errors.Is(err, fs.ErrNotExist):
		return false, nil
	default:
		return false, fmt.Errorf("%w: %w", ErrTarget, err)
	}
}

// workDir creates the temporary directory a project is built in. For a
// missing target it is created next to the target, so that placing the
// project is a single rename. For an existing empty target it is created
// inside the target, so that moving the files up one level stays on one
// file system and needs no write access to the target's parent.
func workDir(target string, exists bool) (string, error) {
	base := filepath.Dir(target)
	if exists {
		base = target
	}
	tmp, err := os.MkdirTemp(base, tempPrefix+"*")
	if err != nil {
		return "", fmt.Errorf("create temporary directory: %w", err)
	}
	// MkdirTemp creates 0700 directories; the project root is a normal
	// directory.
	if err := os.Chmod(tmp, dirMode); err != nil {
		return "", errors.Join(fmt.Errorf("create temporary directory: %w", err), os.Remove(tmp))
	}
	return tmp, nil
}

// place moves the project from tmp into target. A missing target is created
// by renaming tmp. An existing empty target receives each top-level entry of
// tmp, after which tmp is removed; an entry that already exists in the
// target stops the move. It returns the paths it moved into an existing
// target, so they can be removed if this or a later step fails.
func place(ctx context.Context, tmp, target string, exists bool) ([]string, error) {
	if !exists {
		if err := rename(ctx, tmp, target); err != nil {
			return nil, fmt.Errorf("move project into place: %w", err)
		}
		return nil, nil
	}

	entries, err := os.ReadDir(tmp)
	if err != nil {
		return nil, fmt.Errorf("read temporary directory: %w", err)
	}
	var moved []string
	for _, e := range entries {
		dst := filepath.Join(target, e.Name())
		if _, err := os.Lstat(dst); err == nil {
			return moved, fmt.Errorf("move %s into place: %s already exists; the target changed during generation", e.Name(), dst)
		} else if !errors.Is(err, fs.ErrNotExist) {
			return moved, fmt.Errorf("move %s into place: %w", e.Name(), err)
		}
		if err := rename(ctx, filepath.Join(tmp, e.Name()), dst); err != nil {
			return moved, fmt.Errorf("move %s into place: %w", e.Name(), err)
		}
		moved = append(moved, dst)
	}
	if err := os.Remove(tmp); err != nil {
		return moved, fmt.Errorf("remove temporary directory: %w", err)
	}
	return moved, nil
}

// rename renames oldpath to newpath, retrying transient errors.
func rename(ctx context.Context, oldpath, newpath string) error {
	err := retry(ctx, func() error { return os.Rename(oldpath, newpath) }, isTransient, retryDelays())
	if errors.Is(err, syscall.EXDEV) {
		return fmt.Errorf("%w: the temporary directory and the target are on different file systems", err)
	}
	return err
}

// retry calls op until it succeeds, returns an error retryable rejects, the
// delays are used up, or ctx is done.
func retry(ctx context.Context, op func() error, retryable func(error) bool, delays []time.Duration) error {
	err := op()
	for _, d := range delays {
		if err == nil || !retryable(err) {
			return err
		}
		t := time.NewTimer(d)
		select {
		case <-ctx.Done():
			t.Stop()
			return errors.Join(err, ctx.Err())
		case <-t.C:
		}
		err = op()
	}
	return err
}

// cleanup removes the temporary directory and every entry already moved
// into a pre-existing target. The target directory itself is never removed.
// It also runs after cancellation, so it does not use the caller's context.
func cleanup(tmp string, moved []string) error {
	var errs []error
	for _, p := range append([]string{tmp}, moved...) {
		err := retry(context.Background(), func() error { return os.RemoveAll(p) }, isTransient, retryDelays())
		if err != nil {
			errs = append(errs, fmt.Errorf("%w: remove %s: %w", ErrCleanup, p, err))
		}
	}
	return errors.Join(errs...)
}
