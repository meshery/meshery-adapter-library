package common

import (
	"github.com/layer5io/meshery-adapter-library/adapter"
	"github.com/layer5io/meshery-adapter-library/meshes"
)

const (

	// Operation Commands
	BookInfoCommand  = "bookinfo"
	HTTPBinCommand   = "httpbin"
	ImageHubCommand  = "imagehub"
	EmojiVotoCommand = "emojivoto"

	// Validate Operation Commands
	ValidateSmiConformance = "validate_smi_conformance_test"

	// Custom Operation Commands
	CustomOpCommand = "custom"
)

var (
	Operations = adapter.Operations{
		BookInfoCommand: &adapter.Operation{
			Type:        int32(meshes.OpCategory_SAMPLE_APPLICATION),
			Description: "Istio Book Info Application",
			Versions:    adapter.NoneVersion,
			Template:    "templates/bookinfo.yaml",
		},
		HTTPBinCommand: &adapter.Operation{
			Type:        int32(meshes.OpCategory_SAMPLE_APPLICATION),
			Description: "HTTPBin Application",
			Versions:    adapter.NoneVersion,
			Template:    "templates/httpbin.yaml",
		},
		ImageHubCommand: &adapter.Operation{
			Type:        int32(meshes.OpCategory_SAMPLE_APPLICATION),
			Description: "Image Hub Application",
			Versions:    adapter.NoneVersion,
			Template:    "templates/imagehub.yaml",
		},
		EmojiVotoCommand: &adapter.Operation{
			Type:        int32(meshes.OpCategory_SAMPLE_APPLICATION),
			Description: "EmojiVoto Application",
			Versions:    adapter.NoneVersion,
			Template:    "templates/emojivoto.yaml",
		},
		CustomOpCommand: &adapter.Operation{
			Type:        int32(meshes.OpCategory_CUSTOM),
			Description: "Custom YAML",
			Template:    "templates/custom.yaml",
		},

		ValidateSmiConformance: &adapter.Operation{
			Type:        int32(meshes.OpCategory_VALIDATE),
			Description: "SMI Conformance",
		},
	}
)
