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
	"github.com/layer5io/meshkit/utils"
)

// InMem instance for configuration
type InMem struct {
	store map[string]string
}

// NewInMem intializes an in-memory instance for config
func NewInMem(opts Options) (config.Handler, error) {
	store := make(map[string]string)

	val, err := utils.Marshal(opts.ServerConfig)
	if err != nil {
		return nil, config.ErrInMem(err)
	}
	store[adapter.ServerKey] = val

	val, err = utils.Marshal(opts.MeshSpec)
	if err != nil {
		return nil, config.ErrInMem(err)
	}
	store[adapter.MeshSpecKey] = val

	val, err = utils.Marshal(opts.Operations)
	if err != nil {
		return nil, config.ErrInMem(err)
	}
	store[adapter.OperationsKey] = val

	return &InMem{
		store: store,
	}, nil
}

// -------------------------------------------Application config methods----------------------------------------------------------------

// SetKey sets a key value in local store
func (l *InMem) SetKey(key string, value string) {
	l.store[key] = value
}

// GetKey gets a key value from local store
func (l *InMem) GetKey(key string) string {
	return l.store[key]
}

// GetObject gets an object value for the key
func (l *InMem) GetObject(key string, result interface{}) error {
	return utils.Unmarshal(l.store[key], result)
}

// SetObject sets an object value for the key
func (l *InMem) SetObject(key string, value interface{}) error {
	val, err := utils.Marshal(value)
	if err != nil {
		return config.ErrInMem(err)
	}
	l.store[key] = val
	return nil
}
