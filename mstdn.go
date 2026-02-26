package main

import (
	"fmt"
	"sync"

	"tempfed/cfg"
	"tempfed/lib/backoff"
	"tempfed/lib/mstdn"
	"tempfed/srv/log"
)

func mstdn() {
	log := logsrv.Begin()
	defer log.End()

	hosts := cfg.MstdnHosts()

	var wg sync.WaitGroup

	for _, host := range hosts {
		wg.Add(1)
		go func(host string) {
			defer wg.Done()

			libbackoff.KeepAlive(log, fmt.Sprintf("mstdn-compatible server %s", host), func(){
				libmstdn.Accept(log, host)
			})
		}(host)
	}

	wg.Wait()
}
