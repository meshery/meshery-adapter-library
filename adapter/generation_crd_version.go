package adapter

import (
	"encoding/json"
	"fmt"

	"sigs.k8s.io/yaml"
)

type crdVersionInfo struct {
	Name    string `json:"name"`
	Served  bool   `json:"served"`
	Storage bool   `json:"storage"`
	Schema  struct {
		OpenAPIV3Schema json.RawMessage `json:"openAPIV3Schema"`
	} `json:"schema"`
}

// selectServedCRDVersion prepares the CRD for MeshKit's
// index-based generator without modifying its global configuration.
func selectServedCRDVersion(resource string) (string, error) {
	var document map[string]json.RawMessage

	if err := yaml.Unmarshal([]byte(resource), &document); err != nil {
		return "", fmt.Errorf("decode CRD: %w", err)
	}

	var spec map[string]json.RawMessage
	if err := json.Unmarshal(document["spec"], &spec); err != nil {
		return "", fmt.Errorf("decode CRD spec: %w", err)
	}

	var versions []json.RawMessage
	if err := json.Unmarshal(spec["versions"], &versions); err != nil {
		return "", fmt.Errorf("decode CRD versions: %w", err)
	}

	sharedSchema, err := extractSharedCRDSchema(document, spec)
	if err != nil {
		return "", err
	}

	index, err := preferredServedVersion(
		versions, sharedSchema,
	)
	if err != nil {
		return "", err
	}

	if index == 0 {
		return resource, nil
	}

	versions[0], versions[index] = versions[index], versions[0]

	spec["versions"], err = json.Marshal(versions)
	if err != nil {
		return "", err
	}

	document["spec"], err = json.Marshal(spec)
	if err != nil {
		return "", err
	}

	data, err := json.Marshal(document)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func extractSharedCRDSchema(
	document, spec map[string]json.RawMessage,
) (json.RawMessage, error) {
	var apiVersion string
	if err := json.Unmarshal(document["apiVersion"], &apiVersion); err != nil {
		return nil, fmt.Errorf("decode CRD API version: %w", err)
	}

	if apiVersion != "apiextensions.k8s.io/v1beta1" ||
		len(spec["validation"]) == 0 {
		return nil, nil
	}

	var sharedSchema struct {
		OpenAPIV3Schema json.RawMessage `json:"openAPIV3Schema"`
	}

	if err := json.Unmarshal(spec["validation"], &sharedSchema); err != nil {
		return nil, fmt.Errorf("decode shared CRD schema: %w", err)
	}

	return sharedSchema.OpenAPIV3Schema, nil
}

func preferredServedVersion(
	versions []json.RawMessage,
	sharedSchema json.RawMessage,
) (int, error) {
	selected := -1

	for index, raw := range versions {
		var version crdVersionInfo
		if err := json.Unmarshal(raw, &version); err != nil {
			return -1, fmt.Errorf("decode CRD version: %w", err)
		}

		if !version.Served {
			continue
		}

		if selected == -1 {
			selected = index
		}

		if version.Storage {
			selected = index
			break
		}
	}

	if selected == -1 {
		return -1, fmt.Errorf("CRD has no served versions")
	}

	return selected, validateSelectedCRDVersion(versions[selected], sharedSchema)
}

func validateSelectedCRDVersion(
	raw, sharedSchema json.RawMessage,
) error {
	var version crdVersionInfo
	if err := json.Unmarshal(raw, &version); err != nil {
		return err
	}

	if version.Name == "" {
		return fmt.Errorf("selected CRD version has no name")
	}

	schema := version.Schema.OpenAPIV3Schema
	if !hasCRDOpenAPISchema(schema) &&
		!hasCRDOpenAPISchema(sharedSchema) {
		return fmt.Errorf(
			"served CRD version %q has no OpenAPI schema",
			version.Name,
		)
	}

	return nil
}

func hasCRDOpenAPISchema(schema json.RawMessage) bool {
	var properties map[string]json.RawMessage
	return json.Unmarshal(schema, &properties) == nil &&
		properties != nil
}
