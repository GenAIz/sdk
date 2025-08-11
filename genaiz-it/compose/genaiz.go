package compose

import (
	"context"
	"fmt"
	"os/user"
	"slices"
	"strings"

	"github.com/compose-spec/compose-go/v2/types"
	messages "github.com/cucumber/messages/go/v24"
	"github.com/docker/compose/v2/pkg/api"

	"genaiz.com/genaiz-it/cucumber"
	"genaiz.com/genaiz-lib/lang/stringz"
)

const (
	genaizRes = "genaiz-compose"
)

type test struct {
	alterations *alterations
	feature     *messages.Feature
	scenario    *messages.Scenario
	user        *user.User
}

func newTest(alts *alterations, feature *messages.Feature, scenario *messages.Scenario) *test {
	var result = &test{
		alterations: alts,
		feature:     feature,
		scenario:    scenario,
	}

	return result
}

func (t *test) toFeature() *cucumber.Feature {
	var normalized = normalizeName(t.scenario.Name)

	return cucumber.NewFeature(t.alterations.feature.Orchestrator, t.alterations.feature.Name, []string{normalized})
}

type genaiz struct {
	baseService

	includes []cucumber.Feature
	features []cucumber.Feature
	prefix   string
	project  *types.Project
	services []string
	version  string
}

func (g *genaiz) allFeatures() []cucumber.Feature {
	var result []cucumber.Feature

	result = append(result, g.features...)
	return append(result, g.includes...)
}

func (g *genaiz) loadDependency(alts *alterations, doc *messages.GherkinDocument, serviceName string) ([]*types.Project, error) {
	var scenarioProject *types.Project
	var scenarioTest *test
	var result []*types.Project
	var err error

	if scenarioTest, err = g.loadTest(alts, doc, serviceName); err == nil {
		g.includes = append(g.includes, *scenarioTest.toFeature())

		if scenarioProject, err = g.loadScenario(alts, scenarioTest); err == nil {
			var projects []*types.Project

			result = append(result, scenarioProject)

			if projects, err = g.loadDependencies(alts, doc, scenarioProject.Services); err == nil && len(projects) > 0 {
				result = append(result, projects...)
			} else if err != nil {
				return nil, err
			}
		}
	}

	return result, err
}

func (g *genaiz) loadDependencies(alts *alterations, doc *messages.GherkinDocument, services types.Services) ([]*types.Project, error) {
	var result []*types.Project
	var err error

	for _, serviceConfig := range services {
		for dependsOn, _ := range serviceConfig.DependsOn {
			var projects []*types.Project

			if strings.HasPrefix(dependsOn, "genaiz") {
				if projects, err = g.loadDependency(alts, doc, dependsOn); err == nil {
					result = append(result, projects...)
				}
			} else {
				g.services = append(g.services, dependsOn)
			}

			if err != nil {
				return nil, err
			}
		}
	}

	return result, nil
}

func (g *genaiz) loadFeature(feature *cucumber.Feature) ([]*types.Project, error) {
	var doc *messages.GherkinDocument
	var result []*types.Project
	var err error

	if doc, err = cucumber.GetDocument(feature.Filename()); err == nil {
		var testDir = feature.GetWorkDir()

		if err = feature.CreateWorkDir(); err == nil {
			var testAlterations = newAlterations(feature, testDir)
			var project *types.Project

			for _, scenarioTest := range g.loadTests(testAlterations, doc) {
				if project, err = g.loadScenario(testAlterations, scenarioTest); err == nil {
					result = append(result, project)
				} else {
					return nil, err
				}
			}

			for _, project = range result {
				var dependencyProjects []*types.Project

				if dependencyProjects, err = g.loadDependencies(testAlterations, doc, project.Services); err == nil {
					result = append(result, dependencyProjects...)
				} else {
					return nil, err
				}
			}

			feature.ResetWorkDir()
		}
	}

	return result, err
}

func (g *genaiz) loadProjects() ([]*types.Project, error) {
	var result []*types.Project

	for _, feature := range g.features {
		if featureProjects, err := g.loadFeature(&feature); err == nil {
			result = append(result, featureProjects...)
		} else {
			return nil, err
		}
	}

	return result, nil
}

func (g *genaiz) loadScenario(alts *alterations, test *test) (*types.Project, error) {
	var orchestratorType = alts.feature.OrchestratorType()
	var environment = map[string]string{
		"CONTAINER": normalizeContainer(orchestratorType, test.feature.Name, test.scenario.Name),
		"VERSION":   g.version,
	}
	var project *types.Project
	var err error

	if project, err = g.loadProjectWithEnv(genaizRes, environment); err == nil {
		var modifiedServices = map[string]types.ServiceConfig{}

		test.alterations.flushParams()

		for _, serviceConfig := range project.Services {
			var steps = cucumber.NewSteps(test.alterations)

			if err = steps.Visit(test.scenario.Steps); err != nil {
				return nil, err
			}

			serviceConfig.User = test.alterations.flushUser(serviceConfig.User)
			serviceConfig.Name = normalizeService(orchestratorType, test.feature.Name, test.scenario.Name)
			serviceConfig.CustomLabels[api.ServiceLabel] = serviceConfig.Name
			serviceConfig.Command = test.alterations.flushCommand()
			serviceConfig.Environment = test.alterations.flushEnvironment()
			serviceConfig.DependsOn = test.alterations.flushDependsOnConfig()
			serviceConfig.Volumes = []types.ServiceVolumeConfig{
				{
					Type:   types.VolumeTypeBind,
					Source: alts.workDir,
					Target: "/home/genaiz",
					Bind: &types.ServiceVolumeBind{
						CreateHostPath: false,
					},
				},
				{
					Type:   types.VolumeTypeBind,
					Source: "/var/run/docker.sock",
					Target: "/var/run/docker.sock",
					Bind: &types.ServiceVolumeBind{
						CreateHostPath: false,
					},
				},
			}
			modifiedServices[serviceConfig.Name] = serviceConfig
		}

		project.Services = modifiedServices
		return project, nil
	}

	return nil, err
}

func (g *genaiz) loadServices() ([]*types.Project, error) {
	var result []*types.Project
	var err error

	if slices.Contains(g.services, "registry") {
		var rs = newRegistryService(g.ctx, g.user)

		if err = rs.Init(); err == nil {
			result = append(result, rs.project)
		}
	}

	if slices.Contains(g.services, "wiremock") {
		var ws = newWiremockService(g.ctx, g.user, g.allFeatures())

		if err = ws.Init(); err == nil {
			result = append(result, ws.project)
		}
	}

	return result, nil
}

func (g *genaiz) loadTest(alts *alterations, doc *messages.GherkinDocument, scenario string) (*test, error) {
	var orchestratorType = alts.feature.OrchestratorType()
	var scenarioChild *messages.Scenario

	for _, child := range doc.Feature.Children {
		if child.Scenario != nil {
			var normalized = normalizeService(orchestratorType, doc.Feature.Name, child.Scenario.Name)

			if normalized == scenario {
				scenarioChild = child.Scenario
				break
			}
		}
	}

	if scenarioChild == nil {
		return nil, fmt.Errorf("could not resolve scenario service name [%s]", scenario)
	}

	return newTest(alts, doc.Feature, scenarioChild), nil
}

func (g *genaiz) loadTests(alts *alterations, doc *messages.GherkinDocument) []*test {
	var loadAll = alts.feature.IsLoadAll()
	var result []*test

	for _, child := range doc.Feature.Children {
		if child.Scenario != nil {
			if loadAll || alts.feature.IsLoad(normalizeName(child.Scenario.Name)) {
				result = append(result, newTest(alts, doc.Feature, child.Scenario))
			}
		}
	}

	return result
}

func (g *genaiz) Init() error {
	var projects []*types.Project
	var err error

	if projects, err = g.loadProjects(); err == nil {
		var serviceProjects []*types.Project

		if serviceProjects, err = g.loadServices(); err == nil {
			var project *types.Project

			if project, err = g.mergeProjects(append(projects, serviceProjects...)...); err == nil {
				g.project = project
			}
		}
	}

	return err
}

func (g *genaiz) Start() error {
	return g.start(g.project)
}

func (g *genaiz) Stop() error {
	var projects []*types.Project
	var err error

	if projects, err = g.loadProjects(); err == nil {
		var project *types.Project

		if slices.Contains(g.services, "registry") {
			var rs = newRegistryService(g.ctx, g.user)

			if err = rs.setupRegistryProject(registryWd); err != nil {
				return err
			}

			projects = append(projects, rs.project)
		}

		if slices.Contains(g.services, "wiremock") {
			var ws = newWiremockService(g.ctx, g.user, g.allFeatures())

			if err = ws.setupWiremockProject(); err != nil {
				return err
			}

			projects = append(projects, ws.project)
		}

		if project, err = g.mergeProjects(projects...); err == nil {
			g.project = project
		}

		return g.stop(g.project)
	}

	return err
}

func NewGenaizService(ctx context.Context, user *user.User, version string, features ...cucumber.Feature) Service {
	return &genaiz{
		baseService: baseService{
			ctx:      ctx,
			profiles: []string{"services-genaiz"},
			user:     user,
		},
		features: features,
		version:  version,
	}
}

type alterations struct {
	command     string
	environment []string
	execGroup   string
	execUser    string
	feature     *cucumber.Feature
	internals   map[string]string
	params      map[string]string
	scenarios   map[string]string
	services    map[string]string

	workDir string
}

func (a *alterations) getCommand() []string {
	var mapper = a.getMapper()
	var result []string

	for _, token := range strings.Split(a.command, " ") {
		result = append(result, mapper(token))
	}

	return result
}

func (a *alterations) getDependsOnConfig() types.DependsOnConfig {
	var orchestratorType = a.feature.OrchestratorType()
	var result = map[string]types.ServiceDependency{}

	for k, v := range a.scenarios {
		var scenarioName = normalizeService(orchestratorType, a.feature.Name, k)

		result[scenarioName] = types.ServiceDependency{Condition: v}
	}

	for k, v := range a.services {
		result[k] = types.ServiceDependency{Condition: v}
	}

	return result
}

func (a *alterations) getEnvironment() types.MappingWithEquals {
	return types.NewMappingWithEquals(a.environment)
}

func (a *alterations) getMapper() func(string) string {
	return func(s string) string {
		var result = s

		for k, v := range a.params {
			result = strings.ReplaceAll(result, "<"+k+">", v)
		}

		return result
	}
}

func (a *alterations) flushCommand() []string {
	var result = a.getCommand()

	a.command = ""
	return result
}

func (a *alterations) flushDependsOnConfig() types.DependsOnConfig {
	var result = a.getDependsOnConfig()

	a.scenarios = make(map[string]string)
	a.services = make(map[string]string)
	return result
}

func (a *alterations) flushEnvironment() types.MappingWithEquals {
	var result = a.getEnvironment()

	a.environment = nil
	return result
}

func (a *alterations) flushParams() {
	a.params = make(map[string]string)

	for k, v := range a.internals {
		a.params[k] = v
	}
}

func (a *alterations) flushUser(user string) string {
	var currentPair = strings.Split(user, ":")
	var alteredUser = stringz.FirstNonEmpty(a.execUser, currentPair[0])
	var alteredGroup = stringz.FirstNonEmpty(a.execGroup, currentPair[1])
	var mapper = a.getMapper()

	return mapper(alteredUser + ":" + alteredGroup)
}

func (a *alterations) AProvisionedFunctionWithToken() {
	if a.feature.OrchestratorType() == "mock" {

	}
}

func (a *alterations) IRunTheCommand(command string) {
	a.command = command
}

func (a *alterations) IShouldHaveASessionIdWithHostForUsername(host string, username string) {
	var mapper = a.getMapper()

	fmt.Printf("I should have a session id with host [%s] and username [%s]\n",
		mapper(host), mapper(username))
}

func (a *alterations) IShouldHaveADockerImageTagged(tag string) {
	var mapper = a.getMapper()

	fmt.Printf("I should have a docker image tagged [%s]\n", mapper(tag))
}

func (a *alterations) IShouldHaveAFunctionNamed(name string) {
	var mapper = a.getMapper()

	fmt.Printf("I should have a function named [%s]\n", mapper(name))
}

func (a *alterations) IShouldHaveASessionId() {
	fmt.Println("I should have a session id")
}

func (a *alterations) TheEnvironmentContains(keyPair string) {
	a.environment = append(a.environment, keyPair)
}

func (a *alterations) TheExecutionGroup(name string) {
	a.execGroup = name
}

func (a *alterations) TheExecutionUser(name string) {
	a.execUser = name
}

func (a *alterations) TheFollowingParameters(params map[string]string) {
	for k, v := range params {
		a.params[k] = v
	}
}

func (a *alterations) TheOrchestratorIsRunningWithCondition(condition string) {
	if a.feature.OrchestratorType() == "ext" {
		a.services[proxyService] = condition
	} else {
		a.services[a.feature.Orchestrator] = condition
	}
}

func (a *alterations) TheRegistryIsRunningWithCondition(condition string) {
	a.services[registryService] = condition
}

func (a *alterations) TheScenarioRanWithCondition(scenario string, condition string) {
	a.scenarios[scenario] = condition
}

func newAlterations(feature *cucumber.Feature, workDir string) *alterations {
	var result = &alterations{
		feature:   feature,
		internals: newInternals(feature),
		params:    make(map[string]string),
		scenarios: make(map[string]string),
		services:  make(map[string]string),
		workDir:   workDir,
	}

	return result
}

func newInternals(feature *cucumber.Feature) map[string]string {
	var internals = make(map[string]string)
	var group, _ = user.LookupGroup("docker")

	// Because we connect to /var/run/docker.sock
	if group != nil {
		internals["docker_gid"] = group.Gid
	}

	if strings.HasPrefix(feature.Orchestrator, "http") ||
		strings.Index(feature.Orchestrator, ":") > 0 {
		internals["orchestrator"] = feature.Orchestrator
	} else {
		internals["orchestrator"] = feature.Orchestrator + ":8080"
	}

	return internals
}

func normalizeContainer(words ...string) string {
	return normalizeGenaiz("_", words...)
}

func normalizeName(name ...string) string {
	var words []string

	return normalizeWords("_", append(words, name...))
}

func normalizeService(words ...string) string {
	return normalizeGenaiz("-", words...)
}

func normalizeGenaiz(sep string, words ...string) string {
	return normalizeWords(sep, append([]string{"genaiz"}, words...))
}

func normalizeWords(sep string, words []string) string {
	var normalized []string

	for _, part := range words {
		var normalizedPart = strings.ReplaceAll(part, "_", sep)

		normalized = append(normalized, strings.ReplaceAll(normalizedPart, " ", sep))
	}

	return strings.Join(normalized, sep)
}
