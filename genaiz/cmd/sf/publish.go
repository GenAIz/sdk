package sf

import (
	"context"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
	"genaiz.com/genaiz/task/docker"
	"genaiz.com/genaiz/task/layout"
)

type InspectTaskFactory func() *task.Task[docker.BuildParams]

type ProvisionTaskFactory func() *task.Task[broker.ProvisionParams]

type PublishTaskFactory func() *task.Task[broker.ProvisionParams]

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
		&pe.optionFqdn.Option,
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
	workers = append(workers, task.NewPretender(provisionParams, pe.publishTaskFactory()))

	if !noUpdate {
		var builder = makeInitBuilder(pe.Cli)
		var initParams = makePublishInitParams(pe.Ledger, pe.PublishOptions)

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
		task.NewWorker(provisionParams, pe.publishTaskFactory()),
	)

	if !noUpdate {
		var builder = makeInitBuilder(pe.Cli)
		var initParams = makePublishInitParams(pe.Ledger, pe.PublishOptions)

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
		Fqdn:        pe.Ledger.GetString(pe.optionFqdn),
		Handle:      pe.Ledger.GetString(pe.optionHandle),
		Name:        nameDesc,
		Oem:         pe.Ledger.GetString(pe.optionOem),
		Type:        pe.Ledger.GetString(pe.optionType),
		Version:     pe.Ledger.GetString(pe.optionVersion),
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
	optionFqdn     *config.StringOption
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
		po.optionFqdn,
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
		Args:    cobra.MatchAll(cobra.ExactArgs(1)),
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
		provisionTaskFactory: broker.NewProvisionTask,
		pushTaskFactory:      docker.NewPushTask,
		publishTaskFactory:   broker.NewPublishTask,
	}
}

func NewPublishOptions(cli *Cli) *PublishOptions {
	var options = newPublishOptions("Publish", cli)

	options.optionRebuild = newOptionRebuild()
	options.optionNoUpdate = newOptionNoUpdate()
	return options
}

func makePublishInitParams(ledger *config.Ledger, options *PublishOptions) *layout.InitParams {
	var archTypeStrings = ledger.GetList(options.optionArches)
	var functionTypeString = ledger.GetString(options.optionType)
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
			ConfigName: ledger.ConfigName,
		},
		Arches:  archTypes,
		FQDN:    ledger.GetString(options.optionFqdn),
		Type:    *functionType,
		Handle:  ledger.GetString(options.optionHandle),
		Name:    ledger.GetString(options.optionName),
		OEM:     ledger.GetString(options.optionOem),
		Version: ledger.GetString(options.optionVersion),
	}
}

func newOptionArches(cmd string) *config.ListOption {
	return &config.ListOption{
		Option: config.Option{
			Key:       "SF." + cmd + ".Arches",
			Param:     "arch",
			Usage:     "a list of architectures assigned to the function scope of execution. Supported: x86, x86_64, arm and arm64",
			Validator: config.AllFromEnumerated(layout.ArchTypes),
		},
	}
}

func newOptionFqdn(cmd string) *config.StringOption {
	return &config.StringOption{
		Option: config.Option{
			Key:       "SF." + cmd + ".FQDN",
			Param:     "fqdn",
			Usage:     "fully qualified domain name to which the function belongs",
			Validator: config.Validation.DomainName,
		},
	}
}

func newOptionHandle(cmd string) *config.StringOption {
	return &config.StringOption{
		Option: config.Option{
			Key:   "SF." + cmd + ".Handle",
			Param: "handle",
			Usage: "a unique string identifying the function across its oem, valid characters can only be alphanumerics, the dot, dash and underscore",
			DefaultSetter: func(ledger *config.Ledger) any {
				var wd, _ = os.Getwd()

				return filepath.Base(wd)
			},
			Validator: config.Validation.FolderName,
		},
		// TODO: Validator, this needs to be non-empty with certain max length
	}
}

func newOptionName(defaultOption *config.StringOption, cmd string) *config.StringOption {
	return &config.StringOption{
		Option: config.Option{
			Key:   "SF." + cmd + ".Name",
			Param: "name",
			Short: "n",
			Usage: "display name of the function to create, defaults to the handle value if not provided",
			DefaultGetter: func(ledger *config.Ledger) any {
				return ledger.GetString(defaultOption)
			},
		},
		// TODO: Validator, this needs to be non-empty with certain max length
	}
}

func newOptionNoUpdate() *config.BoolOption {
	return &config.BoolOption{
		Option: config.Option{
			Key:          "SF.Publish.No_Update",
			Param:        "noUpdate",
			DefaultValue: "false",
			Usage:        "if set publish will not update configuration values after publishing",
		},
	}
}

func newOptionOem(cmd string) *config.StringOption {
	return &config.StringOption{
		Option: config.Option{
			Key:   "SF." + cmd + ".OEM",
			Param: "oem",
			Usage: "unique id belonging to the publisher of the function",
		},
		// TODO: Validator, this needs to be non-empty with certain max length
	}
}

func newOptionRebuild() *config.BoolOption {
	return &config.BoolOption{
		Option: config.Option{
			Key:          "SF.Publish.Rebuild",
			Param:        "rebuild",
			DefaultValue: "false",
			Usage:        "if set publish will force building the Smart Function before provisioning",
		},
	}
}

func newOptionType(cmd string) *config.StringOption {
	return &config.StringOption{
		Option: config.Option{
			Key:          "SF." + cmd + ".Type",
			Param:        "type",
			Short:        "t",
			DefaultValue: "function",
			Usage:        "type of the function to create, only \"connector\", \"function\" or \"trigger\" are supported",
			Validator:    config.AnyOfEnumerated(layout.FunctionTypes),
		},
	}
}

func newOptionVersion(cmd string, cli *Cli) *config.StringOption {
	return &config.StringOption{
		Option: config.Option{
			Key:   "SF." + cmd + ".Version",
			Param: "version",
			Short: "v",
			Usage: "initial version to use for the smart function",
			DefaultGetter: func(ledger *config.Ledger) any {
				return ledger.GetString(cli.optionDockerVersion)
			},
			Validator: config.Validation.Version,
		},
	}
}

func newPublishOptions(cmd string, cli *Cli, flagOptions ...*config.BoolOption) *PublishOptions {
	var optionHandle = newOptionHandle(cmd)
	var optionRebuild *config.BoolOption
	var optionNoUpdate *config.BoolOption
	var size = len(flagOptions)

	if size > 0 {
		optionRebuild = flagOptions[0]
	}

	if size > 1 {
		optionNoUpdate = flagOptions[1]
	}

	return &PublishOptions{
		optionArches:   newOptionArches(cmd),
		optionFqdn:     newOptionFqdn(cmd),
		optionHandle:   optionHandle,
		optionName:     newOptionName(optionHandle, cmd),
		optionNoUpdate: optionNoUpdate,
		optionRebuild:  optionRebuild,
		optionType:     newOptionType(cmd),
		optionOem:      newOptionOem(cmd),
		optionVersion:  newOptionVersion(cmd, cli),
	}
}
