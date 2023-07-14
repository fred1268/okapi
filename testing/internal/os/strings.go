package os

import (
	"os"
	"strings"
)

func SubstituteString(value string) string {
	var result string
	for {
		start := strings.Index(value, "${env:")
		if start != -1 {
			end := strings.Index(value, "}")
			if end != -1 {
				result = value[0:start]
				start += 6
				result += os.Getenv(value[start:end])
				end++
				value = value[end:]
				continue
			}
			return value
		}
		result += value
		break
	}
	return result
}
