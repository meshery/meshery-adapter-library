package adapter

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
)

// registrationHTTPClient sends serialized MeshModel registration requests.
type registrationHTTPError struct {
	StatusCode int
	Status     string
}

func (e *registrationHTTPError) Error() string {
	return fmt.Sprintf("registration request failed: %s", e.Status)
}

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
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		endpoint,
		bytes.NewReader(payload),
	)
	if err != nil {
		return fmt.Errorf("create registration request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("send registration request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	_, _ = io.Copy(io.Discard, resp.Body)

	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated, http.StatusAccepted:
		return nil
	default:
		return &registrationHTTPError{
			StatusCode: resp.StatusCode,
			Status:     resp.Status,
		}
	}
}
