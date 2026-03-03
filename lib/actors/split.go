package libactors

import (
	"strings"
)

func Split(value string) (string, string, bool) {
	for _, prefix := range prefixes {
		if strings.HasPrefix(value, prefix) {
			var (
				service   string = prefix[:len(prefix)-1]
				parameter string = value[len(prefix):]
			)

			return service, parameter, true
		}
	}

	return "", "", false
}
