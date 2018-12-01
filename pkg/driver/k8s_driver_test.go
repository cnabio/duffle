package driver

import "testing"

func TestKubernetesInterfaces(t *testing.T) {
	var _ Driver = &Kubernetes{}
	var _ Configurable = &Kubernetes{}
}
