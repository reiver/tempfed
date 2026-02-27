package cfg

import (
	"time"

	"codeberg.org/reiver/go-env"
)

func StatsLogInterval() time.Duration {
	seconds := env.GetElse[int]("STATS_LOG_INTERVAL", 30)
	return time.Duration(seconds) * time.Second
}
