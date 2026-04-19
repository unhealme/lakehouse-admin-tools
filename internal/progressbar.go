package internal

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/pterm/pterm"
)

func NewProgressBar() *pterm.ProgressbarPrinter {
	base := pterm.DefaultProgressbar
	prog := &base
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		status := <-c
		prog.Stop()
		os.Exit(int(status.(syscall.Signal)))
	}()
	return prog
}
