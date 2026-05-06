package internal

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/pterm/pterm"
)

func NewProgressBar() *pterm.ProgressbarPrinter {
	prog := pterm.DefaultProgressbar
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		status := <-c
		prog.Stop()
		os.Exit(int(status.(syscall.Signal)))
	}()
	return &prog
}
