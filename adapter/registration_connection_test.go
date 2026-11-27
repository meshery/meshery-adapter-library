package adapter

import (
	"encoding/json"
	"testing"

	"github.com/meshery/meshkit/models/meshmodel/registry"
	"github.com/meshery/schemas/models/v1beta3/connection"
)

func TestAdapterConnectionRouting(t *testing.T) {
	data, err := newAdapterConnection(
		testAdapterHostname, 10005, testContextID,
	)
	if err != nil {
		t.Fatal(err)
	}

	var conn connection.Connection
	if err := json.Unmarshal(data, &conn); err != nil {
		t.Fatal(err)
	}

	legacy := registry.RegistrantHostToV1beta1(conn)
	restored := registry.RegistrantHostToV1beta3(legacy)

	if restored.Metadata["host_name"] != testAdapterHostname {
		t.Fatal("adapter hostname was not preserved")
	}

	if restored.Metadata["port"] != float64(10005) {
		t.Fatal("adapter port was not preserved")
	}

	if restored.Metadata["context_id"] != testContextID {
		t.Fatal("context ID was not preserved")
	}

	id1, err := legacy.GenerateID()
	if err != nil {
		t.Fatal(err)
	}

	id2, err := legacy.GenerateID()
	if err != nil {
		t.Fatal(err)
	}

	if id1 != id2 {
		t.Fatal("connection identity is not deterministic")
	}
}
