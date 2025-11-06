package prop

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"

	"github.com/spf13/cobra"

	"genaiz.com/genaiz-lib/lang/errorz"
	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/lang"
	"genaiz.com/genaiz/schema"
	"genaiz.com/genaiz/task/broker"
)

type EnvExecutor struct {
	*EnvOptions
	Context context.Context
	Ledger  *config.Ledger

	innerPropSpecs *config.Option
	key            string
	value          string
}

func (ee *EnvExecutor) Display() {
	ee.Ledger.DisplayOptionsWithMap(&map[string]string{
		"key":   ee.key,
		"value": ee.value,
	},
		&ee.optionContext.Option,
		&ee.optionEnvFile.Option,
	)
}

func (ee *EnvExecutor) Pretend() {
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

func (ee *EnvExecutor) Proceed() {
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

func (ee *EnvExecutor) validate() error {
	var specs = ee.Ledger.Get(ee.innerPropSpecs)

	if propSpec := broker.FindPropSpec(specs, ee.key); propSpec != nil {
		return propSpec.Validate(ee.value)
	}

	return fmt.Errorf("property specification for key [%s] is not defined", ee.key)
}

type EnvOptions struct {
	optionContext *config.StringOption
	optionEnvFile *config.StringOption
}

func (eo EnvOptions) allDefiners() []config.Definer {
	return []config.Definer{
		eo.optionContext,
		eo.optionEnvFile,
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
	}

	ledger.Register(envCmd, envOptions.allDefiners()...)
	return envCmd
}

func NewEnvExecutor(ledger *config.Ledger, options *EnvOptions, key, value string) *EnvExecutor {
	return &EnvExecutor{
		EnvOptions: options,
		Ledger:     ledger,

		innerPropSpecs: cli.NewOptionBuilder().
			WithKeys(&schema.Genaiz.Function.Publish.PropSpecs).
			BuildOption(),
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
	}
}
