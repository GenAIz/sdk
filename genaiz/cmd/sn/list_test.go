package sn

import (
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
	"genaiz.com/genaiz/lang"
	"genaiz.com/genaiz/mgmt"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
	"genaiz.com/genaiz/task/shared"
)

type stubUserSolutionFacade struct {
	getSolutions []mgmt.UserSolution
	getError     task.Error
	filter       string
	logger       *logrus.Logger
	params       *broker.SolutionListParams
}

func (s *stubUserSolutionFacade) Filtering(filter string) mgmt.Provider[[]mgmt.UserSolution] {
	s.filter = filter
	return s
}

func (s *stubUserSolutionFacade) Get() ([]mgmt.UserSolution, task.Error) {
	return s.getSolutions, s.getError
}

func (s *stubUserSolutionFacade) Provider() mgmt.Provider[[]mgmt.UserSolution] {
	return s
}

func (s *stubUserSolutionFacade) WithLogger(logger *logrus.Logger) mgmt.Facade[[]mgmt.UserSolution, broker.SolutionListParams] {
	s.logger = logger
	return s
}

func (s *stubUserSolutionFacade) WithParams(params *broker.SolutionListParams) mgmt.Facade[[]mgmt.UserSolution, broker.SolutionListParams] {
	s.params = params
	return s
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

func TestListExecutor_List(t *testing.T) {
	var expectedFilter = "filter"
	var expectedId = int64(37)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testUserSolutionFacade = &stubUserSolutionFacade{
		getSolutions: []mgmt.UserSolution{
			{
				Id: expectedId,
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
		userSolutionFacadeProvider: func() mgmt.UserSolutionFacade {
			return testUserSolutionFacade
		},
	}

	testLedger.InitLogging()
	assert.NoError(t, testExecutor.List(expectedFilter))
	assert.Equal(t, expectedFilter, testUserSolutionFacade.filter)
	assert.NotEmpty(t, testUserSolutionFacade.getSolutions)

	if actual, ok := testCliPrinter.printPayload.([]mgmt.UserSolution); ok {
		assert.Equal(t, testUserSolutionFacade.getSolutions, actual)
	} else {
		assert.Fail(t, "did not get actual userSolution values")
	}
}

func TestListExecutor_ListError(t *testing.T) {
	var expectedFilter = "filter"
	var expectedError = task.NewError("expected")
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testUserSolutionFacade = &stubUserSolutionFacade{
		getError: expectedError,
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
		userSolutionFacadeProvider: func() mgmt.UserSolutionFacade {
			return testUserSolutionFacade
		},
	}

	testLedger.InitLogging()
	assert.NoError(t, testExecutor.List(expectedFilter))
	assert.Equal(t, expectedFilter, testUserSolutionFacade.filter)
	assert.ErrorIs(t, testCliPrinter.errorPayload.(error), expectedError)
}

func TestNewList_LocalNoArgs(t *testing.T) {
	var testFolder = t.TempDir()
	var testSolutionFolder = filepath.Join(testFolder, "solution")
	var emptySolutionFolder = filepath.Join(testFolder, "noSolution")
	var expectedSolution = &broker.Solution{
		Id:      lang.Ref(int64(37)),
		Oem:     "expectedOem",
		Handle:  "expectedHandle",
		Version: "expectedVersion",
		Name:    "expectedName",
	}
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		WithUserPath(testFolder).
		WithWorkDir(testFolder).
		Build()
	var testCmd = NewList(testLedger)
	var stdoutRestore = os.Stdout
	var err error

	if err = os.MkdirAll(testSolutionFolder, 0750); err == nil {
		var filename = filepath.Join(testSolutionFolder, testLedger.ConfigName+"."+shared.ConfigTypeYaml)
		var outgoingViper = viper.New()

		outgoingViper.Set("solution", expectedSolution)
		err = outgoingViper.WriteConfigAs(filename)
	}

	if err != nil {
		assert.Fail(t, err.Error())
		return
	}

	if err = os.MkdirAll(emptySolutionFolder, 0750); err == nil {
		var fd *os.File

		fd, err = os.Create(filepath.Join(emptySolutionFolder, testLedger.ConfigName+"."+shared.ConfigTypeYaml))
		defer filez.CloseSilently(fd)
	}

	if err != nil {
		assert.Fail(t, "could not write empty solution")
		return
	}

	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() {
		os.Stdout = stdoutRestore
	}()

	testLedger.InitLogging()
	testCmd.Run(testCmd, []string{})
	_ = w.Close()
	b, _ := io.ReadAll(r)
	output := string(b)
	assert.Contains(t, output, cast.ToString(*expectedSolution.Id))
	assert.Contains(t, output, expectedSolution.GetFqdn())
	assert.Contains(t, output, expectedSolution.GetVersion())
	assert.Contains(t, output, expectedSolution.Name)
}

func TestNewList_NoSession(t *testing.T) {
	var testFolder = t.TempDir()
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		WithUserPath(testFolder).
		WithWorkDir(testFolder).
		Build()
	var testCmd = NewList(testLedger)
	var stdoutRestore = os.Stdout

	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() {
		os.Stdout = stdoutRestore
	}()

	testLedger.InitLogging()
	testCmd.Run(testCmd, []string{"oem/handle"})
	_ = w.Close()
	b, _ := io.ReadAll(r)
	output := string(b)
	assert.Empty(t, output)
}
