package action

import (
	"context"
	"fmt"
	"io"

	"github.com/deis/duffle/pkg/claim"
	"github.com/deis/duffle/pkg/credentials"
	"github.com/deis/duffle/pkg/digest"
	"github.com/deis/duffle/pkg/driver"
)

// Install describes an installation action
type Install struct {
	Driver driver.Driver // Needs to be more than a string
}

// Run performs an installation and updates the Claim accordingly
func (i *Install) Run(c *claim.Claim, creds credentials.Set, w io.Writer) error {
	invocImage, err := selectInvocationImage(i.Driver, c)
	if err != nil {
		return err
	}
	// Validate the Digests of Invocation Image and Bundle Images
	validator, err := digest.NewValidator(invocImage.ImageType)
	if err != nil {
		return fmt.Errorf("unable to get image validator: %s", err)
	}
	ctx := context.Background()
	if invocImage.BaseImage.Digest != "" {
		err = validator.Validate(ctx, invocImage.BaseImage.Digest, invocImage.BaseImage.Image)
		if err != nil {
			return fmt.Errorf("unable to validate invocation image: %s", err)
		}
	}
	// TODO, how do we know what kind this is? default to docker right now
	for _, img := range c.Bundle.Images {
		if img.Digest != "" {
			err = validator.Validate(ctx, img.BaseImage.Digest, img.BaseImage.Image)
			if err != nil {
				return fmt.Errorf("unable to validate image %s: %s", img.BaseImage.Image, err)
			}
		}
	}
	op, err := opFromClaim(claim.ActionInstall, c, invocImage, creds, w)
	if err != nil {
		return err
	}
	if err := i.Driver.Run(op); err != nil {
		c.Update(claim.ActionInstall, claim.StatusFailure)
		c.Result.Message = err.Error()
		return err
	}

	// Update claim:
	c.Update(claim.ActionInstall, claim.StatusSuccess)
	return nil
}
