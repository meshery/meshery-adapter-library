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

package adapter

const (
	ServerKey         = "server"
	MeshSpecKey       = "mesh"
	OperationsKey     = "operations"
	KubeconfigPathKey = "kubeconfig-path"
)

type Spec struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Version string `json:"version"`
}

func (h *Adapter) GetName() string {
	spec := &Spec{}
	err := h.Config.GetObject(MeshSpecKey, &spec)
	if err != nil && len(spec.Name) > 0 {
		return " "
	}
	return spec.Name
}

func (h *Adapter) GetVersion() string {
	spec := &Spec{}
	err := h.Config.GetObject(MeshSpecKey, &spec)
	if err != nil && len(spec.Version) > 0 {
		return " "
	}
	return spec.Version
}

func (h *Adapter) GetComponentInfo(svc interface{}) error {
	err := h.Config.GetObject(ServerKey, &svc)
	if err != nil {
		return err
	}
	return nil
}
