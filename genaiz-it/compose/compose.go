package compose

import (
	"context"
	"embed"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/compose-spec/compose-go/v2/loader"
	"github.com/compose-spec/compose-go/v2/types"
	"github.com/docker/compose/v2/pkg/api"
)

var (
	//go:embed all:compose_res
	embeddedRes  embed.FS
	embeddedPath = "compose_res"
)

type baseService struct {
	ctx         context.Context
	environment map[string]string
	profiles    []string
}

func (bs baseService) loadProject(resource string) (*types.Project, error) {
	var path = filepath.Join(embeddedPath, resource) + ".yaml"
	var result *types.Project
	var currentUser *user.User
	var defBytes []byte
	var err error

	if currentUser, err = user.Current(); err != nil {
		return nil, err
	}

	if defBytes, err = embeddedRes.ReadFile(path); err == nil {
		var configFile = types.ConfigFile{
			Filename: filepath.Base(path),
			Content:  defBytes,
		}
		var details = types.ConfigDetails{
			WorkingDir:  ".",
			ConfigFiles: []types.ConfigFile{configFile},
			Environment: bs.environment,
		}

		if result, err = loader.LoadWithContext(bs.ctx, details, func(options *loader.Options) {
			options.SkipExtends = true
			options.SkipInclude = true
			options.Profiles = bs.profiles
			options.SetProjectName(resource, true)
		}); result != nil {
			for name, s := range result.Services {
				s.User = currentUser.Uid + ":" + currentUser.Gid
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

type Service interface {
	Init() error

	Start() error

	Stop() error
}

type Services interface {
	Feature(string) Service

	Orchestrator() Service

	Registry() Service
}

type serviceProvider func(context.Context) Service

type services struct {
	ctx context.Context

	registryProvider serviceProvider
}

func (s services) Feature(feature string) Service {
	return nil
}

func (s services) Orchestrator() Service {
	return nil
}

func (s services) Registry() Service {
	return s.registryProvider(s.ctx)
}

func NewServices(ctx context.Context) Services {
	return &services{
		ctx: ctx,

		registryProvider: NewRegistryService,
	}
}
