package json

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
)

var ErrJSONMismatched error = errors.New("json mismatched")

func compareSlices(ctx context.Context, src, dst []any) error {
	for n, value := range src {
		dstValue := dst[n]
		if reflect.TypeOf(value) != reflect.TypeOf(dstValue) {
			return ErrJSONMismatched
		}
		switch value.(type) {
		case nil:
		case map[string]any:
			if err := compareMaps(ctx, value.(map[string]any), dstValue.(map[string]any)); err != nil {
				return err
			}
		case []any:
			if err := compareSlices(ctx, value.([]any), dstValue.([]any)); err != nil {
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

func compareMaps(ctx context.Context, src, dst map[string]any) error {
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
			if err := compareMaps(ctx, value.(map[string]any), dstValue.(map[string]any)); err != nil {
				return err
			}
		case []any:
			if err := compareSlices(ctx, value.([]any), dstValue.([]any)); err != nil {
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

func CompareJSONStrings(ctx context.Context, wanted, got string) error {
	// perfectly identical
	if wanted == "" || got == wanted {
		return nil
	}
	// one is not a json object, compare strings
	if !strings.Contains(wanted, "{") || !strings.Contains(got, "{") {
		if wanted != got {
			return ErrJSONMismatched
		}
		return nil
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
			return compareMaps(ctx, wantedMap, gotMap)
		}
	}
	return ErrJSONMismatched
}
