package adapter

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

func TestModernRegistrationRetryPolicy(t *testing.T) {
	cases := []struct {
		name         string
		status       int
		wantRequests int32
		wantError    bool
	}{
		{"BadRequest", 400, 1, true},
		{"Unauthorized", 401, 1, true},
		{"Forbidden", 403, 1, true},
		{"NotFound", 404, 1, true},
		{"RequestTimeout", 408, 2, false},
		{"TooManyRequests", 429, 2, false},
		{"InternalServerError", 500, 2, false},
		{"ServiceUnavailable", 503, 2, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var requests atomic.Int32

			server := httptest.NewServer(http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					if requests.Add(1) == 1 {
						w.WriteHeader(tc.status)
						return
					}
					w.WriteHeader(http.StatusCreated)
				},
			))
			defer server.Close()

			ctx, cancel := context.WithTimeout(
				context.Background(), 5*time.Second,
			)
			defer cancel()

			registrant := &ModernComponentRegistrant{
				Endpoint:   server.URL,
				Connection: []byte(`{"kind":"linkerd"}`),
				Client:     server.Client(),
			}

			err := registrant.RegisterFiles(ctx, []string{
				filepath.Join("testdata", testLinkerdName,
					testLinkerdCoreFilename),
			})

			if (err != nil) != tc.wantError {
				t.Fatalf("unexpected registration result: %v", err)
			}

			if got := requests.Load(); got != tc.wantRequests {
				t.Fatalf("expected %d requests, got %d",
					tc.wantRequests, got)
			}
		})
	}
}
