package adapter

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/layer5io/meshkit/models/meshmodel/core/v1alpha1"
	"github.com/layer5io/meshkit/utils/manifests"
)

type RegistrantMetadata struct {
	// Host is the address of the grpc service of the registrant
	Host string `json:"host,omitempty"`
}

type Capability struct {
	ID                   string `json:"id,omitempty"`
	Restricted           bool   `json:"restricted,omitempty"`
	RegistrantMetadata   `json:"registrant_metadata,omitempty"`
	CapabilityDefinition interface{} `json:"capability_definition,omitempty"`
}

// Registrant provides utility functions for registering
// capabilities to a registry in a reliable way
type Registrant struct {
	// CapabilityDefinitionPaths is a slice for holding the paths of CapabilityDefinitions
	// on the filesystem
	//
	// Registrant will read the definitions from these
	// paths and will register them to the capabilities registry
	CapabilityDefinitionPaths []string
	Host                      string

	// HTTPRegistry is the address of the capabilities registry
	HTTPRegistry string
}

func NewRegistrant(capDefPaths []string, HTTPRegistry string, host string) *Registrant {
	return &Registrant{
		CapabilityDefinitionPaths: capDefPaths,
		HTTPRegistry:              HTTPRegistry,
		Host:                      host,
	}
}

// RegisterCapabilities will register all of the capability definitions
// present in the path oam/workloads
//
// Registration process will send POST request to $runtime/api/oam/workload
func RegisterCapabilities(runtime, host string, paths []string) error {

	return NewRegistrant(paths, fmt.Sprintf("%s/api/oam/workload", runtime), host).Register()
}

// Register will register each capability individually to the Capabilities registry
//
// It sends a POST request to the endpoint in the "HTTPRegistry", if the request
// fails then the request is retried. It uses exponential backoff algorithm to determine
// the interval between in the retries. It will retry only for 10 mins and will stop retrying
// after that.
//
// Register function is a blocking function
func (reg *Registrant) Register() error {
	for _, cpath := range reg.CapabilityDefinitionPaths {
		var capability = Capability{}

		capabilityDef, err := os.Open(cpath)
		if err != nil {
			return ErrOpenOAMDefintionFile(err)
		}
		defer func() {
			_ = capabilityDef.Close()
		}()

		capDefMap := map[string]interface{}{}
		if err := json.NewDecoder(capabilityDef).Decode(&capDefMap); err != nil {
			return ErrJSONMarshal(err)
		}
		capability.CapabilityDefinition = capDefMap
		capability.RegistrantMetadata.Host = reg.Host

		// send request to the register
		backoffOpt := backoff.NewExponentialBackOff()
		backoffOpt.MaxElapsedTime = 10 * time.Minute
		if err := backoff.Retry(func() error {
			contentByt, err := json.Marshal(capDefMap)
			if err != nil {
				return backoff.Permanent(err)
			}
			content := bytes.NewReader(contentByt)

			// host here is given by the application itself and is trustworthy hence,
			// #nosec
			resp, err := http.Post(reg.HTTPRegistry, "application/json", content)
			fmt.Printf("Resp: \n %v \n Err: \n %v \n", resp, err)
			if err != nil {
				return err
			}

			if resp.StatusCode != http.StatusCreated &&
				resp.StatusCode != http.StatusOK &&
				resp.StatusCode != http.StatusAccepted {
				return fmt.Errorf(
					"register process failed, host returned status: %s with status code %d",
					resp.Status,
					resp.StatusCode,
				)
			}

			return nil
		}, backoffOpt); err != nil {
			return ErrOAMRetry(err)
		}
	}

	return nil
}

//StaticCapabilitiesConfig is used to configure CreateCapabilities
type StaticCapabilitiesConfig struct {
	URL     string           //URL
	Method  string           //Use the constants exported by package. Manifests or Helm
	Path    string           //Where to store the directory.(Each directory will have an array of definitions and schemas)
	DirName string           //The directory's name. By convention, it should be the version name
	Config  manifests.Config //Filters required to create definition and schema
	Force   bool             //When set to true, if the file with same name already exists, they will be overridden
}

//CreateComponents generates components for a given configuration and stores them.
func CreateComponents(scfg StaticCapabilitiesConfig) error {
	dir := filepath.Join(scfg.Path, scfg.DirName)
	_, err := os.Stat(dir)
	if err != nil && !os.IsNotExist(err) {
		return ErrCreatingComponents(err)
	}
	if err != nil && os.IsNotExist(err) {
		err = os.Mkdir(dir, 0777)
		if err != nil {
			return ErrCreatingComponents(err)
		}
	}
	var components *([]v1alpha1.Component)
	switch scfg.Method {
	case Manifests:
		components, err = manifests.GetFromManifest(context.Background(), scfg.URL, manifests.SERVICE_MESH, scfg.Config)
	case HelmCHARTS:
		components, err = manifests.GetFromHelm(context.Background(), scfg.URL, manifests.SERVICE_MESH, scfg.Config)
	default:
		return ErrCreatingComponents(errors.New("invalid generation method. Must be either Manifests or HelmCharts"))
	}
	if err != nil {
		return ErrCreatingComponents(err)
	}
	if len(*components) == 0 {
		fmt.Printf("%v\n", *components)
		return ErrCreatingComponents(errors.New("no components found"))
	}
	for _, comp := range *components {
		name := strings.ToLower(comp.Metadata["name"].(string))
		defFileName := name + "_component_definition.json"
		compJson, err := json.Marshal(comp)
		if err != nil {
			return ErrCreatingComponents(err)
		}
		err = writeToFile(filepath.Join(dir, defFileName), []byte(compJson), scfg.Force)
		if err != nil {
			return ErrCreatingComponents(err)
		}
	}
	return nil
}

//create a file with this filename and stuff the string
func writeToFile(path string, data []byte, force bool) error {
	_, err := os.Stat(path)
	if err != nil && !os.IsNotExist(err) { //There some other error than non existence of file
		return err
	}

	if err == nil { //file already exists
		if !force { // Dont override existing file, skip it
			fmt.Println("File already exists,skipping...")
			return nil
		}
		err := os.Remove(path) //Remove the existing file, before overriding it
		if err != nil {
			return err
		}
	}
	return ioutil.WriteFile(path, data, 0777)
}
