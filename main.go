package main

import (
	"tempfed/srv/log"
)

func main() {
	log := logsrv.Begin()
	defer log.End()

	log.Informf("tempfed ⚡")

	mstdn()
}
