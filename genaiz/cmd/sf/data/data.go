package data

import (
	"github.com/spf13/cobra"

	"genaiz.com/genaiz/cmd/sf/data/input"
	"genaiz.com/genaiz/cmd/sf/data/output"
)

type Executor interface {
	input.AddExecutor
	input.RemoveExecutor
	output.AddExecutor
	output.RemoveExecutor
}

type ExecutorFactory func(*cobra.Command) Executor
