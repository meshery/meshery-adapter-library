package adapter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	backoff "github.com/cenkalti/backoff/v4"
	"github.com/layer5io/meshkit/models/meshmodel/core/types"
	"github.com/layer5io/meshkit/models/meshmodel/core/v1alpha1"
	"github.com/layer5io/meshkit/models/meshmodel/registry"
)

// MeshModelRegistrantDefinitionPath - Structure for configuring registrant paths
type MeshModelRegistrantDefinitionPath struct {
	// EntityDefinitionPath holds the path for Entity Definition file
	EntityDefintionPath string

	Type types.CapabilityType
	// Host is the address of the gRPC host capable of processing the request
	Host string
	Port int
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
		var mrd registry.MeshModelRegistrantData
		definition, err := os.Open(dpath.EntityDefintionPath)
		if err != nil {
			return ErrOpenOAMDefintionFile(err)
		}
		mrd.Host = registry.Host{
			Hostname: dpath.Host,
			Port:     dpath.Port,
			Metadata: ctxID,
		}
		mrd.EntityType = dpath.Type
		switch dpath.Type {
		case types.ComponentDefinition:
			var cd v1alpha1.ComponentDefinition
			if err := json.NewDecoder(definition).Decode(&cd); err != nil {
				_ = definition.Close()
				return ErrJSONMarshal(err)
			}
			_ = definition.Close()
			enbyt, _ := json.Marshal(cd)
			mrd.Entity = enbyt
			// send request to the register
			backoffOpt := backoff.NewExponentialBackOff()
			backoffOpt.MaxElapsedTime = 10 * time.Minute
			if err := backoff.Retry(func() error {
				contentByt, err := json.Marshal(mrd)
				if err != nil {
					return backoff.Permanent(err)
				}
				content := bytes.NewReader(contentByt)

				// host here is given by the application itself and is trustworthy hence,
				// #nosec
				resp, err := http.Post(or.HTTPRegistry, "application/json", content)
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
	}

	return nil
}
