package main

import (
	"fmt"
	"time"
)

var SummaryInterval = 10 * time.Second

func main() {
	fmt.Printf("Preparing switch statistics...")
	statsReady := make(chan StatsList)
	go Gather(statsReady)
	stats := <-statsReady
	fmt.Printf("Switch statistics are ready")

	// TODO: Serve prometheus metrics from `stats`

	loopTimer := time.NewTimer(SummaryInterval)
	for {
		fmt.Printf("--- %s ---\n", time.Now().Format(time.RFC3339))
		for _, stat := range stats {
			fmt.Printf("%s (%s) locked to %s offset %d\n",
				stat.Info.Ip, stat.Info.Hostname, stat.LastPtpStatus.GmId, stat.LastPtpStatus.Offset)
			fmt.Printf("  %s %s running %s\n", stat.Info.Vendor, stat.Info.Model, stat.Info.SwVersion)
			fmt.Printf("  Fetch duration: %+v\n", stat.LastDuration)
			fmt.Printf("  Lock reliability: %d/%d = %.1f%%\n", stat.LockCount, stat.PollCount, pct(stat.LockCount, stat.PollCount))
			for gm, c := range stat.GmLockCount {
				fmt.Printf("    %s %d/%d = %.1f%%\n", gm, c, stat.LockCount, pct(c, stat.LockCount))
			}
			offsetHeader, offsetValues, offsetMeans := stat.Offsets.Render()
			fmt.Printf("  %s\n  %s\n  %s\n", offsetHeader, offsetValues, offsetMeans)
		}
		<-loopTimer.C
		loopTimer.Reset(SummaryInterval)
	}
}
