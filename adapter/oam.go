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
	"time"

	backoff "github.com/cenkalti/backoff/v4"
	"github.com/layer5io/meshkit/models/oam/core/v1alpha1"
	"github.com/layer5io/meshkit/utils/manifests"
)

const (
	// OAM Metadata constants
	OAMAdapterNameMetadataKey       = "adapter.meshery.io/name"
	OAMComponentCategoryMetadataKey = "ui.meshery.io/category"

	//Runtime generation methods
	Manifests  = "MANIFESTS"
	HelmCHARTS = "HELM_CHARTS"
)

// ProcessOAM processes OAM components. This is adapter specific and needs to be implemented by each adapter.
func (h *Adapter) ProcessOAM(context.Context, OAMRequest) (string, error) {
	return "", nil
}

// OAMRegistrant provides utility functions for registering
// OAM components to a registry in a reliable way
type OAMRegistrant struct {
	// Paths is a slice for holding the paths of OAMDefitions,
	// OAMRefSchema and Host on the filesystem
	//
	// OAMRegistrant will read the definitions from these
	// paths and will register them to the OAM registry
	Paths []OAMRegistrantDefinitionPath

	// OAMHTTPRegistry is the address of an OAM registry
	OAMHTTPRegistry string
}

// OAMRegistrantDefinitionPath - Structure for configuring registrant paths
type OAMRegistrantDefinitionPath struct {
	// OAMDefinitionPath holds the path for OAM Definition file
	OAMDefintionPath string
	// OAMRefSchemaPath holds the path for the OAM Ref Schema file
	OAMRefSchemaPath string
	// Host is the address of the gRPC host capable of processing the request
	Host string
	// Restricted should be set to true if this capability should be restricted
	// only to the server and shouldn't be exposed to the user for direct usage
	Restricted bool
	// Metadata is the other data which can be attached to the post request body
	//
	// Metadata like name of the component, etc.
	Metadata map[string]string
}

// OAMRegistrantData struct defines the body of the POST request that is sent to the OAM
// registry (Meshery)
//
// The body contains the
// 1. OAM definition, which is in accordance with the OAM spec
// 2. OAMRefSchema, which is json schema draft-4, draft-7 or draft-8 for the corresponding OAM object
// 3. Host is this service's grpc address in the form of `hostname:port`
// 4. Restricted should be set to true if the given capability is meant to be used internally
// 5. Metadata can be a map of key value pairs
type OAMRegistrantData struct {
	OAMDefinition interface{}       `json:"oam_definition,omitempty"`
	OAMRefSchema  string            `json:"oam_ref_schema,omitempty"`
	Host          string            `json:"host,omitempty"`
	Restricted    bool              `json:"restricted,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
}

// NewOAMRegistrant returns an instance of OAMRegistrant
func NewOAMRegistrant(paths []OAMRegistrantDefinitionPath, oamHTTPRegistry string) *OAMRegistrant {
	return &OAMRegistrant{
		Paths:           paths,
		OAMHTTPRegistry: oamHTTPRegistry,
	}
}

// Register will register each capability individually to the OAM Capability registry
//
// It sends a POST request to the endpoint in the "OAMHTTPRegistry", if the request
// fails then the request is retried. It uses exponential backoff algorithm to determine
// the interval between in the retries. It will retry only for 10 mins and will stop retrying
// after that.
//
// Register function is a blocking function
func (or *OAMRegistrant) Register() error {
	for _, dpath := range or.Paths {
		var ord OAMRegistrantData

		definition, err := os.Open(dpath.OAMDefintionPath)
		if err != nil {
			return ErrOpenOAMDefintionFile(err)
		}
		defer func() {
			_ = definition.Close()
		}()

		definitionMap := map[string]interface{}{}
		if err := json.NewDecoder(definition).Decode(&definitionMap); err != nil {
			return ErrJSONMarshal(err)
		}
		ord.OAMDefinition = definitionMap

		schema, err := ioutil.ReadFile(dpath.OAMRefSchemaPath)
		if err != nil {
			return ErrOpenOAMRefFile(err)
		}
		if string(schema) == "" { //since this component is unusable if it doesn't have oam_ref_schema
			continue
		}
		formatTitleInOAMRefSchema(&schema)

		ord.OAMRefSchema = string(schema)

		ord.Host = dpath.Host
		ord.Metadata = dpath.Metadata
		ord.Restricted = dpath.Restricted

		// send request to the register
		backoffOpt := backoff.NewExponentialBackOff()
		backoffOpt.MaxElapsedTime = 10 * time.Minute
		if err := backoff.Retry(func() error {
			contentByt, err := json.Marshal(ord)
			if err != nil {
				return backoff.Permanent(err)
			}
			content := bytes.NewReader(contentByt)

			// host here is given by the application itself and is trustworthy hence,
			// #nosec
			resp, err := http.Post(or.OAMHTTPRegistry, "application/json", content)
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

//StaticCompConfig is used to configure CreateComponents
type StaticCompConfig struct {
	URL     string           //URL
	Method  string           //Use the constants exported by package. Manifests or Helm
	Path    string           //Where to store the directory.(Each directory will have an array of definitions and schemas)
	DirName string           //The directory's name. By convention, it should be the version name
	Config  manifests.Config //Filters required to create definition and schema
	Force   bool             //When set to true, if the file with same name already exists, they will be overridden
}

//CreateComponents generates components for a given configuration and stores them.
func CreateComponents(scfg StaticCompConfig) error {
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
	var comp *manifests.Component
	switch scfg.Method {
	case Manifests:
		comp, err = manifests.GetFromManifest(context.Background(), scfg.URL, manifests.SERVICE_MESH, scfg.Config)
	case HelmCHARTS:
		comp, err = manifests.GetFromHelm(context.Background(), scfg.URL, manifests.SERVICE_MESH, scfg.Config)
	default:
		return ErrCreatingComponents(errors.New("invalid generation method. Must be either Manifests or HelmCharts"))
	}
	if err != nil {
		return ErrCreatingComponents(err)
	}
	if comp == nil {
		return ErrCreatingComponents(errors.New("no components found"))
	}
	for i, def := range comp.Definitions {
		schema := comp.Schemas[i]
		name := getNameFromWorkloadDefinition([]byte(def))
		defFileName := name + "_definition.json"
		schemaFileName := name + ".meshery.layer5io.schema.json"
		err := writeToFile(filepath.Join(dir, defFileName), []byte(def), scfg.Force)
		if err != nil {
			return ErrCreatingComponents(err)
		}
		err = writeToFile(filepath.Join(dir, schemaFileName), []byte(schema), scfg.Force)
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

//getNameFromWorkloadDefinition takes out name from workload definition
func getNameFromWorkloadDefinition(definition []byte) string {
	var wd v1alpha1.WorkloadDefinition
	err := json.Unmarshal(definition, &wd)
	if err != nil {
		return ""
	}
	return wd.Spec.DefinitionRef.Name
}

//This will be depracated once all adapters migrate to new method of component creation( using static config) and registeration
type DynamicComponentsConfig struct {
	TimeoutInMinutes time.Duration
	URL              string
	GenerationMethod string
	Config           manifests.Config
	Operation        string
}

func RegisterWorkLoadsDynamically(runtime, host string, dc *DynamicComponentsConfig) error {
	var comp *manifests.Component
	var err error
	switch dc.GenerationMethod {
	case Manifests:
		comp, err = manifests.GetFromManifest(context.Background(), dc.URL, manifests.SERVICE_MESH, dc.Config)
	case HelmCHARTS:
		comp, err = manifests.GetFromHelm(context.Background(), dc.URL, manifests.SERVICE_MESH, dc.Config)
	default:
		return ErrGenerateComponents(errors.New("failed to generate components"))
	}
	if err != nil {
		return ErrGenerateComponents(err)
	}
	if comp == nil {
		return ErrGenerateComponents(errors.New("failed to generate components"))
	}
	for i, def := range comp.Definitions {
		var ord OAMRegistrantData
		ord.OAMRefSchema = comp.Schemas[i]

		//Marshalling the stringified json
		ord.Host = host
		definitionMap := map[string]interface{}{}
		if err := json.Unmarshal([]byte(def), &definitionMap); err != nil {
			return err
		}
		definitionMap["apiVersion"] = "core.oam.dev/v1alpha1"
		definitionMap["kind"] = "WorkloadDefinition"
		ord.OAMDefinition = definitionMap
		ord.Metadata = map[string]string{
			OAMAdapterNameMetadataKey: dc.Operation,
		}
		// send request to the register
		backoffOpt := backoff.NewExponentialBackOff()
		backoffOpt.MaxElapsedTime = time.Minute * dc.TimeoutInMinutes
		if err := backoff.Retry(func() error {
			contentByt, err := json.Marshal(ord)
			if err != nil {
				return backoff.Permanent(err)
			}
			content := bytes.NewReader(contentByt)
			// host here is given by the application itself and is trustworthy hence,
			// #nosec
			resp, err := http.Post(fmt.Sprintf("%s/api/oam/workload", runtime), "application/json", content)
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
			return err
		}
	}
	return nil
}

func formatTitleInOAMRefSchema(schema *[]byte) {
	var schemamap map[string]interface{}
	err := json.Unmarshal(*schema, &schemamap)
	if err != nil {
		return
	}
	title, ok := schemamap["title"].(string)
	if !ok {
		return
	}

	schemamap["title"] = manifests.FormatToReadableString(title)
	(*schema), err = json.Marshal(schemamap)
	if err != nil {
		return
	}
}
