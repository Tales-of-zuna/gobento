//go:build !windows

package generate

// isTransient reports whether err may go away on its own. Outside Windows,
// file operations do not fail because another process holds a file open.
func isTransient(error) bool {
	return false
}
