// This is a sample config file which is to be added in every adapter according to their config attributes/values
package config

import (
	"fmt"

	configprovider "github.com/layer5io/meshery-adapter-library/config/provider"
	"github.com/layer5io/meshery-adapter-library/status"
	"github.com/layer5io/meshkit/utils"
)

var (
	defaultServerConfig = map[string]string{
		"name":     "sample-adapter",
		"port":     "10000",
		"traceurl": "none",
		"version":  "v0.1.0",
	}

	defaultMeshSpec = map[string]string{
		"name":    "Sample",
		"status":  status.NotInstalled,
		"version": "1.8.2",
	}

	defaultMeshInstance = map[string]string{
		"name":    "Sample",
		"status":  status.NotInstalled,
		"version": "1.8.2",
	}

	defaultProviderConfig = map[string]string{
		"filepath": fmt.Sprintf("%s/.meshery", utils.GetHome()),
		"filename": "sample.yml",
		"filetype": "yaml",
	}

	DefaultOpts = configprovider.Options{
		ServerConfig:   defaultServerConfig,
		MeshSpec:       defaultMeshSpec,
		MeshInstance:   defaultMeshInstance,
		ProviderConfig: defaultProviderConfig,
		Operations:     CommonOperations,
	}
)
