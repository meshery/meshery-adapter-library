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

type Service string

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

// Operation represents an operation of a given Type (see meshes.OpCategory), with a set of properties.
type Operation struct {
	Type                 int32             `json:"type,string,omitempty"`
	Description          string            `json:"description,omitempty"`
	Versions             []Version         `json:"versions,omitempty"`
	Templates            []Template        `json:"templates,omitempty"`
	Services             []Service         `json:"services,omitempty"`
	AdditionalProperties map[string]string `json:"additional_properties,omitempty"`
}

// Operations contains all operations supported by an adapter.
type Operations map[string]*Operation

// OperationRequest contains the request data from meshes.ApplyRuleRequest.
type OperationRequest struct {
	OperationName     string // The identifier of the operation. It is used as key in the Operations map. Avoid using a verb as part of the name, as it designates both provisioning as deprovisioning operations.
	Namespace         string // The namespace to use in the environment, e.g. Kubernetes, where the operation is applied.
	Username          string // User to execute operation as, if any.
	CustomBody        string // Custom operation manifest, in the case of a custom operation (OpCategory_CUSTOM).
	IsDeleteOperation bool   // If true, the operation specified by OperationName is reverted, i.e. all resources created are deleted.
	OperationID       string // ID of the operation, if any. This identifies a specific operation invocation.
}

type OAMRequest struct {
	Username  string
	DeleteOp  bool
	OamComps  []string
	OamConfig string
}

// List all operations an adapter supports.
func (h *Adapter) ListOperations() (Operations, error) {
	operations := make(Operations)
	err := h.Config.GetObject(OperationsKey, &operations)
	if err != nil {
		return nil, ErrListOperations(err)
	}
	return operations, nil
}

// Applies an adapter operation. This is adapter specific and needs to be implemented by each adapter.
func (h *Adapter) ApplyOperation(context.Context, OperationRequest) error {
	return nil
}

// ProcessOAM processes OAM components. This is adapter specific and needs to be implemented by each adapter.
func (h *Adapter) ProcessOAM(context.Context, OAMRequest) error {
	return nil
}
