package adapter

import (
	"os"
	"path/filepath"
)

func GetCapabilityDefinitionPaths(basePath string, versionsHolder *map[string]bool) ([]string, error) {
	var res []string
	if err := filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if matched, err := filepath.Match("*_component_definition.json", filepath.Base(path)); err != nil {
			return err
		} else if matched {

			res = append(res, path)
			(*versionsHolder)[filepath.Base(filepath.Dir(path))] = true // Getting available versions already existing on file system
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return res, nil
}
