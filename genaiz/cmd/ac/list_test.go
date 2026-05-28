package ac

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/mock"
	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/mgmt"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
)

type stubUserAccountFacade struct {
	getAccounts []mgmt.UserAccount
	getError    task.Error
	filter      string
	logger      *logrus.Logger
	params      *broker.AuthParams
}

func (s *stubUserAccountFacade) Filtering(filter string) mgmt.Provider[[]mgmt.UserAccount] {
	s.filter = filter
	return s
}

func (s *stubUserAccountFacade) Get() ([]mgmt.UserAccount, task.Error) {
	return s.getAccounts, s.getError
}

func (s *stubUserAccountFacade) Provider() mgmt.Provider[[]mgmt.UserAccount] {
	return s
}

func (s *stubUserAccountFacade) WithLogger(logger *logrus.Logger) mgmt.Facade[[]mgmt.UserAccount, broker.AuthParams] {
	s.logger = logger
	return s
}

func (s *stubUserAccountFacade) WithParams(params *broker.AuthParams) mgmt.Facade[[]mgmt.UserAccount, broker.AuthParams] {
	s.params = params
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

func TestListExecutor_List(t *testing.T) {
	var expectedFilter = "filter"
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testUserAccountFacade = &stubUserAccountFacade{}
	var testExecutor = &ListExecutor{
		ListOption: NewListOptions(),
		ledger:     testLedger,
		cliPrinterProvider: func() cli.Printer {
			return &stubCliPrinter{}
		},
		userAccountFacadeProvider: func() mgmt.UserAccountFacade {
			return testUserAccountFacade
		},
	}

	assert.NoError(t, testExecutor.List(expectedFilter))
	assert.Equal(t, expectedFilter, testUserAccountFacade.filter)
	assert.Empty(t, testUserAccountFacade.getAccounts)
}

func TestNewListExecutor_ListError(t *testing.T) {
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithUserPath(t.TempDir()).
		WithViper(testViper).
		Build()
	var testList = NewList(testLedger)

	defer patch.Unpatch()
	testLedger.InitLogging()
	testList.Run(testList, []string{"filter"})
	assert.NotEmpty(t, patch.CalledWith)
	assert.EqualValues(t, 1, patch.CalledWith)
}
