package sf

import (
	"context"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"genaiz.com/genaiz-lib/lang/dirz"
	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz-lib/lang/panicz"
	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/lang"
	"genaiz.com/genaiz/schema"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
	"genaiz.com/genaiz/task/layout"
	"genaiz.com/genaiz/task/shared"
)

type InitWriter struct {
	*PublishOptions
	*RunOptions
	buildFileKeys        *schema.Keys
	buildTagKeys         *schema.Keys
	buildVersionKeys     *schema.Keys
	publishPropSpecsKeys *schema.Keys
	removedPropSpec      *broker.PropSpec
	vp                   *viper.Viper
}

func (iw *InitWriter) BuildArches() (string, []string) {
	return iw.optionArches.Key, iw.vp.GetStringSlice(iw.optionArches.Key)
}

func (iw *InitWriter) BuildHandle() (string, string) {
	return iw.optionHandle.Key, iw.vp.GetString(iw.optionHandle.Key)
}

func (iw *InitWriter) BuildInput() (string, string) {
	return iw.optionMountInput.Key, iw.vp.GetString(iw.optionMountInput.Key)
}

func (iw *InitWriter) BuildName() (string, string) {
	return iw.optionName.Key, iw.vp.GetString(iw.optionName.Key)
}

func (iw *InitWriter) BuildOem() (string, string) {
	return iw.optionOem.Key, iw.vp.GetString(iw.optionOem.Key)
}

func (iw *InitWriter) BuildOutput() map[string]string {
	return map[string]string{
		iw.optionMountOutput.Key: iw.vp.GetString(iw.optionMountOutput.Key),
		iw.optionMountLog.Key:    iw.vp.GetString(iw.optionMountLog.Key),
		iw.optionMountVar.Key:    iw.vp.GetString(iw.optionMountVar.Key),
	}
}

func (iw *InitWriter) BuildPropSpecs() (string, []broker.PropSpec) {
	if list, ok := iw.vp.Get(iw.publishPropSpecsKeys.Doc).([]broker.PropSpec); ok {
		return iw.publishPropSpecsKeys.Doc, list
	}

	return iw.publishPropSpecsKeys.Doc, []broker.PropSpec{}
}

func (iw *InitWriter) BuildRemovedPropSpec() (string, *broker.PropSpec) {
	return iw.publishPropSpecsKeys.Doc, iw.removedPropSpec
}

func (iw *InitWriter) BuildType() (string, string) {
	return iw.optionType.Key, iw.vp.GetString(iw.optionType.Key)
}

func (iw *InitWriter) BuildVersion() (string, string) {
	return iw.optionVersion.Key, iw.vp.GetString(iw.optionVersion.Key)
}

func (iw *InitWriter) WithArches(values []string) layout.ConfigWriter {
	if values != nil {
		iw.vp.Set(iw.optionArches.Key, values)
	}

	return iw
}

func (iw *InitWriter) WithConfigFile(file string) layout.ConfigWriter {
	if fd, err := os.Open(file); fd != nil {
		defer filez.CloseSilently(fd)
		iw.vp.SetConfigType(filez.GetFileType(file))
		panicz.PanicIfError(iw.vp.ReadConfig(fd))
	} else {
		panicz.PanicIfError(err)
	}

	return iw
}

func (iw *InitWriter) WithDockerFile(file string) layout.ConfigWriter {
	if file != "" {
		iw.vp.Set(iw.buildFileKeys.Doc, file)
	}

	return iw
}

func (iw *InitWriter) WithHandle(value string) layout.ConfigWriter {
	if value != "" {
		var oem = iw.vp.GetString(iw.optionOem.Key)

		iw.setTag(oem, value)
		iw.vp.Set(iw.optionHandle.Key, value)
	}

	return iw
}

func (iw *InitWriter) WithInput(value string) layout.ConfigWriter {
	if value != "" {
		iw.vp.Set(iw.optionMountInput.Key, value)
	}

	return iw
}

func (iw *InitWriter) WithName(value string) layout.ConfigWriter {
	if value != "" {
		iw.vp.Set(iw.optionName.Key, value)
	}
	return iw
}

func (iw *InitWriter) WithOem(value string) layout.ConfigWriter {
	if value != "" {
		var handle = iw.vp.GetString(iw.optionHandle.Key)

		iw.setTag(value, handle)
		iw.vp.Set(iw.optionOem.Key, value)
	}

	return iw
}

func (iw *InitWriter) WithOutput(value string) layout.ConfigWriter {
	if value != "" {
		iw.vp.Set(iw.optionMountOutput.Key, value)
	}

	return iw
}

func (iw *InitWriter) WithPropSpecs(specs []broker.PropSpec) layout.ConfigWriter {
	if len(specs) > 0 || iw.removedPropSpec != nil {
		iw.vp.Set(iw.publishPropSpecsKeys.Doc, specs)
	}

	return iw
}

func (iw *InitWriter) WithPropSpecRemoved(spec *broker.PropSpec) layout.ConfigWriter {
	iw.removedPropSpec = spec
	return iw
}

func (iw *InitWriter) WithTag(value string) layout.ConfigWriter {
	if value != "" {
		iw.vp.Set(iw.buildTagKeys.Doc, value)
	}

	return iw
}

func (iw *InitWriter) WithType(value string) layout.ConfigWriter {
	if value != "" {
		iw.vp.Set(iw.optionType.Key, value)
	}

	return iw
}

func (iw *InitWriter) WithVersion(value string) layout.ConfigWriter {
	if value != "" {
		iw.vp.Set(iw.optionVersion.Key, value)
		iw.vp.Set(iw.buildVersionKeys.Doc, "latest")
	}

	return iw
}

func (iw *InitWriter) Write(filepath string) error {
	return iw.vp.WriteConfigAs(filepath)
}

func (iw *InitWriter) setTag(oem string, handle string) {
	var tagTokens []string

	if oem != "" {
		tagTokens = append(tagTokens, oem)
	}

	if handle != "" {
		tagTokens = append(tagTokens, handle)
	}

	iw.vp.Set(iw.buildTagKeys.Doc, strings.Join(tagTokens, "/"))
}

type InitTaskFactory func(layout.ConfigWriter) *task.Task[layout.InitParams]

type InitExecutor struct {
	BaseExecutor
	*InitOptions

	initTaskFactory InitTaskFactory
}

func (ie *InitExecutor) Display() {
	ie.Ledger.DisplayOptions(
		&ie.optionArches.Option,
		&ie.optionConfigType.Option,
		&ie.optionHandle.Option,
		&ie.optionName.Option,
		&ie.optionType.Option,
		&ie.optionMountInput.Option,
		&ie.optionMountOutput.Option,
		&ie.optionOem.Option,
		&ie.optionVersion.Option,
	)
}

func (ie *InitExecutor) Pretend() {
	var params = ie.makeInitParams()
	var builder = ie.makeInitBuilder()

	ie.Ledger.DisplayChangeDir()
	ie.initTaskFactory(builder).Pretend(params, ie.Ledger.Logger)
}

func (ie *InitExecutor) Proceed() {
	var builder = ie.makeInitBuilder()
	var params = ie.makeInitParams()
	var plan = task.NewPlan("Init", ie.Ledger.Logger)

	plan.PrintReportsOnly = true
	task.Single(plan, params, ie.initTaskFactory(builder))
}

func (ie *InitExecutor) makeInitBuilder() layout.ConfigWriter {
	return makeInitBuilder(ie.Ledger, ie.Cli)
}

func (ie *InitExecutor) makeInitParams() *layout.InitParams {
	return makeInitParams(ie.Ledger, ie.InitOptions)
}

type InitOptions struct {
	optionArches      *config.ListOption
	optionConfigType  *config.StringOption
	optionHandle      *config.StringOption
	optionInteractive *config.BoolOption
	optionMountInput  *config.StringOption
	optionMountOutput *config.StringOption
	optionName        *config.StringOption
	optionOem         *config.StringOption
	optionType        *config.StringOption
	optionVersion     *config.StringOption
}

func (io *InitOptions) allDefiners() []config.Definer {
	return []config.Definer{
		io.optionArches,
		io.optionConfigType,
		io.optionHandle,
		io.optionInteractive,
		io.optionMountInput,
		io.optionMountOutput,
		io.optionName,
		io.optionOem,
		io.optionType,
		io.optionVersion,
	}
}

func NewInit(ledger *config.Ledger, cli *Cli) *cobra.Command {
	var options = NewInitOptions(cli)
	var init = &cobra.Command{
		Use:     "init",
		Short:   "Initiates an existing Smart Function configuration values",
		Long:    "Initiates an existing Smart Function configuration values, interactively by default",
		Example: "genaiz sf init --oem com.genaiz",
		Run: func(cmd *cobra.Command, args []string) {
			cli.Exec(ledger, NewInitExecutor(cmd.Context(), ledger, cli, options))
		},
	}

	ledger.Register(init, options.allDefiners()...)
	return init
}

func NewInitExecutor(ctx context.Context, ledger *config.Ledger, cli *Cli, options *InitOptions) *InitExecutor {
	return &InitExecutor{
		BaseExecutor: BaseExecutor{
			Context: ctx,
			Ledger:  ledger,
			Cli:     cli,
		},
		InitOptions: options,

		initTaskFactory: layout.NewInitTask,
	}
}

func NewInitOptions(sfCli *Cli) *InitOptions {
	var parentOpt = cli.Options.Configs.SolutionPath().
		WithKeys(&schema.Genaiz.Function.Init.SolutionPath).
		WithDefaultGetter(func(ledger *config.Ledger) any {
			return dirz.WorkingDirParent()
		}).BuildStringOption()
	var handleOpt = cli.Options.Functions.Handle().
		WithKeys(&schema.Genaiz.Function.Init.Handle).
		WithDefaultGetter(func(ledger *config.Ledger) any {
			return dirz.WorkingDirBase()
		}).
		BuildStringOption()

	return &InitOptions{
		optionArches: cli.Options.Functions.Arches().
			WithKeys(&schema.Genaiz.Function.Init.Arches).
			BuildListOption(),
		optionConfigType: cli.Options.Configs.Type().
			WithKeys(&schema.Genaiz.Function.Init.ConfigType).
			WithDefaultGetter(sfCli.ParentConfigType(parentOpt)).
			BuildStringOption(),
		optionHandle: handleOpt,
		optionInteractive: cli.Options.Modes.Interactive().
			BuildBoolOption(),
		optionMountInput: cli.Options.Functions.MountInput().
			WithKeys(&schema.Genaiz.Function.Init.MountInput).
			Validated(false).
			BuildStringOption(),
		optionMountOutput: cli.Options.Functions.MountOutput().
			WithKeys(&schema.Genaiz.Function.Init.MountOutput).
			Validated(false).
			BuildStringOption(),
		optionName: cli.Options.Functions.Name().
			WithKeys(&schema.Genaiz.Function.Init.Name).
			WithUsage("defaults to the handle value if not provided").
			WithDefaultGetter(func(ledger *config.Ledger) any {
				return ledger.GetString(handleOpt)
			}).BuildStringOption(),
		optionOem: cli.Options.Functions.Oem().
			WithKeys(&schema.Genaiz.Function.Init.Oem).
			WithDefaultGetter(sfCli.ParentOem(parentOpt)).
			BuildStringOption(),
		optionType: cli.Options.Functions.Type().
			WithKeys(&schema.Genaiz.Function.Init.Type).
			BuildStringOption(),
		optionVersion: cli.Options.Functions.Version().
			WithKeys(&schema.Genaiz.Function.Init.Version).
			WithDefaultGetter(sfCli.ParentVersion(parentOpt)).
			BuildStringOption(),
	}
}

func makeInitBuilder(ledger *config.Ledger, sfCli *Cli) layout.ConfigWriter {
	var dockerFile = ledger.GetString(sfCli.optionDockerFile)
	var dockerTag = ledger.GetString(sfCli.optionDockerTag)
	var result = &InitWriter{
		PublishOptions:       NewPublishOptions(sfCli),
		RunOptions:           NewRunOptions(sfCli),
		buildFileKeys:        &schema.Genaiz.Function.Build.File,
		buildTagKeys:         &schema.Genaiz.Function.Build.Tag,
		buildVersionKeys:     &schema.Genaiz.Function.Build.Version,
		publishPropSpecsKeys: &schema.Genaiz.Function.Publish.PropSpecs,
		vp:                   viper.New(),
	}

	if dockerFile != sfCli.optionDockerFile.DefaultGetter(ledger) {
		result.WithDockerFile(dockerFile)
	}

	return result.WithTag(dockerTag)
}

func makeInitParams(ledger *config.Ledger, initOptions *InitOptions) *layout.InitParams {
	var archTypeStrings = ledger.GetList(initOptions.optionArches)
	var functionTypeString = ledger.GetString(initOptions.optionType)
	var archTypes []layout.ArchType
	var functionType *layout.FunctionType
	var configType *shared.ConfigType
	var err error

	functionType, err = layout.FunctionTypes.FromString(functionTypeString)
	lang.HandleExit(err)
	configType, err = ledger.GetConfigType(initOptions.optionConfigType)
	lang.HandleExit(err)

	if len(archTypeStrings) > 0 {
		archTypes, err = layout.ArchTypes.AllFromStrings(&archTypeStrings)
		lang.HandleExit(err)
	}

	return &layout.InitParams{
		CreateParams: layout.CreateParams{
			ConfigParams: shared.ConfigParams{
				ConfigType: configType,
				ConfigName: ledger.ConfigName,
			},
		},
		Arches:      archTypes,
		Type:        *functionType,
		Handle:      ledger.GetString(initOptions.optionHandle),
		Name:        ledger.GetString(initOptions.optionName),
		MountInput:  ledger.GetString(initOptions.optionMountInput),
		MountOutput: ledger.GetString(initOptions.optionMountOutput),
		OEM:         ledger.GetString(initOptions.optionOem),
		Version:     ledger.GetString(initOptions.optionVersion),
	}
}
