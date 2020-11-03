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

package provider

import (
	"github.com/layer5io/meshery-adapter-library/adapter"
	"github.com/layer5io/meshery-adapter-library/config"
	"github.com/spf13/viper"
)

const (
	FilePath = "filepath"
	FileType = "filetype"
	FileName = "filename"
)

type Viper struct {
	instance *viper.Viper
}

func NewViper(opts Options) (config.Handler, error) {
	v := viper.New()
	v.AddConfigPath(opts.ProviderConfig[FilePath])
	v.SetConfigType(opts.ProviderConfig[FileType])
	v.SetConfigName(opts.ProviderConfig[FileName])
	v.AutomaticEnv()

	for key, value := range opts.ServerConfig {
		v.SetDefault(adapter.ServerKey+"."+key, value)
	}

	for key, value := range opts.MeshSpec {
		v.SetDefault(adapter.MeshSpecKey+"."+key, value)
	}

	for key, value := range opts.MeshInstance {
		v.SetDefault(adapter.MeshInstanceKey+"."+key, value)
	}

	for key, value := range opts.Operations {
		v.Set(adapter.OperationsKey+"."+key, value)
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

func (v *Viper) GetObject(key string, result interface{}) error {
	return v.instance.Sub(key).Unmarshal(result)
}
