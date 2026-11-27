package adapter

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRegistrationHTTPClientSuccess(t *testing.T) {
	for _, status := range []int{
		http.StatusOK,
		http.StatusCreated,
		http.StatusAccepted,
	} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					if r.Method != http.MethodPost {
						t.Errorf("unexpected method: %s", r.Method)
					}

					if got := r.Header.Get("Content-Type"); got != "application/json" {
						t.Errorf("unexpected content type: %s", got)
					}

					body, err := io.ReadAll(r.Body)
					if err != nil {
						t.Fatal(err)
					}

					if string(body) != `{"test":true}` {
						t.Errorf("unexpected body: %s", body)
					}

					w.WriteHeader(status)
				},
			))
			defer server.Close()

			client := newRegistrationHTTPClient(server.Client())
			err := client.post(
				context.Background(),
				server.URL,
				[]byte(`{"test":true}`),
			)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestRegistrationHTTPClientFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
		},
	))
	defer server.Close()

	client := newRegistrationHTTPClient(server.Client())
	err := client.post(context.Background(), server.URL, []byte(`{}`))

	if err == nil || !strings.Contains(err.Error(), "400") {
		t.Fatalf("expected HTTP 400 error, got %v", err)
	}
}

func TestRegistrationHTTPClientCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	client := newRegistrationHTTPClient(nil)
	err := client.post(ctx, "http://127.0.0.1:1", []byte(`{}`))

	if err == nil {
		t.Fatal("expected canceled-context error")
	}
}
