package httpx

import (
	"log"
	"os"
)

// Logger defines the interface for logging. It is basically a thin wrapper
// around the standard logger which implements only a subset of the logger API.
type Logger interface {
	Printf(format string, v ...any)
}

// DefaultLogger is the default logger used by HTTPX when Client.Debug is true.
func DefaultLogger() *log.Logger {
	return log.New(os.Stderr, "", log.LstdFlags)
}
