package compose

import (
	"context"
	"path/filepath"

	"github.com/compose-spec/compose-go/v2/types"
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/flags"
	"github.com/docker/compose/v2/pkg/api"
	"github.com/docker/compose/v2/pkg/compose"

	"genaiz.com/genaiz-lib/lang/dirz"
	"genaiz.com/genaiz-oauth/cert"
)

const (
	registryService = "registry"
	registryRes     = "registry-compose"
	registryWd      = ".registry"
)

type registry struct {
	baseService

	project *types.Project
}

func (r *registry) setupHost(authFolder string) error {
	var reset func()
	var err error

	if reset, err = dirz.CreateWorkingDir(authFolder); err == nil {
		defer reset()
		var arbiter = cert.NewArbiter().
			WithBundle("bundle.crt").
			WithAuthority("ca.cert", "ca.key").
			WithServer("server.cert", "server.key").
			WithCaCommonName("iss.genaiz.com").
			WithCaLifetime(1).
			WithCommonName("dev.genaiz.com").
			WithCountry("CA").
			WithLifetime(1).
			WithLocality("Montreal").
			WithOrganization("GenAIz").
			WithProvince("QC")

		if err = arbiter.BuildCert(); err == nil {
			if err = arbiter.BuildRootBundle(); err != nil {
				return err
			}
		}
	}

	return err
}

func (r *registry) Init() error {
	var authFolder = filepath.Join(registryWd, "auth")
	var err error

	if err = r.setupHost(authFolder); err == nil {
		if r.project, err = r.loadProject(registryRes); err == nil {
			var serviceConfig types.ServiceConfig

			if serviceConfig, err = r.project.GetService(registryService); err == nil {
				serviceConfig.Volumes = []types.ServiceVolumeConfig{
					{
						Type:   types.VolumeTypeBind,
						Source: authFolder,
						Target: r.environment["AUTH_PATH"],
						Bind: &types.ServiceVolumeBind{
							CreateHostPath: false,
						},
					},
				}
				r.project.Services[registryService] = serviceConfig
			}
		}
	}

	return err
}

func (r *registry) Start() error {
	var cli *command.DockerCli
	var err error

	if cli, err = command.NewDockerCli(); err == nil {
		var opts = flags.ClientOptions{Context: "default", LogLevel: "error"}

		if err = cli.Initialize(&opts); err == nil {
			var composeService = compose.NewComposeService(cli)

			err = composeService.Up(r.ctx, r.project, api.UpOptions{
				Create: api.CreateOptions{
					Recreate:             api.RecreateForce,
					RecreateDependencies: api.RecreateForce,
				},
				Start: api.StartOptions{
					Project: r.project,
				},
			})
		}
	}

	return err
}

func (r *registry) Stop() error {
	var cli *command.DockerCli
	var err error

	if r.project, err = r.loadProject(registryRes); err == nil {
		if cli, err = command.NewDockerCli(); err == nil {
			var opts = flags.ClientOptions{Context: "default", LogLevel: "error"}

			if err = cli.Initialize(&opts); err == nil {
				var composeService = compose.NewComposeService(cli)

				err = composeService.Down(r.ctx, r.project.Name, api.DownOptions{
					RemoveOrphans: true,
					Project:       r.project,
					Volumes:       true,
				})
			}
		}
	}

	return err
}

func NewRegistryService(ctx context.Context) Service {
	return &registry{
		baseService: baseService{
			ctx: ctx,
			environment: map[string]string{
				"AUTH_PATH": "/etc/distribution/auth",
			},
			profiles: []string{"services-registry"},
		},
	}
}
