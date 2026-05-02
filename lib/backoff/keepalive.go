package libbackoff

import (
	"context"
	"math/rand/v2"
	"time"

	"codeberg.org/reiver/go-field"
	"codeberg.org/reiver/go-log"
)

func KeepAlive(ctx context.Context, logger log.Logger, name string, fn func()) {
	log := logger.Begin()
	defer log.End()


	const baseDelay  = 1 * time.Second
	const maxDelay   = 5 * time.Minute
	const multiplier = 2.0
	const jitter     = 0.25
	const resetAfter = 1 * time.Minute

	delay := baseDelay

	for {
		started := time.Now()

		fn()

		if nil != ctx.Err() {
			return
		}

		elapsed := time.Since(started)

		if resetAfter <= elapsed {
			delay = baseDelay
		}

		actualDelay := time.Duration(float64(delay) * (1.0 + jitter*(2.0*rand.Float64()-1.0)))

		log.Warn(
			field.String("", "backoff keep-alive stopped, retrying"),
			field.String("name", name),
			field.String("delay", actualDelay.String()),
		)

		select {
		case <-ctx.Done():
			return
		case <-time.After(actualDelay):
		}

		delay = time.Duration(float64(delay) * multiplier)
		if delay > maxDelay {
			delay = maxDelay
		}
	}
}
