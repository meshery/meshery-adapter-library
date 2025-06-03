// Copyright Meshery Authors
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

package grpc

import (
	"fmt"

	"github.com/meshery/meshkit/errors"
)

var (
	ErrRequestInvalidCode = "1000"
	ErrPanicCode          = "1001"
	ErrGrpcListenerCode   = "1002"
	ErrGrpcServerCode     = "1003"

	ErrRequestInvalid = errors.New(ErrRequestInvalidCode, errors.Alert, []string{"Apply Request invalid"}, []string{}, []string{}, []string{})
)

func ErrPanic(r interface{}) error {
	return errors.New(ErrPanicCode, errors.Alert, []string{fmt.Sprintf("%v", r)}, []string{}, []string{}, []string{})
}

func ErrGrpcListener(err error) error {
	return errors.New(ErrGrpcListenerCode, errors.Alert, []string{"Error during grpc listener initialization"}, []string{err.Error()}, []string{}, []string{})
}

func ErrGrpcServer(err error) error {
	return errors.New(ErrGrpcServerCode, errors.Alert, []string{"Error during grpc server initialization"}, []string{err.Error()}, []string{}, []string{})
}
