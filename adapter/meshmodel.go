package adapter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	backoff "github.com/cenkalti/backoff/v4"
	componentdef "github.com/meshery/schemas/models/v1beta1/component"
)

// MeshModelRegistrantDefinitionPath - Structure for configuring registrant paths
type MeshModelRegistrantDefinitionPath struct {
	// EntityDefinitionPath holds the path for Entity Definition file
	EntityDefintionPath string
}

// MeshModel provides utility functions for registering
// MeshModel components to a registry in a reliable way
type MeshModelRegistrant struct {
	Paths        []MeshModelRegistrantDefinitionPath
	HTTPRegistry string
}

// NewMeshModelRegistrant returns an instance of NewMeshModelRegistrant
func NewMeshModelRegistrant(paths []MeshModelRegistrantDefinitionPath, HTTPRegistry string) *MeshModelRegistrant {
	return &MeshModelRegistrant{
		Paths:        paths,
		HTTPRegistry: HTTPRegistry,
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
func (or *MeshModelRegistrant) Register(ctxID string) error {
	for _, dpath := range or.Paths {
		definition, err := os.Open(dpath.EntityDefintionPath)
		if err != nil {
			return ErrOpenOAMDefintionFile(err)
		}
		var cd componentdef.ComponentDefinition
		if err := json.NewDecoder(definition).Decode(&cd); err != nil {
			_ = definition.Close()
			return ErrJSONMarshal(err)
		}
		_ = definition.Close()
		entityBytes, _ := json.Marshal(cd)
		backoffOpt := backoff.NewExponentialBackOff()
		backoffOpt.MaxElapsedTime = 10 * time.Minute
		if err := backoff.Retry(func() error {
			resp, err := http.Post(or.HTTPRegistry, "application/json", bytes.NewBuffer(entityBytes))
			if err != nil {
				return err
			}
			defer resp.Body.Close()
			if resp.StatusCode < 200 || resp.StatusCode >= 300 {
				return fmt.Errorf("registration failed with status: %s", resp.Status)
			}
			return nil
		}, backoffOpt); err != nil {
			return err
		}
	}
	return nil
}
