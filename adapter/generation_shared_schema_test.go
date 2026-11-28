package adapter

import (
	"strings"
	"testing"
)

const sharedSchemaReviewCRD = `
apiVersion: apiextensions.k8s.io/v1beta1
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
    - name: v1beta1
      served: true
      storage: true
  validation:
    openAPIV3Schema:
      type: object
      properties:
        sharedField:
          type: string
`

func TestServedCRDVersionSharedSchema(t *testing.T) {
	data, err := GenerateNativeComponent(sharedSchemaReviewCRD)
	if err != nil {
		t.Fatal(err)
	}

	assertSelectedCRDVersion(
		t, data, "linkerd.io/v1beta1", "sharedField",
	)
}

func TestServedCRDVersionSharedSchemaWithModel(t *testing.T) {
	data, err := GenerateNativeComponentWithModel(
		sharedSchemaReviewCRD,
		testLinkerdType,
		testLinkerdRelease,
		MeshModelConfig{Category: testModelCategory},
	)
	if err != nil {
		t.Fatal(err)
	}

	assertSelectedCRDVersion(
		t, data, "linkerd.io/v1beta1", "sharedField",
	)
}

func TestServedCRDVersionRejectsModernSharedSchema(t *testing.T) {
	resource := strings.Replace(
		sharedSchemaReviewCRD,
		"apiextensions.k8s.io/v1beta1",
		"apiextensions.k8s.io/v1",
		1,
	)

	_, err := GenerateNativeComponent(resource)
	if err == nil ||
		!strings.Contains(err.Error(), "no OpenAPI schema") {
		t.Fatalf("expected missing-schema error: %v", err)
	}
}

func TestServedCRDVersionRejectsNullSharedSchema(t *testing.T) {
	resource := strings.Replace(
		sharedSchemaReviewCRD,
		"openAPIV3Schema:",
		"openAPIV3Schema: null\n    unusedSchema:",
		1,
	)

	_, err := GenerateNativeComponent(resource)
	if err == nil ||
		!strings.Contains(err.Error(), "no OpenAPI schema") {
		t.Fatalf("expected missing-schema error: %v", err)
	}
}
