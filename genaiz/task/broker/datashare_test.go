package broker

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type stubDataShareClient struct {
	client
	listDataSources      []DataLinkInstance
	listDataSourcesError error
	listDataStores       []DataLinkInstance
	listDataStoresError  error
}

func (dsc stubDataShareClient) ListDataSources() ([]DataLinkInstance, error) {
	return dsc.listDataSources, dsc.listDataSourcesError
}

func (dsc stubDataShareClient) ListDataStores() ([]DataLinkInstance, error) {
	return dsc.listDataStores, dsc.listDataStoresError
}

func TestDataInstanceListParams_getId(t *testing.T) {
	var expectedId = int64(37)
	var testParams = &DataInstanceListParams{}

	assert.Nil(t, testParams.getLinkId())
	testParams.DataLink = &DataLink{
		Id: new(expectedId),
	}
	assert.Equal(t, expectedId, *testParams.getLinkId())
}

func TestDataInstanceListParams_hasDataLinkFilter(t *testing.T) {
	var testParams = &DataInstanceListParams{}

	assert.False(t, testParams.hasDataLinkFilter())
	testParams.DataLink = &DataLink{}
	assert.False(t, testParams.hasDataLinkFilter())
	testParams.DataLink.Oem = "oem"
	assert.True(t, testParams.hasDataLinkFilter())
}

func TestDataInstanceListParams_isVersionEqual(t *testing.T) {
	var testParams = &DataInstanceListParams{}
	var testInstance *DataLinkInstance

	assert.False(t, testParams.isVersionEqual(testInstance))
	testParams.DataLink = &DataLink{}
	assert.False(t, testParams.isVersionEqual(testInstance))
	testInstance = &DataLinkInstance{}
	assert.True(t, testParams.isVersionEqual(testInstance))
	testParams.DataLink.Version = "version"
	assert.False(t, testParams.isVersionEqual(testInstance))
	testInstance = &DataLinkInstance{
		DataLink: DataLink{
			Version: testParams.DataLink.Version,
		},
	}
	assert.True(t, testParams.isVersionEqual(testInstance))
	testParams.DataLink.Seq = new(3)
	assert.False(t, testParams.isVersionEqual(testInstance))
	testInstance = &DataLinkInstance{
		DataLink: DataLink{
			Version: testParams.Version,
			Seq:     testParams.Seq,
		},
	}
	assert.True(t, testParams.isVersionEqual(testInstance))
	testInstance = &DataLinkInstance{
		DataLink: DataLink{
			Version: testParams.Version,
			Seq:     new(4),
		},
	}
	assert.False(t, testParams.isVersionEqual(testInstance))
}

func TestDataInstanceListParams_validateDataLinkFilter(t *testing.T) {
	var testParams = &DataInstanceListParams{
		DataLink: &DataLink{
			Oem:     "oem",
			Handle:  "handle",
			Version: "version",
			Seq:     new(0),
		},
	}

	assert.NoError(t, testParams.validateDataLinkFilter())
}

func TestDataInstanceListParams_validateDataLinkFilter_HandleInvalid(t *testing.T) {
	var testParams = &DataInstanceListParams{
		DataLink: &DataLink{
			Oem:     "oem",
			Version: "version",
			Seq:     new(0),
		},
	}

	assert.ErrorIs(t, testParams.validateDataLinkFilter(), errorDataShareHandleInvalid)
}

func TestDataInstanceListParams_validateDataLinkFilter_InactiveLink(t *testing.T) {
	var testParams = &DataInstanceListParams{
		DataLink: &DataLink{
			Id:      new(int64(37)),
			Oem:     "oem",
			Handle:  "handle",
			Version: "version",
			Seq:     new(0),
			Flags:   new(0),
		},
	}

	assert.ErrorIs(t, testParams.validateDataLinkFilter(), errorDataShareLinkInvalid)
	testParams.Id = nil
	assert.NoError(t, testParams.validateDataLinkFilter())
}

func TestDataInstanceListParams_validateDataLinkFilter_NoLink(t *testing.T) {
	var testParams = &DataInstanceListParams{}

	assert.NoError(t, testParams.validateDataLinkFilter())
}

func TestDataInstanceListParams_validateDataLinkFilter_OemRequired(t *testing.T) {
	var testParams = &DataInstanceListParams{
		DataLink: &DataLink{
			Version: "version",
			Seq:     new(0),
		},
	}

	assert.ErrorIs(t, testParams.validateDataLinkFilter(), errorDataShareOemRequired)
}

func TestDataInstanceListParams_validateDataLinkFilter_VersionInvalid(t *testing.T) {
	var testParams = &DataInstanceListParams{
		DataLink: &DataLink{
			Oem:    "oem",
			Handle: "handle",
			Seq:    new(0),
		},
	}

	assert.ErrorIs(t, testParams.validateDataLinkFilter(), errorDataShareVersionInvalid)
}
