package generate

import (
	"errors"
	"io/fs"
	"syscall"
)

// errorSharingViolation is ERROR_SHARING_VIOLATION: another process has the
// file open.
const errorSharingViolation = syscall.Errno(32)

// isTransient reports whether err may go away on its own: access denied or a
// sharing violation, typically while an antivirus scanner or indexer holds a
// newly created file.
func isTransient(err error) bool {
	return errors.Is(err, fs.ErrPermission) || errors.Is(err, errorSharingViolation)
}
