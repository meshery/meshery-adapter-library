package adapter

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/meshery/meshkit/utils/manifests"
	"github.com/meshery/schemas/models/core"
	"github.com/meshery/schemas/models/v1beta1"
	"github.com/meshery/schemas/models/v1beta1/category"
	"github.com/meshery/schemas/models/v1beta1/model"
)

func newGeneratedModel(name, version string, cfg MeshModelConfig) (*model.ModelDefinition, error) {
	name = strings.TrimSpace(name)
	version = strings.TrimSpace(version)

	if name == "" || version == "" {
		return nil, fmt.Errorf("model name and version are required")
	}

	metadata := &model.ModelDefinition_Metadata{}

	if cfg.Metadata != nil {
		data, err := json.Marshal(cfg.Metadata)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(data, metadata); err != nil {
			return nil, err
		}
	}

	return &model.ModelDefinition{
		SchemaVersion: v1beta1.ModelSchemaVersion,
		Version:       defaultDefinitionVersion,
		Status:        model.Enabled,
		Name:          strings.ToLower(name),
		DisplayName:   manifests.FormatToReadableString(name),
		Category: category.CategoryDefinition{
			Name:     category.CategoryDefinitionName(cfg.Category),
			Metadata: cfg.CategoryMetadata,
		},
		Metadata: metadata,
		Model: model.Model{
			Version: core.SemverString(version),
		},
	}, nil
}
