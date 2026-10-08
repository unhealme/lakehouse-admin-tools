package utils

import "github.com/pterm/pterm"

type ProgressBar struct {
	*pterm.ProgressbarPrinter

	stopId uint32
}

func (p ProgressBar) Start() (*ProgressBar, error) {
	newStart := !p.IsActive
	pp, err := p.ProgressbarPrinter.Start()
	if err == nil && newStart {
		p.stopId = AtInterrupt(func() { p.Stop() })
	}
	p.ProgressbarPrinter = pp
	return &p, err
}

func (p *ProgressBar) Stop() (*ProgressBar, error) {
	if p.IsActive {
		UnInterrupt(p.stopId)
	}
	_, err := p.ProgressbarPrinter.Stop()
	return p, err
}

func NewProgressBar(p *pterm.ProgressbarPrinter) ProgressBar {
	return ProgressBar{ProgressbarPrinter: p}
}
