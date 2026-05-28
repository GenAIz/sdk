package mgmt

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"

	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz/task/broker"
)

func TestUserAccount_MarshalJSON(t *testing.T) {
	var testExpiry = time.Now().Add(time.Hour).UnixMilli()
	var testCreated = time.Now().UnixMilli()
	var testUserAccount = &UserAccount{
		Active:   true,
		Username: "username",
		HostAddr: "hostAddr",
		Created:  testCreated,
		Expiry:   testExpiry,
	}
	var bytes []byte
	var err error

	if bytes, err = testUserAccount.MarshalJSON(); err == nil {
		assert.NoError(t, err)
		assert.NotEmpty(t, bytes)
		actual := string(bytes)
		assert.Contains(t, actual, fmt.Sprintf("\"active\":%s", "true"))
		assert.Contains(t, actual, fmt.Sprintf("\"username\":\"%s\"", testUserAccount.Username))
		assert.Contains(t, actual, fmt.Sprintf("\"hostAddr\":\"%s\"", testUserAccount.HostAddr))
		assert.Contains(t, actual, fmt.Sprintf("\"created\":\"%s\"", createdFormatter.FormatMillis(testCreated)))
		assert.Contains(t, actual, fmt.Sprintf("\"expiry\":\"%s\"", expiryFormatter.FormatMillis(testExpiry)))
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestUserAccount_MarshalJSON_NoCreated(t *testing.T) {
	var testExpiry = time.Now().UnixMilli()
	var testUserAccount = &UserAccount{
		Active:   true,
		Username: "username",
		HostAddr: "hostAddr",
		Expiry:   testExpiry,
	}
	var bytes []byte
	var err error

	if bytes, err = testUserAccount.MarshalJSON(); err == nil {
		assert.NoError(t, err)
		assert.NotEmpty(t, bytes)
		actual := string(bytes)
		assert.Contains(t, actual, fmt.Sprintf("\"active\":%s", "true"))
		assert.Contains(t, actual, fmt.Sprintf("\"username\":\"%s\"", testUserAccount.Username))
		assert.Contains(t, actual, fmt.Sprintf("\"hostAddr\":\"%s\"", testUserAccount.HostAddr))
		assert.Contains(t, actual, fmt.Sprintf("\"expiry\":\"%s\"", expiryFormatter.FormatMillis(testExpiry)))
		assert.NotContains(t, actual, "created")
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestUserAccount_MarshalSlice(t *testing.T) {
	var testExpiry = time.Now().Add(time.Hour).UnixMilli()
	var testCreated = time.Now().UnixMilli()
	var testUserAccount = &UserAccount{
		Active:   true,
		Username: "username",
		HostAddr: "hostAddr",
		Created:  testCreated,
		Expiry:   testExpiry,
	}
	var values []string
	var err error

	if values, err = testUserAccount.MarshalSlice(); err == nil {
		assert.NoError(t, err)
		assert.NotEmpty(t, values)
		assert.Equal(t, values[0], "yes")
		assert.Equal(t, values[1], testUserAccount.Username)
		assert.Equal(t, values[2], testUserAccount.HostAddr)
		assert.Equal(t, values[3], createdFormatter.FormatMillis(testCreated))
		assert.Equal(t, values[4], expiryFormatter.FormatMillis(testExpiry))
		assert.Equal(t, values[5], "no")
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestUserAccount_MarshalSlice_NoCreated(t *testing.T) {
	var testExpiry = time.Now().Add(time.Hour).UnixMilli()
	var testUserAccount = &UserAccount{
		Active:   true,
		Username: "username",
		HostAddr: "hostAddr",
		Expiry:   testExpiry,
	}
	var values []string
	var err error

	if values, err = testUserAccount.MarshalSlice(); err == nil {
		assert.NoError(t, err)
		assert.NotEmpty(t, values)
		assert.Equal(t, values[0], "yes")
		assert.Equal(t, values[1], testUserAccount.Username)
		assert.Equal(t, values[2], testUserAccount.HostAddr)
		assert.Equal(t, values[3], "-")
		assert.Equal(t, values[4], expiryFormatter.FormatMillis(testExpiry))
		assert.Equal(t, values[5], "no")
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestNewUserAccountFacade_Filtering(t *testing.T) {
	var testAuthFile = filepath.Join(t.TempDir(), ".auth")
	var expectedFilter = "filterHost"
	var expectedToken = "1234567890ABCDEFGHIJ"
	var expectedUsername1 = "user1"
	var expectedUsername2 = "user2"
	var expectedTokenUser = "token"
	var testAuthData = &broker.AuthData{
		Active: 1,
		Accounts: []*broker.AuthAccount{
			{
				HostAddr: expectedFilter,
				AuthSession: &broker.AuthSession{
					Username: expectedUsername1,
					Expiry:   time.Now().Add(24 * time.Hour).UnixMilli(),
				},
			},
			{
				HostAddr: "notHost",
				AuthSession: &broker.AuthSession{
					Username: expectedUsername2,
					Expiry:   -1,
				},
			},
			{
				HostAddr: expectedFilter,
				AuthSession: &broker.AuthSession{
					Username: expectedTokenUser,
					Token:    expectedToken,
					Expiry:   time.Now().Add(24 * time.Hour).UnixMilli(),
				},
			},
		},
	}
	var testParams = &broker.AuthParams{
		Broker: &broker.Broker{
			AuthFile: testAuthFile,
		},
		Expired: false,
	}
	var testLogger = &logrus.Logger{}
	var testProvider = NewUserAccountFacade().
		WithParams(testParams).
		WithLogger(testLogger).
		Filtering(expectedFilter)
	var bytes []byte
	var err error

	if bytes, err = yaml.Marshal(testAuthData); err == nil {
		var fd *os.File

		if fd, err = os.Create(testAuthFile); err == nil {
			defer filez.CloseSilently(fd)

			if _, err = fd.Write(bytes); err == nil {
				var i interface{}

				i, err = testProvider.Get()
				assert.NoError(t, err)
				assert.NotNil(t, i)
				actual := i.([]UserAccount)
				assert.Equal(t, 2, len(actual))
				assert.Equal(t, expectedUsername1, actual[0].Username)
				assert.Equal(t, expectedToken[:10], actual[1].Username)
				return
			}
		}
	}

	assert.NoError(t, err)
}

func TestNewUserAccountFacade_Provider(t *testing.T) {
	var testAuthFile = filepath.Join(t.TempDir(), ".auth")
	var testParams = &broker.AuthParams{
		Broker: &broker.Broker{
			AuthFile: testAuthFile,
		},
	}
	var testLogger = &logrus.Logger{}
	var testProvider = NewUserAccountFacade().
		WithParams(testParams).
		WithLogger(testLogger).
		Provider()

	actual, err := testProvider.Get()
	assert.Error(t, err)
	assert.Equal(t, "No sessions found", err.Error())
	assert.Nil(t, actual)
}
