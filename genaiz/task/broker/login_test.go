package broker

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"

	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz-lib/lang/panicz"
	"genaiz.com/genaiz/task"
)

type stubLoginClient struct {
	client
	session   *AuthSession
	loginErr  error
	logoutErr error

	loginUsername string
	loginPassword []byte
	logoutId      string
}

func (s *stubLoginClient) Login(username string, password []byte) (*AuthSession, error) {
	s.loginUsername = username
	s.loginPassword = password
	return s.session, s.loginErr
}

func (s *stubLoginClient) Logout(id string) error {
	s.logoutId = id
	return s.logoutErr
}

func (s *stubLoginClient) Session() (*Session, error) {
	if s.session == nil {
		return nil, errors.New("session error")
	}

	return &Session{
		UserId: s.UserId,
	}, nil
}

func (s *stubLoginClient) SessionValid(session *Session) bool {
	if session.UserId > 0 {
		return s.client.SessionValid(session)

	}

	return true
}

func TestAuthData_Find_NoBrokerSession(t *testing.T) {
	var authData = &AuthData{}
	var _, err = authData.Find("hostAddr")

	assert.ErrorIs(t, ErrorNoSession, err)
}

func TestAuthData_Find(t *testing.T) {
	var expectedAddr = "hostAddr"
	var authData = &AuthData{
		Accounts: []*AuthAccount{
			{
				HostAddr: expectedAddr,
			},
		},
	}
	var actual, err = authData.Find(expectedAddr)

	assert.NoError(t, err)
	assert.Same(t, authData.Accounts[0], actual)
}

func TestAuthData_ForHostUser_NoHostNoSession(t *testing.T) {
	var expectedUser = "username"
	var authData = &AuthData{}
	var _, err = authData.ForHostUser("", expectedUser)

	assert.ErrorIs(t, ErrorNoSession, err)
}

func TestAuthData_ForHostUser_NoHostNoUser(t *testing.T) {
	var authData = &AuthData{}
	var _, err = authData.ForHostUser("", "")

	assert.ErrorIs(t, ErrorNoSession, err)
}

func TestAuthData_ForHostUser_NoHost(t *testing.T) {
	var expectedUser = "username"
	var authData = &AuthData{
		Accounts: []*AuthAccount{
			{
				HostAddr: "hostAddr",
				AuthSession: &AuthSession{
					Username: expectedUser,
				},
			},
		},
	}
	var actual, err = authData.ForHostUser("", expectedUser)

	assert.NoError(t, err)
	assert.Same(t, authData.Accounts[0], actual)
}

func TestAuthData_ForHostUser(t *testing.T) {
	var expectedUser = "username"
	var expectedHost = "hostAddr"
	var authData = &AuthData{
		Accounts: []*AuthAccount{
			{
				HostAddr: "hostAddr2",
				AuthSession: &AuthSession{
					Username: expectedUser,
				},
			},
			{
				HostAddr: expectedHost,
				AuthSession: &AuthSession{
					Username: expectedUser,
				},
			},
		},
	}
	var actual, err = authData.ForHostUser(expectedHost, expectedUser)

	assert.NoError(t, err)
	assert.Same(t, authData.Accounts[1], actual)
}

func TestAuthData_ForToken_NoSession(t *testing.T) {
	var expectedToken = "token"
	var authData = &AuthData{}
	var _, err = authData.ForToken(expectedToken)

	assert.ErrorIs(t, ErrorNoSession, err)
}

func TestAuthData_ForToken(t *testing.T) {
	var expectedToken = "token"
	var authData = &AuthData{
		Accounts: []*AuthAccount{
			{
				HostAddr: "hostAddr",
				AuthSession: &AuthSession{
					Token: expectedToken,
				},
			},
		},
	}
	var actual, err = authData.ForToken(expectedToken)

	assert.NoError(t, err)
	assert.Same(t, authData.Accounts[0], actual)
}

func TestAuthData_Push(t *testing.T) {
	var expectedHost = "hostAddr"
	var authSession = &AuthSession{
		Token: "token",
	}
	var authData = &AuthData{
		Accounts: []*AuthAccount{
			{
				HostAddr: expectedHost,
			},
			{
				HostAddr: "hostAddr2",
			},
		},
	}
	var actual = authData.Push(expectedHost, authSession)

	assert.Equal(t, authSession.Token, actual.Accounts[0].AuthSession.Token)
	assert.Same(t, authData.Accounts[1], actual.Accounts[1])
}

func TestAuthData_Write_PermissionDenied(t *testing.T) {
	var testDir = t.TempDir()
	var testFile = filepath.Join(testDir, "notWriteable", ".auth")
	var authData = NewAuthData()

	if err := os.MkdirAll(filepath.Dir(testFile), 0222); err == nil {
		assert.Error(t, authData.Write(testFile))
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestAuthData_Write(t *testing.T) {
	var expectedFile = filepath.Join(t.TempDir(), "genaiz.login")
	var authData = &AuthData{
		Accounts: []*AuthAccount{
			{
				HostAddr:    "test",
				AuthSession: &AuthSession{},
			},
		},
	}
	var actual *AuthData

	assert.NoError(t, authData.Write(expectedFile))
	assert.NoError(t, filez.IsReadable(expectedFile))
	actual = NewAuthData(expectedFile)
	assert.Equal(t, authData.Accounts[0].HostAddr, actual.Accounts[0].HostAddr)
}

func TestAuthSession_IsExpired(t *testing.T) {
	assert.True(t, NewAuthSession(&Session{Expiry: 0}, "username", "token").IsExpired())
	assert.False(t, NewAuthSession(&Session{Expiry: -1}, "username", "token").IsExpired())
}

func TestNewAuthTask(t *testing.T) {
	var actual = NewAuthTask()

	assert.NotEmpty(t, actual.Name)
	assert.NotEmpty(t, actual.OnPrepare)
	assert.NotEmpty(t, actual.OnComplete)
	assert.Empty(t, actual.OnPretend)
}

func TestNewLoginTask(t *testing.T) {
	var actual = NewLoginTask()

	assert.NotEmpty(t, actual.Name)
	assert.NotEmpty(t, actual.OnPrepare)
	assert.NotEmpty(t, actual.OnComplete)
	assert.Empty(t, actual.OnPretend)
}

func TestNewLogoutTask(t *testing.T) {
	var actual = NewLogoutTask()

	assert.NotEmpty(t, actual.Name)
	assert.NotEmpty(t, actual.OnPrepare)
	assert.NotEmpty(t, actual.OnComplete)
	assert.Empty(t, actual.OnPretend)
}

func TestNewSessionTask(t *testing.T) {
	var actual = NewSessionTask()

	assert.NotEmpty(t, actual.Name)
	assert.NotEmpty(t, actual.OnPrepare)
	assert.NotEmpty(t, actual.OnComplete)
	assert.Empty(t, actual.OnPretend)
}

func Test_handleAuthContext(t *testing.T) {
	var testParams = &AuthParams{
		Broker: &Broker{
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
				assert.NoError(t, handleAuthContext(testParams, testState))
				return
			}
		}
	}

	assert.NoError(t, err)
}

func Test_handleAuthContext_NoAccountsFound(t *testing.T) {
	var testParams = &AuthParams{
		Broker: &Broker{
			AuthFile: filepath.Join(t.TempDir(), ".auth"),
		},
	}
	var testState = &task.State{
		Logger: logrus.New(),
	}

	err := handleAuthContext(testParams, testState)
	assert.Equal(t, errorSessionsNotFound(), err)
}

func Test_handleAuthList(t *testing.T) {
	var testParams = &AuthParams{
		Broker: &Broker{
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
				assert.NoError(t, handleAuthList(testParams, testState))
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

func Test_handleAuthList_OutActiveExpired(t *testing.T) {
	var testParams = &AuthParams{
		Broker: &Broker{
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
				assert.NoError(t, handleAuthList(testParams, testState))
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

func Test_handleAuthList_NegActiveNoExpired(t *testing.T) {
	var testParams = &AuthParams{
		Broker: &Broker{
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
				assert.NoError(t, handleAuthList(testParams, testState))
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

func Test_handleLoginContext_EmptyFile(t *testing.T) {
	var dir = t.TempDir()

	if fd, err := os.CreateTemp(dir, "genaiz.login*"); err == nil {
		var testState = &task.State{Logger: logrus.New()}
		var testParams = &LoginParams{
			Broker: &Broker{
				AuthFile: fd.Name(),
				HostAddr: "hostAddr",
			},
		}

		assert.NoError(t, handleLoginContext(testParams, testState))
	} else {
		assert.Fail(t, err.Error())
	}
}

func Test_handleLoginContext(t *testing.T) {
	var dir = t.TempDir()

	if fd, err := os.CreateTemp(dir, "genaiz.login*"); err == nil {
		var authData = NewAuthData(fd.Name())
		var testState = &task.State{Logger: logrus.New()}
		var testParams = &LoginParams{
			Broker: &Broker{
				AuthFile: fd.Name(),
				HostAddr: "hostAddr",
			},
		}

		authData = authData.Push("hostAddr2", &AuthSession{
			Token: "token",
		})
		panicz.PanicIfError(authData.Write(fd.Name()))
		assert.NoError(t, handleLoginContext(testParams, testState))
	} else {
		assert.Fail(t, err.Error())
	}
}

func Test_handleLoginCreate_LoginErr(t *testing.T) {
	var dir = t.TempDir()

	if fd, err := os.CreateTemp(dir, "genaiz.login*"); err == nil {
		var restoredFactory = clientFactory.New
		var testState = &task.State{Logger: logrus.New()}
		var testParams = &LoginParams{
			Broker: &Broker{
				AuthFile: fd.Name(),
				HostAddr: "hostAddr",
			},
			Username: "username",
			Password: []byte("password"),
		}
		var expectedError = errors.New("expected")

		defer func() {
			clientFactory.New = restoredFactory
		}()
		clientFactory.New = func(hostAddr string) Client {
			return &stubLoginClient{
				loginErr: expectedError,
			}
		}

		assert.ErrorIs(t, handleLoginCreate(testParams, testState), expectedError)
	}
}

func Test_handleLoginCreate_TokenPresent(t *testing.T) {
	assert.NoError(t, handleLoginCreate(&LoginParams{}, &task.State{Output: "token"}))
}

func Test_handleLoginCreate_WriteErr(t *testing.T) {
	var dir = t.TempDir()

	if fd, err := os.CreateTemp(dir, "genaiz.login*"); err == nil {
		var restoredFactory = clientFactory.New
		var testState = &task.State{Logger: logrus.New()}
		var testParams = &LoginParams{
			Broker: &Broker{
				AuthFile: fd.Name(),
				HostAddr: "hostAddr",
			},
			Username: "username",
			Password: []byte("password"),
		}

		panicz.PanicIfError(os.Chmod(fd.Name(), 0400))

		defer func() {
			clientFactory.New = restoredFactory
		}()
		clientFactory.New = func(hostAddr string) Client {
			return &stubLoginClient{
				session: &AuthSession{
					Token:    "token",
					Username: "user",
				},
			}
		}

		assert.Error(t, handleLoginCreate(testParams, testState))
	}
}

func Test_handleLoginCreate(t *testing.T) {
	var dir = t.TempDir()

	if fd, err := os.CreateTemp(dir, "genaiz.login*"); err == nil {
		var expectedHost = "hostAddr"
		var expectedUser = "username"
		var expectedToken = "token"
		var testState = &task.State{Logger: logrus.New()}
		var testParams = &LoginParams{
			Broker: &Broker{
				AuthFile: fd.Name(),
				HostAddr: expectedHost,
			},
			Username: expectedUser,
			Password: []byte("password"),
		}
		var restoredFactory = clientFactory.New
		var actual *AuthAccount

		defer func() {
			clientFactory.New = restoredFactory
		}()
		clientFactory.New = func(hostAddr string) Client {
			return &stubLoginClient{
				session: &AuthSession{
					Token:    expectedToken,
					Username: expectedUser,
				},
			}
		}

		assert.NoError(t, handleLoginCreate(testParams, testState))

		if actual, err = NewAuthData(fd.Name()).ForHostUser(expectedHost, expectedUser); err == nil {
			assert.Equal(t, actual.Token, expectedToken)
		}

		assert.NoError(t, err)
	}
}

func Test_handleLoginDelete_NoSession(t *testing.T) {
	var testState = &task.State{}
	var testParams = &LoginParams{}

	assert.ErrorIs(t, ErrorNoSession, handleLoginDelete(testParams, testState))
}

func Test_handleLoginDelete_LogoutFailed(t *testing.T) {
	var dir = t.TempDir()

	if fd, err := os.CreateTemp(dir, "genaiz.logout*"); err == nil {
		var expectedSessionId = int64(37)
		var testState = &task.State{
			Logger: logrus.New(),
			Output: cast.ToString(expectedSessionId),
		}
		var testParams = &LoginParams{
			Broker: &Broker{
				AuthFile: fd.Name(),
				HostAddr: "hostAddr",
			},
		}
		var restoredFactory = clientFactory.Get

		defer func() {
			clientFactory.Get = restoredFactory
		}()
		clientFactory.Get = func(authFile, addr string) (Client, error) {
			return &stubLoginClient{
				logoutErr: errors.New("unexpected"),
			}, nil
		}

		assert.NoError(t, handleLoginDelete(testParams, testState))
	} else {
		assert.Fail(t, err.Error())
	}
}

func Test_handleLoginDelete(t *testing.T) {
	var dir = t.TempDir()

	if fd, err := os.CreateTemp(dir, "genaiz.logout*"); err == nil {
		var expectedHost = "hostAddr"
		var expectedSessionId = int64(37)
		var testSession = &AuthSession{
			SessionId: expectedSessionId,
		}
		var testState = &task.State{
			Logger: logrus.New(),
			Output: cast.ToString(expectedSessionId),
		}
		var testParams = &LoginParams{
			Broker: &Broker{
				AuthFile: fd.Name(),
				HostAddr: expectedHost,
			},
		}
		var restoredFactory = clientFactory.Get

		defer func() {
			clientFactory.Get = restoredFactory
		}()
		clientFactory.Get = func(authFile, addr string) (Client, error) {
			return &stubLoginClient{
				session: testSession,
			}, nil
		}

		panicz.PanicIfError(NewAuthData().
			Push(expectedHost, testSession).
			Push("hostAddr2", &AuthSession{}).
			Write(fd.Name()))
		assert.NoError(t, handleLoginDelete(testParams, testState))
		assert.Equal(t, 1, len(NewAuthData(fd.Name()).Accounts))
	} else {
		assert.Fail(t, err.Error())
	}
}

func Test_handleLogoutContext_NoAuth(t *testing.T) {
	var dir = t.TempDir()

	if fd, err := os.CreateTemp(dir, "genaiz.logout*"); err == nil {
		var testState = &task.State{
			Logger: logrus.New(),
		}
		var testParams = &LoginParams{
			Broker: &Broker{
				AuthFile: fd.Name(),
			},
		}

		assert.ErrorIs(t, ErrorNoLogin, handleLogoutContext(testParams, testState))
	} else {
		assert.Fail(t, err.Error())
	}
}

func Test_handleLogoutContext_NoForHostUser(t *testing.T) {
	var dir = t.TempDir()

	if fd, err := os.CreateTemp(dir, "genaiz.logout*"); err == nil {
		var testSession = &AuthSession{
			Token: "token",
		}
		var testState = &task.State{
			Logger: logrus.New(),
		}
		var testParams = &LoginParams{
			Broker: &Broker{
				AuthFile: fd.Name(),
				HostAddr: "hostAddr",
			},
			Username: "username",
		}

		panicz.PanicIfError(NewAuthData().
			Push("hostAddr2", testSession).
			Write(fd.Name()))
		assert.ErrorIs(t, ErrorNoSession, handleLogoutContext(testParams, testState))
	} else {
		assert.Fail(t, err.Error())
	}
}

func Test_handleLogoutContext_SingleLogin(t *testing.T) {
	var dir = t.TempDir()

	if fd, err := os.CreateTemp(dir, "genaiz.logout*"); err == nil {
		var expectedSessionId = int64(37)
		var testSession = &AuthSession{
			Token:     "token",
			SessionId: expectedSessionId,
		}
		var testState = &task.State{
			Logger: logrus.New(),
		}
		var testParams = &LoginParams{
			Broker: &Broker{
				AuthFile: fd.Name(),
			},
		}
		var testAuth = NewAuthData().
			Push("hostAddr2", testSession)

		// Intentionally set an invalid active account
		testAuth.Active = 1
		panicz.PanicIfError(testAuth.Write(fd.Name()))
		assert.NoError(t, handleLogoutContext(testParams, testState))
		assert.Equal(t, cast.ToString(expectedSessionId), testState.Output)
	} else {
		assert.Fail(t, err.Error())
	}
}

func Test_handleLogoutContext_TooManyLogins(t *testing.T) {
	var dir = t.TempDir()

	if fd, err := os.CreateTemp(dir, "genaiz.logout*"); err == nil {
		var testSession = &AuthSession{
			Token: "token",
		}
		var testState = &task.State{
			Logger: logrus.New(),
		}
		var testParams = &LoginParams{
			Broker: &Broker{
				AuthFile: fd.Name(),
			},
		}
		var testAuth = NewAuthData().
			Push("hostAddr2", testSession).
			Push("hostAddr3", testSession)

		// Intentionally set an invalid active account
		testAuth.Active = 2
		panicz.PanicIfError(testAuth.Write(fd.Name()))
		assert.ErrorIs(t, handleLogoutContext(testParams, testState), ErrorSessionConflict)
	} else {
		assert.Fail(t, err.Error())
	}
}

func Test_handleLogoutContext_UsernameLogin(t *testing.T) {
	var dir = t.TempDir()

	if fd, err := os.CreateTemp(dir, "genaiz.logout*"); err == nil {
		var expectedSessionId = int64(37)
		var expectedUser = "username"
		var testSession = &AuthSession{
			Token:     "token",
			SessionId: expectedSessionId,
			Username:  expectedUser,
		}
		var testState = &task.State{
			Logger: logrus.New(),
		}
		var testParams = &LoginParams{
			Broker: &Broker{
				AuthFile: fd.Name(),
				HostAddr: "hostAddr",
			},
			Username: expectedUser,
		}

		defer filez.CloseSilently(fd)
		panicz.PanicIfError(NewAuthData().
			Push("hostAddr", testSession).
			Write(fd.Name()))
		assert.NoError(t, handleLogoutContext(testParams, testState))
		assert.Equal(t, cast.ToString(expectedSessionId), testState.Output)
	} else {
		assert.Fail(t, err.Error())
	}
}

func Test_handleLogoutContext(t *testing.T) {
	var dir = t.TempDir()

	if fd, err := os.CreateTemp(dir, "genaiz.logout*"); err == nil {
		var expectedSessionId = int64(37)
		var expectedUser = "username"
		var testSession = &AuthSession{
			Token:     "token",
			SessionId: expectedSessionId,
			Username:  expectedUser,
		}
		var testState = &task.State{
			Logger: logrus.New(),
		}
		var testParams = &LoginParams{
			Broker: &Broker{
				AuthFile: fd.Name(),
			},
		}

		panicz.PanicIfError(NewAuthData().
			Push("hostAddr", testSession).
			Write(fd.Name()))
		assert.NoError(t, handleLogoutContext(testParams, testState))
		assert.Equal(t, cast.ToString(expectedSessionId), testState.Output)
	} else {
		assert.Fail(t, err.Error())
	}
}

func Test_handleSessionActivate(t *testing.T) {
	var testAuthFile = filepath.Join(t.TempDir(), "authTest")

	if fd, err := os.Create(testAuthFile); err == nil {
		var expectedToken = "token"
		var testSession = &AuthSession{
			Expiry: time.Now().Add(time.Hour).UnixMilli(),
			Token:  expectedToken,
		}
		var testState = &task.State{
			Logger: logrus.New(),
			Output: expectedToken,
		}
		var restoredFactory = clientFactory.New
		var testParams = &Broker{
			AuthFile: fd.Name(),
			HostAddr: "hostAddr",
		}

		panicz.PanicIfError(NewAuthData().
			Push("hostAddr", testSession).
			Write(fd.Name()))

		defer filez.CloseSilently(fd)
		defer func() {
			clientFactory.New = restoredFactory
		}()
		clientFactory.New = func(hostAddr string) Client {
			return &stubLoginClient{
				session: &AuthSession{
					Token:    "token",
					Username: "user",
				},
			}
		}

		assert.NoError(t, handleSessionActivate(testParams, testState))
		actual := NewAuthData(fd.Name())
		assert.Equal(t, 1, len(actual.Accounts))
	} else {
		assert.Fail(t, err.Error())
	}
}

func Test_handleSessionActivate_NoSessionAccount(t *testing.T) {
	assert.ErrorIs(t, handleSessionActivate(&Broker{}, &task.State{}), ErrorNoSession)
}

func Test_handleSessionActivate_NoSessionToken(t *testing.T) {
	var testAuthFile = filepath.Join(t.TempDir(), "authTest")

	if fd, err := os.Create(testAuthFile); err == nil {
		var testState = task.State{Output: "token"}

		defer filez.CloseSilently(fd)
		assert.ErrorIs(t, handleSessionActivate(&Broker{}, &testState), ErrorNoSession)
	} else {
		assert.Fail(t, err.Error())
	}
}

func Test_handleSessionActivate_WriteError(t *testing.T) {
	var testAuthFile = filepath.Join(t.TempDir(), "authTest")

	if fd, err := os.Create(testAuthFile); err == nil {
		var expectedToken = "token"
		var testSession = &AuthSession{
			Expiry: time.Now().Add(time.Hour).UnixMilli(),
			Token:  expectedToken,
		}
		var testState = &task.State{
			Logger: logrus.New(),
			Output: expectedToken,
		}
		var restoredFactory = clientFactory.New
		var testParams = &Broker{
			AuthFile: fd.Name(),
			HostAddr: "hostAddr",
		}

		panicz.PanicIfError(NewAuthData().
			Push("hostAddr", testSession).
			Write(fd.Name()))
		panicz.PanicIfError(os.Chmod(fd.Name(), 0400))

		defer filez.CloseSilently(fd)
		defer func() {
			clientFactory.New = restoredFactory
		}()
		clientFactory.New = func(hostAddr string) Client {
			return &stubLoginClient{
				session: &AuthSession{
					Token:    "token",
					Username: "user",
				},
			}
		}

		assert.Error(t, handleSessionActivate(testParams, testState))
	} else {
		assert.Fail(t, err.Error())
	}
}

func Test_handleSessionContext_NoAuth(t *testing.T) {
	var testState = &task.State{
		Logger: logrus.New(),
	}
	var testBroker = &Broker{
		AuthFile: "not_exist",
	}

	assert.ErrorIs(t, ErrorNoAuth, handleSessionContext(testBroker, testState))
}

func Test_handleSessionContext_NoHostSession(t *testing.T) {
	var dir = t.TempDir()

	if fd, err := os.CreateTemp(dir, "genaiz.logout*"); err == nil {
		var expectedToken = "token"
		var testSession = &AuthSession{
			Expiry: time.Now().Add(time.Hour).UnixMilli(),
			Token:  expectedToken,
		}
		var testState = &task.State{
			Logger: logrus.New(),
		}
		var testBroker = &Broker{
			AuthFile: fd.Name(),
			HostAddr: "hostAddr2",
		}

		panicz.PanicIfError(NewAuthData().
			Push("hostAddr", testSession).
			Write(fd.Name()))
		assert.ErrorIs(t, ErrorNoSession, handleSessionContext(testBroker, testState))
	} else {
		assert.Fail(t, err.Error())
	}
}

func Test_handleSessionContext_ActiveExpired(t *testing.T) {
	var dir = t.TempDir()

	if fd, err := os.CreateTemp(dir, "genaiz.logout*"); err == nil {
		var testSession = &AuthSession{
			Expiry: 0,
		}
		var testState = &task.State{
			Logger: logrus.New(),
		}
		var testBroker = &Broker{
			AuthFile: fd.Name(),
		}

		panicz.PanicIfError(NewAuthData().
			Push("hostAddr", testSession).
			Write(fd.Name()))
		assert.ErrorIs(t, ErrorNoAuth, handleSessionContext(testBroker, testState))
	}
}

func Test_handleSessionContext_Active(t *testing.T) {
	var dir = t.TempDir()

	if fd, err := os.CreateTemp(dir, "genaiz.logout*"); err == nil {
		var expectedToken = "token"
		var testSession = &AuthSession{
			Expiry: time.Now().Add(time.Hour).UnixMilli(),
			Token:  expectedToken,
		}
		var testState = &task.State{
			Logger: logrus.New(),
		}
		var testBroker = &Broker{
			AuthFile: fd.Name(),
		}

		panicz.PanicIfError(NewAuthData().
			Push("hostAddr", testSession).
			Write(fd.Name()))
		assert.NoError(t, handleSessionContext(testBroker, testState))
		assert.Equal(t, expectedToken, testState.Output)
	}
}

func Test_handleSessionContext(t *testing.T) {
	var dir = t.TempDir()

	if fd, err := os.CreateTemp(dir, "genaiz.logout*"); err == nil {
		var expectedHost = "hostAddr"
		var expectedToken = "token"
		var testSession = &AuthSession{
			Expiry: time.Now().Add(time.Hour).UnixMilli(),
			Token:  expectedToken,
		}
		var testState = &task.State{
			Logger: logrus.New(),
		}
		var testBroker = &Broker{
			AuthFile: fd.Name(),
			HostAddr: expectedHost,
		}

		panicz.PanicIfError(NewAuthData().
			Push(expectedHost, testSession).
			Write(fd.Name()))
		assert.ErrorIs(t, handleSessionContext(testBroker, testState), ErrorSessionExists)
		assert.Equal(t, expectedToken, testState.Output)
	} else {
		assert.Fail(t, err.Error())
	}
}

func Test_handleSessionValidate_NoAuth(t *testing.T) {
	assert.ErrorIs(t, ErrorNoAuth, handleSessionValidate(&Broker{}, &task.State{}))
}

func Test_handleSessionValidate_NoHostSession(t *testing.T) {
	var expectedError = errors.New("expected")
	var testState = &task.State{
		Error:  errors.New("test"),
		Logger: logrus.New(),
		Output: "output",
	}
	var testParams = &Broker{
		AuthFile: "file",
		HostAddr: "hostAddr",
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return nil, expectedError
	}

	assert.ErrorIs(t, expectedError, handleSessionValidate(testParams, testState))
	assert.Empty(t, testState.Output)
}

func Test_handleSessionValidate_SessionError(t *testing.T) {
	var testState = &task.State{
		Error:  errors.New("test"),
		Logger: logrus.New(),
		Output: "output",
	}
	var testParams = &Broker{
		AuthFile: "file",
		HostAddr: "hostAddr",
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubLoginClient{}, nil
	}

	assert.Error(t, handleSessionValidate(testParams, testState))
	assert.Empty(t, testState.Output)
}

func Test_handleSessionValidate_SessionInvalid(t *testing.T) {
	var testState = &task.State{
		Error:  errors.New("test"),
		Logger: logrus.New(),
		Output: "output",
	}
	var testParams = &Broker{
		AuthFile: "file",
		HostAddr: "hostAddr",
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubLoginClient{
			client: client{
				UserId: 37,
			},
			session: &AuthSession{},
		}, nil
	}

	assert.ErrorIs(t, ErrorSessionInvalid, handleSessionValidate(testParams, testState))
	assert.Empty(t, testState.Output)
}

func Test_handleSessionValidate(t *testing.T) {
	var expectedOutput = "session"
	var testState = &task.State{
		Error:  errors.New("test"),
		Logger: logrus.New(),
		Output: expectedOutput,
	}
	var testParams = &Broker{
		AuthFile: "file",
		HostAddr: "hostAddr",
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubLoginClient{
			client: client{
				UserId: 0,
			},
			session: &AuthSession{},
		}, nil
	}

	assert.NoError(t, handleSessionValidate(testParams, testState))
	assert.Equal(t, expectedOutput, testState.Output)
}
