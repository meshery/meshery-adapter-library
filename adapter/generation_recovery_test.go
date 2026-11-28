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

	generatedFilename := assertIncompleteGeneration(t, destination)
	marker := filepath.Join(destination, incompleteGenerationMarker)

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

	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatalf("expected marker removal after successful retry, got %v", err)
	}

	copied, err := os.ReadFile(filepath.Join(destination, filename))
	if err != nil {
		t.Fatal(err)
	}

	assertPreservedCoreComponent(t, copied, target)

	assertRecoveryPreservesExistingFiles(
		t, resources, cfg, destination, filename,
		generatedFilename, copied, target,
	)
}

func assertIncompleteGeneration(t *testing.T, destination string) string {
	t.Helper()

	entries, err := os.ReadDir(destination)
	if err != nil {
		t.Fatal(err)
	}

	if len(entries) != 2 {
		t.Fatalf("expected generated file and marker, got %d entries", len(entries))
	}

	marker := filepath.Join(destination, incompleteGenerationMarker)
	if _, err := os.Stat(marker); err != nil {
		t.Fatalf("missing incomplete generation marker: %v", err)
	}

	var generatedFilename string
	for _, entry := range entries {
		if entry.Name() != incompleteGenerationMarker {
			generatedFilename = entry.Name()
		}
	}

	if generatedFilename == "" {
		t.Fatal("missing generated component")
	}

	return generatedFilename
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

func TestPreviousComponentVersionSkipsIncomplete(t *testing.T) {
	root := t.TempDir()
	complete := defaultDefinitionVersion
	incomplete := "1.1.0"
	target := "1.2.0"

	for _, version := range []string{
		complete, incomplete, target,
	} {
		if err := os.MkdirAll(
			filepath.Join(root, version), 0755,
		); err != nil {
			t.Fatal(err)
		}
	}

	marker := filepath.Join(
		root, incomplete, incompleteGenerationMarker,
	)

	if err := os.WriteFile(marker, nil, 0644); err != nil {
		t.Fatal(err)
	}

	version, err := previousComponentVersion(root, target)
	if err != nil {
		t.Fatal(err)
	}

	if version != complete {
		t.Fatalf("expected eligible source %q, got %q",
			complete, version)
	}

	if err := os.Remove(marker); err != nil {
		t.Fatal(err)
	}

	version, err = previousComponentVersion(root, target)
	if err != nil {
		t.Fatal(err)
	}

	if version != incomplete {
		t.Fatalf("expected newly eligible source %q, got %q",
			incomplete, version)
	}
}

func TestModernGenerationSkipsIncompleteSource(t *testing.T) {
	root := t.TempDir()

	complete := filepath.Join(root, defaultDefinitionVersion)
	incomplete := filepath.Join(root, "1.1.0")
	target := "1.2.0"

	core := prepareIncompleteRecoverySources(t, complete, incomplete)

	cfg := StaticCompConfig{
		MeshModelPath: root,
		DirName:       target,
	}

	generated := map[string][]byte{
		"example.v1_meshmodel.json": []byte(`{}`),
	}

	if err := persistModernComponentFiles(generated, cfg); err != nil {
		t.Fatal(err)
	}

	destination := filepath.Join(root, target)

	expected, err := modifyMeshmodelVersionInDefinition(core, target)
	if err != nil {
		t.Fatal(err)
	}

	assertAtomicFileContent(t,
		filepath.Join(destination, testLinkerdCoreFilename),
		expected,
	)

	if _, err := os.Stat(filepath.Join(
		destination, incompleteGenerationMarker,
	)); !os.IsNotExist(err) {
		t.Fatalf("expected completed generation, got %v", err)
	}

	if _, err := os.Stat(filepath.Join(
		incomplete, incompleteGenerationMarker,
	)); err != nil {
		t.Fatalf("incomplete source marker was modified: %v", err)
	}
}

func prepareIncompleteRecoverySources(t *testing.T, complete, incomplete string) []byte {
	t.Helper()

	for _, dir := range []string{complete, incomplete} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
	}

	core, err := os.ReadFile(filepath.Join(
		"testdata", "linkerd", testLinkerdCoreFilename,
	))
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(
		filepath.Join(complete, testLinkerdCoreFilename),
		core, 0644,
	); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(
		filepath.Join(incomplete, incompleteGenerationMarker),
		nil, 0644,
	); err != nil {
		t.Fatal(err)
	}

	// A non-component directory must not interfere with core copying.
	if err := os.Mkdir(
		filepath.Join(complete, "temporary"), 0755,
	); err != nil {
		t.Fatal(err)
	}

	return core
}
