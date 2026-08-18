package lk

import (
	"fmt"
	"io"
	"os"

	"github.com/awnumar/memguard"
	"github.com/spf13/cobra"

	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/lang"
	"genaiz.com/genaiz/lang/dialogz"
	"genaiz.com/genaiz/schema"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/locker"
)

type dialogConfirmation func(io.Writer, io.Reader, string) bool
type initLockerTaskFactory func() *task.Task[locker.InitParams]

type InitExecutor struct {
	BaseExecutor
	*InitOptions

	path                  string
	initLockerTaskFactory initLockerTaskFactory
}

func (ie InitExecutor) Display() {
	ie.Ledger.OverrideString(ie.optionPath, ie.path)
	ie.Ledger.DisplayOptions(
		&ie.optionPath.Option,
		&ie.optionOverwrite.Option,
		&ie.optionUpdate.Option,
	)
}

func (ie InitExecutor) Pretend() {
	var params *locker.InitParams
	var err error

	if params, err = ie.newLockerInitParams(); err == nil {
		var plan = task.NewPlan("init", ie.Ledger.Logger)

		plan.Sequence(task.NewPretender(params, ie.initLockerTaskFactory()))
		return
	}

	lang.HandleExit(err)
}

func (ie InitExecutor) Proceed() {
	var params *locker.InitParams
	var err error

	if params, err = ie.newLockerInitParams(); err == nil {
		var plan = task.NewPlan("init", ie.Ledger.Logger)

		plan.PrintReportsOnly = true
		plan.Sequence(task.NewWorker(params, ie.initLockerTaskFactory()))
		return
	}

	lang.HandleExit(err)
}

func (ie InitExecutor) newLockerInitParams() (*locker.InitParams, error) {
	var update = ie.Ledger.GetBool(ie.optionUpdate)
	var message = "enter passphrase: "
	var oldEnclave *memguard.Enclave
	var resolved string
	var err error

	ie.Ledger.OverrideString(ie.optionPath, ie.path)
	resolved = ie.Ledger.GetString(ie.optionPath)

	if err = filez.IsReadable(resolved); err == nil {
		var overwrite = ie.Ledger.GetBool(ie.optionOverwrite)

		if !overwrite && !update {
			if update = ie.InitOptions.confirmUpdate(resolved); !update {
				if ok := ie.InitOptions.confirmOverwrite(resolved); !ok {
					return nil, fmt.Errorf("the locker [%s] is already initialized", resolved)
				}
			}
		}
	} else if os.IsPermission(err) {
		return nil, fmt.Errorf("the locker [%s] can not be read", resolved)
	}

	if update {
		oldEnclave = ie.Ledger.QuerySecret("enter current passphrase: ")
		message = "enter new passphrase: "
	}

	return &locker.InitParams{
		LockerPath:    resolved,
		OldPassphrase: oldEnclave,
		Passphrase:    ie.Ledger.QuerySecret(message),
		Update:        update,
	}, nil
}

type InitOptions struct {
	optionOverwrite *config.BoolOption
	optionPath      *config.StringOption
	optionUpdate    *config.BoolOption
	dialogYes       dialogConfirmation
	stdIn           io.Reader
	stdOut          io.Writer
}

func (o InitOptions) allDefiners() []config.Definer {
	return []config.Definer{
		o.optionOverwrite,
		o.optionUpdate,
	}
}

func (o InitOptions) confirmOverwrite(path string) bool {
	var message = fmt.Sprintf("Overwrite %s ?", path)

	return o.dialogYes(o.stdOut, o.stdIn, message)
}

func (o InitOptions) confirmUpdate(path string) bool {
	var message = fmt.Sprintf("Update %s ?", path)

	return o.dialogYes(o.stdOut, o.stdIn, message)
}

func NewInit(ledger *config.Ledger, dkCli *Cli) *cobra.Command {
	var initOptions = NewInitOptions()
	var initCmd = &cobra.Command{
		Use:     "init [PATH]",
		Short:   "Initializes an encrypted data locker",
		Long:    "Initializes an encrypted data locker file under the specified path or the user default location",
		Example: "genaiz lk init ./myLocker.bin",
		Args:    cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			var path = cli.ArgsOptionalSingle(args)

			dkCli.Exec(ledger, NewInitExecutor(cmd, ledger, dkCli, path, initOptions))
		},
	}

	ledger.Register(initCmd, initOptions.allDefiners()...)
	return initCmd
}

func NewInitExecutor(cmd *cobra.Command, ledger *config.Ledger, dkCli *Cli, path string, options *InitOptions) *InitExecutor {
	return &InitExecutor{
		BaseExecutor: BaseExecutor{
			Cli:     dkCli,
			Context: cmd.Context(),
			Ledger:  ledger,
		},
		InitOptions: options,

		path:                  path,
		initLockerTaskFactory: locker.NewInitLockerTask,
	}
}

func NewInitOptions() *InitOptions {
	return &InitOptions{
		optionOverwrite: cli.Options.Lockers.Overwrite().
			WithKeys(&schema.Genaiz.Locker.Init.Overwrite).
			BuildBoolOption(),
		optionPath: cli.Options.Lockers.Path().
			WithKeys(&schema.Genaiz.Locker.Init.Path).
			BuildStringOption(),
		optionUpdate: cli.Options.Lockers.Update().
			WithKeys(&schema.Genaiz.Locker.Init.Update).
			BuildBoolOption(),

		dialogYes: dialogz.ConfirmYes,
		stdIn:     os.Stdin,
		stdOut:    os.Stdout,
	}
}
