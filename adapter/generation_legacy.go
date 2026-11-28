package adapter

import (
	"encoding/json"
	"fmt"

	"github.com/meshery/schemas/models/core"
	"github.com/meshery/schemas/models/v1beta3"
	"github.com/meshery/schemas/models/v1beta3/component"
)

type legacyComponentDefinition struct {
	Kind        string                 `json:"kind"`
	APIVersion  string                 `json:"apiVersion"`
	DisplayName string                 `json:"displayName"`
	Format      string                 `json:"format"`
	Schema      string                 `json:"schema"`
	Metadata    map[string]interface{} `json:"metadata"`
	Model       struct {
		Name        string                 `json:"name"`
		Version     string                 `json:"version"`
		DisplayName string                 `json:"displayName"`
		Metadata    map[string]interface{} `json:"metadata"`
		Category    struct {
			Name     string                 `json:"name"`
			Metadata map[string]interface{} `json:"metadata"`
		} `json:"category"`
	} `json:"model"`
}

// migrateLegacyComponent converts an existing component definition
// into the modern component structure.
func migrateLegacyComponent(data []byte) ([]byte, error) {
	var legacy legacyComponentDefinition

	if err := json.Unmarshal(data, &legacy); err != nil {
		return nil, fmt.Errorf("decode legacy component: %w", err)
	}

	if err := validateLegacyComponent(legacy); err != nil {
		return nil, err
	}

	metadata := legacy.Model.Metadata
	if metadata == nil {
		metadata = legacy.Metadata
	}

	m, err := newGeneratedModel(
		legacy.Model.Name,
		legacy.Model.Version,
		MeshModelConfig{
			Category:         legacy.Model.Category.Name,
			CategoryMetadata: legacy.Model.Category.Metadata,
			Metadata:         metadata,
		},
	)
	if err != nil {
		return nil, err
	}

	if legacy.Model.DisplayName != "" {
		m.DisplayName = legacy.Model.DisplayName
	}

	definition := component.ComponentDefinition{
		SchemaVersion: v1beta3.ComponentSchemaVersion,
		Version:       defaultDefinitionVersion,
		DisplayName:   core.InputString(legacy.DisplayName),
		Format:        component.JSON,
		Model:         m,
		Component: component.Component{
			Kind:    legacy.Kind,
			Version: legacy.APIVersion,
			Schema:  legacy.Schema,
		},
	}

	if err := applyLegacyComponentMetadata(&definition, legacy.Metadata); err != nil {
		return nil, err
	}

	definition.Styles, err = newComponentStyles(legacy.Metadata)
	if err != nil {
		return nil, err
	}

	status := component.Enabled
	definition.Status = &status

	return json.Marshal(definition)
}

// prepareComponentForRegistration accepts modern component
// definitions and migrates existing legacy definitions.
func prepareComponentForRegistration(data []byte) ([]byte, error) {
	var header struct {
		SchemaVersion string          `json:"schemaVersion"`
		Component     json.RawMessage `json:"component"`
		Kind          string          `json:"kind"`
	}

	if err := json.Unmarshal(data, &header); err != nil {
		return nil, fmt.Errorf("decode component: %w", err)
	}

	if header.SchemaVersion != "" && len(header.Component) > 0 {
		return validateModernComponent(data, header.Component)
	}

	if header.Kind != "" {
		return migrateLegacyComponent(data)
	}

	return nil, fmt.Errorf("unsupported component definition")
}

func validateModernComponent(data, componentJSON []byte) ([]byte, error) {
	var identity struct {
		Kind    string `json:"kind"`
		Version string `json:"version"`
		Schema  string `json:"schema"`
	}

	if err := json.Unmarshal(componentJSON, &identity); err != nil {
		return nil, err
	}

	if identity.Kind == "" || identity.Version == "" {
		return nil, fmt.Errorf("incomplete modern component identity")
	}

	var definition component.ComponentDefinition
	if err := json.Unmarshal(data, &definition); err != nil {
		return nil, fmt.Errorf("decode modern component: %w", err)
	}

	if definition.Model == nil {
		return nil, fmt.Errorf("component model is required")
	}

	if definition.Model.Name == "" ||
		definition.Model.Model.Version == "" {
		return nil, fmt.Errorf("incomplete component model identity")
	}

	if err := validateRegistrationSchema(identity.Schema, definition.Metadata.IsAnnotation); err != nil {
		return nil, err
	}

	return data, nil
}

func validateRegistrationSchema(schema string, isAnnotation bool) error {
	if schema == "" {
		if !isAnnotation {
			return fmt.Errorf("component schema is required")
		}
	} else if !json.Valid([]byte(schema)) {
		return fmt.Errorf("invalid component schema")
	}

	return nil
}

func applyLegacyComponentMetadata(definition *component.ComponentDefinition, metadata map[string]interface{}) error {
	if metadata == nil {
		return nil
	}

	raw, err := json.Marshal(metadata)
	if err != nil {
		return err
	}

	return json.Unmarshal(raw, &definition.Metadata)
}

func validateLegacyComponent(legacy legacyComponentDefinition) error {
	if legacy.Kind == "" || legacy.APIVersion == "" {
		return fmt.Errorf("incomplete legacy component identity")
	}

	// Older component definitions may omit the format field.
	// Accept them only when their schema contains valid JSON.
	if legacy.Format != "" && legacy.Format != string(component.JSON) {
		return fmt.Errorf("unsupported legacy schema format: %q", legacy.Format)
	}

	if !json.Valid([]byte(legacy.Schema)) {
		return fmt.Errorf("invalid legacy JSON schema")
	}

	return nil
}
