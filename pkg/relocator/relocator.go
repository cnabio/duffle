package relocator

import (
	_ "crypto/sha256" // ensure SHA-256 is loaded
	"fmt"

	"github.com/deislabs/cnab-go/bundle"
	"github.com/pivotal/image-relocation/pkg/image"

	"github.com/deislabs/duffle/pkg/imagestore"
)

type Relocator struct {
	bun        *bundle.Bundle
	mapping    Mapping
	imageStore imagestore.Store
}

type Mapping func(image.Name) image.Name

func NewRelocator(bun *bundle.Bundle, mapping Mapping, is imagestore.Store) (*Relocator, error) {
	return &Relocator{
		bun:        bun,
		mapping:    mapping,
		imageStore: is,
	}, nil
}

func (r *Relocator) Relocate() error {
	for i := range r.bun.InvocationImages {
		ii := r.bun.InvocationImages[i]
		err := r.relocateImage(&ii.BaseImage)
		if err != nil {
			return err
		}
		r.bun.InvocationImages[i] = ii
	}

	for k := range r.bun.Images {
		im := r.bun.Images[k]
		err := r.relocateImage(&im.BaseImage)
		if err != nil {
			return err
		}
		r.bun.Images[k] = im
	}

	return nil

}

func (r *Relocator) relocateImage(i *bundle.BaseImage) error {
	if !isOCI(i.ImageType) && !isDocker(i.ImageType) {
		return fmt.Errorf("cannot relocate image %s with imageType %s: only oci and docker image types are currently supported", i.Image, i.ImageType)
	}
	// map the image name
	n, err := image.NewName(i.Image)
	if err != nil {
		return err
	}
	rn := r.mapping(n)

	dig := n.Digest()
	if dig == image.EmptyDigest && i.Digest != "" {
		var err error
		dig, err = image.NewDigest(i.Digest)
		if err != nil {
			return err
		}
	}

	fmt.Printf("writing %s\n", i.Image)
	err = r.imageStore.Push(dig, n, rn)
	if err != nil {
		return err
	}

	// update the imagemap
	i.OriginalImage = i.Image
	i.Image = rn.String()
	return nil
}

func isOCI(imageType string) bool {
	return imageType == "" || imageType == "oci"
}

func isDocker(imageType string) bool {
	return imageType == "docker"
}
