package main

import (
	"context"
)

func www(ctx context.Context) <-chan struct{} {
	done := make(chan struct{})

	go func() {
		defer close(done)

		<-ctx.Done()
	}()

	return done
}
