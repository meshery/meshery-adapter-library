package adapter

import (
	"context"
	"encoding/json"
	"fmt"
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
func NewMeshModelRegistrant(paths []MeshModelRegistrantDefinitionPath, httpRegistry string) *MeshModelRegistrant {
	return &MeshModelRegistrant{
		Paths:        paths,
		HTTPRegistry: httpRegistry,
	}
}

// Register registers capabilities using a background context.
// Existing callers retain their original blocking behavior.
func (or *MeshModelRegistrant) Register(ctxID string) error {
	return or.RegisterWithContext(context.Background(), ctxID)
}

// RegisterWithContext registers capabilities with caller-controlled
// cancellation and bounded exponential retries for each component.
func (or *MeshModelRegistrant) RegisterWithContext(
	ctx context.Context,
	ctxID string,
) error {
	if or == nil {
		return fmt.Errorf("nil meshmodel registrant")
	}

	if ctx == nil {
		return fmt.Errorf("registration context is required")
	}

	client := newRegistrationHTTPClient(nil)

	for _, dpath := range or.Paths {
		if err := ctx.Err(); err != nil {
			return err
		}

		if dpath.Type != types.ComponentDefinition {
			continue
		}

		payload, err := legacyRegistrationPayload(dpath, ctxID)
		if err != nil {
			return err
		}

		if err := or.registerLegacyPayload(ctx, client, payload); err != nil {
			return ErrOAMRetry(err)
		}
	}

	return nil
}

func legacyRegistrationPayload(
	dpath MeshModelRegistrantDefinitionPath,
	ctxID string,
) ([]byte, error) {
	definition, err := os.Open(dpath.EntityDefintionPath)
	if err != nil {
		return nil, ErrOpenOAMDefintionFile(err)
	}
	defer func() {
		_ = definition.Close()
	}()

	var cd v1alpha1.ComponentDefinition
	if err := json.NewDecoder(definition).Decode(&cd); err != nil {
		return nil, ErrJSONMarshal(err)
	}

	entity, err := json.Marshal(cd)
	if err != nil {
		return nil, ErrJSONMarshal(err)
	}

	mrd := registry.MeshModelRegistrantData{
		Host: registry.Host{
			Hostname: dpath.Host,
			Port:     dpath.Port,
			Metadata: ctxID,
		},
		EntityType: dpath.Type,
		Entity:     entity,
	}

	payload, err := json.Marshal(mrd)
	if err != nil {
		return nil, ErrJSONMarshal(err)
	}

	return payload, nil
}

func (or *MeshModelRegistrant) registerLegacyPayload(
	ctx context.Context,
	client *registrationHTTPClient,
	payload []byte,
) error {
	retry := backoff.NewExponentialBackOff()
	retry.MaxElapsedTime = 10 * time.Minute

	return backoff.Retry(func() error {
		if err := ctx.Err(); err != nil {
			return backoff.Permanent(err)
		}

		return classifyRegistrationError(
			client.post(ctx, or.HTTPRegistry, payload),
		)
	}, backoff.WithContext(retry, ctx))
}
