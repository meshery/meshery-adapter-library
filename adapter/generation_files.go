package adapter

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

var componentFilePart = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9.-]*$`)

// modernComponentFilename preserves the legacy service-mesh
// filename convention without changing component API identity.
func modernComponentFilename(data []byte, meshType string) (string, error) {
	var definition struct {
		Component struct {
			Kind    string `json:"kind"`
			Version string `json:"version"`
		} `json:"component"`
	}

	if err := json.Unmarshal(data, &definition); err != nil {
		return "", fmt.Errorf("decode component identity: %w", err)
	}

	kind := strings.ToLower(definition.Component.Kind)
	version := definition.Component.Version

	group, apiVersion, found := strings.Cut(version, "/")
	if !found {
		apiVersion = group
		group = ""
	}

	if err := validateComponentFileIdentity(kind, apiVersion, group, meshType, found); err != nil {
		return "", err
	}

	if group == "core.meshmodel.dev" {
		return kind + "_meshmodel.json", nil
	}

	if meshType != "" {
		kind += "." + meshType
	}

	return kind + ".meshery.layer5.io_meshmodel.json", nil
}

func validateComponentFileIdentity(kind, apiVersion, group, meshType string, found bool) error {
	if !componentFilePart.MatchString(kind) ||
		!componentFilePart.MatchString(apiVersion) {
		return fmt.Errorf("invalid component identity")
	}

	if found && !componentFilePart.MatchString(group) {
		return fmt.Errorf("invalid component group")
	}

	if meshType != "" && !componentFilePart.MatchString(meshType) {
		return fmt.Errorf("invalid mesh type")
	}

	return nil
}
