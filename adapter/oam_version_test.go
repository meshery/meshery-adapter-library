package adapter

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestModifyMeshmodelVersion(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "legacy",
			input: `{"kind":"ServiceProfile","model":{"version":"1.0.0"},"custom":"preserved"}`,
			want:  `{"kind":"ServiceProfile","model":{"version":"2.0.0"},"custom":"preserved"}`,
		},
		{
			name:  "modern",
			input: `{"component":{"kind":"ServiceProfile"},"model":{"version":"1.0.0","model":{"version":"1.2.0"}},"custom":"preserved"}`,
			want:  `{"component":{"kind":"ServiceProfile"},"model":{"version":"1.0.0","model":{"version":"2.0.0"}},"custom":"preserved"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := modifyMeshmodelVersionInDefinition(
				[]byte(tt.input), "2.0.0",
			)
			if err != nil {
				t.Fatal(err)
			}

			var actual, expected interface{}
			if err := json.Unmarshal(got, &actual); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal([]byte(tt.want), &expected); err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(actual, expected) {
				t.Fatalf("got %s, want %s", got, tt.want)
			}
		})
	}
}

func TestModifyMeshmodelVersionInvalidDefinition(t *testing.T) {
	for _, input := range []string{
		`invalid`,
		`{}`,
		`{"model":null}`,
		`{"model":{"model":null}}`,
	} {
		if _, err := modifyMeshmodelVersionInDefinition(
			[]byte(input), "2.0.0",
		); err == nil {
			t.Fatalf("expected error for %s", input)
		}
	}
}
