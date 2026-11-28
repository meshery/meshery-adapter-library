package adapter

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/meshery/schemas/models/v1beta3/component"
)

const multiVersionReviewCRD = `
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: serviceprofiles.linkerd.io
spec:
  group: linkerd.io
  scope: Namespaced
  names:
    kind: ServiceProfile
    plural: serviceprofiles
    singular: serviceprofile
  versions:
    - name: v1alpha1
      served: false
      storage: false
      schema:
        openAPIV3Schema:
          type: object
          properties:
            obsoleteField:
              type: string
    - name: v1beta1
      served: true
      storage: false
      schema:
        openAPIV3Schema:
          type: object
          properties:
            betaField:
              type: string
    - name: v1
      served: true
      storage: true
      schema:
        openAPIV3Schema:
          type: object
          properties:
            stableField:
              type: integer
`

func TestServedCRDVersionNativeGeneration(t *testing.T) {
	data, err := GenerateNativeComponent(multiVersionReviewCRD)
	if err != nil {
		t.Fatal(err)
	}

	assertSelectedCRDVersion(
		t, data, "linkerd.io/v1", "stableField",
	)
}

func TestServedCRDVersionGenerationWithModel(t *testing.T) {
	data, err := GenerateNativeComponentWithModel(
		multiVersionReviewCRD,
		testLinkerdType,
		testLinkerdRelease,
		MeshModelConfig{Category: testModelCategory},
	)
	if err != nil {
		t.Fatal(err)
	}

	assertSelectedCRDVersion(
		t, data, "linkerd.io/v1", "stableField",
	)
}

func TestServedCRDVersionFallback(t *testing.T) {
	resource := strings.ReplaceAll(
		multiVersionReviewCRD,
		"storage: true",
		"storage: false",
	)

	data, err := GenerateNativeComponent(resource)
	if err != nil {
		t.Fatal(err)
	}

	assertSelectedCRDVersion(
		t, data, "linkerd.io/v1beta1", "betaField",
	)
}

func TestServedCRDVersionRejectsUnservedCRD(t *testing.T) {
	resource := strings.ReplaceAll(
		multiVersionReviewCRD,
		"served: true",
		"served: false",
	)

	_, err := GenerateNativeComponent(resource)
	if err == nil ||
		!strings.Contains(err.Error(), "no served versions") {
		t.Fatalf("expected no-served-version error: %v", err)
	}
}

func TestServedCRDVersionRejectsMissingSchema(t *testing.T) {
	resource := strings.Replace(
		multiVersionReviewCRD,
		"storage: true\n      schema:",
		"storage: true\n      unusedSchema:",
		1,
	)

	_, err := GenerateNativeComponent(resource)
	if err == nil ||
		!strings.Contains(err.Error(), "no OpenAPI schema") {
		t.Fatalf("expected missing-schema error: %v", err)
	}
}

func assertSelectedCRDVersion(
	t *testing.T,
	data []byte,
	version string,
	field string,
) {
	t.Helper()

	var definition component.ComponentDefinition
	if err := json.Unmarshal(data, &definition); err != nil {
		t.Fatal(err)
	}

	if definition.Component.Version != version {
		t.Fatalf(
			"expected version %s, got %s",
			version, definition.Component.Version,
		)
	}

	var schema struct {
		Properties map[string]json.RawMessage `json:"properties"`
	}

	if err := json.Unmarshal(
		[]byte(definition.Component.Schema), &schema,
	); err != nil {
		t.Fatal(err)
	}

	if _, ok := schema.Properties[field]; !ok {
		t.Fatalf("selected schema missing field %q", field)
	}

	for _, other := range []string{
		"obsoleteField", "betaField", "stableField",
	} {
		if other == field {
			continue
		}

		if _, ok := schema.Properties[other]; ok {
			t.Fatalf("unexpected schema field %q", other)
		}
	}
}
