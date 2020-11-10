// Copyright 2020 Layer5, Inc.
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

package common

import (
	"github.com/layer5io/meshery-adapter-library/adapter"
	"github.com/layer5io/meshery-adapter-library/meshes"
)

const (

	// Install Operation Commands
	InstallBookInfoCommand = "install_book_info"
	InstallHTTPBinCommand  = "install_http_bin"
	InstallImageHubCommand = "install_image_hub"

	// Validate Operation Commands
	ValidateSmiConformance = "validate_smi_conformance_test"

	// Custom Operation Commands
	CustomOpCommand = "custom"
)

var (
	Operations = adapter.Operations{
		InstallBookInfoCommand: &adapter.Operation{
			Type: int32(meshes.OpCategory_SAMPLE_APPLICATION),
			Properties: map[string]string{
				adapter.OperationDescriptionKey:  "Istio Book Info Application",
				adapter.OperationVersionKey:      "",
				adapter.OperationTemplateNameKey: "templates/bookinfo.yaml",
				adapter.OperationServiceNameKey:  "productpage",
			},
		},
		InstallHTTPBinCommand: &adapter.Operation{
			Type: int32(meshes.OpCategory_SAMPLE_APPLICATION),
			Properties: map[string]string{
				adapter.OperationDescriptionKey:  "HTTPBin Application",
				adapter.OperationVersionKey:      "",
				adapter.OperationTemplateNameKey: "templates/httpbin.yaml",
				adapter.OperationServiceNameKey:  "httpbin",
			},
		},
		InstallImageHubCommand: &adapter.Operation{
			Type: int32(meshes.OpCategory_SAMPLE_APPLICATION),
			Properties: map[string]string{
				adapter.OperationDescriptionKey:  "Image Hub Application",
				adapter.OperationVersionKey:      "",
				adapter.OperationTemplateNameKey: "templates/imagehub.yaml",
				adapter.OperationServiceNameKey:  "web",
			},
		},
		CustomOpCommand: &adapter.Operation{
			Type: int32(meshes.OpCategory_CUSTOM),
			Properties: map[string]string{
				adapter.OperationDescriptionKey:  "Custom YAML",
				adapter.OperationVersionKey:      "",
				adapter.OperationTemplateNameKey: "templates/custom.yaml",
			},
		},

		ValidateSmiConformance: &adapter.Operation{
			Type: int32(meshes.OpCategory_VALIDATE),
			Properties: map[string]string{
				adapter.OperationDescriptionKey: "SMI Conformance Test",
			},
		},
	}
)
