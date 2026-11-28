package adapter

import (
	"encoding/json"
	"fmt"

	"github.com/meshery/meshkit/models/meshmodel/registry"
)

const componentEntityType = "component"

func newComponentRegistrationPayload(connectionJSON, componentJSON []byte) ([]byte, error) {
	if !json.Valid(connectionJSON) {
		return nil, fmt.Errorf("invalid registrant connection JSON")
	}

	if !json.Valid(componentJSON) {
		return nil, fmt.Errorf("invalid component definition JSON")
	}

	var payload registry.MeshModelRegistrantData

	if err := json.Unmarshal(connectionJSON, &payload.Connection); err != nil {
		return nil, fmt.Errorf("decode registrant connection: %w", err)
	}

	payload.EntityType = componentEntityType
	payload.Entity = componentJSON

	result, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal component registration payload: %w", err)
	}

	return result, nil
}
