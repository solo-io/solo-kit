package vault

import (
	"context"
	"fmt"

	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
)

var _ error = new(UnrecoverableConversionError)

type SecretConverter interface {
	// FromSecret accepts the raw value of a Vault secret and returns the Gloo representation
	// of that secret. An error is returned if the conversion failed
	FromSecret(ctx context.Context, secret Secret) (resources.Resource, error)
}

type UnrecoverableConversionError struct {
	Err error
}

func UnrecoverableConversionErr(err error) UnrecoverableConversionError {
	return UnrecoverableConversionError{
		Err: err,
	}
}

func (u UnrecoverableConversionError) Error() string {
	return fmt.Sprintf("UnrecoverableConversionError: %v", u.Err)
}
