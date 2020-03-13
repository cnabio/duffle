module github.com/cnabio/duffle

go 1.13

require (
	cloud.google.com/go v0.53.0 // indirect
	github.com/Azure/go-autorest v13.3.3+incompatible // indirect
	github.com/Masterminds/semver v1.5.0
	github.com/Microsoft/hcsshim v0.8.7 // indirect
	github.com/Nvveen/Gotty v0.0.0-20120604004816-cd527374f1e5 // indirect
	github.com/cnabio/cnab-go v0.8.2-beta1
	github.com/containerd/cgroups v0.0.0-20200204152634-780d21166089 // indirect
	github.com/containerd/continuity v0.0.0-20200107194136-26c1120b8d41 // indirect
	github.com/containerd/cri v1.11.1 // indirect
	github.com/containerd/fifo v0.0.0-20191213151349-ff969a566b00 // indirect
	github.com/containerd/typeurl v0.0.0-20200205145503-b45ef1f1f737 // indirect
	github.com/docker/cli v0.0.0-20191017083524-a8ff7f821017
	github.com/docker/compose-on-kubernetes v0.4.17 // indirect
	github.com/docker/distribution v2.7.1-0.20190205005809-0d3efadf0154+incompatible
	github.com/docker/docker v1.4.2-0.20181229214054-f76d6a078d88
	github.com/docker/go v1.5.1-1
	github.com/docker/go-connections v0.4.0
	github.com/fatih/color v1.9.0 // indirect
	github.com/ghodss/yaml v1.0.0
	github.com/gogo/googleapis v1.3.2 // indirect
	github.com/google/gofuzz v1.1.0 // indirect
	github.com/googleapis/gnostic v0.4.0 // indirect
	github.com/gophercloud/gophercloud v0.8.0 // indirect
	github.com/gorilla/mux v1.7.4 // indirect
	github.com/gosuri/uitable v0.0.4
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79 // indirect
	github.com/hashicorp/go-version v1.2.0 // indirect
	github.com/imdario/mergo v0.3.8 // indirect
	github.com/json-iterator/go v1.1.9 // indirect
	github.com/mattn/go-isatty v0.0.12 // indirect
	github.com/mattn/go-runewidth v0.0.8 // indirect
	github.com/oklog/ulid v1.3.1
	github.com/opencontainers/go-digest v1.0.0-rc1
	github.com/petar/GoLLRB v0.0.0-20190514000832-33fb24c13b99 // indirect
	github.com/pivotal/image-relocation v0.0.0-20191111101224-e94aff6df06c
	github.com/pkg/errors v0.9.1
	github.com/prometheus/common v0.9.1 // indirect
	github.com/prometheus/procfs v0.0.8 // indirect
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.4.0
	github.com/technosophos/moniker v0.0.0-20180509230615-a5dbd03a2245
	golang.org/x/crypto v0.0.0-20200210222208-86ce3cb69678 // indirect
	golang.org/x/net v0.0.0-20200202094626-16171245cfb2
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0 // indirect
	gopkg.in/AlecAivazis/survey.v1 v1.8.8
	gopkg.in/yaml.v2 v2.2.8
	k8s.io/apimachinery v0.0.0-20191004115801-a2eda9f80ab8
)

replace github.com/docker/docker => github.com/moby/moby v0.7.3-0.20190826074503-38ab9da00309

replace golang.org/x/sys => golang.org/x/sys v0.0.0-20190830141801-acfa387b8d69
