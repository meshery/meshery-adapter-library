package adapter

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	meshkiterrors "github.com/layer5io/meshkit/errors"
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

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
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

	if code := meshkiterrors.GetCode(err); code != ErrOAMRetryCode {
		t.Fatalf("expected legacy error code %s, got %s",
			ErrOAMRetryCode, code)
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

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}

	if calls.Load() != 1 {
		t.Fatalf("expected one request, got %d", calls.Load())
	}
}

func TestLegacyRegistrationCallerDeadline(t *testing.T) {
	var calls atomic.Int32
	release := make(chan struct{})

	server := httptest.NewServer(http.HandlerFunc(
		func(_ http.ResponseWriter, r *http.Request) {
			calls.Add(1)
			select {
			case <-r.Context().Done():
			case <-release:
			case <-time.After(3 * time.Second):
			}
		},
	))
	defer server.Close()
	defer close(release)

	ctx, cancel := context.WithTimeout(
		context.Background(), 500*time.Millisecond,
	)
	defer cancel()

	err := newLegacyReviewRegistrant(server.URL).
		RegisterWithContext(ctx, "test-context")

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected context.DeadlineExceeded, got %v", err)
	}

	if calls.Load() != 1 {
		t.Fatalf("expected one request, got %d", calls.Load())
	}
}
