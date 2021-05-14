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
	"fmt"

	"github.com/layer5io/meshkit/errors"
)

const (
	ErrGetNameCode              = "1000"
	ErrCreateInstanceCode       = "1001"
	ErrMeshConfigCode           = "1002"
	ErrValidateKubeconfigCode   = "1003"
	ErrClientConfigCode         = "1004"
	ErrClientSetCode            = "1005"
	ErrStreamEventCode          = "1006"
	ErrOpInvalidCode            = "1007"
	ErrApplyOperationCode       = "1008"
	ErrListOperationsCode       = "1009"
	ErrNewSmiCode               = "1010"
	ErrRunSmiCode               = "1011"
	ErrNoResponseCode           = "1011"
	ErrOpenOAMDefintionFileCode = "1013"
	ErrOpenOAMRefFileCode       = "1014"
	ErrJSONMarshalCode          = "1015"
	ErrOAMRetryCode             = "1016"
)

var (
	ErrGetName    = errors.NewDefault(ErrGetNameCode, "Unable to get mesh name")
	ErrOpInvalid  = errors.NewDefault(ErrOpInvalidCode, "Invalid operation")
	ErrNoResponse = errors.NewDefault(ErrNoResponseCode, "No response from the smi tool")

	// ErrAuthInfosInvalidMsg is the error message when the all of auth infos have invalid or inaccessible paths
	// as there certificate paths
	ErrAuthInfosInvalidMsg = fmt.Errorf("none of the auth infos are valid either the certificate path is invalid or is inaccessible")
)

func ErrCreateInstance(err error) error {
	return errors.NewDefault(ErrCreateInstanceCode, "Error creating adapter instance", err.Error())
}

func ErrMeshConfig(err error) error {
	return errors.NewDefault(ErrMeshConfigCode, "Error configuration mesh", err.Error())
}

func ErrValidateKubeconfig(err error) error {
	return errors.NewDefault(ErrValidateKubeconfigCode, "Error validating kubeconfig", err.Error())
}

func ErrClientConfig(err error) error {
	return errors.NewDefault(ErrClientConfigCode, "Error setting client Config", err.Error())
}

func ErrClientSet(err error) error {
	return errors.NewDefault(ErrClientSetCode, "Error setting clientset", err.Error())
}

func ErrStreamEvent(err error) error {
	return errors.NewDefault(ErrStreamEventCode, "Error streaming event", err.Error())
}
func ErrListOperations(err error) error {
	return errors.NewDefault(ErrListOperationsCode, "Error listing operations", err.Error())
}

func ErrNewSmi(err error) error {
	return errors.NewDefault(ErrNewSmiCode, "Error creating new SMI test client", err.Error())
}

func ErrRunSmi(err error) error {
	return errors.NewDefault(ErrRunSmiCode, "Error running SMI conformance test", err.Error())
}

// ErrSmiInit is the error for smi init method
func ErrSmiInit(des string) error {
	return errors.NewDefault(errors.ErrSmiInit, des)
}

// ErrInstallSmi is the error for installing smi tool
func ErrInstallSmi(err error) error {
	return errors.NewDefault(errors.ErrInstallSmi, fmt.Sprintf("Error installing smi tool: %s", err.Error()))
}

// ErrConnectSmi is the error for connecting to smi tool
func ErrConnectSmi(err error) error {
	return errors.NewDefault(errors.ErrConnectSmi, fmt.Sprintf("Error connecting to smi tool: %s", err.Error()))
}

// ErrDeleteSmi is the error for deleting smi tool
func ErrDeleteSmi(err error) error {
	return errors.NewDefault(errors.ErrDeleteSmi, fmt.Sprintf("Error deleting smi tool: %s", err.Error()))
}

// ErrOpenOAMDefintionFile is the error for opening OAM Definition file
func ErrOpenOAMDefintionFile(err error) error {
	return errors.NewDefault(ErrOpenOAMDefintionFileCode, fmt.Sprintf("error opening OAM Definition File: %s", err.Error()))
}

// ErrOpenOAMRefFile is the error for opening OAM Schema Ref file
func ErrOpenOAMRefFile(err error) error {
	return errors.NewDefault(ErrOpenOAMRefFileCode, fmt.Sprintf("error opening OAM Schema Ref File: %s", err.Error()))
}

// ErrJSONMarshal is the error for json marhal failure
func ErrJSONMarshal(err error) error {
	return errors.NewDefault(ErrOAMRetryCode, fmt.Sprintf("error marshal JSON: %s", err.Error()))
}

func ErrOAMRetry(err error) error {
	return errors.NewDefault(ErrOAMRetryCode, fmt.Sprintf("error marshal JSON: %s", err.Error()))
}
