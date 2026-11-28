package adapter

import (
	"fmt"
	"slices"
)

// PartialRegistrationError describes an interrupted registration batch.
// Acknowledged files received successful HTTP responses, but their
// persistence in the registry is not independently verified.
// The failed file may have been persisted despite a request failure.
// Callers must not assume that replaying the batch is safe.
type PartialRegistrationError struct {
	Acknowledged []string
	Failed       string
	Unattempted  []string
	Err          error
}

func (e *PartialRegistrationError) Error() string {
	return fmt.Sprintf(
		"registration stopped at %q (%d acknowledged, %d unattempted): %v",
		e.Failed,
		len(e.Acknowledged),
		len(e.Unattempted),
		e.Err,
	)
}

func (e *PartialRegistrationError) Unwrap() error {
	return e.Err
}

func newPartialRegistrationError(
	paths []string,
	index int,
	err error,
) *PartialRegistrationError {
	return &PartialRegistrationError{
		Acknowledged: slices.Clone(paths[:index]),
		Failed:       paths[index],
		Unattempted:  slices.Clone(paths[index+1:]),
		Err:          err,
	}
}
