package compose

import (
	"context"
	"os/user"
	"path/filepath"

	"github.com/compose-spec/compose-go/v2/types"

	"genaiz.com/genaiz-it/cucumber"
	wm "genaiz.com/genaiz-it/wiremock"
	"genaiz.com/genaiz-lib/lang/dirz"
)

const (
	wiremockFiles    = "__files"
	wiremockMappings = "mappings"
	wiremockRes      = "wiremock-compose"
	wiremockRoot     = ".wiremock"
	wiremockService  = "wiremock"
)

type wiremock struct {
	baseService

	features []cucumber.Feature
	project  *types.Project
}

func (w *wiremock) featurePrefixes() []string {
	var result []string

	for _, feature := range w.features {
		if len(feature.Scenarios) > 0 {
			for _, scenario := range feature.Scenarios {
				result = append(result, feature.Name+"_"+scenario)
			}
		} else {
			// all scenarios in the feature is the default
			result = append(result, feature.Name)
		}
	}

	return result
}

func (w *wiremock) setupWiremockDir() error {
	var filesDir = filepath.Join(wiremockRoot, wiremockFiles)
	var mappingsDir = filepath.Join(wiremockRoot, wiremockMappings)
	var reset func()
	var err error

	if reset, err = dirz.CreateWorkingDir(filesDir); err == nil {
		defer reset()

		if err = wm.CopyFiles("."); err == nil {
			reset()

			if reset, err = dirz.CreateWorkingDir(mappingsDir); err == nil {
				defer reset()
				err = wm.CopyMappings(".", w.featurePrefixes())
			}
		}
	}

	return err
}

func (w *wiremock) setupWiremockProject() error {
	var err error

	if w.project, err = w.loadProject(wiremockRes); err == nil {
		var serviceConfig types.ServiceConfig

		if serviceConfig, err = w.project.GetService(wiremockService); err == nil {
			serviceConfig.Volumes = []types.ServiceVolumeConfig{
				{
					Type:   types.VolumeTypeBind,
					Source: wiremockRoot,
					Target: "/home/wiremock",
					Bind: &types.ServiceVolumeBind{
						CreateHostPath: false,
					},
				},
			}
			w.project.Services[wiremockService] = serviceConfig
		}
	}

	return err
}

func (w *wiremock) Init() error {
	var err error

	if err = w.setupWiremockProject(); err == nil {
		err = w.setupWiremockDir()
	}

	return err
}

func (w *wiremock) Start() error {
	return w.start(w.project)
}

func (w *wiremock) Stop() error {
	var err error

	if w.project, err = w.loadProject(wiremockRes); err == nil {
		err = w.stop(w.project)
	}

	return err
}

func NewWiremockService(ctx context.Context, user *user.User, features ...cucumber.Feature) Service {
	return newWiremockService(ctx, user, features)
}

func newWiremockService(ctx context.Context, user *user.User, features []cucumber.Feature) *wiremock {
	return &wiremock{
		baseService: baseService{
			ctx:      ctx,
			profiles: []string{"services-wiremock"},
			user:     user,
		},
		features: features,
	}
}
