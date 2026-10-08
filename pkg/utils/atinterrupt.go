package utils

import (
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/atomic"
)

var (
	hooks  = make(map[uint32]func())
	hookId atomic.Uint32

	listenSig atomic.Pointer[chan os.Signal]
)

func AtInterrupt(hook func()) uint32 {
	i := hookId.Inc()
	hooks[i] = hook
	waitSignal()
	return i
}

func UnInterrupt(id uint32) {
	delete(hooks, id)
	if len(hooks) < 1 {
		if sig := listenSig.Swap(nil); sig != nil {
			close(*sig)
		}
	}
}

func waitSignal() {
	if sig := make(chan os.Signal, 1); listenSig.CompareAndSwap(nil, &sig) {
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		go execHooks(sig)
	}
}

func execHooks(sig chan os.Signal) {
	if s, ok := <-sig; ok {
		for _, h := range hooks {
			h()
		}
		os.Exit(int(s.(syscall.Signal)))
	}
	signal.Stop(sig)
}
