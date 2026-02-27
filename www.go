package main

import (
	"context"
	"fmt"
	"net/http"

	"tempfed/cfg"
	"tempfed/srv/log"

	_ "tempfed/www"

	httpsrv "tempfed/srv/http"
)

func www(ctx context.Context) <-chan struct{} {
	log := logsrv.Begin()
	defer log.End()

	done := make(chan struct{})

	var addr string = fmt.Sprintf(":%d", cfg.HTTPTcpPort())

	server := &http.Server{
		Addr:    addr,
		Handler: &httpsrv.Mux,
	}

	go func() {
		defer close(done)
		log.Informf("😈 www spawned, listening to TCP-address %q", server.Addr)
		err := server.ListenAndServe()
		if nil != err && http.ErrServerClosed != err {
			log.Errorf("💀 www server error: %s", err)
		}
		log.Informf("😵 www died, that was listening to TCP-address %q", server.Addr)
	}()

	go func() {
		<-ctx.Done()

		log.Informf("👼 www killed, that was listening to TCP-address %q", server.Addr)
		err := server.Shutdown(context.Background())
		if nil != err {
			log.Errorf("www server shutdown error: %s", err)
		}
	}()

	return done
}
