package libactors

import (
	"strings"
)

func IsValidUserName(value string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(value, prefix) {
			return true
		}
	}

	return false
}
