package adapter

import (
	"encoding/json"
	"fmt"

	generator "github.com/meshery/meshkit/utils/component"
	"github.com/meshery/schemas/models/v1beta3/component"
)

const defaultDefinitionVersion = "1.0.0"

// GenerateNativeComponent generates a component definition using MeshKit.
func GenerateNativeComponent(resource string) ([]byte, error) {
	normalized, err := selectServedCRDVersion(resource)
	if err != nil {
		return nil, fmt.Errorf("select CRD version: %w", err)
	}

	definition, err := generator.Generate(normalized)
	if err != nil {
		return nil, fmt.Errorf("generate component: %w", err)
	}

	if definition.Component.Kind == "" {
		return nil, fmt.Errorf("generated component kind is empty")
	}

	if definition.Component.Version == "" {
		return nil, fmt.Errorf("generated component API version is empty")
	}

	if !json.Valid([]byte(definition.Component.Schema)) {
		return nil, fmt.Errorf("generated component schema is invalid")
	}

	if definition.Version == "" {
		definition.Version = defaultDefinitionVersion
	}

	if definition.Status == nil {
		status := component.Enabled
		definition.Status = &status
	}

	return json.Marshal(definition)
}

// GenerateNativeComponentWithModel generates a component and associates
// it with the adapter's model configuration.
func GenerateNativeComponentWithModel(
	resource, name, version string,
	cfg MeshModelConfig,
) ([]byte, error) {
	normalized, err := selectServedCRDVersion(resource)
	if err != nil {
		return nil, fmt.Errorf("select CRD version: %w", err)
	}

	definition, err := generator.Generate(normalized)
	if err != nil {
		return nil, fmt.Errorf("generate component: %w", err)
	}

	if definition.Component.Kind == "" ||
		definition.Component.Version == "" {
		return nil, fmt.Errorf("incomplete component identity")
	}

	if !json.Valid([]byte(definition.Component.Schema)) {
		return nil, fmt.Errorf("invalid generated component schema")
	}

	definition.Model, err = newGeneratedModel(name, version, cfg)
	if err != nil {
		return nil, fmt.Errorf("enrich component model: %w", err)
	}

	definition.Styles, err = newComponentStyles(cfg.Metadata)
	if err != nil {
		return nil, err
	}

	if definition.Version == "" {
		definition.Version = defaultDefinitionVersion
	}

	if definition.Status == nil {
		status := component.Enabled
		definition.Status = &status
	}

	return json.Marshal(definition)
}
