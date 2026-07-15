package wf

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/mgmt"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
	"genaiz.com/genaiz/task/shared"
)

type stubUserWorkflowFacade struct {
	getWorkflows []mgmt.UserWorkflow
	getError     task.Error
	filter       string
	logger       *logrus.Logger
	params       *broker.WorkflowListParams
	path         string
	pathGraphers map[string]broker.SolutionGrapher
}

func (s *stubUserWorkflowFacade) Filtering(filter string) mgmt.Provider[[]mgmt.UserWorkflow] {
	s.filter = filter
	return s
}

func (s *stubUserWorkflowFacade) Get() ([]mgmt.UserWorkflow, task.Error) {
	return s.getWorkflows, s.getError
}

func (s *stubUserWorkflowFacade) Provider() mgmt.Provider[[]mgmt.UserWorkflow] {
	return s
}

func (s *stubUserWorkflowFacade) WithLogger(logger *logrus.Logger) mgmt.Facade[[]mgmt.UserWorkflow, broker.WorkflowListParams] {
	s.logger = logger
	return s
}

func (s *stubUserWorkflowFacade) WithParams(params *broker.WorkflowListParams) mgmt.Facade[[]mgmt.UserWorkflow, broker.WorkflowListParams] {
	s.params = params
	return s
}

func (s *stubUserWorkflowFacade) WithPathGraphers(path string, pathGraphers map[string]broker.SolutionGrapher) mgmt.UserWorkflowFacade {
	s.path = path
	s.pathGraphers = pathGraphers
	return s
}

type stubCliPrinter struct {
	errorError   error
	errorPayload interface{}
	printError   error
	printPayload interface{}
}

func (s *stubCliPrinter) Error(i interface{}) error {
	s.errorPayload = i
	return s.errorError
}

func (s *stubCliPrinter) Print(i interface{}) error {
	s.printPayload = i
	return s.printError
}

type stubAccountParametric struct {
	brokerParams *broker.Broker
}

func (s stubAccountParametric) BrokerParams() *broker.Broker {
	return s.brokerParams
}

type stubCliPrinterParametric struct {
	cliPrinter cli.Printer
}

func (s stubCliPrinterParametric) Printer() cli.Printer {
	return s.cliPrinter
}

func (s stubCliPrinterParametric) IsDefault() bool {
	return true
}

func TestListExecutor_List(t *testing.T) {
	var expectedArg = "oem/handle:1.0.0"
	var expectedId = int64(37)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testUserWorkflowFacade = &stubUserWorkflowFacade{
		getWorkflows: []mgmt.UserWorkflow{
			{
				Id: &expectedId,
			},
		},
	}
	var testCliPrinter = &stubCliPrinter{}
	var testExecutor = &ListExecutor{
		ListOptions: NewListOptions(),

		ledger: testLedger,
		accountParams: &stubAccountParametric{
			brokerParams: &broker.Broker{
				HostAddr: "testAddr",
			},
		},
		printerParams: &stubCliPrinterParametric{
			cliPrinter: testCliPrinter,
		},
		userWorkflowFacadeProvider: func() mgmt.UserWorkflowFacade {
			return testUserWorkflowFacade
		},
	}

	testLedger.InitLogging()
	assert.NoError(t, testExecutor.List(expectedArg))
	assert.Empty(t, testUserWorkflowFacade.filter)

	if actual, ok := testCliPrinter.printPayload.([]mgmt.UserWorkflow); ok {
		assert.Equal(t, testUserWorkflowFacade.getWorkflows, actual)
	} else {
		assert.Fail(t, "did not get actual userWorkflow values")
	}
}

func TestListExecutor_List_FqdnError(t *testing.T) {
	var expectedArg = "handle:version"
	var expectedId = int64(37)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testUserWorkflowFacade = &stubUserWorkflowFacade{
		getWorkflows: []mgmt.UserWorkflow{
			{
				Id: &expectedId,
			},
		},
	}
	var testCliPrinter = &stubCliPrinter{
		errorError: errors.New("expected"),
	}
	var testExecutor = &ListExecutor{
		ListOptions: NewListOptions(),

		ledger: testLedger,
		accountParams: &stubAccountParametric{
			brokerParams: &broker.Broker{
				HostAddr: "testAddr",
			},
		},
		printerParams: &stubCliPrinterParametric{
			cliPrinter: testCliPrinter,
		},
		userWorkflowFacadeProvider: func() mgmt.UserWorkflowFacade {
			return testUserWorkflowFacade
		},
	}

	testLedger.InitLogging()
	assert.ErrorIs(t, testExecutor.List(expectedArg), testCliPrinter.errorError)
}

func TestListExecutor_List_ListError(t *testing.T) {
	var expectedArg = "oem/handle:1.0.0"
	var expectedId = int64(37)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testUserWorkflowFacade = &stubUserWorkflowFacade{
		getWorkflows: []mgmt.UserWorkflow{
			{
				Id: &expectedId,
			},
		},
		getError: task.NewError("expected"),
	}
	var testCliPrinter = &stubCliPrinter{
		errorError: testUserWorkflowFacade.getError,
	}
	var testExecutor = &ListExecutor{
		ListOptions: NewListOptions(),

		ledger: testLedger,
		accountParams: &stubAccountParametric{
			brokerParams: &broker.Broker{
				HostAddr: "testAddr",
			},
		},
		printerParams: &stubCliPrinterParametric{
			cliPrinter: testCliPrinter,
		},
		userWorkflowFacadeProvider: func() mgmt.UserWorkflowFacade {
			return testUserWorkflowFacade
		},
	}

	testLedger.InitLogging()
	assert.ErrorIs(t, testExecutor.List(expectedArg), testUserWorkflowFacade.getError)
}

func TestListExecutor_List_PathError(t *testing.T) {
	var expectedArg = filepath.Join(t.TempDir(), "notExist")
	var expectedId = int64(37)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testUserWorkflowFacade = &stubUserWorkflowFacade{
		getWorkflows: []mgmt.UserWorkflow{
			{
				Id: &expectedId,
			},
		},
	}
	var testCliPrinter = &stubCliPrinter{
		errorError: errors.New("expected"),
	}
	var testExecutor = &ListExecutor{
		ListOptions: NewListOptions(),

		ledger: testLedger,
		accountParams: &stubAccountParametric{
			brokerParams: &broker.Broker{
				HostAddr: "testAddr",
			},
		},
		printerParams: &stubCliPrinterParametric{
			cliPrinter: testCliPrinter,
		},
		userWorkflowFacadeProvider: func() mgmt.UserWorkflowFacade {
			return testUserWorkflowFacade
		},
	}

	testLedger.InitLogging()
	assert.ErrorIs(t, testExecutor.List(expectedArg), testCliPrinter.errorError)
}

func TestNewList_LocalNoArgs(t *testing.T) {
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithWorkDir(t.TempDir()).
		WithViper(testViper).
		Build()
	var testCmd = NewList(testLedger)
	var fd *os.File
	var err error

	if fd, err = os.Create(filepath.Join(testLedger.WorkDir, testLedger.ConfigName+"."+shared.ConfigTypeYaml)); err == nil {
		var subSolution = filepath.Join(testLedger.WorkDir, "sub")

		defer filez.CloseSilently(fd)

		if err = os.MkdirAll(subSolution, 0750); err == nil {
			var subSolutionFile = filepath.Join(subSolution, testLedger.ConfigName+"."+shared.ConfigTypeYaml)
			var expectedSolution = &broker.Solution{
				Handle: "testHandle",
				Workflows: []broker.Workflow{
					{
						Id: new(int64(37)),
					},
				},
			}
			var v = viper.New()

			v.Set("solution", expectedSolution)

			if err = v.WriteConfigAs(subSolutionFile); err == nil {
				var stdoutRestore = os.Stdout

				r, w, _ := os.Pipe()
				os.Stdout = w
				defer func() {
					os.Stdout = stdoutRestore
				}()

				testLedger.InitLogging()
				t.Chdir(testLedger.WorkDir)
				testCmd.Run(testCmd, []string{})

				_ = w.Close()
				b, _ := io.ReadAll(r)
				output := string(b)
				assert.Contains(t, output, cast.ToString(expectedSolution.Handle))
				return
			}
		}
	}

	assert.NoError(t, err)
}
