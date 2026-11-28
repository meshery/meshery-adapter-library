package adapter

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	registrationAttemptTimeout = 30 * time.Second
	registrationErrorBodyLimit = 4096
)

type registrationRequestError struct {
	err error
}

func (e *registrationRequestError) Error() string {
	return fmt.Sprintf("create registration request: %v", e.err)
}

func (e *registrationRequestError) Unwrap() error {
	return e.err
}

// registrationHTTPError describes an unsuccessful registry response.
type registrationHTTPError struct {
	StatusCode int
	Status     string
	Body       string
}

func (e *registrationHTTPError) Error() string {
	if e.Body != "" {
		return fmt.Sprintf("registration request failed: %s: %s", e.Status, e.Body)
	}
	return fmt.Sprintf("registration request failed: %s", e.Status)
}

// registrationHTTPClient sends serialized MeshModel registration requests.
type registrationHTTPClient struct {
	client *http.Client
}

func newRegistrationHTTPClient(client *http.Client) *registrationHTTPClient {
	if client == nil {
		client = http.DefaultClient
	}
	return &registrationHTTPClient{client: client}
}

func (c *registrationHTTPClient) post(
	ctx context.Context,
	endpoint string,
	payload []byte,
) error {
	if ctx == nil {
		return &registrationRequestError{
			err: fmt.Errorf("registration context is required"),
		}
	}

	ctx, cancel := context.WithTimeout(ctx, registrationAttemptTimeout)
	defer cancel()

	parsed, err := url.Parse(endpoint)
	if err != nil {
		return &registrationRequestError{err: err}
	}

	if (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Hostname() == "" {
		return &registrationRequestError{
			err: fmt.Errorf("registration endpoint must be an absolute HTTP or HTTPS URL"),
		}
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		endpoint,
		bytes.NewReader(payload),
	)
	if err != nil {
		return &registrationRequestError{err: err}
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("send registration request: %w", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated, http.StatusAccepted:
		// A successful status acknowledges the request.
		// A body read failure must not trigger another registration.
		_, _ = io.Copy(io.Discard, resp.Body)
		return nil
	default:
		body, _ := io.ReadAll(io.LimitReader(resp.Body, registrationErrorBodyLimit))
		return &registrationHTTPError{
			StatusCode: resp.StatusCode,
			Status:     resp.Status,
			Body:       strings.TrimSpace(string(body)),
		}
	}
}
