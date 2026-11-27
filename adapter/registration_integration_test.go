package adapter

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestModernComponentRegistrationHTTPContract(t *testing.T) {
	component := []byte(`{
		"schemaVersion": "components.meshery.io/v1beta3",
		"displayName": "Linkerd ServiceProfile"
	}`)

	connection := []byte(`{
		"kind": "linkerd"
	}`)

	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			assertRegistrationHTTPRequest(t, r)

			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Errorf("read request: %v", err)
				http.Error(w, "invalid request", http.StatusBadRequest)
				return
			}

			var payload struct {
				Connection map[string]interface{} `json:"connection"`
				EntityType string                 `json:"entityType"`
				Entity     []byte                 `json:"entity"`
			}

			if err := json.Unmarshal(body, &payload); err != nil {
				t.Errorf("decode request: %v", err)
				http.Error(w, "invalid JSON", http.StatusBadRequest)
				return
			}

			if payload.EntityType != componentEntityType {
				t.Errorf("unexpected entity type: %s", payload.EntityType)
			}

			if payload.Connection["kind"] != testLinkerdName {
				t.Errorf("unexpected connection: %v", payload.Connection)
			}

			var decodedComponent map[string]interface{}
			if err := json.Unmarshal(payload.Entity, &decodedComponent); err != nil {
				t.Errorf("decode component: %v", err)
			}

			if decodedComponent["displayName"] != "Linkerd ServiceProfile" {
				t.Errorf("unexpected component: %v", decodedComponent)
			}

			w.WriteHeader(http.StatusCreated)
		},
	))
	defer server.Close()

	payload, err := newComponentRegistrationPayload(connection, component)
	if err != nil {
		t.Fatalf("create registration payload: %v", err)
	}

	client := newRegistrationHTTPClient(server.Client())

	err = client.post(
		context.Background(),
		server.URL+"/api/registry/components",
		payload,
	)
	if err != nil {
		t.Fatalf("register component: %v", err)
	}
}

func assertRegistrationHTTPRequest(t *testing.T, r *http.Request) {
	t.Helper()

	if r.URL.Path != "/api/registry/components" {
		t.Errorf("unexpected endpoint: %s", r.URL.Path)
	}

	if r.Method != http.MethodPost {
		t.Errorf("unexpected method: %s", r.Method)
	}

	if r.Header.Get("Content-Type") != "application/json" {
		t.Errorf("unexpected content type")
	}

}
