package main

import (
	"sync"
	"time"

	"github.com/lack/switch-metrics/pkg/restconf"
)

type SwitchResult struct {
	Idx       int
	Info      restconf.SwitchInfo
	PtpStatus restconf.PtpStatus
	err       error
	duration  time.Duration
}

type SwitchStats struct {
	Info          restconf.SwitchInfo
	LastErr       error
	LastPtpStatus restconf.PtpStatus
	LastDuration  time.Duration
	Offsets       restconf.Histogram
	PollCount     int
	ErrCount      int
	LockCount     int
	GmLockCount   map[string]int
}

func pct(a, b int) float64 {
	return (float64(a) * 100.0) / float64(b)
}

var OffsetBuckets = []int{
	-200, -100, -75, -50, -25, 0, 25, 50, 75, 100, 200,
}

var PollInterval = 500 * time.Millisecond

type StatsList []SwitchStats

func Gather(ready chan<- StatsList) {
	var signaled bool
	switches, err := restconf.LoadSwitches()
	if err != nil {
		panic(err)
	}
	swlist := make([]restconf.DellSwitch, len(switches))
	for i, s := range switches {
		swlist[i] = restconf.DellSwitch{Switch: s}
	}
	resultChan := make(chan SwitchResult, len(swlist))
	wg := sync.WaitGroup{}
	for i, s := range swlist {
		wg.Add(1)
		go func() {
			defer wg.Done()
			info, err := s.Info()
			resultChan <- SwitchResult{
				Idx:  i,
				Info: info,
				err:  err,
			}
		}()
	}
	wg.Wait()
	close(resultChan)
	stats := make([]SwitchStats, len(swlist))
	for info := range resultChan {
		if info.err != nil {
			panic(info.err)
		}
		stats[info.Idx] = SwitchStats{
			Info: info.Info,
			Offsets: restconf.Histogram{
				Buckets: OffsetBuckets,
			},
			GmLockCount: make(map[string]int),
		}
	}

	pollTimer := time.NewTimer(PollInterval)
	for {
		resultChan = make(chan SwitchResult, len(swlist))
		for i, s := range swlist {
			wg.Add(1)
			go func() {
				defer wg.Done()
				start := time.Now()
				ptpstatus, err := s.GetPtpStatus()
				resultChan <- SwitchResult{
					Idx:       i,
					PtpStatus: ptpstatus,
					err:       err,
					duration:  time.Now().Sub(start),
				}
			}()
		}
		wg.Wait()
		close(resultChan)
		for info := range resultChan {
			stat := stats[info.Idx]
			stat.PollCount += 1
			stat.LastDuration = info.duration
			if info.err != nil {
				stat.LastErr = info.err
				stat.ErrCount += 1
			} else {
				stat.LastPtpStatus = info.PtpStatus
				if info.PtpStatus.SyncState == restconf.SYNC_STATE_LOCKED {
					stat.LockCount += 1
					stat.GmLockCount[info.PtpStatus.GmId] = stat.GmLockCount[info.PtpStatus.GmId] + 1
				}
				stat.Offsets.Add(info.PtpStatus.Offset)
			}
			stats[info.Idx] = stat
		}
		if !signaled {
			ready <- stats
			signaled = true
		}

		// Wait for poll timer to throttle us if needes
		<-pollTimer.C
		pollTimer.Reset(PollInterval)
	}
}
