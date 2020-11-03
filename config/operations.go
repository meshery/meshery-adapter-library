package config

// import (
// 	"github.com/layer5io/meshery-adapter-library/adapter"
// )

// const (

// 	// Install Operation Commands
// 	InstallBookInfoCommand = "install_book_info"
// 	InstallHTTPBinCommand  = "install_http_bin"
// 	InstallImageHubCommand = "install_image_hub"

// 	// Custom Operation Commands
// 	CustomOpCommand = "custom"
// )

// var (
// 	CommonOperations = adapter.Operations{
// 		InstallBookInfoCommand: &adapter.Operation{
// 			Type: int32(meshes.OpCategory_SAMPLE_APPLICATION),
// 			Properties: map[string]string{
// 				adapter.OperationDescriptionKey:  "Istio Book Info Application",
// 				adapter.OperationVersionKey:      "",
// 				adapter.OperationTemplateNameKey: "bookinfo.yaml",
// 				adapter.OperationServiceNameKey:  "productpage",
// 			},
// 		},
// 		InstallHTTPBinCommand: &adapter.Operation{
// 			Type: int32(meshes.OpCategory_SAMPLE_APPLICATION),
// 			Properties: map[string]string{
// 				adapter.OperationDescriptionKey:  "HTTPbin Application",
// 				adapter.OperationVersionKey:      "",
// 				adapter.OperationTemplateNameKey: "httpbin-consul.yaml",
// 				adapter.OperationServiceNameKey:  "httpbin",
// 			},
// 		},
// 		InstallImageHubCommand: &adapter.Operation{
// 			Type: int32(meshes.OpCategory_SAMPLE_APPLICATION),
// 			Properties: map[string]string{
// 				adapter.OperationDescriptionKey:  "Image Hub Application",
// 				adapter.OperationVersionKey:      "",
// 				adapter.OperationTemplateNameKey: "image-hub.yaml",
// 				adapter.OperationServiceNameKey:  "ingess",
// 			},
// 		},
// 		CustomOpCommand: &adapter.Operation{
// 			Type: int32(meshes.OpCategory_CUSTOM),
// 			Properties: map[string]string{
// 				adapter.OperationDescriptionKey:  "Custom YAML",
// 				adapter.OperationVersionKey:      "",
// 				adapter.OperationTemplateNameKey: "image-hub.yaml",
// 			},
// 		},
// 	}
// )
