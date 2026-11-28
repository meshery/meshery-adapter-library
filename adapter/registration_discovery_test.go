package adapter

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestLoadMeshmodelComponentsFiltersUnrelatedFiles(t *testing.T) {
	root := t.TempDir()

	validVersion := "discovery-test-valid"
	ignoredVersion := "discovery-test-ignored"

	t.Cleanup(func() {
		versionLock.Lock()
		defer versionLock.Unlock()

		delete(AvailableVersions, validVersion)
		delete(AvailableVersions, ignoredVersion)
	})

	files := []string{
		validVersion + "/serviceprofile.Linkerd.meshery.layer5.io_meshmodel.json",
		validVersion + "/" + testLinkerdCoreFilename,
		validVersion + "/README.md",
		validVersion + "/metadata.json",
		ignoredVersion + "/.gitkeep",
		ignoredVersion + "/temporary.txt",
	}

	createDiscoveryTestFiles(t, root, ignoredVersion, files)

	paths, err := loadMeshmodelComponents(root)
	if err != nil {
		t.Fatal(err)
	}

	var actual []string
	for _, path := range paths {
		actual = append(actual, path.meshmodelDefinitionPath)
	}

	expected := []string{
		filepath.Join(root, files[0]),
		filepath.Join(root, files[1]),
	}

	slices.Sort(actual)
	slices.Sort(expected)

	if !slices.Equal(actual, expected) {
		t.Fatalf("unexpected discovered paths: %v", actual)
	}

	versionLock.Lock()
	defer versionLock.Unlock()

	if !AvailableVersions[validVersion] {
		t.Fatal("component version was not discovered")
	}

	if AvailableVersions[ignoredVersion] {
		t.Fatal("unrelated files incorrectly marked a version available")
	}
}

func TestLoadMeshmodelComponentsMissingDirectory(t *testing.T) {
	root := filepath.Join(t.TempDir(), "missing")

	if _, err := loadMeshmodelComponents(root); err == nil {
		t.Fatal("expected missing directory error")
	}
}

func createDiscoveryTestFiles(
	t *testing.T,
	root, ignoredVersion string,
	files []string,
) {
	t.Helper()

	for _, name := range files {
		path := filepath.Join(root, name)

		if err := os.MkdirAll(
			filepath.Dir(path), 0755,
		); err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile(
			path, []byte("{}"), 0600,
		); err != nil {
			t.Fatal(err)
		}
	}

	// A directory with a component-like name must not be registered.
	if err := os.MkdirAll(filepath.Join(
		root, ignoredVersion, "directory_meshmodel.json",
	), 0755); err != nil {
		t.Fatal(err)
	}

	// Symlinks must not be discovered as regular component files.
	if err := os.Symlink(
		filepath.Join(root, files[0]),
		filepath.Join(root, ignoredVersion, "alias_meshmodel.json"),
	); err != nil {
		t.Fatal(err)
	}

}
