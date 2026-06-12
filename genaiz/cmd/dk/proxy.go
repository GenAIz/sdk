package dk

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/spf13/cast"
	"github.com/spf13/cobra"

	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/cmd/dk/proxy"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/lang"
	"genaiz.com/genaiz/schema"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
	"genaiz.com/genaiz/task/shared"
)

type ProxyExecutor struct {
	BaseExecutor
	*ProxyOptions

	addedProxy             *broker.Proxy
	removedProxy           *broker.Proxy
	editedLink             *broker.DataLink
	dataLinksWriterFactory DataLinksWriterFactory
	editLinkTaskFactory    EditLinkTaskFactory
}

func (pe *ProxyExecutor) Add(handleArg string, host string, port int) error {
	var writer *DataLinksWriter
	var err error

	pe.setOptions(pe.Ledger, handleArg)

	if writer, err = pe.loadDataLinkWriter(); err == nil {
		var oem = pe.Ledger.GetString(pe.optionOem)
		var handle = pe.Ledger.GetString(pe.optionHandle)
		var version = pe.Ledger.GetString(pe.optionVersion)

		if pe.editedLink = writer.GetDataLink(oem, handle, version); pe.editedLink != nil {
			if slices.ContainsFunc(pe.editedLink.OutboundProxies, func(proxy broker.Proxy) bool {
				return strings.EqualFold(host, proxy.Host) && port == proxy.Port
			}) {
				return fmt.Errorf("proxy [%s:%d] is already defined", host, port)
			}

			pe.addedProxy = pe.newProxy(host, port)
			pe.editedLink.OutboundProxies = append(pe.editedLink.OutboundProxies, *pe.addedProxy)
			pe.Cli.Exec(pe.Ledger, pe)
			return nil
		}

		return fmt.Errorf("data link [%s/%s:%s] not found", oem, handle, version)
	}

	return err
}

func (pe *ProxyExecutor) Display() {
	var pr *broker.Proxy

	if pe.addedProxy != nil {
		_, _ = fmt.Printf("Adding the following proxy:\n")
		pr = pe.addedProxy
	} else if pe.removedProxy != nil {
		_, _ = fmt.Printf("Removing the following proxy:\n")
		pr = pe.removedProxy
	}

	if pr != nil {
		var detailsMap = map[string]string{}

		detailsMap["host"] = pr.Host
		detailsMap["port"] = cast.ToString(pr.Port)
		detailsMap["tcp"] = cast.ToString(pr.IsTcp())
		detailsMap["udp"] = cast.ToString(pr.IsUdp())
		pe.Ledger.DisplayOptionsWithMap(&detailsMap)
	} else {
		_, _ = fmt.Printf("No action would be taken, no proxy could be found\n")
	}
}

func (pe *ProxyExecutor) Pretend() {
	var configParams *shared.ConfigParams
	var err error

	if configParams, err = pe.makeConfigParams(pe.optionConfigType, pe.optionUserDefined); err == nil {
		var params = pe.makeDataLinkParams(*configParams)
		var writer = pe.dataLinksWriterFactory(pe.Ledger, configParams.GetConfigPath())

		pe.Ledger.DisplayChangeDir()
		pe.editLinkTaskFactory(writer).Pretend(params, pe.Ledger.Logger)
		return
	}

	lang.HandleExit(err)
}

func (pe *ProxyExecutor) Proceed() {
	var configParams *shared.ConfigParams
	var err error

	if configParams, err = pe.makeConfigParams(pe.optionConfigType, pe.optionUserDefined); err == nil {
		var params = pe.makeDataLinkParams(*configParams)
		var writer = pe.dataLinksWriterFactory(pe.Ledger, configParams.GetConfigPath())
		var plan = task.NewPlan("DataLink", pe.Ledger.Logger)

		plan.PrintReportsOnly = true
		task.Single(plan, params, pe.editLinkTaskFactory(writer))
		return
	}

	lang.HandleExit(err)
}

func (pe *ProxyExecutor) Remove(handleArg string, host string, port int) error {
	var writer *DataLinksWriter
	var err error

	pe.setOptions(pe.Ledger, handleArg)

	if writer, err = pe.loadDataLinkWriter(); err == nil {
		var oem = pe.Ledger.GetString(pe.optionOem)
		var handle = pe.Ledger.GetString(pe.optionHandle)
		var version = pe.Ledger.GetString(pe.optionVersion)

		if pe.editedLink = writer.GetDataLink(oem, handle, version); pe.editedLink != nil {
			if pe.removedProxy = pe.editedLink.RemoveProxy(host, port); pe.removedProxy != nil {
				pe.Cli.Exec(pe.Ledger, pe)

			}

			return nil
		}

		return fmt.Errorf("data link [%s/%s:%s] not found", oem, handle, version)
	}

	return err
}

func (pe *ProxyExecutor) loadDataLinkWriter() (*DataLinksWriter, error) {
	var configParams *shared.ConfigParams
	var err error

	if configParams, err = pe.makeConfigParams(pe.optionConfigType, pe.optionUserDefined); err != nil {
		return nil, err
	}

	return pe.dataLinksWriterFactory(pe.Ledger, configParams.GetConfigPath()), nil
}

func (pe *ProxyExecutor) makeDataLinkParams(configParams shared.ConfigParams) *broker.DataLinkParams {
	return &broker.DataLinkParams{
		Broker: broker.Broker{
			AuthFile: pe.Ledger.AuthFile,
		},
		ConfigParams: configParams,
		DataLink:     pe.editedLink,
	}
}

func (pe *ProxyExecutor) newProxy(host string, port int) *broker.Proxy {
	var result = &broker.Proxy{Host: host, Port: port}

	result.SetTcp(pe.Ledger.GetBool(pe.optionTcp))
	result.SetUdp(pe.Ledger.GetBool(pe.optionUdp))
	// always active on Data Links
	result.SetActive(true)
	return result
}

type ProxyOptions struct {
	BaseOptions
	optionTcp *config.BoolOption
	optionUdp *config.BoolOption
}

func (po ProxyOptions) getAddDefiners() []config.Definer {
	return []config.Definer{
		po.optionConfigType,
		po.optionOem,
		po.optionVersion,
		po.optionTcp,
		po.optionUdp,
		po.optionUserDefined,
	}
}

func (po ProxyOptions) getRmDefiners() []config.Definer {
	return []config.Definer{
		po.optionConfigType,
		po.optionOem,
		po.optionVersion,
		po.optionUserDefined,
	}
}

func NewProxy(ledger *config.Ledger, dkCli *Cli) *cobra.Command {
	var addOptions = NewProxyAddOptions()
	var rmOptions = NewProxyRemoveOptions()
	var addCmd = proxy.NewAddProxy(newProxyAddFactory(ledger, dkCli, addOptions))
	var rmCmd = proxy.NewRemoveProxy(newProxyRemoveFactory(ledger, dkCli, rmOptions))
	var proxyCmd = &cobra.Command{
		Use:     "proxy",
		Aliases: []string{"po"},
		Short:   "Manages Outbound Proxies specifications for Data Links",
	}

	ledger.Register(addCmd, addOptions.getAddDefiners()...)
	ledger.Register(rmCmd, rmOptions.getRmDefiners()...)
	proxyCmd.AddCommand(addCmd)
	proxyCmd.AddCommand(rmCmd)
	return proxyCmd
}

func NewProxyExecutor(ctx context.Context, ledger *config.Ledger, sfCli *Cli, options *ProxyOptions) *ProxyExecutor {
	return &ProxyExecutor{
		BaseExecutor: BaseExecutor{
			Cli:     sfCli,
			Ledger:  ledger,
			Context: ctx,
		},
		ProxyOptions: options,

		dataLinksWriterFactory: NewDataLinksWriter,
		editLinkTaskFactory:    broker.NewDataLinkEditTask,
	}
}

func NewProxyAddOptions() *ProxyOptions {
	var udpOption = cli.Options.Proxies.Udp().
		WithKeys(&schema.Genaiz.DataLink.ProxyAdd.Udp).
		BuildBoolOption()

	return &ProxyOptions{
		BaseOptions: BaseOptions{
			optionConfigType: cli.Options.Configs.Type().
				WithKeys(&schema.Genaiz.DataLink.ProxyAdd.ConfigType).
				WithDefaultValue("yaml").
				BuildStringOption(),
			optionHandle: cli.Options.DataLinks.Handle().
				WithKeys(&schema.Genaiz.DataLink.ProxyAdd.Handle).
				WithValidator(config.Validation.Handle).
				BuildStringOption(),
			optionOem: cli.Options.DataLinks.Oem().
				WithKeys(&schema.Genaiz.DataLink.ProxyAdd.Oem).
				WithValidator(config.Validation.Oem).
				BuildStringOption(),
			optionUserDefined: cli.Options.DataLinks.UserDefined().
				WithKeys(&schema.Genaiz.DataLink.ProxyAdd.UserDefined).
				WithDefaultValue("True").
				BuildBoolOption(),
			optionVersion: cli.Options.DataLinks.Version().
				WithKeys(&schema.Genaiz.DataLink.ProxyAdd.Version).
				WithValidator(config.Validation.Version).
				Optional(true).
				BuildStringOption(),
		},
		optionTcp: cli.Options.Proxies.Tcp().
			WithKeys(&schema.Genaiz.DataLink.ProxyAdd.Tcp).
			WithDefaultGetter(func(ledger *config.Ledger) any {
				return !ledger.GetBool(udpOption)
			}).
			BuildBoolOption(),
		optionUdp: udpOption,
	}
}

func NewProxyRemoveOptions() *ProxyOptions {
	return &ProxyOptions{
		BaseOptions: BaseOptions{
			optionConfigType: cli.Options.Configs.Type().
				WithKeys(&schema.Genaiz.DataLink.ProxyRm.ConfigType).
				WithDefaultValue("yaml").
				BuildStringOption(),
			optionHandle: cli.Options.DataLinks.Handle().
				WithKeys(&schema.Genaiz.DataLink.ProxyRm.Handle).
				WithValidator(config.Validation.Handle).
				BuildStringOption(),
			optionOem: cli.Options.DataLinks.Oem().
				WithKeys(&schema.Genaiz.DataLink.ProxyRm.Oem).
				WithValidator(config.Validation.Oem).
				BuildStringOption(),
			optionUserDefined: cli.Options.DataLinks.UserDefined().
				WithKeys(&schema.Genaiz.DataLink.ProxyRm.UserDefined).
				WithDefaultValue("True").
				BuildBoolOption(),
			optionVersion: cli.Options.DataLinks.Version().
				WithKeys(&schema.Genaiz.DataLink.ProxyRm.Version).
				WithValidator(config.Validation.Version).
				Optional(true).
				BuildStringOption(),
		},
	}
}

func newProxyAddFactory(ledger *config.Ledger, sfCli *Cli, options *ProxyOptions) proxy.AddExecutorFactory {
	return func(cmd *cobra.Command) proxy.AddExecutor {
		return NewProxyExecutor(cmd.Context(), ledger, sfCli, options)
	}
}

func newProxyRemoveFactory(ledger *config.Ledger, sfCli *Cli, options *ProxyOptions) proxy.RemoveExecutorFactory {
	return func(cmd *cobra.Command) proxy.RemoveExecutor {
		return NewProxyExecutor(cmd.Context(), ledger, sfCli, options)
	}
}
