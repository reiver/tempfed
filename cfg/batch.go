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

func BatchTimeout() time.Duration {
	seconds := env.GetElse[int]("BATCH_TIMEOUT", 10)
	return time.Duration(seconds) * time.Second
}

func StatsLogInterval() time.Duration {
	seconds := env.GetElse[int]("STATS_LOG_INTERVAL", 30)
	return time.Duration(seconds) * time.Second
}
