package utils

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/pterm/pterm"
)

func NewProgressBar() *pterm.ProgressbarPrinter {
	prog := pterm.DefaultProgressbar
	go progressBarStopper(&prog)
	return &prog
}

func progressBarStopper(prog *pterm.ProgressbarPrinter) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	status := <-sig
	prog.Stop()
	os.Exit(int(status.(syscall.Signal)))
}
