package logsrv

import (
	"codeberg.org/reiver/go-log"
)

func Begin(fields ...log.Field) log.Logger {
	return logger.Begin(fields...)
}
