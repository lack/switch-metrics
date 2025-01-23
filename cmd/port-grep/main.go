package main

import (
	"fmt"
	"sync"

	"github.com/lack/switch-metrics/pkg/restconf"
)

type SwitchResult struct {
	Ip     string
	Ifaces []restconf.Iface
	err    error
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
	for _, s := range swlist {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ifaces, err := s.Interfaces()
			resultChan <- SwitchResult{
				Ip:     s.Ip,
				Ifaces: ifaces,
				err:    err,
			}
		}()
	}
	wg.Wait()
	close(resultChan)
	for info := range resultChan {
		if info.err != nil {
			panic(info.err)
		}
		for _, iface := range info.Ifaces {
			if iface.Desc == "" {
				continue
			}
			fmt.Printf("%s %-20s %s\n", info.Ip, iface.Name, iface.Desc)
		}
	}
}
