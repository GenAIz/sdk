package ws

import (
	"context"
	"strconv"

	"github.com/spf13/cobra"

	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/cmd/ws/args"
	"genaiz.com/genaiz/cmd/ws/data"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/lang"
	"genaiz.com/genaiz/mgmt/workspaces"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
)

type DataInstanceExecutor struct {
	BaseExecutor
	*DataInstanceOptions

	accountParams config.AccountParametric
	printerParams cli.PrinterParametric

	nodeArguments          args.NodeResolution
	nodeResolveTaskFactory broker.WorkspaceNodeResolveTaskFactory
}

type DataInstanceOptions struct {
	optionAccount     *config.StringOption
	optionJsonPrinter *config.BoolOption
}

func (dio DataInstanceOptions) allDefiners() []config.Definer {
	return []config.Definer{
		dio.optionAccount,
		dio.optionJsonPrinter,
	}
}

type DataListExecutor struct {
	BaseExecutor
	*data.ListOptions

	accountParams config.AccountParametric
	printerParams cli.PrinterParametric

	nodeArguments args.NodeResolution

	listNodeSetTaskFactory       broker.WorkspaceNodeSetListTaskFactory
	resolveNodeTaskFactory       broker.WorkspaceNodeResolveTaskFactory
	workspaceNodeDataSetsFactory workspaces.UserWorkspaceDataSetsFactory
}

func (dle *DataListExecutor) Display() {
	var optionMap = dle.nodeArguments.MakeOptionMap()

	dle.Ledger.DisplayOptionsWithMap(&optionMap, dle.Options()...)
}

func (dle *DataListExecutor) List(ws, fl, nd string) error {
	dle.nodeArguments = args.NewNodeResolution(ws, fl, nd)
	dle.Cli.Exec(dle.Ledger, dle)
	return nil
}

func (dle *DataListExecutor) Pretend() {
	var brokerParams = dle.accountParams.BrokerParams()
	var resolveNodeParams = dle.nodeArguments.ResolveNodeParams(brokerParams)
	var listParams = dle.ListParams(brokerParams, resolveNodeParams)
	var plan = task.NewPlan("DataSetList", dle.Ledger.Logger)

	plan.Sequence(task.NewPretender(resolveNodeParams, dle.resolveNodeTaskFactory()),
		task.NewPretender(listParams, dle.listNodeSetTaskFactory()))
}

func (dle *DataListExecutor) Proceed() {
	var brokerParams = dle.accountParams.BrokerParams()
	var resolveNodeParams = dle.nodeArguments.ResolveNodeParams(brokerParams)
	var printer = dle.printerParams.Printer()
	var workspaceList, err = dle.workspaceNodeDataSetsFactory.
		WithLogger(dle.Ledger.Logger).
		WithParams(resolveNodeParams).
		Provider().
		Get()

	if err == nil {
		lang.HandleExit(printer.Print(workspaceList))
		return
	}

	lang.HandleExit(printer.Error(err))
}

type DataSetExecutor struct {
	BaseExecutor
	*data.SetOptions
	path string
	port string

	accountParams config.AccountParametric
	printerParams cli.PrinterParametric

	nodeArguments args.NodeResolution

	createDataSetTaskFactory  broker.DataSetCreateTaskFactory
	listDataFilesTaskFactory  broker.DataSetListFilesTaskFactory
	resolveDataSetTaskFactory broker.WorkspaceNodeSetResolveTaskFactory
	resolveNodeTaskFactory    broker.WorkspaceNodeResolveTaskFactory
	updateNodeSetTaskFactory  broker.WorkspaceNodeSetUpdateTaskFactory
	uploadDataSetTaskFactory  broker.DataSetUploadTaskFactory
}

func (dse *DataSetExecutor) Display() {
	var optionMap = dse.nodeArguments.MakeOptionMap()

	if _, err := strconv.ParseInt(dse.port, 10, 64); err == nil {
		optionMap["data set id"] = dse.port
	} else {
		optionMap["input port"] = dse.port
	}

	optionMap["file path"] = dse.path
	dse.Ledger.DisplayOptionsWithMap(&optionMap, dse.Options()...)
}

func (dse *DataSetExecutor) Pretend() {
	var brokerParams = dse.accountParams.BrokerParams()
	var resolveNodeParams = dse.nodeArguments.ResolveNodeParams(brokerParams)
	var resolveSetParams = dse.ResolveSetParams(dse.port, brokerParams, resolveNodeParams)
	var createSetParams = dse.CreateSetParams(dse.Ledger, brokerParams, resolveNodeParams, resolveSetParams)
	var listFilesParams = dse.ListFilesParams(brokerParams, resolveSetParams)
	var uploadFileParams = dse.UploadFileParams(dse.Ledger, dse.path, brokerParams, resolveSetParams)
	var plan = task.NewPlan("DataSetAdd", dse.Ledger.Logger)
	var pretenders = []task.Worker{
		task.NewPretender(resolveNodeParams, dse.resolveNodeTaskFactory()),
		task.NewPretender(resolveSetParams, dse.resolveDataSetTaskFactory()),
		task.NewPretender(createSetParams, dse.createDataSetTaskFactory()),
	}

	if !dse.IsOverwrite(dse.Ledger) {
		pretenders = append(pretenders, task.NewPretender(listFilesParams, dse.listDataFilesTaskFactory()))
	}

	pretenders = append(pretenders, task.NewPretender(uploadFileParams, dse.uploadDataSetTaskFactory()))
	pretenders = append(pretenders, task.NewPretender(resolveSetParams, dse.updateNodeSetTaskFactory()))
	plan.Sequence(pretenders...)
}

func (dse *DataSetExecutor) Proceed() {
	var brokerParams = dse.accountParams.BrokerParams()
	var resolveNodeParams = dse.nodeArguments.ResolveNodeParams(brokerParams)
	var resolveSetParams = dse.ResolveSetParams(dse.port, brokerParams, resolveNodeParams)
	var createSetParams = dse.CreateSetParams(dse.Ledger, brokerParams, resolveNodeParams, resolveSetParams)
	var listFilesParams = dse.ListFilesParams(brokerParams, resolveSetParams)
	var uploadFileParams = dse.UploadFileParams(dse.Ledger, dse.path, brokerParams, resolveSetParams)
	var workers = []task.Worker{
		task.NewWorker(resolveNodeParams, dse.resolveNodeTaskFactory()),
		task.NewWorker(resolveSetParams, dse.resolveDataSetTaskFactory()),
		task.NewWorker(createSetParams, dse.createDataSetTaskFactory()),
	}
	var plan *task.Plan

	if dse.printerParams.IsDefault() {
		plan = task.NewPlan("DataSetAdd", dse.Ledger.Logger)
		plan.PrintReportsOnly = true
	} else {
		var printer = dse.printerParams.Printer()

		plan = task.NewPlanBuilder(dse.Ledger.Logger).
			WithReturn(cli.HandlePrint(printer)).
			WithFailures(cli.HandleError(printer)).
			Build()
	}

	if !dse.IsOverwrite(dse.Ledger) {
		workers = append(workers, task.NewWorker(listFilesParams, dse.listDataFilesTaskFactory()))
	}

	workers = append(workers, task.NewWorker(uploadFileParams, dse.uploadDataSetTaskFactory()))
	workers = append(workers, task.NewWorker(resolveSetParams, dse.updateNodeSetTaskFactory()))
	plan.Sequence(workers...)
}

func (dse *DataSetExecutor) Set(ws, fl, nd, port, path string) error {
	var err error

	if err = filez.IsReadable(path); err == nil {
		dse.nodeArguments = args.NewNodeResolution(ws, fl, nd)
		dse.port = port
		dse.path = path
		dse.Cli.Exec(dse.Ledger, dse)
		return nil
	}

	return err
}

func NewData(ledger *config.Ledger, wsCli *Cli) *cobra.Command {
	var listOptions = data.NewDataListOptions()
	var setOptions = data.NewDataSetOptions()
	var listFactory = newDataListExecutorFactory(ledger, wsCli, listOptions)
	var setFactory = newDataSetExecutorFactory(ledger, wsCli, setOptions)
	var dataCmd = &cobra.Command{
		Use:     "data",
		Aliases: []string{"dt"},
		Short:   "Manages data sources, stores and sets for nodes under workspace flows",
	}

	dataCmd.AddCommand(data.NewDataList(ledger, listOptions, listFactory))
	dataCmd.AddCommand(data.NewDataSet(ledger, setOptions, setFactory))
	dataCmd.AddCommand(NewDataSource(ledger, wsCli))
	dataCmd.AddCommand(NewDataStore(ledger, wsCli))
	return dataCmd
}

func NewDataListExecutor(ctx context.Context, ledger *config.Ledger, wsCli *Cli, options *data.ListOptions) *DataListExecutor {
	return &DataListExecutor{
		BaseExecutor: BaseExecutor{
			Cli:     wsCli,
			Context: ctx,
			Ledger:  ledger,
		},
		ListOptions: options,

		accountParams: options.AccountParams(ledger),
		printerParams: options.PrinterParams(ledger),

		listNodeSetTaskFactory:       broker.WorkspaceTasks.ListSetNodeTaskFactory,
		resolveNodeTaskFactory:       broker.WorkspaceTasks.ResolveNodeTaskFactory,
		workspaceNodeDataSetsFactory: workspaces.NewUserWorkspaceDataSetsFactory(),
	}
}

func NewDataSetExecutor(ctx context.Context, ledger *config.Ledger, wsCli *Cli, options *data.SetOptions) *DataSetExecutor {
	return &DataSetExecutor{
		BaseExecutor: BaseExecutor{
			Cli:     wsCli,
			Context: ctx,
			Ledger:  ledger,
		},
		SetOptions: options,

		accountParams: options.AccountParams(ledger),
		printerParams: options.PrinterParams(ledger),

		createDataSetTaskFactory:  broker.DataSetTasks.CreateTaskFactory,
		listDataFilesTaskFactory:  broker.DataSetTasks.ListFilesTaskFactory,
		resolveDataSetTaskFactory: broker.WorkspaceTasks.ResolveSetNodeTaskFactory,
		resolveNodeTaskFactory:    broker.WorkspaceTasks.ResolveNodeTaskFactory,
		updateNodeSetTaskFactory:  broker.WorkspaceTasks.UpdateSetNodeTaskFactory,
		uploadDataSetTaskFactory:  broker.DataSetTasks.UploadTaskFactory,
	}
}

func newDataListExecutorFactory(ledger *config.Ledger, wsCli *Cli, options *data.ListOptions) data.ListExecutorFactory {
	return func(cmd *cobra.Command) data.ListExecutor {
		return NewDataListExecutor(cmd.Context(), ledger, wsCli, options)
	}
}

func newDataSetExecutorFactory(ledger *config.Ledger, wsCli *Cli, options *data.SetOptions) data.SetExecutorFactory {
	return func(cmd *cobra.Command) data.SetExecutor {
		return NewDataSetExecutor(cmd.Context(), ledger, wsCli, options)
	}
}
