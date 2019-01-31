package driver

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	// Side-effect import
	//_ "k8s.io/client-go/plugin/pkg/client/auth"
)

const secretTypeDuffle = "duffle.sh/cnab"

// Kubernetes is the Kubernetes driver for Duffle
//
// It works by injecting configuration into a config map, credentials into a secret, and
// the container into a pod. By default, it uses the standard Kubernetes configuration
// to connect to the cluster.
type Kubernetes struct {
	// Simulate indicates that the driver should simulate the operation.
	Simulate    bool
	Client      *kubernetes.Clientset
	Namespace   string
	KubeConfig  string
	KubeContext string
	Verbose     bool
}

// Config returns the configuration options for Kubernetes
func (d *Kubernetes) Config() map[string]string {
	return map[string]string{
		"KUBE_NAMESPACE": "Kubernetes namespace in which to run the invocation image",
		"KUBE_CONFIG":    "The path to the Kubernetes configuration file",
		"KUBE_CONTEXT":   "The name of the Kubernetes context to use",
		"VERBOSE":        "If 1, verbose output will be sent.",
	}
}

// SetConfig sets the configuration parameters on Kubernetes
func (d *Kubernetes) SetConfig(config map[string]string) {
	for k, v := range config {
		switch k {
		case "KUBE_NAMESPACE":
			d.Namespace = v
		case "KUBE_CONFIG":
			d.KubeConfig = v
		case "KUBE_CONTEXT":
			d.KubeContext = v
		case "VERBOSE":
			d.Verbose = v == "1"
		case "SIMULATE":
			d.Simulate = v == "1"
		}
	}
}

// Run executes the operation inside of the invocation image
func (d *Kubernetes) Run(op *Operation) error {
	if !d.Handles(op.ImageType) {
		return fmt.Errorf("driver for Kubernetes does not handle type %q", op.ImageType)
	}

	if d.Client == nil {
		c, err := d.kubeClient()
		if err != nil {
			return err
		}
		d.Client = c
	}
	if d.Namespace == "" {
		d.Namespace = "default"
	}

	runName := genName(op.Installation, op.Revision)

	if err := d.createSecret(runName, op); err != nil {
		return err
	}
	defer d.destroySecret(runName)

	return d.runPodAndWait(runName, op)
}

func genName(installation, revision string) string {
	return strings.ToLower(fmt.Sprintf("%s-%s", installation, revision))
}

// Handles returns true if the image type is docker or oci.
func (d *Kubernetes) Handles(imgType string) bool {
	switch strings.ToLower(imgType) {
	case "docker", "oci":
		return true
	}
	return false
}

// createSecret creates a secret that stores all of the envs and files
func (d *Kubernetes) createSecret(name string, op *Operation) error {

	combinedSecrets := map[string]string{}
	for k, v := range op.Environment {
		combinedSecrets[k] = v
	}
	for k, v := range op.Files {
		combinedSecrets[k] = v
	}

	secret := v1.Secret{
		ObjectMeta: meta.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"release":  op.Installation,
				"action":   op.Action,
				"heritage": "duffle",
				"revision": op.Revision,
			},
		},
		Type: secretTypeDuffle,
	}
	secret.StringData = combinedSecrets

	if d.Simulate {
		data, err := json.MarshalIndent(secret, "", "  ")
		fmt.Fprintln(op.Out, "Secret:")
		fmt.Fprintln(op.Out, string(data))
		return err
	}

	_, err := d.Client.CoreV1().Secrets(d.Namespace).Create(&secret)
	return err
}

func (d *Kubernetes) runPodAndWait(name string, op *Operation) error {

	pod := &v1.Pod{
		ObjectMeta: meta.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"release":  op.Installation,
				"action":   op.Action,
				"heritage": "duffle",
				"revision": op.Revision,
			},
		},
		Spec: v1.PodSpec{
			RestartPolicy: v1.RestartPolicyNever,
			Containers: []v1.Container{
				v1.Container{
					Name:         "invocationimage",
					Image:        op.Image,
					VolumeMounts: []v1.VolumeMount{},
					Env:          []v1.EnvVar{},
				},
			},
			Volumes: []v1.Volume{},
		},
	}

	// Copy env var definitions into pod
	// Because credentials may be passed here, we don't put the values in. Instead, we
	// reference the secret.
	vars := []v1.EnvVar{
		{
			Name:  "CNAB_ACTION",
			Value: op.Action,
		},
	}
	trueVal := true
	for k := range op.Environment {
		vars = append(vars, v1.EnvVar{
			Name: k,
			ValueFrom: &v1.EnvVarSource{
				SecretKeyRef: &v1.SecretKeySelector{
					Key: k,
					LocalObjectReference: v1.LocalObjectReference{
						Name: name,
					},
					Optional: &trueVal,
				},
			},
		})
	}
	pod.Spec.Containers[0].Env = vars

	// Copy volumes into pod
	// Again, we use secrets because the data inside of these may be credential info
	for k := range op.Files {
		pod.Spec.Volumes = append(pod.Spec.Volumes, v1.Volume{
			Name: name,
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: name,
					Optional:   &trueVal,
				},
			},
		})

		for i := range pod.Spec.Containers {
			pod.Spec.Containers[i].VolumeMounts = append(
				pod.Spec.Containers[i].VolumeMounts,
				v1.VolumeMount{
					Name:      name,
					MountPath: k,
				},
			)
		}
	}

	if d.Simulate {
		fmt.Fprintln(op.Out, "Pod:")
		data, err := json.MarshalIndent(pod, "", "  ")
		fmt.Fprintln(op.Out, string(data))
		return err
	}

	// Create the pod
	if _, err := d.Client.CoreV1().Pods(d.Namespace).Create(pod); err != nil {
		return err
	}

	// Waid for the pod to run to completion.
	return d.waitForPod(name, op)
}

func (d *Kubernetes) destroySecret(name string) error {
	if d.Simulate {
		return nil
	}
	return d.Client.CoreV1().Secrets(d.Namespace).Delete(name, &meta.DeleteOptions{})
}

// kubeClient returns a Kubernetes clientset.
func (d *Kubernetes) kubeClient() (*kubernetes.Clientset, error) {
	cfg, err := d.getKubeConfig()
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(cfg)
}

func (d *Kubernetes) getKubeConfig() (*rest.Config, error) {
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	rules.DefaultClientConfig = &clientcmd.DefaultClientConfig
	rules.ExplicitPath = d.KubeConfig

	overrides := &clientcmd.ConfigOverrides{
		ClusterDefaults: clientcmd.ClusterDefaults,
		CurrentContext:  d.KubeContext,
	}

	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, overrides).ClientConfig()
}

func (d *Kubernetes) waitForPod(name string, op *Operation) error {
	opts := meta.ListOptions{
		LabelSelector: fmt.Sprintf("heritage=duffle,release=%s,revision=%s", op.Installation, op.Revision),
	}
	req, err := d.Client.CoreV1().Pods(d.Namespace).Watch(opts)
	if err != nil {
		return err
	}
	res := req.ResultChan()

	// Now we block until the Pod is ready
	timeout := time.After(30 * time.Minute)
	for {
		select {
		case e := <-res:
			if d.Verbose {
				d, _ := json.MarshalIndent(e.Object, "", "  ")
				fmt.Fprintf(op.Out, "Event: %s\n %s\n", e.Type, d)
			}
			// If the pod is added or modified, check the phase and see if it is
			// running or complete.
			switch e.Type {
			case "DELETED":
				// This happens if a user directly kills the pod with kubectl.
				return fmt.Errorf("pod %s was just deleted unexpectedly", name)
			case "ADDED", "MODIFIED":
				pod := e.Object.(*v1.Pod)
				switch pod.Status.Phase {
				// Unhandled cases are Unknown and Pending, both of which should
				// cause the loop to spin.
				case "Running", "Succeeded":
					req.Stop()
					return nil
				case "Failed":
					req.Stop()
					return fmt.Errorf("pod failed to schedule: %s", pod.Status.Reason)
				}
			}
		case <-timeout:
			req.Stop()
			return fmt.Errorf("timeout waiting for pod %s to start", name)
		}
	}
}
