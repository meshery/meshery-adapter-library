package adapter

import (
	"encoding/json"
	"testing"
)

func TestNewComponentRegistrationPayload(t *testing.T) {
	connection := []byte(`{"kind":"linkerd"}`)
	component := []byte(`{"schemaVersion":"components.meshery.io/v1beta3"}`)

	result, err := newComponentRegistrationPayload(connection, component)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var decoded struct {
		Connection map[string]interface{} `json:"connection"`
		EntityType string                 `json:"entityType"`
		Entity     []byte                 `json:"entity"`
	}

	if err := json.Unmarshal(result, &decoded); err != nil {
		t.Fatalf("invalid registration payload: %v", err)
	}

	if decoded.EntityType != componentEntityType {
		t.Errorf("unexpected entity type: %s", decoded.EntityType)
	}

	if decoded.Connection["kind"] != testLinkerdName {
		t.Errorf("unexpected registrant connection: %v", decoded.Connection)
	}

	if string(decoded.Entity) != string(component) {
		t.Errorf("component bytes changed during serialization")
	}
}

func TestNewComponentRegistrationPayloadRejectsInvalidJSON(t *testing.T) {
	tests := []struct {
		name       string
		connection []byte
		component  []byte
	}{
		{"invalid connection", []byte(`{invalid}`), []byte(`{}`)},
		{"invalid component", []byte(`{}`), []byte(`{invalid}`)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := newComponentRegistrationPayload(tt.connection, tt.component); err == nil {
				t.Fatal("expected invalid JSON error")
			}
		})
	}
}
