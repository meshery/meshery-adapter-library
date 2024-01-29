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

	"github.com/layer5io/meshery-adapter-library/meshes"
	"github.com/layer5io/meshkit/logger"
)

type adapterLogger struct {
	log  logger.Handler
	next Handler
}

func AddLogger(logger logger.Handler, h Handler) Handler {
	return &adapterLogger{
		log:  logger,
		next: h,
	}
}

func (s *adapterLogger) GetName() string {
	if !(len(s.next.GetName()) > 1) {
		s.log.Error(ErrGetName)
	}
	return s.next.GetName()
}

func (s *adapterLogger) GetComponentInfo(svc interface{}) error {
	err := s.next.GetComponentInfo(svc)
	if err != nil {
		s.log.Error(ErrGetName)
	}
	return s.next.GetComponentInfo(svc)
}

// func (s *adapterLogger) CreateInstance(c *chan interface{}) error {
// 	s.log.Info("Creating instance")
// 	err := s.next.CreateInstance(c)
// 	if err != nil {
// 		s.log.Error(err)
// 	}
// 	return err
// }

func (s *adapterLogger) ApplyOperation(ctx context.Context, op OperationRequest) error {
	s.log.Info("Applying operation ", op.OperationName)
	err := s.next.ApplyOperation(ctx, op)
	if err != nil {
		s.log.Error(err)
	}
	return err
}

// ProcessOAM wraps the Handler's ProcessOAM method along with relevant logging
func (s *adapterLogger) ProcessOAM(ctx context.Context, oamRequest OAMRequest) (string, error) {
	s.log.Info("Process model components")
	msg, err := s.next.ProcessOAM(ctx, oamRequest)
	if err != nil {
		s.log.Error(err)
	}

	return msg, err
}

func (s *adapterLogger) ListOperations() (Operations, error) {
	s.log.Info("Listing Operations")
	ops, err := s.next.ListOperations()
	if err != nil {
		s.log.Error(err)
	}
	return ops, err
}

func (s *adapterLogger) StreamErr(e *meshes.EventsResponse, err error) {
	s.log.Error(err)
}

func (s *adapterLogger) StreamInfo(*meshes.EventsResponse) {
	s.log.Info("Sending event response")
}
