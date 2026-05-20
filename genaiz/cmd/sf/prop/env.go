package prop

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	"genaiz.com/genaiz-lib/lang/errorz"
	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/cmd/dk"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/lang"
	"genaiz.com/genaiz/schema"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
	"genaiz.com/genaiz/task/shared"
)

type EnvExecutor struct {
	*EnvOptions
	dk.SyncBridge
	Context context.Context
	Ledger  *config.Ledger

	innerSources *config.ListOption
	innerStores  *config.ListOption

	key   string
	value string
}

func (ee EnvExecutor) Display() {
	ee.Ledger.DisplayOptionsWithMap(&map[string]string{
		"key":   ee.key,
		"value": ee.value,
	},
		&ee.optionContext.Option,
		&ee.optionEnvFile.Option,
	)
}

func (ee EnvExecutor) List() ([]cobra.Completion, error) {
	var dataLinks = ee.getFunctionDataLinks()
	var specs []broker.PropSpec
	var results []string
	var err error

	if len(dataLinks) > 0 {
		var varSpecs []shared.VarSpec

		if varSpecs, err = ee.collectDataLinks(dataLinks); err == nil {
			for _, spec := range varSpecs {
				if spec.GetDescription() != "" && !strings.EqualFold(spec.GetKey(), spec.GetDescription()) {
					results = append(results, cobra.CompletionWithDesc(spec.GetKey(), spec.GetDescription()))
				} else {
					results = append(results, spec.GetKey())
				}
			}
		} else {
			return nil, err
		}
	}

	if err = ee.Ledger.Unmarshal(schema.Genaiz.Function.Publish.PropSpecs, &specs); err == nil {
		for _, spec := range specs {
			if spec.Description != "" && !strings.EqualFold(spec.Key, spec.Description) {
				results = append(results, cobra.CompletionWithDesc(spec.Key, spec.Description))
			} else {
				results = append(results, spec.Key)
			}
		}
	} else {
		ee.Ledger.Logger.Errorf("Could not read propSpecs: [%s]", err)
	}

	if len(results) > 0 {
		return results, nil
	}

	return nil, err
}

func (ee EnvExecutor) Pretend() {
	var envFile = ee.Ledger.GetString(ee.optionEnvFile)
	var envBytes []byte
	var err error

	if err = ee.validate(); err == nil {
		if envBytes, err = os.ReadFile(envFile); err != nil && !errorz.IsPathError(err) {
			lang.HandleExit(err)
		} else if regexp.MustCompile(`^` + ee.key).Match(envBytes) {
			_, _ = fmt.Printf("sed -i~ '/^%s=/s/=.*/=\"%s\"/' %s", ee.key, ee.value, envFile)
		} else {
			_, _ = fmt.Printf("echo %s=\"%s\" > %s", ee.key, ee.value, envFile)
		}
	} else {
		lang.HandleExit(err)
	}
}

func (ee EnvExecutor) Proceed() {
	var envFile = ee.Ledger.GetString(ee.optionEnvFile)
	var resultEnv []string
	var fd *os.File
	var err error

	if err = ee.validate(); err == nil {
		var replaced = false

		if fd, err = os.Open(envFile); err == nil {
			var scanner = bufio.NewScanner(fd)

			for scanner.Scan() {
				if regexp.MustCompile(`^` + ee.key).Match(scanner.Bytes()) {
					resultEnv = append(resultEnv, fmt.Sprintf("%s=\"%s\"", ee.key, ee.value))
					replaced = true
				} else {
					resultEnv = append(resultEnv, scanner.Text())
				}
			}

			filez.CloseSilently(fd)
		}

		if errorz.IsPathError(err) || err == nil {
			if !replaced {
				resultEnv = append(resultEnv, fmt.Sprintf("%s=\"%s\"", ee.key, ee.value))
			}

			if fd, err = os.Create(envFile); err == nil {
				var writer = bufio.NewWriter(fd)

				defer filez.CloseSilently(fd)

				for _, line := range resultEnv {
					_, _ = fmt.Fprintln(writer, line)
				}

				err = writer.Flush()
			}
		}
	}

	lang.HandleExit(err)
}

func (ee EnvExecutor) collectDataLinks(dataLinks []string) ([]shared.VarSpec, error) {
	var datalinkWorkers []task.Worker
	var err error

	if datalinkWorkers, err = ee.MakeSyncWorkers(dataLinks, ee.Ledger, ee.optionNoPropSync); err == nil &&
		len(datalinkWorkers) > 0 {
		var results []shared.VarSpec
		var plan = task.NewPlanBuilder(ee.Ledger.Logger).
			WithFailures(func(i interface{}) { err = i.(error) }).
			WithReturn(func(i interface{}) {
				if specTracking, ok := i.(shared.VarSpecTracking); ok {
					results = append(results, specTracking.VarSpecs...)
				}
			}).Build()

		plan.Sequence(datalinkWorkers...)
		return results, err
	}

	return nil, err
}

func (ee EnvExecutor) getFunctionDataLinks() []string {
	var dataLink []string

	dataLink = append(dataLink, ee.Ledger.GetList(ee.innerSources)...)
	dataLink = append(dataLink, ee.Ledger.GetList(ee.innerStores)...)
	return dataLink
}

func (ee EnvExecutor) validate() error {
	var specs []broker.PropSpec
	var err error

	if err = ee.Ledger.Unmarshal(schema.Genaiz.Function.Publish.PropSpecs, &specs); err == nil {
		if propSpec := broker.FindPropSpec(specs, ee.key); propSpec != nil {
			return propSpec.Validate(ee.value)
		}
	}

	if dataLinks := ee.getFunctionDataLinks(); len(dataLinks) > 0 {
		var varSpecs []shared.VarSpec

		if varSpecs, err = ee.collectDataLinks(dataLinks); err == nil {
			if varSpec := shared.FindVarSpec(varSpecs, ee.key); varSpec != nil {
				return varSpec.Validate(ee.value)
			}
		} else {
			return err
		}
	}

	return fmt.Errorf("property specification for key [%s] is not defined", ee.key)
}

type EnvOptions struct {
	optionContext    *config.StringOption
	optionEnvFile    *config.StringOption
	optionNoPropSync *config.BoolOption
}

func (eo EnvOptions) allDefiners() []config.Definer {
	return []config.Definer{
		eo.optionContext,
		eo.optionEnvFile,
		eo.optionNoPropSync,
	}
}

func NewEnv(ledger *config.Ledger, cli *cli.BaseCli) *cobra.Command {
	var envOptions = NewEnvOptions()
	var envCmd = &cobra.Command{
		Use:     "env KEY VALUE",
		Short:   "Defines a property spec value",
		Long:    "Defines a property spec value, for the current Smart Function, under the specified context or working directory",
		Example: "genaiz sf prop env MY_KEY value1",
		Args: cobra.MatchAll(cobra.ExactArgs(2), func(cmd *cobra.Command, args []string) error {
			if !config.Validation.EnvKey(args[0]) {
				return errors.New("invalid environment key; [A-Z_][A-Z0-9_]*")
			}

			return nil
		}),
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			var flags = cmd.Flags()

			ledger.ChangeWorkDir(envOptions.optionContext)
			ledger.FromWorkDir(envOptions.optionEnvFile, flags)
		},
		Run: func(cmd *cobra.Command, args []string) {
			cli.Exec(ledger, NewEnvExecutor(ledger, envOptions, args[0], args[1]))
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
			if len(args) < 2 {
				var executor = NewEnvExecutor(ledger, envOptions, toComplete, "")
				var keys []string
				var err error

				if keys, err = executor.List(); err != nil {
					return nil, cobra.ShellCompDirectiveError |
						cobra.ShellCompDirectiveNoFileComp
				}

				return keys, cobra.ShellCompDirectiveNoFileComp
			}

			return nil, cobra.ShellCompDirectiveDefault
		},
	}

	ledger.Register(envCmd, envOptions.allDefiners()...)
	return envCmd
}

func NewEnvExecutor(ledger *config.Ledger, options *EnvOptions, key, value string) *EnvExecutor {
	return &EnvExecutor{
		EnvOptions: options,
		Ledger:     ledger,
		SyncBridge: dk.NewSyncBridgeBuilder().Build(),

		innerSources: cli.NewOptionBuilder().
			WithKeys(&schema.Genaiz.Function.Publish.DataSources).
			BuildListOption(),
		innerStores: cli.NewOptionBuilder().
			WithKeys(&schema.Genaiz.Function.Publish.DataStores).
			BuildListOption(),

		key:   key,
		value: value,
	}
}

func NewEnvOptions() *EnvOptions {
	return &EnvOptions{
		optionContext: cli.Options.Docker.ContextPath().
			WithKeys(&schema.Genaiz.Function.Env.Context).
			WithValidator(config.Validation.DirExists).
			BuildStringOption(),
		optionEnvFile: cli.Options.Docker.EnvFile().
			WithKeys(&schema.Genaiz.Function.Env.File).
			BuildStringOption(),
		optionNoPropSync: cli.Options.Functions.NoPropSync().
			WithKeys(&schema.Genaiz.Function.Env.NoPropSync).
			BuildBoolOption(),
	}
}
