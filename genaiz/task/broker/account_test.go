package broker

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"go.yaml.in/yaml/v3"

	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz/task"
)

func TestAccountTasks_ListTaskFactory(t *testing.T) {
	var actual = AccountTasks.ListTaskFactory()

	assert.NotEmpty(t, actual.Name)
	assert.NotEmpty(t, actual.OnPrepare)
	assert.NotEmpty(t, actual.OnComplete)
	assert.Empty(t, actual.OnPretend)
}

func Test_handleAccountListContext(t *testing.T) {
	var testParams = &AccountListParams{
		Broker: Broker{
			AuthFile: filepath.Join(t.TempDir(), ".auth"),
		},
	}
	var testState = &task.State{
		Logger: logrus.New(),
	}
	var testAuth = &AuthData{
		Active: 0,
		Accounts: []*AuthAccount{
			{
				HostAddr: "testHost",
				AuthSession: &AuthSession{
					Username: "testUsername",
				},
			},
		},
	}
	var fd *os.File
	var err error

	if fd, err = os.Create(testParams.AuthFile); err == nil {
		var bytes []byte

		defer filez.CloseSilently(fd)

		if bytes, err = yaml.Marshal(testAuth); err == nil {
			if _, err = fd.Write(bytes); err == nil {
				assert.NoError(t, handleAccountListContext(testParams, testState))
				return
			}
		}
	}

	assert.NoError(t, err)
}

func Test_handleAccountListContext_NoAccountsFound(t *testing.T) {
	var testParams = &AccountListParams{
		Broker: Broker{
			AuthFile: filepath.Join(t.TempDir(), ".auth"),
		},
	}
	var testState = &task.State{
		Logger: logrus.New(),
	}

	assert.ErrorIs(t, handleAccountListContext(testParams, testState), ErrorSessionsEmpty)
}

func Test_handleAccountListComplete(t *testing.T) {
	var testParams = &AccountListParams{
		Broker: Broker{
			AuthFile: filepath.Join(t.TempDir(), ".auth"),
		},
		Expired: false,
	}
	var testState = &task.State{
		Logger: logrus.New(),
	}
	var testAuth = &AuthData{
		Active: 1,
		Accounts: []*AuthAccount{
			{
				HostAddr: "expiredHost",
				AuthSession: &AuthSession{
					Username: "testUsername",
					Expiry:   1,
				},
			},
			{
				HostAddr: "testHost",
				AuthSession: &AuthSession{
					Username: "testUsername2",
					Expiry:   time.Now().Add(time.Hour).UnixMilli(),
				},
			},
		},
	}
	var fd *os.File
	var err error

	if fd, err = os.Create(testParams.AuthFile); err == nil {
		var bytes []byte

		defer filez.CloseSilently(fd)

		if bytes, err = yaml.Marshal(testAuth); err == nil {
			if _, err = fd.Write(bytes); err == nil {
				assert.NoError(t, handleAccountListComplete(testParams, testState))
				assert.NotEqual(t, testAuth, testState.Internal)
				actual, ok := testState.Internal.(*AuthData)
				assert.True(t, ok)
				assert.Equal(t, 0, actual.Active)
				assert.Equal(t, 1, len(actual.Accounts))
				assert.Equal(t, testAuth.Accounts[1].HostAddr, actual.Accounts[0].HostAddr)
				assert.Equal(t, testAuth.Accounts[1].AuthSession.Username, actual.Accounts[0].AuthSession.Username)
				return
			}
		}
	}

	assert.NoError(t, err)
}

func Test_handleAccountListComplete_OutActiveExpired(t *testing.T) {
	var testParams = &AccountListParams{
		Broker: Broker{
			AuthFile: filepath.Join(t.TempDir(), ".auth"),
		},
		Expired: false,
	}
	var testState = &task.State{
		Logger: logrus.New(),
	}
	var testAuth = &AuthData{
		Active: 3,
		Accounts: []*AuthAccount{
			{
				HostAddr: "expiredHost",
				AuthSession: &AuthSession{
					Username: "testUsername",
					Expiry:   1,
				},
			},
			{
				HostAddr: "testHost",
				AuthSession: &AuthSession{
					Username: "testUsername2",
					Expiry:   time.Now().Add(time.Hour).UnixMilli(),
				},
			},
		},
	}
	var fd *os.File
	var err error

	if fd, err = os.Create(testParams.AuthFile); err == nil {
		var bytes []byte

		defer filez.CloseSilently(fd)

		if bytes, err = yaml.Marshal(testAuth); err == nil {
			if _, err = fd.Write(bytes); err == nil {
				assert.NoError(t, handleAccountListComplete(testParams, testState))
				assert.NotEqual(t, testAuth, testState.Internal)
				actual, ok := testState.Internal.(*AuthData)
				assert.True(t, ok)
				assert.Equal(t, 0, actual.Active)
				assert.Equal(t, 1, len(actual.Accounts))
				assert.Equal(t, testAuth.Accounts[1].HostAddr, actual.Accounts[0].HostAddr)
				assert.Equal(t, testAuth.Accounts[1].AuthSession.Username, actual.Accounts[0].AuthSession.Username)
				return
			}
		}
	}

	assert.NoError(t, err)
}

func Test_handleAccountListComplete_NegActiveNoExpired(t *testing.T) {
	var testParams = &AccountListParams{
		Broker: Broker{
			AuthFile: filepath.Join(t.TempDir(), ".auth"),
		},
		Expired: true,
	}
	var testState = &task.State{
		Logger: logrus.New(),
	}
	var testAuth = &AuthData{
		Active: -1,
		Accounts: []*AuthAccount{
			{
				HostAddr: "testHost",
				AuthSession: &AuthSession{
					Username: "testUsername",
				},
			},
		},
	}
	var fd *os.File
	var err error

	if fd, err = os.Create(testParams.AuthFile); err == nil {
		var bytes []byte

		defer filez.CloseSilently(fd)

		if bytes, err = yaml.Marshal(testAuth); err == nil {
			if _, err = fd.Write(bytes); err == nil {
				assert.NoError(t, handleAccountListComplete(testParams, testState))
				assert.NotEqual(t, testAuth, testState.Internal)
				actual, ok := testState.Internal.(*AuthData)
				assert.True(t, ok)
				assert.Equal(t, 0, actual.Active)
				assert.Equal(t, testAuth.Accounts, actual.Accounts)
				return
			}
		}
	}

	assert.NoError(t, err)
}
