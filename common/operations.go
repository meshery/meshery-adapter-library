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

package common

import (
	"github.com/layer5io/meshery-adapter-library/adapter"
	"github.com/layer5io/meshery-adapter-library/meshes"
)

const (

	// Common Operations
	BookInfoOperation  = "bookinfo"
	HTTPBinOperation   = "httpbin"
	ImageHubOperation  = "imagehub"
	EmojiVotoOperation = "emojivoto"

	// Validate Operations
	SmiConformanceOperation = "smi_conformance"

	// Custom Operation
	CustomOperation = "custom"

	// Additional Properties
	ServiceName = "service_name"
)

var (
	Operations = adapter.Operations{
		BookInfoOperation: &adapter.Operation{
			Type:        int32(meshes.OpCategory_SAMPLE_APPLICATION),
			Description: "Istio Book Info Application",
			Versions:    adapter.NoneVersion,
			Templates: []adapter.Template{
				"https://raw.githubusercontent.com/istio/istio/master/samples/bookinfo/platform/kube/bookinfo.yaml",
			},
			AdditionalProperties: map[string]string{
				ServiceName: BookInfoOperation,
			},
		},
		HTTPBinOperation: &adapter.Operation{
			Type:        int32(meshes.OpCategory_SAMPLE_APPLICATION),
			Description: "HTTPBin Application",
			Versions:    adapter.NoneVersion,
			Templates: []adapter.Template{
				"https://raw.githubusercontent.com/istio/istio/master/samples/httpbin/httpbin.yaml",
			},
			AdditionalProperties: map[string]string{
				ServiceName: HTTPBinOperation,
			},
		},
		ImageHubOperation: &adapter.Operation{
			Type:        int32(meshes.OpCategory_SAMPLE_APPLICATION),
			Description: "Image Hub Application",
			Versions:    adapter.NoneVersion,
			Templates: []adapter.Template{
				"https://raw.githubusercontent.com/layer5io/image-hub/master/deployment.yaml",
			},
			AdditionalProperties: map[string]string{
				ServiceName: ImageHubOperation,
			},
		},
		EmojiVotoOperation: &adapter.Operation{
			Type:        int32(meshes.OpCategory_SAMPLE_APPLICATION),
			Description: "EmojiVoto Application",
			Versions:    adapter.NoneVersion,
			Templates: []adapter.Template{
				"https://raw.githubusercontent.com/BuoyantIO/emojivoto/main/kustomize/deployment/emoji.yml",
				"https://raw.githubusercontent.com/BuoyantIO/emojivoto/main/kustomize/deployment/vote-bot.yml",
				"https://raw.githubusercontent.com/BuoyantIO/emojivoto/main/kustomize/deployment/voting.yml",
				"https://raw.githubusercontent.com/BuoyantIO/emojivoto/main/kustomize/deployment/web.yml",
			},
			AdditionalProperties: map[string]string{
				ServiceName: EmojiVotoOperation,
			},
		},
		CustomOperation: &adapter.Operation{
			Type:        int32(meshes.OpCategory_CUSTOM),
			Description: "Custom YAML",
			Templates:   adapter.NoneTemplate,
		},

		SmiConformanceOperation: &adapter.Operation{
			Type:        int32(meshes.OpCategory_VALIDATE),
			Description: "SMI Conformance",
			Templates: []adapter.Template{
				"https://raw.githubusercontent.com/layer5io/learn-layer5/master/smi-conformance/manifest.yml",
			},
		},
	}
)
