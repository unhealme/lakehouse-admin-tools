package utils

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/pterm/pterm"
)

func NewProgressBar(ctx context.Context) *pterm.ProgressbarPrinter {
	prog := pterm.DefaultProgressbar
	go progressBarStopper(ctx, &prog)
	return &prog
}

func progressBarStopper(ctx context.Context, prog *pterm.ProgressbarPrinter) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	select {
	case status := <-sig:
		prog.Stop()
		os.Exit(int(status.(syscall.Signal)))
	case <-ctx.Done():
		prog.Stop()
		signal.Stop(sig)
		return
	}
}
