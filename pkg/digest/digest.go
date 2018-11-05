package digest

import (
	"context"
	"fmt"
)

type Validator interface {
	Validate(ctx context.Context, digest string, image string) error
}

func NewValidator(imageType string) (Validator, error) {
	if imageType == "docker" {
		return &dockerDigestValidator{}, nil
	}
	return nil, fmt.Errorf("unknown image type")
}
