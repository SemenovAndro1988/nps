//go:build !windows
// +build !windows

package machineid

import "errors"

// readWindows is a stub on non-Windows platforms; the runtime
// dispatcher in machineid.go never calls it on those OSes, but the
// symbol must exist for the package to compile.
func readWindows() (string, error) {
	return "", errors.New("not supported on this platform")
}
