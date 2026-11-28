package adapter

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/layer5io/meshkit/models/meshmodel/core/types"
)

func newLegacyReviewRegistrant(endpoint string) *MeshModelRegistrant {
	return NewMeshModelRegistrant(
		[]MeshModelRegistrantDefinitionPath{
			{
				EntityDefintionPath: filepath.Join(
					"testdata", "linkerd",
					testLinkerdCoreFilename,
				),
				Type: types.ComponentDefinition,
				Host: "meshery-linkerd",
				Port: 10005,
			},
		},
		endpoint,
	)
}

func TestLegacyRegistrationCanceledContext(t *testing.T) {
	var calls atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, _ *http.Request) {
			calls.Add(1)
			w.WriteHeader(http.StatusOK)
		},
	))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := newLegacyReviewRegistrant(server.URL).
		RegisterWithContext(ctx, "test-context")

	if err == nil {
		t.Fatal("expected cancellation error")
	}

	if calls.Load() != 0 {
		t.Fatal("canceled registration sent an HTTP request")
	}
}

func TestLegacyRegistrationPermanentFailure(t *testing.T) {
	var calls atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, _ *http.Request) {
			calls.Add(1)
			w.WriteHeader(http.StatusBadRequest)
		},
	))
	defer server.Close()

	ctx, cancel := context.WithTimeout(
		context.Background(), 3*time.Second,
	)
	defer cancel()

	err := newLegacyReviewRegistrant(server.URL).
		RegisterWithContext(ctx, "test-context")

	if err == nil {
		t.Fatal("expected registration failure")
	}

	if calls.Load() != 1 {
		t.Fatalf("expected one request, got %d", calls.Load())
	}
}

func TestLegacyRegistrationBackwardCompatibility(t *testing.T) {
	var calls atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, _ *http.Request) {
			calls.Add(1)
			w.WriteHeader(http.StatusOK)
		},
	))
	defer server.Close()

	err := newLegacyReviewRegistrant(server.URL).
		Register("test-context")

	if err != nil {
		t.Fatal(err)
	}

	if calls.Load() != 1 {
		t.Fatalf("expected one request, got %d", calls.Load())
	}
}

func TestLegacyRegistrationCancellationDuringRetry(t *testing.T) {
	var calls atomic.Int32

	ctx, cancel := context.WithTimeout(
		context.Background(), 3*time.Second,
	)
	defer cancel()

	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, _ *http.Request) {
			calls.Add(1)
			cancel()
			w.WriteHeader(http.StatusServiceUnavailable)
		},
	))
	defer server.Close()

	err := newLegacyReviewRegistrant(server.URL).
		RegisterWithContext(ctx, "test-context")

	if err == nil {
		t.Fatal("expected cancellation error")
	}

	if calls.Load() != 1 {
		t.Fatalf("expected one request, got %d", calls.Load())
	}
}
