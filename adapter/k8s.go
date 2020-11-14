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
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/layer5io/meshery-adapter-library/meshes"

	"io"
	"io/ioutil"
	"strings"
	"text/template"
	"time"

	gherrors "github.com/pkg/errors"

	v1 "k8s.io/api/core/v1"

	"github.com/ghodss/yaml"
	"github.com/layer5io/meshkit/models"

	"github.com/sirupsen/logrus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// CreateNamespace creates the namespace specified unless it is 'default', or a delete operation.
func (h *Adapter) CreateNamespace(isDelete bool, namespace string) error {
	if !isDelete && namespace != "default" {
		if err := h.createNamespace(context.TODO(), namespace); err != nil {
			logrus.Error(err)
			return err
		}
	}
	return nil
}

// GetServicePorts returns the node port(s) for a specific service in the namespace given.
func (h *Adapter) GetServicePorts(serviceName, namespace string) ([]int64, error) {
	ports, err := h.getServicePorts(context.TODO(), serviceName, namespace)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}
	return ports, nil
}

// ApplyKubernetesManifest merges the file given by templatePath with mergeData and applies it. For a delete operation, the resources are deleted.
// The namespace specified in the operation has to exist.
func (h *Adapter) ApplyKubernetesManifest(request OperationRequest, operation Operation, mergeData map[string]string, templatePath string) error {
	if err := h.applyK8sManifest(context.TODO(), request, operation, mergeData, templatePath); err != nil {
		logrus.Error(err)
		return err
	}
	return nil
}

func (h *Adapter) k8sClientConfig(kubeconfig []byte, contextName string) (*rest.Config, error) {
	if len(kubeconfig) > 0 {
		ccfg, err := clientcmd.Load(kubeconfig)
		if err != nil {
			return nil, err
		}
		if contextName != "" {
			ccfg.CurrentContext = contextName
		}
		err = writeKubeconfig(kubeconfig, contextName, h.KubeConfigPath)
		if err != nil {
			return nil, err
		}
		return clientcmd.NewDefaultClientConfig(*ccfg, &clientcmd.ConfigOverrides{}).ClientConfig()
	}
	return rest.InClusterConfig()
}

// writeKubeconfig creates kubeconfig in local container or file system
func writeKubeconfig(kubeconfig []byte, contextName string, path string) error {
	yamlConfig := models.Kubeconfig{}
	err := yaml.Unmarshal(kubeconfig, &yamlConfig)
	if err != nil {
		return err
	}

	yamlConfig.CurrentContext = contextName

	d, err := yaml.Marshal(yamlConfig)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(path, d, 0600)
	if err != nil {
		return err
	}

	return nil
}

func (h *Adapter) executeRule(ctx context.Context, data *unstructured.Unstructured, namespace string, isDelete, isCustomOp bool) error {
	if namespace != "" {
		data.SetNamespace(namespace)
	}
	groupVersion := strings.Split(data.GetAPIVersion(), "/")
	logrus.Debugf("groupVersion: %v", groupVersion)
	var group, version string
	if len(groupVersion) == 2 {
		group = groupVersion[0]
		version = groupVersion[1]
	} else if len(groupVersion) == 1 {
		version = groupVersion[0]
	}

	kind := strings.ToLower(data.GetKind())
	switch kind {
	case "logentry":
		kind = "logentries"
	case "kubernetes":
		kind = "kuberneteses"
	default:
		kind += "s"
	}

	res := schema.GroupVersionResource{
		Group:    group,
		Version:  version,
		Resource: kind,
	}
	logrus.Debugf("Computed Resource: %+#v", res)

	if isDelete {
		return h.deleteResource(ctx, res, data)
	}

	if err := h.createResource(ctx, res, data); err != nil {
		if isCustomOp {
			if err := h.deleteResource(ctx, res, data); err != nil {
				return err
			}
			time.Sleep(time.Second)
			if err := h.createResource(ctx, res, data); err != nil {
				return err
			}
		} else {
			return err
		}
	}
	return nil
}

func (h *Adapter) createResource(ctx context.Context, res schema.GroupVersionResource, data *unstructured.Unstructured) error {
	_, err := h.DynamicKubeClient.Resource(res).Namespace(data.GetNamespace()).Create(ctx, data, metav1.CreateOptions{})
	if err != nil {
		err = gherrors.Wrapf(err, "unable to create the requested resource, attempting operation without namespace")
		logrus.Warn(err)
		_, err = h.DynamicKubeClient.Resource(res).Create(ctx, data, metav1.CreateOptions{})
		if err != nil {
			err = gherrors.Wrapf(err, "unable to create the requested resource, attempting to update")
			logrus.Error(err)
			return err
		}
	}
	logrus.Infof("Created Resource of type: %s and name: %s", data.GetKind(), data.GetName())
	return nil
}

func (h *Adapter) deleteResource(ctx context.Context, res schema.GroupVersionResource, data *unstructured.Unstructured) error {
	if h.DynamicKubeClient == nil {
		return errors.New("mesh client has not been created")
	}

	if res.Resource == "namespaces" && data.GetName() == "default" { // skipping deletion of default namespace
		return nil
	}

	// in the case with deployments, have to scale it down to 0 first and then delete. . . or else RS and pods will be left behind
	if res.Resource == "deployments" {
		data1, err := h.getResource(ctx, res, data)
		if err != nil {
			return err
		}
		depl := data1.UnstructuredContent()
		spec1 := depl["spec"].(map[string]interface{})
		spec1["replicas"] = 0
		data1.SetUnstructuredContent(depl)
		if err = h.updateResource(ctx, res, data1); err != nil {
			return err
		}
	}

	err := h.DynamicKubeClient.Resource(res).Namespace(data.GetNamespace()).Delete(ctx, data.GetName(), metav1.DeleteOptions{})
	if err != nil {
		err = gherrors.Wrapf(err, "unable to delete the requested resource, attempting operation without namespace")
		logrus.Warn(err)

		err := h.DynamicKubeClient.Resource(res).Delete(ctx, data.GetName(), metav1.DeleteOptions{})
		if err != nil {
			err = gherrors.Wrapf(err, "unable to delete the requested resource")
			logrus.Error(err)
			return err
		}
	}
	logrus.Infof("Deleted Resource of type: %s and name: %s", data.GetKind(), data.GetName())
	return nil
}

func (h *Adapter) getResource(ctx context.Context, res schema.GroupVersionResource, data *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	data1, err := h.DynamicKubeClient.Resource(res).Namespace(data.GetNamespace()).Get(ctx, data.GetName(), metav1.GetOptions{})
	if err != nil {
		err = gherrors.Wrap(err, "unable to retrieve the resource with a matching name, attempting operation without namespace")
		logrus.Warn(err)

		data1, err = h.DynamicKubeClient.Resource(res).Get(ctx, data.GetName(), metav1.GetOptions{})
		if err != nil {
			err = gherrors.Wrap(err, "unable to retrieve the resource with a matching name, while attempting to apply the config")
			logrus.Error(err)
			return nil, err
		}
	}
	logrus.Infof("Retrieved Resource of type: %s and name: %s", data.GetKind(), data.GetName())
	return data1, nil
}

func (h *Adapter) updateResource(ctx context.Context, res schema.GroupVersionResource, data *unstructured.Unstructured) error {
	if _, err := h.DynamicKubeClient.Resource(res).Namespace(data.GetNamespace()).Update(ctx, data, metav1.UpdateOptions{}); err != nil {
		err = gherrors.Wrap(err, "unable to update resource with the given name, attempting operation without namespace")
		logrus.Warn(err)

		if _, err = h.DynamicKubeClient.Resource(res).Update(ctx, data, metav1.UpdateOptions{}); err != nil {
			err = gherrors.Wrap(err, "unable to update resource with the given name, while attempting to apply the config")
			logrus.Error(err)
			return err
		}
	}
	logrus.Infof("Updated Resource of type: %s and name: %s", data.GetKind(), data.GetName())
	return nil
}

func (h *Adapter) applyConfigChange(ctx context.Context, yamlFileContents, namespace string, isDelete, isCustomOp bool) error {
	yamls, err := h.splitYAML(yamlFileContents)
	if err != nil {
		err = gherrors.Wrap(err, "error while splitting yaml")
		logrus.Error(err)
		return err
	}
	for _, yml := range yamls {
		if strings.TrimSpace(yml) != "" {
			if err := h.applyRulePayload(ctx, namespace, []byte(yml), isDelete, isCustomOp); err != nil {
				errStr := strings.TrimSpace(err.Error())
				if isDelete {
					if strings.HasSuffix(errStr, "not found") ||
						strings.HasSuffix(errStr, "the server could not find the requested resource") {
						continue
					}
				} else {
					if strings.HasSuffix(errStr, "already exists") {
						continue
					}
				}
				return err
			}
		}
	}
	return nil
}

func (h *Adapter) applyRulePayload(ctx context.Context, namespace string, newBytes []byte, isDelete, isCustomOp bool) error {
	if h.DynamicKubeClient == nil {
		return errors.New("mesh client has not been created")
	}
	jsonBytes, err := yaml.YAMLToJSON(newBytes)
	if err != nil {
		err = gherrors.Wrapf(err, "unable to convert yaml to json")
		logrus.Error(err)
		return err
	}
	if len(jsonBytes) > 5 { // attempting to skip 'null' json
		data := &unstructured.Unstructured{}
		err = data.UnmarshalJSON(jsonBytes)
		if err != nil {
			err = gherrors.Wrapf(err, "unable to unmarshal json created from yaml")
			logrus.Error(err)
			return err
		}
		if data.IsList() {
			err = data.EachListItem(func(r runtime.Object) error {
				dataL, _ := r.(*unstructured.Unstructured)
				return h.executeRule(ctx, dataL, namespace, isDelete, isCustomOp)
			})
			return err
		}
		return h.executeRule(ctx, data, namespace, isDelete, isCustomOp)
	}
	return nil
}

func (h *Adapter) splitYAML(yamlContents string) ([]string, error) {
	yamlDecoder, ok := NewDocumentDecoder(ioutil.NopCloser(bytes.NewReader([]byte(yamlContents)))).(*YAMLDecoder)
	if !ok {
		err := fmt.Errorf("unable to create a yaml decoder")
		logrus.Error(err)
		return nil, err
	}
	defer func() {
		if err := yamlDecoder.Close(); err != nil {
			logrus.Error(err)
		}
	}()
	var err error
	n := 0
	data := [][]byte{}
	ind := 0
	for err == io.ErrShortBuffer || err == nil {
		d := make([]byte, 1000)
		n, err = yamlDecoder.Read(d)
		if len(data) == 0 || len(data) <= ind {
			data = append(data, []byte{})
		}
		if n > 0 {
			data[ind] = append(data[ind], d...)
		}
		if err == nil {
			logrus.Debugf("..............BOUNDARY................")
			ind++
		}
	}
	result := make([]string, len(data))
	for i, row := range data {
		r := string(row)
		r = strings.Trim(r, "\x00")
		logrus.Debugf("ind: %d, data: %s", i, r)
		result[i] = r
	}
	return result, nil
}

// creates the namespace if it doesn't exist
func (h *Adapter) createNamespace(ctx context.Context, namespace string) error {
	logrus.Debugf("creating namespace: %s", namespace)
	_, errGetNs := h.KubeClient.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if apierrors.IsNotFound(errGetNs) {
		nsSpec := &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: namespace}}
		_, err := h.KubeClient.CoreV1().Namespaces().Create(ctx, nsSpec, metav1.CreateOptions{})
		return err
	}
	return errGetNs
}

func (h *Adapter) executeTemplate(ctx context.Context, data map[string]string, templatePath string) (string, error) {
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		err = gherrors.Wrapf(err, "unable to parse template")
		logrus.Error(err)
		return "", err
	}
	buf := bytes.NewBufferString("")
	err = tmpl.Execute(buf, data)
	if err != nil {
		err = gherrors.Wrapf(err, "unable to execute template")
		logrus.Error(err)
		return "", err
	}
	return buf.String(), nil
}

func (h *Adapter) applyK8sManifest(ctx context.Context, request OperationRequest, operation Operation, data map[string]string, templatePath string) error {
	merged, err := h.executeTemplate(ctx, data, templatePath)
	if err != nil {
		err = gherrors.Wrapf(err, "unable to apply kubernetes manifest (executeTemplate) ")
		logrus.Error(err)
		return err
	}

	isCustomOperation := operation.Type == int32(meshes.OpCategory_CUSTOM)

	if err := h.applyConfigChange(ctx, merged, request.Namespace, request.IsDeleteOperation, isCustomOperation); err != nil {
		err = gherrors.Wrapf(err, "unable to apply kubernetes manifest (applyConfigChange)")
		logrus.Error(err)
		return err
	}

	return nil
}

func (h *Adapter) getServicePorts(ctx context.Context, svc, namespace string) ([]int64, error) {
	ns := &unstructured.Unstructured{}
	res := schema.GroupVersionResource{
		Version:  "v1",
		Resource: "services",
	}
	ns.SetName(svc)
	ns.SetNamespace(namespace)
	ns, err := h.getResource(ctx, res, ns)
	if err != nil {
		err = gherrors.Wrapf(err, "unable to get service details")
		logrus.Error(err)
		return nil, err
	}
	svcInst := ns.UnstructuredContent()
	spec := svcInst["spec"].(map[string]interface{})
	ports, _ := spec["ports"].([]interface{})
	nodePorts := []int64{}
	for _, port := range ports {
		p, _ := port.(map[string]interface{})
		np, ok := p["nodePort"]
		if ok {
			npi, _ := np.(int64)
			nodePorts = append(nodePorts, npi)
		}
	}
	logrus.Debugf("retrieved svc: %+#v", ns)
	return nodePorts, nil
}
