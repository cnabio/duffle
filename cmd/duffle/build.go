package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/deis/duffle/pkg/cmdline"

	"github.com/deis/duffle/pkg/duffle/manifest"

	"github.com/deis/duffle/pkg/builder"
	"github.com/deis/duffle/pkg/builder/docker"
	"github.com/deis/duffle/pkg/duffle/home"

	"github.com/docker/cli/cli/command"
	cliconfig "github.com/docker/cli/cli/config"
	dockerdebug "github.com/docker/cli/cli/debug"
	dockerflags "github.com/docker/cli/cli/flags"
	"github.com/docker/cli/opts"
	"github.com/docker/go-connections/tlsconfig"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
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
		ctx  = context.Background()
		bldr = builder.New()
	)
	bldr.LogsDir = b.home.Logs()

	mfst, err := manifest.Load(filepath.Join(b.src, "duffle.toml"))
	if err != nil {
		return err
	}

	bldr.BundleBuilder, err = lookupBuilder(mfst.Builder, b)
	if err != nil {
		return fmt.Errorf("cannot lookup builder: %v", err)
	}

	app, bf, err := bldr.BundleBuilder.PrepareBuild(bldr, mfst, b.src)
	if err != nil {
		return fmt.Errorf("cannot prepare build: %v", err)
	}

	f, err := os.OpenFile("cnab/bundle.json", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("cannot create or open bundle file: %v", err)
	}

	enc := json.NewEncoder(f)
	enc.SetIndent("", "    ")
	if err := enc.Encode(bf); err != nil {
		return fmt.Errorf("cannot write bundle file: %v", err)
	}

	cmdline.Display(ctx, app.Ctx.Manifest.Name, bldr.BundleBuilder.Build(ctx, app), cmdline.WithBuildID(bldr.ID))
	return nil
}

// lookupBuilder takes a builder name and returns an appropriate builder
func lookupBuilder(b string, cmd *buildCmd) (builder.BundleBuilder, error) {

	var bb builder.BundleBuilder

	// setup docker
	cli := &command.DockerCli{}
	if err := cli.Initialize(cmd.dockerClientOptions); err != nil {
		return bb, fmt.Errorf("failed to create docker client: %v", err)
	}

	switch b {

	case "docker":
		bb = docker.Builder{
			DockerClient: cli,
		}
	default:
		bb = docker.Builder{
			DockerClient: cli,
		}
	}

	return bb, nil
}
