package adapter

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/meshery/meshkit/models/meshmodel/registry"
)

func TestRegisterModernMeshModelComponentFiles(t *testing.T) {
	requests := make(chan registry.MeshModelRegistrantData, 1)

	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("unexpected method: %s", r.Method)
			}

			if r.URL.Path != "/api/registry/components" {
				t.Errorf("unexpected endpoint: %s", r.URL.Path)
			}

			var payload registry.MeshModelRegistrantData

			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Error(err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			select {
			case requests <- payload:
			default:
				t.Error("unexpected duplicate registration request")
			}

			w.WriteHeader(http.StatusCreated)
		},
	))
	defer server.Close()

	ctx, cancel := context.WithTimeout(
		context.Background(), 5*time.Second,
	)
	defer cancel()

	path := filepath.Join(
		"testdata",
		testLinkerdName,
		testLinkerdCoreFilename,
	)

	err := RegisterModernMeshModelComponentFiles(
		ctx,
		testContextID,
		server.URL,
		testAdapterHostname,
		"10005",
		[]string{path},
	)
	if err != nil {
		t.Fatal(err)
	}

	select {
	case payload := <-requests:
		assertRegistrationConnection(t, payload)

		assertRegisteredCoreComponent(t, payload.Entity)

	default:
		t.Fatal("no registration request received")
	}
}

func assertRegistrationConnection(t *testing.T, payload registry.MeshModelRegistrantData) {
	t.Helper()

	if payload.EntityType != componentEntityType {
		t.Fatal("incorrect entity type")
	}

	if payload.Connection.Metadata["host_name"] != testAdapterHostname {
		t.Fatal("incorrect adapter hostname")
	}

	if payload.Connection.Metadata["port"] != float64(10005) {
		t.Fatal("incorrect adapter port")
	}

	if payload.Connection.Metadata["context_id"] != testContextID {
		t.Fatal("incorrect context ID")
	}

}

func assertRegisteredCoreComponent(t *testing.T, data []byte) {
	t.Helper()

	var definition struct {
		SchemaVersion string `json:"schemaVersion"`
		Component     struct {
			Kind    string `json:"kind"`
			Version string `json:"version"`
		} `json:"component"`
	}

	if err := json.Unmarshal(data, &definition); err != nil {
		t.Fatal(err)
	}

	if definition.SchemaVersion == "" {
		t.Fatal("missing component schema version")
	}

	if definition.Component.Kind != "LinkerdMesh" {
		t.Fatal("incorrect registered component")
	}

	if definition.Component.Version != "core.meshmodel.dev/v1alpha1" {
		t.Fatal("incorrect component API version")
	}

}
