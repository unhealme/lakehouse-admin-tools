package cmd

import (
	"fmt"
	"maps"
	"os/user"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pterm/pterm"
	"github.com/shirou/gopsutil/v4/process"
	"github.com/unhealme/lakehouse-admin-tools/internal"
)

const PsAutoKillVersion = "2026.06.30-1"

func PsAutoKill(logger *pterm.Logger, args *PsAutoKillArgs) {
	logger.Debug("using auto kill args.", logger.Args(internal.ToArgs(*args)...))
	procs, err := process.Processes()
	if err != nil {
		logger.Fatal("unable to list processes", logger.Args("error", err))
	}
	memoryThreshold, err := internal.ParseSize(args.MemoryThreshold)
	if err != nil {
		logger.Fatal(err.Error())
	}
	excludeUsersArgs := strings.Split(args.ExcludeUsers, ",")
	excludeUsers := make(map[uint32]internal.EmptyType, len(excludeUsersArgs))
	for _, excludeUser := range excludeUsersArgs {
		var uid uint64
		if u, err := user.Lookup(excludeUser); err != nil {
			if uid, err = strconv.ParseUint(excludeUser, 10, 32); err != nil {
				continue
			}
		} else {
			if uid, err = strconv.ParseUint(u.Uid, 10, 32); err != nil {
				continue
			}
		}
		excludeUsers[uint32(uid)] = internal.Empty
	}
	logger.Debug(fmt.Sprintf("user to exclude: %v", slices.Collect(maps.Keys(excludeUsers))))
	var wg sync.WaitGroup
	for _, p := range procs {
		createTime, err := p.CreateTime()
		if err != nil {
			logger.Warn("unable to get process create time. skipping..", logger.Args("pid", p.Pid, "error", err))
			continue
		}
		memInfo, err := p.MemoryInfo()
		if err != nil {
			logger.Warn("unable to get process memory info. skipping..", logger.Args("pid", p.Pid, "error", err))
			continue
		}
		memPerc, err := p.MemoryPercent()
		if err != nil {
			logger.Warn("unable to get process memory percent. skipping..", logger.Args("pid", p.Pid, "error", err))
			continue
		}
		userIds, err := p.Uids()
		if err != nil {
			logger.Warn("unable to get process user. skipping..", logger.Args("pid", p.Pid, "error", err))
			continue
		}
		uidString := strconv.FormatInt(int64(userIds[0]), 10)
		var userName string
		if userInfo, err := user.LookupId(uidString); err == nil {
			userName = userInfo.Username
		} else {
			userName = uidString
		}
		cmdLine, err := p.Cmdline()
		if err != nil {
			logger.Warn("unable to get process cmdline. skipping..", logger.Args("pid", p.Pid, "error", err))
			continue
		}
		logArgs := logger.Args(
			"pid", p.Pid,
			"createTime", time.UnixMilli(createTime),
			"rss", internal.FormatSize(int64(memInfo.RSS)),
			"percMem", fmt.Sprintf("%2.1f%%", memPerc),
			"user", userName,
			"cmd", cmdLine,
		)
		if _, ok := excludeUsers[userIds[0]]; !ok &&
			(time.Now().UnixMilli()-createTime) > args.RuntimeThreshold.Milliseconds() &&
			memInfo.RSS > uint64(memoryThreshold) {
			wg.Go(func() {
				if !args.DryRun {
					if err := softKill(p); err != nil {
						logger.Warn("failed to kill process.", logArgs)
						return
					}
				}
				logger.Info("process killed.", logArgs)
			})
		}
	}
	wg.Wait()
}

func softKill(process *process.Process) error {
	if running, _ := process.IsRunning(); running {
		if err := process.Terminate(); err != nil {
			return err
		}
	} else {
		return nil
	}
	for i := 1; i <= 15; i++ {
		if running, _ := process.IsRunning(); !running {
			return nil
		}
		time.Sleep(1 * time.Second)
	}
	if running, _ := process.IsRunning(); running {
		if err := process.Kill(); err != nil {
			return err
		}
	}
	return nil
}
