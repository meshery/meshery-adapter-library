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

	// Custom Operation Commands
	CustomOpCommand = "custom"
)

var (
	CommonOperations = adapter.Operations{
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
	}
)
