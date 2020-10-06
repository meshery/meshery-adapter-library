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

	"github.com/layer5io/gokit/errors"
)

var (
	ErrOpInvalid = errors.New(errors.ErrOpInvalid, "Invalid operation")
)

func ErrInstallMesh(err error) error {
	return errors.New(errors.ErrInstallMesh, fmt.Sprintf("Error installing mesh: %s", err.Error()))
}

func ErrMeshConfig(err error) error {
	return errors.New(errors.ErrMeshConfig, fmt.Sprintf("Error configuration mesh: %s", err.Error()))
}

func ErrPortForward(err error) error {
	return errors.New(errors.ErrPortForward, fmt.Sprintf("Error portforwarding mesh gui: %s", err.Error()))
}

func ErrClientConfig(err error) error {
	return errors.New(errors.ErrClientConfig, fmt.Sprintf("Error setting client Config: %s", err.Error()))
}

func ErrClientSet(err error) error {
	return errors.New(errors.ErrClientSet, fmt.Sprintf("Error setting clientset: %s", err.Error()))
}

func ErrStreamEvent(err error) error {
	return errors.New(errors.ErrStreamEvent, fmt.Sprintf("Error streaming event: %s", err.Error()))
}
