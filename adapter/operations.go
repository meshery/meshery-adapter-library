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

package adapter

import (
	"context"
	"net/url"

	"github.com/layer5io/meshkit/utils"
)

var (
	NoneVersion  = []Version{"none"}
	NoneTemplate = []Template{"none"}
)

type Version string

type Template string

func (t Template) String() string {
	_, err := url.ParseRequestURI(string(t))
	if err != nil {
		return string(t)
	}

	st, err := utils.ReadRemoteFile(string(t))
	if err != nil {
		return ""
	}

	return st
}

type Operation struct {
	Type                 int32             `json:"type,string,omitempty"`
	Description          string            `json:"description,omitempty"`
	Versions             []Version         `json:"versions,omitempty"`
	Templates            []Template        `json:"templates,omitempty"`
	AdditionalProperties map[string]string `json:"additional_properties,omitempty"`
}

type Operations map[string]*Operation

type OperationRequest struct {
	OperationName     string
	Namespace         string
	Username          string
	CustomBody        string
	IsDeleteOperation bool
	OperationID       string
}

func (h *Adapter) ListOperations() (Operations, error) {
	operations := make(Operations)
	err := h.Config.GetObject(OperationsKey, &operations)
	if err != nil {
		return nil, ErrListOperations(err)
	}
	return operations, nil
}

func (h *Adapter) ApplyOperation(context.Context, OperationRequest) error {
	return nil
}
