package sf

import (
	"context"
	"fmt"
	"slices"

	"github.com/spf13/cast"
	"github.com/spf13/cobra"

	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/cmd/sf/proxy"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/schema"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
	"genaiz.com/genaiz/task/layout"
	"genaiz.com/genaiz/task/shared"
)

type ProxyExecutor struct {
	BaseExecutor
	*ProxyOptions

	addedProxy     *broker.Proxy
	innerType      *config.StringOption
	innerProxies   *config.Option
	removedProxy   *broker.Proxy
	updatedProxies []broker.Proxy

	initTaskFactory InitTaskFactory
}

func (pe *ProxyExecutor) Add(host string, port int) error {
	var err error

	if err = pe.validateConnector(pe.innerType); err == nil {
		var proxies = pe.Ledger.Get(pe.innerProxies)
		var list = broker.ListProxies(proxies)

		if !slices.ContainsFunc(list, func(pr broker.Proxy) bool {
			return pr.IsEqual(host, port)
		}) {
			pe.addedProxy = pe.makeAddProxy(host, port)
			pe.updatedProxies = append(list, *pe.addedProxy)
			pe.Cli.Exec(pe.Ledger, pe)
			return nil
		}

		return fmt.Errorf("the host [%s] port [%d] is already configured", host, port)
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
		detailsMap["active"] = cast.ToString(pr.IsActive())
		detailsMap["tcp"] = cast.ToString(pr.IsTcp())
		detailsMap["udp"] = cast.ToString(pr.IsUdp())
		pe.Ledger.DisplayOptionsWithMap(&detailsMap)
	}
}

func (pe *ProxyExecutor) Pretend() {
	var params = pe.makeInitParams()
	var builder = makeInitBuilder(pe.Ledger, pe.Cli)

	builder.WithOutboundProxyRemoved(pe.removedProxy)
	pe.initTaskFactory(builder).Pretend(params, pe.Ledger.Logger)
}

func (pe *ProxyExecutor) Proceed() {
	var params = pe.makeInitParams()
	var builder = makeInitBuilder(pe.Ledger, pe.Cli)
	var plan = task.NewPlan("OutboundProxy", pe.Ledger.Logger)

	plan.PrintReportsOnly = true
	builder.WithOutboundProxyRemoved(pe.removedProxy)
	task.Single(plan, params, pe.initTaskFactory(builder))
}

func (pe *ProxyExecutor) Remove(host string, port int) {
	var proxies = pe.Ledger.Get(pe.innerProxies)
	var list = broker.ListProxies(proxies)
	var updated []broker.Proxy

	if i := slices.IndexFunc(list, func(pr broker.Proxy) bool {
		return pr.IsEqual(host, port)
	}); i >= 0 {
		pe.removedProxy = &list[i]

		for j, pr := range list {
			if j != i {
				updated = append(updated, pr)
			}
		}

		pe.updatedProxies = updated
		pe.Cli.Exec(pe.Ledger, pe)
	}
}

func (pe *ProxyExecutor) makeAddProxy(host string, port int) *broker.Proxy {
	var result = &broker.Proxy{
		Host: host,
		Port: port,
	}

	result.SetActive(!pe.Ledger.GetBool(pe.optionInactive))
	result.SetTcp(pe.Ledger.GetBool(pe.optionTcp))
	result.SetUdp(pe.Ledger.GetBool(pe.optionUdp))
	return result
}

func (pe *ProxyExecutor) makeInitParams() *layout.InitParams {
	return &layout.InitParams{
		CreateParams: layout.CreateParams{
			ConfigParams: shared.ConfigParams{
				ConfigName: pe.Ledger.ConfigName,
			},
		},
		OutboundProxies: pe.updatedProxies,
	}
}

type ProxyOptions struct {
	optionInactive *config.BoolOption
	optionTcp      *config.BoolOption
	optionUdp      *config.BoolOption
}

func (po ProxyOptions) allDefiners() []config.Definer {
	return []config.Definer{
		po.optionInactive,
		po.optionTcp,
		po.optionUdp,
	}
}

func NewProxy(ledger *config.Ledger, sfCli *Cli) *cobra.Command {
	var addOptions = NewProxyAddOptions()
	var addCommand = proxy.NewAddProxy(newProxyAddFactory(ledger, sfCli, addOptions))
	var rmCommand = proxy.NewRemoveProxy(newProxyRemoveFactory(ledger, sfCli))
	var proxyCmd = &cobra.Command{
		Use:     "proxy",
		Aliases: []string{"pr"},
		Short:   "Manages outbound proxy configurations for Smart Functions",
	}

	ledger.Register(addCommand, addOptions.allDefiners()...)
	proxyCmd.AddCommand(addCommand)
	proxyCmd.AddCommand(rmCommand)
	return proxyCmd
}

func NewProxyAddOptions() *ProxyOptions {
	var udpOption = cli.Options.Proxies.Udp().
		WithKeys(&schema.Genaiz.Function.Publish.OutboundProxyAdd.Udp).
		BuildBoolOption()

	return &ProxyOptions{
		optionInactive: cli.Options.Proxies.Inactive().
			WithKeys(&schema.Genaiz.Function.Publish.OutboundProxyAdd.Inactive).
			BuildBoolOption(),
		optionTcp: cli.Options.Proxies.Tcp().
			WithKeys(&schema.Genaiz.Function.Publish.OutboundProxyAdd.Tcp).
			WithDefaultGetter(func(ledger *config.Ledger) any {
				return !ledger.GetBool(udpOption)
			}).
			BuildBoolOption(),
		optionUdp: udpOption,
	}
}

func NewProxyExecutor(ctx context.Context, ledger *config.Ledger, sfCli *Cli, options *ProxyOptions) *ProxyExecutor {
	return &ProxyExecutor{
		BaseExecutor: BaseExecutor{
			Cli:     sfCli,
			Ledger:  ledger,
			Context: ctx,
		},
		ProxyOptions: options,

		innerType: cli.Options.Functions.Type().
			WithKeys(&schema.Genaiz.Function.Publish.Type).
			BuildStringOption(),
		innerProxies: cli.NewOptionBuilder().
			WithKeys(&schema.Genaiz.Function.Publish.OutboundProxies).
			BuildOption(),

		initTaskFactory: layout.NewInitTask,
	}
}

func newProxyAddFactory(ledger *config.Ledger, sfCli *Cli, options *ProxyOptions) proxy.AddExecutorFactory {
	return func(cmd *cobra.Command) proxy.AddExecutor {
		return NewProxyExecutor(cmd.Context(), ledger, sfCli, options)
	}
}

func newProxyRemoveFactory(ledger *config.Ledger, sfCli *Cli) proxy.RemoveExecutorFactory {
	return func(cmd *cobra.Command) proxy.RemoveExecutor {
		return NewProxyExecutor(cmd.Context(), ledger, sfCli, &ProxyOptions{})
	}
}
