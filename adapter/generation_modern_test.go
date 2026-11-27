package adapter

import (
	"encoding/json"
	"testing"

	"github.com/meshery/schemas/models/v1beta3/component"
)

func TestGenerateNativeComponent(t *testing.T) {
	resource := `
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
      served: true
      storage: true
      schema:
        openAPIV3Schema:
          type: object
          properties:
            spec:
              type: object
              properties:
                name:
                  type: string
`

	data, err := GenerateNativeComponentWithModel(
		resource, testLinkerdType, testLinkerdRelease,
		MeshModelConfig{Category: testModelCategory},
	)
	if err != nil {
		t.Fatal(err)
	}

	var definition component.ComponentDefinition
	if err := json.Unmarshal(data, &definition); err != nil {
		t.Fatal(err)
	}

	assertGeneratedModel(t, definition)

	if definition.Component.Kind != testServiceProfileKind {
		t.Fatalf("unexpected kind: %s", definition.Component.Kind)
	}

	if definition.Component.Version != testServiceProfileAPIVersion {
		t.Fatalf("unexpected API version: %s", definition.Component.Version)
	}

	if !definition.Metadata.IsNamespaced {
		t.Fatal("expected namespaced component")
	}

	if !json.Valid([]byte(definition.Component.Schema)) {
		t.Fatal("invalid generated schema")
	}
}

func TestGenerateNativeComponentInvalidResource(t *testing.T) {
	if _, err := GenerateNativeComponent("invalid resource"); err == nil {
		t.Fatal("expected generation error")
	}
}

func assertGeneratedModel(t *testing.T, definition component.ComponentDefinition) {
	t.Helper()

	if definition.Model == nil {
		t.Fatal("generated model is missing")
	}

	if definition.Model.Name != testLinkerdName {
		t.Fatal("unexpected model name")
	}

	if definition.Model.Model.Version != testLinkerdRelease {
		t.Fatal("unexpected infrastructure version")
	}

	if definition.Model.Category.Name != testModelCategory {
		t.Fatal("unexpected model category")
	}

}
