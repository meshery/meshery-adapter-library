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

package grpc

import (
	"time"

	"github.com/layer5io/meshery-adapter-library/adapter"
	"github.com/layer5io/meshery-adapter-library/meshes"

	"context"
)

// CreateMeshInstance is the handler function for the method CreateMeshInstance.
func (s *Service) CreateMeshInstance(ctx context.Context, req *meshes.CreateMeshInstanceRequest) (*meshes.CreateMeshInstanceResponse, error) {
	err := s.Handler.CreateInstance(req.K8SConfig, req.ContextName, &s.Channel)
	if err != nil {
		return nil, err
	}
	return &meshes.CreateMeshInstanceResponse{}, nil
}

// MeshName is the handler function for the method MeshName.
func (s *Service) MeshName(ctx context.Context, req *meshes.MeshNameRequest) (*meshes.MeshNameResponse, error) {
	return &meshes.MeshNameResponse{
		Name: s.Handler.GetName(),
	}, nil
}

// ApplyOperation is the handler function for the method ApplyOperation.
func (s *Service) ApplyOperation(ctx context.Context, req *meshes.ApplyRuleRequest) (*meshes.ApplyRuleResponse, error) {
	// TODO: if err is nil then the response is correctly propagated to the client as JSON
	// TODO: Consider whether this is the correct way to handle errors.
	if req == nil {
		return &meshes.ApplyRuleResponse{
			Error:       ErrRequestInvalid.Error(),
			OperationId: "",
		}, ErrRequestInvalid
	}

	operation := adapter.OperationRequest{
		OperationName:     req.OpName,
		Namespace:         req.Namespace,
		Username:          req.Username,
		CustomBody:        req.CustomBody,
		IsDeleteOperation: req.DeleteOp,
		OperationID:       req.OperationId,
	}
	err := s.Handler.ApplyOperation(ctx, operation)
	if err != nil {
		return &meshes.ApplyRuleResponse{
			Error:       err.Error(),
			OperationId: req.OperationId,
		}, err
	}

	return &meshes.ApplyRuleResponse{
		Error:       "",
		OperationId: req.OperationId,
	}, nil
}

// SupportedOperations is the handler function for the method SupportedOperations.
func (s *Service) SupportedOperations(ctx context.Context, req *meshes.SupportedOperationsRequest) (*meshes.SupportedOperationsResponse, error) {
	result, err := s.Handler.ListOperations()
	if err != nil {
		return nil, err
	}

	operations := make([]*meshes.SupportedOperation, 0)
	for key, val := range result {
		operations = append(operations, &meshes.SupportedOperation{
			Key:      key,
			Value:    val.Properties[adapter.OperationDescriptionKey],
			Category: meshes.OpCategory(val.Type),
		})
	}

	return &meshes.SupportedOperationsResponse{
		Ops:   operations,
		Error: "none",
	}, nil
}

// StreamEvents is the handler function for the method StreamEvents.
func (s *Service) StreamEvents(ctx *meshes.EventsRequest, srv meshes.MeshService_StreamEventsServer) error {
	for {
		data := <-s.Channel
		event := &meshes.EventsResponse{
			OperationId: data.(*adapter.Event).Operationid,
			EventType:   meshes.EventType(data.(*adapter.Event).EType),
			Summary:     data.(*adapter.Event).Summary,
			Details:     data.(*adapter.Event).Details,
		}
		if err := srv.Send(event); err != nil {
			// to prevent loosing the event, will re-add to the channel
			go func() {
				s.Channel <- data
			}()
			return err
		}
		time.Sleep(500 * time.Millisecond)
	}
}
