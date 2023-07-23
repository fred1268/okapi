package os

import (
	"fmt"
	"os"
	"strconv"
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
	var index int
	n = strings.Index(key, "[")
	if n != -1 {
		i, err := strconv.ParseInt(key[n+1:len(key)-1], 10, 64)
		if err != nil {
			return key
		}
		index = int(i)
		key = key[:n]
	}
	v, ok := captures[key]
	if !ok {
		return key
	}
	switch value := v.(type) {
	case float64:
		return fmt.Sprintf("%0.0f", value)
	case string:
		return value
	case map[string]any:
		return findValue(nextKey, value)
	case []any:
		if index >= len(value) {
			return "array index out of bounds"
		}
		switch element := value[index].(type) {
		case float64:
			return fmt.Sprintf("%0.0f", element)
		case string:
			return element
		case map[string]any:
			return findValue(nextKey, element)
		}
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
