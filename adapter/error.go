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

var (
	ErrOpInvalid = errors.NewDefault(errors.ErrOpInvalid, "Invalid operation")
	ErrGetName   = errors.NewDefault(errors.ErrGetName, "Unable to get mesh name")
)

func ErrInstallMesh(err error) error {
	return errors.NewDefault(errors.ErrInstallMesh, "Error installing mesh: ", err.Error())
}

func ErrMeshConfig(err error) error {
	return errors.NewDefault(errors.ErrMeshConfig, "Error configuration mesh: ", err.Error())
}

func ErrPortForward(err error) error {
	return errors.NewDefault(errors.ErrPortForward, "Error portforwarding mesh gui: ", err.Error())
}

func ErrClientConfig(err error) error {
	return errors.NewDefault(errors.ErrClientConfig, "Error setting client Config: ", err.Error())
}

func ErrClientSet(err error) error {
	return errors.NewDefault(errors.ErrClientSet, "Error setting clientset: ", err.Error())
}

func ErrStreamEvent(err error) error {
	return errors.NewDefault(errors.ErrStreamEvent, "Error streaming event: ", err.Error())
}
