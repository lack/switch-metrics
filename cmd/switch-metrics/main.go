package main

import (
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/lack/switch-metrics/pkg/restconf"
)

type SwitchResult struct {
	Idx       int
	Info      restconf.SwitchInfo
	PtpStatus restconf.PtpStatus
	err       error
}

type Histogram struct {
	Buckets []int
	Data    []int
}

func (h *Histogram) init() {
	if h.Buckets[len(h.Buckets)-1] != math.MaxInt {
		h.Buckets = append(h.Buckets, math.MaxInt)
	}
	h.Data = make([]int, len(h.Buckets))
}

func (h *Histogram) Add(value int) {
	if len(h.Data) == 0 {
		h.init()
	}
	for i, boundary := range h.Buckets {
		if value < boundary {
			h.Data[i] = h.Data[i] + 1
			return
		}
	}
}

func (h *Histogram) BucketName(i int) string {
	if i == 0 {
		return fmt.Sprintf("<%d", h.Buckets[i])
	}
	if i == len(h.Buckets)-1 {
		return fmt.Sprintf(">=%d", h.Buckets[i-1])
	}
	return fmt.Sprintf("[%d %d)", h.Buckets[i-1], h.Buckets[i])
}

type SwitchStats struct {
	Info          restconf.SwitchInfo
	LastErr       error
	LastPtpStatus restconf.PtpStatus
	Offsets       Histogram
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

func main() {
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
			Offsets: Histogram{
				Buckets: OffsetBuckets,
			},
			GmLockCount: make(map[string]int),
		}
	}

	for {
		resultChan = make(chan SwitchResult, len(swlist))
		for i, s := range swlist {
			wg.Add(1)
			go func() {
				defer wg.Done()
				ptpstatus, err := s.GetPtpStatus()
				resultChan <- SwitchResult{
					Idx:       i,
					PtpStatus: ptpstatus,
					err:       err,
				}
			}()
		}
		wg.Wait()
		close(resultChan)
		for info := range resultChan {
			stat := stats[info.Idx]
			stat.PollCount += 1
			if info.err != nil {
				stat.LastErr = info.err
				stat.ErrCount += 1
			}
			stat.LastPtpStatus = info.PtpStatus
			if info.PtpStatus.SyncState == restconf.SYNC_STATE_LOCKED {
				stat.LockCount += 1
				stat.GmLockCount[info.PtpStatus.GmId] = stat.GmLockCount[info.PtpStatus.GmId] + 1
			}
			stat.Offsets.Add(info.PtpStatus.Offset)
			stats[info.Idx] = stat
		}

		// TODO: Publish Prometheus endpoints...

		fmt.Printf("--- %s ---\n", time.Now().Format(time.RFC3339))
		for _, stat := range stats {
			fmt.Printf("%s (%s) locked to %s offset %d\n",
				stat.Info.Ip, stat.Info.Hostname, stat.LastPtpStatus.GmId, stat.LastPtpStatus.Offset)
			fmt.Printf("  %s %s running %s\n", stat.Info.Vendor, stat.Info.Model, stat.Info.SwVersion)
			fmt.Printf("  Lock reliability: %d/%d = %.1f%%\n", stat.LockCount, stat.PollCount, pct(stat.LockCount, stat.PollCount))
			for gm, c := range stat.GmLockCount {
				fmt.Printf("    %s %d/%d = %.1f%%\n", gm, c, stat.LockCount, pct(c, stat.LockCount))
			}
			offsetHeader := strings.Builder{}
			offsetValues := strings.Builder{}
			for i, v := range stat.Offsets.Data {
				header := stat.Offsets.BucketName(i)
				offsetHeader.WriteString(header)
				offsetHeader.WriteRune(' ')
				offsetValues.WriteString(fmt.Sprintf("%*d ", len(header), v))
			}
			fmt.Printf("  %s\n  %s\n", offsetHeader.String(), offsetValues.String())
		}
	}
}
