package adapter

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/meshery/schemas/models/v1beta3/component"
)

func TestLegacyFormatCompatibility(t *testing.T) {
	original, err := os.ReadFile(filepath.Join(
		"testdata", testLinkerdName, testLinkerdCoreFilename,
	))
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name          string
		format        string
		omitFormat    bool
		invalidSchema bool
		wantError     bool
	}{
		{
			name:   "explicit JSON format",
			format: string(component.JSON),
		},
		{
			name:       "missing format",
			omitFormat: true,
		},
		{
			name: "empty format",
		},
		{
			name:      "unsupported format",
			format:    "YAML",
			wantError: true,
		},
		{
			name:          "missing format with invalid schema",
			omitFormat:    true,
			invalidSchema: true,
			wantError:     true,
		},
		{
			name:          "JSON format with invalid schema",
			format:        string(component.JSON),
			invalidSchema: true,
			wantError:     true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var legacy map[string]interface{}

			if err := json.Unmarshal(original, &legacy); err != nil {
				t.Fatal(err)
			}

			if tc.omitFormat {
				delete(legacy, "format")
			} else {
				legacy["format"] = tc.format
			}

			if tc.invalidSchema {
				legacy["schema"] = testInvalidJSON
			}

			input, err := json.Marshal(legacy)
			if err != nil {
				t.Fatal(err)
			}

			data, err := prepareComponentForRegistration(input)

			if tc.wantError {
				if err == nil {
					t.Fatal("expected legacy validation error")
				}
				return
			}

			if err != nil {
				t.Fatal(err)
			}

			assertLegacyFormatCompatibility(t, data, legacy)
		})
	}
}

func assertLegacyFormatCompatibility(
	t *testing.T,
	data []byte,
	legacy map[string]interface{},
) {
	t.Helper()

	var definition component.ComponentDefinition

	if err := json.Unmarshal(data, &definition); err != nil {
		t.Fatal(err)
	}

	if definition.Format != component.JSON {
		t.Fatalf("unexpected format: %q", definition.Format)
	}

	if definition.Component.Schema != legacy["schema"] {
		t.Fatal("original component schema was modified")
	}

	if definition.Component.Kind != legacy["kind"] {
		t.Fatal("original component kind was modified")
	}

	if definition.Component.Version != legacy["apiVersion"] {
		t.Fatal("original API version was modified")
	}
}
