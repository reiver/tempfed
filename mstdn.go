package main

import (
	"context"
	"fmt"
	"sync"

	"tempfed/cfg"
	"tempfed/lib/backoff"
	"tempfed/lib/mstdn"
	"tempfed/srv/log"
)

func mstdn(ctx context.Context) {
	log := logsrv.Begin()
	defer log.End()

	hosts := cfg.MstdnHosts()

	var wg sync.WaitGroup

	for _, host := range hosts {
		wg.Add(1)
		go func(host string) {
			defer wg.Done()

			libbackoff.KeepAlive(ctx, log, fmt.Sprintf("mstdn-compatible server %s", host), func(){
				libmstdn.Accept(ctx, log, host)
			})
		}(host)
	}

	wg.Wait()
}
