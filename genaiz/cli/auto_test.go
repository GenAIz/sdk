package cli

import (
	"path/filepath"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/mgmt"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
)

type stubUserAccountFacade struct {
	filter      string
	getAccounts []mgmt.UserAccount
	getError    task.Error
	logger      *logrus.Logger
	params      *broker.AuthParams
}

func (s *stubUserAccountFacade) Filtering(filter string) mgmt.Provider[[]mgmt.UserAccount] {
	s.filter = filter
	return &stubUserAccountProvider{
		accounts: s.getAccounts,
		err:      s.getError,
	}
}

func (s *stubUserAccountFacade) Provider() mgmt.Provider[[]mgmt.UserAccount] {
	return &stubUserAccountProvider{
		accounts: s.getAccounts,
		err:      s.getError,
	}
}

func (s *stubUserAccountFacade) WithLogger(logger *logrus.Logger) mgmt.Facade[[]mgmt.UserAccount, broker.AuthParams] {
	s.logger = logger
	return s
}

func (s *stubUserAccountFacade) WithParams(params *broker.AuthParams) mgmt.Facade[[]mgmt.UserAccount, broker.AuthParams] {
	s.params = params
	return s
}

type stubUserAccountProvider struct {
	err      task.Error
	accounts []mgmt.UserAccount
}

func (s stubUserAccountProvider) Get() ([]mgmt.UserAccount, task.Error) {
	return s.accounts, s.err
}

func TestBridgeAccounts_Arguments(t *testing.T) {
	var testCmd = &cobra.Command{}
	var testRegistering = AutoBridge.Accounts()
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithUserPath(t.TempDir()).
		WithViper(testViper).
		Build()

	testLedger.InitLogging()
	testRegistering.Arguments(testCmd, testLedger)
	// it won't read any sessions under t.TempDir and therefor will error out
	results, directive := testCmd.ValidArgsFunction(testCmd, []string{}, "")
	assert.Empty(t, results)
	assert.Equal(t, cobra.ShellCompDirectiveError, directive)
}

func TestBridgeAccounts_Option_Error(t *testing.T) {
	var testCmd = &cobra.Command{}
	var testRegistering = AutoBridge.Accounts()
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOption = &config.StringOption{
		Option: config.Option{
			Param: "testParam",
		},
	}

	testLedger.Register(testCmd, testOption)
	testRegistering.Option(testCmd, testLedger, testOption)
	assert.Panics(t, func() { testRegistering.Option(testCmd, testLedger, testOption) })
}

func TestBridgeAccounts_Complete(t *testing.T) {
	var expectedAuthFile = filepath.Join(t.TempDir(), ".auth")
	var expectedFilter = "testFilter"
	var expectedAddr1 = "expectedAddr1"
	var expectedAddr2 = "expectedAddr2"
	var expectedName2 = "expectedName2"
	var testFacade = &stubUserAccountFacade{
		getAccounts: []mgmt.UserAccount{
			{
				HostAddr: expectedAddr1,
			},
			{
				HostAddr: expectedAddr2,
				Name:     expectedName2,
			},
		},
	}
	var testRegistering = &bridgeAccounts{facadeProvider: func() mgmt.UserAccountFacade {
		return testFacade
	}}
	var testCmd = &cobra.Command{}
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOption = &config.StringOption{
		Option: config.Option{
			Param: "testParam",
		},
	}

	testLedger.AuthFile = expectedAuthFile
	testLedger.Register(testCmd, testOption)
	testRegistering.Option(testCmd, testLedger, testOption)
	testFunc, found := testCmd.GetFlagCompletionFunc(testOption.Param)
	assert.True(t, found)
	results, directive := testFunc(testCmd, []string{}, expectedFilter)

	if len(results) == 2 {
		assert.Equal(t, expectedAddr1, results[0])
		assert.Equal(t, expectedAddr2, results[1])
		assert.Equal(t, cobra.ShellCompDirectiveKeepOrder, directive)
		assert.Equal(t, expectedAuthFile, testFacade.params.AuthFile)
		assert.Equal(t, expectedFilter, testFacade.filter)
	} else {
		assert.Fail(t, "missing some results")
	}
}

func TestBridgeAccounts_Complete_Empty(t *testing.T) {
	var expectedAuthFile = filepath.Join(t.TempDir(), ".auth")
	var expectedFilter = "testFilter"
	var testFacade = &stubUserAccountFacade{
		getAccounts: []mgmt.UserAccount{},
	}
	var testRegistering = &bridgeAccounts{facadeProvider: func() mgmt.UserAccountFacade {
		return testFacade
	}}
	var testCmd = &cobra.Command{}
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOption = &config.StringOption{
		Option: config.Option{
			Param: "testParam",
		},
	}

	testLedger.AuthFile = expectedAuthFile
	testLedger.Register(testCmd, testOption)
	testRegistering.Option(testCmd, testLedger, testOption)
	testFunc, found := testCmd.GetFlagCompletionFunc(testOption.Param)
	assert.True(t, found)
	results, directive := testFunc(testCmd, []string{}, expectedFilter)
	assert.Empty(t, results)
	assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive)
	assert.Equal(t, expectedAuthFile, testFacade.params.AuthFile)
	assert.Equal(t, expectedFilter, testFacade.filter)
}

func TestBridgeAccounts_Complete_Error(t *testing.T) {
	var expectedError = task.NewError("expected")
	var expectedAuthFile = filepath.Join(t.TempDir(), ".auth")
	var expectedFilter = "testFilter"
	var testFacade = &stubUserAccountFacade{
		getError: expectedError,
	}
	var testRegistering = &bridgeAccounts{facadeProvider: func() mgmt.UserAccountFacade {
		return testFacade
	}}
	var testCmd = &cobra.Command{}
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOption = &config.StringOption{
		Option: config.Option{
			Param: "testParam",
		},
	}

	testLedger.AuthFile = expectedAuthFile
	testLedger.Register(testCmd, testOption)
	testRegistering.Option(testCmd, testLedger, testOption)
	testFunc, found := testCmd.GetFlagCompletionFunc(testOption.Param)
	assert.True(t, found)
	results, directive := testFunc(testCmd, []string{}, expectedFilter)
	assert.Empty(t, results)
	assert.Equal(t, cobra.ShellCompDirectiveError, directive)
	assert.Equal(t, expectedAuthFile, testFacade.params.AuthFile)
	assert.Equal(t, expectedFilter, testFacade.filter)
}
