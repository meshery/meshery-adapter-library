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

	"github.com/layer5io/meshery-adapter-library/config"
	"github.com/layer5io/meshkit/logger"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

type Handler interface {
	GetName() string
	CreateInstance([]byte, string, *chan interface{}) error
	ApplyOperation(context.Context, OperationRequest) error
	ListOperations() (Operations, error)

	// Need not implement this method and can be reused
	StreamErr(*Event, error)
	StreamInfo(*Event)
}

type Adapter struct {
	Config config.Handler
	Log    logger.Handler

	Channel *chan interface{}

	KubeClient        *kubernetes.Clientset
	DynamicKubeClient dynamic.Interface
	RestConfig        rest.Config
	ClientcmdConfig   *clientcmdapi.Config
}

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
