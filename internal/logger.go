package internal

import (
	"os"
	"sync"

	"github.com/pterm/pterm"
)

const (
	LogLevelDisabled = pterm.LogLevelDisabled
	LogLevelTrace    = pterm.LogLevelTrace
	LogLevelDebug    = pterm.LogLevelDebug
	LogLevelInfo     = pterm.LogLevelInfo
	LogLevelWarn     = pterm.LogLevelWarn
	LogLevelError    = pterm.LogLevelError
	LogLevelFatal    = pterm.LogLevelFatal
	LogLevelPrint    = pterm.LogLevelPrint
)

type Logger = pterm.Logger

var DefaultLogger = sync.OnceValue(func() *Logger { return pterm.DefaultLogger.WithLevel(LogLevelInfo).WithWriter(os.Stderr) })
