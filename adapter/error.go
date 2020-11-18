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
	"github.com/layer5io/meshkit/errors"
)

const (
	ErrGetNameCode            = "1000"
	ErrCreateInstanceCode     = "1001"
	ErrMeshConfigCode         = "1002"
	ErrValidateKubeconfigCode = "1003"
	ErrClientConfigCode       = "1004"
	ErrClientSetCode          = "1005"
	ErrStreamEventCode        = "1006"
	ErrOpInvalidCode          = "1007"
	ErrApplyOperationCode     = "1008"
	ErrListOperationsCode     = "1009"
	ErrNewSmiCode             = "1010"
	ErrRunSmiCode             = "1011"
)

var (
	ErrGetName   = errors.NewDefault(ErrGetNameCode, "Unable to get mesh name")
	ErrOpInvalid = errors.NewDefault(ErrOpInvalidCode, "Invalid operation")
)

func ErrCreateInstance(err error) error {
	return errors.NewDefault(ErrCreateInstanceCode, "Error creating adapter instance: ", err.Error())
}

func ErrMeshConfig(err error) error {
	return errors.NewDefault(ErrMeshConfigCode, "Error configuration mesh: ", err.Error())
}

func ErrValidateKubeconfig(err error) error {
	return errors.NewDefault(ErrValidateKubeconfigCode, "Error validating kubeconfig: ", err.Error())
}

func ErrClientConfig(err error) error {
	return errors.NewDefault(ErrClientConfigCode, "Error setting client Config: ", err.Error())
}

func ErrClientSet(err error) error {
	return errors.NewDefault(ErrClientSetCode, "Error setting clientset: ", err.Error())
}

func ErrStreamEvent(err error) error {
	return errors.NewDefault(ErrStreamEventCode, "Error streaming event: ", err.Error())
}
func ErrListOperations(err error) error {
	return errors.NewDefault(ErrListOperationsCode, "Error listing operations: ", err.Error())
}

func ErrNewSmi(err error) error {
	return errors.NewDefault(ErrNewSmiCode, "Error creating new SMI test client: ", err.Error())
}

func ErrRunSmi(err error) error {
	return errors.NewDefault(ErrRunSmiCode, "Error running SMI conformance test: ", err.Error())
}
