package os

import (
	"fmt"
	"os"
	"strings"
)

func SubstituteEnvironmentVariable(value string) string {
	var result string
	for {
		start := strings.Index(value, "${env:")
		if start != -1 {
			end := strings.Index(value, "}")
			if end != -1 {
				result += fmt.Sprintf("%s%s", value[0:start], os.Getenv(value[start+6:end]))
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

func findValue(key string, captures map[string]any) string {
	n := strings.Index(key, ".")
	nextKey := ""
	if n != -1 {
		nextKey = key[n+1:]
		key = key[:n]
	}
	v, ok := captures[key]
	if !ok {
		return key
	}
	switch value := v.(type) {
	case float32, float64:
		return fmt.Sprintf("%0.0f", value)
	case string:
		return value
	case map[string]any:
		return findValue(nextKey, value)
	case []any:
		return key
	}
	return key
}

func SubstituteCapturedVariable(value string, captures map[string]any) string {
	var result string
	for {
		start := strings.Index(value, "${")
		if start != -1 {
			end := strings.Index(value, "}")
			if end != -1 {
				result += fmt.Sprintf("%s%s", value[0:start], findValue(value[start+2:end], captures))
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
