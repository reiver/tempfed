package main

import (
	"context"
	"os/signal"
	"syscall"

	"tempfed/srv/db"
	"tempfed/srv/log"
)

func main() {
	log := logsrv.Begin()
	defer log.End()

	log.Highlightf("tempfed ⚡")

	sigCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	ctx, cancel := context.WithCancel(sigCtx)
	defer cancel()

	mstdnDone := mstdn(ctx)
	wwwDone := www(ctx)

	select {
	case <-mstdnDone:
	case <-wwwDone:
	}

	cancel()

	<-mstdnDone
	<-wwwDone

//	dbsrv.Conn.Close()
	log.Highlightf("tempfed 👻")
}
