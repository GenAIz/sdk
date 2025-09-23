package sf

import (
	"context"

	"github.com/spf13/cobra"

	"genaiz.com/genaiz-lib/lang/dirz"
	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/schema"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
	"genaiz.com/genaiz/task/docker"
	"genaiz.com/genaiz/task/layout"
	"genaiz.com/genaiz/task/shared"
)

type InspectTaskFactory func() *task.Task[docker.BuildParams]

type ProvisionTaskFactory func() *task.Task[broker.ProvisionParams]

type PublishTaskFactory func() *task.Task[broker.PublishParams]

type PushTaskFactory func() *task.Task[docker.PushParams]

type PublishExecutor struct {
	BaseExecutor
	*PublishOptions

	brokerAddr           string
	buildTaskFactory     BuildTaskFactory
	initTaskFactory      InitTaskFactory
	inspectTaskFactory   InspectTaskFactory
	provisionTaskFactory ProvisionTaskFactory
	publishTaskFactory   PublishTaskFactory
	pushTaskFactory      PushTaskFactory
}

func (pe *PublishExecutor) Display() {
	pe.Ledger.DisplayOptionsWithMap(&map[string]string{
		"broker": pe.brokerAddr,
	},
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
		var builder = makeInitBuilder(pe.Cli)
		var initParams = pe.makePublishInitParams()

		workers = append(workers, task.NewPretender(initParams, pe.initTaskFactory(builder)))
	}

	plan.ContinueOnFailure = true
	plan.Sequence(workers...)
}

func (pe *PublishExecutor) Proceed() {
	var rebuild = pe.Ledger.GetBool(pe.optionRebuild)
	var noUpdate = pe.Ledger.GetBool(pe.optionNoUpdate)
	var buildParams = makeBuildParams(&pe.BaseExecutor)
	var provisionParams = pe.makeProvisionParams()
	var publishParams = pe.makePublishParams(provisionParams)
	var pushParams = pe.makePushParams()
	var workers []task.Worker
	var plan = task.NewPlan("Publish", pe.Ledger.Logger)

	if rebuild {
		workers = append(workers, task.NewWorker(buildParams, pe.buildTaskFactory()))
	}

	workers = append(workers,
		task.NewWorker(buildParams, pe.inspectTaskFactory()),
		task.NewWorker(provisionParams, pe.provisionTaskFactory()),
		task.NewWorker(pushParams, pe.pushTaskFactory()),
		task.NewWorker(publishParams, pe.publishTaskFactory()),
	)

	if !noUpdate {
		var builder = makeInitBuilder(pe.Cli)
		var initParams = pe.makePublishInitParams()

		workers = append(workers, task.NewWorker(initParams, pe.initTaskFactory(builder)))
	}

	plan.Sequence(workers...)
}

func (pe *PublishExecutor) makeProvisionParams() *broker.ProvisionParams {
	var nameDesc = pe.Ledger.GetString(pe.optionName)

	return &broker.ProvisionParams{
		Broker: broker.Broker{
			AuthFile: pe.Ledger.AuthFile,
			HostAddr: pe.brokerAddr,
		},
		Arches:      pe.Ledger.GetList(pe.optionArches),
		Description: nameDesc,
		Handle:      pe.Ledger.GetString(pe.optionHandle),
		Name:        nameDesc,
		Oem:         pe.Ledger.GetString(pe.optionOem),
		Type:        pe.Ledger.GetString(pe.optionType),
		Version:     pe.Ledger.GetString(pe.optionVersion),
	}
}

func (pe *PublishExecutor) makePublishInitParams() *layout.InitParams {
	var archTypeStrings = pe.Ledger.GetList(pe.optionArches)
	var functionTypeString = pe.Ledger.GetString(pe.optionType)
	var archTypes []layout.ArchType
	var functionType *layout.FunctionType
	var err error

	functionType, err = layout.FunctionTypes.FromString(functionTypeString)
	cobra.CheckErr(err)

	if len(archTypeStrings) > 0 {
		archTypes, err = layout.ArchTypes.AllFromStrings(&archTypeStrings)
		cobra.CheckErr(err)
	}

	return &layout.InitParams{
		CreateParams: layout.CreateParams{
			ConfigParams: shared.ConfigParams{
				ConfigName: pe.Ledger.ConfigName,
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
			HostAddr: pe.brokerAddr,
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
	optionArches   *config.ListOption
	optionHandle   *config.StringOption
	optionName     *config.StringOption
	optionRebuild  *config.BoolOption
	optionNoUpdate *config.BoolOption
	optionOem      *config.StringOption
	optionType     *config.StringOption
	optionVersion  *config.StringOption
}

func (po *PublishOptions) allDefiners() []config.Definer {
	return []config.Definer{
		po.optionArches,
		po.optionHandle,
		po.optionName,
		po.optionRebuild,
		po.optionNoUpdate,
		po.optionOem,
		po.optionType,
		po.optionVersion,
	}
}

func NewPublish(ledger *config.Ledger, cli *Cli) *cobra.Command {
	var options = NewPublishOptions(cli)
	var publish = &cobra.Command{
		Use:     "publish HOST",
		Short:   "Publishes a Smart Function metadata and image to a Broker",
		Long:    "Publishes a Smart Function metadata and image to a Broker, by provisioning its information and then pushing it the Broker's registry",
		Example: "genaiz sf publish --handle=my-function --oem=com.genaiz --version=0.1-dev www.genaiz.com",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cli.Exec(ledger, NewPublishExecutor(cmd.Context(), ledger, cli, options, args[0]))
		},
	}

	ledger.Register(publish, options.allDefiners()...)
	return publish
}

func NewPublishExecutor(ctx context.Context, ledger *config.Ledger, cli *Cli, options *PublishOptions, brokerAddr string) *PublishExecutor {
	return &PublishExecutor{
		BaseExecutor: BaseExecutor{
			Cli:     cli,
			Context: ctx,
			Ledger:  ledger,
		},
		PublishOptions: options,

		brokerAddr:           brokerAddr,
		buildTaskFactory:     docker.NewBuildTask,
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
		optionArches: cli.Options.Functions.Arches().
			WithKeys(&schema.Genaiz.Function.Publish.Arches).
			BuildListOption(),
		optionHandle: handleOpt,
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
