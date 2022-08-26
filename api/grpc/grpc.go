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

// Package grpc implements the MeshServiceServer which is the server API for MeshService service.
//
// A specific adapter creates an instance of the struct Service (see below) and populates it with parameters, the adapter handler, etc.
// The adapter handler extends the default adapter handler (see package adapter).
// The struct Service is used as parameter in the func Start (see below) that starts and runs the MeshServiceServer.
// This is usually implemented in the package main of an adapter.
package grpc

import (
	"net"
	"time"

	"google.golang.org/grpc/reflection"

	"github.com/layer5io/meshery-adapter-library/adapter"
	"github.com/layer5io/meshery-adapter-library/api/tracing"
	"github.com/layer5io/meshery-adapter-library/meshes"
	"github.com/layer5io/meshkit/utils/events"

	"fmt"

	middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	otelgrpc "go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc"

	apitrace "go.opentelemetry.io/otel/api/trace"
	"google.golang.org/grpc"
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
func Start(s *Service, tr tracing.Handler) error {
	address := fmt.Sprintf(":%s", s.Port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return ErrGrpcListener(err)
	}

	middlewares := middleware.ChainUnaryServer(
		grpc_recovery.UnaryServerInterceptor(
			grpc_recovery.WithRecoveryHandler(panicHandler),
		),
	)
	if tr != nil {
		middlewares = middleware.ChainUnaryServer(
			otelgrpc.UnaryServerInterceptor(tr.Tracer(s.Name).(apitrace.Tracer)),
		)
	}

	server := grpc.NewServer(
		grpc.UnaryInterceptor(middlewares),
	)
	// Reflection is enabled to simplify accessing the gRPC service using gRPCurl, e.g.
	//    grpcurl --plaintext localhost:10002 meshes.MeshService.SupportedOperations
	// If the use of reflection is not desirable, the parameters '-import-path ./meshes/ -proto meshops.proto' have
	//    to be added to each grpcurl request, with the appropriate import path.
	reflection.Register(server)

	//Register Proto
	meshes.RegisterMeshServiceServer(server, s)

	// Start serving requests
	if err = server.Serve(listener); err != nil {
		return ErrGrpcServer(err)
	}
	return nil
}
