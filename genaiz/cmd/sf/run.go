package sf

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/cmd/dk"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/lang"
	"genaiz.com/genaiz/schema"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
	"genaiz.com/genaiz/task/docker"
	"genaiz.com/genaiz/task/layout"
	"genaiz.com/genaiz/task/shared"
)

type CollectTaskFactory func(broker.DataLinkWriter) *task.Task[broker.DataLinkParams]
type ExportTaskFactory func(broker.DataLinkWriter) *task.Task[broker.DataLinkParams]
type RunTaskFactory func() *task.Task[docker.ContainerParams]

type SyncExecutor struct {
	innerSources *config.ListOption
	innerStores  *config.ListOption

	collectTaskFactory     CollectTaskFactory
	exportTaskFactory      ExportTaskFactory
	dataLinksWriterFactory dk.DataLinksWriterFactory
}

func (se SyncExecutor) GetFunctionDataLinks(ledger *config.Ledger) []string {
	var dataLink []string

	dataLink = append(dataLink, ledger.GetList(se.innerSources)...)
	dataLink = append(dataLink, ledger.GetList(se.innerStores)...)
	return dataLink
}

func (se SyncExecutor) newBrokerParams(ledger *config.Ledger) *broker.Broker {
	return &broker.Broker{
		AuthFile: ledger.AuthFile,
	}
}

func (se SyncExecutor) newConfigParams(ledger *config.Ledger) *shared.ConfigParams {
	return &shared.ConfigParams{
		ConfigName:   ledger.ConfigName,
		ConfigFolder: ledger.UserPath,
	}
}

type RunExecutor struct {
	BaseExecutor
	SyncExecutor
	*RunOptions

	buildTaskFactory BuildTaskFactory
	runTaskFactory   RunTaskFactory
}

func (re *RunExecutor) Display() {
	displayRunOptions(re.BaseExecutor, re.RunOptions)
}

func (re *RunExecutor) Pretend() {
	var runParams *docker.ContainerParams
	var err error

	if runParams, err = newRunParams(re.BaseExecutor, re.RunOptions); err == nil {
		var plan = task.NewPlan("Run", re.Ledger.Logger)
		var datalinkWorkers []task.Worker
		var workers []task.Worker

		re.Ledger.DisplayChangeDir()

		if re.rebuildImage {
			workers = append(workers, task.NewPretender(makeBuildParams(&re.BaseExecutor), re.buildTaskFactory()))
		}

		if datalinkWorkers, err = makeSyncPretenders(re.Ledger, re.SyncExecutor, re.optionNoPropSync); err == nil {
			workers = append(workers, datalinkWorkers...)
			workers = append(workers, task.NewPretender(runParams, re.runTaskFactory()))
			plan.Sequence(workers...)
			return
		}
	}

	lang.HandleExit(err)
}

func (re *RunExecutor) Proceed() {
	var runParams *docker.ContainerParams
	var err error

	if runParams, err = newRunParams(re.BaseExecutor, re.RunOptions); err == nil {
		var plan = task.NewPlan("Run", re.Ledger.Logger)
		var datalinkWorkers []task.Worker
		var workers []task.Worker

		if re.rebuildImage {
			workers = append(workers, task.NewWorker(makeBuildParams(&re.BaseExecutor), re.buildTaskFactory()))
		}

		if datalinkWorkers, err = makeSyncWorkers(re.Ledger, re.SyncExecutor, re.optionNoPropSync); err == nil {
			workers = append(workers, datalinkWorkers...)
			workers = append(workers, task.NewWorker(runParams, re.runTaskFactory()))
			plan.PrintReportsOnly = true
			plan.Sequence(workers...)
			return
		}
	}

	lang.HandleExit(err)
}

type RunOptions struct {
	EnvOptions
	optionMountInput  *config.StringOption
	optionMountLog    *config.StringOption
	optionMountOutput *config.StringOption
	optionMountVar    *config.StringOption
	optionRunImage    *config.StringOption
	optionRunPrefix   *config.StringOption
	optionNoPropSync  *config.BoolOption
	rebuildImage      bool
}

func (ro *RunOptions) allDefiners() []config.Definer {
	return []config.Definer{
		ro.optionEnvFile,
		ro.optionEnvVars,
		ro.optionMountInput,
		ro.optionMountLog,
		ro.optionMountOutput,
		ro.optionMountVar,
		ro.optionRunImage,
		ro.optionRunPrefix,
	}
}

func NewRun(ledger *config.Ledger, sfCli *Cli) *cobra.Command {
	var options = NewRunOptions(sfCli)
	var run = &cobra.Command{
		Use:     "run",
		Short:   "Runs a Smart Function detached from the current shell",
		Long:    "Runs a Smart Function image detached, building it first if necessary, assigning it a disposable container",
		Example: "genaiz sf run --image=genaiz.com/sf/smartfunc:latest",
		PreRun: func(cmd *cobra.Command, args []string) {
			ledger.FromWorkDir(options.optionEnvFile, cmd.Flags())
			ledger.FromWorkDir(options.optionMountInput, cmd.Flags())
			ledger.FromWorkDir(options.optionMountLog, cmd.Flags())
			ledger.FromWorkDir(options.optionMountOutput, cmd.Flags())
			ledger.FromWorkDir(options.optionMountVar, cmd.Flags())
		},
		Run: func(cmd *cobra.Command, args []string) {
			var imageFlag = cmd.Flags().Lookup(options.optionRunImage.Param)

			options.rebuildImage = imageFlag.Value.String() == ""
			sfCli.Exec(ledger, NewRunExecutor(cmd.Context(), ledger, sfCli, options))
		},
	}

	ledger.Register(run, options.allDefiners()...)
	return run
}

func NewRunExecutor(ctx context.Context, ledger *config.Ledger, sfCli *Cli, options *RunOptions) *RunExecutor {
	return &RunExecutor{
		BaseExecutor: BaseExecutor{
			Cli:     sfCli,
			Context: ctx,
			Ledger:  ledger,
		},
		SyncExecutor: makeSyncExecutor(),
		RunOptions:   options,

		buildTaskFactory: docker.NewBuildTask,
		runTaskFactory:   docker.NewRunTask,
	}
}

func NewRunOptions(sfCli *Cli) *RunOptions {
	var runLayout = layout.NewRunLayout()
	var mountOutputOption = cli.Options.Functions.MountOutput().
		WithKeys(&schema.Genaiz.Function.Run.MountOutput).
		WithDefaultValue(runLayout.DirOutput).
		Optional(false).
		BuildStringOption()

	return &RunOptions{
		EnvOptions: EnvOptions{
			optionEnvFile: cli.Options.Docker.EnvFile().
				WithKeys(&schema.Genaiz.Function.Run.EnvFile).
				BuildStringOption(),
			optionEnvVars: cli.Options.Docker.EnvVar().
				WithKeys(&schema.Genaiz.Function.Run.EnvVars).
				BuildListOption(),
		},
		optionMountInput: cli.Options.Functions.MountInput().
			WithKeys(&schema.Genaiz.Function.Run.MountInput).
			WithDefaultValue(runLayout.DirInput).
			Optional(false).
			BuildStringOption(),
		optionMountLog: cli.Options.Functions.MountLog().
			WithKeys(&schema.Genaiz.Function.Run.MountLog).
			WithDefaultValue(runLayout.DirLog).
			BuildStringOption(),
		optionMountOutput: mountOutputOption,
		optionMountVar: cli.Options.Functions.MountVar().
			WithKeys(&schema.Genaiz.Function.Run.MountVar).
			WithDefaultValue(runLayout.DirVar).
			BuildStringOption(),
		optionNoPropSync: cli.Options.Functions.NoPropSync().
			WithKeys(&schema.Genaiz.Function.Run.NoPropSync).
			BuildBoolOption(),
		optionRunImage: cli.Options.Docker.Image().
			WithKeys(&schema.Genaiz.Function.Run.Image).
			WithDefaultGetter(sfCli.DefaultRunImage).
			BuildStringOption(),
		optionRunPrefix: cli.Options.Docker.ContainerPrefix().
			WithKeys(&schema.Genaiz.Function.Run.Prefix).
			WithDefaultGetter(sfCli.ContainerPrefix).
			BuildStringOption(),
	}
}

func displayRunOptions(be BaseExecutor, ro *RunOptions) {
	var options = []*config.Option{
		&ro.optionRunImage.Option,
		&ro.optionRunPrefix.Option,
		&ro.optionEnvFile.Option,
		&ro.optionEnvVars.Option,
		&ro.optionMountInput.Option,
		&ro.optionMountLog.Option,
		&ro.optionMountOutput.Option,
		&ro.optionMountVar.Option,
	}

	options = append(options, be.Cli.SfOptions()...)
	be.Ledger.DisplayOptions(options...)
}

func makeSyncExecutor() SyncExecutor {
	return SyncExecutor{
		innerSources: cli.NewOptionBuilder().
			WithKeys(&schema.Genaiz.Function.Publish.DataSources).
			BuildListOption(),
		innerStores: cli.NewOptionBuilder().
			WithKeys(&schema.Genaiz.Function.Publish.DataStores).
			BuildListOption(),

		collectTaskFactory:     broker.NewDataLinkCollectTask,
		exportTaskFactory:      broker.NewDataLinkExportTask,
		dataLinksWriterFactory: dk.NewDataLinksWriter,
	}
}

func makeSyncPretenders(ledger *config.Ledger, se SyncExecutor, noSyncOption *config.BoolOption) ([]task.Worker, error) {
	var datalinks = se.GetFunctionDataLinks(ledger)
	var worker []task.Worker
	var err error

	if len(datalinks) > 0 {
		var workerFactory func(*broker.DataLinkParams) task.Worker
		var brokerParams = se.newBrokerParams(ledger)
		var configParams = se.newConfigParams(ledger)

		if ledger.GetBool(noSyncOption) {
			workerFactory = func(params *broker.DataLinkParams) task.Worker {
				var dataLinkWriter = se.dataLinksWriterFactory(ledger, params.GetConfigFile())

				return task.NewPretender(params, se.collectTaskFactory(dataLinkWriter))
			}
		} else {
			workerFactory = func(params *broker.DataLinkParams) task.Worker {
				var dataLinkWriter = se.dataLinksWriterFactory(ledger, params.GetConfigFile())

				return task.NewPretender(params, se.exportTaskFactory(dataLinkWriter))
			}
		}

		return makeDataLinkWorkers(brokerParams, configParams, datalinks, workerFactory)
	}

	return worker, err
}

func makeSyncWorkers(ledger *config.Ledger, se SyncExecutor, noSyncOption *config.BoolOption) ([]task.Worker, error) {
	var datalinks = se.GetFunctionDataLinks(ledger)
	var worker []task.Worker
	var err error

	if len(datalinks) > 0 {
		var workerFactory func(*broker.DataLinkParams) task.Worker
		var brokerParams = se.newBrokerParams(ledger)
		var configParams = se.newConfigParams(ledger)

		if ledger.GetBool(noSyncOption) {
			workerFactory = func(params *broker.DataLinkParams) task.Worker {
				var dataLinkWriter = se.dataLinksWriterFactory(ledger, params.GetConfigFile())

				return task.NewWorker(params, se.collectTaskFactory(dataLinkWriter))
			}
		} else {
			workerFactory = func(params *broker.DataLinkParams) task.Worker {
				var dataLinkWriter = se.dataLinksWriterFactory(ledger, params.GetConfigFile())

				return task.NewWorker(params, se.exportTaskFactory(dataLinkWriter))
			}
		}

		return makeDataLinkWorkers(brokerParams, configParams, datalinks, workerFactory)
	}

	return worker, err
}

func newDataLinkParams(oem, handle, version string, brokerParams *broker.Broker, configParams *shared.ConfigParams) *broker.DataLinkParams {
	return &broker.DataLinkParams{
		Broker:       *brokerParams,
		ConfigParams: *configParams,
		DataLink: &broker.DataLink{
			Oem:     oem,
			Handle:  handle,
			Version: version,
		},
	}
}

func makeDataLinkWorkers(brokerParams *broker.Broker, configParams *shared.ConfigParams, dataLinks []string, workFactory func(*broker.DataLinkParams) task.Worker) ([]task.Worker, error) {
	var result []task.Worker

	for _, link := range dataLinks {
		if o, h, v := dk.ParseDataLinkArgument(link); o != "" && h != "" && v != "" {
			var params = newDataLinkParams(o, h, v, brokerParams, configParams)

			result = append(result, workFactory(params))
		} else {
			return nil, fmt.Errorf("invalid datalink found [%s]", link)
		}
	}

	return result, nil
}

func newRunParams(be BaseExecutor, ro *RunOptions) (*docker.ContainerParams, error) {
	var envVars map[string]string
	var err error

	if envVars, err = ro.makeEnvMap(be.Ledger); err == nil {
		var propSpecs []broker.PropSpec
		var varSpecs []shared.VarSpec

		if err = be.Ledger.Unmarshal(schema.Genaiz.Function.Publish.PropSpecs, &propSpecs); err == nil {
			for _, propSpec := range propSpecs {
				varSpecs = append(varSpecs, propSpec.VarSpec())
			}
		}

		return &docker.ContainerParams{
			RunParams: docker.RunParams{
				Env:      task.Env{Context: be.Context},
				Attached: false,
				Dispose:  true,
			},
			DockerImage: be.Ledger.GetString(ro.optionRunImage),
			EnvVars:     envVars,
			MountInput:  be.Ledger.GetPath(ro.optionMountInput),
			MountLog:    be.Ledger.GetPath(ro.optionMountLog),
			MountOutput: be.Ledger.GetPath(ro.optionMountOutput),
			MountVar:    be.Ledger.GetPath(ro.optionMountVar),
			Prefix:      be.Ledger.GetString(ro.optionRunPrefix),
			VarSpecs:    varSpecs,
		}, nil
	}

	return nil, err
}
