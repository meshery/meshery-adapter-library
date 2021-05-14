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
	// paths and will register them to the OAM regsitry
	Paths []OAMRegistrantDefintionPath

	// OAMHTTPRegistry is the address of an OAM registry
	OAMHTTPRegistry string
}

// OAMRegistrantDefintionPath - Structure for configuring registrant paths
type OAMRegistrantDefintionPath struct {
	// OAMDefinitionPath holds the path for OAM Defintion file
	OAMDefintionPath string
	// OAMRefSchemaPath holds the path for the OAM Ref Schema file
	OAMRefSchemaPath string
	// Host is the address of the gRPC host capabale of processing the request
	Host string
}

// OAMRegistrantData struct defines the body of the POST request that is sent to the OAM
// registry (Meshery)
//
// The body contains the
// 1. OAM definition, which is in accordance with the OAM spec
// 2. OAMRefSchema, which is json schema draft-4, draft-7 or draft-8 for the corresponding OAM object
// 3. Host is this service's grpc address in the form of `hostname:port`
type OAMRegistrantData struct {
	OAMDefinition interface{} `json:"oam_definition,omitempty"`
	OAMRefSchema  string      `json:"oam_ref_schema,omitempty"`
	Host          string      `json:"host,omitempty"`
}

// NewOAMRegistrant returns an instance of OAMRegistrant
func NewOAMRegistrant(paths []OAMRegistrantDefintionPath, oamHTTPRegistry string) *OAMRegistrant {
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

		schema, err := os.ReadFile(dpath.OAMRefSchemaPath)
		if err != nil {
			return ErrOpenOAMRefFile(err)
		}
		ord.OAMRefSchema = string(schema)

		ord.Host = dpath.Host

		// send request to the register
		backoffOpt := backoff.NewExponentialBackOff()
		backoffOpt.MaxElapsedTime = 10 * time.Minute
		backoff.Retry(func() error {
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
		}, backoffOpt)
	}

	return nil
}
