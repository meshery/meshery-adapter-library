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
	ErrSmiInitCode              = "1007"
	ErrInstallSmiCode           = "1008"
	ErrConnectSmiCode           = "1009"
	ErrDeleteSmiCode            = "1010"
	ErrGenerateComponentsCode   = "1011"
	ErrAuthInfosInvalidMsgCode  = "1012"
)

var (
	ErrGetName    = errors.New(ErrGetNameCode, errors.Alert, []string{"Unable to get mesh name"}, []string{}, []string{}, []string{})
	ErrOpInvalid  = errors.New(ErrOpInvalidCode, errors.Alert, []string{"Invalid operation"}, []string{}, []string{}, []string{})
	ErrNoResponse = errors.New(ErrNoResponseCode, errors.Alert, []string{"No response from the smi tool"}, []string{}, []string{}, []string{})

	// ErrAuthInfosInvalidMsg is the error message when the all of auth infos have invalid or inaccessible paths
	// as there certificate paths
	ErrAuthInfosInvalidMsg = errors.New(
		ErrAuthInfosInvalidMsgCode,
		errors.Alert,
		[]string{"none of the auth info are valid - either the certificate path is invalid or is inaccessible"},
		[]string{"kubernetes authentication info are either invalid or the their certificate paths are invalid causing meshery adapter setup failure"},
		[]string{"kubeconfig passed to meshery may be referring to a \"context\" whose auth info is a file path", "adapter may have cached a copy of kubeconfig"},
		[]string{"ensure kubeconfig passed to meshery is flattened", "if running adapter in kubernetes, attempt to restart the pod; in development environment try deleting ~/.meshery s"},
	)
)

func ErrCreateInstance(err error) error {
	return errors.New(ErrCreateInstanceCode, errors.Alert, []string{"Error creating adapter instance"}, []string{err.Error()}, []string{}, []string{})
}

func ErrMeshConfig(err error) error {
	return errors.New(ErrMeshConfigCode, errors.Alert, []string{"Error configuration mesh"}, []string{err.Error()}, []string{}, []string{})
}

func ErrValidateKubeconfig(err error) error {
	return errors.New(ErrValidateKubeconfigCode, errors.Alert, []string{"Error validating kubeconfig"}, []string{err.Error()}, []string{}, []string{})
}

func ErrClientConfig(err error) error {
	return errors.New(ErrClientConfigCode, errors.Alert, []string{"Error setting client Config"}, []string{err.Error()}, []string{}, []string{})
}

func ErrClientSet(err error) error {
	return errors.New(ErrClientSetCode, errors.Alert, []string{"Error setting clientset"}, []string{err.Error()}, []string{}, []string{})
}

func ErrStreamEvent(err error) error {
	return errors.New(ErrStreamEventCode, errors.Alert, []string{"Error streaming event"}, []string{err.Error()}, []string{}, []string{})
}
func ErrListOperations(err error) error {
	return errors.New(ErrListOperationsCode, errors.Alert, []string{"Error listing operations"}, []string{err.Error()}, []string{}, []string{})
}

func ErrNewSmi(err error) error {
	return errors.New(ErrNewSmiCode, errors.Alert, []string{"Error creating new SMI test client"}, []string{err.Error()}, []string{}, []string{})
}

func ErrRunSmi(err error) error {
	return errors.New(ErrRunSmiCode, errors.Alert, []string{"Error running SMI conformance test"}, []string{err.Error()}, []string{}, []string{})
}

// ErrSmiInit is the error for smi init method
func ErrSmiInit(des string) error {
	return errors.New(ErrSmiInitCode, errors.Alert, []string{des}, []string{}, []string{}, []string{})
}

// ErrInstallSmi is the error for installing smi tool
func ErrInstallSmi(err error) error {
	return errors.New(ErrInstallSmiCode, errors.Alert, []string{"Error installing smi tool"}, []string{err.Error()}, []string{}, []string{})
}

// ErrConnectSmi is the error for connecting to smi tool
func ErrConnectSmi(err error) error {
	return errors.New(ErrConnectSmiCode, errors.Alert, []string{"Error connecting to smi tool: %s"}, []string{err.Error()}, []string{}, []string{})
}

// ErrDeleteSmi is the error for deleting smi tool
func ErrDeleteSmi(err error) error {
	return errors.New(ErrDeleteSmiCode, errors.Alert, []string{"Error deleting smi tool: %s"}, []string{err.Error()}, []string{}, []string{})
}

// ErrOpenOAMDefintionFile is the error for opening OAM Definition file
func ErrOpenOAMDefintionFile(err error) error {
	return errors.New(ErrOpenOAMDefintionFileCode, errors.Alert, []string{"error opening OAM Definition File: %s"}, []string{err.Error()}, []string{}, []string{})
}

// ErrOpenOAMRefFile is the error for opening OAM Schema Ref file
func ErrOpenOAMRefFile(err error) error {
	return errors.New(ErrOpenOAMRefFileCode, errors.Alert, []string{"error opening OAM Schema Ref File: %s"}, []string{err.Error()}, []string{}, []string{})
}

// ErrJSONMarshal is the error for json marhal failure
func ErrJSONMarshal(err error) error {
	return errors.New(ErrOAMRetryCode, errors.Alert, []string{"error marshal JSON: %s"}, []string{err.Error()}, []string{}, []string{})
}

func ErrOAMRetry(err error) error {
	return errors.New(ErrOAMRetryCode, errors.Alert, []string{"error marshal JSON: %s"}, []string{err.Error()}, []string{}, []string{})
}

func ErrGenerateComponents(err error) error {
	return errors.New(ErrGenerateComponentsCode, errors.Alert, []string{"error generating components"}, []string{err.Error()}, []string{"Invalid component generation method passed, Some invalid field passed in DynamicComponentsConfig"}, []string{"Pass the correct GenerationMethod in DynamicComponentsConfig", "Pass the correct fields in DynamicComponentsConfig"})
}
