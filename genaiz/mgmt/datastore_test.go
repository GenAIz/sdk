package mgmt

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz/task/broker"
)

func TestUserDataStoresFacade_Filtering(t *testing.T) {
	var expectedFilter = "filter"
	var testProvider = NewUserDataStoreFacade().
		WithLogger(logrus.New()).
		Filtering(expectedFilter)

	// We can't unit test the provider here
	assert.NotNil(t, testProvider)
}

func TestUserDataStoresFacade_Provider(t *testing.T) {
	var testProvider = NewUserDataStoreFacade().
		WithParams(&broker.DataInstanceListParams{}).
		Provider()

	// We can't unit test the provider here
	assert.NotNil(t, testProvider)
}
