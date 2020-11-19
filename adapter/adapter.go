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

	"github.com/layer5io/meshery-adapter-library/config"
	"github.com/layer5io/meshkit/logger"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

// Interface Handler is extended by adapters, and used in package api/grpc that implements the MeshServiceServer.
type Handler interface {
	GetName() string                                        // Returns the name of the adapter.
	CreateInstance([]byte, string, *chan interface{}) error // Instantiates clients used in deploying and managing mesh instances, e.g. Kubernetes clients.
	ApplyOperation(context.Context, OperationRequest) error // Applies an adapter operation. This is adapter specific and needs to be implemented by each adapter.
	ListOperations() (Operations, error)                    // List all operations an adapter supports.

	// Need not implement this method and can be reused
	StreamErr(*Event, error) // Streams an error event, e.g. to a channel
	StreamInfo(*Event)       // Streams an informational event, e.g. to a channel
}

// Adapter contains all handlers, channels, clients, and other parameters for an adapter.
// Use type embedding in a specific adapter to extend it.
type Adapter struct {
	Config config.Handler
	Log    logger.Handler

	Channel *chan interface{}

	KubeClient        *kubernetes.Clientset
	DynamicKubeClient dynamic.Interface
	RestConfig        rest.Config
	ClientcmdConfig   *clientcmdapi.Config
}

// Instantiates clients used in deploying and managing mesh instances, e.g. Kubernetes clients.
// This needs to be called before applying operations.
func (h *Adapter) CreateInstance(kubeconfig []byte, contextName string, ch *chan interface{}) error {
	err := h.validateKubeconfig(kubeconfig)
	if err != nil {
		return ErrCreateInstance(err)
	}

	err = h.createKubeClient(kubeconfig)
	if err != nil {
		return ErrCreateInstance(err)
	}

	h.ClientcmdConfig.CurrentContext = contextName
	h.Channel = ch

	return nil
}

func (h *Adapter) createKubeClient(kubeconfig []byte) error {
	var (
		restConfig *rest.Config
		err        error
	)

	if len(kubeconfig) > 0 {
		restConfig, err = clientcmd.RESTConfigFromKubeConfig(kubeconfig)
		if err != nil {
			return ErrClientSet(err)
		}
	} else {
		restConfig, err = rest.InClusterConfig()
		if err != nil {
			return ErrClientSet(err)
		}
	}

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return ErrClientSet(err)
	}

	dynamicClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return ErrClientSet(err)
	}

	h.KubeClient = clientset
	h.DynamicKubeClient = dynamicClient
	h.RestConfig = *restConfig
	return nil
}

func (h *Adapter) validateKubeconfig(kubeconfig []byte) error {
	clientcmdConfig, err := clientcmd.Load(kubeconfig)
	if err != nil {
		return ErrValidateKubeconfig(err)
	}

	err = clientcmdapi.FlattenConfig(clientcmdConfig)
	if err != nil {
		return ErrValidateKubeconfig(err)
	}

	err = clientcmdapi.MinifyConfig(clientcmdConfig)
	if err != nil {
		return ErrValidateKubeconfig(err)
	}

	h.ClientcmdConfig = clientcmdConfig

	return nil
}
