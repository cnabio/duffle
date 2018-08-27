package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/cli/cli/command"
	cliconfig "github.com/docker/cli/cli/config"
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

const pushDesc = `
This command pushes a CNAB bundle's images to a container registry.
`

type pushCmd struct {
	out  io.Writer
	src  string
	home home.Home

	// options common to the docker client and the daemon.
	dockerClientOptions *dockerflags.ClientOptions
}

func newPushCmd(out io.Writer) *cobra.Command {
	var (
		push = &pushCmd{
			out:                 out,
			dockerClientOptions: dockerflags.NewClientOptions(),
		}
		f *pflag.FlagSet
	)

	cmd := &cobra.Command{
		Use:   "push [path]",
		Short: "pushes a CNAB bundle's images to a container registry",
		Long:  pushDesc,
		PersistentPreRun: func(c *cobra.Command, args []string) {
			push.dockerClientOptions.Common.SetDefaultOptions(f)
			dockerPreRun(push.dockerClientOptions)
		},
		RunE: func(_ *cobra.Command, args []string) (err error) {
			if len(args) > 0 {
				push.src = args[0]
			}
			if push.src == "" || push.src == "." {
				if push.src, err = os.Getwd(); err != nil {
					return err
				}
			}
			push.home = home.Home(homePath())
			return push.run()
		},
	}

	f = cmd.Flags()
	f.BoolVar(&push.dockerClientOptions.Common.Debug, "docker-debug", false, "Enable debug mode")
	f.StringVar(&push.dockerClientOptions.Common.LogLevel, "docker-log-level", "info", `Set the logging level ("debug"|"info"|"warn"|"error"|"fatal")`)
	f.BoolVar(&push.dockerClientOptions.Common.TLS, "docker-tls", defaultDockerTLS(), "Use TLS; implied by --tlsverify")
	f.BoolVar(&push.dockerClientOptions.Common.TLSVerify, fmt.Sprintf("docker-%s", dockerflags.FlagTLSVerify), defaultDockerTLSVerify(), "Use TLS and verify the remote")
	f.StringVar(&push.dockerClientOptions.ConfigDir, "docker-config", cliconfig.Dir(), "Location of client config files")

	push.dockerClientOptions.Common.TLSOptions = &tlsconfig.Options{
		CAFile:   filepath.Join(dockerCertPath, dockerflags.DefaultCaFile),
		CertFile: filepath.Join(dockerCertPath, dockerflags.DefaultCertFile),
		KeyFile:  filepath.Join(dockerCertPath, dockerflags.DefaultKeyFile),
	}

	tlsOptions := push.dockerClientOptions.Common.TLSOptions
	f.Var(opts.NewQuotedString(&tlsOptions.CAFile), "docker-tlscacert", "Trust certs signed only by this CA")
	f.Var(opts.NewQuotedString(&tlsOptions.CertFile), "docker-tlscert", "Path to TLS certificate file")
	f.Var(opts.NewQuotedString(&tlsOptions.KeyFile), "docker-tlskey", "Path to TLS key file")

	hostOpt := opts.NewNamedListOptsRef("docker-hosts", &push.dockerClientOptions.Common.Hosts, opts.ValidateHost)
	f.Var(hostOpt, "docker-host", "Daemon socket(s) to connect to")

	return cmd
}

func (b *pushCmd) run() (err error) {
	var (
		pushctx *builder.Context
		ctx     = context.Background()
		bldr    = builder.New()
	)
	bldr.LogsDir = b.home.Logs()
	if pushctx, err = builder.LoadWithEnv(b.src); err != nil {
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

	app, err := builder.PrepareBuild(bldr, pushctx)
	if err != nil {
		return err
	}

	bf := bundle.Bundle{Name: pushctx.Name}

	for _, c := range pushctx.DockerContexts {

		// TODO - add invocation image as top level field in duffle.toml
		if c.Name == "cnab" {
			bf.InvocationImage = bundle.InvocationImage{
				Image: c.Images[0],
				// TODO - handle image type
				ImageType: "docker",
			}
			bf.Version = strings.Split(c.Images[0], ":")[1]
			continue
		}
		bf.Images = append(bf.Images, bundle.Image{Name: c.Name, URI: c.Images[0]})
	}

	f, err := os.OpenFile("cnab/bundle.json", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}

	enc := json.NewEncoder(f)
	enc.SetIndent("", "    ")
	if err := enc.Encode(bf); err != nil {
		return err
	}

	progressC := bldr.Push(ctx, app, pushctx)
	cmdline.Display(ctx, pushctx.Name, progressC, cmdline.WithBuildID(bldr.ID))

	return nil
}
