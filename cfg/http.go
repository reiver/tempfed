package cfg

import (
	"codeberg.org/reiver/go-env"
)

func HTTPTcpPort() uint16 {
	return env.GetElse[uint16]("PORT", 8080)
}
