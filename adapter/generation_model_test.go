package adapter

import "testing"

func TestNewGeneratedModel(t *testing.T) {
	cfg := MeshModelConfig{
		Category: testModelCategory,
		Metadata: map[string]interface{}{
			"primaryColor":   testLinkerdColor,
			"secondaryColor": testLinkerdColor,
			"shape":          "circle",
			"svgColor":       "<svg>color</svg>",
			"svgWhite":       "<svg>white</svg>",
		},
	}

	m, err := newGeneratedModel(testLinkerdType, testLinkerdRelease, cfg)
	if err != nil {
		t.Fatal(err)
	}

	if m.Name != testLinkerdName {
		t.Fatalf("unexpected model name: %s", m.Name)
	}

	if m.Model.Version != testLinkerdRelease {
		t.Fatalf("unexpected infrastructure version: %s", m.Model.Version)
	}

	if m.Category.Name != testModelCategory {
		t.Fatal("model category was not preserved")
	}

	if m.Metadata.PrimaryColor == nil ||
		*m.Metadata.PrimaryColor != testLinkerdColor {
		t.Fatal("primary color was not preserved")
	}

	if m.Metadata.SvgColor != "<svg>color</svg>" ||
		m.Metadata.SvgWhite != "<svg>white</svg>" {
		t.Fatal("SVG metadata was not preserved")
	}
}

func TestNewGeneratedModelInvalidInput(t *testing.T) {
	for _, input := range [][2]string{
		{"", "1.0.0"},
		{testLinkerdName, ""},
	} {
		if _, err := newGeneratedModel(
			input[0], input[1], MeshModelConfig{},
		); err == nil {
			t.Fatal("expected invalid model error")
		}
	}
}
