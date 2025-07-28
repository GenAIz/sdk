package compose

import (
	"context"
	"embed"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/compose-spec/compose-go/v2/loader"
	"github.com/compose-spec/compose-go/v2/types"
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/flags"
	"github.com/docker/compose/v2/pkg/api"
	"github.com/docker/compose/v2/pkg/compose"

	"genaiz.com/genaiz-it/cucumber"
)

var (
	//go:embed all:compose_res
	embeddedRes  embed.FS
	embeddedPath = "compose_res"
)

type baseService struct {
	ctx      context.Context
	profiles []string
	user     *user.User
}

func (bs baseService) loadOptions(resource string) func(*loader.Options) {
	return func(options *loader.Options) {
		options.SkipExtends = true
		options.SkipInclude = true
		options.Profiles = bs.profiles
		options.SetProjectName(resource, true)
	}
}

func (bs baseService) loadProject(resource string) (*types.Project, error) {
	return bs.loadProjectWithEnv(resource, nil)
}

func (bs baseService) loadProjectWithEnv(resource string, environment map[string]string) (*types.Project, error) {
	var path = filepath.Join(embeddedPath, resource) + ".yaml"
	var result *types.Project
	var defBytes []byte
	var err error

	if defBytes, err = embeddedRes.ReadFile(path); err == nil {
		var details = types.ConfigDetails{
			WorkingDir: ".",
			ConfigFiles: []types.ConfigFile{{
				Content:  defBytes,
				Filename: filepath.Base(path),
			}},
			Environment: environment,
		}

		if result, err = loader.LoadWithContext(bs.ctx, details, bs.loadOptions(resource)); result != nil {
			for name, s := range result.Services {
				s.User = bs.user.Uid + ":" + bs.user.Gid
				s.CustomLabels = map[string]string{
					api.ProjectLabel:     result.Name,
					api.ServiceLabel:     s.Name,
					api.VersionLabel:     api.ComposeVersion,
					api.WorkingDirLabel:  "/",
					api.ConfigFilesLabel: strings.Join(result.ComposeFiles, ","),
					api.OneoffLabel:      "False", // default, will be overridden by `run` command
				}
				result.Services[name] = s
			}
		}
	}

	return result, err
}

func (bs baseService) mergeProjects(projects ...*types.Project) (*types.Project, error) {
	var result *types.Project

	if len(projects) > 0 {
		result = projects[0]

		for _, project := range projects {
			for serviceName, service := range project.Services {
				if _, ok := result.Services[serviceName]; !ok {
					service.Profiles = result.Profiles
					service.CustomLabels[api.ProjectLabel] = result.Name
					result.Services[serviceName] = service
				}
			}

			result.Environment.Merge(project.Environment)
		}
	}

	return result, nil
}

func (bs baseService) start(project *types.Project) error {
	var cli *command.DockerCli
	var err error

	if cli, err = command.NewDockerCli(); err == nil {
		var opts = flags.ClientOptions{Context: "default", LogLevel: "error"}

		if err = cli.Initialize(&opts); err == nil {
			var composeService = compose.NewComposeService(cli)

			err = composeService.Up(bs.ctx, project, api.UpOptions{
				Create: api.CreateOptions{
					Recreate:             api.RecreateForce,
					RecreateDependencies: api.RecreateForce,
				},
				Start: api.StartOptions{
					Project: project,
				},
			})
		}
	}

	return err
}

func (bs baseService) stop(project *types.Project) error {
	var cli *command.DockerCli
	var err error

	if cli, err = command.NewDockerCli(); err == nil {
		var opts = flags.ClientOptions{Context: "default", LogLevel: "error"}

		if err = cli.Initialize(&opts); err == nil {
			var composeService = compose.NewComposeService(cli)

			err = composeService.Down(bs.ctx, project.Name, api.DownOptions{
				RemoveOrphans: true,
				Project:       project,
				Volumes:       true,
			})
		}
	}

	return err
}

type Service interface {
	Init() error

	Start() error

	Stop() error
}

type Services interface {
	Feature(...cucumber.Feature) Service

	Genaiz(string, ...cucumber.Feature) Service

	Registry() Service

	Wiremock(...cucumber.Feature) Service
}

type genaizProvider func(context.Context, *user.User, string, ...cucumber.Feature) Service

type registryProvider func(context.Context, *user.User) Service

type wiremockProvider func(context.Context, *user.User, ...cucumber.Feature) Service

type services struct {
	ctx  context.Context
	user *user.User

	genaizProvider   genaizProvider
	registryProvider registryProvider
	wiremockProvider wiremockProvider
}

func (s services) Genaiz(version string, features ...cucumber.Feature) Service {
	return s.genaizProvider(s.ctx, s.user, version, features...)
}

func (s services) Feature(features ...cucumber.Feature) Service {
	return nil
}

func (s services) Registry() Service {
	return s.registryProvider(s.ctx, s.user)
}

func (s services) Wiremock(features ...cucumber.Feature) Service {
	return s.wiremockProvider(s.ctx, s.user, features...)
}

func NewServices(ctx context.Context) Services {
	var currentUser, _ = user.Current()
	return &services{
		ctx:  ctx,
		user: currentUser,

		genaizProvider:   NewGenaizService,
		registryProvider: NewRegistryService,
		wiremockProvider: NewWiremockService,
	}
}
