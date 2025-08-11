package compose

import (
	"context"
	"os"
	"os/user"
	"path/filepath"

	"github.com/compose-spec/compose-go/v2/types"

	"genaiz.com/genaiz-lib/lang/dirz"
	"genaiz.com/genaiz-oauth/cert"
	"genaiz.com/genaiz-oauth/jwt"
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

func (r *registry) authFolder(rootFolder string) string {
	return filepath.Join(rootFolder, "auth")
}

func (r *registry) libFolder(authFolder string) string {
	return filepath.Join(authFolder, "lib")
}

func (r *registry) setupRegistryDir(rootFolder string) error {
	var authFolder = r.authFolder(rootFolder)
	var libFolder = r.libFolder(rootFolder)
	var reset func()
	var err error

	if err = os.MkdirAll(libFolder, 0750); err != nil {
		return err
	}

	if reset, err = dirz.CreateWorkingDir(authFolder); err == nil {
		defer reset()
		var arbiter = cert.NewArbiter().
			WithBundle("bundle.crt").
			WithAuthority("ca.cert", "ca.key").
			WithServer("signing.cert", "signing.key").
			WithCaCommonName("iss.genaiz.com").
			WithCaLifetime(1).
			WithCommonName("dev.genaiz.com").
			WithCountry("CA").
			WithLifetime(1).
			WithLocality("Montreal").
			WithOrganization("GenAIz").
			WithProvince("QC")

		if err = arbiter.BuildCert(); err == nil {
			if err = arbiter.BuildRootBundle(); err == nil {
				var keyManager = jwt.NewKeyManager().
					WithSetFile("keys.jwks").
					WithPemKeys("signing.cert")

				err = keyManager.WriteKeySet()
			}
		}
	}

	return err
}

func (r *registry) setupRegistryProject(rootFolder string) error {
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
					Source: r.authFolder(rootFolder),
					Target: environment["AUTH_PATH"],
					Bind: &types.ServiceVolumeBind{
						CreateHostPath: false,
					},
				},
				{
					Type:   types.VolumeTypeBind,
					Source: r.libFolder(rootFolder),
					Target: "/var/lib/registry",
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
	var err error

	if err = r.setupRegistryProject(registryWd); err == nil {
		err = r.setupRegistryDir(registryWd)
	}

	return err
}

func (r *registry) Start() error {
	return r.start(r.project)
}

func (r *registry) Stop() error {
	var err error
	var environment = map[string]string{
		"AUTH_PATH": r.authFolder(registryWd),
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
