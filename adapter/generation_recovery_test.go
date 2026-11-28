package adapter

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/layer5io/meshkit/utils/manifests"
)

func TestModernGenerationRecoversCoreOnRetry(t *testing.T) {
	root := t.TempDir()
	target := testLinkerdNextRelease
	filename := testLinkerdCoreFilename

	source := filepath.Join(root, testLinkerdRelease)
	if err := os.MkdirAll(source, 0755); err != nil {
		t.Fatal(err)
	}

	sourceFile := filepath.Join(source, filename)

	// Invalid source content forces the initial core copy to fail.
	if err := os.WriteFile(
		sourceFile, []byte(testInvalidJSON), 0644,
	); err != nil {
		t.Fatal(err)
	}

	cfg := StaticCompConfig{
		MeshModelPath: root,
		DirName:       target,
		Force:         false,
		MeshModelConfig: MeshModelConfig{
			Category: testModelCategory,
		},
		Config: manifests.Config{
			Name: testLinkerdName,
			Type: testLinkerdType,
		},
	}

	resources := []string{pipelineTestCRD}

	if err := generateModernComponentFiles(
		resources, cfg,
	); err == nil {
		t.Fatal("expected initial core copy failure")
	}

	destination := filepath.Join(root, target)

	entries, err := os.ReadDir(destination)
	if err != nil {
		t.Fatal(err)
	}

	if len(entries) != 1 {
		t.Fatalf("expected incomplete target, got %d files", len(entries))
	}

	// Repair the source without removing the incomplete target.
	original, err := os.ReadFile(filepath.Join(
		"testdata", testLinkerdName, filename,
	))
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(sourceFile, original, 0644); err != nil {
		t.Fatal(err)
	}

	if err := generateModernComponentFiles(
		resources, cfg,
	); err != nil {
		t.Fatalf("generation retry failed: %v", err)
	}

	copied, err := os.ReadFile(filepath.Join(destination, filename))
	if err != nil {
		t.Fatal(err)
	}

	assertPreservedCoreComponent(t, copied, target)

	assertRecoveryPreservesExistingFiles(
		t, resources, cfg, destination, filename,
		entries[0].Name(), copied, target,
	)
}

func TestPreviousComponentVersionExcludesTarget(t *testing.T) {
	root := t.TempDir()
	target := testLinkerdNextRelease

	for _, version := range []string{
		"edge-26.5.1",
		testLinkerdRelease,
		target,
	} {
		if err := os.MkdirAll(
			filepath.Join(root, version), 0755,
		); err != nil {
			t.Fatal(err)
		}
	}

	if err := os.WriteFile(
		filepath.Join(root, "edge-99.99.99"),
		[]byte("not a directory"),
		0644,
	); err != nil {
		t.Fatal(err)
	}

	version, err := previousComponentVersion(root, target)
	if err != nil {
		t.Fatal(err)
	}

	if version != testLinkerdRelease {
		t.Fatalf("unexpected previous version: %s", version)
	}
}

func TestPreviousComponentVersionFreshInstall(t *testing.T) {
	root := filepath.Join(t.TempDir(), "components")

	version, err := previousComponentVersion(root, testLinkerdRelease)
	if err != nil || version != "" {
		t.Fatalf("unexpected fresh-install result: %q, %v", version, err)
	}
}

func TestPreviousComponentVersionInvalidRoot(t *testing.T) {
	root := filepath.Join(t.TempDir(), "components")

	if err := os.WriteFile(root, []byte("file"), 0644); err != nil {
		t.Fatal(err)
	}

	if _, err := previousComponentVersion(
		root, testLinkerdRelease,
	); err == nil {
		t.Fatal("expected directory lookup error")
	}
}

func assertRecoveryPreservesExistingFiles(
	t *testing.T,
	resources []string,
	cfg StaticCompConfig,
	destination, filename, generatedFilename string,
	copied []byte,
	target string,
) {
	t.Helper()

	// Add valid JSON whitespace to detect unintended overwrites.
	corePath := filepath.Join(destination, filename)
	preservedCore := append(append([]byte(nil), copied...), '\n', '\n')

	if err := os.WriteFile(corePath, preservedCore, 0644); err != nil {
		t.Fatal(err)
	}

	generatedPath := filepath.Join(destination, generatedFilename)
	generated, err := os.ReadFile(generatedPath)
	if err != nil {
		t.Fatal(err)
	}

	preservedGenerated := append(
		append([]byte(nil), generated...), '\n', '\n',
	)

	if err := os.WriteFile(
		generatedPath, preservedGenerated, 0644,
	); err != nil {
		t.Fatal(err)
	}

	// Confirm another retry preserves both existing files.
	if err := generateModernComponentFiles(
		resources, cfg,
	); err != nil {
		t.Fatal(err)
	}

	unchanged, err := os.ReadFile(corePath)
	if err != nil {
		t.Fatal(err)
	}

	if string(preservedCore) != string(unchanged) {
		t.Fatal("existing core component was overwritten")
	}

	unchangedGenerated, err := os.ReadFile(generatedPath)
	if err != nil {
		t.Fatal(err)
	}

	if string(preservedGenerated) != string(unchangedGenerated) {
		t.Fatal("existing generated component was overwritten")
	}

	assertPreservedCoreComponent(t, unchanged, target)
}
