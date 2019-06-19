package main

import (
	"bytes"
	"testing"

	"github.com/deislabs/cnab-go/credentials"
)

func TestPrintCredentials(t *testing.T) {
	cs := &credentials.CredentialSet{
		Name: "foo",
		Credentials: []credentials.CredentialStrategy{
			{
				Name:   "password",
				Source: credentials.Source{Value: "TOPSECRET"},
			},
			{
				Name: "another-password",
			},
			{
				Name:   "kubeconfig",
				Source: credentials.Source{Path: "/root/.kube/config"},
			},
			{
				Name:   "some-setting",
				Source: credentials.Source{EnvVar: "MYSETTING"},
			},
		},
	}

	testcases := []struct {
		name       string
		unredacted bool
		output     string
	}{
		{name: "reacted", unredacted: false, output: `name: foo
credentials:
- name: password
  source:
    value: REDACTED
- name: another-password
  source: {}
- name: kubeconfig
  source:
    path: /root/.kube/config
- name: some-setting
  source:
    env: MYSETTING
`},
		{name: "unredacted", unredacted: true, output: `name: foo
credentials:
- name: password
  source:
    value: TOPSECRET
- name: another-password
  source: {}
- name: kubeconfig
  source:
    path: /root/.kube/config
- name: some-setting
  source:
    env: MYSETTING
`},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			output := &bytes.Buffer{}
			show := &credentialShowCmd{
				out:        output,
				unredacted: tc.unredacted,
			}

			err := show.printCredentials(*cs)
			if err != nil {
				t.Fatal(err)
			}

			want := tc.output
			got := output.String()
			if want != got {
				t.Fatalf("expected credentials output. WANT:\n%q\n\nGOT:\n%q\n", want, got)
			}
		})
	}
}
