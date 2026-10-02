package auto

import (
	"fmt"
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

type stubUserDataStoreFacade struct {
	filter   string
	logger   *logrus.Logger
	params   *broker.DataInstanceListParams
	provider mgmt.Provider[[]mgmt.UserLinkInstance]
}

func (s *stubUserDataStoreFacade) Filtering(filter string) mgmt.Provider[[]mgmt.UserLinkInstance] {
	s.filter = filter
	return s.provider
}

func (s *stubUserDataStoreFacade) Provider() mgmt.Provider[[]mgmt.UserLinkInstance] {
	return s.provider
}

func (s *stubUserDataStoreFacade) WithLogger(logger *logrus.Logger) mgmt.Facade[[]mgmt.UserLinkInstance, broker.DataInstanceListParams] {
	s.logger = logger
	return s
}

func (s *stubUserDataStoreFacade) WithParams(params *broker.DataInstanceListParams) mgmt.Facade[[]mgmt.UserLinkInstance, broker.DataInstanceListParams] {
	s.params = params
	return s
}

type stubUserDataStoreProvider struct {
	linkInstances []mgmt.UserLinkInstance
	getError      task.Error
}

func (s stubUserDataStoreProvider) Get() ([]mgmt.UserLinkInstance, task.Error) {
	return s.linkInstances, s.getError
}

func TestDataStoreAutoBridge_Bridge(t *testing.T) {
	var expectedId = int64(37)
	var expectedName = "workspaceName"
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		Build()
	var testProvider = stubUserDataStoreProvider{
		linkInstances: []mgmt.UserLinkInstance{
			{
				Id:   new(expectedId),
				Name: expectedName,
			},
		},
	}
	var testAuto = &DataStoreAutoBridge{
		ledger: testLedger,
		dataStoreFacadeProvider: func() mgmt.UserDataStoreFacade {
			return &stubUserDataStoreFacade{
				provider: testProvider,
			}
		},
	}
	var testCompletable = "37"

	if actual, directive := testAuto.Bridge(testCompletable); len(actual) > 0 {
		assert.Equal(t, fmt.Sprintf("%d\t%s", *testProvider.linkInstances[0].Id, testProvider.linkInstances[0].Name), actual[0])
		assert.Equal(t, cobra.ShellCompDirectiveKeepOrder, directive)
	} else {
		assert.Fail(t, "expected actual data store results")
	}
}

func TestDataStoreAutoBridge_Bridge_GetError(t *testing.T) {
	var expectedError = task.NewError("expected")
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		Build()
	var testProvider = stubUserDataStoreProvider{
		getError: expectedError,
	}
	var testAuto = &DataStoreAutoBridge{
		ledger: testLedger,
		dataStoreFacadeProvider: func() mgmt.UserDataStoreFacade {
			return &stubUserDataStoreFacade{
				provider: testProvider,
			}
		},
	}
	var testCompletable = "37"

	if actual, directive := testAuto.Bridge(testCompletable); len(actual) > 0 {
		assert.Fail(t, "did not expect data store results")
	} else {
		assert.Equal(t, cobra.ShellCompDirectiveError, directive)
	}
}

func TestDataStoreAutoBridge_Bridge_NoDataStores(t *testing.T) {
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		Build()
	var testProvider = stubUserDataStoreProvider{}
	var testAuto = &DataStoreAutoBridge{
		ledger: testLedger,
		dataStoreFacadeProvider: func() mgmt.UserDataStoreFacade {
			return &stubUserDataStoreFacade{
				provider: testProvider,
			}
		},
	}
	var testCompletable = "37"

	if actual, directive := testAuto.Bridge(testCompletable); len(actual) > 0 {
		assert.Fail(t, "did not expect data store results")
	} else {
		assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive)
	}
}

func TestNewDataStoreAutoBridge(t *testing.T) {
	var testLedger = config.NewBuilder().WithViper(viper.New()).Build()
	var testBridge = NewDataStoreAutoBridge(testLedger)

	assert.NotNil(t, testBridge)
	assert.Same(t, testLedger, testBridge.ledger)
}
