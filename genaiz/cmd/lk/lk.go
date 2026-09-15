package lk

import (
	"context"
	"os"
	"strconv"
	"strings"

	"github.com/awnumar/memguard"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"

	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/cmd/dk"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
	"genaiz.com/genaiz/task/locker"
	"genaiz.com/genaiz/task/shared"
)

const (
	passphraseEnvKey = "GENAIZ_LK_PASSWORD"
	passphrasePrompt = "enter passphrase: "
)

type FindLinksTaskFactory func() *task.Task[broker.DataLinkParams]

type BaseExecutor struct {
	Cli     *Cli
	Context context.Context
	Ledger  *config.Ledger
}

func (be BaseExecutor) newBaseParams(handle string, optionLocker *config.StringOption) *locker.BaseParams {
	var baseParams = &locker.BaseParams{
		LockerHandle: handle,
		LockerPath:   be.Ledger.GetString(optionLocker),
	}

	if envPwd := os.Getenv(passphraseEnvKey); envPwd != "" {
		baseParams.Passphrase = memguard.NewEnclave([]byte(envPwd))
	} else if pwdEnclave := be.Ledger.QuerySecret(passphrasePrompt); pwdEnclave != nil {
		baseParams.Passphrase = pwdEnclave
	}

	return baseParams
}

type BaseLockerExecutor struct {
	BaseExecutor
	*LockerOptions

	addHandle   string
	addOem      string
	addSequence *int
	addVersion  string
	handleArg   string
	keyArg      string
	secretArg   *memguard.Enclave
	valueArg    string

	accountParams config.AccountParametric

	dataLinksWriterFactory dk.DataLinksWriterFactory
	collectLinkTaskFactory dk.CollectLinkTaskFactory
	exportLinkTaskFactory  dk.ExportLinkTaskFactory
}

func (bpe *BaseLockerExecutor) Add(handleArg string, dataLinkArg string) error {
	bpe.handleArg = handleArg
	bpe.addOem, bpe.addHandle, bpe.addVersion = dk.ParseDataLinkArgument(dataLinkArg)

	if bpe.addOem == "" {
		return errorDataLinkOemRequired
	}

	if bpe.addVersion == "" {
		return errorDataLinkVersionRequired
	}

	if versionParts := strings.Split(bpe.addVersion, "-rc-"); len(versionParts) == 2 {
		var seq int
		var err error

		if seq, err = strconv.Atoi(versionParts[1]); err != nil {
			return errorDataLinkSequenceInvalid
		}

		bpe.addSequence = &seq
	}

	return nil
}

func (bpe *BaseLockerExecutor) Display() {
	var argMap = map[string]string{
		"handle": bpe.handleArg,
	}

	if bpe.addHandle == "" {
		argMap["prop-key"] = bpe.keyArg

		if bpe.secretArg == nil {
			argMap["prop-value"] = bpe.valueArg
		} else {
			// Never display STDIN values, regardless of whether it is used for a secret or not.
			argMap["prop-value"] = "********"
		}
	} else {
		argMap["datalink-oem"] = bpe.addOem
		argMap["datalink-handle"] = bpe.addHandle
		argMap["datalink-version"] = bpe.addVersion
		argMap["datalink-seq"] = cast.ToString(bpe.addSequence)
	}

	bpe.Ledger.DisplayOptionsWithMap(&argMap,
		&bpe.optionAccount.Option,
		&bpe.optionLocker.Option)
}

func (bpe *BaseLockerExecutor) Update(handleArg string, keyArg string, valueArg string) error {
	var err error

	if bpe.secretArg, err = bpe.Ledger.QueryPipe(); err == nil {
		bpe.handleArg = handleArg
		bpe.keyArg = keyArg
		bpe.valueArg = valueArg
		return nil
	}

	return err
}

func (bpe *BaseLockerExecutor) newConfigParams() *shared.ConfigParams {
	return &shared.ConfigParams{
		ConfigName:   bpe.Ledger.ConfigName,
		ConfigFolder: bpe.Ledger.UserPath,
		ConfigType:   new(shared.ConfigTypeYaml),
	}
}

func (bpe *BaseLockerExecutor) newDataLinkParams(brokerParams *broker.Broker, configParams *shared.ConfigParams) *broker.DataLinkParams {
	return &broker.DataLinkParams{
		Broker:       *brokerParams,
		ConfigParams: *configParams,
		DataLink: &broker.DataLink{
			Oem:     bpe.addOem,
			Handle:  bpe.addHandle,
			Version: bpe.addVersion,
		},
	}
}

type BasePublishExecutor struct {
	BaseExecutor
	*PublishOptions

	handleArg string

	accountParams config.AccountParametric

	dataLinkFindTaskFactory FindLinksTaskFactory
}

func (bpe *BasePublishExecutor) Display() {
	var argMap = map[string]string{
		"handle": bpe.handleArg,
	}

	bpe.Ledger.DisplayOptionsWithMap(&argMap,
		&bpe.optionAccount.Option,
		&bpe.optionDescription.Option,
		&bpe.optionLocker.Option,
		&bpe.optionName.Option,
		&bpe.optionVisibility.Option)
}

type Cli struct {
	cli.BaseCli
}

type LockerOptions struct {
	optionAccount *config.StringOption
	optionLocker  *config.StringOption
}

func (so LockerOptions) allDefiners() []config.Definer {
	return []config.Definer{
		so.optionAccount,
		so.optionLocker,
	}
}

type PublishOptions struct {
	optionAccount     *config.StringOption
	optionDescription *config.StringOption
	optionLocker      *config.StringOption
	optionName        *config.StringOption
	optionVisibility  *config.StringOption
}

func (po PublishOptions) allDefiners() []config.Definer {
	return []config.Definer{
		po.optionAccount,
		po.optionDescription,
		po.optionLocker,
		po.optionName,
		po.optionVisibility,
	}
}

func NewLk(ledger *config.Ledger, confirm cli.Interactive, dry, pretend cli.Decisive) *cobra.Command {
	var lkCli = NewLkCli(confirm, dry, pretend)
	var lkCmd = &cobra.Command{
		Use:     "locker",
		Aliases: []string{"lk"},
		Short:   "GenAIz Locker Toolkit",
	}

	lkCmd.AddCommand(NewInit(ledger, lkCli))
	lkCmd.AddCommand(NewSource(ledger, lkCli))
	lkCmd.AddCommand(NewStore(ledger, lkCli))
	return lkCmd
}

func NewLkCli(confirm cli.Interactive, dry, pretend cli.Decisive) *Cli {
	return &Cli{
		BaseCli: cli.BaseCli{
			Confirm: confirm,
			Dry:     dry,
			Pretend: pretend,
		},
	}
}
