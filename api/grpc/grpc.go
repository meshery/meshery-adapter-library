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

// Package grpc implements the MeshServiceServer which is the server API for MeshService service.
//
// A specific adapter creates an instance of the struct Service (see below) and populates it with parameters, the adapter handler, etc.
// The adapter handler extends the default adapter handler (see package adapter).
// The struct Service is used as parameter in the func Start (see below) that starts and runs the MeshServiceServer.
// This is usually implemented in the package main of an adapter.
package grpc

import (
	"fmt"
	"net"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"github.com/layer5io/meshery-adapter-library/adapter"
	"github.com/layer5io/meshery-adapter-library/meshes"
	"github.com/meshery/meshkit/utils/events"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Service object holds all the information about the server parameters.
type Service struct {
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Port      string    `json:"port"`
	Version   string    `json:"version"`
	GitSHA    string    `json:"gitsha"`
	StartedAt time.Time `json:"startedat"`
	TraceURL  string    `json:"traceurl"`

	Handler       adapter.Handler
	EventStreamer *events.EventStreamer

	meshes.UnimplementedMeshServiceServer
}

// panicHandler is the handler function to handle panic errors.
func panicHandler(r interface{}) error {
	fmt.Println("600 Error")
	return ErrPanic(r)
}

// Start starts grpc server.
func Start(s *Service) error {
	address := fmt.Sprintf(":%s", s.Port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return ErrGrpcListener(err)
	}

	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			otelgrpc.UnaryServerInterceptor(),
			recovery.UnaryServerInterceptor(recovery.WithRecoveryHandler(panicHandler)),
		),
	)

	// Reflection is enabled to simplify accessing the gRPC service using gRPCurl, e.g.
	//    grpcurl --plaintext localhost:10002 meshes.MeshService.SupportedOperations
	// If the use of reflection is not desirable, the parameters '-import-path ./meshes/ -proto meshops.proto' have
	//    to be added to each grpcurl request, with the appropriate import path.
	reflection.Register(server)

	meshes.RegisterMeshServiceServer(server, s)

	if err = server.Serve(listener); err != nil {
		return ErrGrpcServer(err)
	}
	return nil
}
