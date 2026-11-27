package adapter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// CreateModernComponents generates native MeshKit components
// while preserving the adapter's existing directory structure.
func CreateModernComponents(cfg StaticCompConfig) error {
	if strings.TrimSpace(cfg.MeshModelPath) == "" {
		return fmt.Errorf("MeshModel path is required")
	}

	if !componentFilePart.MatchString(cfg.DirName) {
		return fmt.Errorf("invalid component directory version")
	}

	resources, err := loadModernComponentResources(cfg)
	if err != nil {
		return err
	}

	return generateModernComponentFiles(resources, cfg)
}

func generateModernComponentFiles(
	resources []string,
	cfg StaticCompConfig,
) error {
	if len(resources) == 0 {
		return fmt.Errorf("no component resources provided")
	}

	if cfg.MeshModelPath == "" ||
		!componentFilePart.MatchString(cfg.DirName) {
		return fmt.Errorf("invalid component destination")
	}

	generated, err := buildModernComponentFiles(resources, cfg)
	if err != nil {
		return err
	}

	return persistModernComponentFiles(generated, cfg)
}

func persistModernComponentFiles(generated map[string][]byte, cfg StaticCompConfig) error {
	previousVersion, _ := getLatestDirectory(cfg.MeshModelPath)

	destination := filepath.Join(
		cfg.MeshModelPath,
		cfg.DirName,
	)

	if err := os.MkdirAll(destination, 0755); err != nil {
		return err
	}

	for filename, data := range generated {
		path := filepath.Join(destination, filename)

		if err := writeToFile(path, data, cfg.Force); err != nil {
			return fmt.Errorf("write component %s: %w", filename, err)
		}
	}

	if previousVersion != "" &&
		previousVersion != cfg.DirName {
		if err := copyCoreComponentsToNewVersion(
			filepath.Join(cfg.MeshModelPath, previousVersion),
			destination,
			cfg.DirName,
			true,
		); err != nil {
			return fmt.Errorf("preserve core components: %w", err)
		}
	}

	return nil
}

// CreateModernComponentsFromSources generates components
// from multiple sources in a single generation operation.
func CreateModernComponentsFromSources(
	sources []string,
	cfg StaticCompConfig,
) error {
	if len(sources) == 0 {
		return fmt.Errorf("no component sources provided")
	}

	if strings.TrimSpace(cfg.MeshModelPath) == "" ||
		!componentFilePart.MatchString(cfg.DirName) {
		return fmt.Errorf("invalid component destination")
	}

	var resources []string

	for _, source := range sources {
		sourceCfg := cfg
		sourceCfg.URL = source

		extracted, err := loadModernComponentResources(sourceCfg)
		if err != nil {
			return fmt.Errorf("load source %q: %w", source, err)
		}

		resources = append(resources, extracted...)
	}

	return generateModernComponentFiles(resources, cfg)
}

func buildModernComponentFiles(resources []string, cfg StaticCompConfig) (map[string][]byte, error) {
	modelName := cfg.MeshModelName
	if modelName == "" {
		modelName = cfg.Config.Name
	}

	generated := make(map[string][]byte)

	for _, resource := range resources {
		data, err := GenerateNativeComponentWithModel(
			resource,
			modelName,
			cfg.DirName,
			cfg.MeshModelConfig,
		)
		if err != nil {
			return nil, err
		}

		filename, err := modernComponentFilename(
			data, cfg.Config.Type,
		)
		if err != nil {
			return nil, err
		}

		if _, exists := generated[filename]; exists {
			return nil, fmt.Errorf(
				"component filename collision: %s",
				filename,
			)
		}

		generated[filename] = data
	}

	return generated, nil
}
