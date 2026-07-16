package dk

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/mgmt"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
	"genaiz.com/genaiz/task/shared"
)

type stubUserDataLinkFacade struct {
	getDatalinks []mgmt.UserDataLink
	getError     task.Error
	filter       string
	logger       *logrus.Logger
	params       *broker.DataLinkListParams
}

func (s *stubUserDataLinkFacade) Filtering(filter string) mgmt.Provider[[]mgmt.UserDataLink] {
	s.filter = filter
	return s
}

func (s *stubUserDataLinkFacade) Get() ([]mgmt.UserDataLink, task.Error) {
	return s.getDatalinks, s.getError
}

func (s *stubUserDataLinkFacade) Provider() mgmt.Provider[[]mgmt.UserDataLink] {
	return s
}

func (s *stubUserDataLinkFacade) WithLogger(logger *logrus.Logger) mgmt.Facade[[]mgmt.UserDataLink, broker.DataLinkListParams] {
	s.logger = logger
	return s
}

func (s *stubUserDataLinkFacade) WithParams(params *broker.DataLinkListParams) mgmt.Facade[[]mgmt.UserDataLink, broker.DataLinkListParams] {
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
	var testUserDatalinkFacade = &stubUserDataLinkFacade{
		getDatalinks: []mgmt.UserDataLink{
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
		userDataLinkFacadeProvider: func() mgmt.UserDataLinkFacade {
			return testUserDatalinkFacade
		},
	}

	testLedger.InitLogging()
	assert.NoError(t, testExecutor.List(expectedFilter))
	assert.Equal(t, expectedFilter, testUserDatalinkFacade.filter)
	assert.NotEmpty(t, testUserDatalinkFacade.getDatalinks)

	if actual, ok := testCliPrinter.printPayload.([]mgmt.UserDataLink); ok {
		assert.Equal(t, testUserDatalinkFacade.getDatalinks, actual)
	} else {
		assert.Fail(t, "did not get actual userDatalink values")
	}
}

func TestListExecutor_ListError(t *testing.T) {
	var expectedFilter = "filter"
	var expectedError = task.NewError("expected")
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testUserSolutionFacade = &stubUserDataLinkFacade{
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
		userDataLinkFacadeProvider: func() mgmt.UserDataLinkFacade {
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
	var expectedDatalink = &broker.DataLink{
		Id:      new(int64(37)),
		Oem:     "expectedOem",
		Handle:  "expectedHandle",
		Version: "expectedVersion",
		Name:    "expectedName",
	}
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		WithUserPath(testFolder).
		Build()
	var testConfigFolder = filepath.Join(testFolder, ".config", "genaiz")
	var filename = filepath.Join(testConfigFolder, testLedger.ConfigName+"."+shared.ConfigTypeYaml)
	var outgoingViper = viper.New()
	var err error

	outgoingViper.Set("datalinks", []broker.DataLink{*expectedDatalink})

	if err = os.MkdirAll(testConfigFolder, 0750); err == nil {
		if err = outgoingViper.WriteConfigAs(filename); err == nil {
			var testCmd = NewList(testLedger)
			var stdoutRestore = os.Stdout

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
			assert.Contains(t, output, cast.ToString(*expectedDatalink.Id))
			assert.Contains(t, output, expectedDatalink.GetFqdn())
			assert.Contains(t, output, expectedDatalink.GetVersion())
			assert.Contains(t, output, expectedDatalink.Name)
			return
		}
	}

	assert.NoError(t, err)
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
