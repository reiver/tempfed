package cfg

import (
	"codeberg.org/reiver/go-env"
)

func RelayPublicKeyFileName() string {
	return env.GetElse[string]("RELAY_PUBLIC_KEY_FILE_NAME", "")
}
