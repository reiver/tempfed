package logsrv

import (
	"io"
	"os"

	"codeberg.org/reiver/go-log"

	"tempfed/cfg"
)

var writer io.Writer = os.Stdout

var logger log.Logger = log.CreateLogger(writer, cfg.LogLevel())
