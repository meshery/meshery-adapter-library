package adapter

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAtomicWritePublication(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, testLinkerdCoreFilename)

	original := []byte(`{"version":"original"}`)
	replacement := []byte(`{"version":"replacement"}`)

	if err := writeToFile(path, original, false); err != nil {
		t.Fatal(err)
	}
	assertAtomicFileContent(t, path, original)

	if err := writeToFile(path, replacement, false); err != nil {
		t.Fatal(err)
	}
	assertAtomicFileContent(t, path, original)

	if err := writeToFile(path, replacement, true); err != nil {
		t.Fatal(err)
	}
	assertAtomicFileContent(t, path, replacement)

	assertNoComponentTempFiles(t, root)
}

func TestAtomicWriteFailedPublication(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, testLinkerdCoreFilename)

	if err := os.Mkdir(path, 0755); err != nil {
		t.Fatal(err)
	}

	if err := writeToFile(path, []byte(`{}`), true); err == nil {
		t.Fatal("expected publication failure")
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if !info.IsDir() {
		t.Fatal("existing destination directory was replaced")
	}

	assertNoComponentTempFiles(t, root)
}

func TestAtomicWriteNonForcedPublicationConflict(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, testLinkerdCoreFilename)

	original := []byte(`{"version":"original"}`)

	if err := os.WriteFile(path, original, 0644); err != nil {
		t.Fatal(err)
	}

	// Exercise publication when the destination already exists,
	// simulating another writer winning the publication race.
	if err := publishComponentFile(path, []byte(`{}`), false); err != nil {
		t.Fatal(err)
	}

	assertAtomicFileContent(t, path, original)
	assertNoComponentTempFiles(t, root)
}

func assertAtomicFileContent(t *testing.T, path string, expected []byte) {
	t.Helper()

	actual, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if string(actual) != string(expected) {
		t.Fatalf("expected %q, got %q", expected, actual)
	}
}

func assertNoComponentTempFiles(t *testing.T, root string) {
	t.Helper()

	files, err := filepath.Glob(
		filepath.Join(root, ".component-*.tmp"),
	)
	if err != nil {
		t.Fatal(err)
	}

	if len(files) != 0 {
		t.Fatalf("temporary files remain: %v", files)
	}
}

func TestAtomicWritePreservesCreationPermissions(t *testing.T) {
	root := t.TempDir()
	reference := filepath.Join(root, "reference.json")
	destination := filepath.Join(root, testLinkerdCoreFilename)
	data := []byte(`{}`)

	if err := os.WriteFile(reference, data, 0666); err != nil {
		t.Fatal(err)
	}

	expected, err := os.Stat(reference)
	if err != nil {
		t.Fatal(err)
	}

	for _, force := range []bool{false, true} {
		if err := writeToFile(destination, data, force); err != nil {
			t.Fatal(err)
		}

		actual, err := os.Stat(destination)
		if err != nil {
			t.Fatal(err)
		}

		if actual.Mode().Perm() != expected.Mode().Perm() {
			t.Fatalf("expected permissions %o, got %o",
				expected.Mode().Perm(), actual.Mode().Perm())
		}
	}

	assertNoComponentTempFiles(t, root)
}
