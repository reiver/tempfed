package cfg

import (
	"codeberg.org/reiver/go-env"
)

var defaultMstdnHosts = []string{"fedi.buzz"}

func MstdnHosts() []string {
	return env.GetElse[[]string]("MSTDN_HOSTS", defaultMstdnHosts)
}
