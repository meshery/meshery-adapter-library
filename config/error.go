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

package config

import (
	"github.com/layer5io/meshkit/errors"
)

var (
	ErrEmptyConfig = errors.NewDefault(errors.ErrEmptyConfig, "Config not initialized")
)

func ErrViper(err error) error {
	return errors.NewDefault(errors.ErrViper, "Viper initialization failed with error: ", err.Error())
}

func ErrInMem(err error) error {
	return errors.NewDefault(errors.ErrInMem, "InMem initialization failed with error: ", err.Error())
}
