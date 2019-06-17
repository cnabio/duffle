// +build integration

package driver

import (
	"bytes"
	"testing"

	"github.com/deislabs/cnab-go/driver"
	"github.com/stretchr/testify/assert"
)

func TestKubernetesDriver_Run_Integration(t *testing.T) {
	namespace := "default"
	k := &KubernetesDriver{}
	k.SetConfig(map[string]string{
		"KUBE_NAMESPACE": namespace,
	})
	k.ActiveDeadlineSeconds = 60

	cases := []struct {
		name   string
		op     *driver.Operation
		output string
		err    error
	}{
		{
			name: "install",
			op: &driver.Operation{
				Installation: "example",
				Action:       "install",
				Image:        "cnab/helloworld@sha256:55f83710272990efab4e076f9281453e136980becfd879640b06552ead751284",
				Environment: map[string]string{
					"PORT": "3000",
				},
			},
			output: "Port parameter was set to 3000\nInstall action\nAction install complete for example\n",
			err:    nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var output bytes.Buffer
			tc.op.Out = &output
			if tc.op.Environment == nil {
				tc.op.Environment = map[string]string{}
			}
			tc.op.Environment["CNAB_ACTION"] = tc.op.Action
			tc.op.Environment["CNAB_INSTALLATION_NAME"] = tc.op.Installation

			err := k.Run(tc.op)

			if tc.err != nil {
				assert.EqualError(t, err, tc.err.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.output, output.String())
		})
	}
}
