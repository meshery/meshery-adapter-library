package smi

import (
	"context"
	"fmt"
	"time"

	"github.com/layer5io/learn-layer5/smi-conformance/conformance"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/layer5io/meshkit/utils"
	mesherykube "github.com/layer5io/meshkit/utils/kubernetes"
)

var (
	name      = "smi-conformance"
	namespace = "meshery"
	// smiManifest is the remote location for the SMI manifest
	smiManifest = "https://raw.githubusercontent.com/layer5io/learn-layer5/master/smi-conformance/manifest.yml"
)

type SmiTest struct {
	id             string
	adaptorVersion string
	adaptorName    string
	ctx            context.Context
	kclient        *mesherykube.Client
	smiAddress     string
	annotations    map[string]string
	labels         map[string]string
}

type Response struct {
	Id                string    `json:"id,omitempty"`
	Date              string    `json:"date,omitempty"`
	MeshName          string    `json:"mesh_name,omitempty"`
	MeshVersion       string    `json:"mesh_version,omitempty"`
	CasesPassed       string    `json:"cases_passed,omitempty"`
	PassingPercentage string    `json:"passing_percentage,omitempty"`
	Status            string    `json:"status,omitempty"`
	MoreDetails       []*Detail `json:"more_details,omitempty"`
}

type Detail struct {
	SmiSpecification string `json:"smi_specification,omitempty"`
	SmiVersion       string `json:"smi_version,omitempty"`
	Time             string `json:"time,omitempty"`
	Assertions       string `json:"assertions,omitempty"`
	Result           string `json:"result,omitempty"`
	Reason           string `json:"reason,omitempty"`
	Capability       string `json:"capability,omitempty"`
	Status           string `json:"status,omitempty"`
}

// RunTest initiates the SMI test on the service mesh of the given adapter
func RunTest(
	ctx context.Context,
	id,
	adapterName,
	adapterVersion string,
	labels,
	annotations map[string]string,
	kubeClient *kubernetes.Clientset,
	restConfig rest.Config,
) (Response, error) {
	// Create meshkit kubernetes client
	kclient, err := mesherykube.New(kubeClient, restConfig)
	if err != nil {
		return Response{}, ErrSmiInit(fmt.Sprintf("error creating meshery kubernetes client: %v", err))
	}

	test := &SmiTest{
		ctx:            ctx,
		id:             id,
		adaptorName:    adapterName,
		adaptorVersion: adapterVersion,
		labels:         labels,
		annotations:    annotations,
		kclient:        kclient,
	}

	response := Response{
		Id:                test.id,
		Date:              time.Now().Format(time.RFC3339),
		MeshName:          test.adaptorName,
		MeshVersion:       test.adaptorVersion,
		CasesPassed:       "0",
		PassingPercentage: "0",
		Status:            "deploying",
	}

	if err = test.installConformanceTool(); err != nil {
		response.Status = "installing"
		return response, ErrInstallSmi(err)
	}

	if err = test.connectConformanceTool(); err != nil {
		response.Status = "connecting"
		return response, ErrConnectSmi(err)
	}

	if err = test.runConformanceTest(&response); err != nil {
		response.Status = "running"
		return response, ErrRunSmi(err)
	}

	if err = test.deleteConformanceTool(); err != nil {
		response.Status = "deleting"
		return response, ErrDeleteSmi(err)
	}

	response.Status = "completed"
	return response, nil
}

// installConformanceTool installs the smi conformance tool
func (test *SmiTest) installConformanceTool() error {
	// Fetch the meanifest
	manifest, err := utils.ReadRemoteFile(smiManifest)
	if err != nil {
		return err
	}

	if err := test.kclient.ApplyManifest([]byte(manifest), mesherykube.ApplyOptions{}); err != nil {
		return err
	}

	time.Sleep(20 * time.Second) // Required for all the resources to be created

	return nil
}

// deleteConformanceTool deletes the smi conformance tool
func (test *SmiTest) deleteConformanceTool() error {
	// Fetch the meanifest
	manifest, err := utils.ReadRemoteFile(smiManifest)
	if err != nil {
		return err
	}

	if err := test.kclient.ApplyManifest([]byte(manifest), mesherykube.ApplyOptions{Delete: true}); err != nil {
		return err
	}
	return nil
}

// connectConformanceTool initiates the connection
func (test *SmiTest) connectConformanceTool() error {
	endpoint, err := test.kclient.GetServiceEndpoint(test.ctx, name, namespace)
	if err != nil {
		return err
	}

	test.smiAddress = fmt.Sprintf("%s:%d", endpoint.Address, endpoint.Port)
	return nil
}

// runConformanceTest runs the conformance test
func (test *SmiTest) runConformanceTest(response *Response) error {

	cClient, err := conformance.CreateClient(context.TODO(), test.smiAddress)
	if err != nil {
		return err
	}

	result, err := cClient.CClient.RunTest(context.TODO(), &conformance.Request{
		Annotations: test.annotations,
		Labels:      test.labels,
		Meshname:    test.adaptorName,
		Meshversion: test.adaptorVersion,
	})
	if err != nil {
		return err
	}

	response.CasesPassed = result.Casespassed
	response.PassingPercentage = result.Passpercent

	details := make([]*Detail, 0)

	for _, d := range result.Details {
		details = append(details, &Detail{
			SmiSpecification: d.Smispec,
			Time:             d.Time,
			Assertions:       d.Assertions,
			Result:           d.Result,
			Reason:           d.Reason,
			Capability:       d.Capability,
			Status:           d.Status,
		})
	}

	response.MoreDetails = details

	err = cClient.Close()
	if err != nil {
		return err
	}

	return nil
}
