package ac

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/mock"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/mgmt"
	"genaiz.com/genaiz/mgmt/accounts"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
)

type stubUserAccountsFactory struct {
	mgmt.ProviderFactory[broker.AccountListParams, []accounts.UserAccount]
}

type stubUserAccountsProvider struct {
	accounts []accounts.UserAccount
	getError task.Error
}

func (s stubUserAccountsProvider) Get() ([]accounts.UserAccount, task.Error) {
	return s.accounts, s.getError
}

func TestListExecutor_List(t *testing.T) {
	var expectedFilter = "filter"
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testFactory = &stubUserAccountsFactory{}
	var testProvider = &stubUserAccountsProvider{}
	var testExecutor = &ListExecutor{
		ListOption: NewListOptions(),

		ledger:        testLedger,
		printerParams: &stubCliPrinterParametric{},
		userAccountsFactory: testFactory.WithProvider(
			func(p accounts.UserAccountsProviderFactory) accounts.UserAccountsProvider {
				return testProvider
			},
		),
	}

	assert.NoError(t, testExecutor.List(expectedFilter))
	assert.Equal(t, &expectedFilter, testFactory.Filter)
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
