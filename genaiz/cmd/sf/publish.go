package sf

import (
	"context"

	"github.com/spf13/cobra"

	"genaiz.com/genaiz-lib/lang/dirz"
	"genaiz.com/genaiz-lib/lang/panicz"
	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/lang"
	"genaiz.com/genaiz/schema"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
	"genaiz.com/genaiz/task/docker"
	"genaiz.com/genaiz/task/layout"
	"genaiz.com/genaiz/task/shared"
)

type GetTaskFactory func() *task.Task[broker.GetParams]

type IdTaskFactory func() *task.Task[broker.GetParams]

type InspectTaskFactory func() *task.Task[docker.BuildParams]

type ProvisionTaskFactory func() *task.Task[broker.ProvisionParams]

type PublishTaskFactory func() *task.Task[broker.PublishParams]

type PushTaskFactory func() *task.Task[docker.PushParams]

type PublishExecutor struct {
	BaseExecutor
	*PublishOptions

	innerDataSources     *config.ListOption
	innerDataStores      *config.ListOption
	innerExtras          *config.Option
	innerInputPorts      *config.Option
	innerOutboundProxies *config.Option
	innerOutputPorts     *config.Option
	innerPropSpecs       *config.Option
	innerResultValues    *config.ListOption

	printerParams cli.PrinterParametric

	buildTaskFactory     BuildTaskFactory
	getTaskFactory       GetTaskFactory
	idTaskFactory        IdTaskFactory
	initTaskFactory      InitTaskFactory
	inspectTaskFactory   InspectTaskFactory
	provisionTaskFactory ProvisionTaskFactory
	publishTaskFactory   PublishTaskFactory
	pushTaskFactory      PushTaskFactory
}

func (pe *PublishExecutor) Display() {
	pe.Ledger.DisplayOptions(
		&pe.optionAccount.Option,
		&pe.optionArches.Option,
		&pe.optionHandle.Option,
		&pe.optionName.Option,
		&pe.optionRebuild.Option,
		&pe.optionNoUpdate.Option,
		&pe.optionOem.Option,
		&pe.optionType.Option,
		&pe.optionVersion.Option,
	)
}

func (pe *PublishExecutor) Pretend() {
	var buildParams = makeBuildParams(&pe.BaseExecutor)
	var provisionParams = pe.makeProvisionParams()
	var publishParams = pe.makePublishParams(provisionParams)
	var pushParams = pe.makePushParams()
	var rebuild = pe.Ledger.GetBool(pe.optionRebuild)
	var noUpdate = pe.Ledger.GetBool(pe.optionNoUpdate)
	var plan = task.NewPlan("Publish", pe.Ledger.Logger)
	var workers []task.Worker

	if rebuild {
		workers = append(workers, task.NewPretender(buildParams, pe.buildTaskFactory()))
	}

	workers = append(workers, task.NewPretender(buildParams, pe.inspectTaskFactory()))
	workers = append(workers, task.NewPretender(provisionParams, pe.provisionTaskFactory()))
	workers = append(workers, task.NewPretender(pushParams, pe.pushTaskFactory()))
	workers = append(workers, task.NewPretender(publishParams, pe.publishTaskFactory()))

	if !noUpdate {
		var writer = newInitWriter(pe.Cli)
		var initParams = pe.makePublishInitParams()

		workers = append(workers, task.NewPretender(initParams, pe.initTaskFactory(writer)))
	}

	plan.ContinueOnFailure = true
	plan.Sequence(workers...)
}

func (pe *PublishExecutor) Proceed() {
	var rebuild = pe.Ledger.GetBool(pe.optionRebuild)
	var noUpdate = pe.Ledger.GetBool(pe.optionNoUpdate)
	var buildParams = makeBuildParams(&pe.BaseExecutor)
	var provisionParams = pe.makeProvisionParams()
	var getParams = provisionParams.ToGetParams()
	var publishParams = pe.makePublishParams(provisionParams)
	var pushParams = pe.makePushParams()
	var plan = task.NewPlan("Publish", pe.Ledger.Logger)
	var workers []task.Worker
	var err error

	if rebuild {
		workers = append(workers, task.NewWorker(buildParams, pe.buildTaskFactory()))
	}

	workers = append(workers,
		task.NewWorker(buildParams, pe.inspectTaskFactory()),
		task.NewWorker(provisionParams, pe.provisionTaskFactory()),
		task.NewWorker(pushParams, pe.pushTaskFactory()),
		task.NewWorker(publishParams, pe.publishTaskFactory()),
		task.NewWorker(getParams, pe.idTaskFactory()),
	)

	if !noUpdate {
		var writer = newInitWriter(pe.Cli)
		var initParams = pe.makePublishInitParams()

		workers = append(workers, task.NewWorker(initParams, pe.initTaskFactory(writer)))
	}

	workers = append(workers, task.NewWorker(getParams, pe.getTaskFactory()))

	if pe.Ledger.GetBool(pe.optionJsonPrinter) {
		var printer = pe.printerParams.Printer()
		var function *broker.Function
		var failure interface{}

		plan.OnReturn = func(i interface{}) { function = i.(*broker.Function) }
		plan.OnFailure = func(i interface{}) { failure = i }
		plan.OnSuccess = nil
		plan.Sequence(workers...)

		if failure == nil {
			err = printer.Print(function)
		} else {
			err = printer.Error(failure)
		}
	} else {
		plan.PrintReportsOnly = true
		plan.Sequence(workers...)
	}

	lang.HandleExit(err)
}

func (pe *PublishExecutor) makeProvisionExtras() map[string]any {
	var raw = pe.Ledger.Get(pe.innerExtras)

	if result, ok := raw.(map[string]any); ok {
		return result
	}

	return make(map[string]any)
}

func (pe *PublishExecutor) makeProvisionParams() *broker.ProvisionParams {
	var nameDesc = pe.Ledger.GetString(pe.optionName)
	var inputPorts = broker.ListDataPorts(pe.Ledger.Get(pe.innerInputPorts))
	var outboundProxies = broker.ListProxies(pe.Ledger.Get(pe.innerOutboundProxies))
	var outputPorts = broker.ListDataPorts(pe.Ledger.Get(pe.innerOutputPorts))
	var propSpecs = broker.ListPropSpecs(pe.Ledger.Get(pe.innerPropSpecs))
	var extraMap = pe.makeProvisionExtras()

	return &broker.ProvisionParams{
		GetParams: broker.GetParams{
			Oem:     pe.Ledger.GetString(pe.optionOem),
			Handle:  pe.Ledger.GetString(pe.optionHandle),
			Version: pe.Ledger.GetString(pe.optionVersion),
		},
		Broker: broker.Broker{
			AuthFile: pe.Ledger.AuthFile,
			HostAddr: pe.Ledger.GetString(pe.optionAccount),
		},
		Arches:          pe.Ledger.GetList(pe.optionArches),
		Extras:          extraMap,
		DataSources:     pe.Ledger.GetList(pe.innerDataSources),
		DataStores:      pe.Ledger.GetList(pe.innerDataStores),
		Description:     nameDesc,
		InputPorts:      inputPorts,
		Name:            nameDesc,
		OutboundProxies: outboundProxies,
		OutputPorts:     outputPorts,
		PropSpecs:       propSpecs,
		ResultValues:    pe.Ledger.GetList(pe.innerResultValues),
		Type:            pe.Ledger.GetString(pe.optionType),
	}
}

func (pe *PublishExecutor) makePublishInitParams() *layout.InitParams {
	var functionTypeString = pe.Ledger.GetString(pe.optionType)
	var functionType, err = shared.FunctionTypes.FromString(functionTypeString)
	var archTypeStrings = pe.Ledger.GetList(pe.optionArches)
	var archTypes []shared.ArchType

	// This should never happen, unless there's a bug with optionType
	panicz.PanicIfError(err)

	if len(archTypeStrings) > 0 {
		archTypes, err = shared.ArchTypes.AllFromStrings(&archTypeStrings)
		// This should never happen, unless there's a bug with optionArches
		panicz.PanicIfError(err)
	}

	return &layout.InitParams{
		CreateParams: layout.CreateParams{
			ConfigParams: shared.ConfigParams{
				ConfigName:   pe.Ledger.ConfigName,
				ConfigFolder: pe.Ledger.WorkDir,
			},
		},
		Arches:  archTypes,
		Type:    *functionType,
		Handle:  pe.Ledger.GetString(pe.optionHandle),
		Name:    pe.Ledger.GetString(pe.optionName),
		OEM:     pe.Ledger.GetString(pe.optionOem),
		Version: pe.Ledger.GetString(pe.optionVersion),
	}
}

func (pe *PublishExecutor) makePublishParams(provisionParams *broker.ProvisionParams) *broker.PublishParams {
	return &broker.PublishParams{
		Broker: broker.Broker{
			AuthFile: pe.Ledger.AuthFile,
			HostAddr: pe.Ledger.GetString(pe.optionAccount),
		},
		Handle: provisionParams.Handle,
		Oem:    provisionParams.Oem,
	}
}

func (pe *PublishExecutor) makePushParams() *docker.PushParams {
	return &docker.PushParams{
		Env: task.Env{
			Context: pe.Context,
		},
	}
}

type PublishOptions struct {
	optionAccount     *config.StringOption
	optionArches      *config.ListOption
	optionHandle      *config.StringOption
	optionJsonPrinter *config.BoolOption
	optionName        *config.StringOption
	optionNoUpdate    *config.BoolOption
	optionOem         *config.StringOption
	optionRebuild     *config.BoolOption
	optionType        *config.StringOption
	optionVersion     *config.StringOption
}

func (po *PublishOptions) allDefiners() []config.Definer {
	return []config.Definer{
		po.optionAccount,
		po.optionArches,
		po.optionHandle,
		po.optionJsonPrinter,
		po.optionName,
		po.optionNoUpdate,
		po.optionOem,
		po.optionRebuild,
		po.optionType,
		po.optionVersion,
	}
}

func NewPublish(ledger *config.Ledger, sfCli *Cli) *cobra.Command {
	var options = NewPublishOptions(sfCli)
	var publishCmd = &cobra.Command{
		Use:     "publish",
		Short:   "Publishes a Smart Function metadata and image to a Broker",
		Long:    "Publishes a Smart Function metadata and image to a Broker, by provisioning its information and then pushing it the Broker's registry",
		Example: "genaiz sf publish --account=www.genaiz.com --handle=my-function --oem=com.genaiz --version=0.1-dev",
		Args:    cobra.MaximumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			sfCli.Exec(ledger, NewPublishExecutor(cmd.Context(), ledger, sfCli, options))
		},
	}

	ledger.Register(publishCmd, options.allDefiners()...)
	cli.AutoBridge.Accounts().Option(publishCmd, ledger, options.optionAccount)
	return publishCmd
}

func NewPublishExecutor(ctx context.Context, ledger *config.Ledger, sfCli *Cli, options *PublishOptions) *PublishExecutor {
	return &PublishExecutor{
		BaseExecutor: BaseExecutor{
			Cli:     sfCli,
			Context: ctx,
			Ledger:  ledger,
		},
		PublishOptions: options,

		innerDataSources: cli.NewOptionBuilder().
			WithKeys(&schema.Genaiz.Function.Publish.DataSources).
			BuildListOption(),
		innerDataStores: cli.NewOptionBuilder().
			WithKeys(&schema.Genaiz.Function.Publish.DataStores).
			BuildListOption(),
		innerExtras: cli.NewOptionBuilder().
			WithKeys(&schema.Genaiz.Function.Publish.Extras).
			BuildOption(),
		innerInputPorts: cli.NewOptionBuilder().
			WithKeys(&schema.Genaiz.Function.Publish.InputPorts).
			BuildOption(),
		innerOutboundProxies: cli.NewOptionBuilder().
			WithKeys(&schema.Genaiz.Function.Publish.OutboundProxies).
			BuildOption(),
		innerOutputPorts: cli.NewOptionBuilder().
			WithKeys(&schema.Genaiz.Function.Publish.OutputPorts).
			BuildOption(),
		innerPropSpecs: cli.NewOptionBuilder().
			WithKeys(&schema.Genaiz.Function.Publish.PropSpecs).
			BuildOption(),
		innerResultValues: cli.NewOptionBuilder().
			WithKeys(&schema.Genaiz.Function.Publish.ResultValues).
			BuildListOption(),

		printerParams: cli.NewPrinterParam(ledger, options.optionJsonPrinter),

		buildTaskFactory:     docker.NewBuildTask,
		getTaskFactory:       broker.NewFunctionGetTask,
		idTaskFactory:        broker.NewFunctionIdTask,
		initTaskFactory:      layout.NewInitTask,
		inspectTaskFactory:   docker.NewInspectTask,
		provisionTaskFactory: broker.NewFunctionProvisionTask,
		pushTaskFactory:      docker.NewPushTask,
		publishTaskFactory:   broker.NewFunctionPublishTask,
	}
}

func NewPublishOptions(sfCli *Cli) *PublishOptions {
	var parentOpt = cli.Options.Configs.SolutionPath().
		WithKeys(&schema.Genaiz.Function.Init.SolutionPath).
		WithDefaultGetter(func(ledger *config.Ledger) any {
			return dirz.WorkingDirParent()
		}).BuildStringOption()
	var handleOpt = cli.Options.Functions.Handle().
		WithKeys(&schema.Genaiz.Function.Publish.Handle).
		WithDefaultGetter(func(ledger *config.Ledger) any {
			return dirz.WorkingDirBase()
		}).
		BuildStringOption()

	return &PublishOptions{
		optionAccount: cli.Options.Functions.Account().
			WithKeys(&schema.Genaiz.Function.Publish.Account).
			BuildStringOption(),
		optionArches: cli.Options.Functions.Arches().
			WithKeys(&schema.Genaiz.Function.Publish.Arches).
			BuildListOption(),
		optionHandle: handleOpt,
		optionJsonPrinter: cli.Options.Printer.JsonPrinter().
			WithKeys(&schema.Genaiz.Function.Publish.Printer).
			BuildBoolOption(),
		optionName: cli.Options.Functions.Name().
			WithKeys(&schema.Genaiz.Function.Publish.Name).
			WithUsage("defaults to the handle value if not provided").
			WithDefaultGetter(func(ledger *config.Ledger) any {
				return ledger.GetString(handleOpt)
			}).BuildStringOption(),
		optionNoUpdate: cli.Options.Configs.NoUpdate().
			WithKeys(&schema.Genaiz.Function.Publish.NoUpdate).
			BuildBoolOption(),
		optionOem: cli.Options.Functions.Oem().
			WithKeys(&schema.Genaiz.Function.Publish.Oem).
			WithDefaultGetter(sfCli.ParentOem(parentOpt)).
			BuildStringOption(),
		optionRebuild: cli.Options.Functions.Rebuild().
			WithKeys(&schema.Genaiz.Function.Publish.Rebuild).
			BuildBoolOption(),
		optionType: cli.Options.Functions.Type().
			WithKeys(&schema.Genaiz.Function.Publish.Type).
			BuildStringOption(),
		optionVersion: cli.Options.Functions.Version().
			WithKeys(&schema.Genaiz.Function.Publish.Version).
			WithDefaultGetter(sfCli.ParentVersion(parentOpt)).
			BuildStringOption(),
	}
}
