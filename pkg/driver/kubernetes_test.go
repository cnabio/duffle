package driver

// TODO(jlegrone): uncomment tests when k8s.io/client-go/kubernetes/fake is available

// import (
// 	"os"
// 	"testing"

// 	"github.com/deislabs/cnab-go/driver"
// 	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
// 	"k8s.io/client-go/kubernetes/fake"
// )

// func TestKubernetesRun(t *testing.T) {
// 	client := fake.NewSimpleClientset()
// 	namespace := "default"
// 	k := KubernetesDriver{
// 		Namespace:          namespace,
// 		jobs:               client.BatchV1().Jobs(namespace),
// 		secrets:            client.CoreV1().Secrets(namespace),
// 		pods:               client.CoreV1().Pods(namespace),
// 		SkipCleanup:        true,
// 		skipJobStatusCheck: true,
// 	}
// 	op := driver.Operation{
// 		Action: "install",
// 		Out:    os.Stdout,
// 		Environment: map[string]string{
// 			"foo": "bar",
// 		},
// 	}

// 	err := k.Run(&op)
// 	if err != nil {
// 		t.Error(err)
// 	}

// 	jobList, _ := k.jobs.List(metav1.ListOptions{})
// 	if len(jobList.Items) != 1 {
// 		t.Errorf("Expected one item in jobList, got %d", len(jobList.Items))
// 	}

// 	secretList, _ := k.secrets.List(metav1.ListOptions{})
// 	if len(secretList.Items) != 1 {
// 		t.Errorf("Expected one item in secretList, got %d", len(secretList.Items))
// 	}
// }
