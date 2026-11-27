package adapter

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/meshery/schemas/models/v1beta3/component"
)

func TestMigrateLegacyLinkerdComponents(t *testing.T) {
	files, err := filepath.Glob("testdata/linkerd/*.json")
	if err != nil {
		t.Fatal(err)
	}

	if len(files) != 2 {
		t.Fatalf("expected 2 fixtures, got %d", len(files))
	}

	for _, path := range files {
		t.Run(filepath.Base(path), func(t *testing.T) {
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}

			var legacy legacyComponentDefinition
			if err := json.Unmarshal(data, &legacy); err != nil {
				t.Fatal(err)
			}

			migrated, err := migrateLegacyComponent(data)
			if err != nil {
				t.Fatal(err)
			}

			var modern component.ComponentDefinition
			if err := json.Unmarshal(migrated, &modern); err != nil {
				t.Fatal(err)
			}

			assertMigratedStyles(t, modern, legacy)
			assertMigratedIdentity(t, modern, legacy)
			assertMigratedModel(t, modern, legacy)
		})
	}
}

func assertMigratedStyles(t *testing.T, modern component.ComponentDefinition, legacy legacyComponentDefinition) {
	t.Helper()

	if modern.Styles == nil {
		t.Fatal("component styles are missing")
	}

	if modern.Styles.PrimaryColor != legacy.Metadata["primaryColor"] {
		t.Fatal("primary color changed")
	}

	if modern.Styles.SvgColor != legacy.Metadata["svgColor"] {
		t.Fatal("colored SVG changed")
	}

	if modern.Styles.SvgWhite != legacy.Metadata["svgWhite"] {
		t.Fatal("white SVG changed")
	}

	if modern.Styles.Shape == nil ||
		*modern.Styles.Shape != "ellipse" {
		t.Fatal("incorrect modern shape")
	}

}

func assertMigratedIdentity(t *testing.T, modern component.ComponentDefinition, legacy legacyComponentDefinition) {
	t.Helper()

	if modern.Component.Kind != legacy.Kind {
		t.Fatal("component kind changed")
	}

	if modern.Component.Version != legacy.APIVersion {
		t.Fatal("API version changed")
	}

	if modern.Component.Schema != legacy.Schema {
		t.Fatal("original schema changed")
	}

	if string(modern.DisplayName) != legacy.DisplayName {
		t.Fatal("display name changed")
	}

}

func assertMigratedModel(t *testing.T, modern component.ComponentDefinition, legacy legacyComponentDefinition) {
	t.Helper()

	if modern.Model == nil {
		t.Fatal("model is missing")
	}

	if string(modern.Model.Model.Version) != legacy.Model.Version {
		t.Fatal("infrastructure version changed")
	}

	if string(modern.Model.Category.Name) != legacy.Model.Category.Name {
		t.Fatal("category changed")
	}
}
