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

package configprovider

import (
	"github.com/layer5io/meshery-adapter-library/adapter"
	"github.com/layer5io/meshery-adapter-library/config"
	"github.com/spf13/viper"
)

const (
	ServerKey       = "server"
	MeshSpecKey     = "mesh"
	MeshInstanceKey = "instance"
	OperationsKey   = "operations"
)

type Viper struct {
	instance *viper.Viper
}

func NewViper(serverConfig map[string]string, meshSpec map[string]string, meshInstance map[string]string, providerConfig map[string]string, operations adapter.Operations) (config.Handler, error) {
	v := viper.New()
	v.AddConfigPath(providerConfig["filepath"])
	v.SetConfigType(providerConfig["filetype"])
	v.SetConfigName(providerConfig["filename"])
	v.AutomaticEnv()

	for key, value := range serverConfig {
		v.SetDefault(ServerKey+"."+key, value)
	}

	for key, value := range meshSpec {
		v.SetDefault(MeshSpecKey+"."+key, value)
	}

	for key, value := range meshInstance {
		v.SetDefault(MeshInstanceKey+"."+key, value)
	}

	for key, value := range operations {
		v.Set(OperationsKey+"."+key, value)
	}

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error
		} else {
			// Config file was found but another error was produced
			return nil, config.ErrViper(err)
		}
	}
	return &Viper{
		instance: v,
	}, nil
}

func (v *Viper) SetKey(key string, value string) {
	v.instance.Set(key, value)
}

func (v *Viper) GetKey(key string) string {
	return v.instance.Get(key).(string)
}

func (v *Viper) Server(result interface{}) error {
	return v.instance.Sub(ServerKey).Unmarshal(result)
}

func (v *Viper) MeshSpec(result interface{}) error {
	return v.instance.Sub(MeshSpecKey).Unmarshal(result)
}

func (v *Viper) MeshInstance(result interface{}) error {
	return v.instance.Sub(MeshInstanceKey).Unmarshal(result)
}

func (v *Viper) Operations(result interface{}) error {
	return v.instance.Sub(OperationsKey).Unmarshal(result)
}
