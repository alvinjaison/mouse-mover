// Package winapi provides CGO-free Win32 cursor and display bindings.
// On non-Windows platforms the stub implementations return ErrUnsupported.
package winapi

import "errors"

// ErrUnsupported is returned by every stub function on non-Windows platforms.
var ErrUnsupported = errors.New("winapi: not supported on this platform")
