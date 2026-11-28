package adapter

import (
	"strings"
	"testing"

	"github.com/layer5io/meshkit/utils/manifests"
)

func TestExtractModernComponentResources(t *testing.T) {
	manifest := `
apiVersion: v1
kind: ConfigMap
metadata:
  name: ignored
---
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: serviceprofiles.linkerd.io
---
`

	cfg := StaticCompConfig{
		Config: manifests.Config{
			ExtractCrds: func(data string) []string {
				return strings.Split(data, "---")
			},
		},
	}

	resources, err := extractModernComponentResources(manifest, cfg)
	if err != nil {
		t.Fatal(err)
	}

	if len(resources) != 1 {
		t.Fatalf("expected one CRD, got %d", len(resources))
	}

	if !strings.Contains(resources[0], "serviceprofiles.linkerd.io") {
		t.Fatal("expected ServiceProfile CRD")
	}

	if strings.Contains(resources[0], "ConfigMap") {
		t.Fatal("non-CRD resource was included")
	}
}

func TestExtractModernComponentResourcesInvalidInput(t *testing.T) {
	cfg := StaticCompConfig{
		Config: manifests.Config{
			ExtractCrds: func(data string) []string {
				return strings.Split(data, "---")
			},
		},
	}

	for _, input := range []string{
		"",
		nullResource,
		"kind: ConfigMap",
		"kind: [invalid",
	} {
		if _, err := extractModernComponentResources(input, cfg); err == nil {
			t.Errorf("expected error for input %q", input)
		}
	}

	if _, err := extractModernComponentResources(
		"kind: CustomResourceDefinition",
		StaticCompConfig{},
	); err == nil {
		t.Error("expected error for missing CRD extractor")
	}
}
