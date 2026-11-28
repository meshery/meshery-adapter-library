package adapter

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	backoff "github.com/cenkalti/backoff/v4"
)

type reviewRoundTripper func(*http.Request) (*http.Response, error)

func (f reviewRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func TestRegistrationAttemptDeadline(t *testing.T) {
	client := &http.Client{
		Transport: reviewRoundTripper(func(req *http.Request) (*http.Response, error) {
			deadline, ok := req.Context().Deadline()
			if !ok {
				t.Fatal("registration request has no deadline")
			}

			remaining := time.Until(deadline)
			if remaining <= 0 || remaining > registrationAttemptTimeout {
				t.Fatalf("unexpected deadline: %v", remaining)
			}

			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader("")),
			}, nil
		}),
	}

	err := newRegistrationHTTPClient(client).post(
		context.Background(),
		"http://example.test/register",
		[]byte("{}"),
	)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRegistrationBoundedErrorBody(t *testing.T) {
	body := "registry validation failed: " +
		strings.Repeat("x", registrationErrorBodyLimit*2)

	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = io.WriteString(w, body)
		},
	))
	defer server.Close()

	err := newRegistrationHTTPClient(server.Client()).post(
		context.Background(), server.URL, []byte("{}"),
	)

	var httpErr *registrationHTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected HTTP registration error, got %v", err)
	}

	if httpErr.StatusCode != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d", httpErr.StatusCode)
	}

	if len(httpErr.Body) != registrationErrorBodyLimit {
		t.Fatalf("unexpected error body length: %d", len(httpErr.Body))
	}

	if !strings.Contains(err.Error(), "registry validation failed") {
		t.Fatal("missing registry error details")
	}
}

func TestRegistrationMalformedEndpointPermanent(t *testing.T) {
	endpoints := []string{
		"://invalid",
		"relative/path",
		"ftp://example.test",
		"http://",
		"http://example.test:invalid",
	}

	for _, endpoint := range endpoints {
		t.Run(endpoint, func(t *testing.T) {
			err := newRegistrationHTTPClient(nil).post(
				context.Background(), endpoint, []byte("{}"),
			)

			var requestErr *registrationRequestError
			if !errors.As(err, &requestErr) {
				t.Fatalf("expected request error, got %v", err)
			}

			var permanentErr *backoff.PermanentError
			if !errors.As(classifyRegistrationError(err), &permanentErr) {
				t.Fatal("malformed endpoint must not be retried")
			}
		})
	}
}

func TestRegistrationNilContext(t *testing.T) {
	r := &ModernComponentRegistrant{
		Endpoint: "http://example.test/register",
	}

	err := r.RegisterFiles(nil, []string{"component.json"}) //nolint:staticcheck // Intentionally verify rejection of a nil context.
	if err == nil || !strings.Contains(err.Error(), "context") {
		t.Fatalf("expected context validation error, got %v", err)
	}
}

func TestRegistrationEarlierCallerDeadline(t *testing.T) {
	ctx, cancel := context.WithTimeout(
		context.Background(), 5*time.Second,
	)
	defer cancel()

	expected, _ := ctx.Deadline()

	client := &http.Client{
		Transport: reviewRoundTripper(func(req *http.Request) (*http.Response, error) {
			actual, ok := req.Context().Deadline()
			if !ok || !actual.Equal(expected) {
				t.Fatalf(
					"expected caller deadline %v, got %v",
					expected, actual,
				)
			}

			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader("")),
			}, nil
		}),
	}

	err := newRegistrationHTTPClient(client).post(
		ctx,
		"http://example.test/register",
		[]byte("{}"),
	)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRegistrationBlockedRequestDeadline(t *testing.T) {
	ctx, cancel := context.WithTimeout(
		context.Background(), 100*time.Millisecond,
	)
	defer cancel()

	client := &http.Client{
		Transport: reviewRoundTripper(func(req *http.Request) (*http.Response, error) {
			select {
			case <-req.Context().Done():
				return nil, req.Context().Err()
			case <-time.After(3 * time.Second):
				return nil, errors.New("request deadline was not propagated")
			}
		}),
	}

	err := newRegistrationHTTPClient(client).post(
		ctx,
		"http://example.test/register",
		[]byte("{}"),
	)

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected deadline exceeded, got %v", err)
	}
}

func TestRegistrationInFlightCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client := &http.Client{
		Transport: reviewRoundTripper(func(req *http.Request) (*http.Response, error) {
			cancel()

			select {
			case <-req.Context().Done():
				return nil, req.Context().Err()
			case <-time.After(3 * time.Second):
				return nil, errors.New("request cancellation was not propagated")
			}
		}),
	}

	err := newRegistrationHTTPClient(client).post(
		ctx,
		"http://example.test/register",
		[]byte("{}"),
	)

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation, got %v", err)
	}
}

type registrationFailingResponseBody struct {
	closed bool
}

func (b *registrationFailingResponseBody) Read([]byte) (int, error) {
	return 0, io.ErrUnexpectedEOF
}

func (b *registrationFailingResponseBody) Close() error {
	b.closed = true
	return nil
}

func TestRegistrationAcknowledgementWithBodyFailure(t *testing.T) {
	for _, status := range []int{
		http.StatusOK,
		http.StatusCreated,
		http.StatusAccepted,
	} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			body := &registrationFailingResponseBody{}
			requests := 0

			client := &http.Client{
				Transport: reviewRoundTripper(func(_ *http.Request) (*http.Response, error) {
					requests++

					return &http.Response{
						StatusCode: status,
						Body:       body,
					}, nil
				}),
			}

			err := newRegistrationHTTPClient(client).post(
				context.Background(),
				"http://example.test/register",
				[]byte("{}"),
			)

			if err != nil {
				t.Fatalf("expected HTTP acknowledgement, got %v", err)
			}

			if requests != 1 {
				t.Fatalf("expected one request, got %d", requests)
			}

			if !body.closed {
				t.Fatal("response body was not closed")
			}
		})
	}
}
