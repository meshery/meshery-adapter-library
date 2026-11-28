package adapter

import (
	"encoding/json"
	"fmt"

	"github.com/meshery/schemas/models/core"
)

func newComponentStyles(metadata map[string]interface{}) (*core.ComponentStyles, error) {
	styles := &core.ComponentStyles{}

	if metadata == nil {
		return styles, nil
	}

	values := make(map[string]interface{})

	for _, key := range []string{
		"primaryColor",
		"secondaryColor",
		"svgColor",
		"svgWhite",
		"svgComplete",
		"shape",
	} {
		if value, exists := metadata[key]; exists {
			values[key] = value
		}
	}

	if shape, ok := values["shape"].(string); ok && shape == "circle" {
		values["shape"] = "ellipse"
	}

	data, err := json.Marshal(values)
	if err != nil {
		return nil, fmt.Errorf("encode component styles: %w", err)
	}

	if err := json.Unmarshal(data, styles); err != nil {
		return nil, fmt.Errorf("decode component styles: %w", err)
	}

	return styles, nil
}
