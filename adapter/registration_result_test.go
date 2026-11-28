package adapter

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"sync/atomic"
	"testing"
	"time"
)

func TestPartialRegistrationHTTPFailure(t *testing.T) {
	var requests atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, _ *http.Request) {
			if requests.Add(1) == 2 {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusCreated)
		},
	))
	defer server.Close()

	paths := partialRegistrationTestPaths(t)

	ctx, cancel := context.WithTimeout(
		context.Background(), 5*time.Second,
	)
	defer cancel()

	r := &ModernComponentRegistrant{
		Endpoint:   server.URL,
		Connection: []byte(`{"kind":"linkerd"}`),
		Client:     server.Client(),
	}

	err := r.RegisterFiles(ctx, paths)

	assertPartialRegistration(t, err, paths)

	var httpErr *registrationHTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected underlying HTTP error: %v", err)
	}

	if httpErr.StatusCode != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d", httpErr.StatusCode)
	}

	if requests.Load() != 2 {
		t.Fatalf("expected two requests, got %d", requests.Load())
	}
}

func TestPartialRegistrationMissingFile(t *testing.T) {
	var requests atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, _ *http.Request) {
			requests.Add(1)
			w.WriteHeader(http.StatusCreated)
		},
	))
	defer server.Close()

	paths := partialRegistrationTestPaths(t)
	paths[1] = filepath.Join(t.TempDir(), "missing.json")

	r := &ModernComponentRegistrant{
		Endpoint:   server.URL,
		Connection: []byte(`{"kind":"linkerd"}`),
		Client:     server.Client(),
	}

	err := r.RegisterFiles(context.Background(), paths)

	assertPartialRegistration(t, err, paths)

	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected missing-file error: %v", err)
	}

	if requests.Load() != 1 {
		t.Fatalf("expected one request, got %d", requests.Load())
	}
}

func TestPartialRegistrationCanceledContext(t *testing.T) {
	paths := partialRegistrationTestPaths(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	r := &ModernComponentRegistrant{
		Endpoint:   "http://example.test/register",
		Connection: []byte(`{"kind":"linkerd"}`),
	}

	err := r.RegisterFiles(ctx, paths)

	var partial *PartialRegistrationError
	if !errors.As(err, &partial) {
		t.Fatalf("expected partial registration error: %v", err)
	}

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected cancellation error: %v", err)
	}

	if len(partial.Acknowledged) != 0 ||
		partial.Failed != paths[0] ||
		!slices.Equal(partial.Unattempted, paths[1:]) {
		t.Fatalf("unexpected cancellation outcome: %+v", partial)
	}
}

func partialRegistrationTestPaths(t *testing.T) []string {
	t.Helper()

	data, err := os.ReadFile(filepath.Join(
		"testdata", "linkerd", testLinkerdCoreFilename,
	))
	if err != nil {
		t.Fatal(err)
	}

	dir := t.TempDir()
	paths := make([]string, 0, 3)

	for _, name := range []string{
		"first.json", "second.json", "third.json",
	} {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, data, 0600); err != nil {
			t.Fatal(err)
		}
		paths = append(paths, path)
	}

	return paths
}

func assertPartialRegistration(
	t *testing.T,
	err error,
	paths []string,
) {
	t.Helper()

	var partial *PartialRegistrationError
	if !errors.As(err, &partial) {
		t.Fatalf("expected partial registration error: %v", err)
	}

	if !slices.Equal(partial.Acknowledged, paths[:1]) {
		t.Fatalf("unexpected acknowledged files: %v", partial.Acknowledged)
	}

	if partial.Failed != paths[1] {
		t.Fatalf("unexpected failed file: %s", partial.Failed)
	}

	if !slices.Equal(partial.Unattempted, paths[2:]) {
		t.Fatalf("unexpected unattempted files: %v", partial.Unattempted)
	}
}
