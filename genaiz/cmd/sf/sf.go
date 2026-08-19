// Package sf provides commands for managing Genaiz Smart Functions.
// Smart Functions commands include build, create, debug, init, list, run, start, stop and test.
//
// See: genaiz sf --help
package sf

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"genaiz.com/genaiz-lib/lang/errorz"
	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/lang"
	"genaiz.com/genaiz/schema"
	"genaiz.com/genaiz/task/broker"
	"genaiz.com/genaiz/task/shared"
)

var (
	errInvalidConnectorType = errors.New("data links can only be configured for connector functions")
)

type BaseExecutor struct {
	Cli     *Cli
	Context context.Context
	Ledger  *config.Ledger
}

func (be *BaseExecutor) validateConnector(typeOption *config.StringOption) error {
	var functionType = be.Ledger.GetString(typeOption)

	if strings.ToLower(functionType) != shared.FunctionTypeConnector {
		return errInvalidConnectorType
	}

	return nil
}

type EnvOptions struct {
	optionEnvFile *config.StringOption
	optionEnvVars *config.ListOption
}

func (eo EnvOptions) makeEnvMap(ledger *config.Ledger) (map[string]string, error) {
	var envFile = ledger.GetString(eo.optionEnvFile)
	var sfType = ledger.GetString(cli.NewOptionBuilder().
		WithKeys(&schema.Genaiz.Function.Publish.Type).
		BuildStringOption())
	var result map[string]string
	var err error

	if result, err = eo.parseEnvFile(envFile); err != nil && !errorz.IsPathError(err) {
		// ignore files that do not exist
		return nil, err
	}

	for _, envPair := range ledger.GetList(eo.optionEnvVars) {
		parts := strings.SplitN(envPair, "=", 2)
		result[parts[0]] = parts[1]
	}

	result["SF_TYPE"] = sfType
	return result, nil
}

func (eo EnvOptions) parseEnvFile(filePath string) (map[string]string, error) {
	var result = make(map[string]string)
	var fd *os.File
	var err error

	if fd, err = os.Open(filePath); err == nil {
		defer filez.CloseSilently(fd)
		var scanner = bufio.NewScanner(fd)

		for scanner.Scan() {
			var keyPair = scanner.Text()

			if !config.Validation.EnvKeyValue(keyPair) {
				return nil, fmt.Errorf("%s, is not a valid key/pair value", keyPair)
			}

			parts := strings.SplitN(keyPair, "=", 2)
			result[parts[0]] = strings.ReplaceAll(strings.Trim(parts[1], "\"'"), "\\n", "\n")
		}
	}

	return result, err
}

type InnerOptions struct {
	innerSources *config.ListOption
	innerStores  *config.ListOption
}

func (io InnerOptions) getFunctionDataLinks(ledger *config.Ledger) []string {
	var dataLink []string

	dataLink = append(dataLink, ledger.GetList(io.innerSources)...)
	dataLink = append(dataLink, ledger.GetList(io.innerStores)...)
	return dataLink
}

type Cli struct {
	cli.BaseCli

	optionDockerContext *config.StringOption
	optionDockerFile    *config.StringOption
	optionDockerRepo    *config.StringOption
	optionDockerVersion *config.StringOption

	parentConfigType *shared.ConfigType
	parentSolution   *broker.Solution
}

func (c *Cli) ContainerPrefix(ledger *config.Ledger) any {
	var repo = strings.ReplaceAll(ledger.GetString(c.optionDockerRepo), "/", "-")
	var workspace = ledger.GetWorkspace()

	if workspace != "" {
		return workspace + "-" + repo
	}

	return repo
}

func (c *Cli) DefaultRunImage(ledger *config.Ledger) any {
	var repo = ledger.GetString(c.optionDockerRepo)
	var version = ledger.GetString(c.optionDockerVersion)
	var parts []string

	if repo != "" {
		parts = append(parts, repo)
	}

	if version != "" {
		parts = append(parts, version)
	}

	return strings.Join(parts, ":")
}

func (c *Cli) ParentConfigType(parentOption *config.StringOption) func(*config.Ledger) any {
	return func(ledger *config.Ledger) any {
		if c.parentConfigType == nil {
			c.parentSolution = c.parseParentSolution(ledger, parentOption)
		}

		if c.parentConfigType == nil {
			// Kludge: Event though we have no default definition under cli/options, defaulting to yaml because the command definitions all default to it
			c.parentConfigType = new(shared.ConfigTypeYaml)
		}

		return *c.parentConfigType
	}
}

func (c *Cli) ParentOem(parentOption *config.StringOption) func(*config.Ledger) any {
	return func(ledger *config.Ledger) any {
		if c.parentSolution == nil {
			c.parentSolution = c.parseParentSolution(ledger, parentOption)
		}

		return c.parentSolution.Oem
	}
}

func (c *Cli) ParentVersion(parentOption *config.StringOption) func(*config.Ledger) any {
	return func(ledger *config.Ledger) any {
		if c.parentSolution == nil {
			c.parentSolution = c.parseParentSolution(ledger, parentOption)
		}

		return c.parentSolution.Version
	}
}

func (c *Cli) SfOptions() []*config.Option {
	return []*config.Option{
		&c.optionDockerContext.Option,
		&c.optionDockerFile.Option,
		&c.optionDockerRepo.Option,
		&c.optionDockerVersion.Option,
	}
}

func (c *Cli) allDefiners() []config.Definer {
	return []config.Definer{
		c.optionDockerContext,
		c.optionDockerFile,
		c.optionDockerRepo,
		c.optionDockerVersion,
	}
}

func (c *Cli) parseParentSolution(ledger *config.Ledger, parentOption *config.StringOption) *broker.Solution {
	var solutionPath = ledger.GetString(parentOption)
	var solutionReader = config.NewSolutionReader(ledger).
		WithConfigPath(solutionPath)

	if result, err := solutionReader.ReadName(ledger.ConfigName); err == nil {
		c.parentConfigType = new(solutionReader.GetConfigType())
		return result
	} else if !errorz.IsPathError(err) {
		lang.HandleExit(err)
	}

	return &broker.Solution{Version: "1.0.0"}
}

func NewSf(ledger *config.Ledger, confirm cli.Interactive, dry, pretend cli.Decisive) *cobra.Command {
	var sfCli = NewSfCli(confirm, dry, pretend)
	var sfCmd = &cobra.Command{
		Use:     "function",
		Aliases: []string{"sf"},
		Short:   "GenAIz Smart Function Toolkit",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			var flags = cmd.Flags()

			ledger.ToWorkDir(sfCli.optionDockerContext, flags)
			ledger.FromWorkDir(sfCli.optionDockerFile, flags)
		},
	}

	ledger.Register(sfCmd, sfCli.allDefiners()...)
	sfCmd.AddCommand(
		NewBuild(ledger, sfCli),
		NewCreate(ledger, sfCli),
		NewData(ledger, sfCli),
		NewInit(ledger, sfCli),
		NewList(ledger, sfCli),
		NewProp(ledger, sfCli),
		NewRun(ledger, sfCli),
		NewTest(ledger, sfCli),
		NewStop(ledger, sfCli),
		NewStart(ledger, sfCli),
		NewPublish(ledger, sfCli),
	)
	// The sf command captures context modifications and those are needed before the ledger sets defaults
	cobra.OnInitialize(func() { ledger.ChangeWorkDir(sfCli.optionDockerContext) })
	return sfCmd
}

func NewSfCli(confirm cli.Interactive, dry, pretend cli.Decisive) *Cli {
	return &Cli{
		BaseCli: cli.BaseCli{
			Confirm: confirm,
			Dry:     dry,
			Pretend: pretend,
		},

		optionDockerContext: cli.Options.Docker.ContextPath().BuildStringOption(),
		optionDockerFile:    cli.Options.Docker.FilePath().BuildStringOption(),
		optionDockerRepo: cli.Options.Docker.Repository().
			WithDefaultGetter(makeResolveSfRepo()).
			BuildStringOption(),
		optionDockerVersion: cli.Options.Docker.Version().BuildStringOption(),
	}
}

func makeResolveSfRepo() func(ledger *config.Ledger) any {
	return func(ledger *config.Ledger) any {
		var parent = filepath.Base(filepath.Dir(ledger.WorkDir))
		var fn = filepath.Base(ledger.WorkDir)

		if config.Validation.Oem(parent) && config.Validation.Handle(fn) {
			// Should not set a default value that is not a valid repository string
			return fmt.Sprintf("%s/%s", parent, fn)
		}

		return ""
	}
}
