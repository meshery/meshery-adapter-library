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

// // CreateMeshInstance is the handler function for the method CreateMeshInstance.
// func (s *Service) CreateMeshInstance(ctx context.Context, req *meshes.CreateMeshInstanceRequest) (*meshes.CreateMeshInstanceResponse, error) {
// 	err := s.Handler.CreateInstance(&s.Channel)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &meshes.CreateMeshInstanceResponse{}, nil
// }

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
		K8sConfigs:        req.KubeConfigs,
		Version:           req.Version,
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
			Value:    val.Description,
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
	clientchan := make(chan interface{}, 10)
	go s.EventStreamer.Subscribe(clientchan)
	for {
		data := <-clientchan
		event := &meshes.EventsResponse{
			OperationId:          data.(*meshes.EventsResponse).OperationId,
			EventType:            meshes.EventType(data.(*meshes.EventsResponse).EventType),
			Summary:              data.(*meshes.EventsResponse).Summary,
			Details:              data.(*meshes.EventsResponse).Details,
			ErrorCode:            data.(*meshes.EventsResponse).ErrorCode,
			ProbableCause:        data.(*meshes.EventsResponse).ProbableCause,
			SuggestedRemediation: data.(*meshes.EventsResponse).SuggestedRemediation,
			Component:            data.(*meshes.EventsResponse).Component,
			ComponentName:        data.(*meshes.EventsResponse).ComponentName,
		}

		if err := srv.Send(event); err != nil {
			return err
		}
		time.Sleep(500 * time.Millisecond)
	}
}

// ProcessOAM is the handler function for the method ProcessOAM
func (s *Service) ProcessOAM(ctx context.Context, srv *meshes.ProcessOAMRequest) (*meshes.ProcessOAMResponse, error) {
	operation := adapter.OAMRequest{
		Username:   srv.Username,
		DeleteOp:   srv.DeleteOp,
		OamComps:   srv.OamComps,
		OamConfig:  srv.OamConfig,
		K8sConfigs: srv.KubeConfigs,
	}

	msg, err := s.Handler.ProcessOAM(ctx, operation)
	return &meshes.ProcessOAMResponse{Message: msg}, err
}

// ProcessOAM is the handler function for the method ProcessOAM
func (s *Service) MeshVersions(context.Context, *meshes.MeshVersionsRequest) (*meshes.MeshVersionsResponse, error) {
	versions := make([]string, 0)
	return &meshes.MeshVersionsResponse{
		Version: versions,
	}, nil
}

// ProcessOAM is the handler function for the method ProcessOAM
func (s *Service) ComponentInfo(context.Context, *meshes.ComponentInfoRequest) (*meshes.ComponentInfoResponse, error) {
	err := s.Handler.GetComponentInfo(s)
	if err != nil {
		return nil, err
	}
	return &meshes.ComponentInfoResponse{
		Type:    s.Type,
		Name:    s.Name,
		Version: s.Version,
		GitSha:  s.GitSHA,
	}, nil
}
