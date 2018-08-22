package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/docker/cli/cli/command"
	cliconfig "github.com/docker/cli/cli/config"
	dockerdebug "github.com/docker/cli/cli/debug"
	dockerflags "github.com/docker/cli/cli/flags"
	"github.com/docker/cli/opts"
	"github.com/docker/go-connections/tlsconfig"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"golang.org/x/net/context"

	"github.com/deis/duffle/pkg/builder"
	dockercontainerbuilder "github.com/deis/duffle/pkg/builder/docker"
	"github.com/deis/duffle/pkg/bundle"
	"github.com/deis/duffle/pkg/cmdline"
	"github.com/deis/duffle/pkg/duffle/home"
)

const buildDesc = `
This command builds a CNAB bundle.
`

const (
	dockerTLSEnvVar       = "DOCKER_TLS"
	dockerTLSVerifyEnvVar = "DOCKER_TLS_VERIFY"
)

var (
	dockerCertPath = os.Getenv("DOCKER_CERT_PATH")
)

type buildCmd struct {
	out  io.Writer
	src  string
	home home.Home
	// storage engine duffle should use for storing builds, logs, etc.
	storageEngine string
	// options common to the docker client and the daemon.
	dockerClientOptions *dockerflags.ClientOptions
}

func defaultDockerTLS() bool {
	return os.Getenv(dockerTLSEnvVar) != ""
}

func defaultDockerTLSVerify() bool {
	return os.Getenv(dockerTLSVerifyEnvVar) != ""
}

func dockerPreRun(opts *dockerflags.ClientOptions) {
	dockerflags.SetLogLevel(opts.Common.LogLevel)

	if opts.ConfigDir != "" {
		cliconfig.SetDir(opts.ConfigDir)
	}

	if opts.Common.Debug {
		dockerdebug.Enable()
	}
}

func newBuildCmd(out io.Writer) *cobra.Command {
	var (
		build = &buildCmd{
			out:                 out,
			dockerClientOptions: dockerflags.NewClientOptions(),
		}
		f *pflag.FlagSet
	)

	cmd := &cobra.Command{
		Use:   "build [path]",
		Short: "builds a CNAB bundle",
		Long:  buildDesc,
		PersistentPreRun: func(c *cobra.Command, args []string) {
			build.dockerClientOptions.Common.SetDefaultOptions(f)
			dockerPreRun(build.dockerClientOptions)
		},
		RunE: func(_ *cobra.Command, args []string) (err error) {
			if len(args) > 0 {
				build.src = args[0]
			}
			if build.src == "" || build.src == "." {
				if build.src, err = os.Getwd(); err != nil {
					return err
				}
			}
			build.home = home.Home(homePath())
			return build.run()
		},
	}

	f = cmd.Flags()
	f.BoolVar(&build.dockerClientOptions.Common.Debug, "docker-debug", false, "Enable debug mode")
	f.StringVar(&build.dockerClientOptions.Common.LogLevel, "docker-log-level", "info", `Set the logging level ("debug"|"info"|"warn"|"error"|"fatal")`)
	f.BoolVar(&build.dockerClientOptions.Common.TLS, "docker-tls", defaultDockerTLS(), "Use TLS; implied by --tlsverify")
	f.BoolVar(&build.dockerClientOptions.Common.TLSVerify, fmt.Sprintf("docker-%s", dockerflags.FlagTLSVerify), defaultDockerTLSVerify(), "Use TLS and verify the remote")
	f.StringVar(&build.dockerClientOptions.ConfigDir, "docker-config", cliconfig.Dir(), "Location of client config files")

	build.dockerClientOptions.Common.TLSOptions = &tlsconfig.Options{
		CAFile:   filepath.Join(dockerCertPath, dockerflags.DefaultCaFile),
		CertFile: filepath.Join(dockerCertPath, dockerflags.DefaultCertFile),
		KeyFile:  filepath.Join(dockerCertPath, dockerflags.DefaultKeyFile),
	}

	tlsOptions := build.dockerClientOptions.Common.TLSOptions
	f.Var(opts.NewQuotedString(&tlsOptions.CAFile), "docker-tlscacert", "Trust certs signed only by this CA")
	f.Var(opts.NewQuotedString(&tlsOptions.CertFile), "docker-tlscert", "Path to TLS certificate file")
	f.Var(opts.NewQuotedString(&tlsOptions.KeyFile), "docker-tlskey", "Path to TLS key file")

	hostOpt := opts.NewNamedListOptsRef("docker-hosts", &build.dockerClientOptions.Common.Hosts, opts.ValidateHost)
	f.Var(hostOpt, "docker-host", "Daemon socket(s) to connect to")

	return cmd
}

func (b *buildCmd) run() (err error) {
	var (
		buildctx *builder.Context
		ctx      = context.Background()
		bldr     = builder.New()
	)
	bldr.LogsDir = b.home.Logs()
	if buildctx, err = builder.LoadWithEnv(b.src); err != nil {
		return fmt.Errorf("failed loading build context: %v", err)
	}

	var cb builder.ContainerBuilder

	// setup docker
	cli := &command.DockerCli{}
	if err := cli.Initialize(b.dockerClientOptions); err != nil {
		return fmt.Errorf("failed to create docker client: %v", err)
	}
	cb = &dockercontainerbuilder.Builder{
		DockerClient: cli,
	}
	bldr.ContainerBuilder = cb

	progressC := bldr.Build(ctx, buildctx)
	cmdline.Display(ctx, buildctx.Name, progressC, cmdline.WithBuildID(bldr.ID))

	bf := bundle.Bundle{
		Name: buildctx.Name,
		// TODO - handle bundle version
		Version: "1.0.0",
	}
	for _, c := range buildctx.DockerContexts {

		// TODO - add invocation image as top level field in duffle.toml
		if c.Name == "cnab" {
			bf.InvocationImage = bundle.InvocationImage{
				Image: c.Images[0],
				// TODO - handle image type
				ImageType: "docker",
			}
			continue
		}
		bf.Images = append(bf.Images, bundle.Image{Name: c.Name, URI: c.Images[0]})
	}

	mb, err := json.Marshal(bf)
	if err != nil {
		return err
	}

	return ioutil.WriteFile("bundle.json", mb, 0644)
}
