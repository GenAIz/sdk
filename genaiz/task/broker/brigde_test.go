package broker

import (
	"net/http"
	"testing"
	"time"

	"github.com/spf13/cast"
	"github.com/stretchr/testify/assert"
	"resty.dev/v3"
)

func TestRestyBridge_Close(t *testing.T) {
	var testResty = resty.New()
	var testBridge = &restyBridge{client: testResty}

	assert.NoError(t, testBridge.Close())
}

func TestRestyBridge_Cookie(t *testing.T) {
	var expectedName = "name"
	var expectedValue = "value"
	var testResty = resty.New()
	var testBridge = &restyBridge{
		client:  testResty,
		request: testResty.R(),
	}

	testBridge.Cookie(&http.Cookie{Name: expectedName, Value: expectedValue})
	assert.Equal(t, expectedName, testBridge.request.Cookies[0].Name)
	assert.Equal(t, expectedValue, testBridge.request.Cookies[0].Value)
}

func TestRestyBridge_Get(t *testing.T) {
	var testResty = resty.New()
	var testBridge = &restyBridge{
		client:  testResty,
		request: testResty.R(),
	}

	_, err := testBridge.Get("invalid_protocol")
	assert.Error(t, err)
}

func TestRestyBridge_Json(t *testing.T) {
	var testResty = resty.New()
	var testBridge = &restyBridge{
		client:  testResty,
		request: testResty.R(),
	}

	testBridge.Json()
	assert.Equal(t, "application/json", testBridge.request.ExpectResponseContentType)
}

func TestRestyBridge_Params(t *testing.T) {
	var expectedKey = "key"
	var expectedValue = "value"
	var testParams = map[string]string{expectedKey: expectedValue}
	var testResty = resty.New()
	var testBridge = &restyBridge{
		client:  testResty,
		request: testResty.R(),
	}

	testBridge.Params(testParams)
	assert.Equal(t, expectedValue, testBridge.request.FormData.Get(expectedKey))
}

func TestRestyBridge_Post(t *testing.T) {
	var testResty = resty.New()
	var testBridge = &restyBridge{
		client:  testResty,
		request: testResty.R(),
	}

	_, err := testBridge.Post("invalid_protocol")
	assert.Error(t, err)
}

func TestRestyBridge_Resulting(t *testing.T) {
	var expectedVariable = "variable"
	var testResty = resty.New()
	var testBridge = &restyBridge{
		client:  testResty,
		request: testResty.R(),
	}

	testBridge.Resulting(&expectedVariable)
	assert.Equal(t, expectedVariable, cast.ToString(testBridge.request.Result))
}

func TestRestyBridge_Timeout(t *testing.T) {
	var expectedTimeout = time.Duration(37)
	var testResty = resty.New()
	var testBridge = &restyBridge{
		client:  testResty,
		request: testResty.R(),
	}

	testResty.SetTimeout(expectedTimeout)
	assert.Equal(t, expectedTimeout, testBridge.Timeout())
}
