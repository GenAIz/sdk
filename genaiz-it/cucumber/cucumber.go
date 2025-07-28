package cucumber

import (
	"bytes"
	"embed"
	"path/filepath"
	"slices"
	"strings"

	gherkin "github.com/cucumber/gherkin/go/v28"
	messages "github.com/cucumber/messages/go/v24"

	"genaiz.com/genaiz-lib/lang/dirz"
)

var (
	//go:embed all:cucumber_res
	embeddedRes  embed.FS
	embeddedPath = "cucumber_res"
)

type execution struct {
	Name         string   `json:"name"`
	Scenarios    []string `json:"scenarios"`
	Orchestrator string   `json:"orchestrator"`
	Registry     string   `json:"registry"`
}

type Feature struct {
	execution

	resetFn func()
}

func (f *Feature) CreateWorkDir() error {
	var workDir = f.GetWorkDir()
	var err error

	f.resetFn, err = dirz.CreateWorkingDir(workDir)
	return err
}

func (f *Feature) Filename() string {
	return normalizedName(f.Name) + ".feature"
}

func (f *Feature) GetWorkDir() string {
	return "." + normalizedName(f.Name)
}

func (f *Feature) IsLoad(scenario string) bool {
	return slices.Contains(f.Scenarios, scenario)
}

func (f *Feature) IsLoadAll() bool {
	return len(f.Scenarios) == 0 || f.IsLoad("*")
}

func (f *Feature) OrchestratorType() string {
	if strings.HasPrefix(f.Orchestrator, "http") {
		return "ext"
	} else {
		return "mock"
	}
}

func (f *Feature) ResetWorkDir() {
	if f.resetFn != nil {
		f.resetFn()
	}
}

func GetDocument(path string) (*messages.GherkinDocument, error) {
	var resPath = filepath.Join(embeddedPath, path)
	var resBytes []byte
	var err error

	if resBytes, err = embeddedRes.ReadFile(resPath); err == nil {
		var reader = bytes.NewReader(resBytes)
		var uuid = &messages.UUID{}
		var result *messages.GherkinDocument

		if result, err = gherkin.ParseGherkinDocument(reader, uuid.NewId); err == nil {
			return result, nil
		}
	}

	return nil, err
}

func NewFeature(orchestrator string, name string, scenarios []string) *Feature {
	return &Feature{
		execution: execution{
			Name:         name,
			Scenarios:    scenarios,
			Orchestrator: orchestrator,
		},
	}
}

func normalizedName(name string) string {
	return strings.ReplaceAll(name, " ", "_")
}
