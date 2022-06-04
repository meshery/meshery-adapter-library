package adapter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	backoff "github.com/cenkalti/backoff/v4"
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
	ComponentPath string
	// Host is the address of the gRPC host capable of processing the request
	Host string
	// Restricted should be set to true if this capability should be restricted
	// only to the server and shouldn't be exposed to the user for direct usage
	Restricted bool
	// Metadata is the other data which can be attached to the post request body
	//
	// Metadata like name of the capability, etc.
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

		component, err := os.Open(dpath.ComponentPath)
		if err != nil {
			return ErrOpenOAMDefintionFile(err)
		}
		defer func() {
			_ = component.Close()
		}()

		compMap := map[string]interface{}{}
		if err := json.NewDecoder(component).Decode(&compMap); err != nil {
			return ErrJSONMarshal(err)
		}

		// formatTitleInOAMRefSchema(&schema)

		// ord.OAMRefSchema = string(schema)

		// ord.Host = dpath.Host
		// ord.Metadata = dpath.Metadata
		// ord.Restricted = dpath.Restricted

		// send request to the register
		backoffOpt := backoff.NewExponentialBackOff()
		backoffOpt.MaxElapsedTime = 10 * time.Minute
		if err := backoff.Retry(func() error {
			contentByt, err := json.Marshal(compMap)
			if err != nil {
				return backoff.Permanent(err)
			}
			content := bytes.NewReader(contentByt)

			// host here is given by the application itself and is trustworthy hence,
			// #nosec
			resp, err := http.Post(or.OAMHTTPRegistry, "application/json", content)
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

//StaticCompConfig is used to configure CreateComponents
type StaticCompConfig struct {
	URL     string           //URL
	Method  string           //Use the constants exported by package. Manifests or Helm
	Path    string           //Where to store the directory.(Each directory will have an array of definitions and schemas)
	DirName string           //The directory's name. By convention, it should be the version name
	Config  manifests.Config //Filters required to create definition and schema
	Force   bool             //When set to true, if the file with same name already exists, they will be overridden
}

//This will be depracated once all adapters migrate to new method of component creation( using static config) and registeration
type DynamicComponentsConfig struct {
	TimeoutInMinutes time.Duration
	URL              string
	GenerationMethod string
	Config           manifests.Config
	Operation        string
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
