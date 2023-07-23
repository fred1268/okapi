package json

import (
	"testing"
)

func TestCompareJSON(t *testing.T) {
	tests := []struct {
		name   string
		src    string
		dst    string
		result error
	}{
		{
			name:   "empty",
			src:    "{}",
			dst:    "{}",
			result: nil,
		},
		{
			name:   "identical",
			src:    "{\"id\":\"10\",\"field1\":\"value1\"}",
			dst:    "{\"id\":\"10\",\"field1\":\"value1\"}",
			result: nil,
		},
		{
			name:   "larger",
			src:    "{\"id\":\"10\",\"field1\":\"value1\"}",
			dst:    "{\"id\":\"10\",\"field1\":\"value1\",\"field2\":\"value2\"}",
			result: nil,
		},
		{
			name:   "smaller",
			src:    "{\"id\":\"10\",\"field1\":\"value1\",\"field2\":\"value2\",\"field3\":\"value3\"}",
			dst:    "{\"id\":\"10\",\"field1\":\"value1\",\"field2\":\"value2\"}",
			result: ErrJSONMismatched,
		},
		{
			name:   "embedded but identical",
			src:    "{\"id\":\"10\",\"field1\":\"value1\",\"field2\":{\"field21\":\"value2\"}}",
			dst:    "{\"id\":\"10\",\"field1\":\"value1\",\"field2\":{\"field21\":\"value2\"}}",
			result: nil,
		},
		{
			name:   "embedded but different",
			src:    "{\"id\":10,\"field1\":\"value1\",\"field2\":{\"field21\":\"value2\"}}",
			dst:    "{\"id\":10,\"field1\":\"value1\",\"field2\":{\"field21\":\"value3\"}}",
			result: ErrJSONMismatched,
		},
	}
	for _, tt := range tests {
		name := tt.name
		src := tt.src
		dst := tt.dst
		res := tt.result
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			err := CompareJSONStrings(src, dst)
			if err != res {
				t.Errorf("wanted: '%s', got '%s'", dst, src)
			}
		})
	}
}
