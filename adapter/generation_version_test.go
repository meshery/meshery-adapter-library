package adapter

import (
	"encoding/json"
	"testing"
)

func TestGeneratedComponentDefinitionVersion(t *testing.T) {
	tests := []struct {
		name     string
		generate func() ([]byte, error)
	}{
		{
			name: "native component",
			generate: func() ([]byte, error) {
				return GenerateNativeComponent(pipelineTestCRD)
			},
		},
		{
			name: "model-enriched component",
			generate: func() ([]byte, error) {
				return GenerateNativeComponentWithModel(
					pipelineTestCRD,
					testLinkerdName,
					testLinkerdRelease,
					MeshModelConfig{
						Category: testModelCategory,
					},
				)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.generate()
			if err != nil {
				t.Fatal(err)
			}

			var definition struct {
				Version   string `json:"version"`
				Component struct {
					Version string `json:"version"`
				} `json:"component"`
			}

			if err := json.Unmarshal(data, &definition); err != nil {
				t.Fatal(err)
			}

			if definition.Version != "1.0.0" {
				t.Fatalf("unexpected definition version: %q",
					definition.Version)
			}

			if definition.Component.Version != testServiceProfileAPIVersion {
				t.Fatalf("unexpected API version: %q",
					definition.Component.Version)
			}
		})
	}
}
