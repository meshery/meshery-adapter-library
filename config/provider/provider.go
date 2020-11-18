// This file needs a better name
package provider

import (
	"github.com/layer5io/meshery-adapter-library/adapter"
)

const (
	//Provider keys
	ViperKey = "viper"
	InMemKey = "in-mem"
)

type Options struct {
	ServerConfig   map[string]string
	MeshSpec       map[string]string
	ProviderConfig map[string]string
	Operations     adapter.Operations
}
