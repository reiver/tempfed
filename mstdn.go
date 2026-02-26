package main

import (
	"sync"

	"tempfed/cfg"
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
			libmstdn.Accept(log, host)
		}(host)
	}

	wg.Wait()
}
