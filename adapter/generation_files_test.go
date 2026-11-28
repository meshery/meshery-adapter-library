package adapter

import "testing"

func TestModernComponentFilename(t *testing.T) {
	cases := []struct {
		name     string
		data     string
		meshType string
		want     string
	}{
		{
			name:     "Linkerd ServiceProfile",
			data:     `{"component":{"kind":"ServiceProfile","version":"linkerd.io/v1alpha1"}}`,
			meshType: testLinkerdType,
			want:     "serviceprofile.Linkerd.meshery.layer5.io_meshmodel.json",
		},
		{
			name:     "Linkerd Server",
			data:     `{"component":{"kind":"Server","version":"policy.linkerd.io/v1alpha1"}}`,
			meshType: testLinkerdType,
			want:     "server.Linkerd.meshery.layer5.io_meshmodel.json",
		},
		{
			name:     "Linkerd AuthorizationPolicy",
			data:     `{"component":{"kind":"AuthorizationPolicy","version":"policy.linkerd.io/v1alpha1"}}`,
			meshType: testLinkerdType,
			want:     "authorizationpolicy.Linkerd.meshery.layer5.io_meshmodel.json",
		},
		{
			name:     "Linkerd core component",
			data:     `{"component":{"kind":"LinkerdMesh","version":"core.meshmodel.dev/v1alpha1"}}`,
			meshType: testLinkerdType,
			want:     testLinkerdCoreFilename,
		},
		{
			name: "Empty mesh type",
			data: `{"component":{"kind":"ServiceProfile","version":"linkerd.io/v1alpha1"}}`,
			want: "serviceprofile.meshery.layer5.io_meshmodel.json",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := modernComponentFilename(
				[]byte(tc.data), tc.meshType,
			)
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.want {
				t.Fatalf("expected %q, got %q", tc.want, got)
			}
		})
	}
}

func TestModernComponentFilenameInvalidInput(t *testing.T) {
	cases := []string{
		testInvalidJSON,
		`{}`,
		`{"component":{"kind":"","version":"v1"}}`,
		`{"component":{"kind":"Server","version":""}}`,
		`{"component":{"kind":"../Server","version":"v1"}}`,
		`{"component":{"kind":"Server","version":"../v1"}}`,
		`{"component":{"kind":"Server","version":"/v1"}}`,
	}

	for _, data := range cases {
		if _, err := modernComponentFilename(
			[]byte(data), testLinkerdType,
		); err == nil {
			t.Errorf("expected rejection for %s", data)
		}
	}

	data := []byte(`{"component":{"kind":"Server","version":"policy.linkerd.io/v1alpha1"}}`)

	if _, err := modernComponentFilename(data, "../Linkerd"); err == nil {
		t.Error("expected rejection of invalid mesh type")
	}
}
