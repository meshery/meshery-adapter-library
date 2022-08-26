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

// Package adapter provides the default implementation of an adapter Handler, called by the MeshServiceServer implementation.
//
// It implements also common Kubernetes operations, leveraging the comparatively low-level client-go library, and handling of YAML-documents.
package adapter

import (
	"context"
	"sync"

	meshkitCfg "github.com/layer5io/meshkit/config"
	"github.com/layer5io/meshkit/logger"
	"github.com/layer5io/meshkit/utils/events"
)

// Interface Handler is extended by adapters, and used in package api/grpc that implements the MeshServiceServer.
type Handler interface {
	GetName() string                    // Returns the name of the adapter.
	GetComponentInfo(interface{}) error // Returns the component info.
	// CreateInstance(*chan interface{}) error                 // Instantiates clients used in deploying and managing mesh instances, e.g. Kubernetes clients.
	ApplyOperation(context.Context, OperationRequest) error // Applies an adapter operation. This is adapter specific and needs to be implemented by each adapter.
	ListOperations() (Operations, error)                    // List all operations an adapter supports.
	ProcessOAM(ctx context.Context, srv OAMRequest) (string, error)

	// Need not implement this method and can be reused
	StreamErr(*Event, error) // Streams an error event, e.g. to a channel
	StreamInfo(*Event)       // Streams an informational event, e.g. to a channel
}

// Adapter contains all handlers, channels, clients, and other parameters for an adapter.
// Use type embedding in a specific adapter to extend it.
type Adapter struct {
	Config            meshkitCfg.Handler
	KubeconfigHandler meshkitCfg.Handler
	Log               logger.Handler
	EventsBuffer      *events.EventStreamer
	mx                sync.Mutex
}
