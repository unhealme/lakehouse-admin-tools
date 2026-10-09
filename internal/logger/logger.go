package logger

import (
	"os"

	"github.com/pterm/pterm"
	"go.uber.org/atomic"
)

var logger = atomic.NewPointer(
	pterm.DefaultLogger.
		WithCaller(true).
		WithCallerOffset(1).
		WithLevel(pterm.LogLevelInfo).
		WithMaxWidth(200).
		WithWriter(os.Stderr),
)

func Logger() *pterm.Logger {
	return logger.Load()
}

func SetLogger(log *pterm.Logger) {
	logger.Store(log)
}

func Args(args ...any) []pterm.LoggerArgument {
	return logger.Load().Args(args...)
}

func Trace(msg string, args ...[]pterm.LoggerArgument) {
	logger.Load().Trace(msg, args...)
}

func Debug(msg string, args ...[]pterm.LoggerArgument) {
	logger.Load().Debug(msg, args...)
}

func Info(msg string, args ...[]pterm.LoggerArgument) {
	logger.Load().Info(msg, args...)
}

func Warn(msg string, args ...[]pterm.LoggerArgument) {
	logger.Load().Warn(msg, args...)
}

func Error(msg string, args ...[]pterm.LoggerArgument) {
	logger.Load().Error(msg, args...)
}

func Fatal(msg string, args ...[]pterm.LoggerArgument) {
	logger.Load().Fatal(msg, args...)
}

func Print(msg string, args ...[]pterm.LoggerArgument) {
	logger.Load().Print(msg, args...)
}
