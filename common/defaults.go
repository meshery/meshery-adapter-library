// Copyright Meshery Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package common contains code and configuration common to all adapters that can either be used directly or as examples.
package common

import (
	"fmt"

	configprovider "github.com/layer5io/meshery-adapter-library/config/provider"
	"github.com/layer5io/meshery-adapter-library/status"
	"github.com/meshery/meshkit/utils"
)

// The values provided here are examples of config objects that can be used as a starting point for adapter specific configuration.
// Note that most of the entries given here are mandatory.
// For more information about the various config option groups, see the config/provider package.
var (
	defaultServerConfig = map[string]string{
		"name":     "sample-adapter",
		"type":     "adapter",
		"port":     "10000",
		"traceurl": "none",
		"version":  "v0.1.0",
	}

	defaultMeshSpec = map[string]string{
		"name":    "Sample",
		"status":  status.NotInstalled,
		"version": "1.8.2",
	}

	defaultProviderConfig = map[string]string{
		"filepath": fmt.Sprintf("%s/.meshery", utils.GetHome()),
		"filename": "sample.yml",
		"filetype": "yaml",
	}

	// DefaultOpts is an example of options to be injected into a config providers.
	DefaultOpts = configprovider.Options{
		ServerConfig:   defaultServerConfig,
		MeshSpec:       defaultMeshSpec,
		ProviderConfig: defaultProviderConfig,
		Operations:     Operations,
	}
)
