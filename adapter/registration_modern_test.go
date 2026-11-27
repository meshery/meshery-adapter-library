package adapter

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestModernComponentRegistrantRegisterFiles(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "serviceprofile.json")

	component := []byte(`{
		"schemaVersion": "components.meshery.io/v1beta3",
		"displayName": "Linkerd ServiceProfile",
		"model": {
			"name": "linkerd",
			"version": "1.0.0",
			"model": {"version": "1.2.3"}
		},
		"version": "1.0.0",
		"format": "JSON",
		"component": {
			"kind": "ServiceProfile",
			"version": "linkerd.io/v1alpha1",
			"schema": "{\"type\":\"object\"}"
		}
	}`)

	if err := os.WriteFile(path, component, 0600); err != nil {
		t.Fatal(err)
	}

	requests := 0

	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			requests++

			var payload struct {
				EntityType string `json:"entityType"`
				Entity     []byte `json:"entity"`
			}

			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Errorf("decode registration: %v", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			if payload.EntityType != componentEntityType {
				t.Errorf("unexpected entity type: %s", payload.EntityType)
			}

			if string(payload.Entity) != string(component) {
				t.Errorf("unexpected component payload")
			}

			w.WriteHeader(http.StatusCreated)
		},
	))
	defer server.Close()

	registrant := &ModernComponentRegistrant{
		Endpoint:   server.URL + "/api/registry/components",
		Connection: []byte(`{"kind":"linkerd"}`),
		Client:     server.Client(),
	}

	if err := registrant.RegisterFiles(
		context.Background(),
		[]string{path},
	); err != nil {
		t.Fatalf("registration failed: %v", err)
	}

	if requests != 1 {
		t.Errorf("expected 1 request, got %d", requests)
	}
}

func TestModernComponentRegistrantRejectsMissingFiles(t *testing.T) {
	registrant := &ModernComponentRegistrant{
		Endpoint:   "http://localhost:9081/api/registry/components",
		Connection: []byte(`{"kind":"linkerd"}`),
	}

	if err := registrant.RegisterFiles(context.Background(), nil); err == nil {
		t.Fatal("expected missing-files error")
	}
}

func TestModernRegistrantMigratesLegacyFiles(t *testing.T) {
	files, err := filepath.Glob("testdata/linkerd/*.json")
	if err != nil {
		t.Fatal(err)
	}

	if len(files) != 2 {
		t.Fatalf("expected 2 fixtures, got %d", len(files))
	}

	received := make(chan string, len(files))

	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			var payload struct {
				EntityType string `json:"entityType"`
				Entity     []byte `json:"entity"`
			}

			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Errorf("decode payload: %v", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			var definition struct {
				SchemaVersion string `json:"schemaVersion"`
				Component     struct {
					Kind    string `json:"kind"`
					Version string `json:"version"`
					Schema  string `json:"schema"`
				} `json:"component"`
			}

			if err := json.Unmarshal(payload.Entity, &definition); err != nil {
				t.Errorf("decode component: %v", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			if payload.EntityType != componentEntityType {
				t.Error("incorrect entity type")
			}

			if definition.SchemaVersion == "" {
				t.Error("missing schema version")
			}

			if definition.Component.Version == "" {
				t.Error("missing API version")
			}

			if !json.Valid([]byte(definition.Component.Schema)) {
				t.Error("invalid component schema")
			}

			received <- definition.Component.Kind
			w.WriteHeader(http.StatusCreated)
		},
	))
	defer server.Close()

	registrant := &ModernComponentRegistrant{
		Endpoint:   server.URL + "/api/registry/components",
		Connection: []byte(`{"kind":"linkerd"}`),
		Client:     server.Client(),
	}

	if err := registrant.RegisterFiles(context.Background(), files); err != nil {
		t.Fatal(err)
	}

	assertReceivedLegacyKinds(t, received)
}

func TestPrepareComponentRejectsInvalidDefinitions(t *testing.T) {
	cases := []string{
		`invalid JSON`,
		`{}`,
		nullResource,
		`{"kind":"ServiceProfile"}`,
		`{"schemaVersion":"v1beta3","component":{}}`,
		`{"schemaVersion":"v1beta3","component":null}`,
		`{"schemaVersion":"v1beta3","component":{"kind":"ServiceProfile","version":"v1","schema":"{}"}}`,
		`{"schemaVersion":"v1beta3","model":null,"component":{"kind":"ServiceProfile","version":"v1","schema":"{}"}}`,
		`{"schemaVersion":"v1beta3","model":{},"component":{"kind":"ServiceProfile","version":"v1","schema":"{}"}}`,
		`{"schemaVersion":"v1beta3","model":{"name":"linkerd","model":{"version":"1.2.3"}},"component":{"kind":"ServiceProfile","version":"v1","schema":"invalid"}}`,
		`{"schemaVersion":"v1beta3","model":{"name":"linkerd","model":{"version":"1.2.3"}},"component":{"kind":"ServiceProfile","version":"v1"}}`,
	}

	for _, data := range cases {
		if _, err := prepareComponentForRegistration([]byte(data)); err == nil {
			t.Errorf("expected rejection for %s", data)
		}
	}
}

func TestPrepareComponentAllowsAnnotationWithoutSchema(t *testing.T) {
	data := []byte(`{
		"schemaVersion": "components.meshery.io/v1beta3",
		"model": {
			"name": "linkerd",
			"version": "1.0.0",
			"model": {"version": "1.2.3"}
		},
		"metadata": {
			"isAnnotation": true
		},
		"component": {
			"kind": "Annotation",
			"version": "v1"
		}
	}`)

	result, err := prepareComponentForRegistration(data)
	if err != nil {
		t.Fatalf("valid annotation rejected: %v", err)
	}

	if string(result) != string(data) {
		t.Fatal("original annotation payload changed")
	}
}

func assertReceivedLegacyKinds(t *testing.T, received chan string) {
	t.Helper()

	close(received)

	kinds := make(map[string]bool)
	for kind := range received {
		kinds[kind] = true
	}

	if !kinds["LinkerdMesh"] || !kinds[testServiceProfileKind] {
		t.Fatalf("missing migrated components: %v", kinds)
	}
}
