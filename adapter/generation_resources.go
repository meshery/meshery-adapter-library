package adapter

import (
	"fmt"
	"strings"

	"github.com/meshery/meshkit/utils"
	k8s "github.com/meshery/meshkit/utils/kubernetes"
	"sigs.k8s.io/yaml"
)

// loadModernComponentResources retrieves and extracts CRDs
// using the adapter's existing preprocessing configuration.
func loadModernComponentResources(cfg StaticCompConfig) ([]string, error) {
	if strings.TrimSpace(cfg.URL) == "" {
		return nil, fmt.Errorf("component source URL is required")
	}

	var manifest string
	var err error

	switch cfg.Method {
	case Manifests:
		manifest, err = utils.ReadFileSource(cfg.URL)
	case HelmCharts:
		manifest, err = k8s.GetManifestsFromHelm(cfg.URL)
	default:
		return nil, fmt.Errorf("unsupported generation method: %s", cfg.Method)
	}

	if err != nil {
		return nil, fmt.Errorf("retrieve component resources: %w", err)
	}

	return extractModernComponentResources(manifest, cfg)
}

func extractModernComponentResources(
	manifest string,
	cfg StaticCompConfig,
) ([]string, error) {
	if cfg.Config.ExtractCrds == nil {
		return nil, fmt.Errorf("CRD extractor is required")
	}

	var resources []string

	for _, resource := range cfg.Config.ExtractCrds(manifest) {
		resource = strings.TrimSpace(resource)

		if resource == "" || resource == nullResource {
			continue
		}

		var header struct {
			Kind string `json:"kind"`
		}

		if err := yaml.Unmarshal([]byte(resource), &header); err != nil {
			return nil, fmt.Errorf("decode resource: %w", err)
		}

		if header.Kind != "CustomResourceDefinition" {
			continue
		}

		resources = append(resources, resource)
	}

	if len(resources) == 0 {
		return nil, fmt.Errorf("no CRD resources found")
	}

	return resources, nil
}

const nullResource = "null"
