package broker

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/spf13/cast"
	"github.com/stretchr/testify/assert"
	"resty.dev/v3"

	"genaiz.com/genaiz/version/env"
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
	var expected = "expectedGet"
	var testServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, expected)
	}))
	var testResty = resty.New()
	var testBridge = &restyBridge{
		client:  testResty,
		request: testResty.R(),
	}
	defer testServer.Close()

	resp, err := testBridge.Get(testServer.URL)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp)
	assert.Equal(t, expected, string(resp.Bytes()))
}

func TestRestyBridge_Get_Error(t *testing.T) {
	var testResty = resty.New()
	var testBridge = &restyBridge{
		client:  testResty,
		request: testResty.R(),
	}

	_, err := testBridge.Get("invalid_protocol")
	assert.Error(t, err)
}

func TestRestyBridge_Form(t *testing.T) {
	var testResty = resty.New()
	var testBridge = &restyBridge{
		client:  testResty,
		request: testResty.R(),
	}

	testBridge.Form()
	assert.Equal(t, "application/x-www-form-urlencoded", testBridge.request.ExpectResponseContentType)
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
	var expected = "expectedPost"
	var testServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, expected)
	}))
	var testResty = resty.New()
	var testBridge = &restyBridge{
		client:  testResty,
		request: testResty.R(),
	}
	defer testServer.Close()

	resp, err := testBridge.Post(testServer.URL)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp)
	assert.Equal(t, expected, string(resp.Bytes()))
}

func TestRestyBridge_Post_Error(t *testing.T) {
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

func Test_protocolGateChecker(t *testing.T) {
	if env.IsAllowedProtocol("http") {
		assert.NoError(t, protocolGateChecker(&resty.Client{}, &resty.Request{
			URL: "http://somehost",
		}))
	} else {
		assert.Error(t, protocolGateChecker(&resty.Client{}, &resty.Request{
			URL: "http://somehost",
		}))
		assert.NoError(t, protocolGateChecker(&resty.Client{}, &resty.Request{
			URL: "https://somehost",
		}))
	}
}
