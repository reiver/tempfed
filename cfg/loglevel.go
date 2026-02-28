package cfg

import (
	"codeberg.org/reiver/go-env"
)

func LogLevel() string {
	const defaultLogLevel string = "vvv"

	value := env.GetElse[string]("LOG_LEVEL", defaultLogLevel)
	switch value {
	case "v","vv","vvv","vvvv","vvvvv","vvvvvv","vvvvvvv":
		// nothing here
	default:
		value = defaultLogLevel
	}

	return value
}
