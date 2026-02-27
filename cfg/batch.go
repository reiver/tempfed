package cfg

import (
	"time"

	"codeberg.org/reiver/go-env"
)

func BatchSize() int {
	return env.GetElse[int]("BATCH_SIZE", 50)
}

func BatchFlushInterval() time.Duration {
	seconds := env.GetElse[int]("BATCH_FLUSH_INTERVAL", 3)
	return time.Duration(seconds) * time.Second
}
