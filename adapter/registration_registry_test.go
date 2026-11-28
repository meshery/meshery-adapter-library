package adapter

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/meshery/meshkit/database"
	"github.com/meshery/meshkit/models/meshmodel/registry"

	regv1beta1 "github.com/meshery/meshkit/models/meshmodel/registry/v1beta1"

	"github.com/meshery/schemas/models/v1beta1/model"
	"github.com/meshery/schemas/models/v1beta3/component"
)

func TestModernComponentRegistryPersistence(t *testing.T) {
	db, err := database.New(database.Options{
		Filename: filepath.Join(t.TempDir(), "registry.db"),
		Engine:   "sqlite",
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		_ = db.DBClose()
	})

	rm, err := registry.NewRegistryManager(&db)
	if err != nil {
		t.Fatal(err)
	}

	payload := buildRegistryTestPayload(t)

	definition := registerRegistryTestComponent(t, rm, payload)

	var stored component.ComponentDefinition

	if err := db.Where(
		"id = ?", definition.ID,
	).First(&stored).Error; err != nil {
		t.Fatal(err)
	}

	assertPersistedComponent(t, stored)

	var storedModel model.ModelDefinition

	if err := db.Where(
		"id = ?", *stored.ModelID,
	).First(&storedModel).Error; err != nil {
		t.Fatal(err)
	}

	assertPersistedModel(t, storedModel)

	host := rm.GetRegistrant(&stored)

	assertPersistedRegistrant(t, host.ID, host.Metadata)

	assertRegistryDiscovery(t, rm, stored)

	repeated := assertRepeatRegistration(t, rm, payload, stored, host.ID)

	for _, table := range []string{
		"connections",
		"model_dbs",
		"component_definition_dbs",
		"registries",
	} {
		var count int64

		if err := db.Table(table).Count(&count).Error; err != nil {
			t.Fatal(err)
		}

		t.Logf("Table %s: %d records", table, count)
	}

	t.Logf("Original component ID: %s", stored.ID)
	t.Logf("Repeated component ID: %s", repeated.ID)

	t.Logf("Persisted component ID: %s", stored.ID)
	t.Logf("Persisted model ID: %s", storedModel.ID)
	t.Logf("Persisted registrant ID: %s", host.ID)
	t.Logf("Model status: %q", storedModel.Status)
	t.Logf("Component metadata: %+v", stored.Metadata)
	t.Logf("Registrant metadata: %+v", host.Metadata)
}

func buildRegistryTestPayload(t *testing.T) registry.MeshModelRegistrantData {
	t.Helper()

	data, err := GenerateNativeComponentWithModel(
		pipelineTestCRD,
		testLinkerdName,
		testLinkerdRelease,
		MeshModelConfig{
			Category: testModelCategory,
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	connectionJSON, err := newAdapterConnection(
		testAdapterHostname,
		10005,
		testContextID,
	)
	if err != nil {
		t.Fatal(err)
	}

	body, err := newComponentRegistrationPayload(
		connectionJSON,
		data,
	)
	if err != nil {
		t.Fatal(err)
	}

	var payload registry.MeshModelRegistrantData
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatal(err)
	}

	return payload
}

func registerRegistryTestComponent(
	t *testing.T,
	rm *registry.RegistryManager,
	payload registry.MeshModelRegistrantData,
) component.ComponentDefinition {
	t.Helper()

	var definition component.ComponentDefinition
	if err := json.Unmarshal(payload.Entity, &definition); err != nil {
		t.Fatal(err)
	}

	registrantErr, modelErr, err := rm.RegisterEntity(
		payload.Connection,
		&definition,
	)

	if err != nil || registrantErr || modelErr {
		t.Fatalf(
			"registration failed: registrant=%v model=%v err=%v",
			registrantErr, modelErr, err,
		)
	}

	if definition.ID == uuid.Nil {
		t.Fatal("component ID was not generated")
	}

	return definition
}

func assertPersistedComponent(t *testing.T, stored component.ComponentDefinition) {
	t.Helper()

	if stored.Component.Kind != testServiceProfileKind {
		t.Fatalf("unexpected component: %s", stored.Component.Kind)
	}

	if stored.Component.Version != testServiceProfileAPIVersion {
		t.Fatal("component API version was not preserved")
	}

	if stored.ModelID == nil {
		t.Fatal("persisted component has no model ID")
	}

}

func assertPersistedModel(t *testing.T, storedModel model.ModelDefinition) {
	t.Helper()

	if storedModel.Name != testLinkerdName {
		t.Fatalf("unexpected model: %s", storedModel.Name)
	}

	if storedModel.Model.Version != testLinkerdRelease {
		t.Fatal("Linkerd release version was not preserved")
	}

	if storedModel.Status != model.Enabled {
		t.Fatalf("unexpected model status: %q", storedModel.Status)
	}

}

func assertRegistryDiscovery(
	t *testing.T,
	rm *registry.RegistryManager,
	stored component.ComponentDefinition,
) {
	t.Helper()

	if stored.Status == nil || *stored.Status != component.Enabled {
		t.Fatalf("component status is not enabled: %v", stored.Status)
	}

	entities, count, uniqueCount, err := rm.GetEntities(
		&regv1beta1.ComponentFilter{
			Name:       testServiceProfileKind,
			ModelName:  testLinkerdName,
			APIVersion: testServiceProfileAPIVersion,
			Version:    testLinkerdRelease,
		},
	)

	if err != nil {
		t.Fatal(err)
	}

	if len(entities) != 1 || count != 1 || uniqueCount != 1 {
		t.Fatalf("unexpected discovery result: entities=%d count=%d unique=%d",
			len(entities), count, uniqueCount)
	}

	if entities[0].GetID() != stored.ID {
		t.Fatal("discovered component ID does not match persisted component")
	}

	t.Log("Component discovery through default registry filter verified")

}

func assertPersistedRegistrant(
	t *testing.T,
	id uuid.UUID,
	metadata map[string]interface{},
) {
	t.Helper()

	if id == uuid.Nil {
		t.Fatal("registrant was not retrieved")
	}

	if metadata["host_name"] != testAdapterHostname {
		t.Fatal("adapter hostname was not preserved")
	}

	if metadata["port"] != float64(10005) {
		t.Fatalf("unexpected adapter port: %#v", metadata["port"])
	}

	if metadata["context_id"] != testContextID {
		t.Fatal("context ID was not preserved")
	}

}

func assertRepeatRegistration(
	t *testing.T,
	rm *registry.RegistryManager,
	payload registry.MeshModelRegistrantData,
	stored component.ComponentDefinition,
	hostID uuid.UUID,
) component.ComponentDefinition {
	t.Helper()

	var repeated component.ComponentDefinition
	if err := json.Unmarshal(payload.Entity, &repeated); err != nil {
		t.Fatal(err)
	}

	registrantErr, modelErr, err := rm.RegisterEntity(
		payload.Connection,
		&repeated,
	)

	if err != nil || registrantErr || modelErr {
		t.Fatalf("repeat registration failed: registrant=%v model=%v err=%v",
			registrantErr, modelErr, err)
	}

	repeatedHost := rm.GetRegistrant(&repeated)

	if repeatedHost.ID != hostID {
		t.Fatal("repeat registration created a different registrant")
	}

	if repeated.ModelID == nil || *repeated.ModelID != *stored.ModelID {
		t.Fatal("repeat registration created a different model")
	}

	return repeated
}
