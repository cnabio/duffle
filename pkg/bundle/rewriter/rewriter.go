package rewriter

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"text/template"

	"github.com/deis/duffle/pkg/bundle"
	"github.com/deis/duffle/pkg/bundle/replacement"
)

// Rewriter is an interface that is used to rewrite elements of a CNAB bundle
type Rewriter interface {
	Rewrite(contents string, image string, ref bundle.LocationRef) (string, error)
	ReplaceRepository(qualifiedImage string, repository string) (string, error)
	TagImage(ctx context.Context, oldImage string, newImage string) error
}

type rewriter struct {
	dockerClient dockerImageClient
}

// NewRewriter returns a bundle rewriter
func NewRewriter() (Rewriter, error) {
	cli, err := getDockerClient()
	if err != nil {
		return nil, err
	}
	r := &rewriter{
		dockerClient: cli,
	}
	return r, nil
}

func getReplacementValue(image string, value string, templateVal string) (string, error) {
	imageRef, err := getDockerReference(image)
	if err != nil {
		return "", err
	}
	if templateVal != "" {
		var replacement bytes.Buffer
		goTemplate, err := template.New("").Option("missingkey=error").Parse(templateVal)
		if err != nil {
			return "", fmt.Errorf("error building template: %s", err)
		}
		if err = goTemplate.Execute(&replacement, imageRef); err != nil {
			return "", fmt.Errorf("error executing template: %s", err)
		}
		return replacement.String(), nil
	}
	// This is the fallback behavior of no template is specified
	// If a tag is not present, just include the repo / image portion of the reference
	r, err := getDockerReference(value)
	if err != nil {
		return "", fmt.Errorf("invalid image reference :%s", err)
	}
	if r.Tag == "" {
		return strings.Join([]string{imageRef.Repo, imageRef.Image}, "/"), nil
	}
	return image, nil
}

// Rewrite updates the file contents at the specificed reference
func (r *rewriter) Rewrite(contents string, image string, ref bundle.LocationRef) (string, error) {
	replacer := replacement.GetReplacer(ref.Path)
	if replacer == nil {
		return "", fmt.Errorf("unknown file type, unable to replace references")
	}
	val, err := replacer.Retrieve(contents, ref.Field)
	if err != nil {
		return "", fmt.Errorf("cannot read reference file: %s, %v", ref.Path, err)
	}
	replacement, err := getReplacementValue(image, val, ref.Template)
	if err != nil {
		return "", fmt.Errorf("error replacing image")
	}
	return replacer.Replace(contents, ref.Field, replacement)
}
