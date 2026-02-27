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

	log.Informf("tempfed ⚡")

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-mstdn(ctx)

	dbsrv.Conn.Close()
	log.Informf("tempfed 👻")
}
