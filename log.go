package httpx

import (
	"git.sr.ht/~jamesponddotco/xstd-go/xlog"
)

// Logger defines the interface for logging. It is basically a thin wrapper
// around the standard logger which implements only a subset of the logger API.
type Logger interface {
	Printf(format string, v ...any)
}

// DefaultLogger is the default logger used by HTTPX when Client.Debug is true.
func DefaultLogger() Logger {
	return xlog.DefaultZeroLogger
}
