// This file needs a better name
package provider

import (
	"github.com/layer5io/meshery-adapter-library/adapter"
)

const (
	//Provider keys
	Viper = "viper"
	InMem = "in-mem"

	// Config keys
	ServerKey       = "server"
	MeshSpecKey     = "mesh"
	MeshInstanceKey = "instance"
	OperationsKey   = "operations"
)

type Options struct {
	ServerConfig   map[string]string
	MeshSpec       map[string]string
	MeshInstance   map[string]string
	ProviderConfig map[string]string
	Operations     adapter.Operations
}
