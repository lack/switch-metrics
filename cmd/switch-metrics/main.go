package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var SummaryInterval = 10 * time.Second

var (
	commonSwitchLabels = []string{
		"hostname",
		"ip",
		"vendor",
		"model",
		"swversion",
	}
	offsetGauge = prometheus.NewDesc(
		"switchmetrics_ptp_offset_ns",
		"PTP offset compared to GrandMaster in ns",
		commonSwitchLabels, nil,
	)
	unlockedCounter = prometheus.NewDesc(
		"switchmetrics_ptp_unlock_events",
		"A counter of events when the PTP time was not \"locked\"",
		commonSwitchLabels, nil,
	)
)

func (s *SwitchStats) LabelValues() []string {
	return []string{
		s.Info.Hostname,
		s.Info.Ip,
		s.Info.Vendor,
		s.Info.Model,
		s.Info.SwVersion,
	}
}

func (sl StatsList) Collect(ch chan<- prometheus.Metric) {
	for _, s := range sl {
		ch <- prometheus.MustNewConstMetric(
			offsetGauge, prometheus.GaugeValue,
			float64(s.LastPtpStatus.Offset),
			s.LabelValues()...)
		ch <- prometheus.MustNewConstMetric(
			unlockedCounter, prometheus.CounterValue,
			float64(s.PollCount-s.LockCount),
			s.LabelValues()...)
	}
}

func (sl StatsList) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(sl, ch)
}

func main() {
	reg := prometheus.NewPedanticRegistry()

	fmt.Printf("Preparing switch statistics...\n")
	statsReady := make(chan StatsList)
	go Gather(statsReady)
	stats := <-statsReady
	fmt.Printf("Switch statistics are ready\n")

	reg.MustRegister(
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		collectors.NewGoCollector(),
		stats,
	)

	http.Handle("/metrics", promhttp.Handler())
	go http.ListenAndServe(":2121", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	fmt.Printf("Serving metrics at http://localhost:2121/metrics\n")

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
