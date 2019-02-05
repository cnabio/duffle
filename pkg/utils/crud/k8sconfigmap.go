package crud

import (
	"fmt"
	"strings"

	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/clientcmd"
)

const configMapKey = "claim"

// NewFileSystemStore creates a Store backed by a kubernetes cluster.
// Each item is represented by a ConfigMap.
func NewK8sConfigMapStore(kubeconfig, masterURL, namespace, namePrefix string) (Store, error) {
	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("error building kubeconfig: %v", err)
	}
	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("error building kubernetes clientset: %v", err)
	}
	client := kubeClient.CoreV1().ConfigMaps(namespace)
	return k8sConfigMapStore{
		client:     client,
		namePrefix: namePrefix,
	}, nil
}

type k8sConfigMapStore struct {
	client     corev1client.ConfigMapInterface
	namePrefix string
}

func (s k8sConfigMapStore) List() ([]string, error) {
	listing, err := s.client.List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	items := []string{}
	prefix := s.fullyQualifiedName("")
	for _, cm := range listing.Items {
		name := cm.Name
		if strings.HasPrefix(name, prefix) {
			items = append(items, name[len(prefix):])
		}
	}
	return items, nil
}

func (s k8sConfigMapStore) Store(name string, data []byte) error {
	cm, err := s.client.Get(s.fullyQualifiedName(name), metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return s.createConfigMap(name, data)
		}
		return err
	}
	return s.updateConfigMap(cm, data)
}

func (s k8sConfigMapStore) createConfigMap(name string, data []byte) error {
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: s.fullyQualifiedName(name),
		},
		BinaryData: map[string][]byte{
			configMapKey: data,
		},
	}
	_, err := s.client.Create(cm)
	return err
}

func (s k8sConfigMapStore) updateConfigMap(cm *corev1.ConfigMap, data []byte) error {
	cm.BinaryData[configMapKey] = data
	_, err := s.client.Update(cm)
	return err
}

func (s k8sConfigMapStore) Read(name string) ([]byte, error) {
	cm, err := s.client.Get(s.fullyQualifiedName(name), metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, ErrDoesNotExist
		}
		return nil, err
	}
	return cm.BinaryData[configMapKey], nil
}

func (s k8sConfigMapStore) Delete(name string) error {
	return s.client.Delete(s.fullyQualifiedName(name), &metav1.DeleteOptions{})
}

func (s k8sConfigMapStore) fullyQualifiedName(name string) string {
	return fmt.Sprintf("%s-%s", s.namePrefix, name)
}
