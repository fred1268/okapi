package json

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
)

var ErrJSONMismatched error = errors.New("json mismatched")

func compareStrings(wanted, got string) error {
	realWanted := strings.Trim(wanted, "%")
	if strings.HasPrefix(wanted, "%") && strings.HasSuffix(wanted, "%") && !strings.Contains(got, realWanted) {
		return ErrJSONMismatched
	} else if strings.HasPrefix(wanted, "%") && !strings.HasPrefix(got, realWanted) {
		return ErrJSONMismatched
	} else if strings.HasSuffix(wanted, "%") && !strings.HasSuffix(got, realWanted) {
		return ErrJSONMismatched
	} else if wanted != got {
		return ErrJSONMismatched
	}
	return nil
}

func compareSlices(src, dst []any) error {
	found := 0
	for _, value := range src {
		for _, dstValue := range dst {
			if reflect.TypeOf(value) != reflect.TypeOf(dstValue) {
				continue
			}
			switch value.(type) {
			case nil:
			case map[string]any:
				if err := compareMaps(value.(map[string]any), dstValue.(map[string]any)); err != nil {
					continue
				}
			case []any:
				if err := compareSlices(value.([]any), dstValue.([]any)); err != nil {
					continue
				}
			default:
				if value != dstValue {
					continue
				}
			}
			found++
		}
	}
	if found != len(src) {
		return ErrJSONMismatched
	}
	return nil
}

func compareMaps(src, dst map[string]any) error {
	for key, value := range src {
		dstValue, found := dst[key]
		if !found {
			return ErrJSONMismatched
		}
		if reflect.TypeOf(value) != reflect.TypeOf(dstValue) {
			return ErrJSONMismatched
		}
		switch value.(type) {
		case nil:
		case map[string]any:
			if err := compareMaps(value.(map[string]any), dstValue.(map[string]any)); err != nil {
				return err
			}
		case []any:
			if err := compareSlices(value.([]any), dstValue.([]any)); err != nil {
				return err
			}
		default:
			if value != dstValue {
				return ErrJSONMismatched
			}
		}
	}
	return nil
}

func CompareJSONStrings(wanted, got string) error {
	// perfectly identical
	if wanted == "" || got == wanted {
		return nil
	}
	// one is not a json object, compare strings
	if !strings.Contains(wanted, "{") || !strings.Contains(got, "{") {
		return compareStrings(wanted, got)
	}
	// both are json, compare json
	var g, w interface{}
	err := json.Unmarshal([]byte(strings.ToLower(got)), &g)
	if err != nil {
		return err
	}
	err = json.Unmarshal([]byte(strings.ToLower(wanted)), &w)
	if err != nil {
		return err
	}
	if wantedMap, ok := w.(map[string]any); ok {
		if gotMap, ok := g.(map[string]any); ok {
			return compareMaps(wantedMap, gotMap)
		}
	}
	return ErrJSONMismatched
}
