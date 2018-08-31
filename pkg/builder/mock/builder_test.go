package mock

import (
	"testing"

	"github.com/deis/duffle/pkg/builder"
)

func TestComponent_implComponent(t *testing.T) {
	var _ builder.Component = (*Component)(nil)
}
