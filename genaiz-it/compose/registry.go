package compose

import (
	"context"
	"os/user"
	"path/filepath"

	"github.com/compose-spec/compose-go/v2/types"

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

func (r *registry) authFolder() string {
	return filepath.Join(registryWd, "auth")
}

func (r *registry) setupRegistryDir(authFolder string) error {
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

func (r *registry) setupRegistryProject(authFolder string) error {
	var environment = map[string]string{
		"AUTH_PATH": "/etc/distribution/auth",
	}
	var err error

	if r.project, err = r.loadProjectWithEnv(registryRes, environment); err == nil {
		var serviceConfig types.ServiceConfig

		if serviceConfig, err = r.project.GetService(registryService); err == nil {
			serviceConfig.Volumes = []types.ServiceVolumeConfig{
				{
					Type:   types.VolumeTypeBind,
					Source: authFolder,
					Target: environment["AUTH_PATH"],
					Bind: &types.ServiceVolumeBind{
						CreateHostPath: false,
					},
				},
			}
			r.project.Services[registryService] = serviceConfig
		}
	}

	return err
}

func (r *registry) Init() error {
	var authFolder = filepath.Join(registryWd, "auth")
	var err error

	if err = r.setupRegistryProject(authFolder); err == nil {
		err = r.setupRegistryDir(authFolder)
	}

	return err
}

func (r *registry) Start() error {
	return r.start(r.project)
}

func (r *registry) Stop() error {
	var err error
	var environment = map[string]string{
		"AUTH_PATH": r.authFolder(),
	}

	if r.project, err = r.loadProjectWithEnv(registryRes, environment); err == nil {
		err = r.stop(r.project)
	}

	return err
}

func NewRegistryService(ctx context.Context, user *user.User) Service {
	return newRegistryService(ctx, user)
}

func newRegistryService(ctx context.Context, user *user.User) *registry {
	return &registry{
		baseService: baseService{
			ctx:      ctx,
			profiles: []string{"services-registry"},
			user:     user,
		},
	}
}
