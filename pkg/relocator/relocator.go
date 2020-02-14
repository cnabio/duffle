package relocator

import (
	_ "crypto/sha256" // ensure SHA-256 is loaded
	"fmt"
	"io"

	"github.com/deislabs/cnab-go/bundle"
	"github.com/pivotal/image-relocation/pkg/image"

	"github.com/cnabio/duffle/pkg/imagestore"
)

type Relocator struct {
	bun        *bundle.Bundle
	mapping    Mapping
	imageStore imagestore.Store
	out        io.Writer
}

type Mapping func(image.Name) image.Name

func NewRelocator(bun *bundle.Bundle, mapping Mapping, is imagestore.Store, logs io.Writer) (*Relocator, error) {
	return &Relocator{
		bun:        bun,
		mapping:    mapping,
		imageStore: is,
		out:        logs,
	}, nil
}

func (r *Relocator) Relocate(relMap map[string]string) error {
	for i := range r.bun.InvocationImages {
		ii := r.bun.InvocationImages[i]
		err := r.relocateImage(&ii.BaseImage, relMap)
		if err != nil {
			return err
		}
	}

	for k := range r.bun.Images {
		im := r.bun.Images[k]
		err := r.relocateImage(&im.BaseImage, relMap)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *Relocator) relocateImage(i *bundle.BaseImage, relMap map[string]string) error {
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

	fmt.Fprintf(r.out, "writing %s to %s\n", i.Image, rn.String())
	err = r.imageStore.Push(dig, n, rn)
	if err != nil {
		return err
	}

	// update the relocation mapping
	relMap[i.Image] = rn.String()
	return nil
}

func isOCI(imageType string) bool {
	return imageType == "" || imageType == "oci"
}

func isDocker(imageType string) bool {
	return imageType == "docker"
}
