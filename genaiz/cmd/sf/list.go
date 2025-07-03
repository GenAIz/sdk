package sf

import (
	"context"

	"github.com/spf13/cobra"

	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/docker"
)

type ListTaskFactory func() *task.Task[docker.BuildParams]

type ListExecutor struct {
	BaseExecutor

	listTaskFactory ListTaskFactory
}

func (le *ListExecutor) list() {
	var buildParams = makeBuildParams(&le.BaseExecutor)
	var plan = task.NewPlan("List", le.Ledger.Logger)

	task.Single(plan, buildParams, le.listTaskFactory())
}

func NewList(ledger *config.Ledger, cli *Cli) *cobra.Command {
	var list = &cobra.Command{
		Use:     "list",
		Short:   "Lists Smart Function instances and containers",
		Long:    "Lists a Smart Function instances and containers by version structurally",
		Example: "genaiz sf list --version=1.0.*",
		Run: func(cmd *cobra.Command, args []string) {
			var exec = NewListExecutor(cmd.Context(), ledger, cli)

			exec.list()
		},
	}

	return list
}

func NewListExecutor(ctx context.Context, ledger *config.Ledger, cli *Cli) *ListExecutor {
	return &ListExecutor{
		BaseExecutor: BaseExecutor{
			Context: ctx,
			Cli:     cli,
			Ledger:  ledger,
		},

		listTaskFactory: docker.NewListTask,
	}
}
