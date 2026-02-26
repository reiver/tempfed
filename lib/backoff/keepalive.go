package libbackoff

import (
	"math/rand/v2"
	"time"

	"codeberg.org/reiver/go-field/stringly"
	"codeberg.org/reiver/go-log"
)

func KeepAlive(logger log.Logger, name string, fn func()) {
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

		elapsed := time.Since(started)

		if resetAfter <= elapsed {
			delay = baseDelay
		}

		actualDelay := time.Duration(float64(delay) * (1.0 + jitter*(2.0*rand.Float64()-1.0)))

		log.Warn(
			stringly.String("", "backoff keep-alive stopped, retrying"),
			stringly.String("name", name),
			stringly.String("delay", actualDelay.String()),
		)

		time.Sleep(actualDelay)

		delay = time.Duration(float64(delay) * multiplier)
		if delay > maxDelay {
			delay = maxDelay
		}
	}
}
