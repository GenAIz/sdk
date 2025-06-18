package sf

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/lang"
	"genaiz.com/genaiz/lang/filez"
	"genaiz.com/genaiz/lang/panicz"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/layout"
)

type InitWriter struct {
	*PublishOptions
	*RunOptions
	baseTag     *config.StringOption
	baseVersion *config.StringOption
	vp          *viper.Viper
}

func (iw *InitWriter) BuildArches() (string, []string) {
	return iw.optionArches.Key, iw.vp.GetStringSlice(iw.optionArches.Key)
}

func (iw *InitWriter) BuildFqdn() (string, string) {
	return iw.optionFqdn.Key, iw.vp.GetString(iw.optionFqdn.Key)
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

func (iw *InitWriter) WithFqdn(value string) layout.ConfigWriter {
	if value != "" {
		var handle = iw.vp.GetString(iw.optionHandle.Key)

		iw.setTag(value, handle)
		iw.vp.Set(iw.optionFqdn.Key, value)
	}

	return iw
}

func (iw *InitWriter) WithHandle(value string) layout.ConfigWriter {
	if value != "" {
		var fqdn = iw.vp.GetString(iw.optionFqdn.Key)

		iw.setTag(fqdn, value)
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
		iw.vp.Set(iw.optionOem.Key, value)
	}

	return iw
}

func (iw *InitWriter) WithOutput(value string) layout.ConfigWriter {
	if value != "" {
		iw.vp.Set(iw.optionMountOutput.Key, value)
		iw.vp.Set(iw.optionMountVar.Key, filepath.Join(value, "var"))
		iw.vp.Set(iw.optionMountLog.Key, filepath.Join(value, "log"))
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
		iw.vp.Set(iw.baseVersion.Key, "latest")
	}

	return iw
}

func (iw *InitWriter) Write(filepath string) error {
	return iw.vp.WriteConfigAs(filepath)
}

func (iw *InitWriter) setTag(fqdn string, handle string) {
	var tagTokens []string

	if fqdn != "" {
		var fqdnTokens = strings.Split(fqdn, ".")
		var size = len(fqdnTokens)

		if size > 1 {
			tagTokens = append(tagTokens, strings.Join(fqdnTokens[size-2:], "."))
		}
	}

	if handle != "" {
		tagTokens = append(tagTokens, handle)
	}

	iw.vp.Set(iw.baseTag.Key, strings.Join(tagTokens, "/"))
}

type InitTaskFactory func(layout.ConfigWriter) *task.Task[layout.InitParams]

type InitExecutor struct {
	BaseExecutor
	*InitOptions

	initTaskFactory InitTaskFactory
}

func (ie *InitExecutor) Display() {
	ie.Repo.DisplayOptions(
		&ie.optionArches.Option,
		&ie.optionConfigType.Option,
		&ie.optionFqdn.Option,
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

	ie.Repo.DisplayChangeDir()
	ie.initTaskFactory(builder).Pretend(params, ie.Repo.Logger)
}

func (ie *InitExecutor) Proceed() {
	var builder = ie.makeInitBuilder()
	var params = ie.makeInitParams()
	var plan = task.NewPlan("Init", ie.Repo.Logger)

	task.Single(plan, params, ie.initTaskFactory(builder))
}

func (ie *InitExecutor) makeInitBuilder() *InitWriter {
	return makeInitBuilder(ie.Cli)
}

func (ie *InitExecutor) makeInitParams() *layout.InitParams {
	return makeInitParams(ie.Repo, ie.InitOptions)
}

type InitOptions struct {
	*PublishOptions
	optionConfigType  *config.StringOption
	optionInteractive *config.BoolOption
	optionMountInput  *config.StringOption
	optionMountOutput *config.StringOption
}

func (io *InitOptions) allDefiners() []config.Definer {
	return []config.Definer{
		io.optionArches,
		io.optionConfigType,
		io.optionFqdn,
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

func NewInit(repo *config.Repo, cli *Cli) *cobra.Command {
	var options = NewInitOptions()
	var init = &cobra.Command{
		Use:     "init",
		Short:   "Initiates an existing Smart Function configuration values",
		Long:    "Initiates an existing Smart Function configuration values, interactively by default",
		Example: "genaiz sf init --oem com.genaiz",
		Run: func(cmd *cobra.Command, args []string) {
			cli.Exec(repo, NewInitExecutor(cmd.Context(), repo, cli, options))
		},
	}

	repo.Register(init, options.allDefiners()...)
	return init
}

func NewInitExecutor(ctx context.Context, repo *config.Repo, cli *Cli, options *InitOptions) *InitExecutor {
	return &InitExecutor{
		BaseExecutor: BaseExecutor{
			Context: ctx,
			Repo:    repo,
			Cli:     cli,
		},
		InitOptions: options,

		initTaskFactory: layout.NewInitTask,
	}
}

func NewInitOptions() *InitOptions {
	var initCmd = "Init"

	return &InitOptions{
		PublishOptions:    newPublishOptions(initCmd),
		optionConfigType:  newOptionConfigType(initCmd),
		optionInteractive: newOptionInteractive(),
		optionMountInput:  newOptionMountInput(initCmd, false),
		optionMountOutput: newOptionMountOutput(initCmd, false),
	}
}

func makeInitBuilder(cli *Cli) *InitWriter {
	return &InitWriter{
		PublishOptions: newPublishOptions("Publish"),
		RunOptions:     NewRunOptions(cli),
		baseTag:        newOptionDockerTag(),
		baseVersion:    newOptionDockerVersion(),
		vp:             viper.New(),
	}
}

func makeInitParams(repo *config.Repo, initOptions *InitOptions) *layout.InitParams {
	var archTypeStrings = repo.GetList(initOptions.optionArches)
	var functionTypeString = repo.GetString(initOptions.optionType)
	var archTypes []layout.ArchType
	var functionType *layout.FunctionType
	var err error

	functionType, err = layout.FunctionTypes.FromString(functionTypeString)
	lang.HandleExit(err)

	if len(archTypeStrings) > 0 {
		archTypes, err = layout.ArchTypes.AllFromStrings(&archTypeStrings)
		lang.HandleExit(err)
	}

	return &layout.InitParams{
		CreateParams: layout.CreateParams{
			ConfigType: toConfigType(repo, initOptions.optionConfigType),
			ConfigName: repo.ConfigName,
		},
		Arches:      archTypes,
		FQDN:        repo.GetString(initOptions.optionFqdn),
		Type:        *functionType,
		Handle:      repo.GetString(initOptions.optionHandle),
		Name:        repo.GetString(initOptions.optionName),
		MountInput:  repo.GetString(initOptions.optionMountInput),
		MountOutput: repo.GetString(initOptions.optionMountOutput),
		OEM:         repo.GetString(initOptions.optionOem),
		Version:     repo.GetString(initOptions.optionVersion),
	}
}

func newOptionConfigType(cmd string) *config.StringOption {
	return &config.StringOption{
		Option: config.Option{
			Key:       "SF." + cmd + ".ConfigType",
			Param:     "configType",
			Usage:     "sets the format of the configuration file created. Supported values are \"yaml\", \"toml\", \"json\" or \"none\"",
			Validator: config.Optionally(config.AnyOfEnumerated(layout.ConfigTypes)),
		},
	}
}

func newOptionInteractive() *config.BoolOption {
	return &config.BoolOption{
		Option: config.Option{
			Param:        "it",
			Usage:        "specify smart function value through STDIN questionnaire",
			DefaultValue: "false",
		},
	}
}
