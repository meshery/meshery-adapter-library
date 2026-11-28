package adapter

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	backoff "github.com/cenkalti/backoff/v4"
)

// ModernComponentRegistrant registers modern MeshModel component definitions.
type ModernComponentRegistrant struct {
	Endpoint   string
	Connection []byte
	Client     *http.Client
}

// RegisterFiles registers the supplied component definition files.
func (r *ModernComponentRegistrant) RegisterFiles(
	ctx context.Context,
	paths []string,
) error {
	if r == nil {
		return fmt.Errorf("nil component registrant")
	}

	if ctx == nil {
		return fmt.Errorf("registration context is required")
	}

	if r.Endpoint == "" {
		return fmt.Errorf("registration endpoint is required")
	}

	if len(paths) == 0 {
		return fmt.Errorf("no component definition files provided")
	}

	client := newRegistrationHTTPClient(r.Client)

	for index, path := range paths {
		if err := ctx.Err(); err != nil {
			return newPartialRegistrationError(paths, index, err)
		}

		if err := r.registerFile(ctx, client, path); err != nil {
			return newPartialRegistrationError(paths, index, err)
		}
	}

	return nil
}

func (r *ModernComponentRegistrant) registerFile(
	ctx context.Context,
	client *registrationHTTPClient,
	path string,
) error {
	componentJSON, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read component definition %q: %w", path, err)
	}

	componentJSON, err = prepareComponentForRegistration(componentJSON)
	if err != nil {
		return fmt.Errorf("prepare component definition %q: %w", path, err)
	}

	payload, err := newComponentRegistrationPayload(
		r.Connection,
		componentJSON,
	)
	if err != nil {
		return fmt.Errorf("prepare component %q: %w", path, err)
	}

	if err := r.registerPayload(ctx, client, payload); err != nil {
		return fmt.Errorf("register component %q: %w", path, err)
	}

	return nil
}

func (r *ModernComponentRegistrant) registerPayload(
	ctx context.Context,
	client *registrationHTTPClient,
	payload []byte,
) error {
	retry := backoff.NewExponentialBackOff()
	retry.MaxElapsedTime = 10 * time.Minute

	return backoff.Retry(func() error {
		err := client.post(ctx, r.Endpoint, payload)
		return classifyRegistrationError(err)
	}, backoff.WithContext(retry, ctx))
}

func classifyRegistrationError(err error) error {
	var requestErr *registrationRequestError
	if errors.As(err, &requestErr) {
		return backoff.Permanent(err)
	}

	var httpErr *registrationHTTPError
	if errors.As(err, &httpErr) {
		code := httpErr.StatusCode

		if code >= 400 && code < 500 &&
			code != http.StatusRequestTimeout &&
			code != http.StatusTooManyRequests {
			return backoff.Permanent(err)
		}
	}

	return err
}
