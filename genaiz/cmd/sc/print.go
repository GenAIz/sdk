package sc

import (
	"io"
	"os"

	"github.com/spf13/cobra"

	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/lang"
	"genaiz.com/genaiz/schema"
)

type PrintExecutor struct {
	outputFile string
}

func (pe PrintExecutor) PrintSchema() error {
	var out io.Writer
	var fd *os.File
	var err error

	if pe.outputFile == "" {
		out = os.Stdout
	} else if fd, err = os.Create(pe.outputFile); err == nil {
		defer filez.CloseSilently(fd)
		out = fd
	} else {
		return err
	}

	_, err = out.Write(schema.GenaizSchema)
	return err
}

func NewPrint() *cobra.Command {
	var printCmd = &cobra.Command{
		Use:     "print [FILE_PATH]",
		Short:   "Prints the GenAIz JSON schema",
		Long:    "Prints the GenAIz JSON schema for all Genaiz.[yaml|json] files",
		Example: "genaiz sc print myOutputFile.json",
		Args:    cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			var outFile = cli.ArgsOptionalSingle(args)
			var exec = NewPrintExecutor(outFile)
			var err = exec.PrintSchema()

			lang.HandleExit(err)
		},
	}

	return printCmd
}

func NewPrintExecutor(outputFile string) *PrintExecutor {
	return &PrintExecutor{
		outputFile: outputFile,
	}
}
