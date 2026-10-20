package mgmt

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/lang/timez"
	"genaiz.com/genaiz/task"
)

type stubStringProvider struct {
	getError task.Error
	getValue string
}

func (s stubStringProvider) Get() (string, task.Error) {
	return s.getValue, s.getError
}

func TestProviderFactory_WithFilter(t *testing.T) {
	var testFactory = &ProviderFactory[string, string]{}
	var expectedFilter = "filter"

	testFactory.WithFilter("")
	assert.Empty(t, testFactory.Filter)
	testFactory.WithFilter(expectedFilter)
	assert.Equal(t, &expectedFilter, testFactory.Filter)
}

func TestProviderFactory_WithLogger(t *testing.T) {
	var expectedLogger = logrus.New()
	var testFactory = &ProviderFactory[string, string]{}

	testFactory.WithLogger(expectedLogger)
	assert.Same(t, expectedLogger, testFactory.Logger)
}

func TestProviderFactory_WithParams(t *testing.T) {
	var testFactory = &ProviderFactory[string, string]{}
	var expectedParams = "filter"

	testFactory.WithParams(&expectedParams)
	assert.Equal(t, &expectedParams, testFactory.Params)
}

func TestProviderFactory_WithProvider(t *testing.T) {
	var expectedValue = "value"
	var testFactory = &ProviderFactory[string, string]{}
	var testProvider = func(pf ProviderFactory[string, string]) Provider[string] {
		return &stubStringProvider{
			getValue: expectedValue,
		}
	}

	testFactory.WithProvider(testProvider)

	if actual, err := testFactory.Provider().Get(); err == nil {
		assert.Equal(t, expectedValue, actual)
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestProviderFactory_WithTimeFormat(t *testing.T) {
	var testFactory = &ProviderFactory[string, string]{}
	var expectedFormatter = timez.NewMidnightFormatter()

	testFactory.WithTimeFormat(expectedFormatter)
	assert.Same(t, expectedFormatter, testFactory.Format)
}
