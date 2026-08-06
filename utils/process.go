package utils

import (
	"time"

	"github.com/shirou/gopsutil/v4/process"
)

func SoftKill(process *process.Process) (e error) {
	if e = process.Terminate(); e != nil {
		return
	}
	timeout := 0
	for range time.Tick(time.Second) {
		if running, _ := process.IsRunning(); !running {
			break
		}
		if timeout >= 15 {
			e = process.Kill()
			break
		}
		timeout++
	}
	return
}
