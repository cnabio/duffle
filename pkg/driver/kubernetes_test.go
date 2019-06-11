package driver

import (
	"os"
	"testing"

	"github.com/deislabs/cnab-go/driver"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestKubernetesDriver_Run(t *testing.T) {
	client := fake.NewSimpleClientset()
	namespace := "default"
	k := KubernetesDriver{
		Namespace:          namespace,
		jobs:               client.BatchV1().Jobs(namespace),
		secrets:            client.CoreV1().Secrets(namespace),
		pods:               client.CoreV1().Pods(namespace),
		SkipCleanup:        true,
		skipJobStatusCheck: true,
	}
	op := driver.Operation{
		Action: "install",
		Out:    os.Stdout,
		Environment: map[string]string{
			"foo": "bar",
		},
	}

	err := k.Run(&op)
	assert.NoError(t, err)

	jobList, _ := k.jobs.List(metav1.ListOptions{})
	assert.Equal(t, len(jobList.Items), 1, "expected one job to be created")

	secretList, _ := k.secrets.List(metav1.ListOptions{})
	assert.Equal(t, len(secretList.Items), 1, "expected one secret to be created")
}
