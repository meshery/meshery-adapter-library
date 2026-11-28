package adapter

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/layer5io/meshkit/utils/manifests"
)

const pipelineTestCRD = `
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: serviceprofiles.linkerd.io
spec:
  group: linkerd.io
  scope: Namespaced
  names:
    kind: ServiceProfile
    plural: serviceprofiles
  versions:
    - name: v1alpha1
      served: true
      storage: true
      schema:
        openAPIV3Schema:
          type: object
          properties:
            spec:
              type: object
`

func TestGenerateModernComponentFiles(t *testing.T) {
	root := t.TempDir()

	cfg := StaticCompConfig{
		MeshModelPath: root,
		DirName:       testLinkerdRelease,
		Force:         true,
		MeshModelConfig: MeshModelConfig{
			Category: testModelCategory,
		},
		Config: manifests.Config{
			Name: testLinkerdName,
			Type: testLinkerdType,
		},
	}

	err := generateModernComponentFiles(
		[]string{pipelineTestCRD}, cfg,
	)
	if err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(
		root,
		cfg.DirName,
		"serviceprofile.Linkerd.meshery.layer5.io_meshmodel.json",
	)

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	var definition struct {
		SchemaVersion string `json:"schemaVersion"`
		Component     struct {
			Kind    string `json:"kind"`
			Version string `json:"version"`
			Schema  string `json:"schema"`
		} `json:"component"`
	}

	if err := json.Unmarshal(data, &definition); err != nil {
		t.Fatal(err)
	}

	if definition.SchemaVersion == "" {
		t.Fatal("missing schema version")
	}

	if definition.Component.Kind != testServiceProfileKind {
		t.Fatal("incorrect component kind")
	}

	if definition.Component.Version != testServiceProfileAPIVersion {
		t.Fatal("incorrect component API version")
	}

	if !json.Valid([]byte(definition.Component.Schema)) {
		t.Fatal("invalid generated schema")
	}
}

func TestGenerateModernComponentFilesCollision(t *testing.T) {
	root := t.TempDir()

	cfg := StaticCompConfig{
		MeshModelPath: root,
		DirName:       testLinkerdRelease,
		Config: manifests.Config{
			Name: testLinkerdName,
			Type: testLinkerdType,
		},
	}

	err := generateModernComponentFiles(
		[]string{pipelineTestCRD, pipelineTestCRD},
		cfg,
	)

	if err == nil {
		t.Fatal("expected filename collision error")
	}

	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}

	if len(entries) != 0 {
		t.Fatal("collision should be detected before writing files")
	}
}

func TestGenerateModernComponentFilesPreservesCore(t *testing.T) {
	root := t.TempDir()
	oldVersion := testLinkerdRelease
	newVersion := testLinkerdNextRelease
	filename := testLinkerdCoreFilename

	sourceDir := filepath.Join(root, oldVersion)

	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatal(err)
	}

	original, err := os.ReadFile(
		filepath.Join("testdata", testLinkerdName, filename),
	)
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(
		filepath.Join(sourceDir, filename),
		original,
		0644,
	); err != nil {
		t.Fatal(err)
	}

	cfg := StaticCompConfig{
		MeshModelPath: root,
		DirName:       newVersion,
		Force:         true,
		MeshModelConfig: MeshModelConfig{
			Category: testModelCategory,
		},
		Config: manifests.Config{
			Name: testLinkerdName,
			Type: testLinkerdType,
		},
	}

	if err := generateModernComponentFiles(
		[]string{pipelineTestCRD}, cfg,
	); err != nil {
		t.Fatal(err)
	}

	copied, err := os.ReadFile(
		filepath.Join(root, newVersion, filename),
	)
	if err != nil {
		t.Fatal(err)
	}

	assertPreservedCoreComponent(t, copied, newVersion)

	unchanged, err := os.ReadFile(
		filepath.Join(sourceDir, filename),
	)
	if err != nil {
		t.Fatal(err)
	}

	if string(original) != string(unchanged) {
		t.Fatal("original core component was modified")
	}
}

func TestCreateModernComponentsFromSources(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			resource := pipelineTestCRD

			if r.URL.Path == "/server" {
				resource = strings.ReplaceAll(
					resource, testServiceProfileKind, "Server",
				)
				resource = strings.ReplaceAll(
					resource, "serviceprofiles", "servers",
				)
			}

			_, _ = w.Write([]byte(resource))
		},
	))
	defer server.Close()

	cfg := StaticCompConfig{
		Method:        Manifests,
		MeshModelPath: t.TempDir(),
		DirName:       testLinkerdRelease,
		Force:         true,
		Config: manifests.Config{
			Name: testLinkerdName,
			Type: testLinkerdType,
			ExtractCrds: func(data string) []string {
				return strings.Split(data, "---")
			},
		},
	}

	t.Run("multiple sources", func(t *testing.T) {
		err := CreateModernComponentsFromSources(
			[]string{
				server.URL + "/serviceprofile",
				server.URL + "/server",
			},
			cfg,
		)
		if err != nil {
			t.Fatal(err)
		}

		files, err := os.ReadDir(
			filepath.Join(cfg.MeshModelPath, cfg.DirName),
		)
		if err != nil {
			t.Fatal(err)
		}

		if len(files) != 2 {
			t.Fatalf("expected 2 components, got %d", len(files))
		}
	})

	t.Run("cross-source collision", func(t *testing.T) {
		collisionCfg := cfg
		collisionCfg.MeshModelPath = t.TempDir()

		err := CreateModernComponentsFromSources(
			[]string{
				server.URL + "/serviceprofile",
				server.URL + "/duplicate",
			},
			collisionCfg,
		)

		if err == nil ||
			!strings.Contains(err.Error(), "filename collision") {
			t.Fatalf("expected collision error, got %v", err)
		}

		files, err := os.ReadDir(collisionCfg.MeshModelPath)
		if err != nil {
			t.Fatal(err)
		}

		if len(files) != 0 {
			t.Fatal("collision must be detected before writing")
		}
	})
}

func assertPreservedCoreComponent(t *testing.T, copied []byte, newVersion string) {
	t.Helper()

	var definition struct {
		Kind       string `json:"kind"`
		APIVersion string `json:"apiVersion"`
		Schema     string `json:"schema"`
		Model      struct {
			Version string `json:"version"`
		} `json:"model"`
	}

	if err := json.Unmarshal(copied, &definition); err != nil {
		t.Fatal(err)
	}

	if definition.Kind != "LinkerdMesh" {
		t.Fatal("core component kind changed")
	}

	if definition.APIVersion != "core.meshmodel.dev/v1alpha1" {
		t.Fatal("core component API version changed")
	}

	if definition.Model.Version != newVersion {
		t.Fatalf("expected version %s, got %s",
			newVersion, definition.Model.Version)
	}

	if !json.Valid([]byte(definition.Schema)) {
		t.Fatal("core component schema is invalid")
	}

}
