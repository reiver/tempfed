package main

import (
	"context"
	"fmt"
	"sync"

	"tempfed/cfg"
	"tempfed/lib/backoff"
	"tempfed/lib/mstdn"
	"tempfed/srv/db"
	"tempfed/srv/log"
)

func mstdn(ctx context.Context) <-chan struct{} {
	log := logsrv.Begin()
	defer log.End()

	hosts := cfg.MstdnHosts()
	batchSize := cfg.BatchSize()
	flushInterval := cfg.BatchFlushInterval()
	statsInterval := cfg.StatsLogInterval()

	done := make(chan struct{})
	var wg sync.WaitGroup

	for _, host := range hosts {
		wg.Add(1)
		go func(host string) {
			defer wg.Done()

			libbackoff.KeepAlive(ctx, log, fmt.Sprintf("mstdn-compatible server %s", host), func(){
				libmstdn.Accept(ctx, log, host, dbsrv.Conn, batchSize, flushInterval, statsInterval)
			})
		}(host)
	}

	go func() {
		wg.Wait()
		close(done)
	}()

	return done
}
