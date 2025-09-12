package sn

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"genaiz.com/genaiz/cmd/sf"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/lang"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
	"genaiz.com/genaiz/task/docker"
	"genaiz.com/genaiz/task/layout"
	"genaiz.com/genaiz/task/shared"
)

type SolutionPublishTaskFactory func() *task.Task[broker.SolutionPublishParams]

type FunctionOptions struct {
	optionArches      *config.ListOption
	optionDescription *config.StringOption
	optionHandle      *config.StringOption
	optionName        *config.StringOption
	optionOem         *config.StringOption
	optionType        *config.StringOption
	optionVersion     *config.StringOption
}

type FunctionParams struct {
	buildParams     *docker.BuildParams
	provisionParams *broker.ProvisionParams
	pushParams      *docker.PushParams
}

func (fo FunctionOptions) allDefiners() []config.Definer {
	return []config.Definer{
		fo.optionArches,
		fo.optionDescription,
		fo.optionHandle,
		fo.optionName,
		fo.optionOem,
		fo.optionType,
		fo.optionVersion,
	}
}

type PublishExecutor struct {
	BaseExecutor
	*PublishOptions

	cmd            *cobra.Command
	brokerAddr     string
	solutionReader *config.SolutionReader

	inspectTaskFactory         sf.BuildTaskFactory
	provisionTaskFactory       sf.ProvisionTaskFactory
	publishTaskFactory         sf.PublishTaskFactory
	pushTaskFactory            sf.PushTaskFactory
	solutionPublishTaskFactory SolutionPublishTaskFactory
}

func (pe *PublishExecutor) Display() {
	var configType *shared.ConfigType
	var err error

	if configType, err = pe.Ledger.GetConfigType(pe.optionConfigType); err == nil {
		if err = pe.solutionReader.WithConfigPath(pe.folderPath).Read(*configType); err == nil {
			var layoutKeys = map[string]string{}
			var solution = pe.solutionReader.GetSolution()
			var solutionKey = "solutionFile"

			if solution != nil {
				pe.Ledger.InitValue(pe.optionVersion, solution.Version)
				solutionKey = solution.Handle
			}

			layoutKeys[solutionKey] = pe.solutionReader.GetSolutionFile()

			for key, vp := range pe.solutionReader.FindFunctionValues() {
				var value *broker.Function

				if err = vp.UnmarshalKey("Sf.Publish", &value); err == nil && value != nil {
					layoutKeys[key] = value.Handle
				} else {
					pe.Ledger.Logger.Warnf("could not extract function publishing data from file %s", key)
				}
			}

			pe.Ledger.DisplayOptionsWithMap(&layoutKeys,
				&pe.PublishOptions.optionVersion.Option,
			)
		}
	}

	lang.HandleExit(err)
}

func (pe *PublishExecutor) Pretend() {
	var err = pe.collectAndCall(func(snParams *broker.SolutionPublishParams, fnParams []FunctionParams) {
		var plan = task.NewPlan("Publish", pe.Ledger.Logger)
		var pretenders []task.Worker

		for _, fp := range fnParams {
			pretenders = append(pretenders, task.NewPretender(fp.buildParams, pe.inspectTaskFactory()))
			pretenders = append(pretenders, task.NewPretender(fp.provisionParams, pe.provisionTaskFactory()))
			pretenders = append(pretenders, task.NewPretender(fp.pushParams, pe.pushTaskFactory()))
			pretenders = append(pretenders, task.NewPretender(fp.provisionParams, pe.publishTaskFactory()))
		}

		pretenders = append(pretenders, task.NewPretender(snParams, pe.solutionPublishTaskFactory()))
		plan.ContinueOnFailure = true
		plan.Sequence(pretenders...)
	})

	lang.HandleExit(err)
}

func (pe *PublishExecutor) Proceed() {
	var err = pe.collectAndCall(func(snParams *broker.SolutionPublishParams, fnParams []FunctionParams) {
		var plan = task.NewPlan("Publish", pe.Ledger.Logger)
		var workers []task.Worker

		for _, fp := range fnParams {
			workers = append(workers, task.NewWorker(fp.buildParams, pe.inspectTaskFactory()))
			workers = append(workers, task.NewWorker(fp.provisionParams, pe.provisionTaskFactory()))
			workers = append(workers, task.NewWorker(fp.pushParams, pe.pushTaskFactory()))
			workers = append(workers, task.NewWorker(fp.provisionParams, pe.publishTaskFactory()))
		}

		workers = append(workers, task.NewWorker(snParams, pe.solutionPublishTaskFactory()))
		plan.Sequence(workers...)
	})

	lang.HandleExit(err)
}

func (pe *PublishExecutor) collectAndCall(fn func(*broker.SolutionPublishParams, []FunctionParams)) error {
	var configParams = pe.makeConfigParams(pe.optionConfigType)
	var reader = pe.solutionReader.WithConfigPath(pe.folderPath)
	var err error

	if err = reader.Read(*configParams.ConfigType); err == nil {
		if solution := reader.GetSolution(); solution != nil {
			var fnParams []FunctionParams
			var snParams *broker.SolutionPublishParams

			for k, values := range reader.FindFunctionValues() {
				fnParams = append(fnParams, FunctionParams{
					buildParams:     pe.makeFunctionBuildParams(filepath.Dir(k), values),
					provisionParams: pe.makeFunctionProvisionParams(values, solution),
					pushParams:      pe.makeFunctionPushParams(),
				})
			}

			fnParams = pe.filterPublishedFunctions(solution, fnParams)
			snParams = pe.makeSolutionPublishParams(solution, fnParams)
			fn(snParams, fnParams)
			return nil
		} else {
			return fmt.Errorf("no solution could be read from [%s]", reader.GetSolutionFile())
		}
	}

	return err
}

func (pe *PublishExecutor) filterPublishedFunctions(solution *broker.Solution, fnParams []FunctionParams) []FunctionParams {
	var filtered []FunctionParams

	for _, fp := range fnParams {
		var path = fmt.Sprintf("%s/%s", fp.provisionParams.Oem, fp.provisionParams.Handle)

		if slices.ContainsFunc(solution.Workflows, func(workflow broker.Workflow) bool {
			return slices.ContainsFunc(workflow.Nodes, func(node broker.WorkflowNode) bool {
				if node.Sf != nil {
					var nodePath = fmt.Sprintf("%s/%s", node.Sf.Oem, node.Sf.Handle)

					return path == nodePath
				}

				return false
			})
		}) {
			filtered = append(filtered, fp)
		}
	}

	return filtered
}

func (pe *PublishExecutor) makeFunctionBuildParams(path string, vp *viper.Viper) *docker.BuildParams {
	var confDockerContext = vp.GetString("Sf.DockerContext")

	if confDockerContext == "" {
		confDockerContext = path
	}

	return &docker.BuildParams{
		Env: task.Env{
			Context: pe.Context,
		},
		DockerContext: confDockerContext,
		Dockerfile:    vp.GetString("Sf.Dockerfile"),
		DockerTag:     vp.GetString("Sf.Build.Tag"),
		DockerVersion: vp.GetString("Sf.Build.Version"),
	}
}

func (pe *PublishExecutor) makeFunctionProvisionParams(vp *viper.Viper, solution *broker.Solution) *broker.ProvisionParams {
	var ledger = config.NewBuilder().WithViper(vp).Build()
	var options = NewFunctionOptions(solution)

	ledger.Register(pe.cmd, options.allDefiners()...)
	ledger.InitDefaults()
	return &broker.ProvisionParams{
		Broker: broker.Broker{
			AuthFile: pe.Ledger.AuthFile,
			HostAddr: pe.brokerAddr,
		},
		Arches:      ledger.GetList(options.optionArches),
		Description: ledger.GetString(options.optionDescription),
		Handle:      ledger.GetString(options.optionHandle),
		Name:        ledger.GetString(options.optionName),
		Oem:         ledger.GetString(options.optionOem),
		Type:        ledger.GetString(options.optionType),
		Version:     ledger.GetString(options.optionVersion),
	}
}

func (pe *PublishExecutor) makeFunctionPushParams() *docker.PushParams {
	return &docker.PushParams{
		Env: task.Env{
			Context: pe.Context,
		},
	}
}

func (pe *PublishExecutor) makeSolutionPublishParams(solution *broker.Solution, fnParams []FunctionParams) *broker.SolutionPublishParams {
	var provisions []broker.ProvisionParams

	pe.Ledger.InitValue(pe.optionDescription, solution.Description)
	pe.Ledger.InitValue(pe.optionHandle, solution.Handle)
	pe.Ledger.InitValue(pe.optionOem, solution.Oem)
	pe.Ledger.InitValue(pe.optionName, solution.Name)
	pe.Ledger.InitValue(pe.optionVersion, solution.Version)

	for _, fp := range fnParams {
		provisions = append(provisions, *fp.provisionParams)
	}

	return &broker.SolutionPublishParams{
		Broker: broker.Broker{
			AuthFile: pe.Ledger.AuthFile,
			HostAddr: pe.brokerAddr,
		},
		Solution: &broker.Solution{
			Description: pe.Ledger.GetString(pe.optionDescription),
			Handle:      pe.Ledger.GetString(pe.optionHandle),
			Oem:         pe.Ledger.GetString(pe.optionOem),
			Name:        pe.Ledger.GetString(pe.optionName),
			Version:     pe.Ledger.GetString(pe.optionVersion),
			Workflows:   solution.Workflows,
		},
		Provisions: provisions,
	}
}

type PublishOptions struct {
	optionConfigType  *config.StringOption
	optionDescription *config.StringOption
	optionHandle      *config.StringOption
	optionName        *config.StringOption
	optionOem         *config.StringOption
	optionVersion     *config.StringOption
}

func (po PublishOptions) allDefiners() []config.Definer {
	return []config.Definer{
		po.optionConfigType,
		po.optionDescription,
		po.optionHandle,
		po.optionName,
		po.optionOem,
		po.optionVersion,
	}
}

func NewPublish(ledger *config.Ledger, snCli *Cli) *cobra.Command {
	var solutionReader = config.NewSolutionReader(ledger)
	var publishOptions = NewPublishOptions()
	var publishCmd = &cobra.Command{
		Use:     "publish HOST",
		Short:   "Publishes a solution",
		Long:    "Publishes a solution and all smart functions found under the solution path or the current working directory if not specified",
		Example: "genaiz sn publish solution-1 --version=0.1.1",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if path, err := os.Getwd(); err == nil {
				snCli.Exec(ledger, NewPublishExecutor(cmd, ledger, snCli, publishOptions,
					solutionReader, args[0], path))
			} else {
				lang.HandleExit(err)
			}
		},
	}

	ledger.Register(publishCmd, publishOptions.allDefiners()...)
	return publishCmd
}

func NewPublishExecutor(cmd *cobra.Command, ledger *config.Ledger, cli *Cli,
	options *PublishOptions, reader *config.SolutionReader, brokerAddr, folderPath string) *PublishExecutor {

	return &PublishExecutor{
		BaseExecutor: BaseExecutor{
			Cli:        cli,
			Ledger:     ledger,
			Context:    cmd.Context(),
			folderPath: folderPath,
		},
		PublishOptions: options,

		cmd:            cmd,
		brokerAddr:     brokerAddr,
		solutionReader: reader,

		inspectTaskFactory:         docker.NewInspectTask,
		provisionTaskFactory:       broker.NewFunctionProvisionTask,
		publishTaskFactory:         broker.NewFunctionPublishTask,
		pushTaskFactory:            docker.NewPushTask,
		solutionPublishTaskFactory: broker.NewSolutionPublishTask,
	}
}

func NewFunctionOptions(solution *broker.Solution) *FunctionOptions {
	var optionHandle = newFunctionHandleOption()
	var optionName = newFunctionNameOption(optionHandle)

	return &FunctionOptions{
		optionArches:      newFunctionArchesOption(),
		optionDescription: newFunctionDescriptionOption(optionName),
		optionHandle:      optionHandle,
		optionName:        optionName,
		optionOem:         newFunctionOemOption(solution.Oem),
		optionVersion:     newFunctionVersionOption(solution.Version),
		optionType:        newFunctionTypeOption(),
	}
}

func NewPublishOptions() *PublishOptions {
	var cmd = "Publish"
	var optionHandle = newOptionHandle(cmd)
	var optionName = newOptionName(optionHandle, cmd)
	var optionOem = newOptionOem(cmd)
	var optionVersion = newOptionVersion(cmd)

	optionOem.Validator = config.Validation.Oem
	optionVersion.Validator = config.Validation.Version
	return &PublishOptions{
		optionConfigType:  newOptionConfigType(cmd),
		optionDescription: newOptionDescription(optionName, cmd),
		optionHandle:      optionHandle,
		optionName:        optionName,
		optionOem:         optionOem,
		optionVersion:     optionVersion,
	}
}

func newFunctionArchesOption() *config.ListOption {
	return &config.ListOption{
		Option: config.Option{
			Key:       "Sf.Publish.Arches",
			Validator: config.AllFromEnumerated(layout.ArchTypes),
		},
	}
}

func newFunctionDescriptionOption(defaultOption *config.StringOption) *config.StringOption {
	return &config.StringOption{
		Option: config.Option{
			Key: "Sf.Publish.Description",
			DefaultGetter: func(ledger *config.Ledger) any {
				return ledger.GetString(defaultOption)
			},
			Validator: config.Validation.Blob,
		},
	}
}

func newFunctionNameOption(defaultOption *config.StringOption) *config.StringOption {
	return &config.StringOption{
		Option: config.Option{
			Key: "Sf.Publish.Name",
			DefaultGetter: func(ledger *config.Ledger) any {
				return ledger.GetString(defaultOption)
			},
			Validator: config.Validation.Name,
		},
	}
}

func newFunctionHandleOption() *config.StringOption {
	return &config.StringOption{
		Option: config.Option{
			Key:       "Sf.Publish.Handle",
			Validator: config.Validation.Handle,
		},
	}
}

func newFunctionOemOption(defaultValue string) *config.StringOption {
	return &config.StringOption{
		Option: config.Option{
			Key:          "Sf.Publish.Oem",
			DefaultValue: defaultValue,
			Validator:    config.Validation.Oem,
		},
	}
}

func newFunctionTypeOption() *config.StringOption {
	return &config.StringOption{
		Option: config.Option{
			Key:       "Sf.Publish.Type",
			Validator: config.AnyOfEnumerated(layout.FunctionTypes),
		},
	}
}

func newFunctionVersionOption(defaultValue string) *config.StringOption {
	return &config.StringOption{
		Option: config.Option{
			Key:          "Sf.Publish.Version",
			DefaultValue: defaultValue,
			Validator:    config.Validation.Version,
		},
	}
}
