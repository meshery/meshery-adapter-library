package adapter

import (
	"os"
	"strings"
	"testing"

	"github.com/layer5io/meshkit/utils/manifests"
)

func TestModernCollisionDiagnostics(t *testing.T) {
	cases := []struct {
		name     string
		resource string
		version  string
	}{
		{
			name:     "identical resources",
			resource: pipelineTestCRD,
			version:  "linkerd.io/v1alpha1",
		},
		{
			name: "different API groups",
			resource: strings.ReplaceAll(
				pipelineTestCRD, "linkerd.io", "example.io",
			),
			version: "example.io/v1alpha1",
		},
		{
			name: "different API versions",
			resource: strings.ReplaceAll(
				pipelineTestCRD, "v1alpha1", "v1beta1",
			),
			version: "linkerd.io/v1beta1",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
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
				[]string{pipelineTestCRD, tc.resource},
				cfg,
			)

			if err == nil {
				t.Fatal("expected component filename collision")
			}

			expected := []string{
				"component filename collision",
				"serviceprofile.Linkerd.meshery.layer5.io_meshmodel.json",
				`resource #1 (kind="ServiceProfile", apiVersion="linkerd.io/v1alpha1")`,
				`resource #2 (kind="ServiceProfile", apiVersion="` +
					tc.version + `")`,
			}

			for _, fragment := range expected {
				if !strings.Contains(err.Error(), fragment) {
					t.Errorf("missing %q in error: %v", fragment, err)
				}
			}

			entries, err := os.ReadDir(root)
			if err != nil {
				t.Fatal(err)
			}

			if len(entries) != 0 {
				t.Fatal("collision must be detected before writing files")
			}
		})
	}
}
