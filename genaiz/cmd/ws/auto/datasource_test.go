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

type stubUserDataSourceFacade struct {
	filter   string
	logger   *logrus.Logger
	params   *broker.DataInstanceListParams
	provider mgmt.Provider[[]mgmt.UserLinkInstance]
}

func (s *stubUserDataSourceFacade) Filtering(filter string) mgmt.Provider[[]mgmt.UserLinkInstance] {
	s.filter = filter
	return s.provider
}

func (s *stubUserDataSourceFacade) Provider() mgmt.Provider[[]mgmt.UserLinkInstance] {
	return s.provider
}

func (s *stubUserDataSourceFacade) WithLogger(logger *logrus.Logger) mgmt.Facade[[]mgmt.UserLinkInstance, broker.DataInstanceListParams] {
	s.logger = logger
	return s
}

func (s *stubUserDataSourceFacade) WithParams(params *broker.DataInstanceListParams) mgmt.Facade[[]mgmt.UserLinkInstance, broker.DataInstanceListParams] {
	s.params = params
	return s
}

type stubUserDataSourceProvider struct {
	linkInstances []mgmt.UserLinkInstance
	getError      task.Error
}

func (s stubUserDataSourceProvider) Get() ([]mgmt.UserLinkInstance, task.Error) {
	return s.linkInstances, s.getError
}

func TestDataSourceAutoBridge_Bridge(t *testing.T) {
	var expectedId = int64(37)
	var expectedName = "workspaceName"
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		Build()
	var testProvider = stubUserDataSourceProvider{
		linkInstances: []mgmt.UserLinkInstance{
			{
				Id:   new(expectedId),
				Name: expectedName,
			},
		},
	}
	var testAuto = &DataSourceAutoBridge{
		ledger: testLedger,
		dataSourceFacadeProvider: func() mgmt.UserDataSourceFacade {
			return &stubUserDataSourceFacade{
				provider: testProvider,
			}
		},
	}
	var testCompletable = "37"

	if actual, directive := testAuto.Bridge(testCompletable); len(actual) > 0 {
		assert.Equal(t, fmt.Sprintf("%d\t%s", *testProvider.linkInstances[0].Id, testProvider.linkInstances[0].Name), actual[0])
		assert.Equal(t, cobra.ShellCompDirectiveKeepOrder, directive)
	} else {
		assert.Fail(t, "expected actual data source results")
	}
}

func TestDataSourceAutoBridge_Bridge_GetError(t *testing.T) {
	var expectedError = task.NewError("expected")
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		Build()
	var testProvider = stubUserDataSourceProvider{
		getError: expectedError,
	}
	var testAuto = &DataSourceAutoBridge{
		ledger: testLedger,
		dataSourceFacadeProvider: func() mgmt.UserDataSourceFacade {
			return &stubUserDataSourceFacade{
				provider: testProvider,
			}
		},
	}
	var testCompletable = "37"

	if actual, directive := testAuto.Bridge(testCompletable); len(actual) > 0 {
		assert.Fail(t, "did not expect data source results")
	} else {
		assert.Equal(t, cobra.ShellCompDirectiveError, directive)
	}
}

func TestDataSourceAutoBridge_Bridge_NoDataSources(t *testing.T) {
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		Build()
	var testProvider = stubUserDataSourceProvider{}
	var testAuto = &DataSourceAutoBridge{
		ledger: testLedger,
		dataSourceFacadeProvider: func() mgmt.UserDataSourceFacade {
			return &stubUserDataSourceFacade{
				provider: testProvider,
			}
		},
	}
	var testCompletable = "37"

	if actual, directive := testAuto.Bridge(testCompletable); len(actual) > 0 {
		assert.Fail(t, "did not expect data source results")
	} else {
		assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive)
	}
}

func TestNewDataSourceAutoBridge(t *testing.T) {
	var testLedger = config.NewBuilder().WithViper(viper.New()).Build()
	var testBridge = NewDataSourceAutoBridge(testLedger)

	assert.NotNil(t, testBridge)
	assert.Same(t, testLedger, testBridge.ledger)
}
